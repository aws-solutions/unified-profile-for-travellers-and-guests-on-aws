import { CfnOutput } from 'aws-cdk-lib';
import { Construct } from 'constructs';

export class Output {
    static add(scope: Construct, key: string, val: string): CfnOutput {
        return new CfnOutput(scope, key, { value: val })
    }
}