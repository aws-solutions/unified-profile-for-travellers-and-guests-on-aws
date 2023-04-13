// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { Stack, CfnOutput, RemovalPolicy, StackProps, Duration, Tags } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { KinesisStreamsToLambda } from '@aws-solutions-constructs/aws-kinesisstreams-lambda';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as glue from 'aws-cdk-lib/aws-glue';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as kms from 'aws-cdk-lib/aws-kms';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import * as athena from 'aws-cdk-lib/aws-athena';
import * as cognito from 'aws-cdk-lib/aws-cognito';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as aws_events from 'aws-cdk-lib/aws-events';
import * as aws_events_targets from 'aws-cdk-lib/aws-events-targets';
import * as lambda_event_sources from 'aws-cdk-lib/aws-lambda-event-sources';


import { CorsHttpMethod, HttpApi, HttpMethod, HttpRoute, HttpStage, PayloadFormatVersion } from '@aws-cdk/aws-apigatewayv2-alpha';
import { HttpUserPoolAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha';
import { HttpLambdaIntegration, } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { Database, Table, DataFormat, Schema } from '@aws-cdk/aws-glue-alpha';
import { CfnCrawler, CfnJob, CfnTrigger, CfnWorkflow } from 'aws-cdk-lib/aws-glue';
import { Queue, IQueue } from 'aws-cdk-lib/aws-sqs';
import * as tah_s3 from '../tah-cdk-common/s3';
import * as tah_core from '../tah-cdk-common/core';
import { BusinessObjectPipelineOutput, GlueSchema } from "./model"

//importing object schemas
import glueSchemaStay from '../tah-common-glue-schemas/hotel_stay_revenue.glue.json';
import glueSchemaHotelBooking from '../tah-common-glue-schemas/hotel_booking.glue.json';
import glueSchemaAirBooking from '../tah-common-glue-schemas/air_booking.glue.json';
import glueSchemaPaxProfile from '../tah-common-glue-schemas/pax_profile.glue.json';
import glueSchemaClickEvent from '../tah-common-glue-schemas/clickevent.glue.json';
import glueSchemaGuestProfile from '../tah-common-glue-schemas/guest_profile.glue.json';

import * as kinesis from 'aws-cdk-lib/aws-kinesis';

/////////////////////////////////////////////////////////////////////
//Infrastructure Script for the AWS DynamoDB Blog Demo
//This script takes the name of the Environement to spawn and create 
// A lambda function, A DynamoDB table and a set of Api Gateway enpoints.
/////////////////////////////////////////////////////////////////////
export class UCPInfraStack extends Stack {

  constructor(scope: Construct, id: string, props?: StackProps) {

    super(scope, id, props);

    const envName = this.node.tryGetContext("envName");
    const artifactBucket = this.node.tryGetContext("artifactBucket");
    const region = (props && props.env) ? props.env.region : ""
    const account = (props && props.env) ? props.env.account : ""
    if (!envName) {
      throw new Error('No environemnt name provided for stack');
    }
    if (!artifactBucket) {
      throw new Error('No bucket name provided for stack');
    }
    if (!account) {
      throw new Error('No account ID provided for stack');
    }

    //////////////////////
    // Datalake admin Role
    /////////////////////
    const datalakeAdminRole = new iam.Role(this, "ucp-data-admin-role-" + envName, {
      assumedBy: new iam.ServicePrincipal("glue.amazonaws.com"),
      description: "Glue role for UCP data",
      roleName: "ucp-data-admin-role-" + envName
    })
    new CfnOutput(this, 'ucpDataAdminRoleArn', {
      value: datalakeAdminRole.roleArn
    });
    new CfnOutput(this, 'ucpDataAdminRoleName', {
      value: datalakeAdminRole.roleName
    });
    //we need to gran the data admin (which will be the rol for all glue jobs) access to the artifact bucket
    //since the python scripts are stored there
    const artifactsBucket = s3.Bucket.fromBucketName(this, 'ucpArtifactBucket', artifactBucket);
    artifactsBucket.grantReadWrite(datalakeAdminRole)
    //granting general logging
    datalakeAdminRole.addManagedPolicy(iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSGlueServiceRole"))
    datalakeAdminRole.addToPolicy(new iam.PolicyStatement({
      resources: ["arn:aws:logs:" + this.region + ":" + this.account + ":log-group:/*"],
      actions: ["logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogStreams"]
    }))

    datalakeAdminRole.addToPolicy(new iam.PolicyStatement({
      resources: [datalakeAdminRole.roleArn],
      actions: ["iam:PassRole"]
    }))
    const accessLogBucket = new tah_s3.AccessLogBucket(this, "ucp-access-logging")

    /*************************
     * Datalake 3rd party users
     ********************/
    //Amperity
    //TODO: move to cross account role when Amperity supports it
    let amperityUser = new iam.User(this, "ucp3pUserAmperity", {
      userName: "ucp3pUserAmperity" + envName
    })

    /**************
     * SQS Queues
     ********************/
    const connectorCrawlerQueue = new Queue(this, "ucp-connector-crawler-queue-" + envName)
    const connectorCrawlerDlq = new Queue(this, "ucp-connector-crawler-dlq-" + envName)
    const accpDomainErrorQueue = new Queue(this, "ucp-acc-domain-errors-" + envName)

    new CfnOutput(this, 'accpDomainErrorQueue', { value: accpDomainErrorQueue.queueUrl });

    /**************
     * Glue Database
     ********************/
    const glueDb = new Database(this, "ucp-data-glue-database-" + envName, {
      databaseName: "ucp_db_" + envName
    })

    new CfnOutput(this, 'glueDBArn', {
      value: glueDb.databaseArn
    });
    new CfnOutput(this, 'glueDBname', {
      value: glueDb.databaseName
    });


    ////////////////////
    //Athena Workgroup
    ////////////////////////
    const athenaResultsBucket = new tah_s3.Bucket(this, "ucp-athena-output", accessLogBucket)
    let athenaWorkgroupName = "ucp-athena-workgroup-" + envName
    let athenaWorkgroup = new athena.CfnWorkGroup(this, "ucp-athena-workgroup", {
      name: athenaWorkgroupName,
      workGroupConfiguration: {
        publishCloudWatchMetricsEnabled: true,
        resultConfiguration: {
          encryptionConfiguration: {
            encryptionOption: "SSE_S3"
          },
          outputLocation: "s3://" + athenaResultsBucket.bucketName
        }
      }
    })

    /***************************
   * Data Buckets for temporary processing
   *****************************/
    //Target Bucket for Amazon connect profile import
    const connectProfileImportBucket = new tah_s3.Bucket(this, "connectProfileImportBucket", accessLogBucket)
    const connectProfileImportBucketTest = new tah_s3.Bucket(this, "connectProfileImportBucketTest", accessLogBucket)
    //temp bucket for Amazon connect profile identity resolution matches
    const idResolution = new tah_s3.Bucket(this, "ucp-connect-id-resolution-temp", accessLogBucket)

    let hotelBookingOutput = this.buildBusinessObjectPipeline("hotel-booking", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)
    let airBookingOutput = this.buildBusinessObjectPipeline("air_booking", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)
    let guestProfileOutput = this.buildBusinessObjectPipeline("guest-profile", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)
    let paxProfileOutput = this.buildBusinessObjectPipeline("pax-profile", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)
    let clickstreamOutput = this.buildBusinessObjectPipeline("clickstream", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)
    let hotelStayOutput = this.buildBusinessObjectPipeline("hotel-stay", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)

    //Target Bucket for Amazon connect profile export
    let connectProfileExportBucket = new tah_s3.Bucket(this, "ucp-mazon-connect-profile-export", accessLogBucket);
    let amperityImportBucket = new tah_s3.Bucket(this, "ucp-amperity-import", accessLogBucket);
    let amperityExportBucket = new tah_s3.Bucket(this, "ucp-amperity-export", accessLogBucket);

    /*****************************
     * Bucket Permission
     **************************/
    connectProfileImportBucket.grantReadWrite(datalakeAdminRole)
    connectProfileImportBucketTest.grantReadWrite(datalakeAdminRole)
    connectProfileExportBucket.grantReadWrite(datalakeAdminRole)
    amperityExportBucket.grantReadWrite(datalakeAdminRole)

    //Special Permissions for Integration Business Objects
    connectProfileExportBucket.addToResourcePolicy(new iam.PolicyStatement({
      resources: [connectProfileExportBucket.arnForObjects("*"), connectProfileExportBucket.bucketArn],
      actions: ["s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy"],
      principals: [new iam.ServicePrincipal("appflow.amazonaws.com")]
    }))

    connectProfileImportBucket.addToResourcePolicy(new iam.PolicyStatement({
      resources: [connectProfileImportBucket.arnForObjects("*"), connectProfileImportBucket.bucketArn],
      actions: ["s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy"],
      principals: [new iam.ServicePrincipal("appflow.amazonaws.com")]
    }))

    //Amperity
    amperityUser.addToPolicy(new iam.PolicyStatement({
      resources: ["arn:aws:s3:::" + amperityImportBucket.bucketName + "*"],
      actions: ["s3:*"]
    }))

    //Amperity job
    let amperityImportJob = this.job("ucp-amperity-import", envName, artifactBucket, "connectProfileToAmperity", glueDb, datalakeAdminRole, new Map([
      ["SOURCE_TABLE", connectProfileExportBucket.toAthenaTable()],
      ["DEST_BUCKET", amperityImportBucket.bucketName]
    ]))
    let amperityExportJob = this.job("ucp-amperity-export", envName, artifactBucket, "amperityToMatches", glueDb, datalakeAdminRole, new Map([
      ["SOURCE_TABLE", amperityExportBucket.toAthenaTable()],
      ["DEST_DYNAMO_TABLE", "to_be_added"]
    ]))

    /*******************
   * 3- Job Triggers
   ******************/
    this.jobOnDemandTrigger("ucp-amperity-data-import", envName, [amperityImportJob])
    this.jobOnDemandTrigger("ucp-amperity-data-export", envName, [amperityExportJob])

    /////////////////////////
    //Appflow and Amazon Connect Customer profile
    ///////////////////////////
    /*******
     * KMS Key
     */
    //TODO: to rename the key into something moroe explicit
    const kmsKeyProfileDomain = new kms.Key(this, "ucpDomainKey", {
      removalPolicy: RemovalPolicy.DESTROY,
      pendingWindow: Duration.days(20),
      alias: 'alias/ucpDomainKey' + envName,
      description: 'KMS key for encrypting business object S3 buckets',
      enableKeyRotation: true,
    });

    //Customer Profile Outputs for Domain Testing
    new CfnOutput(this, "connectProfileImportBucketOut", { value: connectProfileImportBucket.bucketName })
    new CfnOutput(this, "connectProfileImportBucketTestOut", { value: connectProfileImportBucketTest.bucketName })
    new CfnOutput(this, "connectProfileExportBucket", { value: connectProfileExportBucket.bucketName })
    new CfnOutput(this, "kmsKeyProfileDomain", { value: kmsKeyProfileDomain.keyArn })

    //////////////////////////////////////
    // DYNAMO DB
    /////////////////////////////////////
    const dynamo_pk = "item_id"
    const dynamo_sk = "item_type"
    const configTable = new dynamodb.Table(this, "ucpConfigTable", {
      tableName: "ucp-config-table-" + envName,
      partitionKey: { name: dynamo_pk, type: dynamodb.AttributeType.STRING },
      sortKey: { name: dynamo_sk, type: dynamodb.AttributeType.STRING },
      removalPolicy: RemovalPolicy.DESTROY,
    });
    const dynamo_error_pk = "error_type"
    const dynamo_error_sk = "error_id"
    const errorTable = new dynamodb.Table(this, "ucpErrorTable", {
      tableName: "ucp-error-table-" + envName,
      partitionKey: { name: dynamo_error_pk, type: dynamodb.AttributeType.STRING },
      sortKey: { name: dynamo_error_sk, type: dynamodb.AttributeType.STRING },
      removalPolicy: RemovalPolicy.DESTROY,
    });

    //////////////////////////
    //LAMBDA FUNCTIONS
    /////////////////////////

    //Main application backend
    //////////////////////////////

    const lambdaArtifactRepositoryBucket = s3.Bucket.fromBucketName(this, 'BucketByName', artifactBucket);
    const ucpBackEndLambdaPrefix = 'ucpBackEnd'
    const ucpBackEndLambda = new lambda.Function(this, 'ucpBackEnd' + envName, {
      code: new lambda.S3Code(lambdaArtifactRepositoryBucket, [envName, ucpBackEndLambdaPrefix, 'main.zip'].join("/")),
      functionName: ucpBackEndLambdaPrefix + envName,
      handler: 'main',
      runtime: lambda.Runtime.GO_1_X,
      tracing: lambda.Tracing.ACTIVE,
      timeout: Duration.seconds(60),
      environment: {
        LAMBDA_ACCOUNT_ID: account,
        LAMBDA_ENV: envName,
        ATHENA_WORKGROUP: "",
        ATHENA_DB: "",
        HOTEL_BOOKING_JOB_NAME: hotelBookingOutput.connectorJobName,
        AIR_BOOKING_JOB_NAME: airBookingOutput.connectorJobName,
        GUEST_PROFILE_JOB_NAME: guestProfileOutput.connectorJobName,
        PAX_PROFILE_JOB_NAME: paxProfileOutput.connectorJobName,
        CLICKSTREAM_JOB_NAME: clickstreamOutput.connectorJobName,
        HOTEL_STAY_JOB_NAME: hotelStayOutput.connectorJobName,
        HOTEL_BOOKING_JOB_NAME_CUSTOMER: hotelBookingOutput.customerJobName,
        AIR_BOOKING_JOB_NAME_CUSTOMER: airBookingOutput.customerJobName,
        GUEST_PROFILE_JOB_NAME_CUSTOMER: guestProfileOutput.customerJobName,
        PAX_PROFILE_JOB_NAME_CUSTOMER: paxProfileOutput.customerJobName,
        CLICKSTREAM_JOB_NAME_CUSTOMER: clickstreamOutput.customerJobName,
        HOTEL_STAY_JOB_NAME_CUSTOMER: hotelStayOutput.customerJobName,
        CONNECTOR_CRAWLER_QUEUE: connectorCrawlerQueue.queueArn,
        CONNECTOR_CRAWLER_DLQ: connectorCrawlerDlq.queueArn,
        GLUE_DB: glueDb.databaseName,
        DATALAKE_ADMIN_ROLE_ARN: datalakeAdminRole.roleArn,
        ERROR_TABLE_NAME: errorTable.tableName,
        ERROR_TABLE_PK: dynamo_error_pk,
        ERROR_TABLE_SK: dynamo_error_sk,
        S3_HOTEL_BOOKING: hotelBookingOutput.bucket.bucketName,
        S3_AIR_BOOKING: airBookingOutput.bucket.bucketName,
        S3_GUEST_PROFILE: guestProfileOutput.bucket.bucketName,
        S3_PAX_PROFILE: paxProfileOutput.bucket.bucketName,
        S3_STAY_REVENUE: hotelStayOutput.bucket.bucketName,
        S3_CLICKSTREAM: clickstreamOutput.bucket.bucketName,
        CONNECT_PROFILE_SOURCE_BUCKET: connectProfileImportBucket.bucketName,
        KMS_KEY_PROFILE_DOMAIN: kmsKeyProfileDomain.keyArn,
      }
    });

    hotelBookingOutput.bucket.grantRead(ucpBackEndLambda)
    airBookingOutput.bucket.grantRead(ucpBackEndLambda)
    guestProfileOutput.bucket.grantRead(ucpBackEndLambda)
    paxProfileOutput.bucket.grantRead(ucpBackEndLambda)
    hotelStayOutput.bucket.grantRead(ucpBackEndLambda)
    clickstreamOutput.bucket.grantRead(ucpBackEndLambda)
    hotelStayOutput.bucket.grantRead(ucpBackEndLambda)
    connectProfileImportBucket.grantRead(ucpBackEndLambda)

    errorTable.grantReadWriteData(ucpBackEndLambda)

    ucpBackEndLambda.addToRolePolicy(new iam.PolicyStatement({
      resources: ["*"],
      actions: [
        'appflow:CreateFlow',
        'appflow:StartFlow',
        'appflow:DescribeFlow',
        'appflow:DeleteFlow',
        'glue:CreateCrawler',
        'glue:DeleteCrawler',
        'glue:CreateTrigger',
        'glue:GetTrigger',
        'glue:UpdateTrigger',
        'glue:DeleteTrigger',
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
        // TODO: remove iam actions and set up permission boundary instead
        'iam:AttachRolePolicy',
        'iam:CreatePolicy',
        'iam:CreateRole',
        'iam:DeletePolicy',
        'iam:DeleteRole',
        'iam:DetachRolePolicy',
        'iam:GetPolicy',
        'iam:ListAttachedRolePolicies',
        'iam:PassRole',
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
        'appflow:DescribeFlow',
        'servicecatalog:ListApplications',
        's3:ListBucket',
        's3:ListAllMyBuckets',
        's3:GetBucketLocation',
        's3:GetBucketPolicy',
        's3:PutBucketPolicy',
        'sqs:CreateQueue',
        'sqs:ReceiveMessage',
        'sqs:SetQueueAttributes',
        'sqs:GetQueueAttributes',
        'sqs:DeleteQueue']
    }));


    //Partition sync lambda
    ///////////////////////

    const ucpSyncLambdaPrefix = 'ucpSync'
    const ucpSyncLambda = new lambda.Function(this, 'ucpSync' + envName, {
      code: new lambda.S3Code(lambdaArtifactRepositoryBucket, [envName, ucpSyncLambdaPrefix, 'main.zip'].join("/")),
      functionName: ucpSyncLambdaPrefix + envName,
      handler: 'main',
      runtime: lambda.Runtime.GO_1_X,
      tracing: lambda.Tracing.ACTIVE,
      timeout: Duration.seconds(900),
      environment: {
        LAMBDA_ACCOUNT_ID: account,
        LAMBDA_ENV: envName,
        LAMBDA_REGION: region || "",
        ATHENA_DB: glueDb.databaseName,
        ATHENA_WORKGROUP: athenaWorkgroupName,

        DYNAMO_TABLE: configTable.tableName,
        DYNAMO_PK: dynamo_pk,
        DYNAMO_SK: dynamo_sk,

        S3_HOTEL_BOOKING: hotelBookingOutput.bucket.bucketName,
        S3_AIR_BOOKING: airBookingOutput.bucket.bucketName,
        S3_GUEST_PROFILE: guestProfileOutput.bucket.bucketName,
        S3_PAX_PROFILE: paxProfileOutput.bucket.bucketName,
        S3_STAY_REVENUE: hotelStayOutput.bucket.bucketName,
        S3_CLICKSTREAM: clickstreamOutput.bucket.bucketName,

        HOTEL_BOOKING_JOB_NAME_CUSTOMER: hotelBookingOutput.customerJobName,
        AIR_BOOKING_JOB_NAME_CUSTOMER: airBookingOutput.customerJobName,
        GUEST_PROFILE_JOB_NAME_CUSTOMER: guestProfileOutput.customerJobName,
        PAX_PROFILE_JOB_NAME_CUSTOMER: paxProfileOutput.customerJobName,
        CLICKSTREAM_JOB_NAME_CUSTOMER: clickstreamOutput.customerJobName,
        HOTEL_STAY_JOB_NAME_CUSTOMER: hotelStayOutput.customerJobName,

        HOTEL_BOOKING_DLQ: hotelBookingOutput.errorQueue.queueUrl,
        AIR_BOOKING_DLQ: airBookingOutput.errorQueue.queueUrl,
        GUEST_PROFILE_DLQ: guestProfileOutput.errorQueue.queueUrl,
        PAX_PROFILE_DLQ: paxProfileOutput.errorQueue.queueUrl,
        CLICKSTREAM_DLQ: clickstreamOutput.errorQueue.queueUrl,
        HOTEL_STAY_DLQ: hotelStayOutput.errorQueue.queueUrl,
      }
    });

    const rule = new aws_events.Rule(this, 'ucpSyncJob', {
      schedule: aws_events.Schedule.expression('rate(1 hour)'),
      eventPattern: {}
    });
    rule.addTarget(new aws_events_targets.LambdaFunction(ucpSyncLambda, {}));
    //TODO: narrow this down to the business object tables
    ucpSyncLambda.addToRolePolicy(new iam.PolicyStatement({
      resources: ["*"],
      actions: [
        'athena:StartQueryExecution',
        'athena:GetQueryExecution',
        "athena:GetQueryResults",
        "glue:GetTable",
        "glue:GetDatabase",
        "glue:GetPartition",
        "glue:GetPartitions",
        "glue:BatchCreatePartition",
        "glue:CreatePartition",
      ]
    }));
    //granting permission to read and write on the athena result bucket thus allowing lambda function
    //to successfully execute athena queries using the workgroup created in this stack
    athenaResultsBucket.grantReadWrite(ucpSyncLambda);
    hotelBookingOutput.bucket.grantReadWrite(ucpSyncLambda)
    airBookingOutput.bucket.grantReadWrite(ucpSyncLambda)
    guestProfileOutput.bucket.grantReadWrite(ucpSyncLambda)
    paxProfileOutput.bucket.grantReadWrite(ucpSyncLambda)
    hotelStayOutput.bucket.grantReadWrite(ucpSyncLambda)
    clickstreamOutput.bucket.grantReadWrite(ucpSyncLambda)

    configTable.grantReadWriteData(ucpSyncLambda)

    // Real time flow lambdas
    ////////////////////////

    //The flow here is the following:
    // Kinesis -> Python lmabda -> Kinesis -> Go Lambda -> ACCP
    // we use AWS Solutions Constructs to build this flow from pretested constructs

    const ucpEtlRealTimeLambdaPrefix = "ucpRealTimeTransformer"
    const ucpEtlRealTimeACCPPrefix = "ucpRealTimeTransformerAccp"

    let dlqs: Map<string, Queue> = new Map<string, Queue>();
    dlqs.set("dlgGo", new Queue(this, "ucp-real-time-error-go-" + envName))
    dlqs.set("dldPython", new Queue(this, "ucp-real-time-error-python-" + envName))
    dlqs.set("dlgGoTest", new Queue(this, "ucp-real-time-error-go-test-" + envName))
    dlqs.set("dldPythonTest", new Queue(this, "ucp-real-time-error-python-test-" + envName))

    for (let type of ["", "Test"]) {
      const kinesisLambdaACCP = new KinesisStreamsToLambda(this, ucpEtlRealTimeACCPPrefix + type + envName, {
        kinesisEventSourceProps: {
          startingPosition: lambda.StartingPosition.TRIM_HORIZON,
          batchSize: 10,
          maximumRetryAttempts: 0
        },
        lambdaFunctionProps: {
          runtime: lambda.Runtime.GO_1_X,
          handler: 'main',
          code: new lambda.S3Code(lambdaArtifactRepositoryBucket, [envName, ucpEtlRealTimeACCPPrefix, 'mainAccp.zip'].join("/")),
          deadLetterQueueEnabled: true,
          deadLetterQueue: dlqs.get("dldGo" + type),
          functionName: ucpEtlRealTimeACCPPrefix + type + envName,
          environment: {
            LAMBDA_REGION: region || ""
          }
        }
      });

      const kinesisLambdaStart = new KinesisStreamsToLambda(this, ucpEtlRealTimeLambdaPrefix + type + envName, {
        kinesisEventSourceProps: {
          startingPosition: lambda.StartingPosition.TRIM_HORIZON,
          batchSize: 10,
          maximumRetryAttempts: 0
        },
        lambdaFunctionProps: {
          runtime: lambda.Runtime.PYTHON_3_7,
          handler: 'index.handler',
          code: new lambda.S3Code(lambdaArtifactRepositoryBucket, [envName, ucpEtlRealTimeLambdaPrefix, 'main.zip'].join("/")),
          deadLetterQueueEnabled: true,
          deadLetterQueue: dlqs.get("dldPython" + type),
          functionName: ucpEtlRealTimeLambdaPrefix + type + envName,
          environment: {
            output_stream: kinesisLambdaACCP.kinesisStream.streamName,
            LAMBDA_REGION: region || ""
          }
        }
      });

      //there are 2 cases when we right to DLD.
      //1- if lambda fails for any reason,  the incoming message will automatically be sent to DLQ after the configured numbre of retries
      //2- in cases where we gracefuly fail insude the function, we manually put the message into the queue. this allows better error management
      kinesisLambdaACCP.kinesisStream.grantReadWrite(kinesisLambdaStart.lambdaFunction)
      const pythonDlQ = kinesisLambdaStart.lambdaFunction.deadLetterQueue
      const goDLQ = kinesisLambdaACCP.lambdaFunction.deadLetterQueue
      if (pythonDlQ) {
        pythonDlQ.grantConsumeMessages(kinesisLambdaStart.lambdaFunction)
        pythonDlQ.grantSendMessages(kinesisLambdaStart.lambdaFunction)
        kinesisLambdaStart.lambdaFunction.addEnvironment("DEAD_LETTER_QUEUE_URL", pythonDlQ.queueUrl)
      }
      if (goDLQ) {
        goDLQ.grantConsumeMessages(kinesisLambdaACCP.lambdaFunction)
        goDLQ.grantSendMessages(kinesisLambdaACCP.lambdaFunction)
        kinesisLambdaACCP.lambdaFunction.addEnvironment("DEAD_LETTER_QUEUE_URL", goDLQ.queueUrl)
      }


      kinesisLambdaACCP.lambdaFunction.addToRolePolicy(new iam.PolicyStatement({
        resources: ["*"],
        actions: [
          'profile:PutProfileObject',
          'profile:ListProfileObjects',
          'profile:ListProfileObjectTypes',
          'profile:GetProfileObjectType',
          'profile:PutProfileObjectType',
        ]
      }))



      new CfnOutput(this, "lambdaFunctionNameRealTime" + type, { value: kinesisLambdaStart.lambdaFunction.functionName });
      new CfnOutput(this, "kinesisStreamNameRealTime" + type, { value: kinesisLambdaStart.kinesisStream.streamName });
      new CfnOutput(this, "kinesisStreamOutputNameRealTime" + type, { value: kinesisLambdaACCP.kinesisStream.streamName })
    }


    //Error management Lambda 
    //////////////////////////////
    const ucpErrorLambdaPrefix = 'ucpError'
    const ucpErrorLambda = new lambda.Function(this, 'ucpError' + envName, {
      code: new lambda.S3Code(lambdaArtifactRepositoryBucket, [envName, ucpErrorLambdaPrefix, 'main.zip'].join("/")),
      functionName: ucpErrorLambdaPrefix + envName,
      handler: 'main',
      runtime: lambda.Runtime.GO_1_X,
      tracing: lambda.Tracing.ACTIVE,
      //timeout should be aligned with SQS queue visibility timeout
      timeout: Duration.seconds(30),
      environment: {
        LAMBDA_ACCOUNT_ID: account,
        LAMBDA_ENV: envName,
        LAMBDA_REGION: region || "",
        DYNAMO_TABLE: errorTable.tableName,
        DYNAMO_PK: dynamo_error_pk,
        DYNAMO_SK: dynamo_error_sk,
      }
    });

    errorTable.grantReadWriteData(ucpErrorLambda)
    ucpErrorLambda.addEventSource(new lambda_event_sources.SqsEventSource(accpDomainErrorQueue));
    ucpErrorLambda.addEventSource(new lambda_event_sources.SqsEventSource(hotelBookingOutput.errorQueue));
    ucpErrorLambda.addEventSource(new lambda_event_sources.SqsEventSource(airBookingOutput.errorQueue));
    ucpErrorLambda.addEventSource(new lambda_event_sources.SqsEventSource(guestProfileOutput.errorQueue));
    ucpErrorLambda.addEventSource(new lambda_event_sources.SqsEventSource(paxProfileOutput.errorQueue));
    ucpErrorLambda.addEventSource(new lambda_event_sources.SqsEventSource(clickstreamOutput.errorQueue));
    ucpErrorLambda.addEventSource(new lambda_event_sources.SqsEventSource(hotelStayOutput.errorQueue));
    let pQueue = dlqs.get("dldPython")
    if (pQueue) {
      ucpErrorLambda.addEventSource(new lambda_event_sources.SqsEventSource(pQueue));
    }
    let gQueue = dlqs.get("dldGo")
    if (gQueue) {
      ucpErrorLambda.addEventSource(new lambda_event_sources.SqsEventSource(gQueue));
    }

    new CfnOutput(this, 'dlgRealTimeGo', { value: dlqs.get('dlgGo')?.queueUrl || "" });
    new CfnOutput(this, 'dldRealTimePython', { value: dlqs.get('dldPython')?.queueUrl || "" });
    new CfnOutput(this, 'dlgRealTimeGoTest', { value: dlqs.get('dlgGoTest')?.queueUrl || "" });
    new CfnOutput(this, 'dldRealTimePythonTest', { value: dlqs.get('dldPythonTest')?.queueUrl || "" });



    //////////////////////////
    // COGNITO USER POOL
    ///////////////////////////////////////


    const userPool: cognito.UserPool = new cognito.UserPool(this, "ucpUserpool", {
      signInAliases: { email: true },
      userPoolName: "ucpUserpool" + envName
    });

    const cognitoAppClient = new cognito.UserPoolClient(this, "ucpUserPoolClient", {
      userPool: userPool,
      oAuth: {
        flows: { authorizationCodeGrant: true, implicitCodeGrant: true },
        scopes: [cognito.OAuthScope.PHONE,
        cognito.OAuthScope.EMAIL,
        cognito.OAuthScope.OPENID,
        cognito.OAuthScope.PROFILE,
        cognito.OAuthScope.COGNITO_ADMIN],
        callbackUrls: ["http://localhost:4200"],
        logoutUrls: ["http://localhost:4200"]
      },
      supportedIdentityProviders: [cognito.UserPoolClientIdentityProvider.COGNITO],
      authFlows: {
        userPassword: true,
        userSrp: true,
        adminUserPassword: true
      },
      generateSecret: false,
      refreshTokenValidity: Duration.days(30),
      userPoolClientName: "ucpPortal",
    })

    //Adding the account ID in order to ensure uniqueness per account per region per env
    const domain = new cognito.CfnUserPoolDomain(this, id + "ucpCognitoDomain", {
      userPoolId: userPool.userPoolId,
      domain: "ucp-domain-" + account + "-" + envName
    })

    //Generating outputs used to obtain refresh token
    new CfnOutput(this, "userPoolId", { value: userPool.userPoolId })
    new CfnOutput(this, "cognitoAppClientId", { value: cognitoAppClient.userPoolClientId })
    new CfnOutput(this, "tokenEnpoint", { value: "https://" + domain.domain + ".auth." + region + ".amazoncognito.com/oauth2/token" })
    new CfnOutput(this, "cognitoDomain", { value: domain.domain })


    //////////////////////////
    //API GATEWAY
    /////////////////////////

    let apiV2 = new HttpApi(this, "apiGatewayV2", {
      apiName: "ucpBackEnd" + envName,
      corsPreflight: {
        allowOrigins: ["*"],
        allowMethods: [CorsHttpMethod.OPTIONS, CorsHttpMethod.GET, CorsHttpMethod.POST, CorsHttpMethod.PUT, CorsHttpMethod.DELETE],
        allowHeaders: ["*"]
      }
    })


    const authorizer = new HttpUserPoolAuthorizer("ucpPortalUserPoolAuthorizer", userPool, {
      userPoolClients: [cognitoAppClient]
    });


    const ucpEndpointName = "ucp"
    const ucpEndpointProfile = "profile"
    const ucpEndpointMerge = "merge"
    const ucpEndpointAdmin = "admin"
    const ucpEndpointIndustryConnector = "connector"
    const ucpEndpointDataValidation = "data"
    const ucpEndpointErrors = "error"
    const stageName = "api"
    //partner api enpoint
    let allRoutes: Array<HttpRoute>;
    allRoutes = apiV2.addRoutes({
      path: '/' + ucpEndpointName + "/" + ucpEndpointProfile,
      authorizer: authorizer,
      methods: [HttpMethod.GET],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationProfile', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    })
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + "/" + ucpEndpointProfile + "/{id}",
      authorizer: authorizer,
      methods: [HttpMethod.GET],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationProfileId', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + "/" + ucpEndpointMerge,
      authorizer: authorizer,
      methods: [HttpMethod.GET],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationMerge', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + "/" + ucpEndpointMerge + "/{id}",
      authorizer: authorizer,
      methods: [HttpMethod.GET],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationMergeId', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + "/" + ucpEndpointAdmin,
      authorizer: authorizer,
      methods: [HttpMethod.GET, HttpMethod.POST],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationProfileAdmin', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + "/" + ucpEndpointDataValidation,
      authorizer: authorizer,
      methods: [HttpMethod.GET],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationProfileAdmin', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + "/" + ucpEndpointAdmin + "/{id}",
      authorizer: authorizer,
      methods: [HttpMethod.GET, HttpMethod.DELETE],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationProfileAdminId', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + "/" + ucpEndpointIndustryConnector,
      authorizer: authorizer,
      methods: [HttpMethod.GET],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationIndustryConnector', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + '/' + ucpEndpointIndustryConnector + '/link',
      authorizer: authorizer,
      methods: [HttpMethod.POST],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationLinkIndustryConnector', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + '/' + ucpEndpointIndustryConnector + '/crawler',
      authorizer: authorizer,
      methods: [HttpMethod.POST],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationCreateConnectorCrawler', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + "/" + ucpEndpointErrors,
      authorizer: authorizer,
      methods: [HttpMethod.GET],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationErrors', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + "/" + ucpEndpointErrors + "/{id}",
      authorizer: authorizer,
      methods: [HttpMethod.DELETE],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationErrors', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    apiV2.addRoutes({
      path: '/' + ucpEndpointName + "/" + ucpEndpointErrors + "/{id}",
      authorizer: authorizer,
      methods: [HttpMethod.GET],
      integration: new HttpLambdaIntegration('UCPBackendLambdaIntegrationErrorsId', ucpBackEndLambda, {
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
      })
    }).forEach(route => {
      allRoutes.push(route)
    });
    new HttpStage(this, "apiGarewayv2Stage", {
      httpApi: apiV2,
      stageName: stageName,
      autoDeploy: true
    })
    new CfnOutput(this, 'httpApiUrl', {
      value: apiV2.apiEndpoint || ""
    });
    new CfnOutput(this, 'ucpApiId', {
      value: apiV2.apiId || ""
    });

    /////////////////////////////////
    // WEBSITE, CDN, DNS
    /////////////////////////////////

    /*********************************
    * S3 Bucket for static content
    ******************************/

    const websiteStaticContentBucket = new tah_s3.Bucket(this, "ucp-connector-fe-" + envName, accessLogBucket)
    const cloudfrontLogBucket = new tah_s3.Bucket(this, "ucp-connector-fe-logs-" + envName, accessLogBucket)

    tah_core.Output.add(this, "websiteBucket", websiteStaticContentBucket.bucketName)

    /*************************
     * CloudFront Distribution
     ****************************/


    const oai = new cloudfront.OriginAccessIdentity(this, 'websiteDistributionOAI' + envName, {
      comment: "Origin access identity for website in " + envName
    })

    let distributionConfig: cloudfront.CloudFrontWebDistributionProps = {
      originConfigs: [
        {
          s3OriginSource: {
            s3BucketSource: websiteStaticContentBucket,
            originAccessIdentity: oai
          },
          behaviors: [
            {
              isDefaultBehavior: true,
              forwardedValues: {
                "queryString": true,
                "cookies": {
                  "forward": "none"
                }
              }

            }]
        }
      ],
      loggingConfig: {
        bucket: cloudfrontLogBucket,
        includeCookies: false,
        prefix: 'prefix',
      },
      viewerCertificate: cloudfront.ViewerCertificate.fromCloudFrontDefaultCertificate(),
      errorConfigurations: [
        {
          "errorCachingMinTtl": 300,
          "errorCode": 403,
          "responseCode": 200,
          "responsePagePath": "/index.html"
        },
        {
          "errorCachingMinTtl": 300,
          "errorCode": 404,
          "responseCode": 200,
          "responsePagePath": "/index.html"
        }
      ]
    }




    const websiteDistribution = new cloudfront.CloudFrontWebDistribution(this, 'ucpWebsiteDistribution' + envName, distributionConfig);

    //We create adedicated response ehader policy to include the AWS recommanded HTTP Headers
    const responseHeaderPolicy = new cloudfront.ResponseHeadersPolicy(this, 'ResponseHeadersPolicy', {
      customHeadersBehavior: {
        customHeaders: [
          { header: 'Cache-Control', value: 'no-store, no-cache', override: true },
          { header: 'Pragma', value: 'no-cache', override: false },
        ],
      },
      securityHeadersBehavior: {
        contentTypeOptions: { override: true },
        frameOptions: { frameOption: cloudfront.HeadersFrameOption.DENY, override: true },
        contentSecurityPolicy: { contentSecurityPolicy: "default-src 'none'; img-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; object-src 'none'; connect-src *.amazonaws.com", override: true },
        strictTransportSecurity: { accessControlMaxAge: Duration.seconds(600), includeSubdomains: true, override: true },
      },
    });
    //adding the response policy ID to cloudfront distributiono using an escape hatch
    const cfnDistribution = websiteDistribution.node.defaultChild as cloudfront.CfnDistribution;
    cfnDistribution.addPropertyOverride(
      'DistributionConfig.DefaultCacheBehavior.ResponseHeadersPolicyId',
      responseHeaderPolicy.responseHeadersPolicyId
    );

    websiteStaticContentBucket.addToResourcePolicy(new iam.PolicyStatement({
      actions: ['s3:GetObject'],
      effect: iam.Effect.ALLOW,
      resources: [websiteStaticContentBucket.arnForObjects("*")],
      principals: [new iam.CanonicalUserPrincipal(oai.cloudFrontOriginAccessIdentityS3CanonicalUserId)],
    }));


    tah_core.Output.add(this, "accessLogging", accessLogBucket.bucketName)
    tah_core.Output.add(this, "websiteDistributionId", websiteDistribution.distributionId)
    tah_core.Output.add(this, "websiteDomainName", websiteDistribution.distributionDomainName)

  }


  /*******************
  * HELPER FUNCTIONS
  ******************/

  buildBusinessObjectPipeline(businessObjectName: string, envName: string, dataLakeAdminRole: iam.Role, glueDb: Database, artifactBucketName: string, accessLogBucket: s3.Bucket, connectProfileImportBucket: s3.Bucket): BusinessObjectPipelineOutput {
    const glueSchemas = new Map<string, GlueSchema>();
    glueSchemas.set("hotel-booking", glueSchemaHotelBooking)
    glueSchemas.set("air_booking", glueSchemaAirBooking)
    glueSchemas.set("guest-profile", glueSchemaGuestProfile)
    glueSchemas.set("pax-profile", glueSchemaPaxProfile)
    glueSchemas.set("clickstream", glueSchemaClickEvent)
    glueSchemas.set("hotel-stay", glueSchemaStay)

    //0-create bucket
    let bucketRaw = new tah_s3.Bucket(this, "ucp" + businessObjectName, accessLogBucket);
    //we create a test bucket to be able to test  the job run
    let testBucketRaw = new tah_s3.Bucket(this, "ucp" + businessObjectName + "Test", accessLogBucket);
    //1-Bucket permission
    bucketRaw.grantReadWrite(dataLakeAdminRole)
    testBucketRaw.grantReadWrite(dataLakeAdminRole)

    //3-Creating SQS error queue
    const errQueue = new Queue(this, businessObjectName + "-errors-" + envName)
    errQueue.grantSendMessages(dataLakeAdminRole)

    let table = this.table(this, businessObjectName, envName, glueDb, glueSchemas, bucketRaw)
    let testTable = this.table(this, businessObjectName, "Test" + envName, glueDb, glueSchemas, testBucketRaw)

    //4- Creating Jobs
    let toUcpScript = businessObjectName.replace('-', '_') + "ToUcp"
    let job = this.job(businessObjectName + "FromCustomer", envName, artifactBucketName, toUcpScript, glueDb, dataLakeAdminRole, new Map([
      ["SOURCE_TABLE", table.tableName],
      ["DEST_BUCKET", connectProfileImportBucket.bucketName],
      ["ERROR_QUEUE_URL", errQueue.queueUrl],
      ["extra-py-files", "s3://" + artifactBucketName + "/" + envName + "/etl/tah_lib.zip"]
    ]))
    let industryConnectorJob = this.job(businessObjectName + "FromConnector", envName, artifactBucketName, toUcpScript, glueDb, dataLakeAdminRole, new Map([
      // SOURCE_TABLE provided by customer when linking connector
      ["DEST_BUCKET", connectProfileImportBucket.bucketName],
      ["BUSINESS_OBJECT", businessObjectName],
      ["ERROR_QUEUE_URL", errQueue.queueUrl],
      ["extra-py-files", "s3://" + artifactBucketName + "/" + envName + "/etl/tah_lib.zip"]
    ]))
    //6- Job Triggers
    this.scheduledJobTrigger("ucp" + businessObjectName, envName, job, "cron(0 * * * ? *)")
    //7-Cfn Output
    new CfnOutput(this, 'customerBucket' + businessObjectName, {
      value: bucketRaw.bucketName
    });
    new CfnOutput(this, 'customerTestBucket' + businessObjectName, {
      value: testBucketRaw.bucketName
    });
    new CfnOutput(this, 'tableName' + businessObjectName, {
      value: table.tableName
    });
    new CfnOutput(this, 'testTableName' + businessObjectName, {
      value: testTable.tableName
    });
    new CfnOutput(this, 'customerJobName' + businessObjectName, {
      value: job.name || "",
    });
    new CfnOutput(this, 'industryConnectorJobName' + businessObjectName, {
      value: job.name || "",
    });

    return {
      connectorJobName: industryConnectorJob.name ?? "",
      customerJobName: job.name ?? "",
      bucket: bucketRaw,
      tableName: table.tableName,
      errorQueue: errQueue,
    }
  }

  job(prefix: string, envName: string, artifactBucket: string, scriptName: string, glueDb: Database, dataLakeAdminRole: iam.Role, envVar: Map<string, string>): CfnJob {
    let job = new CfnJob(this, prefix + "Job" + envName, {
      command: {
        name: "glueetl",
        scriptLocation: "s3://" + artifactBucket + "/" + envName + "/etl/" + scriptName + ".py"
      },
      glueVersion: "4.0",
      defaultArguments: {
        '--enable-continuous-cloudwatch-log': 'true',
        "--job-bookmark-option": "job-bookmark-enable",
        "--enable-metrics": "true",
        "--GLUE_DB": glueDb.databaseName,
      },
      executionProperty: {
        maxConcurrentRuns: 2
      },
      maxRetries: 0,
      name: prefix + "Job" + envName,
      role: dataLakeAdminRole.roleArn
    })
    for (let [key, value] of envVar) {
      job.defaultArguments["--" + key] = value
    }
    return job
  }

  table(scope: Construct, businessObjectName: string, envName: string, glueDb: Database, glueSchemas: Map<string, GlueSchema>, bucketRaw: s3.IBucket): Table {
    return new Table(scope, "ucpTable" + businessObjectName + envName, {
      database: glueDb,
      columns: glueSchemas.get(businessObjectName)?.columns || [],
      partitionKeys: [
        {
          name: 'year',
          type: Schema.SMALL_INT,
        },
        {
          name: 'month',
          type: Schema.SMALL_INT,
        },
        {
          name: 'day',
          type: Schema.SMALL_INT,
        }
      ],
      partitionIndexes: [{ keyNames: ['year', 'month', 'day'], indexName: "last_updated" }],
      dataFormat: DataFormat.JSON,
      bucket: bucketRaw,
    })
  }


  crawler(scope: Construct, id: string, glueDb: Database, table: Table, dataLakeAdminRole: iam.Role): glue.CfnCrawler {
    return new glue.CfnCrawler(scope, id, {
      role: dataLakeAdminRole.roleArn,
      name: id,
      description: `Glue crawler for ${table.tableName} (${id})`,
      targets: {
        catalogTargets: [{
          databaseName: glueDb.databaseName,
          tables: [table.tableName]
        }]
      },
      schemaChangePolicy: {
        deleteBehavior: "LOG",
        updateBehavior: 'LOG',
      },
      configuration: `{
        "Version": 1.0,
        "CrawlerOutput": {
            "Partitions": { "AddOrUpdateBehavior": "InheritFromTable" }
         }
     }`,
      databaseName: glueDb.databaseName
    })
  }


  jobTriggerFromCrawler(prefix: string, envName: string, crawlers: Array<CfnCrawler>, job: CfnJob, workflow?: CfnWorkflow): CfnTrigger {
    let conditions: CfnTrigger.ConditionProperty[] = []
    for (let crawler of crawlers) {
      conditions.push({
        crawlerName: crawler.name,
        crawlState: "SUCCEEDED",
        logicalOperator: "EQUALS"
      })
    }
    let trigger = new CfnTrigger(this, prefix + "jobTriggerFromCrawler" + envName, {
      type: "CONDITIONAL",
      name: prefix + "jobTriggerFromCrawler" + envName,
      predicate: {
        logical: "AND",
        conditions: conditions
      },
      startOnCreation: true,
      actions: [
        { jobName: job.name }
      ]
    })
    if (workflow) {
      trigger.addDependency(workflow)
      trigger.workflowName = workflow.name
    }
    return trigger
  }

  jobOnDemandTrigger(prefix: string, envName: string, jobs: Array<CfnJob>, workflow?: CfnWorkflow): glue.CfnTrigger {
    let actions: glue.CfnTrigger.ActionProperty[] = []
    for (let job of jobs) {
      actions.push({ jobName: job.name })
    }
    let trigger = new CfnTrigger(this, prefix + "jobOnDemandTrigger" + envName, {
      type: "ON_DEMAND",
      name: prefix + "jobOnDemandTrigger" + envName,
      actions: actions
    })
    if (workflow) {
      trigger.addDependency(workflow)
      trigger.workflowName = workflow.name
    }
    return trigger
  }

  crawlerOnDemandTrigger(prefix: string, envName: string, crawler: CfnCrawler, workflow?: CfnWorkflow): glue.CfnTrigger {
    let trigger = new CfnTrigger(this, prefix + "crawlerOnDemandTrigger" + envName, {
      type: "ON_DEMAND",
      name: crawler.name + "crawlerOnDemandTrigger" + envName,
      actions: [
        { crawlerName: crawler.name }
      ]
    })
    if (workflow) {
      trigger.addDependency(workflow)
      trigger.workflowName = workflow.name
    }
    return trigger
  }

  crawlerTriggerFromJob(prefix: string, envName: string, job: CfnJob, crawler: CfnCrawler, workflow?: CfnWorkflow): CfnTrigger {
    let trigger = new CfnTrigger(this, prefix + "crawlerTriggerFromJob" + envName, {
      type: "CONDITIONAL",
      name: prefix + "crawlerTriggerFromJob" + envName,
      predicate: {
        conditions: [
          {
            jobName: job.name,
            state: "SUCCEEDED",
            logicalOperator: "EQUALS"
          }
        ]
      },
      startOnCreation: true,
      actions: [
        { crawlerName: crawler.name }
      ]
    })
    if (workflow) {
      trigger.addDependency(workflow)
      trigger.workflowName = workflow.name
    }
    return trigger
  }

  scheduledCrawlerTrigger(prefix: string, envName: string, crawler: CfnCrawler, cron: string, workflow?: CfnWorkflow): CfnTrigger {
    let trigger = new CfnTrigger(this, prefix + "scheduledCrawlerTrigger" + envName, {
      type: "SCHEDULED",
      name: prefix + "scheduledCrawlerTrigger" + envName,
      //every hour:  "cron(10 * * * ? *)"
      schedule: cron,
      startOnCreation: true,
      actions: [
        { crawlerName: crawler.name }
      ]
    })
    if (workflow) {
      trigger.addDependency(workflow)
      trigger.workflowName = workflow.name
    }
    return trigger
  }

  scheduledJobTrigger(prefix: string, envName: string, job: CfnJob, cron: string, workflow?: CfnWorkflow): CfnTrigger {
    let trigger = new CfnTrigger(this, prefix + "scheduledJobTrigger" + envName, {
      type: "SCHEDULED",
      schedule: cron,
      startOnCreation: true,
      actions: [
        { jobName: job.name }
      ]
    })
    if (workflow) {
      trigger.addDependency(workflow)
      trigger.workflowName = workflow.name
    }
    return trigger
  }


  addTargetBucketToDatalake(prefix: string, envName: string, dataLakeAdminRole: iam.Role, servicePrincipal?: iam.ServicePrincipal): s3.Bucket {
    const bucket = new s3.Bucket(this, prefix + "-" + envName, {
      removalPolicy: RemovalPolicy.DESTROY,
      bucketName: prefix + "-" + envName,
      versioned: true
    })
    Tags.of(bucket).add('cloudrack-data-zone', 'gold');
    bucket.grantReadWrite(dataLakeAdminRole)
    return bucket
  }

  addExistingBucketToDatalake(bucketName: string, envName: string, glueDb: Database, dataLakeAdminRole: iam.Role, crawlerConfig?: any): s3.IBucket {
    const bucket = s3.Bucket.fromBucketName(this, bucketName, bucketName);
    dataLakeAdminRole.addToPolicy(new iam.PolicyStatement({
      resources: ["arn:aws:s3:::" + bucket.bucketName + "*"],
      actions: ["s3:GetObject", "s3:PutObject"]
    }))

    let crawler = new CfnCrawler(this, bucket + "-crawler-", {
      role: dataLakeAdminRole.roleArn,
      targets: {
        s3Targets: [{ path: "s3://" + bucket.bucketName }]
      },
      configuration: `{
          "Version": 1.0,
          "Grouping": {
             "TableGroupingPolicy": "CombineCompatibleSchemas" }
       }`,
      databaseName: glueDb.databaseName,
      name: bucketName + "-crawler-" + envName,
    })

    if (crawlerConfig && crawlerConfig.type === "ondemand") {
      new CfnTrigger(this, crawler.name + "Trigger", {
        type: "ON_DEMAND",
        name: crawler.name + "Trigger",
        actions: [
          { crawlerName: crawler.name }
        ]
      })
    }
    return bucket
  }

}
