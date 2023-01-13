import { ManagedPolicy, Role, ServicePrincipal } from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';
import { NagSuppressions } from 'cdk-nag';

export class BuildRole extends Role {
    constructor(scope: Construct, id: string) {
        super(scope, id, {
            assumedBy: new ServicePrincipal('codebuild.amazonaws.com'),
            managedPolicies: [ManagedPolicy.fromAwsManagedPolicyName("AdministratorAccess")]
        })
        NagSuppressions.addResourceSuppressions(this, [
            {
                id: 'AwsSolutions-IAM4',
                reason: 'The code pipeline project required administration role to work since it deploys a large spectrum of resources'
            },
        ]);
    }
}