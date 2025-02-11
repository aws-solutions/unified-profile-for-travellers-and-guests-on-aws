// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import { NodejsFunction } from 'aws-cdk-lib/aws-lambda-nodejs';
import path from 'path';
import * as tahCore from '../../tah-cdk-common/core';
import { CdkBase, CdkBaseProps, LambdaProps } from '../cdk-base';

type FrontEndConfig = {
    readonly cognitoClientId: string;
    readonly cognitoUserPoolId: string;
    readonly deployed: string;
    readonly staticContentBucketName: string;
    readonly ucpApiUrl: string;
    readonly usePermissionSystem: string;
};

type CustomResourceProps = {
    readonly frontEndConfig: FrontEndConfig;
    readonly solutionName: string;
} & LambdaProps;

export enum CustomActions {
    CREATE_UUID = 'CreateUuid',
    DEPLOY_FRONT_END = 'DeployFrontEnd'
}

export class CustomResource extends CdkBase {
    public customResourceLambda: lambda.Function;
    public uuid: string;

    private readonly props: CustomResourceProps;

    constructor(cdkProps: CdkBaseProps, props: CustomResourceProps) {
        super(cdkProps);

        this.props = props;

        // Custom resource Lambda function
        this.customResourceLambda = this.createLambdaFunction();
        const customResourceUuid = this.createCustomResource('CustomResourceUuid', {
            CustomAction: CustomActions.CREATE_UUID
        });

        // UUID custom resource
        this.uuid = customResourceUuid.getAttString('UUID');

        // Frontend asset custom resource
        this.createCustomResource('CustomResourceFrontEnd', {
            CustomAction: CustomActions.DEPLOY_FRONT_END,
            FrontEndConfig: {
                ...this.props.frontEndConfig,
                artifactBucketName: this.props.artifactBucket.bucketName,
                artifactBucketPath: this.props.artifactBucketPath + '/ui/dist.zip'
            }
        });

        // Outputs
        tahCore.Output.add(this.stack, 'deploymentId', this.uuid);
    }

    /**
     * Creates a custom resource Lambda function.
     * @returns Custom resource Lambda function
     */
    private createLambdaFunction() {
        return new NodejsFunction(this.stack, 'CustomResourceFunction', {
            entry: path.join(__dirname, '..', '..', 'custom_resource', 'index.ts'),
            runtime: lambda.Runtime.NODEJS_20_X,
            handler: 'index.handler',
            description: `${this.props.solutionName} (${this.props.solutionVersion}): Custom resource`,
            environment: {
                SOLUTION_ID: this.props.solutionId,
                SOLUTION_VERSION: this.props.solutionVersion,
                RETRY_SECONDS: '5'
            },
            memorySize: 128,
            timeout: cdk.Duration.minutes(1)
        });
    }

    /**
     * Creates a CloudFormation custom resource.
     * @param id Custom resource ID
     * @param props Custom resource properties
     * @returns CloudFormation custom resource
     */
    private createCustomResource(id: string, props?: Record<string, unknown>): cdk.CustomResource {
        return new cdk.CustomResource(this.stack, id, {
            serviceToken: this.customResourceLambda.functionArn,
            properties: props
        });
    }
}
