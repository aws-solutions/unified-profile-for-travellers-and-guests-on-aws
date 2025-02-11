// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Stack } from 'aws-cdk-lib';
import { Repository } from 'aws-cdk-lib/aws-ecr';
import { AccountRootPrincipal, PolicyStatement, Role } from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

export class UptEcrRepository extends Construct {
    public repository: Repository;

    constructor(scope: Construct, id: string) {
        super(scope, id);

        this.repository = new Repository(this, 'Repository');
        Stack.of(this).exportValue(this.repository.repositoryName);

        const publishRole = new Role(this, 'PublishRole', {
            assumedBy: new AccountRootPrincipal(),
            description: 'Role for CodeBuild to publish ECR images to our private repository.'
        });
        Stack.of(this).exportValue(publishRole.roleArn);

        this.repository.grantRead(publishRole);
        this.repository.grantPush(publishRole);

        // needed for publishing multi-architecture images
        publishRole.addToPrincipalPolicy(
            new PolicyStatement({
                actions: ['ecr:BatchGetImage'],
                resources: [this.repository.repositoryArn]
            })
        );
    }
}
