// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Aws, Fn, RemovalPolicy } from 'aws-cdk-lib';
import * as s3 from 'aws-cdk-lib/aws-s3';
import { NagSuppressions } from 'cdk-nag';
import { Construct } from 'constructs';

export class Bucket extends s3.Bucket {
    //this constructor creates a default bucket that complies with all cdk_nag roles for AWS solutions
    constructor(scope: Construct, id: string, logsLocation: s3.Bucket, propOverride?: any) {
        let props: any = {
            removalPolicy: RemovalPolicy.DESTROY,
            versioned: true,
            serverAccessLogsBucket: logsLocation,
            serverAccessLogsPrefix: 'bucket-' + id,
            encryption: s3.BucketEncryption.S3_MANAGED,
            blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
            enforceSSL: true,
            env: { account: Aws.ACCOUNT_ID, region: Aws.REGION },
            bucketKeyEnabled: true
        };
        //overriding props with propOverride
        for (const key in propOverride) {
            props[key] = propOverride[key];
        }
        super(scope, id, props);
    }

    //this function returns the name of the Athena table created by the Gule Crawler for this bucket
    toAthenaTable(): string {
        return Fn.join('_', Fn.split('-', this.bucketName));
    }
}

//bucket to store access logs created
export class AccessLogBucket extends s3.Bucket {
    //this constructor creates a default bucket that complies with all cdk_nag roles for AWS solutions
    constructor(scope: Construct, id: string) {
        super(scope, id, {
            removalPolicy: RemovalPolicy.RETAIN,
            versioned: true,
            encryption: s3.BucketEncryption.S3_MANAGED,
            blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
            enforceSSL: true
        });
        NagSuppressions.addResourceSuppressions(this, [
            {
                id: 'AwsSolutions-S1',
                reason: 'bucket used to store access logs'
            }
        ]);
    }
}
