// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Match, Template } from 'aws-cdk-lib/assertions';
import { ChangeProcessor } from '../lib/change-processor';
import {
    artifactBucketPath,
    createGlueDatabase,
    createGlueTable,
    createProfileStorageOutput,
    createS3Bucket,
    createVpc,
    envName,
    sendAnonymizedData,
    solutionId,
    solutionVersion
} from './mock';

test('ChangeProcessor', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const bucket = createS3Bucket(stack, 'Bucket');
    const bucketPrefix = 'profiles';
    const eventBridgeEnabled = 'true';
    const exportStreamMode = 'ON_DEMAND';
    const exportStreamShardCount = 10;
    const glueDatabase = createGlueDatabase(stack, 'glueDatabase');
    const glueTravelerTable = createGlueTable(stack, 'glueTravelerTable', glueDatabase);
    const artifactBucket = createS3Bucket(stack, 'artifactBucket');
    const vpc = createVpc(stack, 'vpc');
    const profileStorageOutput = createProfileStorageOutput(stack, 'changeProcessorStorageOutput', vpc);

    // Call
    new ChangeProcessor(
        { envName, stack },
        {
            bucket,
            bucketPrefix,
            eventBridgeEnabled,
            exportStreamMode,
            exportStreamShardCount,
            glueDatabase,
            glueTravelerTable,
            profileStorageOutput,
            sendAnonymizedData,
            artifactBucket,
            artifactBucketPath,
            solutionId,
            solutionVersion,
            uptVpc: vpc
        }
    );

    // Verify
    const template = Template.fromStack(stack);
    template.resourceCountIs('AWS::SQS::Queue', 2);
    template.hasResourceProperties('AWS::KinesisFirehose::DeliveryStream', {
        ExtendedS3DestinationConfiguration: {
            DataFormatConversionConfiguration: {
                Enabled: true,
                InputFormatConfiguration: {
                    Deserializer: {
                        HiveJsonSerDe: {}
                    }
                },
                OutputFormatConfiguration: {
                    Serializer: {
                        ParquetSerDe: {
                            Compression: 'GZIP'
                        }
                    }
                },
                SchemaConfiguration: {
                    DatabaseName: { Ref: Match.stringLikeRegexp('glueDatabase') },
                    Region: { Ref: 'AWS::Region' },
                    RoleARN: { 'Fn::GetAtt': [Match.stringLikeRegexp('firehoseFormatConversionRole'), 'Arn'] },
                    TableName: { Ref: Match.stringLikeRegexp('glueTravelerTable') },
                    VersionId: 'LATEST'
                }
            }
        }
    });
    template.hasResourceProperties('AWS::Events::EventBus', {
        Name: 'ucp-traveller-changes-' + envName
    });
    template.hasResourceProperties('AWS::Lambda::Function', {
        FunctionName: 'ucpChangeProcessor' + envName
    });
    template.hasResourceProperties('AWS::Kinesis::Stream', {
        StreamModeDetails: {
            StreamMode: exportStreamMode
        }
    });
});

test('ChangeProcessor when export stream mode is PROVISIONED', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const bucket = createS3Bucket(stack, 'Bucket');
    const bucketPrefix = 'profiles';
    const eventBridgeEnabled = 'true';
    const exportStreamMode = 'PROVISIONED';
    const exportStreamShardCount = 10;
    const glueDatabase = createGlueDatabase(stack, 'glueDatabase');
    const glueTravelerTable = createGlueTable(stack, 'glueTravelerTable', glueDatabase);
    const artifactBucket = createS3Bucket(stack, 'artifactBucket');
    const vpc = createVpc(stack, 'vpc');
    const profileStorageOutput = createProfileStorageOutput(stack, 'changeProcessorProvisionedStorageOutput', vpc);

    // Call
    new ChangeProcessor(
        { envName, stack },
        {
            bucket,
            bucketPrefix,
            eventBridgeEnabled,
            exportStreamMode,
            exportStreamShardCount,
            glueDatabase,
            glueTravelerTable,
            profileStorageOutput,
            sendAnonymizedData,
            artifactBucket,
            artifactBucketPath,
            solutionId,
            solutionVersion,
            uptVpc: vpc
        }
    );

    // Verify
    const template = Template.fromStack(stack);
    template.hasResourceProperties('AWS::Kinesis::Stream', {
        ShardCount: exportStreamShardCount,
        StreamModeDetails: {
            StreamMode: exportStreamMode
        }
    });
});
