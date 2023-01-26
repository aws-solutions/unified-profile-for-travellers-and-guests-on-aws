#!/usr/bin/env node
import 'source-map-support/register';
import { App } from 'aws-cdk-lib';
import { Aspects } from 'aws-cdk-lib';
import { UCPCodePipelinesStack } from '../lib/ucp-code-pipelines-stack';
import { AwsSolutionsChecks } from 'cdk-nag';


const app = new App();
const envName = app.node.tryGetContext("envName");
let stack = new UCPCodePipelinesStack(app, 'UCPCodePipelinesStack' + envName, {
    description: "Deploys the travel and hospitality connector solution pipeline to run the CDK deployment (SO0230)",
});

Aspects.of(app).add(new AwsSolutionsChecks)