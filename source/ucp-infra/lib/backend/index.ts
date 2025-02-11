// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as glueAlpha from '@aws-cdk/aws-glue-alpha';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import * as cdk from 'aws-cdk-lib';
import * as apiGatewayV2 from 'aws-cdk-lib/aws-apigatewayv2';
import * as apiAuthorizer from 'aws-cdk-lib/aws-apigatewayv2-authorizers';
import * as apiIntegrations from 'aws-cdk-lib/aws-apigatewayv2-integrations';
import { CfnPreparedStatement, CfnWorkGroup } from 'aws-cdk-lib/aws-athena';
import * as cognito from 'aws-cdk-lib/aws-cognito';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { Cluster } from 'aws-cdk-lib/aws-ecs';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as kinesis from 'aws-cdk-lib/aws-kinesis';
import * as kms from 'aws-cdk-lib/aws-kms';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import { LogGroup } from 'aws-cdk-lib/aws-logs';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import path from 'path';
import * as permissions from '../../../appPermissions.json';
import * as tahCore from '../../tah-cdk-common/core';
import * as tahS3 from '../../tah-cdk-common/s3';
import { CdkBase, CdkBaseProps, DynamoDbTableKey, LambdaProps } from '../cdk-base';
import { BatchProcessingOutput } from '../data-ingestion';
import { ProfileStorageOutput } from '../profile-storage';

export type BackendProps = {
    readonly configTable: dynamodb.Table;
    readonly connectProfileDomainErrorQueue: sqs.Queue;
    readonly connectProfileImportBucket: s3.Bucket;
    readonly dataLakeAdminRole: iam.Role;
    readonly errorRetryLambdaFunction: lambda.Function;
    readonly errorTable: dynamodb.Table;
    readonly glueDatabase: glueAlpha.Database;
    readonly glueTravelerTable: glueAlpha.S3Table;
    readonly industryConnectorDataSyncRole: iam.Role;
    readonly isPlaceholderConnectorBucket: string;
    readonly matchBucket: s3.Bucket;
    readonly matchTable: dynamodb.Table;
    readonly privacySearchResultsTable: dynamodb.Table;
    readonly athenaSearchProfileS3PathsPreparedStatement: CfnPreparedStatement;
    readonly athenaGetS3PathsByConnectIdsPreparedStatement: CfnPreparedStatement;
    readonly athenaResultsBucket: tahS3.Bucket;
    readonly athenaWorkGroup: CfnWorkGroup;
    readonly outputBucket: s3.Bucket;
    readonly portalConfigTable: dynamodb.Table;
    readonly stackLogicalId: string;
    readonly syncLambdaFunction: lambda.Function;
    readonly travelerS3RootPath: string;
    // Batch processing
    readonly airBookingBatch: BatchProcessingOutput;
    readonly clickStreamBatch: BatchProcessingOutput;
    readonly customerServiceInteractionBatch: BatchProcessingOutput;
    readonly guestProfileBatch: BatchProcessingOutput;
    readonly hotelBookingBatch: BatchProcessingOutput;
    readonly hotelStayBatch: BatchProcessingOutput;
    readonly paxProfileBatch: BatchProcessingOutput;
    // Low cost storage
    readonly profileStorageOutput: ProfileStorageOutput;
    readonly changeProcessorKinesisStream: kinesis.Stream;
    // GDPR
    readonly s3ExciseQueue: sqs.Queue;
    readonly gdprPurgeLogGroup: LogGroup;
    readonly ssmParamNamespace: string;
    readonly irCluster: Cluster;
    // VPC
    readonly uptVpc: ec2.IVpc;
} & LambdaProps;

type ApiRoute = {
    readonly methods: apiGatewayV2.HttpMethod[];
    readonly path: string;
};

export class Backend extends CdkBase {
    public readonly api: apiGatewayV2.HttpApi;
    public readonly asyncLambdaFunction: lambda.Function;
    public readonly backendLambdaFunction: lambda.Function;
    public readonly userPool: cognito.UserPool;
    public readonly userPoolClient: cognito.UserPoolClient;
    public readonly stage: apiGatewayV2.HttpStage;
    public readonly domainKey: kms.Key;

    private readonly props: BackendProps;

