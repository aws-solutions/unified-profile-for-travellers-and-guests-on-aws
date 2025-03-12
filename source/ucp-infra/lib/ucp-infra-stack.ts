// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Aspects, Aws, CfnCondition, CfnMapping, CfnParameter, CfnResource, Fn, Resource, Stack, StackProps } from 'aws-cdk-lib';
import * as apiGatewayV2 from 'aws-cdk-lib/aws-apigatewayv2';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as firehose from 'aws-cdk-lib/aws-kinesisfirehose';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as lambdaEventSources from 'aws-cdk-lib/aws-lambda-event-sources';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as rds from 'aws-cdk-lib/aws-rds';
import * as s3 from 'aws-cdk-lib/aws-s3';

import { Construct } from 'constructs';
import * as tahCore from '../tah-cdk-common/core';

import { AppRegistry } from './app-registry';
import { Backend } from './backend';
import { CdkBaseProps, ConditionalPermissionsBoundary, LambdaProps, UIDeploymentOptions } from './cdk-base';
import { ChangeProcessor } from './change-processor';
import { CommonResource } from './common-resource';
import { CustomResource } from './custom-resource';
import { Dashboard } from './dashboard';
import { DataIngestion, DataIngestionProps } from './data-ingestion';
import { ErrorManagement } from './error-management';
import { Frontend } from './frontend';
import { MatchProcessor } from './match-processor';
import { MergeProcessor } from './merge-processor';
import { ProfileStorage, ProfileStorageOutput } from './profile-storage';

import { StringParameter } from 'aws-cdk-lib/aws-ssm';
import * as tahS3 from '../tah-cdk-common/s3';
import * as tahSqs from '../tah-cdk-common/sqs';
import { CPIndexer } from './cp-indexer';
import { CPWriter } from './cp-writer';
import { FargateResource } from './fargate-resource/fargate-resource';
import { S3ExciseQueueProcessor } from './s3-excise-queue-processor';
import { UptVpc, UptVpcProps } from './vpc/vpc-index';

const TRAVELER_S3_ROOT_PATH = 'profiles';
const BACKUP_S3_ROOT_PATH = 'data';

export type EcrProps = { publishEcrAssets: false; publicEcrRegistry: string; publicEcrTag: string } | { publishEcrAssets: true };

interface UCPInfraStackProps extends StackProps {
    readonly solutionId: string;
    readonly solutionName: string;
    readonly solutionVersion: string;
    readonly ecrProps: EcrProps;
    readonly isByoVpcTemplate: boolean;
}

type CfnNagSuppressRule = {
    readonly id: string;
    readonly reason: string;
};

