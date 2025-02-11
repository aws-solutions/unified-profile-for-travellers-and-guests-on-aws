// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { StatusTypes } from './enums';

export type CompletionStatus = {
    Status: StatusTypes;
    Data: Record<string, unknown> | { Error?: { Code: string; Message: string } };
};

export type FrontEndConfig = {
    readonly artifactBucketName: string;
    readonly artifactBucketPath: string;
    readonly cognitoClientId: string;
    readonly cognitoUserPoolId: string;
    readonly deployed: string;
    readonly staticContentBucketName: string;
    readonly ucpApiUrl: string;
    readonly usePermissionSystem: string;
};
