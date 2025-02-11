// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Match, Template } from 'aws-cdk-lib/assertions';
import { Cluster } from 'aws-cdk-lib/aws-ecs';
import { Backend } from '../lib/backend';
import {
    artifactBucketPath,
    createAthenaPreparedStatement,
    createAthenaWorkGroup,
    createBatch,
    createDynamoDbTable,
    createGlueDatabase,
    createGlueTable,
    createIamRole,
    createKinesisStream,
    createLambdaFunction,
    createLogGroup,
    createProfileStorageOutput,
    createS3Bucket,
    createSqsQueue,
    createTAHS3Bucket,
    createVpc,
    envName,
    sendAnonymizedData,
    solutionId,
    solutionVersion
} from './mock';

test('Backend', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');
    const configTable = createDynamoDbTable(stack, 'configTable', false);
    const connectProfileDomainErrorQueue = createSqsQueue(stack, 'connectProfileDomainErrorQueue');
    const connectProfileImportBucket = createS3Bucket(stack, 'connectProfileImportBucket');
    const dataLakeAdminRole = createIamRole(stack, 'dataLakeAdminRole');
    const errorRetryLambdaFunction = createLambdaFunction(stack, 'errorRetryLambdaFunction');
    const errorTable = createDynamoDbTable(stack, 'errorTable', false);
    const glueDatabase = createGlueDatabase(stack, 'glueDatabase');
    const glueTravelerTable = createGlueTable(stack, 'glueTravelerTable', glueDatabase);
    const industryConnectorDataSyncRole = createIamRole(stack, 'industryConnectorDataSyncRole');
    const isPlaceholderConnectorBucket = 'true';
    const matchBucket = createS3Bucket(stack, 'matchBucket');
    const matchTable = createDynamoDbTable(stack, 'matchTable', false);
    const outputBucket = createS3Bucket(stack, 'outputBucket');
    const portalConfigTable = createDynamoDbTable(stack, 'portalConfigTable', false);
    const stackLogicalId = 'stack';
    const syncLambdaFunction = createLambdaFunction(stack, 'syncLambdaFunction');
    const travelerS3RootPath = 'traveler';
    const airBookingBatch = createBatch(stack, 'airBookingBatch');
    const clickStreamBatch = createBatch(stack, 'clickStreamBatch');
    const customerServiceInteractionBatch = createBatch(stack, 'customerServiceInteractionBatch');
    const guestProfileBatch = createBatch(stack, 'guestProfileBatch');
    const hotelBookingBatch = createBatch(stack, 'hotelBookingBatch');
    const hotelStayBatch = createBatch(stack, 'hotelStayBatch');
    const paxProfileBatch = createBatch(stack, 'paxProfileBatch');
    const artifactBucket = createS3Bucket(stack, 'artifactBucket');
    const athenaResultsBucket = createTAHS3Bucket(stack, 'athenaResultsBucket');
    const athenaWorkGroup = createAthenaWorkGroup(stack, 'athenaWorkGroup');
    const athenaSearchProfileS3PathsPreparedStatement = createAthenaPreparedStatement(stack, 'athenaSearchProfileS3PathsPreparedStatement');
    const athenaGetS3PathsByConnectIdsPreparedStatement = createAthenaPreparedStatement(
        stack,
        'athenaGetS3PathsByConnectIdsPreparedStatement'
    );
    const privacySearchResultsTable = createDynamoDbTable(stack, 'privacyResultsTable', false);
    const vpc = createVpc(stack, 'vpc');
    const profileStorageOutput = createProfileStorageOutput(stack, 'storageOutput', vpc);
    const changeProcessorKinesisStream = createKinesisStream(stack, 'changeProcessorKinesisStream');
    const s3ExciseQueue = createSqsQueue(stack, 's3ExciseQueue');
    const gdprPurgeLogGroup = createLogGroup(stack, 'gdprPurgeLogGroup');
    const irCluster = new Cluster(stack, 'IRCluster');

    // Call
    new Backend(
        { envName, stack },
        {
            configTable,
            connectProfileDomainErrorQueue,
            connectProfileImportBucket,
            dataLakeAdminRole,
            errorRetryLambdaFunction,
            errorTable,
            glueDatabase,
            glueTravelerTable,
            industryConnectorDataSyncRole,
            isPlaceholderConnectorBucket,
            matchBucket,
            matchTable,
            outputBucket,
            portalConfigTable,
            stackLogicalId,
            syncLambdaFunction,
            travelerS3RootPath,
            airBookingBatch,
            clickStreamBatch,
            customerServiceInteractionBatch,
            guestProfileBatch,
            hotelBookingBatch,
            hotelStayBatch,
            paxProfileBatch,
            profileStorageOutput,
            changeProcessorKinesisStream,
            // LambdaProps
            sendAnonymizedData,
            artifactBucket,
            artifactBucketPath,
            solutionId,
            solutionVersion,
            athenaResultsBucket,
            athenaWorkGroup,
            privacySearchResultsTable,
            athenaSearchProfileS3PathsPreparedStatement,
            athenaGetS3PathsByConnectIdsPreparedStatement,
            s3ExciseQueue,
            gdprPurgeLogGroup,
            ssmParamNamespace: `/uptprod/`,
            irCluster,
            uptVpc: vpc
        }
    );

    // Verify
    const template = Template.fromStack(stack);
    template.hasResourceProperties('AWS::Cognito::UserPool', {
        UserPoolName: 'ucpUserpool' + envName
    });
    template.hasResourceProperties('AWS::Cognito::UserPoolClient', {
        ClientName: 'ucpPortal'
    });
    template.hasResourceProperties('AWS::Cognito::UserPoolDomain', {
        Domain: {
            'Fn::Join': [
                '-',
                [
                    'ucp-domain',
                    {
                        Ref: 'AWS::AccountId'
                    },
                    envName
                ]
            ]
        }
    });
    template.hasResourceProperties('AWS::KMS::Alias', {
        AliasName: 'alias/ucpDomainKey' + envName
    });
    template.hasResourceProperties('AWS::Lambda::Function', {
        FunctionName: 'ucpBackEnd' + envName
    });
    template.hasResourceProperties('AWS::Lambda::Function', {
        FunctionName: 'ucpAsync' + envName
    });
    template.hasResourceProperties('AWS::ApiGatewayV2::Api', {
        Name: 'ucpBackEnd' + envName,
        ProtocolType: 'HTTP'
    });
    template.hasResourceProperties('AWS::ApiGatewayV2::Integration', {
        IntegrationType: 'AWS_PROXY'
    });
    template.hasResourceProperties('AWS::ApiGatewayV2::Authorizer', {
        AuthorizerType: 'JWT',
        IdentitySource: ['$request.header.Authorization'],
        JwtConfiguration: {
            Audience: [{ Ref: Match.stringLikeRegexp('UserPoolClient') }],
            Issuer: {
                'Fn::Join': [
                    '',
                    ['https://cognito-idp.', { Ref: 'AWS::Region' }, '.amazonaws.com/', { Ref: Match.stringLikeRegexp('Userpool') }]
                ]
            }
        },
        Name: 'ucpPortalUserPoolAuthorizer'
    });

    const routes = [
        'GET /ucp/admin',
        'POST /ucp/admin',
        'DELETE /ucp/admin/{id}',
        'GET /ucp/admin/{id}',
        'PUT /ucp/admin/{id}',
        'GET /ucp/async',
        'GET /ucp/error',
        'DELETE /ucp/error/{id}',
        'GET /ucp/error/{id}',
        'GET /ucp/flows',
        'POST /ucp/flows',
        'GET /ucp/connector',
        'POST /ucp/connector/link',
        'GET /ucp/jobs',
        'POST /ucp/jobs',
        'GET /ucp/merge',
        'POST /ucp/merge',
        'POST /ucp/unmerge',
        'GET /ucp/merge/{id}',
        'GET /ucp/portalConfig',
        'POST /ucp/portalConfig',
        'GET /ucp/profile',
        'DELETE /ucp/profile/{id}',
        'GET /ucp/profile/{id}',
        'GET /ucp/ruleSet',
        'POST /ucp/ruleSet',
        'POST /ucp/ruleSet/activate',
        'GET /ucp/ruleSetCache',
        'POST /ucp/ruleSetCache',
        'POST /ucp/ruleSetCache/activate',
        'GET /ucp/privacy',
        'GET /ucp/privacy/{id}',
        'POST /ucp/privacy',
        'DELETE /ucp/privacy',
        'POST /ucp/privacy/purge',
        'GET /ucp/privacy/purge',
        'GET /ucp/interactionHistory/{id}',
        'POST /ucp/cache',
        'GET /ucp/promptConfig',
        'POST /ucp/promptConfig',
        'GET /ucp/profile/summary/{id}'
    ];
    template.resourceCountIs('AWS::ApiGatewayV2::Route', routes.length);

    for (const routeKey of routes) {
        template.hasResourceProperties('AWS::ApiGatewayV2::Route', {
            RouteKey: routeKey
        });
    }
});
