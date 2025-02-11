// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App } from 'aws-cdk-lib';
import { Match, Template } from 'aws-cdk-lib/assertions';
import { UCPInfraStack } from '../lib/ucp-infra-stack';
import { solutionId, solutionVersion } from './mock';

// Set up
const envName = 'dev';
const artifactBucket = 'unit-test-bucket';
const artifactBucketPath = 'unit-test-path';

// Setting system time to 0 to avoid date mismatch in snapshot tests
jest.useFakeTimers();
jest.setSystemTime(0);

const app = new App({
    context: {
        envName,
        artifactBucket,
        artifactBucketPath,
        errorTTL: '7',
        ecrRepository: '',
        ecrPublishRoleArn: ''
    }
});

const stack = new UCPInfraStack(app, 'UCPInfraStack', {
    solutionId,
    solutionName: 'Unified Profiles for Travelers and Guests on AWS',
    solutionVersion,
    ecrProps: { publishEcrAssets: true },
    isByoVpcTemplate: false
});
const byoStack = new UCPInfraStack(app, 'UCPInfraStackByoVpc', {
    solutionId,
    solutionName: 'Unified Profiles for Travelers and Guests on AWS',
    solutionVersion,
    ecrProps: { publishEcrAssets: true },
    isByoVpcTemplate: true
});

const stackTemplate = Template.fromStack(stack);
const byoStackTemplate = Template.fromStack(byoStack);

const stackTemplateJson = stackToTemplateJson(stackTemplate);
const byoVpcStackTemplateJson = stackToTemplateJson(byoStackTemplate);

function stackToTemplateJson(template: Template): { [key: string]: any } {
    const templateJson = template.toJSON();
    // Omit hashed cdk assets
    const taskDefinitions = Object.keys(template.findResources('AWS::ECS::TaskDefinition'));
    for (const td of taskDefinitions) {
        templateJson.Resources[td].Properties.ContainerDefinitions.forEach((container: { Image: unknown }) => {
            container.Image = 'Omitted';
        });
    }

    // Omit hashed zip files
    const glueAssets = Object.keys(template.findResources('AWS::Glue::Job'));
    for (const ga of glueAssets) {
        templateJson.Resources[ga].Properties.DefaultArguments['--extra-py-files'] = 'Omitted';
    }
    const lambdas = Object.keys(template.findResources('AWS::Lambda::Function'));
    for (const lambda of lambdas) {
        templateJson.Resources[lambda].Properties.Code = 'Omitted';
    }
    return templateJson;
}

test('UCPInfraStack', () => {
    // Prepare
    const app = new App({
        context: {
            envName,
            artifactBucket,
            artifactBucketPath,
            ecrRepository: 'my-repository',
            imageAssetPublishingRoleArn: 'my-role'
        }
    });

    // Call & Verify
    expect(
        () =>
            new UCPInfraStack(app, 'UCPInfraStack', {
                env: {
                    account: '123456789012',
                    region: 'us-east-1'
                },
                solutionId: 'SO0244',
                solutionName: 'Unified Profiles for Travelers and Guests on AWS',
                solutionVersion: 'v1.0.0',
                ecrProps: { publishEcrAssets: true },
                isByoVpcTemplate: false
            })
    ).not.toThrow();
});

test('UCPInfraStackByoVpc', () => {
    // Prepare
    const app = new App({
        context: {
            envName,
            artifactBucket,
            artifactBucketPath,
            ecrRepository: 'my-repository',
            imageAssetPublishingRoleArn: 'my-role'
        }
    });

    // Call & Verify
    expect(
        () =>
            new UCPInfraStack(app, 'UCPInfraStack', {
                env: {
                    account: '123456789012',
                    region: 'us-east-1'
                },
                solutionId: 'SO0244',
                solutionName: 'Unified Profiles for Travelers and Guests on AWS',
                solutionVersion: 'v1.0.0',
                ecrProps: { publishEcrAssets: true },
                isByoVpcTemplate: true
            })
    ).not.toThrow();
});

test('UCPInfraStack check additional dependencies', () => {
    // Prepare
    const app = new App({
        context: {
            envName,
            artifactBucket,
            artifactBucketPath,
            ecrRepository: 'my-repository',
            imageAssetPublishingRoleArn: 'my-role'
        }
    });

    // Call
    const stack = new UCPInfraStack(app, 'UCPInfraStack', {
        env: {
            account: '123456789012',
            region: 'us-east-1'
        },
        solutionId: 'SO0244',
        solutionName: 'Unified Profiles for Travelers and Guests on AWS',
        solutionVersion: 'v1.0.0',
        ecrProps: { publishEcrAssets: true },
        isByoVpcTemplate: false
    });

    // Verify
    const template = Template.fromStack(stack);
    template.hasResource('AWS::Lambda::Function', {
        DependsOn: Match.arrayWith([
            Match.stringLikeRegexp('ucpindconnectorplaceholderPolicy'),
            Match.stringLikeRegexp('ucpMatch(.*)BucketPolicy')
        ])
    });
    template.hasResource('AWS::CloudFront::Distribution', {
        DependsOn: Match.arrayWith([Match.stringLikeRegexp('ucpconnectorfelogs(.*)Policy')])
    });
});

test('Error when envName is not provided', () => {
    // Prepare
    const app = new App({
        context: {
            artifactBucket,
            artifactBucketPath,
            ecrRepository: 'my-repository',
            imageAssetPublishingRoleArn: 'my-role'
        }
    });

    // Call & Verify
    expect(
        () =>
            new UCPInfraStack(app, 'UCPInfraStack', {
                env: {
                    account: '123456789012',
                    region: 'us-east-1'
                },
                solutionId: 'SO0244',
                solutionName: 'Unified Profiles for Travelers and Guests on AWS',
                solutionVersion: 'v1.0.0',
                ecrProps: { publishEcrAssets: true },
                isByoVpcTemplate: false
            })
    ).toThrow(Error('No environment name provided for stack'));
});

test('Error when artifactBucket is not provided', () => {
    // Prepare
    const app = new App({
        context: {
            envName,
            artifactBucketPath,
            ecrRepository: 'my-repository',
            imageAssetPublishingRoleArn: 'my-role'
        }
    });

    // Call & Verify
    expect(
        () =>
            new UCPInfraStack(app, 'UCPInfraStack', {
                env: {
                    account: '123456789012',
                    region: 'us-east-1'
                },
                solutionId: 'SO0244',
                solutionName: 'Unified Profiles for Travelers and Guests on AWS',
                solutionVersion: 'v1.0.0',
                ecrProps: { publishEcrAssets: true },
                isByoVpcTemplate: false
            })
    ).toThrow(Error('No bucket name provided for stack'));
});

test('permissions boundaries set conditionally', () => {
    const roles = Object.keys(stackTemplate.findResources('AWS::IAM::Role'));
    const templateJson = stackTemplate.toJSON();
    for (const role of roles) {
        expect(templateJson.Resources[role].Properties.PermissionsBoundary).toEqual({
            'Fn::If': [
                'permissionBoundaryCondition',
                {
                    Ref: 'permissionBoundaryArn'
                },
                {
                    Ref: 'AWS::NoValue'
                }
            ]
        });
    }
});

test('default stack matches snapshot', () => {
    expect(stackTemplateJson).toMatchSnapshot('default');
});

test('byo vpc stack matches snapshot', () => {
    expect(byoVpcStackTemplateJson).toMatchSnapshot('byo');
});
