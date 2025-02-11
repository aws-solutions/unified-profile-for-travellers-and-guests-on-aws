// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

export const DATE_FORMAT = 'YYYY-MM-DD hh:mm:ss';

export const INVALID_DATE = '0001-01-01 12:00:00';

export const ROOT_TITLE = 'AWS Solutions for Travel and Hospitality';

export const S3_BUCKET_URL = '.console.aws.amazon.com/s3/buckets/';

export const HTTPS_PREFIX = 'https://';

export const AWS_GLUE_PREFIX = 'https://console.aws.amazon.com/gluestudio/home#/editor/job/';

export const ROUTES = {
    CREATE: 'create',
    SETTINGS: 'settings',
    ERRORS: 'errors',
    JOBS: 'jobs',
    CONFIG: 'config',
    PROFILE: 'profile',
    PROFILE_PROFILE_ID: 'profile/:profileId',
    IDENTITY_RESOLUTION: 'identity-resolution',
    DATA_QUALITY: 'data-quality',
    SETUP: 'setup',
    PRIVACY: 'privacy',
    PRIVACY_RESULTS: 'privacy/:connectId',
    RULES: 'rules',
    RULES_EDIT: 'rules/edit',
    CACHE: 'cache',
    CACHE_EDIT: 'cache/edit',
    EDIT: 'edit',
    AI_MATCHES: 'matches',
} as const;

export const TIMESTAMP = 'timestamp';
