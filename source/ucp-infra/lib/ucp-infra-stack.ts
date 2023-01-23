// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { Stack, CfnOutput, RemovalPolicy, StackProps, Duration, Tags } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as kms from 'aws-cdk-lib/aws-kms';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import * as cognito from 'aws-cdk-lib/aws-cognito';

import { CorsHttpMethod, HttpApi, HttpMethod, HttpRoute, HttpStage, PayloadFormatVersion } from '@aws-cdk/aws-apigatewayv2-alpha';
import { HttpUserPoolAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha';
import { HttpLambdaIntegration, } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { Database } from '@aws-cdk/aws-glue-alpha';
import { CfnCrawler, CfnJob, CfnTrigger, CfnWorkflow } from 'aws-cdk-lib/aws-glue';
import { Queue } from 'aws-cdk-lib/aws-sqs';
import * as tah_s3 from '../tah-cdk-common/s3';
import * as tah_glue from '../tah-cdk-common/glue';
import * as tah_core from '../tah-cdk-common/core';


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


    /*************
     * Datalake admin Role
     */
    const datalakeAdminRole = new iam.Role(this, "ucp-data-admin-role-" + envName, {
      assumedBy: new iam.ServicePrincipal("glue.amazonaws.com"),
      description: "Glue role for UCP data",
      roleName: "ucp-data-admin-role-" + envName
    })
    let roleArn = new CfnOutput(this, 'ucpDataAdminRoleArn', {
      value: datalakeAdminRole.roleArn
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

    /**************
     * Glue Database
     ********************/
    const glueDb = new Database(this, "ucp-data-glue-database-" + envName, {
      databaseName: "ucp_db_" + envName
    })

    new CfnOutput(this, 'glueDBArn', {
      value: glueDb.databaseArn
    });

    /***************************
   * Data Buckets for temporary processing
   *****************************/
    //Target Bucket for Amazon connect profile import
    const connectProfileImportBucket = new tah_s3.Bucket(this, "connectProfileImportBucket", accessLogBucket)
    //temp bucket for Amazon connect profile identity resolution matches
    const idResolution = new tah_s3.Bucket(this, "ucp-connect-id-resolution-temp", accessLogBucket)

    //Source bucket for travel business objects
    this.buildBusinessObjectPipeline("hotel-booking", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)
    this.buildBusinessObjectPipeline("air-booking", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)
    this.buildBusinessObjectPipeline("guest-profile", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)
    this.buildBusinessObjectPipeline("pax-profile", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)
    this.buildBusinessObjectPipeline("clickstream", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)
    this.buildBusinessObjectPipeline("hotel-stay", envName, datalakeAdminRole, glueDb, artifactBucket, accessLogBucket, connectProfileImportBucket)

    //Target Bucket for Amazon connect profile export
    let connectProfileExportBucket = new tah_s3.Bucket(this, "ucp-mazon-connect-profile-export", accessLogBucket);
    let amperityImportBucket = new tah_s3.Bucket(this, "ucp-amperity-import", accessLogBucket);
    let amperityExportBucket = new tah_s3.Bucket(this, "ucp-amperity-export", accessLogBucket);

    /*****************************
     * Bucket Permission
     **************************/
    connectProfileImportBucket.grantReadWrite(datalakeAdminRole)
    connectProfileExportBucket.grantReadWrite(datalakeAdminRole)
    amperityExportBucket.grantReadWrite(datalakeAdminRole)

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
    const kmsKey = new kms.Key(this, "new_kms_key", {
      removalPolicy: RemovalPolicy.DESTROY,
      pendingWindow: Duration.days(20),
      alias: 'alias/mykey',
      description: 'KMS key for encrypting business object S3 buckets',
      enableKeyRotation: true,
    });


    //////////////////////////
    //LAMBDA FUNCTION
    /////////////////////////
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
        LAMBDA_ENV: envName,
        ATHENA_WORKGROUP: "",
        ATHENA_DB: "",
        CONNECTOR_CRAWLER_QUEUE: connectorCrawlerQueue.queueArn,
        CONNECTOR_CRAWLER_DLQ: connectorCrawlerDlq.queueArn,
        GLUE_DB: glueDb.databaseName,
        UCP_GUEST360_TABLE_NAME: "",
        UCP_GUEST360_TABLE_PK: "",
        UCP_GUEST360_ATHENA_TABLE: "",
      }
    });

    ucpBackEndLambda.addToRolePolicy(new iam.PolicyStatement({
      resources: ["*"],
      actions: ['glue:CreateCrawler',
        'glue:DeleteCrawler',
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
        'profile:PutProfileObjectType',
        'profile:DeleteProfileObjectType',
        'profile:GetMatches',
        'profile:ListProfileObjects',
        'profile:ListProfileObjectTypes',
        'profile:GetProfileObjectType',
        'appflow:DescribeFlow',
        'servicecatalog:ListApplications',
        'sqs:CreateQueue',
        'sqs:ReceiveMessage',
        'sqs:SetQueueAttributes',
        'sqs:GetQueueAttributes',
        'sqs:DeleteQueue']
    }));

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

  buildBusinessObjectPipeline(businessObjectName: string, envName: string, dataLakeAdminRole: iam.Role, glueDb: Database, artifactBucketName: string, accessLogBucket: s3.Bucket, connectProfileImportBucket: s3.Bucket) {
    //0-create bucket
    let bucketRaw = new tah_s3.Bucket(this, "ucp" + businessObjectName, accessLogBucket);
    //1-Bucket permission
    bucketRaw.grantReadWrite(dataLakeAdminRole)
    //2-Creating workflow to visualize
    let workflow = new CfnWorkflow(this, businessObjectName, {
      name: "ucp" + businessObjectName + envName
    })
    //3-Raw Data Crawlers
    let crawler = new tah_glue.S3Crawler(this, "ucp" + businessObjectName + envName, glueDb, bucketRaw, dataLakeAdminRole)
    //4-Raw Data Crawler Triggers
    this.scheduledCrawlerTrigger("ucp" + businessObjectName, envName, crawler, "cron(0 * * * ? *)", workflow)
    this.crawlerOnDemandTrigger("ucp" + businessObjectName, envName, crawler)
    //5- Jobs
    let job = this.job(businessObjectName, envName, artifactBucketName, businessObjectName + "ToUcp", glueDb, dataLakeAdminRole, new Map([
      ["SOURCE_TABLE", bucketRaw.toAthenaTable()],
      ["DEST_BUCKET", connectProfileImportBucket.bucketName]
    ]))
    //6- Job Triggers
    let jobTrigger = this.jobTriggerFromCrawler("ucp" + businessObjectName, envName, [crawler], job, workflow)
  }

  job(prefix: string, envName: string, artifactBucket: string, scriptName: string, glueDb: Database, dataLakeAdminRole: iam.Role, envVar: Map<string, string>): CfnJob {
    let job = new CfnJob(this, prefix + "Job" + envName, {
      command: {
        name: "glueetl",
        scriptLocation: "s3://" + artifactBucket + "/" + envName + "/etl/" + scriptName + ".py"
      },
      glueVersion: "2.0",
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




  jobTriggerFromCrawler(prefix: string, envName: string, crawlers: Array<CfnCrawler>, job: CfnJob, workflow?: CfnWorkflow): CfnTrigger {
    let conditions = []
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
      trigger.addDependsOn(workflow)
      trigger.workflowName = workflow.name
    }
    return trigger
  }

  jobOnDemandTrigger(prefix: string, envName: string, jobs: Array<CfnJob>, workflow?: CfnWorkflow) {
    let actions = []
    for (let job of jobs) {
      actions.push({ jobName: job.name })
    }
    let trigger = new CfnTrigger(this, prefix + "crawlerOnDemandTrigger" + envName, {
      type: "ON_DEMAND",
      name: prefix + "jobOnDemandTrigger" + envName,
      actions: actions
    })
    if (workflow) {
      trigger.addDependsOn(workflow)
      trigger.workflowName = workflow.name
    }
    return trigger
  }

  crawlerOnDemandTrigger(prefix: string, envName: string, crawler: CfnCrawler, workflow?: CfnWorkflow) {
    let trigger = new CfnTrigger(this, prefix + "crawlerOnDemandTrigger" + envName, {
      type: "ON_DEMAND",
      name: crawler.name + "crawlerOnDemandTrigger" + envName,
      actions: [
        { crawlerName: crawler.name }
      ]
    })
    if (workflow) {
      trigger.addDependsOn(workflow)
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
      trigger.addDependsOn(workflow)
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
      trigger.addDependsOn(workflow)
      trigger.workflowName = workflow.name
    }
    return trigger
  }

  scheduledJobTrigger(prefix: string, envName: string, job: CfnJob, cron: string, workflow?: CfnWorkflow): CfnTrigger {
    let trigger = new CfnTrigger(this, prefix + "scheduledJobTrigger" + envName, {
      type: "SCHEDULED",
      name: prefix + "scheduledJobTrigger" + envName,
      //every hour:  "cron(10 * * * ? *)"
      schedule: cron,
      startOnCreation: true,
      actions: [
        { jobName: job.name }
      ]
    })
    if (workflow) {
      trigger.addDependsOn(workflow)
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
