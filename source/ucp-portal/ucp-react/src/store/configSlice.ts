// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { SelectProps } from '@cloudscape-design/components';
import type { PayloadAction } from '@reduxjs/toolkit';
import { createSlice } from '@reduxjs/toolkit';
import { AccpRecord, HyperLinkMapping } from '../models/config.ts';
import { configApiSlice } from './configApiSlice.ts';
import { RootState } from './store';

interface ConfigState {
    dataRecordInput: SelectProps.Option | null;
    dataFieldInput: SelectProps.Option | null;
    urlTemplateInput: string;
    hyperlinkMappings: HyperLinkMapping[];
    accpRecords: AccpRecord[];
}

const initialState: ConfigState = {
    dataRecordInput: null,
    dataFieldInput: null,
    urlTemplateInput: '',
    hyperlinkMappings: [],
    accpRecords: [],
};

export const configSlice = createSlice({
    name: 'config',
    initialState,
    reducers: {
        setDataRecordInput(state, action: PayloadAction<SelectProps.Option | null>) {
            Object.assign(state, { dataRecordInput: action.payload });
        },
        setDataFieldInput(state, action: PayloadAction<SelectProps.Option | null>) {
            Object.assign(state, { dataFieldInput: action.payload });
        },
        setUrlTemplateInput(state, action: PayloadAction<string>) {
            state.urlTemplateInput = action.payload;
        },
        clearFormFields(state, action: PayloadAction<void>) {
            state.dataRecordInput = null;
            state.dataFieldInput = null;
            state.urlTemplateInput = '';
        },
    },
    extraReducers: builder => {
        builder
            .addMatcher(configApiSlice.endpoints.getPortalConfig.matchFulfilled, (state, { payload }) => {
                state.hyperlinkMappings = payload.portalConfig.hyperlinkMappings;
                state.accpRecords = payload.accpRecords ?? [];
            })
            .addMatcher(configApiSlice.endpoints.postPortalConfig.matchFulfilled, (state, { payload }) => {
                state.hyperlinkMappings = payload.portalConfig.hyperlinkMappings;
            });
    },
});
export const { setDataRecordInput, setDataFieldInput, setUrlTemplateInput, clearFormFields } = configSlice.actions;

export default configSlice.reducer;

export const selectDataRecordInput = (state: RootState) => state.config.dataRecordInput;
export const selectDataFieldInput = (state: RootState) => state.config.dataFieldInput;
export const selectUrlTemplateInput = (state: RootState) => state.config.urlTemplateInput;
export const selectHyperLinkMappings = (state: RootState) => state.config.hyperlinkMappings;
export const selectAccpRecords = (state: RootState) => state.config.accpRecords;
