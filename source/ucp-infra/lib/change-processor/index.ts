// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as glueAlpha from '@aws-cdk/aws-glue-alpha';
import { KinesisStreamsToLambda } from '@aws-solutions-constructs/aws-kinesisstreams-lambda';
import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as events from 'aws-cdk-lib/aws-events';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as kinesis from 'aws-cdk-lib/aws-kinesis';
import * as firehose from 'aws-cdk-lib/aws-kinesisfirehose';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as tahCore from '../../tah-cdk-common/core';
import * as tahSqs from '../../tah-cdk-common/sqs';
import { CdkBase, CdkBaseProps, FirehoseOutput, LambdaProps } from '../cdk-base';
import { ProfileStorageOutput } from '../profile-storage';

type ChangeProcessorProps = {
    readonly bucket: s3.IBucket;
    readonly bucketPrefix: string;
    readonly eventBridgeEnabled: string;
    readonly exportStreamMode: string;
    readonly exportStreamShardCount: number;
    readonly glueDatabase: glueAlpha.Database;
    readonly glueTravelerTable: glueAlpha.S3Table;
    readonly profileStorageOutput: ProfileStorageOutput;
    readonly uptVpc: ec2.IVpc;
} & LambdaProps;

export class ChangeProcessor extends CdkBase {
    public changeProcessorKinesisLambda: KinesisStreamsToLambda;
    public dlqChangeProcessor: tahSqs.Queue;
    public deliveryStreamLogGroup: logs.LogGroup;
    public deliveryStream: firehose.CfnDeliveryStream;

    private bus: events.EventBus;
    public firehoseFormatConversionRole: iam.Role;
    private props: ChangeProcessorProps;

    constructor(cdkProps: CdkBaseProps, props: ChangeProcessorProps) {
        super(cdkProps);

        this.props = props;

        // Dead letter queue
        this.dlqChangeProcessor = this.createDeadLetterQueue();

        // Firehose format conversion role
        this.firehoseFormatConversionRole = this.createFirehoseRole();

        // Firehose delivery stream
        const output = this.createChangeProcessorFirehoseDeliveryStream();

        // Firehose delivery stream
        this.deliveryStream = output.deliveryStream;
        this.deliveryStreamLogGroup = output.logGroup;

        // Event bus
        this.bus = this.createEventBus();

        // Change processor Kinesis/Lambda
        this.changeProcessorKinesisLambda = this.createChangeProcessorKinesisLambda();

        // Permissions
        this.dlqChangeProcessor.grantConsumeMessages(this.changeProcessorKinesisLambda.lambdaFunction);
        this.dlqChangeProcessor.grantSendMessages(this.changeProcessorKinesisLambda.lambdaFunction);
        this.bus.grantPutEventsTo(this.changeProcessorKinesisLambda.lambdaFunction);

        // Outputs
        tahCore.Output.add(this.stack, 'travellerEventBus', this.bus.eventBusName);
        tahCore.Output.add(
            this.stack,
            'kinesisStreamOutputNameChangeProcessor',
            this.changeProcessorKinesisLambda.kinesisStream.streamName
        );
        tahCore.Output.add(
            this.stack,
            'kinesisStreamOutputNameChangeProcessorArn',
            this.changeProcessorKinesisLambda.kinesisStream.streamArn
        );
    }

    /**
     * Creates a dead letter queue.
     * @returns Dead letter queue for change processor
     */
    private createDeadLetterQueue(): tahSqs.Queue {
        return new tahSqs.Queue(this.stack, 'ucp-change-processor-' + this.envName);
    }

    /**
     * Creates a Kinesis Firehose format conversion role.
     * This allows Firehose to access Glue table.
     * @returns Firehose format conversion role
     */
    private createFirehoseRole(): iam.Role {
        return new iam.Role(this.stack, 'firehoseFormatConversionRole', {
            assumedBy: new iam.ServicePrincipal('firehose.amazonaws.com'),
            description: 'Role for Firehose to convert data formats',
            inlinePolicies: {
                firehoseFormatConversionPolicy: new iam.PolicyDocument({
                    statements: [
                        new iam.PolicyStatement({
                            actions: ['glue:GetTable', 'glue:GetTableVersion', 'glue:GetTableVersions'],
                            resources: [
                                this.props.glueDatabase.catalogArn,
                                this.props.glueDatabase.databaseArn,
                                this.props.glueTravelerTable.tableArn
                            ]
                        })
                    ]
                })
            }
        });
    }

    /**
     * Creates a change processor Kinesis Firehose delivery stream.
     * @returns Kinesis Firehose delivery stream
     */
    private createChangeProcessorFirehoseDeliveryStream(): FirehoseOutput {
        const dataFormatConversionConfiguration: firehose.CfnDeliveryStream.DataFormatConversionConfigurationProperty = {
            enabled: true,
            inputFormatConfiguration: {
                deserializer: {
                    hiveJsonSerDe: {}
                }
            },
            outputFormatConfiguration: {
                serializer: {
                    parquetSerDe: {
                        compression: 'GZIP'
                    }
                }
            },
            schemaConfiguration: {
                databaseName: this.props.glueDatabase.databaseName,
                region: cdk.Aws.REGION,
                roleArn: this.firehoseFormatConversionRole.roleArn,
                tableName: this.props.glueTravelerTable?.tableName,
                versionId: 'LATEST'
            }
        };

        return this.createFirehoseDeliveryStream({
            bucket: this.props.bucket,
            bucketPrefix: this.props.bucketPrefix,
            idPrefix: 'ucpChangeProcessor',
            idSuffix: this.envName,
            partition: { jqString: '.domain', name: 'domainname' },
            dataFormatConversionConfiguration
        });
    }

