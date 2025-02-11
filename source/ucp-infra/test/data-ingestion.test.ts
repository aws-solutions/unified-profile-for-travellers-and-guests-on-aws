// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { DataIngestion } from '../lib/data-ingestion';
import { Bucket } from '../tah-cdk-common/s3';
import {
    artifactBucketPath,
    createAthenaWorkGroup,
    createDynamoDbTable,
    createGlueDatabase,
    createIamRole,
    createKinesisStream,
    createProfileStorageOutput,
    createS3Bucket,
    createSqsQueue,
    createVpc,
    envName,
    sendAnonymizedData,
    solutionId,
    solutionVersion
} from './mock';

test('DataIngestion', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const accessLogBucket = createS3Bucket(stack, 'accessLogBucket');
    const athenaWorkGroup = createAthenaWorkGroup(stack, 'athenaWorkGroup');
    const configTable = createDynamoDbTable(stack, 'configTable', false);
    const connectProfileDomainErrorQueue = createSqsQueue(stack, 'connectProfileDomainErrorQueue');
    const connectProfileImportBucket = <Bucket>createS3Bucket(stack, 'connectProfileImportBucket');
    const dataLakeAdminRole = createIamRole(stack, 'dataLakeAdminRole');
    const deliveryStreamBucketPrefix = 'prefix';
    const entryStreamMode = 'ON_DEMAND';
    const entryStreamShardCount = 10;
    const glueDatabase = createGlueDatabase(stack, 'glueDatabase');
    const industryConnectorBucketName = 'industry-connector-bucket';
    const ingestorShardCount = 20;
    const matchBucket = createS3Bucket(stack, 'matchBucket');
    const partitionStartDate = '2024/02/01';
    const region = 'us-west-2';
    const skipJobRun = 'true';
    const artifactBucket = createS3Bucket(stack, 'artifactBucket');
    const enableRealTimeBackup = false;
    const vpc = createVpc(stack, 'vpc');
    const profileStorageOutput = createProfileStorageOutput(stack, 'storageOutput', vpc);
    const changeProcessorKinesisStream = createKinesisStream(stack, 'changeProcessorKinesisStream');

    // Call
    new DataIngestion(
        { envName, stack },
        {
            accessLogBucket,
            athenaWorkGroup,
            configTable,
            connectProfileDomainErrorQueue,
            connectProfileImportBucket,
            dataLakeAdminRole,
            deliveryStreamBucketPrefix,
            entryStreamMode,
            entryStreamShardCount,
            glueDatabase,
            industryConnectorBucketName,
            ingestorShardCount,
            matchBucket,
            partitionStartDate,
            region,
            skipJobRun,
            profileStorageOutput,
            changeProcessorKinesisStream,
            // LambdaProps
            sendAnonymizedData,
            artifactBucket,
            artifactBucketPath,
            solutionId,
            solutionVersion,
            enableRealTimeBackup,
            uptVpc: vpc
        }
    );

    // Verify
    const template = Template.fromStack(stack);
    template.hasCondition('EnableRealTimeBackup', {
        'Fn::Equals': [false, true]
    });
    template.resourceCountIs('AWS::KinesisFirehose::DeliveryStream', 1);
    template.hasResourceProperties('AWS::Kinesis::Stream', {
        StreamModeDetails: {
            StreamMode: entryStreamMode
        }
    });
    template.hasResourceProperties('AWS::Kinesis::Stream', {
        ShardCount: ingestorShardCount,
        StreamModeDetails: {
            StreamMode: 'PROVISIONED'
        }
    });

    for (const lambdaFunctionNamePrefix of ['ucpRealtimeTransformerAccp', 'ucpRealtimeTransformer', 'ucpIndustryConnector', 'ucpSync']) {
        template.hasResourceProperties('AWS::Lambda::Function', {
            FunctionName: lambdaFunctionNamePrefix + envName
        });
    }

    for (const glueJobNamePrefix of [
        'air_booking',
        'clickstream',
        'customer_service_interaction',
        'guest-profile',
        'hotel-booking',
        'hotel-stay',
        'pax-profile'
    ]) {
        template.hasResourceProperties('AWS::Glue::Job', {
            Name: glueJobNamePrefix + 'Job' + envName
        });
    }
});

