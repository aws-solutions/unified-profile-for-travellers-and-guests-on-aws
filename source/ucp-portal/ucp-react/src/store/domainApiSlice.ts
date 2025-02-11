// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ApiEndpoints, solutionApi } from './solutionApi.ts';
import { DomainCreateRequest } from '../models/domain.ts';
import { ConfigResponse, PromptConfigRq } from '../models/config.ts';
import { AsyncEventResponse } from '../models/config.ts';

export const domainApiSlice = solutionApi.injectEndpoints({
    endpoints: builder => ({
        getAllDomains: builder.query<ConfigResponse, void>({
            query: () => ApiEndpoints.ADMIN,
            providesTags: ['Domain'],
        }),
        getDomain: builder.query<ConfigResponse, string>({
            query: name => ({
                url: ApiEndpoints.ADMIN + `/${name}`,
                method: 'GET',
            }),
            providesTags: ['Domain'],
        }),
        getPromptConfig: builder.query<ConfigResponse, void>({
            query: () => ApiEndpoints.PROMPT_CONFIG,
            providesTags: ['PromptConfig'],
        }),
        createDomain: builder.mutation<AsyncEventResponse, { createRequest: DomainCreateRequest }>({
            query: ({ createRequest }) => ({
                url: ApiEndpoints.ADMIN,
                method: 'POST',
                body: createRequest,
            }),
            transformResponse: (response: ConfigResponse) => {
                return response.asyncEvent;
            },
            invalidatesTags: ['Domain'],
        }),
        deleteDomain: builder.mutation<AsyncEventResponse, string>({
            query: name => ({
                url: ApiEndpoints.ADMIN + `/${name}`,
                method: 'DELETE',
                headers: {
                    'customer-profiles-domain': name,
                },
            }),
            transformResponse: (response: ConfigResponse) => {
                return response.asyncEvent;
            },
            invalidatesTags: ['Domain'],
        }),
        savePromptConfig: builder.mutation<ConfigResponse, PromptConfigRq>({
            query: mergeRq => ({
                url: ApiEndpoints.PROMPT_CONFIG,
                method: 'POST',
                body: mergeRq,
            }),
            invalidatesTags: ['PromptConfig'],
        }),
    }),
});

export const {
    useGetAllDomainsQuery,
    useGetDomainQuery,
    useGetPromptConfigQuery,
    useLazyGetDomainQuery,
    useCreateDomainMutation,
    useDeleteDomainMutation,
    useSavePromptConfigMutation,
} = domainApiSlice;
