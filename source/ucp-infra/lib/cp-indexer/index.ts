// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import * as cdk from 'aws-cdk-lib';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as kinesis from 'aws-cdk-lib/aws-kinesis';
import * as kms from 'aws-cdk-lib/aws-kms';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as lambdaEventSources from 'aws-cdk-lib/aws-lambda-event-sources';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import path from 'path';
import { Output } from '../../tah-cdk-common/core';
import * as tahSqs from '../../tah-cdk-common/sqs';
import { CdkBase, CdkBaseProps, DynamoDbTableKey, LambdaProps } from '../cdk-base';

type CpIndexerProps = {
    maxConcurrency: number;
    readonly dynamoDbKey: kms.Key;
} & LambdaProps;

export type CPIndexerOutput = {
    cpExportStream: kinesis.Stream;
    cpIndexTable: dynamodb.Table;
    indexLambda: lambda.Function;
    cpIndexerDLQ: tahSqs.Queue;
};

type CpIndexDynamoDbTables = {
    readonly cpIndexTable: dynamodb.Table;
};

export class CPIndexer extends CdkBase {
    public cpIndexerDLQ: sqs.Queue;
    public indexLambda: lambda.Function;
    public readonly cpIndexTable: dynamodb.Table;
    public readonly cpExportStream: kinesis.Stream;
    private readonly kinesisBatchSize = 10;

    private props: CpIndexerProps;

    constructor(cdkProps: CdkBaseProps, props: CpIndexerProps) {
        super(cdkProps);

        this.props = props;

        //CP Index Dynamo table
        this.cpIndexTable = this.createCpIndexDynamoDbTable().cpIndexTable;

        //CP export stream
        this.cpExportStream = this.createCpExportStream();

        // Indexer DLQ
        this.cpIndexerDLQ = this.createIndexerDLQ();

        // Indexer Lambda
        this.indexLambda = this.createCpIndexerLambda();

        // Kinesis to Lambda event source
        this.indexLambda.addEventSource(
            new lambdaEventSources.KinesisEventSource(this.cpExportStream, {
                startingPosition: lambda.StartingPosition.TRIM_HORIZON,
                batchSize: this.kinesisBatchSize,
                maxBatchingWindow: cdk.Duration.seconds(5),
                reportBatchItemFailures: true,
                bisectBatchOnError: true,
                retryAttempts: 2
            })
        );
        this.cpExportStream.grantReadWrite(this.indexLambda);
        this.cpIndexTable.grantReadWriteData(this.indexLambda);

        // Add outputs
        Output.add(this.stack, 'cpIndexTable', this.cpIndexTable.tableName);
        Output.add(this.stack, 'cpExportStream', this.cpExportStream.streamName);
        Output.add(this.stack, 'cpIndexerDLQ', this.cpIndexerDLQ.queueName);
    }

    private createCpIndexDynamoDbTable(): CpIndexDynamoDbTables {
        const cpIndexTable = new dynamodb.Table(this.stack, 'ucp', {
            tableName: 'ucp-cp-index-table-' + this.envName,
            partitionKey: { name: DynamoDbTableKey.CP_INDEX_TABLE_PK, type: dynamodb.AttributeType.STRING },
            sortKey: { name: DynamoDbTableKey.CP_INDEX_TABLE_SK, type: dynamodb.AttributeType.STRING },
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            encryptionKey: this.props.dynamoDbKey,
            pointInTimeRecovery: true
        });
        return { cpIndexTable };
    }

    private createCpExportStream(): kinesis.Stream {
        const stream = new kinesis.Stream(this.stack, 'ucpCpExportStream', {
            streamName: 'ucp-cp-export-stream-' + this.envName,
            streamMode: kinesis.StreamMode.ON_DEMAND,
            retentionPeriod: cdk.Duration.days(30),
            encryption: kinesis.StreamEncryption.KMS,
            encryptionKey: this.props.dynamoDbKey,
            removalPolicy: cdk.RemovalPolicy.DESTROY
        });

        (stream.node.defaultChild as kinesis.CfnStream).cfnOptions.metadata = {
            cfn_nag: {
                rules_to_suppress: [
                    {
                        id: 'W28',
                        reason: 'Required'
                    }
                ]
            }
        };

        return stream;
    }

    private createIndexerDLQ(): sqs.Queue {
        return new tahSqs.Queue(this.stack, 'ucp-cp-indexer-dlq' + this.envName);
    }

    private createCpIndexerLambda(): lambda.Function {
        const ucpCpIndexerLambdaPrefix = 'ucpCpIndexer';
        const ucpCpIndexerLambda = new GoFunction(this.stack, ucpCpIndexerLambdaPrefix + this.envName, {
            entry: path.join(__dirname, '..', '..', '..', 'ucp-cp-indexer', 'src', 'main', 'main.go'),
            bundling: {
                environment: { GOOS: 'linux', GOARCH: 'arm64', GOWORK: 'off' }
            },
            functionName: ucpCpIndexerLambdaPrefix + this.envName,
            runtime: lambda.Runtime.PROVIDED_AL2,
            architecture: lambda.Architecture.ARM_64,
            environment: {
                // General info
                LAMBDA_REGION: cdk.Aws.REGION,
                // Solution metrics
                METRICS_SOLUTION_ID: this.props.solutionId,
                METRICS_SOLUTION_VERSION: this.props.solutionVersion,
                // Dynamo
                DYNAMO_PK: DynamoDbTableKey.CP_INDEX_TABLE_PK,
                DYNAMO_SK: DynamoDbTableKey.CP_INDEX_TABLE_SK,
                DYNAMO_TABLE: this.cpIndexTable.tableName
            },
            tracing: lambda.Tracing.ACTIVE,
            deadLetterQueue: this.cpIndexerDLQ,
            deadLetterQueueEnabled: true,
            timeout: cdk.Duration.seconds(300)
        });
        return ucpCpIndexerLambda;
    }
}