    constructor(cdkProps: CdkBaseProps, props: BackendProps) {
        super(cdkProps);

        this.props = props;

        // Cognito
        const { domain, userPool, userPoolClient } = this.createCognito();
        this.userPool = userPool;
        this.userPoolClient = userPoolClient;

        // Backend Lambda function
        this.domainKey = this.createDomainKey();
        this.asyncLambdaFunction = this.createAsyncLambdaFunction();
        this.backendLambdaFunction = this.createBackendLambdaFunction();

        // API
        const { api, stage } = this.createApi();
        this.api = api;
        this.stage = stage;

        // Permission
        this.asyncLambdaFunction.grantInvoke(this.backendLambdaFunction);

        // Outputs
        tahCore.Output.add(this.stack, 'userPoolId', this.userPool.userPoolId);
        tahCore.Output.add(this.stack, 'cognitoAppClientId', this.userPoolClient.userPoolClientId);
        tahCore.Output.add(
            this.stack,
            'tokenEndpoint',
            cdk.Fn.join('', ['https://', domain.domain, '.auth.', cdk.Aws.REGION, '.amazoncognito.com/oauth2/token'])
        );
        tahCore.Output.add(this.stack, 'cognitoDomain', domain.domain);
        tahCore.Output.add(this.stack, 'httpApiUrl', this.api.apiEndpoint ?? '');
        tahCore.Output.add(this.stack, 'ucpApiId', this.api.apiId ?? '');
        tahCore.Output.add(this.stack, 'apiStage', this.stage.stageName);
        tahCore.Output.add(this.stack, 'kmsKeyProfileDomain', this.domainKey.keyArn);
    }

    /**
     * Creates Cognito resources for the frontend and the backend.
     * @returns Cognito domain, user pool and user pool client
     */
    private createCognito(): { domain: cognito.CfnUserPoolDomain; userPool: cognito.UserPool; userPoolClient: cognito.UserPoolClient } {
        const userPool: cognito.UserPool = new cognito.UserPool(this.stack, 'ucpUserpool', {
            signInAliases: { email: true },
            userPoolName: 'ucpUserpool' + this.envName
        });

        const userPoolClient = new cognito.UserPoolClient(this.stack, 'ucpUserPoolClient', {
            userPool,
            oAuth: {
                flows: { authorizationCodeGrant: true, implicitCodeGrant: true },
                scopes: [
                    cognito.OAuthScope.PHONE,
                    cognito.OAuthScope.EMAIL,
                    cognito.OAuthScope.OPENID,
                    cognito.OAuthScope.PROFILE,
                    cognito.OAuthScope.COGNITO_ADMIN
                ],
                callbackUrls: ['http://localhost:4200'],
                logoutUrls: ['http://localhost:4200']
            },
            supportedIdentityProviders: [cognito.UserPoolClientIdentityProvider.COGNITO],
            authFlows: {
                userPassword: true,
                userSrp: true,
                adminUserPassword: true
            },
            generateSecret: false,
            refreshTokenValidity: cdk.Duration.days(30),
            userPoolClientName: 'ucpPortal'
        });

        // Adding the account ID in order to ensure uniqueness per account per env
        const domain = new cognito.CfnUserPoolDomain(this.stack, this.props.stackLogicalId + 'ucpCognitoDomain', {
            userPoolId: userPool.userPoolId,
            domain: cdk.Fn.join('-', ['ucp-domain', cdk.Aws.ACCOUNT_ID, this.envName])
        });

        const AdminRole = 'ffffffff';
        const CreateDomainPermission = 1 << permissions.CreateDomainPermission;
        const DeleteDomainPermission = 1 << permissions.DeleteDomainPermission;
        const DomainManagerRole = (CreateDomainPermission | DeleteDomainPermission).toString(16);

        const SaveHyperlinkPermission = 1 << permissions.SaveHyperlinkPermission;
        const RunGlueJobsPermission = 1 << permissions.RunGlueJobsPermission;
        const ClearAllErrorsPermission = 1 << permissions.ClearAllErrorsPermission;
        const GlobalSettingsRole = (RunGlueJobsPermission | ClearAllErrorsPermission | SaveHyperlinkPermission).toString(16);

        const globalAdminRoleGroup = new cognito.CfnUserPoolGroup(this.stack, 'GlobalAdminRole', {
            userPoolId: userPool.userPoolId,
            groupName: 'app-global-adminRole/' + AdminRole
        });
        const domainManagerRoleGroup = new cognito.CfnUserPoolGroup(this.stack, 'DomainManagerRole', {
            userPoolId: userPool.userPoolId,
            groupName: 'app-global-domainManagerRole/' + DomainManagerRole
        });

        const globalSettingsRoleGroup = new cognito.CfnUserPoolGroup(this.stack, 'GlobalSettingsRole', {
            userPoolId: userPool.userPoolId,
            groupName: 'app-global-settingsRole/' + GlobalSettingsRole
        });
        tahCore.Output.add(this.stack, 'domainManagerRoleGroup', domainManagerRoleGroup.groupName || '');
        tahCore.Output.add(this.stack, 'globalSettingsRoleGroup', globalSettingsRoleGroup.groupName || '');
        tahCore.Output.add(this.stack, 'globalAdminRoleGroup', globalAdminRoleGroup.groupName || '');

        return {
            domain,
            userPool,
            userPoolClient
        };
    }

