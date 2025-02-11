// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { PayloadAction, createSlice } from '@reduxjs/toolkit';
import { PrivacyProfileLocationsTableItem, PrivacySearchResponse, PrivacySearchesTableItem } from '../models/privacy';
import { privacyApiSlice } from './privacyApiSlice';
import { RootState } from './store';

export type PrivacySliceState = {
    searchResults: PrivacySearchesTableItem[];
    searchResult: PrivacyProfileLocationsTableItem[];
};

const InitialState: PrivacySliceState = {
    searchResults: [],
    searchResult: [],
};

export const privacySlice = createSlice({
    name: 'privacy',
    initialState: InitialState,
    reducers: {
        clearSearchResult: state => {
            state.searchResult = [];
        },
    },
    extraReducers: builder => {
        builder.addMatcher(
            privacyApiSlice.endpoints.listSearches.matchFulfilled,
            (state: PrivacySliceState, action: PayloadAction<PrivacySearchResponse[]>) => {
                state.searchResults = action.payload.map(res => {
                    return {
                        connectId: res.connectId,
                        status: res.status,
                        searchDate: res.searchDate,
                        totalResultsFound: res.totalResultsFound,
                        errorMessage: res.errorMessage,
                    };
                });
            },
        );
        builder.addMatcher(
            privacyApiSlice.endpoints.getProfileSearchResult.matchFulfilled,
            (state: PrivacySliceState, action: PayloadAction<PrivacySearchResponse>) => {
                const records: PrivacyProfileLocationsTableItem[] = [];

                //  Map Over All Locations
                if (action.payload.locationResults) {
                    Object.keys(action.payload.locationResults).forEach((key: string) => {
                        if (action.payload.locationResults[key]) {
                            records.push(...action.payload.locationResults[key].map(value => ({ source: key, path: value })));
                        }
                    });
                }
                state.searchResult = records;
            },
        );
    },
});

export default privacySlice.reducer;
export const { clearSearchResult } = privacySlice.actions;
export const selectPrivacySearchResults = (state: RootState) => state.privacy.searchResults;
export const selectPrivacySearchResult = (state: RootState) => state.privacy.searchResult;
