// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as cdk from 'aws-cdk-lib';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as firehose from 'aws-cdk-lib/aws-kinesisfirehose';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as s3 from 'aws-cdk-lib/aws-s3';
import { IConstruct } from 'constructs';

export type CdkBaseProps = {
    readonly envName: string;
    readonly stack: cdk.Stack;
};

export type LambdaProps = {
    readonly artifactBucket: s3.IBucket;
    readonly artifactBucketPath: string;
    // Usage metrics
    readonly sendAnonymizedData: string;
    readonly solutionId: string;
    readonly solutionVersion: string;
};

export type FirehoseDynamicPartition = {
    readonly jqString: string;
    readonly name: string;
};

export type KinesisFirehoseProps = {
    readonly bucket: s3.IBucket;
    readonly bucketPrefix: string;
    readonly idPrefix: string;
    readonly idSuffix: string;
    readonly partition: FirehoseDynamicPartition;
    readonly dataFormatConversionConfiguration?: firehose.CfnDeliveryStream.DataFormatConversionConfigurationProperty;
    readonly resourceCondition?: cdk.CfnCondition;
};

export type FirehoseOutput = {
    readonly deliveryStream: firehose.CfnDeliveryStream;
    readonly logGroup: logs.LogGroup;
};

export const UIDeploymentOptions = {
    CloudFront: 'CloudFront',
    ECS: 'ECS',
    Headless: 'Headless'
} as const;

export enum DynamoDbTableKey {
    CONFIG_TABLE_PK = 'item_id',
    CONFIG_TABLE_SK = 'item_type',
    ERROR_RETRY_TABLE_PK = 'traveller_id',
    ERROR_RETRY_TABLE_SK = 'object_id',
    ERROR_TABLE_PK = 'error_type',
    ERROR_TABLE_SK = 'error_id',
    MATCH_GSI_PK = 'runId',
    MATCH_GSI_SK = 'scoreTargetId',
    MATCH_TABLE_PK = 'domain_sourceProfileId',
    MATCH_TABLE_SK = 'match_targetProfileId',
    PORTAL_CONFIG_TABLE_PK = 'config_item',
    PORTAL_CONFIG_TABLE_SK = 'config_item_category',
    PRIVACY_SEARCH_RESULTS_TABLE_PK = 'domainName',
    PRIVACY_SEARCH_RESULTS_TABLE_SK = 'connectId',
    DISTRIBUTED_MUTEX_TABLE_PK = 'pk',
    DISTRIBUTED_MUTEX_TABLE_TTL_ATTR = 'ttl',
    CP_INDEX_TABLE_PK = 'domainName',
    CP_INDEX_TABLE_SK = 'connectId'
}

export class CdkBase {
    protected envName: string;
    protected stack: cdk.Stack;

    /**
     * Generally, `CdkBase` has an environment name and a CDK stack as properties.
     * - Environment name: Environment name is to distinguish which environment a stack is for.
     * - CDK stack: As resources are created by the stack itself, the CDK stack should be passed.
     *
     * The class is meant to be inherited by other classes and provide common resource creations functions.
     * As the class is not meant to be used standalone, all resources are `protected` so classes inheriting this can access.
     * @param props Environment name and CDK stack
     */
    constructor(props: CdkBaseProps) {
        this.envName = props.envName;
        this.stack = props.stack;
    }