    /**
     * Creates a HTTP API for backend service.
     * @returns HTTP API and stage
     */
    private createApi(): { api: apiGatewayV2.HttpApi; stage: apiGatewayV2.HttpStage } {
        const endpointName = 'ucp';
        const endpointAdmin = 'admin';
        const endpointAsyncEvent = 'async';
        const endpointErrors = 'error';
        const endpointFlows = 'flows';
        const endpointIndustryConnector = 'connector';
        const endpointJobs = 'jobs';
        const endpointLink = 'link';
        const endpointMerge = 'merge';
        const endpointUnmerge = 'unmerge';
        const endpointPortalConfig = 'portalConfig';
        const endpointProfile = 'profile';
        const endpointProfileSummary = `${endpointProfile}/summary`;
        const endpointRule = 'ruleSet';
        const endpointRuleCache = 'ruleSetCache';
        const endpointPrivacy = 'privacy';
        const endpointPrivacyPurge = `${endpointPrivacy}/purge`;
        const endpointInteractionHistory = 'interactionHistory';
        const endpointCache = 'cache';
        const endpointPromptConfig = 'promptConfig';
        const pathParameterId = '{id}';
        const apiRoutes: ApiRoute[] = [
            {
                path: `/${endpointName}/${endpointAdmin}`,
                methods: [apiGatewayV2.HttpMethod.GET, apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointAdmin}/${pathParameterId}`,
                methods: [apiGatewayV2.HttpMethod.DELETE, apiGatewayV2.HttpMethod.GET, apiGatewayV2.HttpMethod.PUT]
            },
            {
                path: `/${endpointName}/${endpointAsyncEvent}`,
                methods: [apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointErrors}`,
                methods: [apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointErrors}/${pathParameterId}`,
                methods: [apiGatewayV2.HttpMethod.DELETE, apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointFlows}`,
                methods: [apiGatewayV2.HttpMethod.GET, apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointIndustryConnector}`,
                methods: [apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointIndustryConnector}/${endpointLink}`,
                methods: [apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointJobs}`,
                methods: [apiGatewayV2.HttpMethod.GET, apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointMerge}`,
                methods: [apiGatewayV2.HttpMethod.GET, apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointMerge}/${pathParameterId}`,
                methods: [apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointUnmerge}`,
                methods: [apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointPortalConfig}`,
                methods: [apiGatewayV2.HttpMethod.GET, apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointProfile}`,
                methods: [apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointProfile}/${pathParameterId}`,
                methods: [apiGatewayV2.HttpMethod.DELETE, apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointProfileSummary}/${pathParameterId}`,
                methods: [apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointRule}`,
                methods: [apiGatewayV2.HttpMethod.GET, apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointRule}/activate`,
                methods: [apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointRuleCache}`,
                methods: [apiGatewayV2.HttpMethod.GET, apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointRuleCache}/activate`,
                methods: [apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointPrivacy}`,
                methods: [apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointPrivacy}/${pathParameterId}`,
                methods: [apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointPrivacy}`,
                methods: [apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointPrivacy}`,
                methods: [apiGatewayV2.HttpMethod.DELETE]
            },
            {
                path: `/${endpointName}/${endpointPrivacyPurge}`,
                methods: [apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointPrivacyPurge}`,
                methods: [apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointInteractionHistory}/${pathParameterId}`,
                methods: [apiGatewayV2.HttpMethod.GET]
            },
            {
                path: `/${endpointName}/${endpointCache}`,
                methods: [apiGatewayV2.HttpMethod.POST]
            },
            {
                path: `/${endpointName}/${endpointPromptConfig}`,
                methods: [apiGatewayV2.HttpMethod.GET, apiGatewayV2.HttpMethod.POST]
            }
        ];

        const authorizer = new apiAuthorizer.HttpUserPoolAuthorizer('ucpPortalUserPoolAuthorizer', this.userPool, {
            userPoolClients: [this.userPoolClient]
        });
        const api = new apiGatewayV2.HttpApi(this.stack, 'apiGatewayV2', {
            apiName: 'ucpBackEnd' + this.envName,
            corsPreflight: {
                allowOrigins: ['*'],
                allowMethods: [
                    apiGatewayV2.CorsHttpMethod.OPTIONS,
                    apiGatewayV2.CorsHttpMethod.GET,
                    apiGatewayV2.CorsHttpMethod.POST,
                    apiGatewayV2.CorsHttpMethod.PUT,
                    apiGatewayV2.CorsHttpMethod.DELETE
                ],
                allowHeaders: ['*']
            }
        });
        const apiLambdaIntegration = new apiIntegrations.HttpLambdaIntegration('UCPBackendLambdaIntegration', this.backendLambdaFunction, {
            payloadFormatVersion: apiGatewayV2.PayloadFormatVersion.VERSION_1_0
        });

        // Creating API Gateway routes in batches of 5 as too many route create requests at the same time will throttle the api
        let routeCount = 1;
        let lastRouteSet: apiGatewayV2.HttpRoute[];
        let currentRouteSet: apiGatewayV2.HttpRoute[] = [];
        for (const { path, methods } of apiRoutes) {
            const currentRoute = api.addRoutes({
                path,
                authorizer: authorizer,
                methods,
                integration: apiLambdaIntegration
            });
            if (lastRouteSet!) {
                for (const routeDep of lastRouteSet) {
                    currentRoute.forEach(route => route.node.addDependency(routeDep));
                }
            }
            for (const currRoute of currentRoute) {
                currentRouteSet.push(currRoute);
            }
            if (routeCount % 2 === 0) {
                lastRouteSet = currentRouteSet;
                currentRouteSet = [];
            }
            routeCount++;
        }

        const stageName = 'api';
        const stage = new apiGatewayV2.HttpStage(this.stack, 'apiGarewayv2Stage', {
            httpApi: api,
            stageName,
            autoDeploy: true,
            throttle: {
                //average request per second over a long period of time. the higher count in e2e test today is 360 per minute
                rateLimit: 1000,
                //max api rate over a short time (up to a few seconds)
                burstLimit: 5000
            }
        });

        return { api, stage };
    }

