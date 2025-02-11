// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { createSlice } from '@reduxjs/toolkit';
import type { PayloadAction } from '@reduxjs/toolkit';
import { Domain, DomainSetting, PromptConfig } from '../models/config.ts';
import { RootState } from './store';
import { domainApiSlice } from './domainApiSlice.ts';
import { asyncApiSlice } from './asyncApiSlice.ts';
import { AsyncEventType } from '../models/async.ts';
import { removeDomainValueStorage, setDomainValueStorage } from '../utils/localStorageUtil.ts';

interface DomainState {
    selectedDomain: Domain;
    allDomains: Domain[];
    promptConfig: PromptConfig;
    matchConfig: PromptConfig;
}

const initDomain: Domain = {
    customerProfileDomain: '',
    numberOfObjects: 0,
    numberOfProfiles: 0,
    matchingEnabled: false,
    cacheMode: 0,
    created: '',
    lastUpdated: '',
    mappings: null,
    integrations: null,
    needsMappingUpdate: false,
};

const initialState: DomainState = {
    selectedDomain: initDomain,
    allDomains: [],
    matchConfig: {
        isActive: false,
        value: '',
    },
    promptConfig: {
        value: '',
        isActive: false,
    },
};

export const domainSlice = createSlice({
    name: 'domains',
    initialState,
    reducers: {
        setSelectedDomain(state, action: PayloadAction<Domain>) {
            setDomainValueStorage(action.payload.customerProfileDomain);
            state.selectedDomain = action.payload;
        },
    },
    extraReducers: builder => {
        builder
            .addMatcher(domainApiSlice.endpoints.getDomain.matchFulfilled, (state, { payload }) => {
                state.selectedDomain = payload.config.domains[0];
            })
            .addMatcher(asyncApiSlice.endpoints.getAsyncStatus.matchFulfilled, (state, { payload }) => {
                if (payload.item_id === AsyncEventType.DELETE_DOMAIN) {
                    if (payload.status === 'success' || payload.status === 'failed') {
                        removeDomainValueStorage();
                        state.selectedDomain = initDomain;
                    }
                }
            })
            .addMatcher(domainApiSlice.endpoints.getAllDomains.matchFulfilled, (state, { payload }) => {
                state.allDomains = payload.config.domains;
            })
            .addMatcher(domainApiSlice.endpoints.getPromptConfig.matchFulfilled, (state, { payload }) => {
                state.promptConfig.value = payload.domainSetting.promptConfig.value;
                state.promptConfig.isActive = payload.domainSetting.promptConfig.isActive;
                state.matchConfig.value = payload.domainSetting.matchConfig.value || '0';
                state.matchConfig.isActive = payload.domainSetting.matchConfig.isActive;
            });
    },
});
export const { setSelectedDomain } = domainSlice.actions;

export default domainSlice.reducer;

export const selectCurrentDomain = (state: RootState) => state.domains.selectedDomain;
export const selectAllDomains = (state: RootState) => state.domains.allDomains;
export const selectPromptValue = (state: RootState) => state.domains.promptConfig.value;
export const selectIsPromptActive = (state: RootState) => state.domains.promptConfig.isActive;
export const selectMatchValue = (state: RootState) => state.domains.matchConfig.value;
export const selectIsAutoMatchActive = (state: RootState) => state.domains.matchConfig.isActive;
