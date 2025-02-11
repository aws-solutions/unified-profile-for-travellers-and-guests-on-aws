// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { PayloadAction, createSlice } from '@reduxjs/toolkit';
import { MatchDiff, MatchPair } from '../models/match';
import { RootState } from './store';
import { matchesApiSlice } from './matchesApiSlice';
import { convertMatchPairsToTableItems } from '../utils/matchUtils';

export interface MatchesState {
    profileDataDiff: MatchDiff;
    matchPairs: MatchPair[];
    matchTableItems: MatchDiff[];
    totalMatches: number;
}

const initialState: MatchesState = {
    profileDataDiff: {} as MatchDiff,
    matchPairs: [],
    matchTableItems: [],
    totalMatches: 0,
};

export const matchesSlice = createSlice({
    name: 'matches',
    initialState: initialState,
    reducers: {
        setProfileDataDiff: (state: MatchesState, action: PayloadAction<MatchDiff>) => {
            state.profileDataDiff = action.payload;
        },
    },
    extraReducers: builder => {
        builder.addMatcher(matchesApiSlice.endpoints.getMatches.matchFulfilled, (state, { payload }) => {
            state.matchPairs = payload.matchPairs;
            state.totalMatches = payload.totalMatchPairs;
            state.matchTableItems = convertMatchPairsToTableItems(payload.matchPairs);
        });
    },
});

export const { setProfileDataDiff } = matchesSlice.actions;
export const selectProfileDataDiff = (state: RootState) => state.matches.profileDataDiff;
export const selectMatchPairs = (state: RootState) => state.matches.matchPairs;
export const selectMatchTableItems = (state: RootState) => state.matches.matchTableItems;
export const selectTotalMatchCount = (state: RootState) => state.matches.totalMatches;
