// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import type { PayloadAction } from '@reduxjs/toolkit';
import { createSlice } from '@reduxjs/toolkit';
import { RuleSet, SkipCondition } from '../models/config';
import { DisplayRule, MatchFilterCondition } from '../models/ruleSet';
import { ruleSetApiSlice } from './ruleSetApiSlice';
import { RootState } from './store';

interface RuleSetState {
    allRuleSets: RuleSet[];
    allRuleSetsCache: RuleSet[];
    selectedRuleSet: string;
    selectedRuleSetCache: string;
    editableRuleSet: DisplayRule[];
    isEditPage: boolean;
    profileMappings: string[];
}

const initialState: RuleSetState = {
    allRuleSets: [],
    allRuleSetsCache: [],
    editableRuleSet: [],
    selectedRuleSet: 'active',
    selectedRuleSetCache: 'active',
    isEditPage: false,
    profileMappings: [],
};

export const ruleSetSlice = createSlice({
    name: 'ruleSet',
    initialState,
    reducers: {
        setEditableRuleSet(state, action: PayloadAction<DisplayRule[]>) {
            state.editableRuleSet = action.payload;
        },
        clearEditableRuleSet(state, action: PayloadAction<void>) {
            state.editableRuleSet = [];
        },
        setSkipCondition(
            state,
            action: PayloadAction<{
                condition: Partial<SkipCondition>;
                ruleIndex: number;
                conditionIndex: number;
            }>,
        ) {
            const existingCondition = state.editableRuleSet[action.payload.ruleIndex].skipConditions[action.payload.conditionIndex];
            state.editableRuleSet[action.payload.ruleIndex].skipConditions[action.payload.conditionIndex] = {
                ...existingCondition,
                ...action.payload.condition,
            };
        },
        setMatchFilterCondition(
            state,
            action: PayloadAction<{
                condition: Partial<MatchFilterCondition>;
                ruleIndex: number;
                conditionIndex: number;
            }>,
        ) {
            const skipConditionLength = state.editableRuleSet[action.payload.ruleIndex].skipConditions.length;
            const matchFilterIndex = action.payload.conditionIndex - skipConditionLength;
            const existingCondition = state.editableRuleSet[action.payload.ruleIndex].matchFilterConditions[matchFilterIndex];

            state.editableRuleSet[action.payload.ruleIndex].matchFilterConditions[matchFilterIndex] = {
                ...existingCondition,
                ...action.payload.condition,
            };
        },
        addSkipCondition(
            state,
            action: PayloadAction<{
                condition: SkipCondition;
                ruleIndex: number;
            }>,
        ) {
            state.editableRuleSet[action.payload.ruleIndex].skipConditions.push(action.payload.condition);
            state.editableRuleSet[action.payload.ruleIndex].matchFilterConditions = state.editableRuleSet[
                action.payload.ruleIndex
            ].matchFilterConditions.map(cond => {
                return { ...cond, index: cond.index + 1 };
            });
        },
        addMatchFilterCondition(
            state,
            action: PayloadAction<{
                condition: MatchFilterCondition;
                ruleIndex: number;
            }>,
        ) {
            state.editableRuleSet[action.payload.ruleIndex].matchFilterConditions.push(action.payload.condition);
        },
        deleteSkipCondition(state, action: PayloadAction<{ ruleIndex: number; conditionIndex: number }>) {
            const conditions = state.editableRuleSet[action.payload.ruleIndex].skipConditions;
            conditions.splice(action.payload.conditionIndex, 1);
            for (let i = action.payload.conditionIndex; i < conditions.length; i++) {
                conditions[i].index--;
            }
            state.editableRuleSet[action.payload.ruleIndex].matchFilterConditions = state.editableRuleSet[
                action.payload.ruleIndex
            ].matchFilterConditions.map(cond => {
                return { ...cond, index: cond.index - 1 };
            });
        },
        deleteMatchFilterCondition(state, action: PayloadAction<{ ruleIndex: number; conditionIndex: number }>) {
            const skipConditionLength = state.editableRuleSet[action.payload.ruleIndex].skipConditions.length;
            const matchFilterIndex = action.payload.conditionIndex - skipConditionLength;
            const conditions = state.editableRuleSet[action.payload.ruleIndex].matchFilterConditions;
            conditions.splice(matchFilterIndex, 1);
            for (let i = matchFilterIndex; i < conditions.length; i++) {
                conditions[i].index--;
            }
        },
        addRule(state, action: PayloadAction<void>) {
            const newRule: DisplayRule = {
                index: state.editableRuleSet.length,
                name: '',
                description: '',
                skipConditions: [],
                matchFilterConditions: [],
            };
            state.editableRuleSet.push(newRule);
        },
        deleteRule(state, action: PayloadAction<{ ruleIndex: number }>) {
            state.editableRuleSet.splice(action.payload.ruleIndex, 1);
            for (let i = action.payload.ruleIndex; i < state.editableRuleSet.length; i++) {
                state.editableRuleSet[i].index--;
            }
        },
        setSelectedRuleSet(state, action: PayloadAction<string>) {
            state.selectedRuleSet = action.payload;
        },
        setSelectedRuleSetCache(state, action: PayloadAction<string>) {
            state.selectedRuleSetCache = action.payload;
        },
        setIsEditPage(state, action: PayloadAction<boolean>) {
            state.isEditPage = action.payload;
        },
    },
    extraReducers: builder => {
        builder.addMatcher(ruleSetApiSlice.endpoints.listRuleSets.matchFulfilled, (state, { payload }) => {
            state.allRuleSets = payload.ruleSets;
            state.profileMappings = payload.profileMappings;
        });
        builder.addMatcher(ruleSetApiSlice.endpoints.listRuleSetsCache.matchFulfilled, (state, { payload }) => {
            state.allRuleSetsCache = payload.ruleSets;
            state.profileMappings = payload.profileMappings;
        });
        builder.addMatcher(ruleSetApiSlice.endpoints.listRuleSets.matchRejected, (state, _) => {
            state.allRuleSets = [];
            state.profileMappings = [];
        });
        builder.addMatcher(ruleSetApiSlice.endpoints.listRuleSetsCache.matchRejected, (state, _) => {
            state.allRuleSetsCache = [];
            state.profileMappings = [];
        });
        builder.addMatcher(ruleSetApiSlice.endpoints.listRuleSets.matchPending, (state, _) => {
            state.allRuleSets = [];
            state.profileMappings = [];
        });
        builder.addMatcher(ruleSetApiSlice.endpoints.listRuleSetsCache.matchPending, (state, _) => {
            state.allRuleSetsCache = [];
            state.profileMappings = [];
        });
    },
});
export const {
    setEditableRuleSet,
    clearEditableRuleSet,
    setSkipCondition,
    setMatchFilterCondition,
    addSkipCondition,
    addMatchFilterCondition,
    deleteSkipCondition,
    deleteMatchFilterCondition,
    addRule,
    deleteRule,
    setSelectedRuleSet,
    setSelectedRuleSetCache,
    setIsEditPage,
} = ruleSetSlice.actions;

export default ruleSetSlice.reducer;

export const selectAllRuleSets = (state: RootState) => state.ruleSet.allRuleSets;
export const selectAllRuleSetsCache = (state: RootState) => state.ruleSet.allRuleSetsCache;
export const selectSelectedRuleSet = (state: RootState) => state.ruleSet.selectedRuleSet;
export const selectSelectedRuleSetCache = (state: RootState) => state.ruleSet.selectedRuleSetCache;
export const selectEditableRuleSet = (state: RootState) => state.ruleSet.editableRuleSet;
export const selectIsEditPage = (state: RootState) => state.ruleSet.isEditPage;
export const selectProfileMappings = (state: RootState) => state.ruleSet.profileMappings;
