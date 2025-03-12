// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as cdk from 'aws-cdk-lib';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { Cluster, ContainerImage, FargateService, FargateTaskDefinition, LogDrivers, OperatingSystemFamily } from 'aws-cdk-lib/aws-ecs';
import * as iam from 'aws-cdk-lib/aws-iam';
import { RetentionDays } from 'aws-cdk-lib/aws-logs';
import * as s3 from 'aws-cdk-lib/aws-s3';
import { Construct } from 'constructs';
import path from 'path';
import * as tahCore from '../../tah-cdk-common/core';
import * as tahS3 from '../../tah-cdk-common/s3';
import { CdkBase, CdkBaseProps, ConditionalResource } from '../cdk-base';
import { SuppressLogGroup } from '../fargate-resource/fargate-resource';
import { EcrProps } from '../ucp-infra-stack';

export type FrontendProps = {
    readonly accessLogBucket: s3.Bucket;
    readonly deployFrontendToCloudFrontCondition: cdk.CfnCondition;
    readonly deployFrontendToEcsCondition: cdk.CfnCondition;
    readonly uptVpc: ec2.IVpc;
    readonly ecrProps: EcrProps;
    readonly ucpApiUrl: string;
    readonly cognitoUserPoolId: string;
    readonly cognitoClientId: string;
    readonly usePermissionSystem: string;
};

export class Frontend extends CdkBase {
    public readonly websiteDistribution: cloudfront.CloudFrontWebDistribution;
    public readonly websiteStaticContentBucket: s3.Bucket;

    private oai: cloudfront.OriginAccessIdentity;
    private props: FrontendProps;

    constructor(cdkProps: CdkBaseProps, props: FrontendProps) {
        super(cdkProps);
        this.props = props;

        this.oai = new cloudfront.OriginAccessIdentity(this.stack, 'websiteDistributionOAI' + this.envName, {
            comment: 'Origin access identity for website in ' + this.envName
        });
        this.websiteStaticContentBucket = this.createFrontendBucket();
        this.websiteDistribution = this.createCloudFrontDistribution();
        this.createWebsiteEcr(cdkProps.stack);

        // Outputs
        tahCore.Output.add(this.stack, 'websiteBucket', this.websiteStaticContentBucket.bucketName);
        tahCore.Output.add(this.stack, 'websiteDistributionId', this.websiteDistribution.distributionId).condition =
            this.props.deployFrontendToCloudFrontCondition;
        tahCore.Output.add(this.stack, 'websiteDomainName', this.websiteDistribution.distributionDomainName).condition =
            this.props.deployFrontendToCloudFrontCondition;
    }

    /**
     * Creates a frontend content S3 bucket.
     * @returns Frontend content S3 bucket
     */
    private createFrontendBucket(): s3.Bucket {
        const websiteStaticContentBucket = new tahS3.Bucket(this.stack, 'ucp-connector-fe-' + this.envName, this.props.accessLogBucket);
        websiteStaticContentBucket.addToResourcePolicy(
            new iam.PolicyStatement({
                actions: ['s3:GetObject'],
                effect: iam.Effect.ALLOW,
                principals: [new iam.CanonicalUserPrincipal(this.oai.cloudFrontOriginAccessIdentityS3CanonicalUserId)],
                resources: [websiteStaticContentBucket.arnForObjects('*')]
            })
        );

        return websiteStaticContentBucket;
    }

