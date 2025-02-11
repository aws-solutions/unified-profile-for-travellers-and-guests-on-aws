// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { createSlice } from '@reduxjs/toolkit';
import type { PayloadAction } from '@reduxjs/toolkit';
import { RootState } from './store';

interface UserState {
    appAccessPermission: number;
}

const initialState: UserState = {
    appAccessPermission: 0,
};

export const userSlice = createSlice({
    name: 'user',
    initialState,
    reducers: {
        setAppAccessPermission(state, action: PayloadAction<number>) {
            state.appAccessPermission = action.payload;
        },
    },
});

export const { setAppAccessPermission } = userSlice.actions;
export const selectAppAccessPermission = (state: RootState) => state.user.appAccessPermission;
