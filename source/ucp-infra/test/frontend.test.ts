// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Match, Template } from 'aws-cdk-lib/assertions';
import { Frontend } from '../lib/frontend';
import { createCfnCondition, createS3Bucket, createVpc, envName } from './mock';

test('Frontend', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const accessLogBucket = createS3Bucket(stack, 'accessLogBucket');
    const deployFrontendToCloudFrontCondition = createCfnCondition(stack, 'deployFeToCfCondition');
    const deployFrontendToEcsCondition = createCfnCondition(stack, 'deployFeToEcsCondition');
    const uptVpc = createVpc(stack, 'vpc');

    // Call
    new Frontend(
        { envName, stack },
        {
            accessLogBucket,
            deployFrontendToCloudFrontCondition,
            deployFrontendToEcsCondition,
            ucpApiUrl: '',
            uptVpc: uptVpc,
            ecrProps: { publishEcrAssets: true },
            cognitoClientId: '',
            cognitoUserPoolId: '',
            usePermissionSystem: ''
        }
    );

    // Verify
    const template = Template.fromStack(stack);
    template.resourceCountIs('AWS::CloudFront::CloudFrontOriginAccessIdentity', 1);
    template.resourceCountIs('AWS::CloudFront::Distribution', 1);
    template.resourceCountIs('AWS::CloudFront::ResponseHeadersPolicy', 1);
    template.hasResourceProperties('AWS::S3::Bucket', {
        LoggingConfiguration: Match.objectLike({
            LogFilePrefix: 'bucket-ucp-connector-fe-' + envName
        })
    });
});