    /**
     * Creates a Kinesis Firehose delivery stream.
     * @returns Kinesis Firehose delivery stream
     */
    protected createFirehoseDeliveryStream(props: KinesisFirehoseProps): FirehoseOutput {
        const logGroup = new logs.LogGroup(this.stack, props.idPrefix + 'FirehoseLogGroup' + props.idSuffix, {
            retention: logs.RetentionDays.TEN_YEARS
        });

        const firehoseRole = new iam.Role(this.stack, props.idPrefix + 'firhoseRoleArn' + props.idSuffix, {
            assumedBy: new iam.ServicePrincipal('firehose.amazonaws.com')
        });
        const logGroupPolicy = new iam.Policy(this.stack, props.idPrefix + 'FirehoseLogGroupPolicy' + props.idSuffix, {
            statements: [
                new iam.PolicyStatement({
                    effect: iam.Effect.ALLOW,
                    actions: ['logs:CreateLogStream', 'logs:PutLogEvents'],
                    resources: [logGroup.logGroupArn]
                }),
                new iam.PolicyStatement({
                    effect: iam.Effect.ALLOW,
                    actions: [
                        's3:GetObject*',
                        's3:GetBucket*',
                        's3:List*',
                        's3:DeleteObject*',
                        's3:PutObject',
                        's3:PutObjectLegalHold',
                        's3:PutObjectRetention',
                        's3:PutObjectTagging',
                        's3:PutObjectVersionTagging',
                        's3:Abort*'
                    ],
                    resources: [props.bucket.bucketArn, props.bucket.arnForObjects('*')]
                })
            ]
        });
        logGroupPolicy.attachToRole(firehoseRole);
        const cfnLogGroupPolicy = logGroupPolicy.node.defaultChild as iam.CfnPolicy;
        cfnLogGroupPolicy.cfnOptions.condition = props.resourceCondition;

        const cfnFirehoseRole = firehoseRole.node.defaultChild as iam.CfnRole;
        cfnFirehoseRole.cfnOptions.condition = props.resourceCondition;

        const cfnLogGroup = logGroup.node.defaultChild as logs.CfnLogGroup;
        cfnLogGroup.cfnOptions.condition = props.resourceCondition;

        const deliveryStream = new firehose.CfnDeliveryStream(this.stack, props.idPrefix + 'firehosDeliveryStream' + props.idSuffix, {
            deliveryStreamName: props.idPrefix + 'firehosDeliveryStream' + props.idSuffix,
            deliveryStreamEncryptionConfigurationInput: {
                keyType: 'AWS_OWNED_CMK'
            },
            extendedS3DestinationConfiguration: {
                bucketArn: props.bucket.bucketArn,
                roleArn: firehoseRole.roleArn,
                dataFormatConversionConfiguration: props.dataFormatConversionConfiguration,
                dynamicPartitioningConfiguration: {
                    enabled: true
                },
                prefix: `${props.bucketPrefix}/${props.partition.name}=!{partitionKeyFromQuery:${props.partition.name}}/!{timestamp:yyyy/MM/dd}/`,
                errorOutputPrefix: 'error/!{firehose:error-output-type}/',
                bufferingHints: {
                    intervalInSeconds: 60
                },
                processingConfiguration: {
                    enabled: true,
                    processors: [
                        {
                            type: 'MetadataExtraction',
                            parameters: [
                                {
                                    parameterName: 'MetadataExtractionQuery',
                                    parameterValue: `{${props.partition.name}: ${props.partition.jqString}}`
                                },
                                {
                                    parameterName: 'JsonParsingEngine',
                                    parameterValue: 'JQ-1.6'
                                }
                            ]
                        },
                        {
                            type: 'AppendDelimiterToRecord',
                            parameters: [
                                {
                                    parameterName: 'Delimiter',
                                    parameterValue: '\\n'
                                }
                            ]
                        }
                    ]
                }
            }
        });
        deliveryStream.cfnOptions.condition = props.resourceCondition;

        return { deliveryStream, logGroup };
    }
}

export class ConditionalResource implements cdk.IAspect {
    public constructor(public readonly condition: cdk.CfnCondition) {}
    public visit(node: IConstruct): void {
        if (node instanceof cdk.CfnResource) {
            node.cfnOptions.condition = this.condition;
        }
    }
}

export class ConditionalPermissionsBoundary implements cdk.IAspect {
    public constructor(
        public readonly condition: cdk.CfnCondition,
        public readonly pbArn: string
    ) {}
    public visit(node: IConstruct): void {
        if (node instanceof iam.CfnRole) {
            if (node.permissionsBoundary !== undefined) {
                throw new Error('Role already has a permission boundary');
            }
            node.permissionsBoundary = cdk.Fn.conditionIf(this.condition.logicalId, this.pbArn, cdk.Aws.NO_VALUE).toString();
        }
    }
}
