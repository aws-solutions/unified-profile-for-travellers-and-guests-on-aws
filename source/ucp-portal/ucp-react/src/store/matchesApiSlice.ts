// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ConfigResponse } from '../models/config.ts';
import { PaginationOptions } from '../models/pagination.ts';
import { MergeRq } from '../models/traveller.ts';
import { ApiEndpoints, solutionApi } from './solutionApi.ts';

export const matchesApiSlice = solutionApi.injectEndpoints({
    endpoints: builder => ({
        getMatches: builder.query<ConfigResponse, PaginationOptions>({
            query: pagination => ApiEndpoints.PROFILE + `?page=` + pagination.page + `&pageSize=` + pagination.pageSize + `&matches=true`,
            providesTags: ['Matches'],
        }),
        performMerge: builder.mutation<ConfigResponse, MergeRq[]>({
            query: mergeRq => ({
                url: ApiEndpoints.MERGE,
                method: 'POST',
                body: { mergeRq },
            }),
            invalidatesTags: ['Matches'],
        }),
    }),
});

export const { useGetMatchesQuery, usePerformMergeMutation } = matchesApiSlice;
