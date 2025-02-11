// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { DataFormat, Database, Code as GlueCode, GlueVersion, Job, JobExecutable, PythonVersion, S3Table } from '@aws-cdk/aws-glue-alpha';
import { CfnCondition, Stack } from 'aws-cdk-lib';
import { HttpApi } from 'aws-cdk-lib/aws-apigatewayv2';
import { CfnPreparedStatement, CfnWorkGroup } from 'aws-cdk-lib/aws-athena';
import { AttributeType, StreamViewType, Table } from 'aws-cdk-lib/aws-dynamodb';
import { InstanceType, SecurityGroup, SubnetType, Vpc } from 'aws-cdk-lib/aws-ec2';
import { Role, ServicePrincipal } from 'aws-cdk-lib/aws-iam';
import { Stream } from 'aws-cdk-lib/aws-kinesis';
import { Key } from 'aws-cdk-lib/aws-kms';
import { Code, Function as LambdaFunction, Runtime } from 'aws-cdk-lib/aws-lambda';
import { LogGroup } from 'aws-cdk-lib/aws-logs';
import { DatabaseCluster, DatabaseClusterEngine } from 'aws-cdk-lib/aws-rds';
import { Bucket } from 'aws-cdk-lib/aws-s3';
import { Queue } from 'aws-cdk-lib/aws-sqs';
import path from 'path';
import { BatchProcessingOutput } from '../lib/data-ingestion';
import { ProfileStorageOutput } from '../lib/profile-storage';
import { Bucket as TAHS3Bucket } from '../tah-cdk-common/s3';

export const sendAnonymizedData = 'Yes';
export const applicationName = 'unified-profile-for-travelers-and-guests-on-aws';
export const applicationType = 'AWS-Solutions';
export const artifactBucketPath = 'artifact';
export const envName = 'dev';
export const solutionId = 'SO0244';
export const solutionName = 'Unified Profiles for Travelers and Guests on AWS';
export const solutionVersion = 'v1.0.0';

export const createAthenaWorkGroup = (stack: Stack, id: string): CfnWorkGroup => new CfnWorkGroup(stack, id, { name: 'mock' });
export const createAthenaPreparedStatement = (stack: Stack, id: string): CfnPreparedStatement =>
    new CfnPreparedStatement(stack, id, { queryStatement: '', statementName: 'mock', workGroup: '' });
export const createDynamoDbTable = (stack: Stack, id: string, hasStream: boolean): Table =>
    new Table(stack, id, {
        partitionKey: { name: 'id', type: AttributeType.STRING },
        stream: hasStream ? StreamViewType.NEW_AND_OLD_IMAGES : undefined
    });
export const createGlueDatabase = (stack: Stack, id: string): Database => new Database(stack, id);
export const createGlueJob = (stack: Stack, id: string): Job =>
    new Job(stack, id, {
        executable: JobExecutable.pythonEtl({
            glueVersion: GlueVersion.V4_0,
            pythonVersion: PythonVersion.THREE,
            script: GlueCode.fromAsset(path.join(__dirname, '..', '..', 'ucp-etls', 'etls', 'air_bookingToUcp.py'))
        }),
        role: createIamRole(stack, `${id}Role`)
    });
export const createGlueTable = (stack: Stack, id: string, database: Database): S3Table =>
    new S3Table(stack, id, {
        database,
        columns: [{ name: 'id', type: { isPrimitive: true, inputString: 'string' } }],
        dataFormat: DataFormat.JSON
    });
export const createHttpApi = (stack: Stack, id: string): HttpApi => new HttpApi(stack, id);
export const createIamRole = (stack: Stack, id: string): Role =>
    new Role(stack, id, { assumedBy: new ServicePrincipal('lambda.amazonaws.com') });
export const createKinesisStream = (stack: Stack, id: string): Stream => new Stream(stack, id);
export const createKmsKey = (stack: Stack, id: string): Key => new Key(stack, id);
export const createLambdaFunction = (stack: Stack, id: string): LambdaFunction =>
    new LambdaFunction(stack, id, {
        code: Code.fromInline('exports.handler = async () => console.log("Hello world");'),
        handler: 'index.handler',
        runtime: Runtime.NODEJS_LATEST
    });
export const createS3Bucket = (stack: Stack, id: string): Bucket => new Bucket(stack, id);
export const createTAHS3Bucket = (stack: Stack, id: string): TAHS3Bucket => new TAHS3Bucket(stack, id, createS3Bucket(stack, `${id}-logs`));
export const createSqsQueue = (stack: Stack, id: string): Queue => new Queue(stack, id);
export const createBatch = (stack: Stack, id: string): BatchProcessingOutput => ({
    bucket: createS3Bucket(stack, id + 'Bucket'),
    errorQueue: createSqsQueue(stack, id + 'Queue'),
    glueJob: createGlueJob(stack, id + 'Job')
});

export const createVpc = (stack: Stack, id: string): Vpc =>
    new Vpc(stack, id, {
        maxAzs: 3
    });

export const createStorageDb = (stack: Stack, id: string, vpc: Vpc): DatabaseCluster =>
    new DatabaseCluster(stack, id, {
        engine: DatabaseClusterEngine.AURORA,
        instanceProps: {
            instanceType: new InstanceType('t3.medium'), // Change as per your needs
            vpc,
            vpcSubnets: {
                subnetType: SubnetType.PUBLIC // Use PRIVATE for RDS by default
            }
        },
        defaultDatabaseName: 'MyDatabase' // Specify the initial database name
    });

export const createSecurityGroup = (stack: Stack, id: string, vpc: Vpc): SecurityGroup =>
    new SecurityGroup(stack, id, {
        vpc,
        description: 'Security Group for web application',
        allowAllOutbound: true // Typically set to true, modify according to your needs
    });

export const createStorageConfigTable = (stack: Stack, id: string): Table =>
    new Table(stack, id, {
        partitionKey: { name: 'id', type: AttributeType.STRING },
        stream: StreamViewType.NEW_AND_OLD_IMAGES,
        tableName: 'mockConfigTable'
    });

export const createProfileStorageOutput = (stack: Stack, id: string, vpc: Vpc): ProfileStorageOutput => ({
    lambdaToProxyGroup: createSecurityGroup(stack, id, vpc),
    storageDb: createStorageDb(stack, id + 'storageDb', vpc),
    storageDbName: '',
    storageProxy: null,
    storageProxyEndpoint: '',
    storageProxyReadonlyEndpoint: '',
    storageSecretArn: '',
    storageConfigTable: createStorageConfigTable(stack, id + 'configTable'),
    storageConfigTablePk: '',
    storageConfigTableSk: ''
});
export const createLogGroup = (stack: Stack, id: string): LogGroup => new LogGroup(stack, id);
export const createCfnCondition = (stack: Stack, id: string): CfnCondition => new CfnCondition(stack, id);
