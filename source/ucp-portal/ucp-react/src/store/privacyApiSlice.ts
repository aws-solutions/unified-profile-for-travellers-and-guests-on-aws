// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { FetchBaseQueryError, QueryActionCreatorResult, QueryDefinition } from '@reduxjs/toolkit/query';
import { AxiosError } from 'axios';
import { AsyncEventStatus } from '../models/async';
import { AsyncEventResponse, AsyncEventResponseWithTxID, ConfigResponse, TxIdResponse } from '../models/config';
import { PrivacyPurgeRequest, PrivacyPurgeStatus, PrivacySearchRequest, PrivacySearchResponse } from '../models/privacy';
import { asyncApiSlice } from './asyncApiSlice';
import { notificationsSlice } from './notificationsSlice';
import { ApiEndpoints, solutionApi } from './solutionApi';

export const privacyApiSlice = solutionApi.injectEndpoints({
    endpoints: builder => ({
        listSearches: builder.query<PrivacySearchResponse[], void>({
            query: () => ({
                url: `${ApiEndpoints.PRIVACY}`,
                method: 'GET',
            }),
            providesTags: ['PrivacySearchList'],
            transformResponse: (response: ConfigResponse) => {
                return response.privacySearchResults;
            },
        }),
        getProfileSearchResult: builder.query<PrivacySearchResponse, PrivacySearchRequest>({
            query: ({ connectId }) => ({
                url: `${ApiEndpoints.PRIVACY}/${connectId}`,
                method: 'GET',
            }),
            providesTags: (_result, _error, { connectId }) => {
                return [{ type: 'PrivacySearchProfile', id: connectId }];
            },
            transformResponse: (response: ConfigResponse) => {
                return response.privacySearchResult;
            },
        }),
        deleteSearches: builder.mutation<ConfigResponse, string[]>({
            async queryFn(arg, _queryApi, _extraOptions, baseQuery) {
                try {
                    const result = await baseQuery({ url: ApiEndpoints.PRIVACY, method: 'DELETE', body: arg });
                    if (result.error) {
                        return { error: result.error as FetchBaseQueryError };
                    }
                    const data = result.data as ConfigResponse;
                    return { data };
                } catch (error: unknown) {
                    if (error instanceof AxiosError) {
                        const e: FetchBaseQueryError = {
                            status: 'CUSTOM_ERROR',
                            error: error.message,
                            data: { TxID: error.response?.data.TxID } as TxIdResponse,
                        };
                        return { error: e };
                    }
                    return { error: error as FetchBaseQueryError };
                }
            },
            invalidatesTags: ['PrivacySearchList'],
        }),
        createSearch: builder.mutation<AsyncEventResponseWithTxID, { privacySearchRq: PrivacySearchRequest[] }>({
            async queryFn(arg, queryApi, _extraOptions, baseQuery) {
                try {
                    const result = await baseQuery({ url: ApiEndpoints.PRIVACY, method: 'POST', body: arg });
                    if (result.error) {
                        return { error: result.error as FetchBaseQueryError };
                    }
                    const data = result.data as ConfigResponse;
                    return { data: { TxID: data.TxID, ...data.asyncEvent } };
                } catch (error: unknown) {
                    if (error instanceof AxiosError) {
                        const e: FetchBaseQueryError = {
                            status: 'CUSTOM_ERROR',
                            error: error.message,
                            data: { TxID: error.response?.data.TxID } as TxIdResponse,
                        };
                        return { error: e };
                    }
                    return { error: error as FetchBaseQueryError };
                }
            },
            async onQueryStarted(arg, api) {
                try {
                    const { data } = await api.queryFulfilled;
                    if (data.status === AsyncEventStatus.EVENT_STATUS_INVOKED) {
                        api.dispatch(solutionApi.util.invalidateTags(['PrivacySearchList']));
                        arg.privacySearchRq.forEach(search => {
                            api.dispatch(solutionApi.util.invalidateTags([{ type: 'PrivacySearchProfile', id: search.connectId }]));
                        });
                    }
                } catch (error) {
                    console.error(error);
                }
            },
        }),
        createSearchCustom: builder.mutation<AsyncEventResponse, { privacySearchRq: PrivacySearchRequest; waitForResults?: boolean }>({
            async queryFn(arg, queryApi, _extraOptions, baseQuery) {
                const result = await baseQuery({ url: ApiEndpoints.PRIVACY, method: 'POST', body: arg });
                if (result.error) {
                    return { error: result.error as FetchBaseQueryError };
                }
                const data = result.data as ConfigResponse;
                if (arg.waitForResults) {
                    queryApi.dispatch(
                        notificationsSlice.actions.addNotification({
                            id: 'searchForProfile',
                            type: 'in-progress',
                            content: 'Searching for Profile Data',
                            loading: true,
                        }),
                    );
                    queryApi.dispatch(privacyApiSlice.endpoints.listSearches.initiate());
                    let pollingAsyncResult: Awaited<QueryActionCreatorResult<QueryDefinition<unknown, any, string, AsyncEventResponse>>>;
                    do {
                        pollingAsyncResult = await queryApi.dispatch(
                            asyncApiSlice.endpoints.getAsyncStatus.initiate({
                                id: data.asyncEvent.item_type,
                                useCase: data.asyncEvent.item_id,
                            }),
                        );
                        await new Promise(resolve => setTimeout(resolve, 2000));
                    } while (
                        pollingAsyncResult.data?.status === AsyncEventStatus.EVENT_STATUS_INVOKED ||
                        pollingAsyncResult.data?.status === AsyncEventStatus.EVENT_STATUS_RUNNING
                    );

                    queryApi.dispatch(notificationsSlice.actions.deleteNotification({ id: 'searchForProfile' }));
                    queryApi.dispatch(privacyApiSlice.endpoints.listSearches.initiate());
                    return { data: pollingAsyncResult.data! };
                } else {
                    queryApi.dispatch(
                        notificationsSlice.actions.addNotification({
                            id: 'searchForProfile',
                            type: 'success',
                            content: 'Profile Data Search Initiated',
                        }),
                    );
                }
                return { data: data.asyncEvent };
            },
        }),
        purgeProfile: builder.mutation<AsyncEventResponseWithTxID, { createPrivacyPurgeRq: PrivacyPurgeRequest }>({
            query: createPurgeRequest => ({
                url: `${ApiEndpoints.PRIVACY}/purge`,
                method: 'POST',
                body: createPurgeRequest,
            }),
            invalidatesTags: ['PrivacySearchList', 'PrivacyPurgeIsRunning'],
            transformResponse: (response: ConfigResponse) => {
                return {
                    ...response.asyncEvent,
                    TxID: response.TxID,
                };
            },
        }),
        getPurgeIsRunning: builder.query<PrivacyPurgeStatus, void | null>({
            query: () => ({
                url: `${ApiEndpoints.PRIVACY}/purge`,
                method: 'GET',
            }),
            providesTags: ['PrivacyPurgeIsRunning'],
            transformResponse: (response: ConfigResponse) => {
                return response.privacyPurgeStatus;
            },
            async onQueryStarted(arg, api) {
                try {
                    const { data } = await api.queryFulfilled;
                    if (data.isPurgeRunning === false) {
                        api.dispatch(solutionApi.util.invalidateTags(['PrivacySearchList']));
                    }
                } catch (error) {
                    console.error(error);
                }
            },
        }),
    }),
});

export const {
    useListSearchesQuery,
    useCreateSearchMutation,
    useGetProfileSearchResultQuery,
    useDeleteSearchesMutation,
    usePurgeProfileMutation,
    useCreateSearchCustomMutation,
    useGetPurgeIsRunningQuery,
} = privacyApiSlice;
