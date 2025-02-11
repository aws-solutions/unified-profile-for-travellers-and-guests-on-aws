// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as kinesis from 'aws-cdk-lib/aws-kinesis';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as lambdaEventSources from 'aws-cdk-lib/aws-lambda-event-sources';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import path from 'path';
import { Output } from '../../tah-cdk-common/core';
import * as tahSqs from '../../tah-cdk-common/sqs';
import { CdkBase, CdkBaseProps, LambdaProps } from '../cdk-base';
import { ProfileStorageOutput } from '../profile-storage';

type MergerProps = {
    readonly profileStorageOutput: ProfileStorageOutput;
    readonly changeProcessorKinesisStream: kinesis.Stream;
    readonly uptVpc: ec2.IVpc;
} & LambdaProps;

export class MergeProcessor extends CdkBase {
    public dlqMerger: tahSqs.Queue;
    public mergeQueue: tahSqs.Queue;
    public mergerLambda: lambda.Function;

    private props: MergerProps;

    constructor(cdkProps: CdkBaseProps, props: MergerProps) {
        super(cdkProps);

        this.props = props;

        // Dead letter queue
        this.dlqMerger = this.createDeadLetterQueue('ucp-merger-dlq-');

        // Merge queue
        this.mergeQueue = this.createMergeQueue(this.dlqMerger);

        // Profile Merger Lambda
        this.mergerLambda = this.createMergerLambda();

        // Permissions
        this.mergerLambda.addEventSource(
            new lambdaEventSources.SqsEventSource(this.mergeQueue, {
                reportBatchItemFailures: true,
                batchSize: 1
            })
        );
        this.dlqMerger.grantConsumeMessages(this.mergerLambda);
        this.dlqMerger.grantSendMessages(this.mergerLambda);

        this.mergeQueue.grantSendMessages(this.mergerLambda);

        // Low-cost storage permissions
        // Tables are created/deleted via API and cannot be scoped down
        this.mergerLambda.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['dynamodb:DescribeTable', 'dynamodb:Query', 'dynamodb:DeleteItem', 'dynamodb:BatchWriteItem', 'dynamodb:PutItem'],
                effect: iam.Effect.ALLOW,
                resources: [`arn:${cdk.Aws.PARTITION}:dynamodb:${cdk.Aws.REGION}:${cdk.Aws.ACCOUNT_ID}:table/*`]
            })
        );

        Output.add(this.stack, 'mergeQueueUrl', this.mergeQueue.queueUrl);
        Output.add(this.stack, 'mergerLambdaName', this.mergerLambda.functionName);
    }

    /**
     * Create an SQS queue for handling merge requests
     * @returns SQS queue
     */
    private createMergeQueue(dlq: tahSqs.Queue): sqs.Queue {
        return new tahSqs.Queue(this.stack, 'ucp-merge-queue-' + this.envName, {
            visibilityTimeout: cdk.Duration.seconds(60),
            deadLetterQueue: {
                maxReceiveCount: 3,
                queue: dlq
            }
        });
    }

    /**
     * Creates a merge queue.
     * @returns Merge queue for handling merge requests
     */
    private createDeadLetterQueue(queueName: string): tahSqs.Queue {
        return new tahSqs.Queue(this.stack, queueName + this.envName);
    }

    /**
     * Creates a Lambda function for handling merge requests.
     * @returns Merge lambda function
     */
    private createMergerLambda(): lambda.Function {
        const ucpMergerLambdaPrefix = 'ucpMerger';
        const mergerLambdaFunction = new GoFunction(this.stack, ucpMergerLambdaPrefix + this.envName, {
            entry: path.join(__dirname, '..', '..', '..', 'ucp-merger', 'src', 'main', 'main.go'),
            bundling: {
                environment: { GOOS: 'linux', GOARCH: 'arm64', GOWORK: 'off' }
            },
            functionName: ucpMergerLambdaPrefix + this.envName,
            runtime: lambda.Runtime.PROVIDED_AL2,
            architecture: lambda.Architecture.ARM_64,
            environment: {
                // General info
                LAMBDA_REGION: cdk.Aws.REGION,
                // Solution metrics
                METRICS_SOLUTION_ID: this.props.solutionId,
                METRICS_SOLUTION_VERSION: this.props.solutionVersion,
                // For low-cost storage
                AURORA_PROXY_ENDPOINT: this.props.profileStorageOutput.storageProxyEndpoint,
                AURORA_DB_NAME: this.props.profileStorageOutput.storageDbName,
                AURORA_DB_SECRET_ARN: this.props.profileStorageOutput.storageSecretArn,
                STORAGE_CONFIG_TABLE_NAME: this.props.profileStorageOutput.storageConfigTable?.tableName ?? '',
                STORAGE_CONFIG_TABLE_PK: this.props.profileStorageOutput.storageConfigTablePk,
                STORAGE_CONFIG_TABLE_SK: this.props.profileStorageOutput.storageConfigTableSk,
                CHANGE_PROC_KINESIS_STREAM_NAME: this.props.changeProcessorKinesisStream.streamName,
                // Merge Queue
                MERGE_QUEUE_URL: this.mergeQueue.queueUrl,
                DEAD_LETTER_QUEUE_URL: this.dlqMerger.queueUrl
            },
            retryAttempts: 0,
            timeout: cdk.Duration.seconds(60),
            memorySize: 128,
            tracing: lambda.Tracing.ACTIVE,
            // Low cost storage security config
            vpc: this.props.uptVpc,
            securityGroups: [this.props.profileStorageOutput.lambdaToProxyGroup]
        });

        mergerLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    'profile:PutProfileObject',
                    'profile:ListProfileObjects',
                    'profile:ListProfileObjectTypes',
                    'profile:GetProfileObjectType',
                    'profile:PutProfileObjectType',
                    'profile:SearchProfiles',
                    'profile:MergeProfiles',
                    'profile:DeleteProfile'
                ],
                effect: iam.Effect.ALLOW,
                resources: ['*']
            })
        );

        return mergerLambdaFunction;
    }
}
