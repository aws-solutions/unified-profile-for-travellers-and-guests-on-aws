// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { AsyncEventType, GetAsyncStatusRequest } from '../models/async.ts';
import { AsyncEventResponseWithTxID, ConfigResponse } from '../models/config.ts';
import { ApiEndpoints, solutionApi } from './solutionApi.ts';

export const asyncApiSlice = solutionApi.injectEndpoints({
    endpoints: builder => ({
        getAsyncStatus: builder.query<AsyncEventResponseWithTxID, GetAsyncStatusRequest>({
            query: req => ApiEndpoints.ASYNC + `?id=${req.id}&usecase=${req.useCase}`,
            providesTags: ['Async'],
            transformResponse: (response: ConfigResponse) => {
                return {
                    TxID: response.TxID,
                    ...response.asyncEvent,
                };
            },
            async onQueryStarted(arg, api) {
                const { data } = await api.queryFulfilled;
                if (data.status === 'success' || data.status === 'failure' || data.status === 'invoked') {
                    if (data.item_type === AsyncEventType.CREATE_PRIVACY_SEARCH) {
                        api.dispatch(solutionApi.util.invalidateTags(['PrivacySearchList']));
                    }
                }
            },
        }),
    }),
});

export const { useGetAsyncStatusQuery, useLazyGetAsyncStatusQuery } = asyncApiSlice;