    /**
     * Creates a frontend CloudFront distribution.
     * @returns Frontend CloudFront distribution
     */
    private createCloudFrontDistribution(): cloudfront.CloudFrontWebDistribution {
        const cloudfrontLogBucket = new tahS3.Bucket(this.stack, 'ucp-connector-fe-logs-' + this.envName, this.props.accessLogBucket, {
            objectOwnership: s3.ObjectOwnership.OBJECT_WRITER
        });
        const cloudfrontLogsBucketPolicy = cloudfrontLogBucket.node.tryFindChild('Policy');

        const distributionConfig: cloudfront.CloudFrontWebDistributionProps = {
            originConfigs: [
                {
                    s3OriginSource: {
                        s3BucketSource: this.websiteStaticContentBucket,
                        originAccessIdentity: this.oai
                    },
                    behaviors: [
                        {
                            isDefaultBehavior: true,
                            forwardedValues: {
                                queryString: true,
                                cookies: {
                                    forward: 'none'
                                }
                            }
                        }
                    ]
                }
            ],
            loggingConfig: {
                bucket: cloudfrontLogBucket,
                includeCookies: false,
                prefix: 'prefix'
            },
            viewerCertificate: cloudfront.ViewerCertificate.fromCloudFrontDefaultCertificate(),
            errorConfigurations: [
                {
                    errorCachingMinTtl: 300,
                    errorCode: 403,
                    responseCode: 200,
                    responsePagePath: '/index.html'
                },
                {
                    errorCachingMinTtl: 300,
                    errorCode: 404,
                    responseCode: 200,
                    responsePagePath: '/index.html'
                }
            ]
        };

        const websiteDistribution = new cloudfront.CloudFrontWebDistribution(
            this.stack,
            'ucpWebsiteDistribution' + this.envName,
            distributionConfig
        );

        if (cloudfrontLogsBucketPolicy) {
            websiteDistribution.node.addDependency(cloudfrontLogsBucketPolicy);
        }

        // This is for a dedicated response header policy to include the AWS recommended HTTP headers.
        const responseHeaderPolicy = new cloudfront.ResponseHeadersPolicy(this.stack, 'ResponseHeadersPolicy', {
            responseHeadersPolicyName: cdk.Fn.join('-', ['ucp-fe-response-headers-policy', cdk.Aws.REGION, this.envName]),
            customHeadersBehavior: {
                customHeaders: [
                    { header: 'Cache-Control', value: 'no-store, no-cache', override: true },
                    { header: 'Pragma', value: 'no-cache', override: false }
                ]
            },
            securityHeadersBehavior: {
                contentTypeOptions: { override: true },
                frameOptions: { frameOption: cloudfront.HeadersFrameOption.DENY, override: true },
                contentSecurityPolicy: {
                    contentSecurityPolicy:
                        "default-src 'none'; img-src 'self'; script-src 'self'; style-src 'unsafe-inline' 'self' ; object-src 'none'; connect-src *.amazonaws.com *.cloudfront.net",
                    override: true
                },
                strictTransportSecurity: { accessControlMaxAge: cdk.Duration.seconds(31536000), includeSubdomains: true, override: true }
            }
        });

        // Adding the response policy ID to CloudFront distribution using an escape hatch.
        const cfnDistribution = <cloudfront.CfnDistribution>websiteDistribution.node.defaultChild;
        cfnDistribution.addPropertyOverride(
            'DistributionConfig.DefaultCacheBehavior.ResponseHeadersPolicyId',
            responseHeaderPolicy.responseHeadersPolicyId
        );
        cdk.Aspects.of(websiteDistribution).add(new ConditionalResource(this.props.deployFrontendToCloudFrontCondition));

        return websiteDistribution;
    }

    private createWebsiteEcr(scope: Construct) {
        const cluster = new Cluster(scope, 'websiteCluster', {
            vpc: this.props.uptVpc,
            enableFargateCapacityProviders: true
        });
        cdk.Aspects.of(cluster).add(new ConditionalResource(this.props.deployFrontendToEcsCondition));

        const taskRole = new iam.Role(scope, 'websiteEcrRole', {
            assumedBy: new iam.ServicePrincipal('ecs-tasks.amazonaws.com')
        });
        cdk.Aspects.of(taskRole).add(new ConditionalResource(this.props.deployFrontendToEcsCondition));

        const taskDefinition = new FargateTaskDefinition(scope, 'websiteTaskDefinition', {
            cpu: 256,
            memoryLimitMiB: 512,
            taskRole,
            runtimePlatform: {
                operatingSystemFamily: OperatingSystemFamily.LINUX
            }
        });
        cdk.Aspects.of(taskDefinition).add(new ConditionalResource(this.props.deployFrontendToEcsCondition));
        cdk.Aspects.of(taskDefinition).add(new SuppressLogGroup());

        let image: ContainerImage | undefined;
        if (this.props.ecrProps.publishEcrAssets) {
            // under normal circumstances, build the image asset
            image = ContainerImage.fromAsset(path.join(__dirname, '..', '..', '..'), {
                file: 'Dockerfile.frontend'
            });
        } else {
            // in the solutions pipeline, the asset is built by the tooling and must be pulled from a public registry
            image = ContainerImage.fromRegistry(`${this.props.ecrProps.publicEcrRegistry}/upt-fe:${this.props.ecrProps.publicEcrTag}`);
        }

        taskDefinition.addContainer('websiteContainer', {
            image,
            logging: LogDrivers.awsLogs({
                streamPrefix: 'upt-fe-ecs',
                logRetention: RetentionDays.TEN_YEARS
            }),
            environment: {
                UCP_REGION: this.stack.region,
                UCP_API_URL: this.props.ucpApiUrl,
                UCP_COGNITO_POOL_ID: this.props.cognitoUserPoolId,
                UCP_COGNITO_CLIENT_ID: this.props.cognitoClientId,
                USE_PERMISSION_SYSTEM: this.props.usePermissionSystem
            },
            portMappings: [
                {
                    containerPort: 8080,
                    hostPort: 8080
                }
            ]
        });

        const websiteService = new FargateService(scope, 'websiteService', {
            cluster,
            taskDefinition,
            desiredCount: 1
        });
        cdk.Aspects.of(websiteService).add(new ConditionalResource(this.props.deployFrontendToEcsCondition));

        websiteService.node.children.forEach(child => {
            child.node.children.forEach(grandchild => {
                if (grandchild instanceof ec2.CfnSecurityGroup) {
                    grandchild.cfnOptions.metadata = {
                        cfn_nag: {
                            rules_to_suppress: [
                                {
                                    id: 'W5',
                                    reason: 'Not applicable'
                                },
                                {
                                    id: 'W40',
                                    reason: 'Not applicable'
                                }
                            ]
                        }
                    };
                }
            });
        });
    }
}
