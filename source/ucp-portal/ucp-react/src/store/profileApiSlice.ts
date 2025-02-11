// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ApiEndpoints, solutionApi } from './solutionApi.ts';
import { AsyncEventResponse, ConfigResponse } from '../models/config.ts';
import { ProfileRequest } from './profileSlice.ts';
import { MultiPaginationProp } from '../models/pagination.ts';
import { UnmergeRq } from '../models/traveller.ts';

function buildSearchParams(params: ProfileRequest | MultiPaginationProp) {
    const searchParams: URLSearchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, val]) => {
        if (Array.isArray(val)) {
            if (val.length > 0) {
                searchParams.append(key, val.join(','));
            }
        } else {
            if (val) {
                searchParams.append(key, val);
            }
        }
    });
    return searchParams;
}

export const profileApiSlice = solutionApi.injectEndpoints({
    endpoints: builder => ({
        getProfilesByParam: builder.query<ConfigResponse, ProfileRequest>({
            query: arg => {
                return {
                    url: ApiEndpoints.PROFILE + `?${buildSearchParams(arg).toString()}`,
                    method: 'GET',
                };
            },
            providesTags: ['Profile'],
        }),
        getProfilesById: builder.query<ConfigResponse, { id: string; params?: MultiPaginationProp }>({
            query: ({ id, params }) => {
                let url = ApiEndpoints.PROFILE + `/${id}`;
                if (params) url = url + `?${buildSearchParams(params).toString()}`;
                return {
                    url: url,
                    method: 'GET',
                };
            },
            providesTags: ['Profile'],
        }),
        getProfilesSummary: builder.query<string, string>({
            query: id => ApiEndpoints.PROFILE_SUMMARY + `/${id}`,
            transformResponse: (response: ConfigResponse) => {
                return response.profileSummary;
            },
            providesTags: ['ProfileSummary'],
        }),
        getInteractionHistory: builder.query<ConfigResponse, { id: string; params?: MultiPaginationProp | any }>({
            query: ({ id, params }) => {
                let url = ApiEndpoints.INTERACTION_HISTORY + `/${id}`;
                if (params) url = url + `?${buildSearchParams(params).toString()}`;
                return {
                    url: url,
                    method: 'GET',
                };
            },
            providesTags: ['Profile'],
        }),
        performUnmerge: builder.mutation<ConfigResponse, UnmergeRq>({
            query: unmergeRq => ({
                url: ApiEndpoints.UNMERGE,
                method: 'POST',
                body: { unmergeRq },
            }),
        }),
    }),
});

export const {
    useGetProfilesByParamQuery,
    useGetProfilesByIdQuery,
    useGetInteractionHistoryQuery,
    usePerformUnmergeMutation,
    useLazyGetProfilesSummaryQuery,
} = profileApiSlice;
