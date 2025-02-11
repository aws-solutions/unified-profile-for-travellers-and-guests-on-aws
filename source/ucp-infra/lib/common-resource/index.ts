// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as glueAlpha from '@aws-cdk/aws-glue-alpha';
import * as cdk from 'aws-cdk-lib';
import * as athena from 'aws-cdk-lib/aws-athena';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as kms from 'aws-cdk-lib/aws-kms';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import glueSchemaTraveler from '../../../tah-common/tah-common-glue-schemas/traveller.glue.json';
import * as tahCore from '../../tah-cdk-common/core';
import * as tahS3 from '../../tah-cdk-common/s3';
import * as tahSqs from '../../tah-cdk-common/sqs';
import { CdkBase, CdkBaseProps, DynamoDbTableKey } from '../cdk-base';

type CommonResourceProps = {
    readonly outputBucketPrefix: string;
};

type DynamoDbOutput = {
    readonly configTable: dynamodb.Table;
    readonly portalConfigTable: dynamodb.Table;
    readonly privacySearchResultsTable: dynamodb.Table;
    readonly distributedMutexTable: dynamodb.Table;
};

type ConnectProfileOutput = {
    readonly connectProfileDomainErrorQueue: tahSqs.Queue;
    readonly connectProfileExportBucket: tahS3.Bucket;
    readonly connectProfileImportBucket: tahS3.Bucket;
};

/**
 * This class creates resources which either don't belong to anywhere or are commonly used by others.
 * For example, commonly shared data resources including Glue resources, DynamoDB tables and so on.
 */
export class CommonResource extends CdkBase {
    public readonly accessLogBucket: tahS3.AccessLogBucket;
    public readonly athenaResultsBucket: tahS3.Bucket;
    public readonly athenaWorkGroup: athena.CfnWorkGroup;
    public readonly athenaSearchProfileS3PathsPreparedStatement: athena.CfnPreparedStatement;
    public readonly athenaGetS3PathsByConnectIdsPreparedStatement: athena.CfnPreparedStatement;
    public readonly configTable: dynamodb.Table;
    public readonly connectProfileDomainErrorQueue: tahSqs.Queue;
    public readonly connectProfileExportBucket: tahS3.Bucket;
    public readonly connectProfileImportBucket: tahS3.Bucket;
    public readonly dataLakeAdminRole: iam.Role;
    public readonly dataSyncLogGroup: logs.LogGroup;
    public readonly gdprPurgeLogGroup: logs.LogGroup;
    public readonly dynamoDbKey: kms.Key;
    public readonly glueDatabase: glueAlpha.Database;
    public readonly glueTravelerTable: glueAlpha.S3Table;
    public readonly outputBucket: tahS3.Bucket;
    public readonly portalConfigTable: dynamodb.Table;
    public readonly privacySearchResultsTable: dynamodb.Table;

    //  This is necessary to prevent multiple Lambda invocations of the
    //  S3ExciseQueueProcessor from overwriting data to S3 from a previous
    //  run, and inadvertently re-add data that was supposed to be removed
    public readonly distributedMutexTable: dynamodb.Table;

    private readonly props: CommonResourceProps;

