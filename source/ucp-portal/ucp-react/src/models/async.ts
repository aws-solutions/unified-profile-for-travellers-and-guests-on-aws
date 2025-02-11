// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

export interface GetAsyncStatusRequest {
    id: AsyncEventType;
    useCase: string;
}

export enum AsyncEventType {
    CREATE_DOMAIN = 'createDomain',
    DELETE_DOMAIN = 'deleteDomain',
    EMPTY_TABLE = 'emptyTable',
    NULL = '',
    CREATE_PRIVACY_SEARCH = 'createPrivacySearch',
    CREATE_PRIVACY_PURGE = 'purgeProfileData',
    UNMERGE_PROFILES = 'unmergeProfiles',
    MERGE_PROFILES = 'mergeProfiles',
    REBUILD_CACHE = 'rebuildCache',
}

export enum AsyncEventStatus {
    EVENT_STATUS_INVOKED = 'invoked',
    EVENT_STATUS_RUNNING = 'running',
    EVENT_STATUS_SUCCESS = 'success',
    EVENT_STATUS_FAILED = 'failed',
}
