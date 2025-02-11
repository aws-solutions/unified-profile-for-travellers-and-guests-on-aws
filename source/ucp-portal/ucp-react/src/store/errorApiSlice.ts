// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ApiEndpoints, solutionApi } from './solutionApi.ts';
import { AsyncEventResponse, ConfigResponse } from '../models/config.ts';
import { DeleteErrorRequest, GetErrorsRequest } from '../models/error.ts';

export const errorApiSlice = solutionApi.injectEndpoints({
    endpoints: builder => ({
        getAllErrors: builder.query<ConfigResponse, GetErrorsRequest>({
            query: req => {
                const query: string[] = [ApiEndpoints.ERROR + '?'];
                if (req.page !== 0) {
                    query.push(`page=${req.page}&`);
                }
                query.push(`pageSize=${req.pageSize}`);
                return query.join('');
            },
            providesTags: ['Error'],
        }),
        deleteError: builder.mutation<ConfigResponse, { deleteErrorRequest: DeleteErrorRequest }>({
            query: ({ deleteErrorRequest }) => ({
                url: ApiEndpoints.ERROR + `/${deleteErrorRequest.errorId}`,
                method: 'DELETE',
            }),
            invalidatesTags: ['Error'],
        }),
        deleteAllErrors: builder.mutation<AsyncEventResponse, void>({
            query: () => ({
                url: ApiEndpoints.ERROR + `/*`,
                method: 'DELETE',
            }),
            transformResponse: (response: ConfigResponse) => {
                return response.asyncEvent;
            },
            invalidatesTags: ['Error'],
        }),
    }),
});

export const { useGetAllErrorsQuery, useDeleteErrorMutation, useDeleteAllErrorsMutation } = errorApiSlice;
