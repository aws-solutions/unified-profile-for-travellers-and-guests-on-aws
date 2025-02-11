// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Match, Template } from 'aws-cdk-lib/assertions';
import { Dashboard } from '../lib/dashboard';
import { createDynamoDbTable, createHttpApi, createKinesisStream, createLambdaFunction, createSqsQueue, envName } from './mock';

test('Dashboard', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const entryKinesisStream = createKinesisStream(stack, 'entryKinesisStream');
    const entryLambdaFunction = createLambdaFunction(stack, 'entryLambdaFunction');
    const ingestorKinesisStream = createKinesisStream(stack, 'ingestorKinesisStream');
    const ingestorLambdaFunction = createLambdaFunction(stack, 'ingestorLambdaFunction');
    const changeProcessorKinesisStream = createKinesisStream(stack, 'changeProcessorKinesisStream');
    const changeProcessorLambdaFunction = createLambdaFunction(stack, 'changeProcessorLambdaFunction');
    const airBookingBatchErrorQueue = createSqsQueue(stack, 'airBookingBatchErrorQueue');
    const clickStreamBatchErrorQueue = createSqsQueue(stack, 'clickStreamBatchErrorQueue');
    const customerServiceInteractionBatchErrorQueue = createSqsQueue(stack, 'customerServiceInteractionBatchErrorQueue');
    const guestProfileBatchErrorQueue = createSqsQueue(stack, 'guestProfileBatchErrorQueue');
    const hotelBookingBatchErrorQueue = createSqsQueue(stack, 'hotelBookingBatchErrorQueue');
    const hotelStayBatchErrorQueue = createSqsQueue(stack, 'hotelStayBatchErrorQueue');
    const paxProfileBatchErrorQueue = createSqsQueue(stack, 'paxProfileBatchErrorQueue');
    const dlqChangeProcessor = createSqsQueue(stack, 'dlqChangeProcessor');
    const dlqGo = createSqsQueue(stack, 'dlqGo');
    const dlqPython = createSqsQueue(stack, 'dlqPython');
    const dlqBatchProcessor = createSqsQueue(stack, 'dlqBatchProcessor');
    const connectProfileDomainErrorQueue = createSqsQueue(stack, 'connectProfileDomainErrorQueue');
    const errorRetryLambdaFunction = createLambdaFunction(stack, 'errorRetryLambdaFunction');
    const errorTable = createDynamoDbTable(stack, 'errorTable', true);
    const api = createHttpApi(stack, 'api');
    const syncLambdaFunction = createLambdaFunction(stack, 'syncLambdaFunction');
    const cpWriterQueue = createSqsQueue(stack, 'cpWriterQueue');

    // Call
    new Dashboard(
        { envName, stack },
        {
            entryKinesisStream,
            entryLambdaFunction,
            ingestorKinesisStream,
            ingestorLambdaFunction,
            changeProcessorKinesisStream,
            changeProcessorLambdaFunction,
            airBookingBatchErrorQueue,
            clickStreamBatchErrorQueue,
            customerServiceInteractionBatchErrorQueue,
            guestProfileBatchErrorQueue,
            hotelBookingBatchErrorQueue,
            hotelStayBatchErrorQueue,
            paxProfileBatchErrorQueue,
            dlqChangeProcessor,
            dlqGo,
            dlqPython,
            dlqBatchProcessor,
            connectProfileDomainErrorQueue,
            errorRetryLambdaFunction,
            errorTable,
            api,
            syncLambdaFunction,
            cpWriterQueue
        }
    );

    // Verify
    const template = Template.fromStack(stack);
    template.resourceCountIs('AWS::CloudWatch::Dashboard', 1);
    template.hasResourceProperties('AWS::CloudWatch::Dashboard', {
        DashboardBody: {
            'Fn::Join': [
                '',
                Match.arrayWith([
                    {
                        'Fn::Select': [3, { 'Fn::Split': ['/', { 'Fn::GetAtt': [Match.anyValue(), 'StreamArn'] }] }]
                    }
                ])
            ]
        }
    });
});