    constructor(cdkProps: CdkBaseProps, props: CommonResourceProps) {
        super(cdkProps);

        this.props = props;

        // S3 access log bucket
        this.accessLogBucket = new tahS3.AccessLogBucket(this.stack, 'ucp-access-logging');

        // Glue resources
        this.outputBucket = new tahS3.Bucket(this.stack, 'ucp-output-bucket', this.accessLogBucket);
        this.athenaResultsBucket = new tahS3.Bucket(this.stack, 'ucp-athena-output', this.accessLogBucket);
        this.athenaWorkGroup = this.createAthenaWorkGroup();
        this.dataLakeAdminRole = this.createDataLakeAdminRole();
        this.glueDatabase = this.createGlueDatabase();
        this.glueTravelerTable = this.createGlueTravelerTable();
        this.athenaSearchProfileS3PathsPreparedStatement = new athena.CfnPreparedStatement(this.stack, 'ucp-athena-get-profile-s3-paths', {
            queryStatement: `SELECT "$path" FROM "${this.glueDatabase.databaseName}"."${this.glueTravelerTable.tableName}" where connectid = ? and domainname = ? group by "$path"`,
            statementName: 'get_profile_s3_paths',
            workGroup: cdk.Fn.ref(this.athenaWorkGroup.logicalId)
        });
        this.athenaGetS3PathsByConnectIdsPreparedStatement = new athena.CfnPreparedStatement(
            this.stack,
            'ucp-athena-get-s3-paths-by-connect-ids',
            {
                queryStatement: ` SELECT "$path", connectid
                        FROM UNNEST(
                                CAST(JSON_PARSE(?) as ARRAY<VARCHAR>)
                        ) AS t(cId)
                          JOIN "${this.glueDatabase.databaseName}"."${this.glueTravelerTable.tableName}" on cId = connectid
                        WHERE domainname = ?
                        GROUP BY "$path", connectid`,
                statementName: 'get_s3_paths_by_connect_id',
                workGroup: cdk.Fn.ref(this.athenaWorkGroup.logicalId)
            }
        );

        // DynamoDB resources
        this.dynamoDbKey = this.createDynamoDbKey();
        const { configTable, portalConfigTable, privacySearchResultsTable, distributedMutexTable } = this.createDynamoDbTables();
        this.configTable = configTable;
        this.portalConfigTable = portalConfigTable;
        this.privacySearchResultsTable = privacySearchResultsTable;
        this.distributedMutexTable = distributedMutexTable;

        // Connect profile resources
        const { connectProfileDomainErrorQueue, connectProfileExportBucket, connectProfileImportBucket } =
            this.createConnectProfileResources();
        this.connectProfileDomainErrorQueue = connectProfileDomainErrorQueue;
        this.connectProfileExportBucket = connectProfileExportBucket;
        this.connectProfileImportBucket = connectProfileImportBucket;

        // DataSync log group
        this.dataSyncLogGroup = this.createDataSyncLogGroup();
        this.gdprPurgeLogGroup = new logs.LogGroup(this.stack, 'gdpr-purge-log-group', {
            retention: logs.RetentionDays.TEN_YEARS,
            logGroupClass: logs.LogGroupClass.INFREQUENT_ACCESS
        });

        // Permissions
        // Data admin role needs read/write permission to keep track of Glue job runs,
        // which lets us incrementally run jobs on new data (based on partition of last run)
        this.configTable.grantReadWriteData(this.dataLakeAdminRole);
        this.connectProfileImportBucket.grantReadWrite(this.dataLakeAdminRole);
        this.connectProfileExportBucket.grantReadWrite(this.dataLakeAdminRole);

        // Outputs
        tahCore.Output.add(this.stack, 'accessLogging', this.accessLogBucket.bucketName);
        tahCore.Output.add(this.stack, 'outputBucket', this.outputBucket.bucketName);
        tahCore.Output.add(this.stack, 'ucpDataAdminRoleArn', this.dataLakeAdminRole.roleArn);
        tahCore.Output.add(this.stack, 'ucpDataAdminRoleName', this.dataLakeAdminRole.roleName);
        tahCore.Output.add(this.stack, 'glueDBArn', this.glueDatabase.databaseArn);
        tahCore.Output.add(this.stack, 'glueDBname', this.glueDatabase.databaseName);
        tahCore.Output.add(this.stack, 'accpDomainErrorQueue', this.connectProfileDomainErrorQueue.queueUrl);
        tahCore.Output.add(this.stack, 'connectProfileImportBucketOut', this.connectProfileImportBucket.bucketName);
        tahCore.Output.add(this.stack, 'connectProfileExportBucket', this.connectProfileExportBucket.bucketName);
        tahCore.Output.add(this.stack, 'dynamodbConfigTableName', configTable.tableName);
        tahCore.Output.add(this.stack, 'dynamodbPortalConfigTableName', portalConfigTable.tableName);
    }

    /**
     * Creates an Athena work group.
     * @returns Athena work group
     */
    private createAthenaWorkGroup(): athena.CfnWorkGroup {
        return new athena.CfnWorkGroup(this.stack, 'ucp-athena-workgroup', {
            name: 'ucp-athena-workgroup-' + this.envName,
            recursiveDeleteOption: true,
            workGroupConfiguration: {
                publishCloudWatchMetricsEnabled: true,
                resultConfiguration: {
                    encryptionConfiguration: {
                        encryptionOption: 'SSE_S3'
                    },
                    outputLocation: `s3://${this.athenaResultsBucket.bucketName}`
                }
            }
        });
    }

