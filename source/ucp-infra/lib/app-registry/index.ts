// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as appreg from '@aws-cdk/aws-servicecatalogappregistry-alpha';
import * as cdk from 'aws-cdk-lib';
import { CdkBase, CdkBaseProps } from '../cdk-base';

type AppRegistryProps = {
    readonly applicationName: string;
    readonly applicationType: string;
    readonly solutionId: string;
    readonly solutionName: string;
    readonly solutionVersion: string;
};

export class AppRegistry extends CdkBase {
    public readonly props: AppRegistryProps;

    constructor(cdkProps: CdkBaseProps, props: AppRegistryProps) {
        super(cdkProps);

        this.props = props;

        const application = this.createApplication();
        this.createAttributeGroup(application);
    }

    /**
     * Creates an AppRegistry application for the solution.
     * @returns AppRegistry application
     */
    private createApplication(): appreg.Application {
        /**
         * There is a character limit for this name.
         * Please refer to point # 5 in the above section "Use cases not supported with below setup" for more details.
         */
        const application = new appreg.Application(this.stack, 'AppRegistry', {
            applicationName: cdk.Fn.join('-', [
                this.props.applicationName,
                cdk.Aws.REGION,
                cdk.Aws.ACCOUNT_ID,
                cdk.Aws.STACK_NAME // If your solution supports multiple deployments in the same , add stack name to the application name to make it unique.
            ]),
            description: `Service Catalog application to track and manage all your resources for the solution ${this.props.solutionName}`
        });
        application.associateApplicationWithStack(this.stack);
        cdk.Tags.of(application).add('Solutions:SolutionID', this.props.solutionId);
        cdk.Tags.of(application).add('Solutions:SolutionName', this.props.solutionName);
        cdk.Tags.of(application).add('Solutions:SolutionVersion', this.props.solutionVersion);
        cdk.Tags.of(application).add('Solutions:ApplicationType', this.props.applicationType);

        return application;
    }

    /**
     * Creates an AppRegistry attribute group.
     * @param application AppRegistry application
     */
    private createAttributeGroup(application: appreg.Application): void {
        const attributeGroup = new appreg.AttributeGroup(this.stack, 'DefaultApplicationAttributes', {
            attributeGroupName: 'Tah' + cdk.Aws.STACK_NAME,
            description: 'Attribute group for solution information',
            attributes: {
                applicationType: this.props.applicationType,
                solutionID: this.props.solutionId,
                solutionName: this.props.solutionName,
                version: this.props.solutionVersion
            }
        });
        attributeGroup.associateWith(application);
    }
}
