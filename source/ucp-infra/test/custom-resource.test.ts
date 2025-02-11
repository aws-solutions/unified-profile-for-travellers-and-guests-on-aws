// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Match, Template } from 'aws-cdk-lib/assertions';
import { Runtime } from 'aws-cdk-lib/aws-lambda';
import { CustomActions, CustomResource } from '../lib/custom-resource';
import { artifactBucketPath, createS3Bucket, envName, sendAnonymizedData, solutionId, solutionName, solutionVersion } from './mock';

test('CustomResource', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const cognitoClientId = 'client-id';
    const cognitoUserPoolId = 'userpool';
    const deployed = new Date().toISOString();
    const ucpApiUrl = 'https://example.com';
    const staticContentBucketName = 'bucket';
    const artifactBucket = createS3Bucket(stack, 'artifactBucket');
    const usePermissionSystem = 'true';

    // Call
    new CustomResource(
        { envName, stack },
        {
            frontEndConfig: { cognitoClientId, cognitoUserPoolId, deployed, ucpApiUrl, staticContentBucketName, usePermissionSystem },
            solutionName,
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
        Runtime: Runtime.NODEJS_20_X.toString(),
        Description: `${solutionName} (${solutionVersion}): Custom resource`
    });
    template.hasResourceProperties('AWS::CloudFormation::CustomResource', {
        CustomAction: CustomActions.CREATE_UUID
    });
    template.hasResourceProperties('AWS::CloudFormation::CustomResource', {
        CustomAction: CustomActions.DEPLOY_FRONT_END,
        FrontEndConfig: {
            cognitoClientId,
            cognitoUserPoolId,
            deployed,
            ucpApiUrl,
            staticContentBucketName,
            artifactBucketName: Match.anyValue(),
            artifactBucketPath: artifactBucketPath + '/ui/dist.zip',
            usePermissionSystem
        }
    });
});