test('DataIngestion when entry stream mode is PROVISIONED', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const accessLogBucket = createS3Bucket(stack, 'accessLogBucket');
    const athenaWorkGroup = createAthenaWorkGroup(stack, 'athenaWorkGroup');
    const configTable = createDynamoDbTable(stack, 'configTable', false);
    const connectProfileDomainErrorQueue = createSqsQueue(stack, 'connectProfileDomainErrorQueue');
    const connectProfileImportBucket = <Bucket>createS3Bucket(stack, 'connectProfileImportBucket');
    const dataLakeAdminRole = createIamRole(stack, 'dataLakeAdminRole');
    const deliveryStreamBucketPrefix = 'prefix';
    const entryStreamMode = 'PROVISIONED';
    const entryStreamShardCount = 10;
    const glueDatabase = createGlueDatabase(stack, 'glueDatabase');
    const industryConnectorBucketName = 'industry-connector-bucket';
    const ingestorShardCount = 20;
    const matchBucket = createS3Bucket(stack, 'matchBucket');
    const partitionStartDate = '2024/02/01';
    const region = 'us-west-2';
    const skipJobRun = 'true';
    const artifactBucket = createS3Bucket(stack, 'artifactBucket');
    const enableRealTimeBackup = false;
    const vpc = createVpc(stack, 'vpc');
    const profileStorageOutput = createProfileStorageOutput(stack, 'storageOutput', vpc);
    const changeProcessorKinesisStream = createKinesisStream(stack, 'changeProcessorKinesisStream');

    // Call
    new DataIngestion(
        { envName, stack },
        {
            accessLogBucket,
            athenaWorkGroup,
            configTable,
            connectProfileDomainErrorQueue,
            connectProfileImportBucket,
            dataLakeAdminRole,
            deliveryStreamBucketPrefix,
            entryStreamMode,
            entryStreamShardCount,
            glueDatabase,
            industryConnectorBucketName,
            ingestorShardCount,
            matchBucket,
            partitionStartDate,
            region,
            skipJobRun,
            profileStorageOutput,
            changeProcessorKinesisStream,
            // LambdaProps
            sendAnonymizedData,
            artifactBucket,
            artifactBucketPath,
            solutionId,
            solutionVersion,
            enableRealTimeBackup,
            uptVpc: vpc
        }
    );

    // Verify
    const template = Template.fromStack(stack);
    template.hasCondition('EnableRealTimeBackup', {
        'Fn::Equals': [false, true]
    });
    template.hasResourceProperties('AWS::Kinesis::Stream', {
        ShardCount: entryStreamShardCount,
        StreamModeDetails: {
            StreamMode: entryStreamMode
        }
    });
    template.hasResourceProperties('AWS::Kinesis::Stream', {
        ShardCount: ingestorShardCount,
        StreamModeDetails: {
            StreamMode: 'PROVISIONED'
        }
    });
});

test('DataIngestion Realtime backup enabled', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const accessLogBucket = createS3Bucket(stack, 'accessLogBucket');
    const athenaWorkGroup = createAthenaWorkGroup(stack, 'athenaWorkGroup');
    const configTable = createDynamoDbTable(stack, 'configTable', false);
    const connectProfileDomainErrorQueue = createSqsQueue(stack, 'connectProfileDomainErrorQueue');
    const connectProfileImportBucket = <Bucket>createS3Bucket(stack, 'connectProfileImportBucket');
    const dataLakeAdminRole = createIamRole(stack, 'dataLakeAdminRole');
    const deliveryStreamBucketPrefix = 'prefix';
    const entryStreamMode = 'ON_DEMAND';
    const entryStreamShardCount = 10;
    const glueDatabase = createGlueDatabase(stack, 'glueDatabase');
    const industryConnectorBucketName = 'industry-connector-bucket';
    const ingestorShardCount = 20;
    const matchBucket = createS3Bucket(stack, 'matchBucket');
    const partitionStartDate = '2024/02/01';
    const region = 'us-west-2';
    const skipJobRun = 'true';
    const artifactBucket = createS3Bucket(stack, 'artifactBucket');
    const enableRealTimeBackup = true;
    const vpc = createVpc(stack, 'vpc');
    const profileStorageOutput = createProfileStorageOutput(stack, 'storageOutput', vpc);
    const changeProcessorKinesisStream = createKinesisStream(stack, 'changeProcessorKinesisStream');

    // Call
    new DataIngestion(
        { envName, stack },
        {
            accessLogBucket,
            athenaWorkGroup,
            configTable,
            connectProfileDomainErrorQueue,
            connectProfileImportBucket,
            dataLakeAdminRole,
            deliveryStreamBucketPrefix,
            entryStreamMode,
            entryStreamShardCount,
            glueDatabase,
            industryConnectorBucketName,
            ingestorShardCount,
            matchBucket,
            partitionStartDate,
            region,
            skipJobRun,
            profileStorageOutput,
            changeProcessorKinesisStream,
            // LambdaProps
            sendAnonymizedData,
            artifactBucket,
            artifactBucketPath,
            solutionId,
            solutionVersion,
            enableRealTimeBackup,
            uptVpc: vpc
        }
    );

    // Verify
    const template = Template.fromStack(stack);
    template.resourceCountIs('AWS::KinesisFirehose::DeliveryStream', 1);
    template.hasCondition('EnableRealTimeBackup', {
        'Fn::Equals': [true, true]
    });
    template.hasResourceProperties('AWS::Kinesis::Stream', {
        StreamModeDetails: {
            StreamMode: entryStreamMode
        }
    });
    template.hasResourceProperties('AWS::Kinesis::Stream', {
        ShardCount: ingestorShardCount,
        StreamModeDetails: {
            StreamMode: 'PROVISIONED'
        }
    });

    for (const lambdaFunctionNamePrefix of ['ucpRealtimeTransformerAccp', 'ucpRealtimeTransformer', 'ucpIndustryConnector', 'ucpSync']) {
        template.hasResourceProperties('AWS::Lambda::Function', {
            FunctionName: lambdaFunctionNamePrefix + envName
        });
    }

    for (const glueJobNamePrefix of [
        'air_booking',
        'clickstream',
        'customer_service_interaction',
        'guest-profile',
        'hotel-booking',
        'hotel-stay',
        'pax-profile'
    ]) {
        template.hasResourceProperties('AWS::Glue::Job', {
            Name: glueJobNamePrefix + 'Job' + envName
        });
    }
});