    /**
     * Creates a domain key which is to encrypt business objects in S3 buckets.
     * @returns Domain KMS key
     */
    private createDomainKey(): kms.Key {
        return new kms.Key(this.stack, 'ucpDomainKey', {
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            pendingWindow: cdk.Duration.days(20),
            alias: 'alias/ucpDomainKey' + this.envName,
            description: 'KMS key for encrypting business object S3 buckets',
            enableKeyRotation: true
        });
    }

    /**
     * Creates a backend Lambda function.
     * @returns Backend Lambda function
     */
    private createBackendLambdaFunction(): lambda.Function {
        const ucpBackEndLambdaPrefix = 'ucpBackEnd';
        const backendLambdaFunction = new GoFunction(this.stack, 'ucpBackEnd' + this.envName, {
            entry: path.join(__dirname, '..', '..', '..', 'ucp-backend', 'src', 'main', 'main.go'),
            bundling: {
                environment: { GOOS: 'linux', GOARCH: 'arm64', GOWORK: 'off' },
                commandHooks: {
                    afterBundling: (inputDir: string, outputDir: string): string[] => [
                        `mkdir -p ${outputDir}/tah-common-glue-schemas`,
                        `cp ${inputDir}/source/tah-common/tah-common-glue-schemas/* ${outputDir}/tah-common-glue-schemas/`
                    ],
                    beforeBundling: (): string[] => []
                }
            },
            functionName: ucpBackEndLambdaPrefix + this.envName,
            runtime: lambda.Runtime.PROVIDED_AL2,
            architecture: lambda.Architecture.ARM_64,
            environment: {
                LAMBDA_ACCOUNT_ID: cdk.Aws.ACCOUNT_ID,
                LAMBDA_ENV: this.envName,
                ATHENA_DB: '',
                ATHENA_WORKGROUP: '',
                GLUE_DB: this.props.glueDatabase.databaseName,
                DATALAKE_ADMIN_ROLE_ARN: this.props.dataLakeAdminRole.roleArn,
                // Batch processing
                AIR_BOOKING_JOB_NAME_CUSTOMER: this.props.airBookingBatch.glueJob.jobName,
                CLICKSTREAM_JOB_NAME_CUSTOMER: this.props.clickStreamBatch.glueJob.jobName,
                CSI_JOB_NAME_CUSTOMER: this.props.customerServiceInteractionBatch.glueJob.jobName,
                GUEST_PROFILE_JOB_NAME_CUSTOMER: this.props.guestProfileBatch.glueJob.jobName,
                HOTEL_BOOKING_JOB_NAME_CUSTOMER: this.props.hotelBookingBatch.glueJob.jobName,
                HOTEL_STAY_JOB_NAME_CUSTOMER: this.props.hotelStayBatch.glueJob.jobName,
                PAX_PROFILE_JOB_NAME_CUSTOMER: this.props.paxProfileBatch.glueJob.jobName,
                S3_AIR_BOOKING: this.props.airBookingBatch.bucket.bucketName,
                S3_CLICKSTREAM: this.props.clickStreamBatch.bucket.bucketName,
                S3_CSI: this.props.customerServiceInteractionBatch.bucket.bucketName,
                S3_GUEST_PROFILE: this.props.guestProfileBatch.bucket.bucketName,
                S3_HOTEL_BOOKING: this.props.hotelBookingBatch.bucket.bucketName,
                S3_PAX_PROFILE: this.props.paxProfileBatch.bucket.bucketName,
                S3_STAY_REVENUE: this.props.hotelStayBatch.bucket.bucketName,
                // DynamoDB tables
                CONFIG_TABLE_NAME: this.props.configTable.tableName,
                CONFIG_TABLE_PK: DynamoDbTableKey.CONFIG_TABLE_PK,
                CONFIG_TABLE_SK: DynamoDbTableKey.CONFIG_TABLE_SK,
                ERROR_TABLE_NAME: this.props.errorTable.tableName,
                ERROR_TABLE_PK: DynamoDbTableKey.ERROR_TABLE_PK,
                ERROR_TABLE_SK: DynamoDbTableKey.ERROR_TABLE_SK,
                PORTAL_CONFIG_TABLE_NAME: this.props.portalConfigTable.tableName,
                PORTAL_CONFIG_TABLE_PK: DynamoDbTableKey.PORTAL_CONFIG_TABLE_PK,
                PORTAL_CONFIG_TABLE_SK: DynamoDbTableKey.PORTAL_CONFIG_TABLE_SK,
                DYNAMO_TABLE_MATCH: this.props.matchTable.tableName,
                DYNAMO_TABLE_MATCH_PK: DynamoDbTableKey.MATCH_TABLE_PK,
                DYNAMO_TABLE_MATCH_SK: DynamoDbTableKey.MATCH_TABLE_SK,
                PRIVACY_RESULTS_TABLE_NAME: this.props.privacySearchResultsTable.tableName,
                PRIVACY_RESULTS_TABLE_PK: DynamoDbTableKey.PRIVACY_SEARCH_RESULTS_TABLE_PK,
                PRIVACY_RESULTS_TABLE_SK: DynamoDbTableKey.PRIVACY_SEARCH_RESULTS_TABLE_SK,

                S3_OUTPUT: this.props.outputBucket.bucketName,
                CONNECT_PROFILE_SOURCE_BUCKET: this.props.connectProfileImportBucket.bucketName,
                KMS_KEY_PROFILE_DOMAIN: this.domainKey.keyArn,
                SYNC_LAMBDA_NAME: this.props.syncLambdaFunction.functionName,
                ACCP_DOMAIN_DLQ: this.props.connectProfileDomainErrorQueue.queueUrl,
                COGNITO_USER_POOL_ID: this.userPool.userPoolId,
                MATCH_BUCKET_NAME: this.props.matchBucket.bucketName,
                ASYNC_LAMBDA_NAME: this.asyncLambdaFunction.functionName,
                RETRY_LAMBDA_NAME: this.props.errorRetryLambdaFunction.functionName,
                IS_PLACEHOLDER_CONNECTOR_BUCKET: this.props.isPlaceholderConnectorBucket,
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
                CHANGE_PROC_KINESIS_STREAM_NAME: this.props.changeProcessorKinesisStream.streamName,
                SSM_PARAM_NAMESPACE: this.props.ssmParamNamespace
            },
            tracing: lambda.Tracing.ACTIVE,
            timeout: cdk.Duration.seconds(60),
            // Low cost storage security config
            vpc: this.props.uptVpc,
            securityGroups: [this.props.profileStorageOutput.lambdaToProxyGroup]
        });
        backendLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    'appflow:CreateFlow',
                    'appflow:StartFlow',
                    'appflow:DescribeFlow',
                    'appflow:DeleteFlow',
                    'cognito-idp:GetGroup',
                    'cognito-idp:AdminListGroupsForUser',
                    'cognito-idp:CreateGroup',
                    'datasync:CreateLocationS3',
                    'datasync:CreateTask',
                    'datasync:StartTaskExecution',
                    'logs:DescribeLogGroups',
                    'glue:CreateTable',
                    'glue:DeleteTable',
                    'glue:GetTags',
                    'glue:TagResource',
                    'glue:GetJob',
                    'glue:GetJobRuns',
                    'glue:UpdateJob',
                    'kms:GenerateDataKey',
                    'kms:Decrypt',
                    'kms:CreateGrant',
                    'kms:ListGrants',
                    'kms:ListAliases',
                    'kms:DescribeKey',
                    'kms:ListKeys',
                    'profile:SearchProfiles',
                    'profile:GetDomain',
                    'profile:CreateDomain',
                    'profile:ListDomains',
                    'profile:ListIntegrations',
                    'profile:DeleteDomain',
                    'profile:DeleteProfile',
                    'profile:PutIntegration',
                    'profile:PutProfileObjectType',
                    'profile:DeleteProfileObjectType',
                    'profile:GetMatches',
                    'profile:ListProfileObjects',
                    'profile:ListProfileObjectTypes',
                    'profile:GetProfileObjectType',
                    'profile:TagResource',
                    'profile:MergeProfiles',
                    'appflow:DescribeFlow',
                    // This is needed for industry connector integration.
                    'servicecatalog:ListApplications',
                    's3:ListBucket',
                    's3:ListAllMyBuckets',
                    's3:GetBucketLocation',
                    's3:GetBucketPolicy',
                    's3:PutBucketPolicy',
                    // This is needed for domain creation
                    'sqs:CreateQueue',
                    'sqs:ReceiveMessage',
                    'sqs:SetQueueAttributes',
                    'sqs:GetQueueAttributes',
                    'sqs:DeleteQueue',
                    // This is needed to be able to empty the table.
                    'dynamodb:GetItem',
                    'dynamodb:Query',
                    'dynamodb:CreateTable',
                    'dynamodb:DeleteTable',
                    'dynamodb:ListTagsOfResource',
                    'dynamodb:TagResource',
                    'dynamodb:DescribeTimeToLive',
                    'dynamodb:UpdateTimeToLive',
                    // This is needed to update the event source mapping in case error table has been emptied.
                    'lambda:ListEventSourceMappings',
                    'bedrock:InvokeModel'
                ],
                effect: iam.Effect.ALLOW,
                resources: ['*']
            })
        );

        // Specific policy for adding DynamoDB stream as event source.
        backendLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['lambda:CreateEventSourceMapping'],
                effect: iam.Effect.ALLOW,
                resources: ['*'] // NOSONAR (typescript:S6317) - Unable to scope this policy to specific ARNs https://docs.aws.amazon.com/lambda/latest/dg/lambda-api-permissions-ref.html
            })
        );

        // Grant permission to the backend Lambda function to trigger the sync Lambda function.
        backendLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['lambda:InvokeFunction'],
                effect: iam.Effect.ALLOW,
                resources: [this.props.syncLambdaFunction.functionArn]
            })
        );

        // Low-cost storage permissions
        // Tables are created/deleted via API and cannot be scoped down
        backendLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    'dynamodb:CreateTable',
                    'dynamodb:DescribeTable',
                    'dynamodb:DeleteTable',
                    'dynamodb:Query',
                    'dynamodb:Scan',
                    'dynamodb:BatchWriteItem'
                ],
                effect: iam.Effect.ALLOW,
                resources: [`arn:${cdk.Aws.PARTITION}:dynamodb:${cdk.Aws.REGION}:${cdk.Aws.ACCOUNT_ID}:table/*`]
            })
        );

        backendLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['secretsmanager:GetSecretValue'],
                effect: iam.Effect.ALLOW,
                resources: [this.props.profileStorageOutput.storageSecretArn]
            })
        );

        backendLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['iam:PassRole'],
                effect: iam.Effect.ALLOW,
                resources: [this.props.industryConnectorDataSyncRole.roleArn]
            })
        );

        this.props.privacySearchResultsTable.grantReadWriteData(backendLambdaFunction);

        backendLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['ssm:GetParameter', 'ssm:GetParameters', 'ssm:GetParametersByPath'],
                resources: [
                    `arn:${cdk.Aws.PARTITION}:ssm:${cdk.Aws.REGION}:${cdk.Aws.ACCOUNT_ID}:parameter${this.props.ssmParamNamespace}*`
                ]
            })
        );

        backendLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['ecs:ListTasks'],
                conditions: { ArnEquals: { 'ecs:cluster': this.props.irCluster.clusterArn } },
                resources: [
                    `arn:${cdk.Aws.PARTITION}:ecs:${cdk.Aws.REGION}:${cdk.Aws.ACCOUNT_ID}:container-instance/${this.props.irCluster.clusterName}/*`
                ]
            })
        );

        return backendLambdaFunction;
    }

    /**
     * Creates an asynchronous Lambda function which handles requests asynchronously.
     * @returns Async Lambda function
     */
    private createAsyncLambdaFunction(): lambda.Function {
        const ucpAsyncLambdaPrefix = 'ucpAsync';
        const asyncLambdaFunction = new GoFunction(this.stack, ucpAsyncLambdaPrefix + this.envName, {
            entry: path.join(__dirname, '..', '..', '..', 'ucp-async', 'src', 'main', 'main.go'),
            bundling: {
                environment: { GOOS: 'linux', GOARCH: 'arm64', GOWORK: 'off' },
                commandHooks: {
                    afterBundling: (inputDir: string, outputDir: string): string[] => [
                        `mkdir -p ${outputDir}/tah-common-glue-schemas`,
                        `cp ${inputDir}/source/tah-common/tah-common-glue-schemas/* ${outputDir}/tah-common-glue-schemas/`
                    ],
                    beforeBundling: (): string[] => []
                }
            },
            functionName: ucpAsyncLambdaPrefix + this.envName,
            runtime: lambda.Runtime.PROVIDED_AL2,
            architecture: lambda.Architecture.ARM_64,
            environment: {
                // General info
                LAMBDA_ACCOUNT_ID: cdk.Aws.ACCOUNT_ID,
                LAMBDA_ENV: this.envName,
                LAMBDA_REGION: cdk.Aws.REGION,
                // Solution metrics
                SEND_ANONYMIZED_DATA: this.props.sendAnonymizedData,
                METRICS_SOLUTION_ID: this.props.solutionId,
                METRICS_SOLUTION_VERSION: this.props.solutionVersion,
                // Config table info
                CONFIG_TABLE_NAME: this.props.configTable.tableName,
                CONFIG_TABLE_PK: DynamoDbTableKey.CONFIG_TABLE_PK,
                CONFIG_TABLE_SK: DynamoDbTableKey.CONFIG_TABLE_SK,
                MATCH_TABLE_NAME: this.props.matchTable.tableName,
                MATCH_TABLE_PK: DynamoDbTableKey.MATCH_TABLE_PK,
                MATCH_TABLE_SK: DynamoDbTableKey.MATCH_TABLE_SK,
                ERROR_TABLE_NAME: this.props.errorTable.tableName,
                ERROR_TABLE_PK: DynamoDbTableKey.ERROR_TABLE_PK,
                ERROR_TABLE_SK: DynamoDbTableKey.ERROR_TABLE_SK,
                PRIVACY_RESULTS_TABLE_NAME: this.props.privacySearchResultsTable.tableName,
                PRIVACY_RESULTS_TABLE_PK: DynamoDbTableKey.PRIVACY_SEARCH_RESULTS_TABLE_PK,
                PRIVACY_RESULTS_TABLE_SK: DynamoDbTableKey.PRIVACY_SEARCH_RESULTS_TABLE_SK,
                PORTAL_CONFIG_TABLE_NAME: this.props.portalConfigTable.tableName,
                PORTAL_CONFIG_TABLE_PK: DynamoDbTableKey.PORTAL_CONFIG_TABLE_PK,
                PORTAL_CONFIG_TABLE_SK: DynamoDbTableKey.PORTAL_CONFIG_TABLE_SK,
                RETRY_LAMBDA_NAME: this.props.errorRetryLambdaFunction.functionName,
                // Cognito user pool
                COGNITO_USER_POOL_ID: this.userPool.userPoolId,
                // Glue db
                GLUE_DB: this.props.glueDatabase.databaseName,
                S3_EXPORT_BUCKET: this.props.glueTravelerTable.bucket.bucketName,
                GLUE_EXPORT_TABLE_NAME: this.props.glueTravelerTable.tableName,
                TRAVELER_S3_ROOT_PATH: this.props.travelerS3RootPath,
                ATHENA_WORKGROUP_NAME: this.props.athenaWorkGroup.name,
                ATHENA_OUTPUT_BUCKET: this.props.athenaResultsBucket.bucketName,
                GET_S3_PATHS_PREPARED_STATEMENT_NAME: this.props.athenaSearchProfileS3PathsPreparedStatement.statementName,
                GET_S3PATH_MAPPING_PREPARED_STATEMENT_NAME: this.props.athenaGetS3PathsByConnectIdsPreparedStatement.statementName,
                AURORA_PROXY_ENDPOINT: this.props.profileStorageOutput.storageProxyEndpoint,
                AURORA_DB_NAME: this.props.profileStorageOutput.storageDbName,
                AURORA_DB_SECRET_ARN: this.props.profileStorageOutput.storageSecretArn,
                STORAGE_CONFIG_TABLE_NAME: this.props.profileStorageOutput.storageConfigTable?.tableName ?? '',
                STORAGE_CONFIG_TABLE_PK: this.props.profileStorageOutput.storageConfigTablePk,
                STORAGE_CONFIG_TABLE_SK: this.props.profileStorageOutput.storageConfigTableSk,
                CHANGE_PROC_KINESIS_STREAM_NAME: this.props.changeProcessorKinesisStream.streamName,
                // GDPR
                S3_EXCISE_QUEUE_URL: this.props.s3ExciseQueue.queueUrl,
                GDPR_PURGE_LOG_GROUP_NAME: this.props.gdprPurgeLogGroup.logGroupName,
                SSM_PARAM_NAMESPACE: this.props.ssmParamNamespace
            },
            retryAttempts: 0,
            timeout: cdk.Duration.seconds(900),
            tracing: lambda.Tracing.ACTIVE,
            // Low cost storage security config
            vpc: this.props.uptVpc,
            securityGroups: [this.props.profileStorageOutput.lambdaToProxyGroup]
        });

        // Additional permissions
        asyncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    'appflow:CreateFlow',
                    'appflow:StartFlow',
                    'appflow:DescribeFlow',
                    'appflow:DeleteFlow',
                    'cognito-idp:GetGroup',
                    'cognito-idp:CreateGroup',
                    'cognito-idp:DeleteGroup',
                    'logs:DescribeLogGroups',
                    'glue:CreateTable',
                    'glue:DeleteTable',
                    'glue:GetTags',
                    'glue:TagResource',
                    'iam:CreateServiceLinkedRole',
                    'kinesis:DescribeStreamSummary',
                    'kms:GenerateDataKey',
                    'kms:Decrypt',
                    'kms:CreateGrant',
                    'kms:ListGrants',
                    'kms:ListAliases',
                    'kms:DescribeKey',
                    'kms:ListKeys',
                    'profile:SearchProfiles',
                    'profile:GetDomain',
                    'profile:CreateDomain',
                    'profile:CreateEventStream',
                    'profile:DeleteEventStream',
                    'profile:ListDomains',
                    'profile:ListIntegrations',
                    'profile:DeleteDomain',
                    'profile:DeleteProfile',
                    'profile:PutIntegration',
                    'profile:PutProfileObjectType',
                    'profile:DeleteProfileObjectType',
                    'profile:GetMatches',
                    'profile:ListProfileObjects',
                    'profile:ListProfileObjectTypes',
                    'profile:GetProfileObjectType',
                    'profile:TagResource',
                    'profile:PutProfileObject',
                    'profile:MergeProfiles',
                    'appflow:DescribeFlow',
                    's3:ListBucket',
                    's3:ListAllMyBuckets',
                    's3:GetBucketLocation',
                    's3:GetBucketPolicy',
                    's3:PutBucketPolicy',
                    'sqs:CreateQueue',
                    'sqs:ReceiveMessage',
                    'sqs:SetQueueAttributes',
                    'sqs:GetQueueAttributes',
                    'sqs:DeleteQueue',
                    'secretsmanager:GetSecretValue',
                    'lambda:ListEventSourceMappings',
                    'lambda:GetFunction',
                    'lambda:ListTags'
                ],
                effect: iam.Effect.ALLOW,
                resources: ['*']
            })
        );

        asyncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    'dynamodb:CreateTable',
                    'dynamodb:DeleteTable',
                    'dynamodb:DescribeTable',
                    'dynamodb:ListTagsOfResource',
                    'dynamodb:TagResource',
                    'dynamodb:PutItem',
                    'dynamodb:BatchWriteItem',
                    'dynamodb:GetItem',
                    'dynamodb:DeleteItem',
                    'dynamodb:Query',
                    'dynamodb:DescribeTimeToLive',
                    'dynamodb:UpdateTimeToLive',
                    'dynamodb:PutItem'
                ],
                resources: [`arn:${cdk.Aws.PARTITION}:dynamodb:${cdk.Aws.REGION}:${cdk.Aws.ACCOUNT_ID}:table/*`]
            })
        );

        asyncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['ssm:GetParameter', 'ssm:GetParameters', 'ssm:GetParametersByPath'],
                resources: [
                    `arn:${cdk.Aws.PARTITION}:ssm:${cdk.Aws.REGION}:${cdk.Aws.ACCOUNT_ID}:parameter${this.props.ssmParamNamespace}*`
                ]
            })
        );

        this.props.s3ExciseQueue.grantSendMessages(asyncLambdaFunction);
        this.props.privacySearchResultsTable.grantReadWriteData(asyncLambdaFunction);

        // Specific policy for adding DynamoDB stream as event source
        asyncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['lambda:CreateEventSourceMapping'],
                effect: iam.Effect.ALLOW,
                resources: ['*'] //NOSONAR (typescript:S6317) - Unable to scope this policy to specific ARNs https://docs.aws.amazon.com/lambda/latest/dg/lambda-api-permissions-ref.html
            })
        );

        asyncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['iam:PutRolePolicy'],
                effect: iam.Effect.ALLOW,
                resources: [
                    this.stack.formatArn({
                        arnFormat: cdk.ArnFormat.SLASH_RESOURCE_NAME,
                        region: '',
                        resource: 'role',
                        resourceName: 'aws-service-role/profile.amazonaws.com/AWSServiceRoleForProfile_*',
                        service: 'iam'
                    })
                ]
            })
        );
        asyncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    'athena:GetPreparedStatement',
                    'athena:GetQueryExecution',
                    'athena:StartQueryExecution',
                    'athena:GetQueryResults'
                ],
                effect: iam.Effect.ALLOW,
                resources: [
                    this.stack.formatArn({
                        arnFormat: cdk.ArnFormat.SLASH_RESOURCE_NAME,
                        service: 'athena',
                        resource: 'workgroup',
                        resourceName: this.props.athenaWorkGroup.name
                    })
                ]
            })
        );

        asyncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['dynamodb:DeleteTable'],
                effect: iam.Effect.ALLOW,
                resources: [this.props.errorTable.tableArn]
            })
        );

        asyncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    // Creating a partition on glue table for each domain
                    'glue:BatchCreatePartition',
                    'glue:BatchDeletePartition',
                    'glue:CreatePartition',
                    'glue:GetTable',
                    'glue:GetPartition',
                    'glue:GetPartitions'
                ],
                effect: iam.Effect.ALLOW,
                resources: [this.props.glueDatabase.catalogArn, this.props.glueDatabase.databaseArn, this.props.glueTravelerTable.tableArn]
            })
        );

        asyncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: [
                    's3:GetBucketLocation',
                    's3:GetObject',
                    's3:ListBucket',
                    's3:ListBucketMultipartUploads',
                    's3:AbortMultipartUpload',
                    's3:PutObject',
                    's3:ListMultipartUploadParts'
                ],
                effect: iam.Effect.ALLOW,
                resources: [this.props.athenaResultsBucket.bucketArn, `${this.props.athenaResultsBucket.bucketArn}/*`]
            })
        );

        // Permission to send change events to Kinesis (e.g. when merging profiles)
        asyncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['kinesis:PutRecords'],
                effect: iam.Effect.ALLOW,
                resources: [this.props.changeProcessorKinesisStream.streamArn]
            })
        );

        this.props.outputBucket.grantRead(asyncLambdaFunction);
        asyncLambdaFunction.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['secretsmanager:GetSecretValue'],
                effect: iam.Effect.ALLOW,
                resources: [this.props.profileStorageOutput.storageSecretArn]
            })
        );

        return asyncLambdaFunction;
    }
}
