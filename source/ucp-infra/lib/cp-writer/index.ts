// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import * as cdk from 'aws-cdk-lib';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as lambdaEventSources from 'aws-cdk-lib/aws-lambda-event-sources';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import path from 'path';
import { Output } from '../../tah-cdk-common/core';
import * as tahSqs from '../../tah-cdk-common/sqs';
import { CdkBase, CdkBaseProps, DynamoDbTableKey, LambdaProps } from '../cdk-base';

type CpWriterProps = {
    maxConcurrency: number;
} & LambdaProps;

export type CPWriterOutput = {
    writerQueue: sqs.Queue;
    writerLambda: lambda.Function;
};

export class CPWriter extends CdkBase {
    public writerDLQ: sqs.Queue;
    public writerQueue: sqs.Queue;
    public writerLambda: lambda.Function;

    private props: LambdaProps;

    constructor(cdkProps: CdkBaseProps, props: CpWriterProps) {
        super(cdkProps);

        this.props = props;

        // Writer DLQ
        this.writerDLQ = this.createWriterDLQ();

        // Writer Queue
        this.writerQueue = this.createWriterQueue();

        // Writer Lambda
        this.writerLambda = this.createWriterLambda();

        // SQS to Lambda event source
        this.writerLambda.addEventSource(
            new lambdaEventSources.SqsEventSource(this.writerQueue, {
                batchSize: 1,
                maxConcurrency: props.maxConcurrency
            })
        );

        // Add permissions
        this.writerQueue.grantConsumeMessages(this.writerLambda);
        this.writerQueue.grantSendMessages(this.writerLambda);

        // Add outputs
        Output.add(this.stack, 'lcsCpWriterQueueUrl', this.writerDLQ.queueName);
    }

    private createWriterDLQ(): sqs.Queue {
        return new tahSqs.Queue(this.stack, 'ucp-cp-writer-dlq' + this.envName);
    }

    private createWriterQueue(): sqs.Queue {
        return new tahSqs.Queue(this.stack, 'ucp-cp-writer-queue' + this.envName, {
            visibilityTimeout: cdk.Duration.seconds(300)
        });
    }

    private createWriterLambda(): lambda.Function {
        const ucpCpWriterLambdaPrefix = 'ucpCpWriter';
        const ucpCpWriterLambdaFunction = new GoFunction(this.stack, ucpCpWriterLambdaPrefix + this.envName, {
            entry: path.join(__dirname, '..', '..', '..', 'ucp-cp-writer', 'src', 'main', 'main.go'),
            bundling: {
                environment: { GOOS: 'linux', GOARCH: 'arm64', GOWORK: 'off' }
            },
            functionName: ucpCpWriterLambdaPrefix + this.envName,
            runtime: lambda.Runtime.PROVIDED_AL2,
            architecture: lambda.Architecture.ARM_64,
            environment: {
                // General info
                LAMBDA_REGION: cdk.Aws.REGION,
                // Solution metrics
                METRICS_SOLUTION_ID: this.props.solutionId,
                METRICS_SOLUTION_VERSION: this.props.solutionVersion,
                // DLQ
                DLQ_URL: this.writerDLQ.queueUrl,
                //CP Writer Queue URL
                CP_WRITER_QUEUE_URL: this.writerQueue.queueUrl,
                //CP Indexer Keys
                INDEX_PK: DynamoDbTableKey.CP_INDEX_TABLE_PK,
                INDEX_SK: DynamoDbTableKey.CP_INDEX_TABLE_SK
            },
            tracing: lambda.Tracing.ACTIVE,
            deadLetterQueue: this.writerDLQ,
            deadLetterQueueEnabled: true,
            timeout: cdk.Duration.seconds(300)
        });

        ucpCpWriterLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['profile:PutProfileObject', 'profile:DeleteProfile', 'profile:MergeProfiles'],
                effect: iam.Effect.ALLOW,
                resources: ['*']
            })
        );
        return ucpCpWriterLambdaFunction;
    }
}
