// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ApiEndpoints, solutionApi } from './solutionApi.ts';
import { ConfigResponse } from '../models/config.ts';
import { SaveRuleSetRequest } from '../models/ruleSet.ts';

export const ruleSetApiSlice = solutionApi.injectEndpoints({
    endpoints: builder => ({
        listRuleSets: builder.query<ConfigResponse, void>({
            query: () => ApiEndpoints.RULE_SET + '?includesHistorical=true',
            providesTags: ['RuleSet'],
        }),
        saveRuleSet: builder.mutation<ConfigResponse, SaveRuleSetRequest>({
            query: rules => ({
                url: ApiEndpoints.RULE_SET,
                method: 'POST',
                body: { saveRuleSetRq: rules },
            }),
            invalidatesTags: ['RuleSet'],
        }),
        activateRuleSet: builder.mutation<ConfigResponse, void>({
            query: () => ({
                url: ApiEndpoints.RULE_SET + '/activate',
                method: 'POST',
            }),
            invalidatesTags: ['RuleSet'],
        }),
        listRuleSetsCache: builder.query<ConfigResponse, void>({
            query: () => ApiEndpoints.RULE_SET_CACHE + '?includesHistorical=true',
            providesTags: ['RuleSetCache'],
        }),
        SaveCacheRuleSet: builder.mutation<ConfigResponse, SaveRuleSetRequest>({
            query: rules => ({
                url: ApiEndpoints.RULE_SET_CACHE,
                method: 'POST',
                body: { saveRuleSetRq: rules },
            }),
            invalidatesTags: ['RuleSetCache'],
        }),
        ActivateCacheRuleSet: builder.mutation<ConfigResponse, void>({
            query: () => ({
                url: ApiEndpoints.RULE_SET_CACHE + '/activate',
                method: 'POST',
            }),
            invalidatesTags: ['RuleSetCache'],
        }),
    }),
});

export const { useListRuleSetsQuery, useSaveRuleSetMutation, useActivateRuleSetMutation, useListRuleSetsCacheQuery, useSaveCacheRuleSetMutation, useActivateCacheRuleSetMutation } = ruleSetApiSlice;
