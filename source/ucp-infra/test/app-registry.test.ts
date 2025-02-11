// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { App, Stack } from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { AppRegistry } from '../lib/app-registry';
import { applicationName, applicationType, envName, solutionId, solutionName, solutionVersion } from './mock';

test('AppRegistry', () => {
    // Prepare
    const app = new App();
    const stack = new Stack(app, 'Stack');

    // Call
    new AppRegistry({ envName, stack }, { applicationName, applicationType, solutionId, solutionName, solutionVersion });

    // Verify
    const template = Template.fromStack(stack);
    template.resourceCountIs('AWS::ServiceCatalogAppRegistry::Application', 1);
    template.resourceCountIs('AWS::ServiceCatalogAppRegistry::ResourceAssociation', 1);
    template.resourceCountIs('AWS::ServiceCatalogAppRegistry::AttributeGroup', 1);
});
