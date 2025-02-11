// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { createSlice } from '@reduxjs/toolkit';
import { RootState } from './store';

export interface PermissionSystem {
    /**
     * The permission system type being used.
     *
     * Current Supported Options
     * - Cognito: default behavior, uses Cognito User Groups to manage fine grained permissions
     * - Disabled: no permissions are used, all features are enabled for authenticated users with access to UPT
     */
    type: string;
}

export enum PermissionSystemType {
    Cognito = 'Cognito',
    Disabled = 'Disabled',
}

export interface PermissionSystemState {
    permissionSystem: PermissionSystem;
}

const initialState: PermissionSystemState = {
    permissionSystem: {
        type: PermissionSystemType.Cognito,
    },
};

export const permissionSystemSlice = createSlice({
    name: 'permissionSystem',
    initialState,
    reducers: {},
});

export default permissionSystemSlice.reducer;

export const selectPermissionSystem = (state: RootState) => state.permissionSystem.permissionSystem;
