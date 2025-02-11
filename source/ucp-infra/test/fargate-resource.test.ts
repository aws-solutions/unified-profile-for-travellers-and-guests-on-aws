// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { FargateResource } from '../lib/fargate-resource/fargate-resource';
import { createKinesisStream, createProfileStorageOutput, createVpc, solutionId, solutionVersion } from './mock';

describe('fargate resources', () => {
    it('matches snapshot', () => {
        const app = new App();
        const stack = new Stack(app);
        const vpc = createVpc(stack, 'vpc');
        const profileStorageOutput = createProfileStorageOutput(stack, 'profileStorageOutput', vpc);
        const changeProcessorKinesisStream = createKinesisStream(stack, 'kinesisStream');

        new FargateResource(stack, {
            profileStorageOutput,
            mergeQueueUrl: 'mergeQueueUrl',
            cpWriterQueueUrl: 'cpWriterQueueUrl',
            changeProcessorKinesisStream,
            ecrProps: { publishEcrAssets: true },
            solutionId,
            solutionVersion,
            uptVpc: vpc,
            logLevel: 'DEBUG'
        });

        const template = Template.fromStack(stack);
        const taskIds = Object.keys(template.findResources('AWS::ECS::TaskDefinition'));
        const templateJson = template.toJSON();
        for (const taskId of taskIds) {
            templateJson.Resources[taskId].Properties.ContainerDefinitions.forEach((container: { Image: unknown }) => {
                container.Image = 'Omitted';
            });
        }

        expect(Template.fromStack(stack).toJSON()).toMatchSnapshot();
    });
});
