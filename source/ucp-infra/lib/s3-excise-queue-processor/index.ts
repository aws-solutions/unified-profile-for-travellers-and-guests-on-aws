// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { Database as GlueDatabase, S3Table as GlueS3Table } from '@aws-cdk/aws-glue-alpha';
import { Duration } from 'aws-cdk-lib';
import { Table as DynamoTable } from 'aws-cdk-lib/aws-dynamodb';
import { Effect, PolicyStatement } from 'aws-cdk-lib/aws-iam';
import { Architecture, Function as LambdaFunction, LayerVersion, Runtime, S3Code } from 'aws-cdk-lib/aws-lambda';
import { SqsEventSource } from 'aws-cdk-lib/aws-lambda-event-sources';
import { Bucket, IBucket } from 'aws-cdk-lib/aws-s3';
import { Queue, QueueEncryption } from 'aws-cdk-lib/aws-sqs';
import { StringParameter } from 'aws-cdk-lib/aws-ssm';
import { CdkBase, CdkBaseProps, DynamoDbTableKey } from '../cdk-base';

export type S3ExciseQueueProcessorProps = {
    readonly artifactBucket: IBucket;
    readonly artifactBucketPath: string;
    readonly outputBucket: Bucket;
    readonly travelerS3RootPath: string;
    readonly glueDatabase: GlueDatabase;
    readonly glueTravelerTable: GlueS3Table;
    readonly distributedMutexTable: DynamoTable;
    readonly privacySearchResultsTable: DynamoTable;
} & CdkBaseProps;

export class S3ExciseQueueProcessor extends CdkBase {
    private readonly props: S3ExciseQueueProcessorProps;
    private readonly functionAndVisibilityTimeout: Duration;

    // GDPR S3 Excise
    public readonly s3ExciseQueue: Queue;
    public readonly s3ExciseQueueLambdaProcessor: LambdaFunction;

    constructor(props: S3ExciseQueueProcessorProps) {
        super(props);

        this.props = props;
        this.functionAndVisibilityTimeout = Duration.minutes(2);

        this.s3ExciseQueue = this.createS3ExciseQueue();
        this.s3ExciseQueueLambdaProcessor = this.createS3ExciseLambdaFunction();
    }

    private createS3ExciseQueue(): Queue {
        const exciseQueue = new Queue(this.stack, 'ucpExciseQueue' + this.envName, {
            encryption: QueueEncryption.SQS_MANAGED,
            visibilityTimeout: this.functionAndVisibilityTimeout,
            retentionPeriod: Duration.days(14)
        });

        return exciseQueue;
    }

    private createS3ExciseLambdaFunction(): LambdaFunction {
        const sdkForPandasLayerArn = StringParameter.fromStringParameterAttributes(this.stack, 'sdkForPandasLayerArn', {
            parameterName: '/aws/service/aws-sdk-pandas/3.5.1/3.11/arm64/layer-arn'
        });
        const sdkForPandasLayer = LayerVersion.fromLayerVersionArn(
            this.stack,
            'aws-sdk-for-pandas-layer',
            sdkForPandasLayerArn.stringValue
        );

        const ucpAsyncLambdaPrefix = 'ucpS3ExciseQueueProcessor';
        const s3ExciseLambdaFunction = new LambdaFunction(this.stack, ucpAsyncLambdaPrefix + this.envName, {
            functionName: `${ucpAsyncLambdaPrefix}${this.envName}`,
            code: new S3Code(this.props.artifactBucket, `${this.props.artifactBucketPath}/${ucpAsyncLambdaPrefix}.zip`),
            handler: 'ucp_s3_excise_queue_processor.index.handler',
            runtime: Runtime.PYTHON_3_11,
            architecture: Architecture.ARM_64,
            environment: {
                S3_ROOT_PATH: this.props.travelerS3RootPath,
                GLUE_DATABASE: this.props.glueDatabase.databaseName,
                GLUE_TABLE: this.props.glueTravelerTable.tableName,
                DISTRIBUTED_MUTEX_TABLE: this.props.distributedMutexTable.tableName,
                DISTRIBUTED_MUTEX_TABLE_PK_NAME: DynamoDbTableKey.DISTRIBUTED_MUTEX_TABLE_PK,
                TTL_ATTRIBUTE_NAME: DynamoDbTableKey.DISTRIBUTED_MUTEX_TABLE_TTL_ATTR,
                PRIVACY_RESULTS_TABLE_NAME: this.props.privacySearchResultsTable.tableName,
                PRIVACY_RESULTS_PK_NAME: DynamoDbTableKey.PRIVACY_SEARCH_RESULTS_TABLE_PK,
                PRIVACY_RESULTS_SK_NAME: DynamoDbTableKey.PRIVACY_SEARCH_RESULTS_TABLE_SK
            },
            //  AWS SDK for Pandas recommends no less than 512 MB to prevent most OOM errors
            memorySize: 512,
            timeout: this.functionAndVisibilityTimeout,
            layers: [sdkForPandasLayer]
        });

        this.s3ExciseQueue.grantConsumeMessages(s3ExciseLambdaFunction);
        s3ExciseLambdaFunction.addEventSource(new SqsEventSource(this.s3ExciseQueue));
        this.props.outputBucket.grantReadWrite(s3ExciseLambdaFunction);
        s3ExciseLambdaFunction.addToRolePolicy(
            new PolicyStatement({
                actions: [
                    // Creating a partition on glue table for each domain
                    'glue:BatchCreatePartition',
                    'glue:BatchDeletePartition',
                    'glue:CreatePartition',
                    'glue:GetTable',
                    'glue:GetPartition',
                    'glue:GetPartitions',
                    'glue:UpdateTable'
                ],
                effect: Effect.ALLOW,
                resources: [this.props.glueDatabase.catalogArn, this.props.glueDatabase.databaseArn, this.props.glueTravelerTable.tableArn]
            })
        );

        this.props.privacySearchResultsTable.grantReadWriteData(s3ExciseLambdaFunction);
        this.props.distributedMutexTable.grantReadWriteData(s3ExciseLambdaFunction);

        return s3ExciseLambdaFunction;
    }
}
