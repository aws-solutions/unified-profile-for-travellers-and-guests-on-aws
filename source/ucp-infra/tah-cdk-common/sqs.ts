// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as sqs from 'aws-cdk-lib/aws-sqs';
import { Construct } from 'constructs';

export class Queue extends sqs.Queue {
    //this constructor creates a default bucket that complies with all cdk_nag roles for AWS solutions
    constructor(scope: Construct, id: string, propOverride?: any) {
        let props: any = {
            enforceSSL: true,
            encryption: sqs.QueueEncryption.KMS_MANAGED
        }
        //overriding props with propOverride
        for (const key in propOverride) {
            props[key] = propOverride[key];
        }
        super(scope, id, props)
    }
}
