// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { Stack, CfnOutput, RemovalPolicy, StackProps, Duration, Tags, Fn } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { aws_lambda as lambda } from 'aws-cdk-lib';
import { aws_s3 as s3 } from 'aws-cdk-lib';
import { aws_dynamodb as dynamodb } from 'aws-cdk-lib';
import { aws_apigateway as apigateway } from 'aws-cdk-lib';
import { aws_iam as iam } from 'aws-cdk-lib';
import { aws_sqs as sqs } from 'aws-cdk-lib';
import { aws_cognito as cognito } from 'aws-cdk-lib';
import { CorsHttpMethod, HttpApi, HttpMethod, HttpRoute, HttpStage, PayloadFormatVersion } from '@aws-cdk/aws-apigatewayv2-alpha';
import { HttpUserPoolAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha';
import { HttpLambdaIntegration, } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { Database } from '@aws-cdk/aws-glue-alpha';
import { CfnCrawler, CfnJob, CfnTrigger, CfnWorkflow } from 'aws-cdk-lib/aws-glue';
import { IBucket } from 'aws-cdk-lib/aws-s3';


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


    /*************************
     * Datalake 3rd party users
     ********************/
    //Amperity
    //TODO: move to cross account role when Amperity supports it
    let amperityUser = new iam.User(this, "ucp3pUserAmperity", {
      userName: "ucp3pUserAmperity" + envName
    })

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
   * Data Buckets for teporary processinng
   *****************************/
    //temp bucket for Amazon connect profile identity resolution matches
    const idResolution = new s3.Bucket(this, "ucp-connect-id-resolution-temp-" + envName, {
      removalPolicy: RemovalPolicy.DESTROY,
    })

    //Sourrce bvucket for travel business objects
    let bookingBucketRaw = this.addBucketToDatalake("ucp-data-booking", envName, datalakeAdminRole);
    let clickStreamBucketRaw = this.addBucketToDatalake("ucp-data-clickstream", envName, datalakeAdminRole);
    let loyaltyBucketRaw = this.addBucketToDatalake("ucp-data-loyalty", envName, datalakeAdminRole);

    //Target Bucket for Amazon connect profile import
    let connectProfileImportBucket = this.addBucketToDatalake("ucp-amazon-connect-profile-import", envName, datalakeAdminRole);
    //Target Bucket for Amazon connect profile export
    let connectProfileExportBucket = this.addBucketToDatalake("aucp-mazon-connect-profile-export", envName, datalakeAdminRole);
    let amperityExportBucket = this.addBucketToDatalake("ucp-amperity-export", envName, datalakeAdminRole);


    /**********************
     * ATHENA TABLES
     ************************/
    const bookingAthenaTable = Fn.join("_", Fn.split('-', bookingBucketRaw.bucketName))
    const loyaltyAthenaTable = Fn.join("_", Fn.split('-', loyaltyBucketRaw.bucketName))
    const clickStreamAthenaTable = Fn.join("_", Fn.split('-', clickStreamBucketRaw.bucketName))
    const ucpProfileAthenaTable = Fn.join("_", Fn.split('-', connectProfileExportBucket.bucketName))

    /*****************************
     * Bucket special permissions for partners
     **************************/
    //Amperity
    amperityUser.addToPolicy(new iam.PolicyStatement({
      resources: ["arn:aws:s3:::" + connectProfileExportBucket.bucketName + "*"],
      actions: ["s3:*"]
    }))

    /*********
     * Creating workflows to visualize
     ****************/
    let bookingWorflow = new CfnWorkflow(this, "ucp-workflow-booking", {
      name: "ucp-workflow-booking-" + envName
    })

    let clickstreamWorkflow = new CfnWorkflow(this, "ucp-workflow-clickstream", {
      name: "ucp-workflow-clickstream-" + envName
    })

    let loyaltyWorkflow = new CfnWorkflow(this, "ucp-workflow-loyalty", {
      name: "ucp-workflow-loyalty-" + envName
    })


    /*******************
     * 0-Raw Data Crawlers
     ******************/
    let bookingCrawler = this.crawler("ucp-data-booking", envName, glueDb, bookingBucketRaw, datalakeAdminRole)
    let clickStreamCrawler = this.crawler("ucp-data-clickstream", envName, glueDb, clickStreamBucketRaw, datalakeAdminRole)
    let loyaltyCrawler = this.crawler("ucp-data-loyalty", envName, glueDb, loyaltyBucketRaw, datalakeAdminRole)

    /*******************
     * 1-Raw Data Crawler Triggers
     ******************/
    let bookingCrawlerTrigger = this.scheduledCrawlerTrigger("ucp-data-booking", envName, bookingCrawler, "cron(0 * * * ? *)", bookingWorflow)
    this.crawlerOnDemandTrigger("ucp-data-booking", envName, bookingCrawler)
    let clickStreamCrawlerTrigger = this.scheduledCrawlerTrigger("ucp-data-clickstream", envName, clickStreamCrawler, "cron(0 * * * ? *)", clickstreamWorkflow)
    this.crawlerOnDemandTrigger("ucp-data-clickstream", envName, clickStreamCrawler)
    let loyaltyCrawlerTrigger = this.scheduledCrawlerTrigger("ucp-data-loyalty", envName, loyaltyCrawler, "cron(0 * * * ? *)", loyaltyWorkflow)
    this.crawlerOnDemandTrigger("ucp-data-loyalty", envName, loyaltyCrawler)

    /**********************
     * 2- Jobs
     *****************************/
    //booking pre-processing (flattening)
    let bookingJob = this.job("ucp-data-booking", envName, artifactBucket, "bookingToUcp", glueDb, datalakeAdminRole, new Map([
      ["SOURCE_TABLE", bookingAthenaTable],
      ["DEST_BUCKET", connectProfileImportBucket.bucketName]
    ]))
    //clickstream processing
    let clickStreamJob = this.job("ucp-data-clicktream", envName, artifactBucket, "clickstreamToUcp", glueDb, datalakeAdminRole, new Map([
      ["SOURCE_TABLE", clickStreamAthenaTable],
      ["DEST_BUCKET", connectProfileImportBucket.bucketName]
    ]))
    let loyaltyJob = this.job("ucp-data-loyalty", envName, artifactBucket, "loyaltyToUcp", glueDb, datalakeAdminRole, new Map([
      ["SOURCE_TABLE", loyaltyAthenaTable],
      ["DEST_BUCKET", connectProfileImportBucket.bucketName]
    ]))
    //Amperity job
    let amperityImportJob = this.job("ucp-amperity-import", envName, artifactBucket, "connectProfileToAmperity", glueDb, datalakeAdminRole, new Map([
      ["SOURCE_TABLE", ucpProfileAthenaTable],
      ["DEST_BUCKET", amperityExportBucket.bucketName]
    ]))
    let amperityExportJob = this.job("ucp-amperity-export", envName, artifactBucket, "amperityToMatches", glueDb, datalakeAdminRole, new Map([
      ["SOURCE_TABLE", ucpProfileAthenaTable],
      ["DEST_BUCKET", amperityExportBucket.bucketName]
    ]))

    /*******************
   * 3- Job Triggers
   ******************/
    let bookingJobTrigger = this.jobTriggerFromCrawler("ucp-data-booking", envName, [bookingCrawler], bookingJob, bookingWorflow)
    let clickStreamJobTrigger = this.jobTriggerFromCrawler("ucp-data-clickstream", envName, [clickStreamCrawler], clickStreamJob, clickstreamWorkflow)
    let loyaltyJobTrigger = this.jobTriggerFromCrawler("ucp-data-loyalty", envName, [loyaltyCrawler], loyaltyJob, loyaltyWorkflow)
    this.jobOnDemandTrigger("ucp-amperity-data-import", envName, [amperityImportJob])
    this.jobOnDemandTrigger("ucp-amperity-data-export", envName, [amperityExportJob])


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
        UCP_GUEST360_TABLE_NAME: "",
        UCP_GUEST360_TABLE_PK: "",
        UCP_GUEST360_ATHENA_TABLE: "",
      }
    });

    ucpBackEndLambda.addToRolePolicy(new iam.PolicyStatement({
      resources: ["*"],
      actions: ['profile:SearchProfiles',
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
  }


  /*******************
 * HELPER FUNCTIONS
 * TODO: to move as constructs
 ******************/

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

  crawler(prefix: string, envName: string, glueDb: Database, bucket: IBucket, dataLakeAdminRole: iam.Role): CfnCrawler {
    return new CfnCrawler(this, prefix + "-crawler-" + envName, {
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
      name: prefix + "-crawler-" + envName
    })
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

  addExistingBucketToDatalake(bucketName: string, envName: string, glueDb: Database, dataLakeAdminRole: iam.Role, crawlerConfig?: any): IBucket {
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

  addBucketToDatalake(prefix: string, envName: string, dataLakeAdminRole: iam.Role): s3.Bucket {
    const bucket = new s3.Bucket(this, prefix + "-" + envName, {
      removalPolicy: RemovalPolicy.DESTROY,
      bucketName: prefix + "-" + envName
    })
    Tags.of(bucket).add('cloudrack-data-zone', 'bronze');

    dataLakeAdminRole.addToPolicy(new iam.PolicyStatement({
      resources: ["arn:aws:s3:::" + bucket.bucketName + "*"],
      actions: ["s3:GetObject", "s3:PutObject"]
    }))
    return bucket
  }

}
