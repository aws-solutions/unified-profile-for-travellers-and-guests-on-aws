// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { BaseQueryApi, BaseQueryFn, createApi, FetchArgs, FetchBaseQueryError } from '@reduxjs/toolkit/query/react';
import { API } from 'aws-amplify';

/**
 * Utilize Amplify's API methods to make request. Amplify will take care of authentication, base url etc.
 */
export const dynamicBaseQuery: BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError> = async (
    args: string | FetchArgs,
    api: BaseQueryApi,
    extraOptions: any,
) => {
    function runAmplifyAxiosRequest(): Promise<any> {
        if (typeof args === 'string') {
            return API.get('solution-api', args, extraOptions);
        } else {
            switch (args.method) {
                case 'POST':
                    return API.post('solution-api', args.url, { body: args.body, ...extraOptions });
                case 'PUT':
                    return API.put('solution-api', args.url, { body: args.body, ...extraOptions });
                case 'DELETE':
                    return API.del('solution-api', args.url, { body: args.body, ...extraOptions });
                case 'PATCH':
                    return API.patch('solution-api', args.url, {
                        body: args.body,
                        ...extraOptions,
                    });
                case 'HEAD':
                    return API.head('solution-api', args.url, extraOptions);
                default:
                    return API.get('solution-api', args.url, extraOptions);
            }
        }
    }

    const data = await runAmplifyAxiosRequest();
    return { data };
};

/**
 * Create 1 api per base URL.
 * (Most of the time, that's 1 api per solution.)
 */
export const solutionApi = createApi({
    reducerPath: 'solution-api',
    baseQuery: dynamicBaseQuery,
    endpoints: () => ({}),
    refetchOnMountOrArgChange: true,
    tagTypes: [
        'Domain',
        'Profile',
        'Async',
        'Error',
        'Job',
        'Cache',
        'Config',
        'PrivacySearchList',
        'PrivacySearchProfile',
        'PrivacyPurgeIsRunning',
        'RuleSet',
        'RuleSetCache',
        'Matches',
        'PromptConfig',
        'ProfileSummary',
    ],
});

export enum ApiEndpoints {
    ADMIN = 'api/ucp/admin',
    PROFILE = 'api/ucp/profile',
    PROFILE_SUMMARY = 'api/ucp/profile/summary',
    ASYNC = 'api/ucp/async',
    ERROR = 'api/ucp/error',
    JOB = 'api/ucp/jobs',
    PORTAL_CONFIG = 'api/ucp/portalConfig',
    PROMPT_CONFIG = 'api/ucp/promptConfig',
    PRIVACY = 'api/ucp/privacy',
    RULE_SET = 'api/ucp/ruleSet',
    RULE_SET_CACHE = 'api/ucp/ruleSetCache',
    INTERACTION_HISTORY = 'api/ucp/interactionHistory',
    UNMERGE = 'api/ucp/unmerge',
    MERGE = 'api/ucp/merge',
    CACHE = 'api/ucp/cache',
}
