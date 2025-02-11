// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Match, Template } from 'aws-cdk-lib/assertions';
import { DynamoDbTableKey } from '../lib/cdk-base';
import { ErrorManagement } from '../lib/error-management';
import {
    artifactBucketPath,
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

test('ErrorManagement', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const dynamoDbKey = createKmsKey(stack, 'dynamoDbKey');
    const errorTtl = '10';
    const artifactBucket = createS3Bucket(stack, 'artifactBucket');
    const vpc = createVpc(stack, 'vpc');
    const profileStorageOutput = createProfileStorageOutput(stack, 'errorMgmStorageOutput', vpc);
    const changeProcessorKinesisStream = createKinesisStream(stack, 'kinesisStream');

    // Call
    new ErrorManagement(
        { envName, stack },
        {
            dynamoDbKey,
            errorTtl,
            profileStorageOutput,
            changeProcessorKinesisStream,
            sendAnonymizedData,
            artifactBucket,
            artifactBucketPath,
            solutionId,
            solutionVersion
        }
    );

    // Verify
    const template = Template.fromStack(stack);
    template.hasResourceProperties('AWS::Lambda::Function', {
        FunctionName: 'ucpError' + envName,
        Environment: {
            Variables: Match.objectLike({
                DYNAMO_PK: DynamoDbTableKey.ERROR_TABLE_PK,
                DYNAMO_SK: DynamoDbTableKey.ERROR_TABLE_SK,
                TTL: errorTtl
            })
        }
    });
    template.hasResourceProperties('AWS::Lambda::Function', {
        FunctionName: 'ucpRetry' + envName,
        Environment: {
            Variables: Match.objectLike({
                ERROR_PK: DynamoDbTableKey.ERROR_TABLE_PK,
                ERROR_SK: DynamoDbTableKey.ERROR_TABLE_SK,
                RETRY_PK: DynamoDbTableKey.ERROR_RETRY_TABLE_PK,
                RETRY_SK: DynamoDbTableKey.ERROR_RETRY_TABLE_SK
            })
        }
    });
    template.hasResourceProperties('AWS::DynamoDB::Table', {
        KeySchema: [
            { AttributeName: DynamoDbTableKey.ERROR_TABLE_PK, KeyType: 'HASH' },
            { AttributeName: DynamoDbTableKey.ERROR_TABLE_SK, KeyType: 'RANGE' }
        ]
    });
    template.hasResourceProperties('AWS::DynamoDB::Table', {
        KeySchema: [
            { AttributeName: DynamoDbTableKey.ERROR_RETRY_TABLE_PK, KeyType: 'HASH' },
            { AttributeName: DynamoDbTableKey.ERROR_RETRY_TABLE_SK, KeyType: 'RANGE' }
        ]
    });
});
