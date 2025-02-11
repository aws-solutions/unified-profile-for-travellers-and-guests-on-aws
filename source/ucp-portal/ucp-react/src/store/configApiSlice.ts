// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ApiEndpoints, solutionApi } from './solutionApi.ts';
import { ConfigResponse, HyperLinkMapping } from '../models/config.ts';

export const configApiSlice = solutionApi.injectEndpoints({
    endpoints: builder => ({
        getPortalConfig: builder.query<ConfigResponse, void>({
            query: () => ApiEndpoints.PORTAL_CONFIG,
            providesTags: ['Config'],
        }),
        postPortalConfig: builder.mutation<ConfigResponse, HyperLinkMapping[]>({
            query: hyperlinkMappings => ({
                url: ApiEndpoints.PORTAL_CONFIG,
                method: 'POST',
                body: { portalConfig: { hyperlinkMappings } },
            }),
            invalidatesTags: ['Config'],
        }),
    }),
});

export const { useGetPortalConfigQuery, usePostPortalConfigMutation } = configApiSlice;
