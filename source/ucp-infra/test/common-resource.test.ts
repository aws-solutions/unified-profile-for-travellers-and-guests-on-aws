// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { DynamoDbTableKey } from '../lib/cdk-base';
import { CommonResource } from '../lib/common-resource';
import { envName } from './mock';

test('CommonResource', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const outputBucketPrefix = 'profiles';

    // Call
    new CommonResource({ envName, stack }, { outputBucketPrefix });

    // Verify
    const template = Template.fromStack(stack);
    template.resourceCountIs('AWS::S3::Bucket', 5);
    template.resourceCountIs('AWS::SQS::Queue', 1);
    template.resourceCountIs('AWS::Logs::LogGroup', 2);
    template.resourceCountIs('AWS::KMS::Key', 1);
    template.resourceCountIs('AWS::Glue::Database', 1);
    template.hasResourceProperties('AWS::Athena::WorkGroup', {
        Name: 'ucp-athena-workgroup-' + envName
    });
    template.hasResourceProperties('AWS::Glue::Database', {
        DatabaseInput: {
            Name: 'ucp_db_' + envName
        }
    });
    template.hasResourceProperties('AWS::Glue::Table', {
        TableInput: {
            Name: 'ucp_traveler_' + envName
        }
    });
    template.hasResourceProperties('AWS::IAM::Role', {
        AssumeRolePolicyDocument: {
            Statement: [
                {
                    Action: 'sts:AssumeRole',
                    Effect: 'Allow',
                    Principal: {
                        Service: 'glue.amazonaws.com'
                    }
                }
            ],
            Version: '2012-10-17'
        }
    });
    template.hasResourceProperties('AWS::DynamoDB::Table', {
        KeySchema: [
            { AttributeName: DynamoDbTableKey.CONFIG_TABLE_PK, KeyType: 'HASH' },
            { AttributeName: DynamoDbTableKey.CONFIG_TABLE_SK, KeyType: 'RANGE' }
        ]
    });
    template.hasResourceProperties('AWS::DynamoDB::Table', {
        KeySchema: [
            { AttributeName: DynamoDbTableKey.PORTAL_CONFIG_TABLE_PK, KeyType: 'HASH' },
            { AttributeName: DynamoDbTableKey.PORTAL_CONFIG_TABLE_SK, KeyType: 'RANGE' }
        ]
    });
});
