// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ApiEndpoints, solutionApi } from './solutionApi.ts';
import { AsyncEventResponse, ConfigResponse } from '../models/config.ts';
import { CacheMode } from '../models/domain.ts';

export const cacheApiSlice = solutionApi.injectEndpoints({
    endpoints: builder => ({
        rebuildCache: builder.mutation<AsyncEventResponse, CacheMode>({
            query: cacheMode => ({
                url: ApiEndpoints.CACHE + `?type=${cacheMode}`,
                method: 'POST',
            }),
            transformResponse: (response: ConfigResponse) => {
                return response.asyncEvent;
            },
            invalidatesTags: ['Cache'],
        }),
    }),
});

export const { useRebuildCacheMutation } = cacheApiSlice;
