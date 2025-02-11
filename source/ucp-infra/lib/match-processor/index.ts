// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import * as cdk from 'aws-cdk-lib';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as kinesis from 'aws-cdk-lib/aws-kinesis';
import * as kms from 'aws-cdk-lib/aws-kms';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as s3Notifications from 'aws-cdk-lib/aws-s3-notifications';
import path from 'path';
import * as tahCore from '../../tah-cdk-common/core';
import * as tahS3 from '../../tah-cdk-common/s3';
import { CdkBase, CdkBaseProps, DynamoDbTableKey, LambdaProps } from '../cdk-base';
import { ProfileStorageOutput } from '../profile-storage';

type MatchProcessorProps = {
    readonly accessLogBucket: s3.Bucket;
    readonly dynamoDbKey: kms.Key;
    readonly profileStorageOutput: ProfileStorageOutput;
    readonly changeProcessorKinesisStream: kinesis.Stream;
    readonly portalConfigTable: dynamodb.Table;
    readonly uptVpc: ec2.IVpc;
} & LambdaProps;

export class MatchProcessor extends CdkBase {
    public readonly matchBucket: tahS3.Bucket;
    public readonly matchLambdaFunction: lambda.Function;
    public readonly matchTable: dynamodb.Table;

    private readonly prefix = 'ucpMatch';
    private readonly props: MatchProcessorProps;

    constructor(cdkProps: CdkBaseProps, props: MatchProcessorProps) {
        super(cdkProps);

        this.props = props;

        // Match processor resources
        this.matchBucket = this.createMatchBucket();
        this.matchTable = this.createMatchTable();
        this.matchLambdaFunction = this.createMatchLambdaFunction();

        // Permissions
        this.matchBucket.grantReadWrite(this.matchLambdaFunction);
        this.matchTable.grantReadWriteData(this.matchLambdaFunction);

        // Lambda to sync the matches to the match table
        this.matchBucket.addEventNotification(s3.EventType.OBJECT_CREATED, new s3Notifications.LambdaDestination(this.matchLambdaFunction));

        // Outputs
        tahCore.Output.add(this.stack, 'matchBucket', this.matchBucket.bucketName);
        tahCore.Output.add(this.stack, 'dynamoTable', this.matchTable.tableName);
    }

    /**
     * Creates a match S3 bucket.
     * @returns Match S3 bucket
     */
    private createMatchBucket(): tahS3.Bucket {
        const matchBucket = new tahS3.Bucket(this.stack, this.prefix + this.envName + 'Bucket', this.props.accessLogBucket);
        matchBucket.addToResourcePolicy(
            new iam.PolicyStatement({
                actions: ['s3:GetBucketLocation', 's3:GetBucketPolicy', 's3:GetObject', 's3:ListBucket', 's3:PutObject'],
                effect: iam.Effect.ALLOW,
                principals: [new iam.ServicePrincipal('profile.amazonaws.com')],
                resources: [matchBucket.bucketArn, matchBucket.arnForObjects('*')]
            })
        );

        return matchBucket;
    }

    /**
     * Creates a match DynamoDB table.
     * @returns Match DynamoDB table
     */
    private createMatchTable(): dynamodb.Table {
        const matchTable = new dynamodb.Table(this.stack, 'ucpMatchTable', {
            tableName: 'ucp-match-table-new-' + this.envName,
            partitionKey: { name: DynamoDbTableKey.MATCH_TABLE_PK, type: dynamodb.AttributeType.STRING },
            sortKey: { name: DynamoDbTableKey.MATCH_TABLE_SK, type: dynamodb.AttributeType.STRING },
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            pointInTimeRecovery: true,
            encryptionKey: this.props.dynamoDbKey
        });
        matchTable.addGlobalSecondaryIndex({
            indexName: 'matchesByConfidenceScore',
            partitionKey: { name: DynamoDbTableKey.MATCH_GSI_PK, type: dynamodb.AttributeType.STRING },
            sortKey: { name: DynamoDbTableKey.MATCH_GSI_SK, type: dynamodb.AttributeType.STRING }
        });

        return matchTable;
    }

    /**
     * Creates a match Lambda function.
     * @returns Match Lambda function
     */
    private createMatchLambdaFunction(): lambda.Function {
        const ucpMatchLambda = new GoFunction(this.stack, this.prefix + this.envName, {
            entry: path.join(__dirname, '..', '..', '..', 'ucp-match', 'src', 'main', 'main.go'),
            bundling: {
                environment: { GOOS: 'linux', GOARCH: 'arm64', GOWORK: 'off' }
            },
            functionName: this.prefix + this.envName,
            runtime: lambda.Runtime.PROVIDED_AL2,
            architecture: lambda.Architecture.ARM_64,
            environment: {
                LAMBDA_ACCOUNT_ID: cdk.Aws.ACCOUNT_ID,
                LAMBDA_ENV: this.envName,
                MATCH_BUCKET_NAME: this.matchBucket.bucketName,
                MATCH_PREFIX: this.prefix,
                LAMBDA_REGION: cdk.Aws.REGION,
                DYNAMO_TABLE: this.matchTable.tableName,
                DYNAMO_PK: DynamoDbTableKey.MATCH_TABLE_PK,
                DYNAMO_SK: DynamoDbTableKey.MATCH_TABLE_SK,
                PORTAL_CONFIG_TABLE_NAME: this.props.portalConfigTable.tableName,
                PORTAL_CONFIG_TABLE_PK: DynamoDbTableKey.PORTAL_CONFIG_TABLE_PK,
                PORTAL_CONFIG_TABLE_SK: DynamoDbTableKey.PORTAL_CONFIG_TABLE_SK,
                // Usage metrics
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
                CHANGE_PROC_KINESIS_STREAM_NAME: this.props.changeProcessorKinesisStream.streamName,
                //CP Indexer Keys
                INDEX_PK: DynamoDbTableKey.CP_INDEX_TABLE_PK,
                INDEX_SK: DynamoDbTableKey.CP_INDEX_TABLE_SK
            },
            timeout: cdk.Duration.seconds(30),
            tracing: lambda.Tracing.ACTIVE,
            vpc: this.props.uptVpc,
            securityGroups: [this.props.profileStorageOutput.lambdaToProxyGroup]
        });

        ucpMatchLambda.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['secretsmanager:GetSecretValue'],
                effect: iam.Effect.ALLOW,
                resources: [this.props.profileStorageOutput.storageSecretArn]
            })
        );

        return ucpMatchLambda;
    }
}
