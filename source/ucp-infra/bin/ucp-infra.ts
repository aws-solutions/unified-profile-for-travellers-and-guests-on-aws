#!/usr/bin/env node
//import 'source-map-support/register';
import { App } from 'aws-cdk-lib';
import { UCPInfraStack } from '../lib/ucp-infra-stack';
import { Tags } from "aws-cdk-lib";

const app = new App();
const envName = app.node.tryGetContext("envName");

let stack = new UCPInfraStack(app, 'UCPInfraStack' + envName, {
    env: {
        account: process.env.CDK_DEFAULT_ACCOUNT,
        region: process.env.CDK_DEFAULT_REGION,

    },
    description: "Deploys the Restaurant Edge Playtfrom infrastructure",
});


/**************
 * Tagging all resources in stack
 */
Tags.of(stack).add('application-name', 'ucp');
Tags.of(stack).add('application-env', envName);

app.synth();
