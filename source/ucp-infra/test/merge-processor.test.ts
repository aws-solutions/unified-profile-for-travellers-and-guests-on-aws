// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { MergeProcessor } from '../lib/merge-processor';
import {
    artifactBucketPath,
    createKinesisStream,
    createProfileStorageOutput,
    createS3Bucket,
    createVpc,
    envName,
    sendAnonymizedData,
    solutionId,
    solutionVersion
} from './mock';

test('MergeProcessor', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const vpc = createVpc(stack, 'vpc');
    const profileStorageOutput = createProfileStorageOutput(stack, 'profileStorageOutput', vpc);
    const changeProcessorKinesisStream = createKinesisStream(stack, 'kinesisStream');
    const artifactBucket = createS3Bucket(stack, 'artifactBucket');

    // Call
    new MergeProcessor(
        { envName, stack },
        {
            profileStorageOutput,
            changeProcessorKinesisStream,
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
    template.hasResourceProperties('AWS::Lambda::Function', {
        FunctionName: 'ucpMerger' + envName
    });
});
