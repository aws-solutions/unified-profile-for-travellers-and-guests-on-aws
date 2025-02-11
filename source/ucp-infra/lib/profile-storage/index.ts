// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as cdk from 'aws-cdk-lib';
import * as dynamo from 'aws-cdk-lib/aws-dynamodb';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as kms from 'aws-cdk-lib/aws-kms';
import * as rds from 'aws-cdk-lib/aws-rds';
import { Output } from '../../tah-cdk-common/core';
import { CdkBase, CdkBaseProps } from '../cdk-base';
import { UptVpc } from '../vpc/vpc-index';

type ProfileStorageProps = {
    readonly uptVpc: UptVpc;
    readonly dynamoDbKey: kms.Key;
};

export type ProfileStorageOutput = {
    storageConfigTable: dynamo.Table;
    storageConfigTablePk: string;
    storageConfigTableSk: string;
    lambdaToProxyGroup: ec2.SecurityGroup;
    storageDbName: string;
    storageDb: rds.DatabaseCluster;
    storageProxy: rds.DatabaseProxy | null;
    storageProxyEndpoint: string;
    storageProxyReadonlyEndpoint: string;
    storageSecretArn: string;
};

export type DBClusterOutput = {
    cluster: rds.DatabaseCluster;
    writer: rds.IClusterInstance;
    reader1: rds.IClusterInstance;
    reader2: rds.IClusterInstance;
};

export class ProfileStorage extends CdkBase {
    public storageConfigTable: dynamo.Table;
    public storageConfigTablePk: string;
    public storageConfigTableSk: string;
    public storageAuroraCluster: rds.DatabaseCluster;
    public storageAuroraWriter: rds.IClusterInstance;
    public storageAuroraReader1: rds.IClusterInstance;
    public storageAuroraReader2: rds.IClusterInstance;
    public storageRdsProxy: rds.DatabaseProxy | null;
    public storageRdsProxyReadonlyEndpoint: string | null;
    public dbName: string;
    public dbConnectionGroup: ec2.SecurityGroup;
    public lambdaToProxyGroup: ec2.SecurityGroup;
    public secretManagerEndpoint: ec2.InterfaceVpcEndpoint;

    private rdsKey: kms.Key;
    private props: ProfileStorageProps;

    constructor(cdkProps: CdkBaseProps, props: ProfileStorageProps) {
        super(cdkProps);

        this.props = props;

        // Config DynamoDB table
        this.storageConfigTablePk = 'pk';
        this.storageConfigTableSk = 'sk';
        this.storageConfigTable = this.createConfigTable(this.storageConfigTablePk, this.storageConfigTableSk);

        // Security groups for Aurora cluster
        this.dbConnectionGroup = this.createDbConnectionGroup();
        this.lambdaToProxyGroup = this.createLambdaToProxyGroup();

        // KMS key for Aurora cluster
        this.rdsKey = this.createRdsKey();

        // Aurora cluster for the storage
        this.dbName = 'ucp' + this.envName;
        //remove all non alphanumeric char
        this.dbName = this.dbName.replace(/[^a-zA-Z0-9]/g, '');
        const clusterOutput = this.createAuroraCluster(this.dbName);
        this.storageAuroraCluster = clusterOutput.cluster;
        this.storageAuroraWriter = clusterOutput.writer;
        this.storageAuroraReader1 = clusterOutput.reader1;
        this.storageAuroraReader2 = clusterOutput.reader2;

        // RDS Proxy
        this.storageRdsProxy = this.createRdsProxy(this.dbName);
        this.storageRdsProxyReadonlyEndpoint = this.createReadonlyRdsProxyEndpoint(this.storageRdsProxy);

        // Permissions
        this.dbConnectionGroup.addIngressRule(this.dbConnectionGroup, ec2.Port.tcp(5432), 'allow database connection');
        this.dbConnectionGroup.addIngressRule(this.lambdaToProxyGroup, ec2.Port.tcp(5432), 'allow Lambda connection');
        this.dbConnectionGroup.addEgressRule(this.lambdaToProxyGroup, ec2.Port.tcp(5432), 'allow Lambda connection (egress)');
        this.dbConnectionGroup.addEgressRule(this.dbConnectionGroup, ec2.Port.tcp(5432), 'allow db connection (egress)');

        // Outputs
        Output.add(this.stack, 'lcsAuroraProxyEndpoint', this.storageRdsProxy.endpoint);
        Output.add(this.stack, 'lcsAuroraDbName', this.dbName);
        Output.add(this.stack, 'lcsAuroraDbSecretArn', this.storageAuroraCluster.secret!.secretArn);
        Output.add(this.stack, 'lcsConfigTableName', this.storageConfigTable.tableName);
        Output.add(this.stack, 'lcsConfigTablePk', this.storageConfigTablePk);
        Output.add(this.stack, 'lcsConfigTableSk', this.storageConfigTableSk);
    }

