// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as glueAlpha from '@aws-cdk/aws-glue-alpha';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { KinesisStreamsToLambda } from '@aws-solutions-constructs/aws-kinesisstreams-lambda';
import * as cdk from 'aws-cdk-lib';
import * as athena from 'aws-cdk-lib/aws-athena';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as events from 'aws-cdk-lib/aws-events';
import * as eventsTargets from 'aws-cdk-lib/aws-events-targets';
import { CfnJob } from 'aws-cdk-lib/aws-glue';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as kinesis from 'aws-cdk-lib/aws-kinesis';
import * as firehose from 'aws-cdk-lib/aws-kinesisfirehose';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as s3 from 'aws-cdk-lib/aws-s3';
import { Asset } from 'aws-cdk-lib/aws-s3-assets';
import * as s3Notifications from 'aws-cdk-lib/aws-s3-notifications';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import path from 'path';
import * as tahCore from '../../tah-cdk-common/core';
import * as tahS3 from '../../tah-cdk-common/s3';
import * as tahSqs from '../../tah-cdk-common/sqs';
import { CdkBase, CdkBaseProps, DynamoDbTableKey, LambdaProps } from '../cdk-base';
import { ProfileStorageOutput } from '../profile-storage';

export type DataIngestionProps = {
    readonly accessLogBucket: s3.Bucket;
    readonly athenaWorkGroup: athena.CfnWorkGroup;
    readonly configTable: dynamodb.Table;
    readonly connectProfileDomainErrorQueue: tahSqs.Queue;
    readonly connectProfileImportBucket: tahS3.Bucket;
    readonly dataLakeAdminRole: iam.Role;
    readonly deliveryStreamBucketPrefix: string;
    readonly entryStreamMode: string;
    readonly entryStreamShardCount: number;
    readonly glueDatabase: glueAlpha.Database;
    readonly industryConnectorBucketName: string;
    readonly ingestorShardCount: number;
    readonly matchBucket: s3.Bucket;
    readonly partitionStartDate: string;
    readonly region: string;
    readonly skipJobRun: string;
    readonly enableRealTimeBackup: boolean;
    // Low cost storage
    readonly profileStorageOutput: ProfileStorageOutput;
    readonly changeProcessorKinesisStream: kinesis.Stream;
    // VPC
    readonly uptVpc: ec2.IVpc;
} & LambdaProps;

type DeadLetterQueueOutput = {
    readonly dlqGo: tahSqs.Queue;
    readonly dlqPython: tahSqs.Queue;
    readonly dlqBatchProcessor: tahSqs.Queue;
};

type RealTimeProcessingOutput = {
    readonly entryKinesisLambda: KinesisStreamsToLambda;
    readonly ingestorKinesisLambda: KinesisStreamsToLambda;
};

export type BatchProcessingOutput = {
    readonly bucket: s3.Bucket;
    readonly errorQueue: sqs.Queue;
    readonly glueJob: glueAlpha.Job;
};

export class DataIngestion extends CdkBase {
    // Dead letter queues
    public readonly dlqGo: tahSqs.Queue;
    public readonly dlqPython: tahSqs.Queue;
    public readonly dlqBatchProcessor: tahSqs.Queue;
    // Real-time processing
    public readonly entryKinesisLambda: KinesisStreamsToLambda;
    public readonly ingestorKinesisLambda: KinesisStreamsToLambda;
    // Batch processing

    public readonly airBookingBatch: BatchProcessingOutput;
    public readonly clickStreamBatch: BatchProcessingOutput;
    public readonly customerServiceInteractionBatch: BatchProcessingOutput;
    public readonly guestProfileBatch: BatchProcessingOutput;
    public readonly hotelBookingBatch: BatchProcessingOutput;
    public readonly hotelStayBatch: BatchProcessingOutput;
    public readonly paxProfileBatch: BatchProcessingOutput;
    public readonly syncLambdaFunction: lambda.Function;
    public readonly batchLambdaFunction: lambda.Function;
    // Industry connector
    public readonly industryConnectorBucket: s3.IBucket;
    public readonly industryConnectorDataSyncRole: iam.Role;
    public readonly industryConnectorLambdaFunction: lambda.Function;
    private readonly kinesisBatchSize = 1;
    private readonly kinesisEventSourceProps = {
        batchSize: this.kinesisBatchSize,
        parallelizationFactor: 10,
        maximumRetryAttempts: 3,
        /**
         * CDK does not support -1 as a value so we use the max allowed. this is important because the default value (1 day) may prevent very large ingestions to succeed.
         * e.g. Customer ingests many millions of records that take more than a day to be processed. After a day some records will not be visible from Lambda and be skipped.
         */
        maxRecordAge: cdk.Duration.days(7),
        startingPosition: lambda.StartingPosition.TRIM_HORIZON
    };
    private readonly props: DataIngestionProps;
    private readonly realTimeFeedBucket: tahS3.Bucket;
    public readonly realTimeDeliveryStream: firehose.CfnDeliveryStream;
    public readonly realTimeDeliveryStreamLogGroup: cdk.aws_logs.LogGroup;
    private readonly enableRealtimeBackupCondition: cdk.CfnCondition;
    private readonly extraPyFiles: Asset;