    /**
     * Creates a data lake admin role.
     * @returns Data lake admin role
     */
    private createDataLakeAdminRole(): iam.Role {
        const dataLakeAdminRole = new iam.Role(this.stack, 'ucp-data-admin-role-' + this.envName, {
            assumedBy: new iam.ServicePrincipal('glue.amazonaws.com'),
            description: 'Glue role for UCP data',
            managedPolicies: [iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AWSGlueServiceRole')]
        });

        // Granting general logging
        dataLakeAdminRole.addToPolicy(
            new iam.PolicyStatement({
                resources: [
                    this.stack.formatArn({
                        arnFormat: cdk.ArnFormat.COLON_RESOURCE_NAME,
                        resource: 'log-group',
                        resourceName: '/*',
                        service: 'logs'
                    })
                ],
                actions: ['logs:CreateLogGroup', 'logs:CreateLogStream', 'logs:DescribeLogStreams', 'logs:PutLogEvents']
            })
        );

        dataLakeAdminRole.addToPolicy(
            new iam.PolicyStatement({
                resources: [dataLakeAdminRole.roleArn],
                actions: ['iam:PassRole']
            })
        );

        return dataLakeAdminRole;
    }

    /**
     * Creates a KMS key for all DynamoDB tables.
     * @returns DynamoDB KMS key
     */
    private createDynamoDbKey(): kms.Key {
        return new kms.Key(this.stack, 'ucpDynamoEncryptionKey', {
            description: 'KMS key for encrypting DynamoDB table',
            enableKeyRotation: true
        });
    }

    /**
     * Creates a Glue database.
     * @returns Glue database
     */
    private createGlueDatabase(): glueAlpha.Database {
        return new glueAlpha.Database(this.stack, 'ucp-data-glue-database-' + this.envName, {
            databaseName: 'ucp_db_' + this.envName
        });
    }

    /**
     * Creates a Glue table for traveler data.
     * @returns Glue traveler table
     */
    private createGlueTravelerTable(): glueAlpha.S3Table {
        const travelerSchema: glueAlpha.Column[] = glueSchemaTraveler.columns;
        return new glueAlpha.S3Table(this.stack, 'ucp-traveler-glue-table-' + this.envName, {
            columns: travelerSchema,
            database: this.glueDatabase,
            dataFormat: glueAlpha.DataFormat.PARQUET,
            bucket: this.outputBucket,
            compressed: true,
            description: 'Traveler data',
            partitionKeys: [{ name: 'domainname', type: glueAlpha.Schema.STRING }],
            s3Prefix: this.props.outputBucketPrefix,
            tableName: 'ucp_traveler_' + this.envName
        });
    }

    /**
     * Creates globally used DynamoDB tables including a config table and a portal config table.
     * @returns Config table and portal config table
     */
    private createDynamoDbTables(): DynamoDbOutput {
        const configTable = new dynamodb.Table(this.stack, 'ucpConfigTable', {
            tableName: 'ucp-config-table-' + this.envName,
            partitionKey: { name: DynamoDbTableKey.CONFIG_TABLE_PK, type: dynamodb.AttributeType.STRING },
            sortKey: { name: DynamoDbTableKey.CONFIG_TABLE_SK, type: dynamodb.AttributeType.STRING },
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            encryptionKey: this.dynamoDbKey,
            pointInTimeRecovery: true
        });

        // This table contains the user-defined configuration for the dashboard
        const portalConfigTable = new dynamodb.Table(this.stack, 'ucpPortalConfigTable', {
            tableName: 'ucp-portal-config-table-' + this.envName,
            partitionKey: { name: DynamoDbTableKey.PORTAL_CONFIG_TABLE_PK, type: dynamodb.AttributeType.STRING },
            sortKey: { name: DynamoDbTableKey.PORTAL_CONFIG_TABLE_SK, type: dynamodb.AttributeType.STRING },
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            pointInTimeRecovery: true,
            encryptionKey: this.dynamoDbKey
        });

        // This table contains privacy search results
        const privacySearchResultsTable = new dynamodb.Table(this.stack, 'ucpPrivacySearchResultsTable', {
            partitionKey: { name: DynamoDbTableKey.PRIVACY_SEARCH_RESULTS_TABLE_PK, type: dynamodb.AttributeType.STRING },
            sortKey: { name: DynamoDbTableKey.PRIVACY_SEARCH_RESULTS_TABLE_SK, type: dynamodb.AttributeType.STRING },
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            pointInTimeRecovery: true,
            encryptionKey: this.dynamoDbKey
        });

        // This table works as a distributed mutex (based on: https://github.com/awslabs/amazon-dynamodb-lock-client)
        const distributedMutexTable = new dynamodb.Table(this.stack, 'ucpDistributedMutexTable', {
            partitionKey: { name: DynamoDbTableKey.DISTRIBUTED_MUTEX_TABLE_PK, type: dynamodb.AttributeType.STRING },
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            pointInTimeRecovery: true,
            encryptionKey: this.dynamoDbKey,
            timeToLiveAttribute: DynamoDbTableKey.DISTRIBUTED_MUTEX_TABLE_TTL_ATTR
        });

        return { configTable, portalConfigTable, privacySearchResultsTable, distributedMutexTable };
    }

    /**
     * Creates Amazon Connect Customer Profiles resources including S3 buckets and an error queue.
     * Buckets are used to import and export data to/from Amazon Connect Customer Profiles.
     * A queue is used to send errors to when any error happens during manipulating data in Amazon Connect Customer Profiles.
     * @returns Amazon Connect Customer Profiles S3 buckets
     */
    private createConnectProfileResources(): ConnectProfileOutput {
        const connectProfileDomainErrorQueue = new sqs.Queue(this.stack, 'ucp-acc-domain-errors-' + this.envName, {
            enforceSSL: true,
            //we need SQS managed key here to ensure CP can use it (if using KMS_MANAGED, the key needs to be given additional policy)
            encryption: sqs.QueueEncryption.SQS_MANAGED
        });
        connectProfileDomainErrorQueue.addToResourcePolicy(
            new iam.PolicyStatement({
                actions: ['SQS:SendMessage'],
                effect: iam.Effect.ALLOW,
                principals: [new iam.ServicePrincipal('profile.amazonaws.com')],
                resources: [connectProfileDomainErrorQueue.queueArn]
            })
        );

        // Target Bucket for Amazon Connect Customer Profiles export
        const connectProfileExportBucket = new tahS3.Bucket(this.stack, 'ucp-mazon-connect-profile-export', this.accessLogBucket);
        connectProfileExportBucket.addToResourcePolicy(
            new iam.PolicyStatement({
                actions: ['s3:GetBucketLocation', 's3:GetBucketPolicy', 's3:GetObject', 's3:ListBucket', 's3:PutObject'],
                effect: iam.Effect.ALLOW,
                principals: [new iam.ServicePrincipal('appflow.amazonaws.com')],
                resources: [connectProfileExportBucket.arnForObjects('*'), connectProfileExportBucket.bucketArn]
            })
        );

        // Amazon Connect Customer Profiles import bucket where all batch processing sends data to
        const connectProfileImportBucket = new tahS3.Bucket(this.stack, 'connectProfileImportBucket', this.accessLogBucket);
        connectProfileImportBucket.addToResourcePolicy(
            new iam.PolicyStatement({
                actions: ['s3:GetBucketLocation', 's3:GetBucketPolicy', 's3:GetObject', 's3:ListBucket', 's3:PutObject'],
                effect: iam.Effect.ALLOW,
                principals: [new iam.ServicePrincipal('appflow.amazonaws.com')],
                resources: [connectProfileImportBucket.arnForObjects('*'), connectProfileImportBucket.bucketArn]
            })
        );

        return { connectProfileDomainErrorQueue, connectProfileExportBucket, connectProfileImportBucket };
    }

    /**
     * Creates a DataSync CloudWatch logs log group.
     * @returns DataSync CloudWatch logs log group
     */
    private createDataSyncLogGroup(): logs.LogGroup {
        return new logs.LogGroup(this.stack, 'datasyncLogGroup', {
            logGroupName: '/aws/datasync/ucpConnectorTransfer' + this.envName,
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            retention: logs.RetentionDays.TEN_YEARS
        });
    }
}