    /**
     * Creates a DynamoDB table for storage config
     * @param pk partition key
     * @param sk sort key
     * @returns DynamoDB table for storage config
     */
    private createConfigTable(pk: string, sk: string): dynamo.Table {
        return new dynamo.Table(this.stack, 'ucpStorageConfigTable', {
            tableName: 'ucp-storage-config-' + this.envName,
            partitionKey: { name: pk, type: dynamo.AttributeType.STRING },
            sortKey: { name: sk, type: dynamo.AttributeType.STRING },
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            billingMode: dynamo.BillingMode.PAY_PER_REQUEST,
            pointInTimeRecovery: true,
            encryptionKey: this.props.dynamoDbKey
        });
    }

    /**
     * Creates a Security Group for connecting to the storage VPC
     * @returns Security Group for connecting to the storage VPC
     */
    private createDbConnectionGroup(): ec2.SecurityGroup {
        return new ec2.SecurityGroup(this.stack, 'Proxy to DB Connection', {
            securityGroupName: 'ucp-db-connection-' + this.envName,
            description: 'Allows proxy to connect to the database',
            vpc: this.props.uptVpc.uptVpc,
            allowAllOutbound: false
        });
    }

    /**
     * Creates a Security Group for connecting Lambda to RDS Proxy
     * @returns Security Group for connecting Lambda to RDS Proxy
     */
    private createLambdaToProxyGroup(): ec2.SecurityGroup {
        return new ec2.SecurityGroup(this.stack, 'Lambda to Proxy Connection', {
            securityGroupName: 'ucp-lambda-proxy-connection-' + this.envName,
            description: 'Allows Lambda to connect to the proxy',
            vpc: this.props.uptVpc.uptVpc
        });
    }

    /**
     * Creates a KMS key for RDS cluster.
     * @returns KMS key
     */
    private createRdsKey(): kms.Key {
        return new kms.Key(this.stack, 'ucpRdsEncryptionKey', {
            alias: 'ucp-rds-encryption-key-' + this.envName,
            description: 'KMS key for encrypting RDS cluster',
            enableKeyRotation: true
        });
    }

    /**
     * Creates a PostgreSQL Aurora cluster
     * @returns RDS Database Cluster
     */
    private createAuroraCluster(dbName: string): DBClusterOutput {
        const auroraDbName = dbName;
        const writer = rds.ClusterInstance.serverlessV2('writer', { publiclyAccessible: false });
        const reader1 = rds.ClusterInstance.serverlessV2('reader1', { scaleWithWriter: true, publiclyAccessible: false });
        const reader2 = rds.ClusterInstance.serverlessV2('reader2', { publiclyAccessible: false });
        const cluster = new rds.DatabaseCluster(this.stack, 'ucpDbCluster' + dbName, {
            defaultDatabaseName: auroraDbName,
            engine: rds.DatabaseClusterEngine.auroraPostgres({
                version: rds.AuroraPostgresEngineVersion.VER_15_5
            }),
            serverlessV2MinCapacity: 1,
            serverlessV2MaxCapacity: 128,
            storageEncrypted: true,
            storageEncryptionKey: this.rdsKey,
            deletionProtection: true,
            vpc: this.props.uptVpc.uptVpc,
            vpcSubnets: {
                subnets: this.props.uptVpc.getPrivateSubnetIds().map(subnetId => ec2.Subnet.fromSubnetId(this.stack, subnetId, subnetId))
            },
            credentials: rds.Credentials.fromGeneratedSecret('postgres', { encryptionKey: this.rdsKey }),
            securityGroups: [this.dbConnectionGroup],
            writer: writer,
            //withing a transaction we want a short timeout (30s) to avoid deadlock
            //target latency for profile insert is <1s p99 so 30s has a good buffer
            parameters: { idle_in_transaction_session_timeout: '30000' },
            readers: [
                // will be put in promotion tier 1 and will scale with the writer
                reader1,
                // will be put in promotion tier 2 and will not scale with the writer
                reader2
            ]
        });
        return { cluster, writer, reader1, reader2 };
    }

    /**
     * Creates an RDS Proxy for managed database connections
     * @returns RDS Proxy
     */
    private createRdsProxy(suffix: string): rds.DatabaseProxy {
        return new rds.DatabaseProxy(this.stack, 'ucpDbProxy' + suffix, {
            proxyTarget: rds.ProxyTarget.fromCluster(this.storageAuroraCluster),
            secrets: [this.storageAuroraCluster.secret!],
            vpc: this.props.uptVpc.uptVpc,
            securityGroups: [this.dbConnectionGroup]
        });
    }

    /**
     * RDS Proxy CDK does not directly support creating a readonly endpoint. This will manually add a readonly endpoint with CloudFormation.
     * @param proxy RDS Proxy that we will add a readonly endpoint to
     * @returns readonly endpoint
     */
    private createReadonlyRdsProxyEndpoint(proxy: rds.DatabaseProxy): string {
        const endpoint = new rds.CfnDBProxyEndpoint(this.stack, 'ProxyReaderEndpoint', {
            dbProxyEndpointName: `ucp-reader-${this.envName}`,
            dbProxyName: proxy.dbProxyName,
            targetRole: 'READ_ONLY',
            vpcSubnetIds: this.props.uptVpc.getPrivateSubnetIds(),
            vpcSecurityGroupIds: [this.dbConnectionGroup.securityGroupId]
        });

        return endpoint.attrEndpoint;
    }
}