export class UCPInfraStack extends Stack {
    constructor(scope: Construct, id: string, props: UCPInfraStackProps) {
        super(scope, id, props);

        const envName = this.node.tryGetContext('envName');
        const artifactBucketPath = this.node.tryGetContext('artifactBucketPath');
        const artifactBucketName = this.node.tryGetContext('artifactBucket');

        if (!envName) {
            throw new Error('No environment name provided for stack');
        }

        if (!artifactBucketName) {
            throw new Error('No bucket name provided for stack');
        }

        const solutionName = props.solutionName;
        const solutionId = props.solutionId;
        const solutionVersion = props.solutionVersion;

        /**
         * CloudFormation parameters
         */
        const eventBridgeEnabled = new CfnParameter(this, 'eventbridgeActivated', {
            type: 'String',
            default: 'true',
            allowedValues: ['true', 'false'],
            description:
                'If set to true, any change within a traveler profile will trigger a event in Amazon EvenBridge Bus created by the solution'
        });
        const industryConnectorBucketNameParam = new CfnParameter(this, 'industryConnectorBucketName', {
            type: 'String',
            default: '',
            description:
                '[DEPRECATED] If provided the solution will support ingesting data from the AWS Travel and Hospitality connector Catalog'
        });
        const indConBucketCondition = new CfnCondition(this, 'indConnectorBucketCondition', {
            expression: Fn.conditionEquals(industryConnectorBucketNameParam.valueAsString, '')
        });
        const isPlaceholderConnectorBucket = Fn.conditionIf(indConBucketCondition.logicalId, 'true', 'false').toString();
        const partitionStartDateParam = new CfnParameter(this, 'partitionStartDate', {
            type: 'String',
            // Set default to the solution launch date if not provided
            default: '2023/07/01',
            description: 'Date to start ingesting customer records from (YYYY/MM/DD)',
            allowedPattern: '^(?!0000)[0-9]{4}/(0?[1-9]|1[0-2])/(0?[1-9]|[1-2][0-9]|3[0-1])$',
            constraintDescription: 'Date with yyyy/mm/dd format'
        });
        const errorTTL = new CfnParameter(this, 'errorTTL', {
            type: 'Number',
            default: 7,
            description: 'The number of days to retain ingestion error records'
        });
        const skipJobRun = new CfnParameter(this, 'skipJobRun', {
            type: 'String',
            default: 'true',
            description: 'Whether or not to disable scheduled job runs, true to disable'
        });
        const inputStreamMode = new CfnParameter(this, 'inputStreamMode', {
            type: 'String',
            default: 'ON_DEMAND',
            allowedValues: ['ON_DEMAND', 'PROVISIONED'],
            description: 'Solution input stream capacity mode'
        });
        const inputStreamShards = new CfnParameter(this, 'inputStreamShards', {
            type: 'Number',
            default: 100,
            maxValue: 10000,
            description: 'Solution input stream shards count (if capacity mode is PROVISIONED)'
        });
        const ingestorShardsCount = new CfnParameter(this, 'ingestorShardsCount', {
            type: 'Number',
            default: 10,
            maxValue: 100,
            description: 'Shard count for the ingestor stream. Reach out to your account team to validate any value above 10.'
        });
        const exportStreamMode = new CfnParameter(this, 'exportStreamMode', {
            type: 'String',
            default: 'ON_DEMAND',
            allowedValues: ['ON_DEMAND', 'PROVISIONED'],
            description: 'Solution Export Stream Capacity Mode'
        });
        const exportStreamShardsCount = new CfnParameter(this, 'exportStreamShards', {
            type: 'Number',
            default: 10,
            maxValue: 10000,
            description: 'Solution export stream shards count (if capacity mode is PROVISIONED)'
        });
        const enableRealtimeBackup = new CfnParameter(this, 'enableRealtimeBackup', {
            type: 'String',
            allowedValues: [`${true}`, `${false}`],
            default: `${false}`,
            description:
                'Enable or disable (default) the Real Time Backup. Enabling this option requires manual GDPR compliance for the RealTime Export S3 Bucket'
        });
        const isRealTimeBackupEnabled = enableRealtimeBackup.valueAsString === `${true}`;
        const customerProfileStorageMode = new CfnParameter(this, 'customerProfileStorageMode', {
            type: 'String',
            allowedValues: [`${true}`, `${false}`],
            default: `${true}`,
            description: 'Flag for customer profiles storage mode'
        });
        const dynamoStorageMode = new CfnParameter(this, 'dynamoStorageMode', {
            type: 'String',
            allowedValues: [`${true}`, `${false}`],
            default: `${true}`,
            description: 'Flag for dynamo storage mode'
        });
        const maxCustomerProfileConcurrency = new CfnParameter(this, 'maxCustomerProfileConcurrency', {
            type: 'Number',
            default: 40,
            description:
                'Max concurrent writes to Amazon Connect Customer Profiles (if Customer Profiles storage mode is enabled). Please check with your account team before changing the default value.'
        });
        const frontendDeploymentOption = new CfnParameter(this, 'FrontendDeploymentOption', {
            type: 'String',
            allowedValues: Object.keys(UIDeploymentOptions),
            default: UIDeploymentOptions.CloudFront,
            description: 'Frontend deployment option'
        });
        const deployFrontendToCloudFrontCondition = new CfnCondition(this, 'deployFrontendToCloudFront', {
            expression: Fn.conditionEquals(frontendDeploymentOption.valueAsString, UIDeploymentOptions.CloudFront)
        });
        const deployFrontendToEcsCondition = new CfnCondition(this, 'deployFrontendToEcs', {
            expression: Fn.conditionEquals(frontendDeploymentOption.valueAsString, UIDeploymentOptions.ECS)
        });
        const permissionBoundaryArn = new CfnParameter(this, 'permissionBoundaryArn', {
            type: 'String',
            default: '',
            description: 'If provided, the solution will apply a permission boundary to all created roles'
        });
        const permissionBoundaryCondition = new CfnCondition(this, 'permissionBoundaryCondition', {
            expression: Fn.conditionNot(Fn.conditionEquals(permissionBoundaryArn.valueAsString, ''))
        });
        Aspects.of(this).add(new ConditionalPermissionsBoundary(permissionBoundaryCondition, permissionBoundaryArn.valueAsString));
        const artifactBucketPrefix = new CfnParameter(this, 'artifactBucketPrefix', {
            type: 'String',
            default: artifactBucketName,
            description:
                'By default, assets required to deploy the solution are in a public AWS Solutions S3 bucket. If required, you can provide an internal S3 bucket instead (all assets must be copied over to your internal S3 bucket, see Implementation Guide for guidance).'
        });
        const usePermissionSystem = new CfnParameter(this, 'usePermissionSystem', {
            type: 'String',
            allowedValues: [`${true}`, `${false}`],
            default: `${true}`,
            description:
                'By default, UPT deploys with a permission system to manage granular data access and admin task permissions via Amazon Cognito. If you wish to manage access without Amazon Cognito user groups for granular permissions, disable this feature. See Implementation Guide for guidance.'
        });

        const logLevel = new CfnParameter(this, 'logLevel', {
            type: 'String',
            allowedValues: ['DEBUG', 'INFO', 'WARN', 'ERROR'],
            default: 'INFO',
            description: 'Sets logging verbosity, from most (DEBUG) to least (ERROR) verbose.'
        });

        // Conditional BYO VPC params
        let uptVpcProps: UptVpcProps;
        if (props.isByoVpcTemplate) {
            const vpcIdParam = new CfnParameter(this, 'vpcId', {
                type: 'String',
                description: 'VPC ID that UPT will be deployed in.',
                minLength: 1
            });
            const vpcCidrBlockParam = new CfnParameter(this, 'vpcCidrBlock', {
                type: 'String',
                description: 'VPC CIDR block that UPT will be deployed in.',
                minLength: 1
            });
            const privateSubnetId1Param = new CfnParameter(this, 'privateSubnetId1', {
                type: 'String',
                description: 'Private subnet ID for AZ 1',
                minLength: 1
            });
            const privateSubnetId2Param = new CfnParameter(this, 'privateSubnetId2', {
                type: 'String',
                description: 'Private subnet ID for AZ 2',
                minLength: 1
            });
            const privateSubnetId3Param = new CfnParameter(this, 'privateSubnetId3', {
                type: 'String',
                description: 'Private subnet ID for AZ 3',
                minLength: 1
            });
            const privateSubnetRouteTableId1Param = new CfnParameter(this, 'privateSubnetRouteTableId1', {
                type: 'String',
                description: 'Private subnet route table IDs.',
                minLength: 1
            });
            const privateSubnetRouteTableId2Param = new CfnParameter(this, 'privateSubnetRouteTableId2', {
                type: 'String',
                description: 'Private subnet route table IDs.',
                minLength: 1
            });
            const privateSubnetRouteTableId3Param = new CfnParameter(this, 'privateSubnetRouteTableId3', {
                type: 'String',
                description: 'Private subnet route table IDs.',
                minLength: 1
            });

            uptVpcProps = {
                isByoVpcTemplate: true,
                privateSubnetId1Param,
                privateSubnetId2Param,
                privateSubnetId3Param,
                privateSubnetRouteTableId1Param,
                privateSubnetRouteTableId2Param,
                privateSubnetRouteTableId3Param,
                vpcCidrBlockParam,
                vpcIdParam
            };
        } else {
            uptVpcProps = {
                isByoVpcTemplate: false
            };
        }

        const map = new CfnMapping(this, 'Solution');
        map.setValue('Data', 'ID', solutionId);
        map.setValue('Data', 'Version', solutionVersion);
        map.setValue('Data', 'AppRegistryApplicationName', 'unified-profile-for-travellers-and-guests-on-aws'); // Use a short name for application as there is 128 char limit currently. This can be a shorter version of legal solution name
        map.setValue('Data', 'SolutionName', solutionName);
        map.setValue('Data', 'ApplicationType', 'AWS-Solutions');
        map.setValue('Data', 'SendAnonymizedData', 'Yes');

        const sendAnonymizedData = map.findInMap('Data', 'SendAnonymizedData');
        const regionalBucketName = Fn.join('-', [artifactBucketPrefix.valueAsString, Aws.REGION]);
        const artifactBucket = s3.Bucket.fromBucketName(this, 'ArtifactBucket', regionalBucketName);
        const cdkProps: CdkBaseProps = { envName, stack: this };
        const lambdaProps: LambdaProps = { artifactBucketPath, artifactBucket, sendAnonymizedData, solutionId, solutionVersion };

        /**
         * VPC
         */
        const uptVpcObj = new UptVpc(cdkProps, uptVpcProps);
        const uptVpc = uptVpcObj.uptVpc;

        /**
         * App Registry
         */
        const appReg = new AppRegistry(cdkProps, {
            applicationName: map.findInMap('Data', 'AppRegistryApplicationName'),
            applicationType: map.findInMap('Data', 'ApplicationType'),
            solutionId: map.findInMap('Data', 'ID'),
            solutionName: map.findInMap('Data', 'SolutionName'),
            solutionVersion: map.findInMap('Data', 'Version')
        });
        tahCore.Output.add(this, 'appRegistry', appReg.props.applicationName);

        /**
         * Common resources used widely
         */
        const {
            accessLogBucket,
            athenaResultsBucket,
            athenaWorkGroup,
            athenaSearchProfileS3PathsPreparedStatement,
            athenaGetS3PathsByConnectIdsPreparedStatement,
            configTable,
            connectProfileDomainErrorQueue,
            connectProfileImportBucket,
            dataLakeAdminRole,
            dataSyncLogGroup,
            gdprPurgeLogGroup,
            dynamoDbKey,
            glueDatabase,
            glueTravelerTable,
            outputBucket,
            portalConfigTable,
            privacySearchResultsTable,
            distributedMutexTable
        } = new CommonResource(cdkProps, { outputBucketPrefix: TRAVELER_S3_ROOT_PATH });

        /**
         * Low cost storage
         */
        const profileStorage = new ProfileStorage(cdkProps, { uptVpc: uptVpcObj, dynamoDbKey });
        const profileStorageOutput: ProfileStorageOutput = {
            lambdaToProxyGroup: profileStorage.lambdaToProxyGroup,
            storageDbName: profileStorage.dbName,
            storageDb: profileStorage.storageAuroraCluster,
            storageProxy: profileStorage.storageRdsProxy!,
            storageProxyEndpoint: profileStorage.storageRdsProxy ? profileStorage.storageRdsProxy.endpoint : '',
            storageProxyReadonlyEndpoint: profileStorage.storageRdsProxyReadonlyEndpoint!,
            storageSecretArn: profileStorage.storageAuroraCluster.secret!.secretArn,
            storageConfigTable: profileStorage.storageConfigTable,
            storageConfigTablePk: profileStorage.storageConfigTablePk,
            storageConfigTableSk: profileStorage.storageConfigTableSk
        };

        /**
         * Amazon Connect Customer Profiles Writer
         */
        const { writerQueue, writerLambda, writerDLQ } = new CPWriter(cdkProps, {
            maxConcurrency: maxCustomerProfileConcurrency.valueAsNumber,
            ...lambdaProps
        });

        const { cpIndexTable, cpExportStream, cpIndexerDLQ, indexLambda } = new CPIndexer(cdkProps, {
            maxConcurrency: maxCustomerProfileConcurrency.valueAsNumber,
            dynamoDbKey: dynamoDbKey,
            ...lambdaProps
        });

        /**
         * Change processor
         */
        const { changeProcessorKinesisLambda, dlqChangeProcessor, deliveryStreamLogGroup, deliveryStream, firehoseFormatConversionRole } =
            new ChangeProcessor(cdkProps, {
                bucket: outputBucket,
                bucketPrefix: TRAVELER_S3_ROOT_PATH,
                eventBridgeEnabled: eventBridgeEnabled.valueAsString,
                exportStreamMode: exportStreamMode.valueAsString,
                exportStreamShardCount: exportStreamShardsCount.valueAsNumber,
                glueDatabase,
                glueTravelerTable,
                profileStorageOutput,
                uptVpc,
                ...lambdaProps
            });

        /**
         * Match processor
         */
        const { matchBucket, matchLambdaFunction, matchTable } = new MatchProcessor(cdkProps, {
            accessLogBucket,
            dynamoDbKey,
            profileStorageOutput,
            changeProcessorKinesisStream: changeProcessorKinesisLambda.kinesisStream,
            portalConfigTable,
            uptVpc,
            ...lambdaProps
        });

        /**
         * Data ingestion
         */
        // Industry connector placeholder bucket (in case industry connector is not provided)
        const industryConnectorPlaceholderBucket = new tahS3.Bucket(this, 'ucp-ind-connector-placeholder', accessLogBucket);
        const industryConnectorBucketName = Fn.conditionIf(
            indConBucketCondition.logicalId,
            industryConnectorPlaceholderBucket.bucketName,
            industryConnectorBucketNameParam.valueAsString
        ).toString();
        const dataIngestionProps: DataIngestionProps = {
            accessLogBucket,
            athenaWorkGroup,
            configTable,
            connectProfileDomainErrorQueue,
            connectProfileImportBucket,
            dataLakeAdminRole,
            deliveryStreamBucketPrefix: BACKUP_S3_ROOT_PATH,
            entryStreamMode: inputStreamMode.valueAsString,
            entryStreamShardCount: inputStreamShards.valueAsNumber,
            glueDatabase,
            industryConnectorBucketName,
            ingestorShardCount: ingestorShardsCount.valueAsNumber,
            matchBucket,
            partitionStartDate: partitionStartDateParam.valueAsString,
            region: Aws.REGION,
            skipJobRun: skipJobRun.valueAsString,
            enableRealTimeBackup: isRealTimeBackupEnabled,
            profileStorageOutput,
            changeProcessorKinesisStream: changeProcessorKinesisLambda.kinesisStream,
            uptVpc,
            ...lambdaProps
        };

        const {
            // Dead letter queues
            dlqGo,
            dlqPython,
            dlqBatchProcessor,
            // Real-time processing
            entryKinesisLambda,
            ingestorKinesisLambda,
            realTimeDeliveryStreamLogGroup,
            realTimeDeliveryStream,
            // Batch processing
            airBookingBatch,
            clickStreamBatch,
            customerServiceInteractionBatch,
            hotelBookingBatch,
            hotelStayBatch,
            guestProfileBatch,
            paxProfileBatch,
            syncLambdaFunction,
            batchLambdaFunction,
            // Industry connector
            industryConnectorBucket,
            industryConnectorDataSyncRole,
            industryConnectorLambdaFunction
        } = new DataIngestion(cdkProps, dataIngestionProps);

        /**
         * Error management
         */
        const { errorLambdaFunction, errorRetryLambdaFunction, errorTable, errorRetryTable } = new ErrorManagement(cdkProps, {
            dynamoDbKey,
            errorTtl: errorTTL.valueAsString,
            profileStorageOutput,
            changeProcessorKinesisStream: changeProcessorKinesisLambda.kinesisStream,
            ...lambdaProps
        });

        /**
         * Merge Processor
         */
        const { dlqMerger, mergeQueue, mergerLambda } = new MergeProcessor(cdkProps, {
            profileStorageOutput,
            changeProcessorKinesisStream: changeProcessorKinesisLambda.kinesisStream,
            uptVpc,
            ...lambdaProps
        });

        mergeQueue.grantConsumeMessages(ingestorKinesisLambda.lambdaFunction);
        mergeQueue.grantSendMessages(ingestorKinesisLambda.lambdaFunction);
        mergeQueue.grantSendMessages(batchLambdaFunction);
        mergeQueue.grantSendMessages(matchLambdaFunction);

        // SSM parameters are to make resource IDs available to Lambdas
        // our Lambda environments are near the 4KiB limit
        const ssmParamNamespace = `/upt${envName}/`;
        tahCore.Output.add(this, 'ssmParamNamespace', ssmParamNamespace);
        const { irResources, rebuildCacheResources } = new FargateResource(this, {
            profileStorageOutput,
            mergeQueueUrl: mergeQueue.queueUrl,
            cpWriterQueueUrl: writerQueue.queueUrl,
            changeProcessorKinesisStream: changeProcessorKinesisLambda.kinesisStream,
            ecrProps: props.ecrProps,
            solutionId,
            solutionVersion,
            uptVpc,
            logLevel: logLevel.valueAsString
        });

        const ssmIdResClusterArn = new StringParameter(this, 'IdResClusterArn', {
            stringValue: irResources.cluster.clusterArn,
            parameterName: `${ssmParamNamespace}IdResClusterArn`
        });
        const ssmIdResTaskDefinitionArn = new StringParameter(this, 'IdResTaskDefinitionArn', {
            stringValue: irResources.taskDefinition.taskDefinitionArn,
            parameterName: `${ssmParamNamespace}IdResTaskDefinitionArn`
        });
        if (!irResources.taskDefinition.defaultContainer) {
            throw new Error('No default container for ID resolution');
        }
        const ssmIdResContainerName = new StringParameter(this, 'IdResContainerName', {
            stringValue: irResources.taskDefinition.defaultContainer.containerName,
            parameterName: `${ssmParamNamespace}IdResContainerName`
        });

        const ssmRebuildCacheClusterArn = new StringParameter(this, 'RebuildCacheClusterArn', {
            stringValue: rebuildCacheResources.cluster.clusterArn,
            parameterName: `${ssmParamNamespace}RebuildCacheClusterArn`
        });
        const ssmRebuildCacheTaskDefinitionArn = new StringParameter(this, 'RebuildCacheTaskDefinitionArn', {
            stringValue: rebuildCacheResources.taskDefinition.taskDefinitionArn,
            parameterName: `${ssmParamNamespace}RebuildCacheTaskDefinitionArn`
        });
        if (!rebuildCacheResources.taskDefinition.defaultContainer) {
            throw new Error('No default container for ID resolution');
        }
        const ssmRebuildCacheContainerName = new StringParameter(this, 'RebuildCacheContainerName', {
            stringValue: rebuildCacheResources.taskDefinition.defaultContainer.containerName,
            parameterName: `${ssmParamNamespace}RebuildCacheContainerName`
        });

        mergeQueue.grantConsumeMessages(irResources.taskRole);
        mergeQueue.grantSendMessages(irResources.taskRole);

        /**
         * S3 Excise Queue Lambda Processor
         */
        const { s3ExciseQueue, s3ExciseQueueLambdaProcessor } = new S3ExciseQueueProcessor({
            artifactBucket,
            artifactBucketPath,
            glueDatabase,
            glueTravelerTable,
            outputBucket,
            travelerS3RootPath: TRAVELER_S3_ROOT_PATH,
            distributedMutexTable,
            privacySearchResultsTable,
            ...cdkProps
        });

        const ssmVpcSubnets = new StringParameter(this, 'VpcSubnets', {
            stringValue: Fn.join(
                ',',
                uptVpc.privateSubnets.map(subnet => subnet.subnetId)
            ),
            parameterName: `${ssmParamNamespace}VpcSubnets`
        });
        const ssmVpcSecurityGroup = new StringParameter(this, 'VpcSecurityGroup', {
            stringValue: profileStorageOutput.lambdaToProxyGroup.securityGroupId,
            parameterName: `${ssmParamNamespace}VpcSecurityGroup`
        });

        //fixes the constructor SonarQube issue and does not hurt to have these as output.
        tahCore.Output.add(this, 'ssmIdResClusterArn', ssmIdResClusterArn.parameterArn);
        tahCore.Output.add(this, 'ssmIdResTaskDefinitionArn', ssmIdResTaskDefinitionArn.parameterArn);
        tahCore.Output.add(this, 'ssmIdResContainerName', ssmIdResContainerName.parameterArn);
        tahCore.Output.add(this, 'ssmRebuildCacheClusterArn', ssmRebuildCacheClusterArn.parameterArn);
        tahCore.Output.add(this, 'ssmRebuildCacheTaskDefinitionArn', ssmRebuildCacheTaskDefinitionArn.parameterArn);
        tahCore.Output.add(this, 'ssmRebuildCacheContainerName', ssmRebuildCacheContainerName.parameterArn);
        tahCore.Output.add(this, 'ssmVpcSubnets', ssmVpcSubnets.parameterArn);
        tahCore.Output.add(this, 'ssmVpcSecurityGroup', ssmVpcSecurityGroup.parameterArn);

        /**
         * Backend
         */
        const { api, asyncLambdaFunction, backendLambdaFunction, userPool, userPoolClient, stage, domainKey } = new Backend(cdkProps, {
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
            privacySearchResultsTable,
            athenaResultsBucket,
            athenaWorkGroup,
            athenaSearchProfileS3PathsPreparedStatement,
            athenaGetS3PathsByConnectIdsPreparedStatement,
            outputBucket,
            portalConfigTable,
            stackLogicalId: id,
            syncLambdaFunction,
            travelerS3RootPath: TRAVELER_S3_ROOT_PATH,
            // Batch processing
            airBookingBatch,
            clickStreamBatch,
            customerServiceInteractionBatch,
            guestProfileBatch,
            hotelBookingBatch,
            hotelStayBatch,
            paxProfileBatch,
            // Low cost storage
            profileStorageOutput,
            changeProcessorKinesisStream: changeProcessorKinesisLambda.kinesisStream,
            // GDPR
            s3ExciseQueue,
            gdprPurgeLogGroup,
            ssmParamNamespace,
            irCluster: irResources.cluster,
            uptVpc,
            ...lambdaProps
        });

        if (!asyncLambdaFunction.role) {
            throw new Error('No role for async lambda');
        }
        irResources.taskDefinition.grantRun(asyncLambdaFunction.role);
        rebuildCacheResources.taskDefinition.grantRun(asyncLambdaFunction.role);

        /**
         * Frontend
         */
        const { websiteDistribution, websiteStaticContentBucket } = new Frontend(cdkProps, {
            accessLogBucket,
            deployFrontendToCloudFrontCondition,
            deployFrontendToEcsCondition,
            uptVpc: uptVpc,
            ecrProps: props.ecrProps,
            ucpApiUrl: api.url ?? '',
            cognitoUserPoolId: userPool.userPoolId,
            cognitoClientId: userPoolClient.userPoolClientId,
            usePermissionSystem: usePermissionSystem.valueAsString
        });

        /**
         * Custom resource that generate UUID for usage metrics and deploys frontend assets.
         */
        const { customResourceLambda, uuid } = new CustomResource(cdkProps, {
            frontEndConfig: {
                cognitoClientId: userPoolClient.userPoolClientId,
                cognitoUserPoolId: userPool.userPoolId,
                deployed: new Date().toISOString(),
                ucpApiUrl: api.url ?? '',
                staticContentBucketName: websiteStaticContentBucket.bucketName,
                usePermissionSystem: usePermissionSystem.valueAsString
            },
            solutionName,
            ...lambdaProps
        });

        /**
         * Dashboard
         */
        const dashboard = new Dashboard(cdkProps, {
            // Real-time ingestion
            entryKinesisStream: entryKinesisLambda.kinesisStream,
            entryLambdaFunction: entryKinesisLambda.lambdaFunction,
            ingestorKinesisStream: ingestorKinesisLambda.kinesisStream,
            ingestorLambdaFunction: ingestorKinesisLambda.lambdaFunction,
            // Change processor
            changeProcessorKinesisStream: changeProcessorKinesisLambda.kinesisStream,
            changeProcessorLambdaFunction: changeProcessorKinesisLambda.lambdaFunction,
            // Error queues, Lambda function, and DynamoDB table
            airBookingBatchErrorQueue: airBookingBatch.errorQueue,
            clickStreamBatchErrorQueue: clickStreamBatch.errorQueue,
            customerServiceInteractionBatchErrorQueue: customerServiceInteractionBatch.errorQueue,
            guestProfileBatchErrorQueue: guestProfileBatch.errorQueue,
            hotelBookingBatchErrorQueue: hotelBookingBatch.errorQueue,
            hotelStayBatchErrorQueue: hotelStayBatch.errorQueue,
            paxProfileBatchErrorQueue: paxProfileBatch.errorQueue,
            dlqChangeProcessor,
            dlqGo,
            dlqPython,
            dlqBatchProcessor,
            connectProfileDomainErrorQueue,
            errorRetryLambdaFunction,
            errorTable,
            // API
            api,
            // Sync Lambda function
            syncLambdaFunction,
            // CP Writer
            cpWriterQueue: writerQueue
        });

        tahCore.Output.add(this, 'cloudwatchDashboard', dashboard.dashboard.dashboardName);

        /**
         * Error Lambda function event sources
         */
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(connectProfileDomainErrorQueue));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(airBookingBatch.errorQueue));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(clickStreamBatch.errorQueue));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(customerServiceInteractionBatch.errorQueue));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(guestProfileBatch.errorQueue));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(hotelBookingBatch.errorQueue));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(hotelStayBatch.errorQueue));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(paxProfileBatch.errorQueue));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(dlqChangeProcessor));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(dlqMerger));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(dlqGo));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(dlqPython));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(dlqBatchProcessor));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(writerDLQ));
        errorLambdaFunction.addEventSource(new lambdaEventSources.SqsEventSource(cpIndexerDLQ));

        /**
         * Additional cross functional permissions
         */
        // Async Lambda function
        configTable.grantReadWriteData(asyncLambdaFunction);
        matchTable.grantReadWriteData(asyncLambdaFunction);

        // Backend Lambda function
        airBookingBatch.bucket.grantRead(backendLambdaFunction);
        clickStreamBatch.bucket.grantRead(backendLambdaFunction);
        customerServiceInteractionBatch.bucket.grantRead(backendLambdaFunction);
        guestProfileBatch.bucket.grantRead(backendLambdaFunction);
        hotelBookingBatch.bucket.grantRead(backendLambdaFunction);
        hotelStayBatch.bucket.grantRead(backendLambdaFunction);
        paxProfileBatch.bucket.grantRead(backendLambdaFunction);
        connectProfileImportBucket.grantRead(backendLambdaFunction);
        configTable.grantReadWriteData(backendLambdaFunction);
        errorTable.grantReadWriteData(backendLambdaFunction);
        matchTable.grantReadWriteData(backendLambdaFunction);
        portalConfigTable.grantReadWriteData(backendLambdaFunction);
        matchBucket.grantReadWrite(backendLambdaFunction);

        // Match Lambda function
        configTable.grantReadWriteData(matchLambdaFunction);
        portalConfigTable.grantReadWriteData(matchLambdaFunction);
        cpIndexTable.grantReadData(matchLambdaFunction);

        //CP Writer Lambda
        cpIndexTable.grantReadData(writerLambda);

        // Granting permission to read and write on the Athena result bucket thus allowing Lambda function
        // to successfully execute Athena queries using the workgroup created in this stack.
        athenaResultsBucket.grantReadWrite(syncLambdaFunction);
        configTable.grantReadWriteData(syncLambdaFunction);

        // Low Cost Storage
        if (profileStorageOutput.storageProxy && profileStorageOutput.storageDb && profileStorageOutput.storageConfigTable) {
            // Interact with Aurora cluster
            profileStorageOutput.storageProxy.grantConnect(backendLambdaFunction);
            profileStorageOutput.storageProxy.grantConnect(asyncLambdaFunction);
            profileStorageOutput.storageProxy.grantConnect(ingestorKinesisLambda.lambdaFunction);
            profileStorageOutput.storageProxy.grantConnect(mergerLambda);
            profileStorageOutput.storageProxy.grantConnect(irResources.taskRole);
            profileStorageOutput.storageProxy.grantConnect(rebuildCacheResources.taskRole);
            profileStorageOutput.storageProxy.grantConnect(changeProcessorKinesisLambda.lambdaFunction);
            profileStorageOutput.storageProxy.grantConnect(matchLambdaFunction);
            profileStorageOutput.storageDb.secret?.grantRead(backendLambdaFunction);
            profileStorageOutput.storageDb.secret?.grantRead(asyncLambdaFunction);
            profileStorageOutput.storageDb.secret?.grantRead(ingestorKinesisLambda.lambdaFunction);
            profileStorageOutput.storageDb.secret?.grantRead(mergerLambda);
            profileStorageOutput.storageDb.secret?.grantRead(irResources.taskRole);
            profileStorageOutput.storageDb.secret?.grantRead(rebuildCacheResources.taskRole);
            profileStorageOutput.storageDb.secret?.grantRead(changeProcessorKinesisLambda.lambdaFunction);
            profileStorageOutput.storageDb.secret?.grantRead(matchLambdaFunction);
            // Update config table
            profileStorageOutput.storageConfigTable.grantReadWriteData(backendLambdaFunction);
            profileStorageOutput.storageConfigTable.grantReadWriteData(asyncLambdaFunction);
            profileStorageOutput.storageConfigTable.grantReadWriteData(ingestorKinesisLambda.lambdaFunction);
            profileStorageOutput.storageConfigTable.grantReadData(changeProcessorKinesisLambda.lambdaFunction);
            profileStorageOutput.storageConfigTable.grantReadData(mergerLambda);
            profileStorageOutput.storageConfigTable.grantReadData(irResources.taskRole);
            profileStorageOutput.storageConfigTable.grantReadData(rebuildCacheResources.taskRole);
            profileStorageOutput.storageConfigTable.grantReadData(matchLambdaFunction);
            // Send events to Kinesis
            changeProcessorKinesisLambda.kinesisStream.grantWrite(ingestorKinesisLambda.lambdaFunction);
            changeProcessorKinesisLambda.kinesisStream.grantWrite(mergerLambda);
        }

        // Change processor
        outputBucket.grantReadWrite(changeProcessorKinesisLambda.lambdaFunction);

        // Industry connector Lambda function
        configTable.grantReadData(<iam.Role>industryConnectorLambdaFunction.role);

        // We need to grant the data admin (which will be the role for all glue jobs) access to the artifact bucket
        // since the Python scripts are stored there.
        artifactBucket.grantReadWrite(dataLakeAdminRole);

        // Custom resource Lambda function
        artifactBucket.grantReadWrite(customResourceLambda);
        websiteStaticContentBucket.grantReadWrite(customResourceLambda);

        /**
         * Lambda function environment variables
         */
        // Add Lambda env variables required to create DataSync tasks
        backendLambdaFunction.addEnvironment('ACCP_DESTINATION_STREAM', changeProcessorKinesisLambda.kinesisStream.streamArn);
        backendLambdaFunction.addEnvironment('INDUSTRY_CONNECTOR_BUCKET_NAME', industryConnectorBucket.bucketName);
        backendLambdaFunction.addEnvironment('DATASYNC_ROLE_ARN', industryConnectorDataSyncRole.roleArn);
        backendLambdaFunction.addEnvironment('DATASYNC_LOG_GROUP_ARN', dataSyncLogGroup.logGroupArn);
        backendLambdaFunction.addEnvironment('MERGE_QUEUE_URL', mergeQueue.queueUrl);
        asyncLambdaFunction.addEnvironment('MERGE_QUEUE_URL', mergeQueue.queueUrl);
        asyncLambdaFunction.addEnvironment('CP_EXPORT_STREAM', cpExportStream.streamArn);
        changeProcessorKinesisLambda.lambdaFunction.addEnvironment('MERGE_QUEUE_URL', mergeQueue.queueUrl);
        ingestorKinesisLambda.lambdaFunction.addEnvironment('MERGE_QUEUE_URL', mergeQueue.queueUrl);
        syncLambdaFunction.addEnvironment('MERGE_QUEUE_URL', mergeQueue.queueUrl);
        errorRetryLambdaFunction.addEnvironment('MERGE_QUEUE_URL', mergeQueue.queueUrl);
        batchLambdaFunction.addEnvironment('MERGE_QUEUE_URL', mergeQueue.queueUrl);
        matchLambdaFunction.addEnvironment('MERGE_QUEUE_URL', mergeQueue.queueUrl);
        writerLambda.addEnvironment('CP_INDEX_TABLE', cpIndexTable.tableName);

        // UUID for metrics
        asyncLambdaFunction.addEnvironment('METRICS_UUID', uuid);
        backendLambdaFunction.addEnvironment('METRICS_UUID', uuid);
        changeProcessorKinesisLambda.lambdaFunction.addEnvironment('METRICS_UUID', uuid);
        entryKinesisLambda.lambdaFunction.addEnvironment('METRICS_UUID', uuid);
        errorLambdaFunction.addEnvironment('METRICS_UUID', uuid);
        errorRetryLambdaFunction.addEnvironment('METRICS_UUID', uuid);
        industryConnectorLambdaFunction.addEnvironment('METRICS_UUID', uuid);
        ingestorKinesisLambda.lambdaFunction.addEnvironment('METRICS_UUID', uuid);
        matchLambdaFunction.addEnvironment('METRICS_UUID', uuid);
        syncLambdaFunction.addEnvironment('METRICS_UUID', uuid);
        batchLambdaFunction.addEnvironment('METRICS_UUID', uuid);
        writerLambda.addEnvironment('METRICS_UUID', uuid);

        //storage modes
        asyncLambdaFunction.addEnvironment('CUSTOMER_PROFILE_STORAGE_MODE', customerProfileStorageMode.valueAsString);
        asyncLambdaFunction.addEnvironment('DYNAMO_STORAGE_MODE', dynamoStorageMode.valueAsString);

        // Granular permission system usage
        backendLambdaFunction.addEnvironment('USE_PERMISSION_SYSTEM', usePermissionSystem.valueAsString);
        asyncLambdaFunction.addEnvironment('USE_PERMISSION_SYSTEM', usePermissionSystem.valueAsString);

        // Log level
        backendLambdaFunction.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        asyncLambdaFunction.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        changeProcessorKinesisLambda.lambdaFunction.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        entryKinesisLambda.lambdaFunction.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        errorLambdaFunction.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        errorRetryLambdaFunction.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        industryConnectorLambdaFunction.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        ingestorKinesisLambda.lambdaFunction.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        matchLambdaFunction.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        syncLambdaFunction.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        batchLambdaFunction.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        writerLambda.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        indexLambda.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        mergerLambda.addEnvironment('LOG_LEVEL', logLevel.valueAsString);
        customResourceLambda.addEnvironment('LOG_LEVEL', logLevel.valueAsString);

        // CP Writer Queue
        writerQueue.grantSendMessages(ingestorKinesisLambda.lambdaFunction);
        writerQueue.grantSendMessages(batchLambdaFunction);
        writerQueue.grantSendMessages(backendLambdaFunction);
        writerQueue.grantSendMessages(asyncLambdaFunction);
        writerQueue.grantSendMessages(mergerLambda);
        writerQueue.grantSendMessages(s3ExciseQueueLambdaProcessor);
        writerQueue.grantSendMessages(rebuildCacheResources.taskRole);

        writerLambda.addToRolePolicy(
            new iam.PolicyStatement({
                actions: ['kms:GenerateDataKey'],
                effect: iam.Effect.ALLOW,
                resources: [domainKey.keyArn]
            })
        );

        asyncLambdaFunction.addEnvironment('CP_WRITER_QUEUE_URL', writerQueue.queueUrl);
        backendLambdaFunction.addEnvironment('CP_WRITER_QUEUE_URL', writerQueue.queueUrl);
        batchLambdaFunction.addEnvironment('CP_WRITER_QUEUE_URL', writerQueue.queueUrl);
        changeProcessorKinesisLambda.lambdaFunction.addEnvironment('CP_WRITER_QUEUE_URL', writerQueue.queueUrl);
        mergerLambda.addEnvironment('CP_WRITER_QUEUE_URL', writerQueue.queueUrl);
        ingestorKinesisLambda.lambdaFunction.addEnvironment('CP_WRITER_QUEUE_URL', writerQueue.queueUrl);
        errorRetryLambdaFunction.addEnvironment('CP_WRITER_QUEUE_URL', writerQueue.queueUrl);
        syncLambdaFunction.addEnvironment('CP_WRITER_QUEUE_URL', writerQueue.queueUrl);
        matchLambdaFunction.addEnvironment('CP_WRITER_QUEUE_URL', writerQueue.queueUrl);
        matchLambdaFunction.addEnvironment('INDEX_TABLE', cpIndexTable.tableName);

        /**
         * CFN NAG suppressions
         */
        this.suppressLambdaFunction(entryKinesisLambda.lambdaFunction);
        this.suppressLambdaFunction(ingestorKinesisLambda.lambdaFunction);
        this.suppressLambdaFunction(asyncLambdaFunction);
        this.suppressLambdaFunction(backendLambdaFunction);
        this.suppressLambdaFunction(customResourceLambda);
        this.suppressLambdaFunction(errorLambdaFunction);
        this.suppressLambdaFunction(errorRetryLambdaFunction);
        this.suppressLambdaFunction(industryConnectorLambdaFunction);
        this.suppressLambdaFunction(matchLambdaFunction);
        this.suppressLambdaFunction(syncLambdaFunction);
        this.suppressLambdaFunction(batchLambdaFunction);
        this.suppressLambdaFunction(s3ExciseQueueLambdaProcessor);
        this.suppressLambdaFunction(writerLambda);
        this.suppressLambdaFunction(indexLambda);
        this.suppressLambdaFunction(mergerLambda);
        this.suppressLambdaFunction(changeProcessorKinesisLambda.lambdaFunction);
        this.suppressLambdaFunction(ingestorKinesisLambda.lambdaFunction);
        this.suppressAccessLogBucket(accessLogBucket);
        this.suppressDynamoDb(configTable);
        this.suppressDynamoDb(errorTable);
        this.suppressDynamoDb(errorRetryTable);
        this.suppressDynamoDb(matchTable);
        this.suppressDynamoDb(portalConfigTable);
        this.suppressDynamoDb(profileStorageOutput.storageConfigTable);
        this.suppressDynamoDb(cpIndexTable);
        this.suppressLog(dataSyncLogGroup);
        this.suppressLog(gdprPurgeLogGroup);
        this.suppressLog(deliveryStreamLogGroup);
        this.suppressLog(realTimeDeliveryStreamLogGroup);
        this.suppressCloudfront(websiteDistribution);
        this.suppressSqs(connectProfileDomainErrorQueue);
        this.suppressSqs(airBookingBatch.errorQueue);
        this.suppressSqs(clickStreamBatch.errorQueue);
        this.suppressSqs(customerServiceInteractionBatch.errorQueue);
        this.suppressSqs(guestProfileBatch.errorQueue);
        this.suppressSqs(hotelStayBatch.errorQueue);
        this.suppressSqs(hotelBookingBatch.errorQueue);
        this.suppressSqs(paxProfileBatch.errorQueue);
        this.suppressSqs(cpIndexerDLQ);
        this.suppressSqs(dlqChangeProcessor);
        this.suppressSqs(dlqGo);
        this.suppressSqs(dlqPython);
        this.suppressSqs(s3ExciseQueue);
        this.suppressSqs(dlqMerger);
        this.suppressSqs(mergeQueue);
        this.suppressAccessLogHttp(stage);
        this.suppressAccessLogHttp(api.defaultStage as apiGatewayV2.HttpStage);
        this.suppressRole(dataLakeAdminRole);
        this.suppressRole(industryConnectorDataSyncRole);
        this.suppressRole(firehoseFormatConversionRole);
        this.suppressRdsWarnings(uptVpc, profileStorage);
        this.suppressDeliveryStream(deliveryStream);
        this.suppressDeliveryStream(realTimeDeliveryStream);
        this.addDependencies(matchBucket, industryConnectorPlaceholderBucket);
    }

    private addDependencies(matchBucket: tahS3.Bucket, industryConnectorPlaceholderBucket: tahS3.Bucket): void {
        for (const child of this.node.children) {
            /**
             * As "BucketNotificationsHandler" Lambda function has all other resources as children,
             * this is going to be run just once eventually.
             */
            if (child.node.id.startsWith('BucketNotificationsHandler')) {
                // Run your function on the child here
                this.suppressConstruct(child);

                /**
                 * To fix occasional CloudFormation stack deletion failure issue,
                 * bucket policy should be a dependency of this Lambda function.
                 * `buckets` are S3 buckets which needs to create bucket notifications.
                 */
                const buckets = [matchBucket, industryConnectorPlaceholderBucket];

                for (const bucket of buckets) {
                    const policy = bucket.node.tryFindChild('Policy');

                    if (policy) child.node.addDependency(policy);
                }

                break;
            }
        }
    }

    private addSuppressRules(resource: Resource | CfnResource, rules: CfnNagSuppressRule[]): void {
        if (resource instanceof Resource) {
            resource = <CfnResource>resource.node.defaultChild;
        }

        const cfnNagMetadata = resource.getMetadata('cfn_nag');

        if (cfnNagMetadata) {
            const existingRules = cfnNagMetadata.rules_to_suppress;

            if (Array.isArray(existingRules)) {
                for (const rule of existingRules) {
                    if (typeof rules.find(newRule => newRule.id === rule.id) === 'undefined') {
                        rules.push(rule);
                    }
                }
            }
        }

        resource.addMetadata('cfn_nag', {
            rules_to_suppress: rules
        });
    }

    private addGuardSuppressRules(resource: Resource | CfnResource, rules: string[]): void {
        if (resource instanceof Resource) {
            resource = <CfnResource>resource.node.defaultChild;
        }

        const cfnGuardMetadata = resource.getMetadata('guard');

        if (cfnGuardMetadata) {
            const existingRules = cfnGuardMetadata.SuppressedRules;

            if (Array.isArray(existingRules)) {
                for (const rule of existingRules) {
                    if (typeof rules.find(newRule => newRule === rule) === 'undefined') {
                        rules.push(rule);
                    }
                }
            }
        }

        resource.addMetadata('guard', {
            SuppressedRules: rules
        });
    }

    private suppressRdsWarnings(uptVpc: ec2.IVpc, storage: ProfileStorage): void {
        const cluster: rds.DatabaseCluster = storage.storageAuroraCluster;
        this.addGuardSuppressRules(cluster, ['RDS_CLUSTER_MASTER_USER_PASSWORD_NO_PLAINTEXT_PASSWORD']);
        this.addGuardSuppressRules(storage.lambdaToProxyGroup, ['SECURITY_GROUP_EGRESS_ALL_PROTOCOLS_RULE']);
        this.addGuardSuppressRules(storage.lambdaToProxyGroup, ['EC2_SECURITY_GROUP_EGRESS_OPEN_TO_WORLD_RULE']);

        (storage.lambdaToProxyGroup.node.defaultChild as ec2.CfnSecurityGroup).cfnOptions.metadata!.cfn_nag = {
            rules_to_suppress: [
                {
                    id: 'W28',
                    reason: 'Required'
                }
            ]
        };

        this.addGuardSuppressRules(storage.dbConnectionGroup, ['SECURITY_GROUP_MISSING_EGRESS_RULE']);

        (storage.dbConnectionGroup.node.defaultChild as ec2.CfnSecurityGroup).cfnOptions.metadata!.cfn_nag = {
            rules_to_suppress: [
                {
                    id: 'W28',
                    reason: 'Required'
                }
            ]
        };

        uptVpc.publicSubnets.forEach(subnet => {
            this.addGuardSuppressRules(<ec2.PublicSubnet>subnet, ['SUBNET_AUTO_ASSIGN_PUBLIC_IP_DISABLED']);
        });
        /*
        let writer: rds.ClusterInstance = <rds.IClusterInstance>storage.storageAuroraWriter
        this.addGuardSuppressRules(writer, ["RDS_MASTER_USER_PASSWORD_NO_PLAINTEXT_PASSWORD"])
        //let reader1: rds.IAuroraClusterInstance = storage.storageAuroraReader1<rds.IAuroraClusterInstance>
        //let reader2: rds.IAuroraClusterInstance = storage.storageAuroraReader2<rds.IAuroraClusterInstance>
        this.addGuardSuppressRules(writer, ["RDS_MASTER_USER_PASSWORD_NO_PLAINTEXT_PASSWORD"])
        this.addGuardSuppressRules(reader1, ["RDS_MASTER_USER_PASSWORD_NO_PLAINTEXT_PASSWORD"])
        this.addGuardSuppressRules(reader2, ["RDS_MASTER_USER_PASSWORD_NO_PLAINTEXT_PASSWORD"])
        this.addGuardSuppressRules(writer, ["RDS_STORAGE_ENCRYPTED"])
        this.addGuardSuppressRules(reader1, ["RDS_STORAGE_ENCRYPTED"])
        this.addGuardSuppressRules(reader2, ["RDS_STORAGE_ENCRYPTED"])*/
    }

    private suppressDeliveryStream(deliveryStream: firehose.CfnDeliveryStream): void {
        this.addGuardSuppressRules(deliveryStream, [
            'KINESIS_FIREHOSE_REDSHIFT_DESTINATION_CONFIGURATION_NO_PLAINTEXT_PASSWORD',
            'KINESIS_FIREHOSE_SPLUNK_DESTINATION_CONFIGURATION_NO_PLAINTEXT_PASSWORD'
        ]);
    }

    private suppressLambdaFunction(lambda: lambda.Function): void {
        const lambdaRules: CfnNagSuppressRule[] = [
            { id: 'W58', reason: 'Lambda functions do have permissions to write cloudwatch logs, through custom policy' },
            { id: 'W89', reason: 'Solution deployed locally, no need to encase in VPC. Not a general rule' },
            { id: 'F10', reason: 'Policy needed' },
            {
                id: 'W12',
                reason: 'Role uses minimum permissions and needs to communicate with client generated resources, such as ACCP domains'
            },
            { id: 'W92', reason: 'Impossible for us to define the correct concurrency for clients' }
        ];
        this.addSuppressRules(lambda, lambdaRules);

        if (lambda.role) {
            const lambdaRoleRules: CfnNagSuppressRule[] = [
                {
                    id: 'W12',
                    reason: 'Role uses minimum permissions and needs to communicate with client generated resources, such as ACCP domains'
                },
                { id: 'W76', reason: 'Permissions are needed for solution to work' }
            ];
            this.addGuardSuppressRules(<iam.Role>lambda.role, ['IAM_NO_INLINE_POLICY_CHECK']);

            lambda.role.node.children.forEach(element => {
                if (element.node.id === 'DefaultPolicy') {
                    this.addSuppressRules(<CfnResource>element.node.defaultChild, lambdaRoleRules);
                }
            });
        }
    }

    private suppressConstruct(lambda: Construct): void {
        const rules: CfnNagSuppressRule[] = [
            { id: 'W58', reason: 'This is a hidden resource created by an API call the solution itself has no control or access to' },
            { id: 'W89', reason: 'This is a hidden resource created by an API call the solution itself has no control or access to' },
            { id: 'W92', reason: 'This is a hidden resource created by an API call the solution itself has no control or access to' }
        ];
        this.addSuppressRules(<CfnResource>lambda.node.defaultChild, rules);

        lambda.node.children.forEach(element => {
            if (element.node.id === 'Role') {
                this.suppressRole(<iam.Role>element);
            }
        });
    }

    private suppressRole(role: iam.Role): void {
        const rules: CfnNagSuppressRule[] = [
            {
                id: 'W12',
                reason: 'Role uses minimum permissions and needs to communicate with client generated resources, such as ACCP domains'
            },
            { id: 'W76', reason: 'Permissions are needed for solution to work' }
        ];

        role.node.children.forEach(element => {
            if (element.node.id === 'DefaultPolicy') {
                this.addSuppressRules(<CfnResource>element.node.defaultChild, rules);
            }
        });
        this.addGuardSuppressRules(role, ['IAM_NO_INLINE_POLICY_CHECK']);
    }

    private suppressAccessLogBucket(bucket: tahS3.AccessLogBucket): void {
        const rules: CfnNagSuppressRule[] = [
            { id: 'W35', reason: 's3 bucket does have access logging configured, special wrapper for that purpose' }
        ];
        this.addSuppressRules(bucket, rules);
    }

    private suppressAccessLogHttp(httpStage: apiGatewayV2.HttpStage): void {
        const rules: CfnNagSuppressRule[] = [
            { id: 'W46', reason: 'access is logged through API widgets, which record metrics for API logging' }
        ];
        this.addSuppressRules(httpStage, rules);
    }

    private suppressDynamoDb(table: dynamodb.Table): void {
        const rules: CfnNagSuppressRule[] = [
            {
                id: 'W28',
                reason: 'Resource must have a name, and there is no workaround to renaming upon update, so that is what will be done'
            }
        ];
        this.addSuppressRules(table, rules);
        this.addGuardSuppressRules(table, ['CFN_NO_EXPLICIT_RESOURCE_NAMES']);
    }

    private suppressLog(log: logs.LogGroup): void {
        const rules: CfnNagSuppressRule[] = [
            { id: 'W84', reason: 'Local deployment, logs do not contain sensitive information, no need for encryption' }
        ];
        this.addSuppressRules(log, rules);
    }

    private suppressCloudfront(cloudfront: cloudfront.CloudFrontWebDistribution): void {
        const rules: CfnNagSuppressRule[] = [
            {
                id: 'W70',
                reason: 'Using default viewer certificate, cannot enforce higher TLS version without creating custom alias or certificate'
            }
        ];
        this.addSuppressRules(cloudfront, rules);
    }

    private suppressSqs(queue: tahSqs.Queue): void {
        const rules: CfnNagSuppressRule[] = [
            { id: 'W48', reason: 'Queues are encrypted using sqs managed encryption, no need to specify kms master key' }
        ];
        this.addSuppressRules(queue, rules);
    }
}