    constructor(cdkBasPros: CdkBaseProps, props: DataIngestionProps) {
        super(cdkBasPros);

        this.props = props;

        // Setup Realtime Backup Enable/Disable condition
        this.enableRealtimeBackupCondition = new cdk.CfnCondition(this.stack, 'EnableRealTimeBackup', {
            expression: cdk.Fn.conditionEquals(props.enableRealTimeBackup, true)
        });

        // Dead letter queues
        const { dlqGo, dlqPython, dlqBatchProcessor } = this.createDeadLetterQueues();
        this.dlqGo = dlqGo;
        this.dlqPython = dlqPython;
        this.dlqBatchProcessor = dlqBatchProcessor;

        // Real-time processing
        this.realTimeFeedBucket = new tahS3.Bucket(this.stack, 'ucp-real-time-backup', this.props.accessLogBucket);
        const output = this.createFirehoseDeliveryStream({
            bucket: this.realTimeFeedBucket,
            bucketPrefix: this.props.deliveryStreamBucketPrefix,
            idPrefix: 'ucpRealtimeBackup',
            idSuffix: this.envName,
            partition: { jqString: '.domain', name: 'domainname' },
            resourceCondition: this.enableRealtimeBackupCondition
        });

        this.realTimeDeliveryStream = output.deliveryStream;
        this.realTimeDeliveryStreamLogGroup = output.logGroup;

        const { ingestorKinesisLambda, entryKinesisLambda } = this.createRealTimeProcessing();
        this.ingestorKinesisLambda = ingestorKinesisLambda;
        this.entryKinesisLambda = entryKinesisLambda;

        // Batch processing
        this.extraPyFiles = new Asset(this.stack, 'ExtraETLFiles', {
            path: path.join(__dirname, '..', '..', '..', 'tah_lib'),
            exclude: [
                '.mypy_cache',
                '.tox',
                '.venv',
                '__pycache__',
                'e2e',
                'etls',
                'test',
                '.gitignore',
                'build.sh',
                'build-local.sh',
                'deploy-local.sh',
                'update-test-data.sh',
                'tah_lib.egg-info'
            ]
        });

        this.airBookingBatch = this.createBatchProcessing('air_booking');
        this.clickStreamBatch = this.createBatchProcessing('clickstream');
        this.customerServiceInteractionBatch = this.createBatchProcessing('customer_service_interaction');
        this.guestProfileBatch = this.createBatchProcessing('guest-profile');
        this.hotelBookingBatch = this.createBatchProcessing('hotel-booking');
        this.hotelStayBatch = this.createBatchProcessing('hotel-stay');
        this.paxProfileBatch = this.createBatchProcessing('pax-profile');
        this.syncLambdaFunction = this.createSyncLambdaFunction();
        this.batchLambdaFunction = this.createBatchLambdaFunction();

        // Industry connector
        this.industryConnectorBucket = s3.Bucket.fromBucketName(
            this.stack,
            'industryConnectorBucket',
            this.props.industryConnectorBucketName
        );
        this.industryConnectorLambdaFunction = this.createIndustryConnectorLambdaFunction();
        this.industryConnectorDataSyncRole = this.createIndustryConnectorDataSyncRole();

        // Permissions
        this.airBookingBatch.bucket.grantReadWrite(this.syncLambdaFunction);
        this.clickStreamBatch.bucket.grantReadWrite(this.syncLambdaFunction);
        this.customerServiceInteractionBatch.bucket.grantReadWrite(this.syncLambdaFunction);
        this.guestProfileBatch.bucket.grantReadWrite(this.syncLambdaFunction);
        this.hotelBookingBatch.bucket.grantReadWrite(this.syncLambdaFunction);
        this.hotelStayBatch.bucket.grantReadWrite(this.syncLambdaFunction);
        this.paxProfileBatch.bucket.grantReadWrite(this.syncLambdaFunction);

        // Outputs
        tahCore.Output.add(this.stack, 'realTimeFeedBucket', this.realTimeFeedBucket.bucketName);
        tahCore.Output.add(this.stack, 'realTimeBackupEnabled', `${props.enableRealTimeBackup}`);
        tahCore.Output.add(this.stack, 'lambdaFunctionNameRealTime', this.entryKinesisLambda.lambdaFunction.functionName);
        tahCore.Output.add(this.stack, 'kinesisStreamNameRealTime', this.entryKinesisLambda.kinesisStream.streamName);
        tahCore.Output.add(this.stack, 'kinesisStreamOutputNameRealTime', this.ingestorKinesisLambda.kinesisStream.streamName);
        tahCore.Output.add(this.stack, 'dlgRealTimeGo', this.dlqGo.queueUrl);
        tahCore.Output.add(this.stack, 'dldRealTimePython', this.dlqPython.queueUrl);
    }

