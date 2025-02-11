// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Match, Template } from 'aws-cdk-lib/assertions';
import { DynamoDbTableKey } from '../lib/cdk-base';
import { MatchProcessor } from '../lib/match-processor';
import {
    artifactBucketPath,
    createDynamoDbTable,
    createKinesisStream,
    createKmsKey,
    createProfileStorageOutput,
    createS3Bucket,
    createVpc,
    envName,
    sendAnonymizedData,
    solutionId,
    solutionVersion
} from './mock';

test('MatchProcessor', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const accessLogBucket = createS3Bucket(stack, 'accessLogBucket');
    const dynamoDbKey = createKmsKey(stack, 'dynamoDbKey');
    const artifactBucket = createS3Bucket(stack, 'artifactBucket');
    const vpc = createVpc(stack, 'vpc');
    const profileStorageOutput = createProfileStorageOutput(stack, 'profileStorageOutput', vpc);
    const changeProcessorKinesisStream = createKinesisStream(stack, 'changeProcessorKinesisStream');
    const portalConfigTable = createDynamoDbTable(stack, 'portalConfigTable', false);

    // Call
    new MatchProcessor(
        { envName, stack },
        {
            accessLogBucket,
            dynamoDbKey,
            sendAnonymizedData,
            artifactBucket,
            artifactBucketPath,
            profileStorageOutput,
            changeProcessorKinesisStream,
            solutionId,
            solutionVersion,
            portalConfigTable,
            uptVpc: vpc
        }
    );

    // Verify
    const template = Template.fromStack(stack);
    template.hasResourceProperties('AWS::S3::Bucket', {
        LoggingConfiguration: Match.objectLike({
            LogFilePrefix: 'bucket-ucpMatch' + envName + 'Bucket'
        })
    });
    template.hasResourceProperties('AWS::Lambda::Function', {
        FunctionName: 'ucpMatch' + envName
    });
    template.hasResourceProperties('AWS::DynamoDB::Table', {
        KeySchema: [
            { AttributeName: DynamoDbTableKey.MATCH_TABLE_PK, KeyType: 'HASH' },
            { AttributeName: DynamoDbTableKey.MATCH_TABLE_SK, KeyType: 'RANGE' }
        ],
        GlobalSecondaryIndexes: [
            Match.objectLike({
                KeySchema: [
                    { AttributeName: DynamoDbTableKey.MATCH_GSI_PK, KeyType: 'HASH' },
                    { AttributeName: DynamoDbTableKey.MATCH_GSI_SK, KeyType: 'RANGE' }
                ]
            })
        ]
    });
});
