#!/usr/bin/env node
// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, DefaultStackSynthesizer, IStackSynthesizer, Stack, Tags } from 'aws-cdk-lib';
import { UptEcrRepository } from '../ecr-repo-stack';
import { EcrProps, UCPInfraStack } from '../lib/ucp-infra-stack';
import { Output } from '../tah-cdk-common/core';

const app = new App();

const envName = app.node.tryGetContext('envName');
const solutionId = app.node.tryGetContext('solutionId');
const solutionName = app.node.tryGetContext('solutionName');

const solutionVersion = '2.1.1';

const artifactBucketPath = app.node.tryGetContext('artifactBucketPath');
const artifactReferenceBucket = app.node.tryGetContext('artifactBucket');

if (!envName) {
    throw new Error('No environment name provided for stack');
}

if (!artifactReferenceBucket) {
    throw new Error('No bucket name provided for stack');
}

const artifactBucketName = `\${artifactBucketPrefix}-\${AWS::Region}`;

const { PUBLIC_ECR_REGISTRY, PUBLIC_ECR_TAG } = process.env;
const solutionsPipeline = Boolean(PUBLIC_ECR_REGISTRY && PUBLIC_ECR_TAG);

let synthesizer: IStackSynthesizer | undefined;
let ecrProps: EcrProps = { publishEcrAssets: true };
if (solutionsPipeline) {
    // set target bucket and ECR registry/repository for solutions pipeline
    synthesizer = new DefaultStackSynthesizer({
        generateBootstrapVersionRule: false,
        fileAssetsBucketName: artifactBucketName,
        bucketPrefix: `${artifactBucketPath}/`
    });
    // the solutions pipeline publishes ECR assets in the public repository
    if (!PUBLIC_ECR_REGISTRY || !PUBLIC_ECR_TAG) {
        throw new Error('Must provide ECR registry and tag');
    }
    ecrProps = {
        publishEcrAssets: false,
        publicEcrRegistry: PUBLIC_ECR_REGISTRY,
        publicEcrTag: PUBLIC_ECR_TAG
    };
} else {
    // set target bucket and ECR registry/repository for all other situations
    const imageAssetsRepositoryName = app.node.tryGetContext('ecrRepository');
    if (!imageAssetsRepositoryName) {
        throw new Error('No ECR repository');
    }

    const imageAssetPublishingRoleArn = app.node.tryGetContext('ecrPublishRoleArn');
    if (!imageAssetPublishingRoleArn) {
        throw new Error('No image asset publish role');
    }

    synthesizer = new DefaultStackSynthesizer({
        generateBootstrapVersionRule: false,
        fileAssetsBucketName: artifactBucketName,
        bucketPrefix: `${artifactBucketPath}/`,
        imageAssetsRepositoryName,
        imageAssetPublishingRoleArn
    });
}

// Standard UPT stack
const stack = new UCPInfraStack(app, 'UCPInfraStack' + envName, {
    synthesizer,
    description: `(${solutionId}) ${solutionName} v${solutionVersion}`,
    solutionId,
    solutionName,
    solutionVersion,
    ecrProps,
    isByoVpcTemplate: false
});
Tags.of(stack).add('application-name', 'ucp');
Tags.of(stack).add('application-env', envName);

// Customer provided (bring your own) UPT stack
const byoVpcStack = new UCPInfraStack(app, 'UCPInfraStackByoVpc' + envName, {
    synthesizer,
    description: `(${solutionId}) ${solutionName} v${solutionVersion}`,
    solutionId,
    solutionName,
    solutionVersion,
    ecrProps,
    isByoVpcTemplate: true
});
Tags.of(byoVpcStack).add('application-name', 'ucp');
Tags.of(byoVpcStack).add('application-env', envName);

const ecrRepoStack = new Stack(app, 'UptInfra', {
    description: 'Stack with resources to allow publishing ECR images for UPT.'
});
const ercRepo = new UptEcrRepository(ecrRepoStack, 'Ecr');
Output.add(ecrRepoStack, 'ecrRepoName', ercRepo.repository.repositoryName);