    /**
     * Creates dead letter queues.
     * @returns Dead letter queues
     */
    private createDeadLetterQueues(): DeadLetterQueueOutput {
        return {
            dlqGo: new tahSqs.Queue(this.stack, 'ucp-real-time-error-go-' + this.envName),
            dlqPython: new tahSqs.Queue(this.stack, 'ucp-real-time-error-python-' + this.envName),
            dlqBatchProcessor: new tahSqs.Queue(this.stack, 'ucp-batch-processor-queue-' + this.envName)
        };
    }

    /**
     * Creates real time processing resources.
     * Generally, it creates 2 KinesisToLambda Solutions Constructs resources.
     * The flow here is the following:
     * Entry Kinesis -> Python Lambda -> Second Kinesis -> Go Lambda -> Amazon Connect Customer Profiles
     * @returns Realtime processing KinesisToLambda functions
     */
    private createRealTimeProcessing(): RealTimeProcessingOutput {
        const ingestorKinesisLambda = this.createIngestorKinesisLambda();
        const entryKinesisLambda = this.createEntryKinesisLambda();

        /**
         * There are 2 cases when we write to DLQ.
         * 1. If Lambda fails for any reason, the incoming message will automatically be sent to DLQ after the configured number of retries.
         * 2. In cases where we gracefully fail inside the function, we manually put the message into the queue. This allows better error management.
         */
        ingestorKinesisLambda.kinesisStream.grantReadWrite(entryKinesisLambda.lambdaFunction);
        ingestorKinesisLambda.lambdaFunction.addEnvironment('DEAD_LETTER_QUEUE_URL', this.dlqGo.queueUrl);
        ingestorKinesisLambda.lambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    'profile:PutProfileObject',
                    'profile:ListProfileObjects',
                    'profile:ListProfileObjectTypes',
                    'profile:GetProfileObjectType',
                    'profile:PutProfileObjectType',
                    'profile:SearchProfiles',
                    'profile:MergeProfiles'
                ],
                effect: iam.Effect.ALLOW,
                resources: ['*']
            })
        );

        ingestorKinesisLambda.lambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['secretsmanager:GetSecretValue'],
                effect: iam.Effect.ALLOW,
                resources: [this.props.profileStorageOutput.storageSecretArn]
            })
        );

        entryKinesisLambda.lambdaFunction.addEnvironment('OUTPUT_STREAM', ingestorKinesisLambda.kinesisStream.streamName);
        entryKinesisLambda.lambdaFunction.addEnvironment('DEAD_LETTER_QUEUE_URL', this.dlqPython.queueUrl);
        const firehosePolicy = new iam.Policy(this.stack, 'realTimeProcessingFirehosePolicy', {
            statements: [
                new iam.PolicyStatement({
                    actions: ['firehose:PutRecord'],
                    effect: iam.Effect.ALLOW,
                    resources: [this.realTimeDeliveryStream.attrArn]
                })
            ]
        });
        const firehoseCfnPolicy = firehosePolicy.node.defaultChild as iam.CfnPolicy;
        firehoseCfnPolicy.cfnOptions.condition = this.enableRealtimeBackupCondition;
        entryKinesisLambda.lambdaFunction.role?.attachInlinePolicy(firehosePolicy);

        this.dlqPython.grantConsumeMessages(entryKinesisLambda.lambdaFunction);
        this.dlqPython.grantSendMessages(entryKinesisLambda.lambdaFunction);
        this.realTimeFeedBucket.grantReadWrite(entryKinesisLambda.lambdaFunction);

        this.dlqGo.grantConsumeMessages(ingestorKinesisLambda.lambdaFunction);
        this.dlqGo.grantSendMessages(ingestorKinesisLambda.lambdaFunction);

        return { ingestorKinesisLambda, entryKinesisLambda };
    }

    /**
     * Creates an Amazon Connect Customer Profiles (ACCP) Kinesis data stream and a Lambda function.
     * The Kinesis stream is needed to manage the real time ingestion of the ACCP data.
     * Given the hard 100TPS limit of the ACCP SDK, we set the stream in provisioned capacity mode with 10 shards, 10 batch size and a parallelization factor of 1.
     * This will allow max of 100 concurrent insertions (which we measured at about 0.9s in duration).
     * Failed insertion due to throttling by ACCP are failing the function and triggering a retry.
     * @returns KinesisSteamsToLambda Solutions Construct
     */
    private createIngestorKinesisLambda(): KinesisStreamsToLambda {
        const ucpEtlRealTimeACCPPrefix = 'ucpRealtimeTransformerAccp';
        const accpKinesisStreamProps: kinesis.StreamProps = {
            shardCount: this.props.ingestorShardCount,
            streamMode: kinesis.StreamMode.PROVISIONED,
            retentionPeriod: cdk.Duration.days(30),
            removalPolicy: cdk.RemovalPolicy.DESTROY
        };

        const streamToLambda = new KinesisStreamsToLambda(this.stack, ucpEtlRealTimeACCPPrefix + this.envName, {
            kinesisEventSourceProps: this.kinesisEventSourceProps,
            kinesisStreamProps: accpKinesisStreamProps,
            lambdaFunctionProps: {
                code: new lambda.S3Code(this.props.artifactBucket, `${this.props.artifactBucketPath}/${ucpEtlRealTimeACCPPrefix}.zip`),
                handler: 'main',
                runtime: lambda.Runtime.PROVIDED_AL2,
                architecture: lambda.Architecture.ARM_64,
                deadLetterQueue: this.dlqGo,
                deadLetterQueueEnabled: true,
                environment: {
                    LAMBDA_REGION: this.props.region ?? '',
                    // For usage metrics
                    SEND_ANONYMIZED_DATA: this.props.sendAnonymizedData,
                    METRICS_SOLUTION_ID: this.props.solutionId,
                    METRICS_SOLUTION_VERSION: this.props.solutionVersion,
                    // For low-cost storage
                    AURORA_PROXY_ENDPOINT: this.props.profileStorageOutput.storageProxyEndpoint,
                    AURORA_PROXY_READONLY_ENDPOINT: this.props.profileStorageOutput.storageProxyReadonlyEndpoint,
                    AURORA_DB_NAME: this.props.profileStorageOutput.storageDbName,
                    AURORA_DB_SECRET_ARN: this.props.profileStorageOutput.storageSecretArn,
                    STORAGE_CONFIG_TABLE_NAME: this.props.profileStorageOutput.storageConfigTable?.tableName ?? '',
                    STORAGE_CONFIG_TABLE_PK: this.props.profileStorageOutput.storageConfigTablePk,
                    STORAGE_CONFIG_TABLE_SK: this.props.profileStorageOutput.storageConfigTableSk,
                    CHANGE_PROC_KINESIS_STREAM_NAME: this.props.changeProcessorKinesisStream.streamName
                },
                functionName: ucpEtlRealTimeACCPPrefix + this.envName,
                timeout: cdk.Duration.seconds(900),
                memorySize: 128,
                // Low cost storage security config
                vpc: this.props.uptVpc,
                securityGroups: [this.props.profileStorageOutput.lambdaToProxyGroup]
            }
        });

        streamToLambda.lambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    'kinesis:PutRecord',
                    'kinesis:PutRecords',
                    'kms:Decrypt',
                    'rds-data:BatchExecuteStatement',
                    'rds-data:BeginTransaction',
                    'rds-data:CommitTransaction',
                    'rds-data:ExecuteStatement',
                    'rds-data:RollbackTransaction'
                ],
                effect: iam.Effect.ALLOW,
                resources: ['*']
            })
        );

        streamToLambda.lambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    'dynamodb:DescribeTable',
                    'dynamodb:DeleteTable',
                    'dynamodb:Query',
                    'dynamodb:BatchWriteItem',
                    'dynamodb:PutItem'
                ],
                effect: iam.Effect.ALLOW,
                resources: [`arn:${cdk.Aws.PARTITION}:dynamodb:${cdk.Aws.REGION}:${cdk.Aws.ACCOUNT_ID}:table/*`]
            })
        );

        return streamToLambda;
    }

    /**
     * Creates an entry Kinesis data stream and a Lambda function.
     * The entry stream can scale up to the requested demand using ON_DEMAND capacity provisioning.
     * Throttling will be managed in the second stream (the ACCP stream). Customers are encouraged to provide
     * random primary key to spread the load over multiple shards.
     * @returns KinesisSteamsToLambda Solutions Construct
     */
    private createEntryKinesisLambda(): KinesisStreamsToLambda {
        const ucpEtlRealTimeLambdaPrefix = 'ucpRealtimeTransformer';
        let entryStreamMode = kinesis.StreamMode.ON_DEMAND;
        let entryShardCount: number | undefined = undefined;

        if (this.props.entryStreamMode === kinesis.StreamMode.PROVISIONED.toString()) {
            entryStreamMode = kinesis.StreamMode.PROVISIONED;
            entryShardCount = this.props.entryStreamShardCount;
        }

        const entryKinesisStreamProps: kinesis.StreamProps = {
            retentionPeriod: cdk.Duration.days(30),
            shardCount: entryShardCount,
            streamMode: entryStreamMode,
            removalPolicy: cdk.RemovalPolicy.DESTROY
        };
        return new KinesisStreamsToLambda(this.stack, ucpEtlRealTimeLambdaPrefix + this.envName, {
            kinesisEventSourceProps: this.kinesisEventSourceProps,
            kinesisStreamProps: entryKinesisStreamProps,
            lambdaFunctionProps: {
                code: new lambda.S3Code(this.props.artifactBucket, `${this.props.artifactBucketPath}/${ucpEtlRealTimeLambdaPrefix}.zip`),
                handler: 'index.handler',
                runtime: lambda.Runtime.PYTHON_3_12,
                deadLetterQueueEnabled: true,
                deadLetterQueue: this.dlqPython,
                functionName: ucpEtlRealTimeLambdaPrefix + this.envName,
                environment: {
                    FIREHOSE_BACKUP_STREAM: this.realTimeDeliveryStream.deliveryStreamName ?? '',
                    LAMBDA_REGION: this.props.region ?? '',
                    ENABLE_BACKUP: `${this.props.enableRealTimeBackup}`,
                    //For usage metrics
                    SEND_ANONYMIZED_DATA: this.props.sendAnonymizedData,
                    METRICS_SOLUTION_ID: this.props.solutionId,
                    METRICS_SOLUTION_VERSION: this.props.solutionVersion
                },
                timeout: cdk.Duration.seconds(60)
            }
        });
    }

    /**
     * Creates an industry connector Lambda function.
     * @returns Industry connector Lambda function
     */
    private createIndustryConnectorLambdaFunction(): lambda.Function {
        const industryConnectorLambdaFunctionRole = new iam.Role(this.stack, 'dataTransferRole', {
            assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com'),
            managedPolicies: [iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AWSLambdaBasicExecutionRole')]
        });
        const ucpIndustryConnectorPrefix = 'ucpIndustryConnector';
        const industryConnectorLambdaFunction = new GoFunction(this.stack, ucpIndustryConnectorPrefix + this.envName, {
            entry: path.join(__dirname, '..', '..', '..', 'ucp-industry-connector-transfer', 'src', 'main.go'),
            bundling: {
                environment: { GOOS: 'linux', GOARCH: 'arm64', GOWORK: 'off' }
            },
            functionName: ucpIndustryConnectorPrefix + this.envName,
            runtime: lambda.Runtime.PROVIDED_AL2,
            architecture: lambda.Architecture.ARM_64,
            environment: {
                LAMBDA_ENV: this.envName,
                LAMBDA_REGION: cdk.Aws.REGION,
                INDUSTRY_CONNECTOR_BUCKET_NAME: this.props.industryConnectorBucketName,
                KINESIS_STREAM_NAME: this.entryKinesisLambda.kinesisStream.streamName,
                DYNAMO_TABLE: this.props.configTable.tableName,
                DYNAMO_PK: DynamoDbTableKey.CONFIG_TABLE_PK,
                DYNAMO_SK: DynamoDbTableKey.CONFIG_TABLE_SK,
                // For usage metrics
                SEND_ANONYMIZED_DATA: this.props.sendAnonymizedData,
                METRICS_SOLUTION_ID: this.props.solutionId,
                METRICS_SOLUTION_VERSION: this.props.solutionVersion
            },
            role: industryConnectorLambdaFunctionRole,
            timeout: cdk.Duration.seconds(30),
            tracing: lambda.Tracing.ACTIVE
        });

        const lambdaFunctionRole = <iam.Role>industryConnectorLambdaFunction.role;
        this.industryConnectorBucket.grantRead(lambdaFunctionRole);
        this.industryConnectorBucket.addEventNotification(
            s3.EventType.OBJECT_CREATED,
            new s3Notifications.LambdaDestination(industryConnectorLambdaFunction)
        );

        return industryConnectorLambdaFunction;
    }

    /**
     * Creates an industry connector data sync role.
     * @returns Industry connector data sync role
     */
    private createIndustryConnectorDataSyncRole(): iam.Role {
        const industryConnectorDataSyncRole = new iam.Role(this.stack, 'connectorDataSyncRole', {
            assumedBy: new iam.ServicePrincipal('datasync.amazonaws.com'),
            description: 'Role for DataSync to transfer Industry Connector data to UCP'
        });

        // DataSync role to transfer data from Connector bucket to business object buckets
        industryConnectorDataSyncRole.addToPolicy(
            new iam.PolicyStatement({
                actions: ['s3:GetBucketLocation', 's3:ListBucket', 's3:ListBucketMultipartUploads'],
                effect: iam.Effect.ALLOW,
                resources: [
                    this.industryConnectorBucket.bucketArn,
                    this.airBookingBatch.bucket.bucketArn,
                    this.clickStreamBatch.bucket.bucketArn,
                    this.customerServiceInteractionBatch.bucket.bucketArn,
                    this.guestProfileBatch.bucket.bucketArn,
                    this.hotelBookingBatch.bucket.bucketArn,
                    this.hotelStayBatch.bucket.bucketArn,
                    this.paxProfileBatch.bucket.bucketArn
                ]
            })
        );
        industryConnectorDataSyncRole.addToPolicy(
            new iam.PolicyStatement({
                actions: [
                    's3:AbortMultipartUpload',
                    's3:DeleteObject',
                    's3:GetObject',
                    's3:GetObjectTagging',
                    's3:ListMultipartUploadParts',
                    's3:PutObject',
                    's3:PutObjectTagging'
                ],
                effect: iam.Effect.ALLOW,
                resources: [
                    this.industryConnectorBucket.arnForObjects('*'),
                    this.airBookingBatch.bucket.arnForObjects('*'),
                    this.clickStreamBatch.bucket.arnForObjects('*'),
                    this.customerServiceInteractionBatch.bucket.arnForObjects('*'),
                    this.hotelBookingBatch.bucket.arnForObjects('*'),
                    this.hotelStayBatch.bucket.arnForObjects('*'),
                    this.guestProfileBatch.bucket.arnForObjects('*'),
                    this.paxProfileBatch.bucket.arnForObjects('*')
                ]
            })
        );

        return industryConnectorDataSyncRole;
    }

    /**
     * Creates a sync Lambda function which runs every hour.
     * @returns Sync Lambda function
     */
    private createSyncLambdaFunction(): lambda.Function {
        const ucpSyncLambdaPrefix = 'ucpSync';
        const syncLambdaFunction = new GoFunction(this.stack, ucpSyncLambdaPrefix + this.envName, {
            entry: path.join(__dirname, '..', '..', '..', 'ucp-sync', 'src', 'main', 'main.go'),
            bundling: {
                environment: { GOOS: 'linux', GOARCH: 'arm64', GOWORK: 'off' }
            },
            functionName: ucpSyncLambdaPrefix + this.envName,
            runtime: lambda.Runtime.PROVIDED_AL2,
            architecture: lambda.Architecture.ARM_64,
            environment: {
                LAMBDA_ACCOUNT_ID: cdk.Aws.ACCOUNT_ID,
                LAMBDA_ENV: this.envName,
                LAMBDA_REGION: cdk.Aws.REGION,
                ATHENA_DB: this.props.glueDatabase.databaseName,
                ATHENA_WORKGROUP: this.props.athenaWorkGroup.name,
                PARTITION_START_DATE: this.props.partitionStartDate,
                DYNAMO_TABLE: this.props.configTable.tableName,
                DYNAMO_PK: DynamoDbTableKey.CONFIG_TABLE_PK,
                DYNAMO_SK: DynamoDbTableKey.CONFIG_TABLE_SK,
                S3_AIR_BOOKING: this.airBookingBatch.bucket.bucketName,
                S3_CLICKSTREAM: this.clickStreamBatch.bucket.bucketName,
                S3_CSI: this.customerServiceInteractionBatch.bucket.bucketName,
                S3_GUEST_PROFILE: this.guestProfileBatch.bucket.bucketName,
                S3_HOTEL_BOOKING: this.hotelBookingBatch.bucket.bucketName,
                S3_PAX_PROFILE: this.paxProfileBatch.bucket.bucketName,
                S3_STAY_REVENUE: this.hotelStayBatch.bucket.bucketName,
                CONNECT_PROFILE_SOURCE_BUCKET: this.props.connectProfileImportBucket.bucketName,
                AIR_BOOKING_JOB_NAME_CUSTOMER: this.airBookingBatch.glueJob.jobName,
                CLICKSTREAM_JOB_NAME_CUSTOMER: this.clickStreamBatch.glueJob.jobName,
                CSI_JOB_NAME_CUSTOMER: this.customerServiceInteractionBatch.glueJob.jobName,
                GUEST_PROFILE_JOB_NAME_CUSTOMER: this.guestProfileBatch.glueJob.jobName,
                HOTEL_BOOKING_JOB_NAME_CUSTOMER: this.hotelBookingBatch.glueJob.jobName,
                HOTEL_STAY_JOB_NAME_CUSTOMER: this.hotelStayBatch.glueJob.jobName,
                PAX_PROFILE_JOB_NAME_CUSTOMER: this.paxProfileBatch.glueJob.jobName,
                AIR_BOOKING_DLQ: this.airBookingBatch.errorQueue.queueUrl,
                CLICKSTREAM_DLQ: this.clickStreamBatch.errorQueue.queueUrl,
                CSI_DLQ: this.customerServiceInteractionBatch.errorQueue.queueUrl,
                GUEST_PROFILE_DLQ: this.guestProfileBatch.errorQueue.queueUrl,
                HOTEL_BOOKING_DLQ: this.hotelBookingBatch.errorQueue.queueUrl,
                HOTEL_STAY_DLQ: this.hotelStayBatch.errorQueue.queueUrl,
                PAX_PROFILE_DLQ: this.paxProfileBatch.errorQueue.queueUrl,
                ACCP_DOMAIN_DLQ: this.props.connectProfileDomainErrorQueue.queueUrl,
                SKIP_JOB_RUN: this.props.skipJobRun,
                MATCH_BUCKET_NAME: this.props.matchBucket.bucketName,
                // For low-cost storage
                AURORA_PROXY_ENDPOINT: this.props.profileStorageOutput.storageProxyEndpoint,
                AURORA_DB_NAME: this.props.profileStorageOutput.storageDbName,
                AURORA_DB_SECRET_ARN: this.props.profileStorageOutput.storageSecretArn,
                STORAGE_CONFIG_TABLE_NAME: this.props.profileStorageOutput.storageConfigTable?.tableName ?? '',
                STORAGE_CONFIG_TABLE_PK: this.props.profileStorageOutput.storageConfigTablePk,
                STORAGE_CONFIG_TABLE_SK: this.props.profileStorageOutput.storageConfigTableSk,
                CHANGE_PROC_KINESIS_STREAM_NAME: this.props.changeProcessorKinesisStream.streamName,
                // For usage metrics
                SEND_ANONYMIZED_DATA: this.props.sendAnonymizedData,
                METRICS_SOLUTION_ID: this.props.solutionId,
                METRICS_SOLUTION_VERSION: this.props.solutionVersion
            },
            timeout: cdk.Duration.seconds(900),
            tracing: lambda.Tracing.ACTIVE,
            // Low cost storage security config
            vpc: this.props.uptVpc,
            securityGroups: [this.props.profileStorageOutput.lambdaToProxyGroup]
        });
        this.props.profileStorageOutput.storageConfigTable?.grantReadWriteData(syncLambdaFunction);
        syncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['secretsmanager:GetSecretValue'],
                effect: iam.Effect.ALLOW,
                resources: [this.props.profileStorageOutput.storageSecretArn]
            })
        );

        // Narrow this down to the business object tables
        syncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    'athena:GetQueryExecution',
                    'athena:GetQueryResults',
                    'athena:StartQueryExecution',
                    'glue:BatchCreatePartition',
                    'glue:CreatePartition',
                    'glue:GetDatabase',
                    'glue:GetTable',
                    'glue:GetPartition',
                    'glue:GetPartitions',
                    'glue:StartJobRun',
                    'profile:ListDomains'
                ],
                effect: iam.Effect.ALLOW,
                resources: ['*']
            })
        );

        const rule = new events.Rule(this.stack, 'ucpSyncJob', {
            schedule: events.Schedule.expression('rate(1 hour)'),
            eventPattern: {}
        });
        rule.addTarget(new eventsTargets.LambdaFunction(syncLambdaFunction));

        return syncLambdaFunction;
    }

    //////////////////////////////
    // BATCH PROCESSING
    /////////////////////////

    /**
     * Creates a batch processor lambda to ingest data from S3 in LCS mode.
     * @returns Sync Lambda function
     */
    private createBatchLambdaFunction(): lambda.Function {
        const ucpBatchLambdaPrefix = 'ucpBatch';
        const ucpBatchLambdaFunction = new GoFunction(this.stack, ucpBatchLambdaPrefix + this.envName, {
            entry: path.join(__dirname, '..', '..', '..', 'ucp-batch', 'src', 'main', 'main.go'),
            bundling: {
                environment: { GOOS: 'linux', GOARCH: 'arm64', GOWORK: 'off' }
            },
            functionName: ucpBatchLambdaPrefix + this.envName,
            runtime: lambda.Runtime.PROVIDED_AL2,
            architecture: lambda.Architecture.ARM_64,
            environment: {
                LAMBDA_ACCOUNT_ID: cdk.Aws.ACCOUNT_ID,
                LAMBDA_ENV: this.envName,
                LAMBDA_REGION: cdk.Aws.REGION,
                DYNAMO_TABLE: this.props.configTable.tableName,
                DYNAMO_PK: DynamoDbTableKey.CONFIG_TABLE_PK,
                DYNAMO_SK: DynamoDbTableKey.CONFIG_TABLE_SK,
                CONNECT_PROFILE_SOURCE_BUCKET: this.props.connectProfileImportBucket.bucketName,
                BATCH_PROCESSOR_DLQ: this.dlqBatchProcessor.queueName, // name of the error queue to send
                KINESIS_INGESTOR_STREAM_NAME: this.ingestorKinesisLambda.kinesisStream.streamName,
                // For usage metrics
                SEND_ANONYMIZED_DATA: this.props.sendAnonymizedData,
                METRICS_SOLUTION_ID: this.props.solutionId,
                METRICS_SOLUTION_VERSION: this.props.solutionVersion
            },
            timeout: cdk.Duration.seconds(900),
            memorySize: 128,
            tracing: lambda.Tracing.ACTIVE,
            // Low cost storage security config
            vpc: this.props.uptVpc,
            securityGroups: [this.props.profileStorageOutput.lambdaToProxyGroup]
        });

        this.dlqBatchProcessor.grantSendMessages(ucpBatchLambdaFunction);

        this.props.connectProfileImportBucket.grantRead(ucpBatchLambdaFunction);

        this.props.connectProfileImportBucket.addEventNotification(
            s3.EventType.OBJECT_CREATED,
            new s3Notifications.LambdaDestination(ucpBatchLambdaFunction)
        );

        this.ingestorKinesisLambda.kinesisStream.grantWrite(ucpBatchLambdaFunction);
        return ucpBatchLambdaFunction;
    }

    /**
     * Creates batch processing resources for a business use case.
     * 1. A bucket to get business objects
     * 2. An error queue to get errors
     * 3. A Glue job to process the bucket objects
     * @returns S3 bucket, Glue job, and error SQS queue
     */
    private createBatchProcessing(businessObjectName: string): BatchProcessingOutput {
        const bucket = new tahS3.Bucket(this.stack, 'ucp' + businessObjectName, this.props.accessLogBucket);
        bucket.grantReadWrite(this.props.dataLakeAdminRole);

        const errorQueue = new tahSqs.Queue(this.stack, businessObjectName + '-errors-' + this.envName);
        errorQueue.grantSendMessages(this.props.dataLakeAdminRole);

        const glueJob = this.createGlueJob(businessObjectName, errorQueue);

        // Outputs
        tahCore.Output.add(this.stack, 'customerBucket' + businessObjectName, bucket.bucketName);
        tahCore.Output.add(this.stack, 'customerJobName' + businessObjectName, glueJob.jobName ?? '');

        return {
            bucket,
            errorQueue,
            glueJob
        };
    }

    /**
     * Creates a Glue job for a business use case.
     * @param businessObjectName Business object name
     * @param errorQueue Error SQS queue
     * @returns Glue job
     */
    private createGlueJob(businessObjectName: string, errorQueue: tahSqs.Queue): glueAlpha.Job {
        const underScoreBusinessObjectName = businessObjectName.replace(/-/g, '_');
        const scriptName = underScoreBusinessObjectName + 'ToUcp';
        const args: Record<string, string> = {
            'enable-continuous-cloudwatch-log': 'true',
            'enable-metrics': 'true',
            'extra-py-files': this.extraPyFiles.s3ObjectUrl,
            'job-bookmark-option': 'job-bookmark-enable',
            BIZ_OBJECT_NAME: underScoreBusinessObjectName,
            DEST_BUCKET: this.props.connectProfileImportBucket.bucketName,
            DYNAMO_TABLE: this.props.configTable.tableName,
            ERROR_QUEUE_URL: errorQueue.queueUrl,
            GLUE_DB: this.props.glueDatabase.databaseName,
            // Usage metrics
            SEND_ANONYMIZED_DATA: this.props.sendAnonymizedData,
            METRICS_SOLUTION_ID: this.props.solutionId,
            METRICS_SOLUTION_VERSION: this.props.solutionVersion,
            // SOURCE_TABLE and METRICS_UUID are set by the Sync Lambda when job is triggered
            SOURCE_TABLE: '',
            METRICS_UUID: ''
        };
        const defaultArguments: Record<string, string> = {};

        for (const [key, value] of Object.entries(args)) {
            defaultArguments[`--${key}`] = value;
        }

        const logicalId = businessObjectName + 'Job' + this.envName;
        const job = new glueAlpha.Job(this.stack, logicalId, {
            executable: glueAlpha.JobExecutable.pythonEtl({
                glueVersion: glueAlpha.GlueVersion.V4_0,
                pythonVersion: glueAlpha.PythonVersion.THREE,
                script: glueAlpha.Code.fromAsset(path.join(__dirname, '..', '..', '..', 'ucp-etls', 'etls', `${scriptName}.py`))
            }),
            defaultArguments,
            maxConcurrentRuns: 1000,
            maxRetries: 0,
            jobName: businessObjectName + 'Job' + this.envName,
            role: this.props.dataLakeAdminRole
        });
        (job.node.defaultChild as CfnJob).overrideLogicalId(logicalId.replace(/[_-]/g, ''));

        return job;
    }
}
