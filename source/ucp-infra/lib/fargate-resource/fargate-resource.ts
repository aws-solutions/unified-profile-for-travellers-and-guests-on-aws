// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Aws } from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { Cluster, ContainerImage, CpuArchitecture, FargateTaskDefinition, LogDrivers, OperatingSystemFamily } from 'aws-cdk-lib/aws-ecs';
import { Effect, IRole, PolicyStatement, Role, ServicePrincipal } from 'aws-cdk-lib/aws-iam';
import { Stream } from 'aws-cdk-lib/aws-kinesis';
import { RetentionDays } from 'aws-cdk-lib/aws-logs';
import { Construct } from 'constructs';
import path from 'path';
import { ProfileStorageOutput } from '../profile-storage';
import { EcrProps } from '../ucp-infra-stack';

export interface FargateResourceProps {
    readonly profileStorageOutput: ProfileStorageOutput;
    readonly mergeQueueUrl: string;
    readonly cpWriterQueueUrl: string;
    readonly changeProcessorKinesisStream: Stream;
    readonly ecrProps: EcrProps;
    readonly solutionId: string;
    readonly solutionVersion: string;
    readonly uptVpc: ec2.IVpc;
    readonly logLevel: string;
}

interface FargateResourceOutput {
    readonly cluster: Cluster;
    readonly taskRole: IRole;
    readonly taskDefinition: FargateTaskDefinition;
}

enum FargateRequestType {
    ID_RES = 'id-res',
    REBUILD_CACHE = 'rebuild-cache'
}

export class FargateResource {
    readonly irResources: FargateResourceOutput;
    readonly rebuildCacheResources: FargateResourceOutput;

    readonly props: FargateResourceProps;

    constructor(scope: Construct, props: FargateResourceProps) {
        this.props = props;

        this.irResources = this.createFargateResources(scope, 'IdentityResolution');
        this.rebuildCacheResources = this.createFargateResources(scope, 'RebuildCache');

        let image: ContainerImage | undefined;
        if (props.ecrProps.publishEcrAssets) {
            // under normal circumstances, build the image asset
            image = ContainerImage.fromAsset(path.join(__dirname, '..', '..', '..', '..'));
        } else {
            // in the solutions pipeline, the asset is built by the tooling and must be pulled from a public registry
            image = ContainerImage.fromRegistry(`${props.ecrProps.publicEcrRegistry}/upt-id-res:${props.ecrProps.publicEcrTag}`);
        }

        this.irResources.taskDefinition.addContainer('IdentityResolutionContainer', {
            image,
            logging: LogDrivers.awsLogs({ streamPrefix: 'upt-ir', logRetention: RetentionDays.TEN_YEARS }),
            environment: {
                // Solution metrics
                METRICS_SOLUTION_ID: props.solutionId,
                METRICS_SOLUTION_VERSION: props.solutionVersion,
                // For low-cost storage
                AURORA_PROXY_ENDPOINT: props.profileStorageOutput.storageProxyEndpoint,
                AURORA_DB_NAME: props.profileStorageOutput.storageDbName,
                AURORA_DB_SECRET_ARN: props.profileStorageOutput.storageSecretArn,
                STORAGE_CONFIG_TABLE_NAME: props.profileStorageOutput.storageConfigTable?.tableName ?? '',
                STORAGE_CONFIG_TABLE_PK: props.profileStorageOutput.storageConfigTablePk,
                STORAGE_CONFIG_TABLE_SK: props.profileStorageOutput.storageConfigTableSk,
                CHANGE_PROC_KINESIS_STREAM_NAME: props.changeProcessorKinesisStream.streamName,
                // Merge Queue
                MERGE_QUEUE_URL: props.mergeQueueUrl,
                // CP Writer Queue
                CP_WRITER_QUEUE_URL: props.cpWriterQueueUrl,
                // Image Identifier
                TASK_REQUEST_TYPE: FargateRequestType.ID_RES,
                LOG_LEVEL: props.logLevel
            }
        });

        this.rebuildCacheResources.taskDefinition.addContainer('RebuildCacheContainer', {
            image,
            logging: LogDrivers.awsLogs({ streamPrefix: 'upt-cache', logRetention: RetentionDays.TEN_YEARS }),
            environment: {
                // Solution metrics
                METRICS_SOLUTION_ID: props.solutionId,
                METRICS_SOLUTION_VERSION: props.solutionVersion,
                // For low-cost storage
                AURORA_PROXY_ENDPOINT: props.profileStorageOutput.storageProxyEndpoint,
                AURORA_DB_NAME: props.profileStorageOutput.storageDbName,
                AURORA_DB_SECRET_ARN: props.profileStorageOutput.storageSecretArn,
                STORAGE_CONFIG_TABLE_NAME: props.profileStorageOutput.storageConfigTable?.tableName ?? '',
                STORAGE_CONFIG_TABLE_PK: props.profileStorageOutput.storageConfigTablePk,
                STORAGE_CONFIG_TABLE_SK: props.profileStorageOutput.storageConfigTableSk,
                CHANGE_PROC_KINESIS_STREAM_NAME: props.changeProcessorKinesisStream.streamName,
                // Merge Queue
                MERGE_QUEUE_URL: props.mergeQueueUrl,
                // CP Writer Queue
                CP_WRITER_QUEUE_URL: props.cpWriterQueueUrl,
                // Image Identifier
                TASK_REQUEST_TYPE: FargateRequestType.REBUILD_CACHE,
                LOG_LEVEL: props.logLevel
            }
        });

        // Low-cost storage permissions
        // Tables are created/deleted via API and cannot be scoped down
        this.rebuildCacheResources.taskRole.addToPrincipalPolicy(
            new PolicyStatement({
                actions: [
                    'dynamodb:CreateTable',
                    'dynamodb:DescribeTable',
                    'dynamodb:DeleteTable',
                    'dynamodb:Query',
                    'dynamodb:Scan',
                    'dynamodb:BatchWriteItem',
                    'dynamodb:PutItem'
                ],
                effect: Effect.ALLOW,
                resources: [`arn:${Aws.PARTITION}:dynamodb:${Aws.REGION}:${Aws.ACCOUNT_ID}:table/ucp_domain_*`]
            })
        );

        this.rebuildCacheResources.taskRole.addToPrincipalPolicy(
            new PolicyStatement({
                actions: ['profile:PutProfileObject'],
                effect: Effect.ALLOW,
                resources: [`arn:${Aws.PARTITION}:profile:${Aws.REGION}:${Aws.ACCOUNT_ID}:domains/*`]
            })
        );
    }

    private createFargateResources(scope: Construct, resourceName: string): FargateResourceOutput {
        const cluster = new Cluster(scope, resourceName + 'Cluster', {
            vpc: this.props.uptVpc,
            enableFargateCapacityProviders: true
        });

        const taskRole = new Role(scope, resourceName + 'Role', {
            assumedBy: new ServicePrincipal('ecs-tasks.amazonaws.com')
        });

        const taskDefinition = new FargateTaskDefinition(scope, resourceName + 'TaskDefinition', {
            taskRole: taskRole,
            runtimePlatform: { cpuArchitecture: CpuArchitecture.ARM64, operatingSystemFamily: OperatingSystemFamily.LINUX }
        });
        return { cluster, taskRole, taskDefinition };
    }
}
