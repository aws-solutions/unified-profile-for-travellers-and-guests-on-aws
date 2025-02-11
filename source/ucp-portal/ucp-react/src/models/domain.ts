// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

export interface DomainCreateRequest {
    domain: DomainNameWrapper;
}

export interface DomainNameWrapper {
    customerProfileDomain: string;
}

// This constant is also defined in LCS (see source/storage/customerprofiles_lowcost_constant.go)
export enum CacheMode {
    NONE = 0,
    CP_MODE = 1 << 0,
    DYNAMO_MODE = 1 << 1,
}
