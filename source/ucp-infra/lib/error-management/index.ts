// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import * as cdk from 'aws-cdk-lib';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as kinesis from 'aws-cdk-lib/aws-kinesis';
import * as kms from 'aws-cdk-lib/aws-kms';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import path from 'path';
import { Output } from '../../tah-cdk-common/core';
import { CdkBase, CdkBaseProps, DynamoDbTableKey, LambdaProps } from '../cdk-base';
import { ProfileStorageOutput } from '../profile-storage';

export type ErrorManagementProps = {
    readonly dynamoDbKey: kms.Key;
    readonly errorTtl: string;
    // Low cost storage
    readonly profileStorageOutput: ProfileStorageOutput;
    readonly changeProcessorKinesisStream: kinesis.Stream;
} & LambdaProps;

type ErrorDynamoDbTables = {
    readonly errorTable: dynamodb.Table;
    readonly errorRetryTable: dynamodb.Table;
};

export class ErrorManagement extends CdkBase {
    public readonly errorLambdaFunction: lambda.Function;
    public readonly errorTable: dynamodb.Table;
    public readonly errorRetryLambdaFunction: lambda.Function;
    public readonly errorRetryTable: dynamodb.Table;
    private readonly props: ErrorManagementProps;

    constructor(cdkBaseProps: CdkBaseProps, props: ErrorManagementProps) {
        super(cdkBaseProps);

        this.props = props;

        // DynamoDB tables
        const { errorTable, errorRetryTable } = this.createErrorDynamoDbTables();
        this.errorTable = errorTable;
        this.errorRetryTable = errorRetryTable;

        // Lambda functions
        this.errorLambdaFunction = this.createErrorProcessorLambdaFunction();
        this.errorRetryLambdaFunction = this.createErrorRetryLambdaFunction();

        // Additional permissions
        this.errorTable.grantReadWriteData(this.errorLambdaFunction);
        this.errorTable.grantReadWriteData(this.errorRetryLambdaFunction);
        this.errorRetryTable.grantReadWriteData(this.errorRetryLambdaFunction);
        this.errorRetryLambdaFunction.addToRolePolicy(
            // Permission to retry adding profile objects
            new iam.PolicyStatement({
                actions: ['profile:PutProfileObject'],
                effect: iam.Effect.ALLOW,
                resources: ['*']
            })
        );
        this.errorRetryLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['dynamodb:DescribeStream', 'dynamodb:GetRecords', 'dynamodb:GetShardIterator', 'dynamodb:ListStreams'],
                effect: iam.Effect.ALLOW,
                resources: [cdk.Fn.join('', [errorTable.tableArn, '/stream/*'])]
            })
        );

        // Outputs
        Output.add(this.stack, 'dynamodbErrorTableName', errorTable.tableName);
        Output.add(this.stack, 'retryLambdaName', this.errorLambdaFunction.functionName);
    }

    /**
     * Creates an error processor Lambda function.
     * This function processes all messages in all error queues and stores errors in an error DynamoDB table.
     * @returns Error processor Lambda function
     */
    private createErrorProcessorLambdaFunction(): lambda.Function {
        const ucpErrorLambdaPrefix = 'ucpError';
        const errorProcessorLambdaFunction = new GoFunction(this.stack, ucpErrorLambdaPrefix + this.envName, {
            entry: path.join(__dirname, '..', '..', '..', 'ucp-error', 'src', 'main', 'main.go'),
            bundling: {
                environment: { GOOS: 'linux', GOARCH: 'arm64', GOWORK: 'off' }
            },
            functionName: ucpErrorLambdaPrefix + this.envName,
            runtime: lambda.Runtime.PROVIDED_AL2,
            architecture: lambda.Architecture.ARM_64,
            tracing: lambda.Tracing.ACTIVE,
            timeout: cdk.Duration.seconds(30), // Timeout should be aligned with SQS queue visibility timeout.
            environment: {
                DYNAMO_PK: DynamoDbTableKey.ERROR_TABLE_PK,
                DYNAMO_SK: DynamoDbTableKey.ERROR_TABLE_SK,
                DYNAMO_TABLE: this.errorTable.tableName,
                LAMBDA_ACCOUNT_ID: cdk.Aws.ACCOUNT_ID,
                LAMBDA_ENV: this.envName,
                LAMBDA_REGION: cdk.Aws.REGION,
                TTL: this.props.errorTtl,
                // For usage metrics
                SEND_ANONYMIZED_DATA: this.props.sendAnonymizedData,
                METRICS_SOLUTION_ID: this.props.solutionId,
                METRICS_SOLUTION_VERSION: this.props.solutionVersion
            }
        });

        return errorProcessorLambdaFunction;
    }

    /**
     * Creates an error retry Lambda function.
     * This function subscribes the error table stream and retries to add profiles into Amazon Connect Customer Profiles.
     * @returns Error retry Lambda function
     */
    private createErrorRetryLambdaFunction(): lambda.Function {
        const ucpErrorRetryLambdaPrefix = 'ucpRetry';
        return new GoFunction(this.stack, 'ucpErrorRetry' + this.envName, {
            entry: path.join(__dirname, '..', '..', '..', 'ucp-retry', 'src', 'main', 'main.go'),
            functionName: ucpErrorRetryLambdaPrefix + this.envName,
            runtime: lambda.Runtime.PROVIDED_AL2,
            architecture: lambda.Architecture.ARM_64,
            tracing: lambda.Tracing.ACTIVE,
            timeout: cdk.Duration.seconds(900),
            environment: {
                ERROR_PK: DynamoDbTableKey.ERROR_TABLE_PK,
                ERROR_SK: DynamoDbTableKey.ERROR_TABLE_SK,
                ERROR_TABLE: this.errorTable.tableName,
                LAMBDA_REGION: cdk.Aws.REGION,
                RETRY_TABLE: this.errorRetryTable.tableName,
                RETRY_PK: DynamoDbTableKey.ERROR_RETRY_TABLE_PK,
                RETRY_SK: DynamoDbTableKey.ERROR_RETRY_TABLE_SK,
                // For usage metrics
                SEND_ANONYMIZED_DATA: this.props.sendAnonymizedData,
                METRICS_SOLUTION_ID: this.props.solutionId,
                METRICS_SOLUTION_VERSION: this.props.solutionVersion,
                // For low-cost storage
                AURORA_PROXY_ENDPOINT: this.props.profileStorageOutput.storageProxyEndpoint,
                AURORA_DB_NAME: this.props.profileStorageOutput.storageDbName,
                AURORA_DB_SECRET_ARN: this.props.profileStorageOutput.storageSecretArn,
                STORAGE_CONFIG_TABLE_NAME: this.props.profileStorageOutput.storageConfigTable?.tableName ?? '',
                STORAGE_CONFIG_TABLE_PK: this.props.profileStorageOutput.storageConfigTablePk,
                STORAGE_CONFIG_TABLE_SK: this.props.profileStorageOutput.storageConfigTableSk,
                CHANGE_PROC_KINESIS_STREAM_NAME: this.props.changeProcessorKinesisStream.streamName
            }
        });
    }

    /**
     * Creates error DynamoDB tables.
     * @returns Error DynamoDB tables
     */
    private createErrorDynamoDbTables(): ErrorDynamoDbTables {
        const errorTable = new dynamodb.Table(this.stack, 'ucpErrorTable', {
            tableName: 'ucp-error-table-' + this.envName,
            partitionKey: { name: DynamoDbTableKey.ERROR_TABLE_PK, type: dynamodb.AttributeType.STRING },
            sortKey: { name: DynamoDbTableKey.ERROR_TABLE_SK, type: dynamodb.AttributeType.STRING },
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            pointInTimeRecovery: true,
            encryptionKey: this.props.dynamoDbKey,
            timeToLiveAttribute: 'ttl',
            stream: dynamodb.StreamViewType.NEW_AND_OLD_IMAGES
        });

        const errorRetryTable = new dynamodb.Table(this.stack, 'ucpRetryTable', {
            tableName: 'ucp-retry-table-' + this.envName,
            partitionKey: { name: DynamoDbTableKey.ERROR_RETRY_TABLE_PK, type: dynamodb.AttributeType.STRING },
            sortKey: { name: DynamoDbTableKey.ERROR_RETRY_TABLE_SK, type: dynamodb.AttributeType.STRING },
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            encryptionKey: this.props.dynamoDbKey,
            pointInTimeRecovery: true
        });

        return { errorTable, errorRetryTable };
    }
}
