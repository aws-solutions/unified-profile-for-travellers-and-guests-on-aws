// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Construct } from 'constructs';
import { Database } from '@aws-cdk/aws-glue-alpha';
import * as glue from 'aws-cdk-lib/aws-glue';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as s3 from 'aws-cdk-lib/aws-s3';


export class S3Crawler extends glue.CfnCrawler {
    constructor(scope: Construct, id: string, glueDb: Database, bucket: s3.IBucket, dataLakeAdminRole: iam.Role) {
        super(scope, id, {
            role: dataLakeAdminRole.roleArn,
            name: id,
            description: `Glue crawler for ${bucket.bucketName} (${id})`,
            targets: {
                s3Targets: [{ path: "s3://" + bucket.bucketName }]
            },
            configuration: `{
            "Version": 1.0,
            "Grouping": {
               "TableGroupingPolicy": "CombineCompatibleSchemas" }
         }`,
            databaseName: glueDb.databaseName
        })
    }
}