    /**
     * Creates an EventBridge event bus.
     * @returns EventBridge event bus
     */
    private createEventBus(): events.EventBus {
        return new events.EventBus(this.stack, 'ucpTravellerChangeBus', {
            eventBusName: 'ucp-traveller-changes-' + this.envName
        });
    }

    /**
     * Creates a Kinesis stream and a Lambda function for change processing.
     * @returns KinesisSteamsToLambda Solutions Construct
     */
    private createChangeProcessorKinesisLambda(): KinesisStreamsToLambda {
        let streamMode = kinesis.StreamMode.ON_DEMAND;
        let shardCount: number | undefined = undefined;

        if (this.props.exportStreamMode === kinesis.StreamMode.PROVISIONED.toString()) {
            streamMode = kinesis.StreamMode.PROVISIONED;
            shardCount = this.props.exportStreamShardCount;
        }

        const streamProp: kinesis.StreamProps = {
            retentionPeriod: cdk.Duration.days(30),
            shardCount,
            streamMode,
            removalPolicy: cdk.RemovalPolicy.DESTROY
        };

        // Kinesis Stream to Lambda constructs to manager ACCP change notifications
        const ucpChangeProcessorPrefix = 'ucpChangeProcessor';
        const changeProcessorKinesisLambda = new KinesisStreamsToLambda(this.stack, ucpChangeProcessorPrefix + this.envName, {
            kinesisEventSourceProps: {
                batchSize: 100,
                parallelizationFactor: 10,
                maximumRetryAttempts: 3,
                maxRecordAge: cdk.Duration.days(7),
                startingPosition: lambda.StartingPosition.TRIM_HORIZON
            },
            kinesisStreamProps: streamProp,
            lambdaFunctionProps: {
                runtime: lambda.Runtime.PROVIDED_AL2,
                architecture: lambda.Architecture.ARM_64,
                handler: 'main',
                timeout: cdk.Duration.seconds(900),
                memorySize: 1024,
                code: new lambda.S3Code(this.props.artifactBucket, `${this.props.artifactBucketPath}/${ucpChangeProcessorPrefix}.zip`),
                deadLetterQueueEnabled: true,
                deadLetterQueue: this.dlqChangeProcessor,
                functionName: ucpChangeProcessorPrefix + this.envName,
                environment: {
                    LAMBDA_REGION: cdk.Aws.REGION,
                    S3_OUTPUT: this.props.bucket.bucketName,
                    FIREHOSE_STREAM_NAME: this.deliveryStream.deliveryStreamName ?? '',
                    // For usage metrics
                    SEND_ANONYMIZED_DATA: this.props.sendAnonymizedData,
                    METRICS_SOLUTION_ID: this.props.solutionId,
                    METRICS_SOLUTION_VERSION: this.props.solutionVersion,
                    // For low-cost storage
                    AURORA_PROXY_READONLY_ENDPOINT: this.props.profileStorageOutput.storageProxyReadonlyEndpoint,
                    AURORA_DB_NAME: this.props.profileStorageOutput.storageDbName,
                    AURORA_DB_SECRET_ARN: this.props.profileStorageOutput.storageSecretArn,
                    STORAGE_CONFIG_TABLE_NAME: this.props.profileStorageOutput.storageConfigTable?.tableName ?? '',
                    STORAGE_CONFIG_TABLE_PK: this.props.profileStorageOutput.storageConfigTablePk,
                    STORAGE_CONFIG_TABLE_SK: this.props.profileStorageOutput.storageConfigTableSk
                },
                // Low cost storage security config
                vpc: this.props.uptVpc,
                securityGroups: [this.props.profileStorageOutput.lambdaToProxyGroup]
            }
        });
        changeProcessorKinesisLambda.lambdaFunction.addEnvironment(
            'CHANGE_PROC_KINESIS_STREAM_NAME',
            changeProcessorKinesisLambda.kinesisStream.streamName
        );

        changeProcessorKinesisLambda.lambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['profile:GetMatches', 'profile:ListProfileObjects', 'profile:SearchProfiles'],
                effect: iam.Effect.ALLOW,
                resources: ['*']
            })
        );
        changeProcessorKinesisLambda.lambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['firehose:PutRecordBatch'],
                effect: iam.Effect.ALLOW,
                resources: [this.deliveryStream.attrArn]
            })
        );

        // Low-cost storage permissions
        // Tables are created/deleted via API and cannot be scoped down
        changeProcessorKinesisLambda.lambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['dynamodb:DescribeTable', 'dynamodb:Query', 'dynamodb:Scan'],
                effect: iam.Effect.ALLOW,
                resources: ['*']
            })
        );

        changeProcessorKinesisLambda.lambdaFunction.addEnvironment('EVENTBRIDGE_ENABLED', this.props.eventBridgeEnabled);
        changeProcessorKinesisLambda.lambdaFunction.addEnvironment('TRAVELLER_EVENT_BUS', this.bus.eventBusName);
        changeProcessorKinesisLambda.lambdaFunction.addEnvironment('DEAD_LETTER_QUEUE_URL', this.dlqChangeProcessor.queueUrl);

        return changeProcessorKinesisLambda;
    }
}
