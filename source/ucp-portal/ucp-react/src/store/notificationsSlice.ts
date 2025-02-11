// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { FlashbarProps } from '@cloudscape-design/components';
import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import React from 'react';
import { AsyncEventStatus, AsyncEventType } from '../models/async.ts';
import { TxIdResponse } from '../models/config.ts';
import { asyncApiSlice } from './asyncApiSlice.ts';
import { cacheApiSlice } from './cacheApiSlice.ts';
import { configApiSlice } from './configApiSlice.ts';
import { domainApiSlice } from './domainApiSlice.ts';
import { errorApiSlice } from './errorApiSlice.ts';
import { jobsApiSlice } from './jobsApiSlice.ts';
import { privacyApiSlice } from './privacyApiSlice.ts';
import { profileApiSlice } from './profileApiSlice.ts';
import { ruleSetApiSlice } from './ruleSetApiSlice.ts';
import { RootState } from './store.ts';
import { matchesApiSlice } from './matchesApiSlice.ts';

export type NotificationPayload = {
    id: string;
    header?: React.ReactNode;
    content?: React.ReactNode;
    type: FlashbarProps.Type;
    loading?: boolean;
};

export type NotificationState = {
    notifications: NotificationPayload[];
};

const initialState: NotificationState = {
    notifications: [],
};

export const notificationsSlice = createSlice({
    name: 'notifications',
    initialState,
    reducers: {
        addNotification: (state: NotificationState, action: PayloadAction<NotificationPayload>) => {
            const notification = action.payload;
            if (!state.notifications.find(it => it.id === notification.id)) state.notifications.push(notification);
        },
        deleteNotification: (state: NotificationState, action: PayloadAction<{ id: string }>) => {
            state.notifications = state.notifications.filter(it => it.id !== action.payload.id);
        },
        clearNotifications: (state: NotificationState) => {
            state.notifications = [];
        },
    },
    extraReducers(builder) {
        builder
            .addMatcher(domainApiSlice.endpoints.createDomain.matchFulfilled, (state, { payload }) => {
                state.notifications.push({
                    id: payload.item_type,
                    type: 'in-progress',
                    content: 'Creating Domain',
                    loading: true,
                });
            })
            .addMatcher(domainApiSlice.endpoints.deleteDomain.matchFulfilled, (state, { payload }) => {
                state.notifications.push({
                    id: payload.item_type,
                    type: 'in-progress',
                    content: 'Deleting Domain',
                    loading: true,
                });
            })
            .addMatcher(errorApiSlice.endpoints.deleteAllErrors.matchFulfilled, (state, { payload }) => {
                state.notifications.push({
                    id: payload.item_type,
                    type: 'in-progress',
                    content: 'Deleting All Errors',
                    loading: true,
                });
            })
            .addMatcher(asyncApiSlice.endpoints.getAsyncStatus.matchFulfilled, (state, { payload }) => {
                if (payload.status === AsyncEventStatus.EVENT_STATUS_SUCCESS) {
                    notificationsSlice.caseReducers.deleteNotification(state, {
                        payload: { id: payload.item_type },
                        type: 'notificationsSlice/deleteNotification',
                    });
                    const notificationPayload: NotificationPayload = {
                        id: payload.item_type,
                        type: 'success',
                    };
                    if (payload.item_id === AsyncEventType.CREATE_DOMAIN) {
                        notificationPayload.content = 'Successfully Created Domain';
                    } else if (payload.item_id === AsyncEventType.DELETE_DOMAIN) {
                        notificationPayload.content = 'Successfully Deleted Domain';
                    } else if (payload.item_id === AsyncEventType.EMPTY_TABLE) {
                        notificationPayload.content = 'Successfully Emptied Error Table';
                    } else if (payload.item_id === AsyncEventType.CREATE_PRIVACY_SEARCH) {
                        notificationPayload.content = 'Profile Data Search Completed';
                    } else if (payload.item_id === AsyncEventType.CREATE_PRIVACY_PURGE) {
                        notificationPayload.content = 'Purge Completed';
                    } else if (payload.item_id === AsyncEventType.UNMERGE_PROFILES) {
                        notificationPayload.content = 'Unmerge Completed';
                    } else if (payload.item_id === AsyncEventType.MERGE_PROFILES) {
                        notificationPayload.content = 'Merge Completed';
                    } else if (payload.item_id === AsyncEventType.REBUILD_CACHE) {
                        notificationPayload.content = 'Rebuild Cache Fargate Tasks Started';
                    } else {
                        notificationPayload.content = 'Unknown Event Type Occurred';
                        notificationPayload.type = 'warning';
                    }
                    notificationsSlice.caseReducers.addNotification(state, {
                        payload: notificationPayload,
                        type: 'notificationsSlice/addNotification',
                    });
                } else if (payload.status === AsyncEventStatus.EVENT_STATUS_FAILED) {
                    notificationsSlice.caseReducers.deleteNotification(state, {
                        payload: { id: payload.item_type },
                        type: 'notificationsSlice/deleteNotification',
                    });
                    const notificationPayload: NotificationPayload = {
                        id: payload.item_type,
                        type: 'error',
                    };
                    if (payload.item_id === AsyncEventType.CREATE_DOMAIN) {
                        notificationPayload.content = 'Unable to Create Domain';
                    } else if (payload.item_id === AsyncEventType.DELETE_DOMAIN) {
                        notificationPayload.content = 'Unable to Delete Domain';
                    } else if (payload.item_id === AsyncEventType.EMPTY_TABLE) {
                        notificationPayload.content = 'Unable to Empty Error Table';
                    } else if (payload.item_id === AsyncEventType.CREATE_PRIVACY_SEARCH) {
                        notificationPayload.content = `An error occurred while performing the Privacy Search. See errors in grid for details. EventID: ${payload.item_type}`;
                    } else if (payload.item_id === AsyncEventType.CREATE_PRIVACY_PURGE) {
                        notificationPayload.content = 'Purge Failed';
                    } else if (payload.item_id === AsyncEventType.UNMERGE_PROFILES) {
                        notificationPayload.content = 'Unmerge Failed';
                    } else if (payload.item_id === AsyncEventType.MERGE_PROFILES) {
                        notificationPayload.content = 'Merge Failed';
                    } else if (payload.item_id === AsyncEventType.REBUILD_CACHE) {
                        notificationPayload.content = 'Failed to Start Rebuild Cache Fargate Tasks';
                    } else {
                        notificationPayload.content = 'Unknown Event Type Occurred';
                        notificationPayload.type = 'warning';
                    }
                    notificationsSlice.caseReducers.addNotification(state, {
                        payload: notificationPayload,
                        type: 'notificationsSlice/addNotification',
                    });
                } else if (
                    payload.status === AsyncEventStatus.EVENT_STATUS_INVOKED ||
                    payload.status === AsyncEventStatus.EVENT_STATUS_RUNNING
                ) {
                    if (payload.item_id === AsyncEventType.CREATE_PRIVACY_SEARCH) {
                        if (!state.notifications.some(notification => notification.id === payload.item_type)) {
                            state.notifications.push({
                                id: payload.item_type,
                                type: 'in-progress',
                                loading: true,
                                content: 'Searching for Profile Data',
                            });
                        }
                    } else if (payload.item_id === AsyncEventType.UNMERGE_PROFILES) {
                        if (!state.notifications.some(notification => notification.id === payload.item_type)) {
                            state.notifications.push({
                                id: payload.item_type,
                                type: 'in-progress',
                                loading: true,
                                content: 'Unmerging profiles',
                            });
                        }
                    } else if (payload.item_id === AsyncEventType.MERGE_PROFILES) {
                        if (!state.notifications.some(notification => notification.id === payload.item_type)) {
                            state.notifications.push({
                                id: payload.item_type,
                                type: 'in-progress',
                                loading: true,
                                content: 'Merging profiles',
                            });
                        }
                    }
                }
            })
            .addMatcher(jobsApiSlice.endpoints.runJob.matchFulfilled, (state, { payload }) => {
                state.notifications.push({
                    id: payload.TxID,
                    type: 'success',
                    content: 'Successfully Ran Glue Job',
                });
            })
            .addMatcher(jobsApiSlice.endpoints.runAllJobs.matchFulfilled, (state, { payload }) => {
                state.notifications.push({
                    id: payload.TxID,
                    type: 'success',
                    content: 'Successfully Ran All Glue Jobs',
                });
            })
            .addMatcher(jobsApiSlice.endpoints.runAllJobs.matchRejected, (state, { payload }) => {
                state.notifications.push({
                    id: 'FAILED RUN ALL',
                    type: 'error',
                    content: 'Run All Jobs Request Failed',
                });
            })
            .addMatcher(configApiSlice.endpoints.postPortalConfig.matchRejected, (state, action) => {
                state.notifications.push({
                    id: 'Failed to Update Config',
                    type: 'error',
                    content: 'Failed to Update Config: ' + action.error,
                });
            })
            .addMatcher(privacyApiSlice.endpoints.createSearch.matchFulfilled, (state, { payload }) => {
                state.notifications.push({
                    id: payload.item_type,
                    type: 'in-progress',
                    loading: true,
                    content: 'Searching for Profile Data. Click the Refresh icon above the results grid to get the most up to date status.',
                });
            })
            .addMatcher(privacyApiSlice.endpoints.createSearch.matchRejected, (state, action) => {
                let errMsg = 'Unknown Error';
                let txId = '';
                if (action.payload?.status === 'CUSTOM_ERROR') {
                    errMsg = action.payload?.error;
                    if (
                        action.payload?.data !== undefined &&
                        action.payload?.data !== null &&
                        Object.hasOwn(action.payload?.data, 'TxID')
                    ) {
                        const d = action.payload.data as TxIdResponse;
                        txId = `TxID: ${d.TxID}`;
                    }
                }
                state.notifications.push({
                    id: action.meta.requestId,
                    type: 'error',
                    content: [
                        `Failed to create search: ${errMsg}`,
                        React.createElement('br', { key: `${action.meta.requestId}br1` }),
                        txId,
                    ],
                });
            })
            .addMatcher(privacyApiSlice.endpoints.deleteSearches.matchFulfilled, (state, { meta, payload }) => {
                state.notifications = state.notifications.filter(it => it.id !== meta.requestId);
                state.notifications.push({
                    id: payload.TxID,
                    type: 'success',
                    loading: false,
                    content: 'Successfully deleted privacy searches',
                });
            })
            .addMatcher(privacyApiSlice.endpoints.deleteSearches.matchPending, (state, action) => {
                state.notifications.push({
                    id: action.meta.requestId,
                    type: 'in-progress',
                    loading: true,
                    content: 'Deleting privacy searches',
                });
            })
            .addMatcher(privacyApiSlice.endpoints.deleteSearches.matchRejected, (state, action) => {
                let errMsg = 'Unknown Error';
                let txId = '';
                if (action.payload?.status === 'CUSTOM_ERROR') {
                    errMsg = action.payload?.error;
                    if (
                        action.payload?.data !== undefined &&
                        action.payload?.data !== null &&
                        Object.hasOwn(action.payload?.data, 'TxID')
                    ) {
                        const d = action.payload.data as TxIdResponse;
                        txId = `TxID: ${d.TxID}`;
                    }
                }
                state.notifications.push({
                    id: action.meta.requestId,
                    type: 'error',
                    loading: false,
                    content: `Failed to delete privacy searches: ${errMsg} ${txId}`,
                });
            })
            .addMatcher(privacyApiSlice.endpoints.purgeProfile.matchPending, (state, { payload }) => {
                state.notifications.push({
                    id: 'purgeRequestInitiated',
                    type: 'info',
                    content: 'Profile Data Purge Request Received',
                });
            })
            .addMatcher(privacyApiSlice.endpoints.getPurgeIsRunning.matchFulfilled, (state, { payload }) => {
                if (state.notifications.some(n => n.id === 'purgeRequestInitiated')) {
                    notificationsSlice.caseReducers.deleteNotification(state, {
                        payload: { id: 'purgeRequestInitiated' },
                        type: 'notificationsSlice/deleteNotification',
                    });
                }

                const notificationId = 'purgeInProgressNotification';
                if (payload.isPurgeRunning === true) {
                    if (!state.notifications.some(n => n.id === notificationId)) {
                        state.notifications.push({
                            id: notificationId,
                            type: 'in-progress',
                            loading: true,
                            content: 'Purge in Progress. Click the Refresh icon above the results grid to get the most up to date status.',
                        });
                    }
                } else {
                    notificationsSlice.caseReducers.deleteNotification(state, {
                        payload: { id: notificationId },
                        type: 'notificationsSlice/deleteNotification',
                    });
                    state.notifications.push({
                        id: notificationId,
                        type: 'info',
                        content: 'Purge Run Completed. See Grid for Details.',
                    });
                }
            })
            .addMatcher(ruleSetApiSlice.endpoints.saveRuleSet.matchFulfilled, (state, { payload }) => {
                state.notifications.push({
                    id: payload.TxID,
                    type: 'success',
                    content: 'Successfully Saved Rule Set to Draft',
                });
            })
            .addMatcher(ruleSetApiSlice.endpoints.saveRuleSet.matchRejected, (state, { meta, error }) => {
                state.notifications.push({
                    id: meta.requestId,
                    type: 'error',
                    content: `Failed to save rule set: ${error.message ?? 'Unknown Error'}`,
                });
            })
            .addMatcher(profileApiSlice.endpoints.getProfilesSummary.matchPending, state => {
                state.notifications.push({
                    id: 'getProfilesSummaryInitiated',
                    type: 'in-progress',
                    loading: true,
                    content: 'Generating profile summary',
                });
            })
            .addMatcher(profileApiSlice.endpoints.getProfilesSummary.matchFulfilled, (state, { payload }) => {
                if (state.notifications.some(n => n.id === 'getProfilesSummaryInitiated')) {
                    notificationsSlice.caseReducers.deleteNotification(state, {
                        payload: { id: 'getProfilesSummaryInitiated' },
                        type: 'notificationsSlice/deleteNotification',
                    });
                }
                state.notifications.push({
                    id: 'getProfilesSummaryFulfilled',
                    type: 'success',
                    content: 'Profile summary generated',
                });
            })
            .addMatcher(
                profileApiSlice.endpoints.getProfilesById.matchPending || profileApiSlice.endpoints.getProfilesByParam.matchPending,
                state => {
                    state.notifications.push({
                        id: 'getProfilesByIdInitiated',
                        type: 'in-progress',
                        loading: true,
                        content: 'Searching for profiles',
                    });
                },
            )
            .addMatcher(
                profileApiSlice.endpoints.getProfilesById.matchFulfilled || profileApiSlice.endpoints.getProfilesByParam.matchFulfilled,
                (state, { payload }) => {
                    if (state.notifications.some(n => n.id === 'getProfilesByIdInitiated')) {
                        notificationsSlice.caseReducers.deleteNotification(state, {
                            payload: { id: 'getProfilesByIdInitiated' },
                            type: 'notificationsSlice/deleteNotification',
                        });
                    }
                    state.notifications.push({
                        id: 'getProfilesByIdFulfilled',
                        type: 'success',
                        content: payload.profiles.length + ' profile' + (payload.profiles.length !== 1 ? 's' : '') + ' found',
                    });
                },
            )
            .addMatcher(
                profileApiSlice.endpoints.getProfilesById.matchRejected || profileApiSlice.endpoints.getProfilesByParam.matchRejected,
                (state, { payload }) => {
                    if (state.notifications.some(n => n.id === 'getProfilesByIdInitiated')) {
                        notificationsSlice.caseReducers.deleteNotification(state, {
                            payload: { id: 'getProfilesByIdInitiated' },
                            type: 'notificationsSlice/deleteNotification',
                        });
                    }
                    state.notifications.push({
                        id: 'getProfilesByIdFulfilled',
                        type: 'error',
                        content: 'Profile not found',
                    });
                },
            )

            .addMatcher(domainApiSlice.endpoints.savePromptConfig.matchPending, state => {
                state.notifications.push({
                    id: 'savePromptConfigInitiated',
                    type: 'in-progress',
                    loading: true,
                    content: 'Saving prompt configuration',
                });
            })
            .addMatcher(cacheApiSlice.endpoints.rebuildCache.matchFulfilled, state => {
                state.notifications.push({
                    id: 'rebuildCacheInitiated',
                    type: 'in-progress',
                    loading: true,
                    content: 'Rebuild Cache Initiated',
                });
            })
            .addMatcher(domainApiSlice.endpoints.savePromptConfig.matchFulfilled, (state, { payload }) => {
                if (state.notifications.some(n => n.id === 'savePromptConfigInitiated')) {
                    notificationsSlice.caseReducers.deleteNotification(state, {
                        payload: { id: 'savePromptConfigInitiated' },
                        type: 'notificationsSlice/deleteNotification',
                    });
                }
                state.notifications.push({
                    id: 'savePromptConfigFulfilled',
                    type: 'success',
                    content: 'Prompt configuration saved',
                });
            })
            .addMatcher(matchesApiSlice.endpoints.performMerge.matchPending, state => {
                state.notifications.push({
                    id: 'performMerge',
                    type: 'in-progress',
                    loading: true,
                    content: 'Operation initiated',
                });
            })
            .addMatcher(matchesApiSlice.endpoints.performMerge.matchRejected, (state, { payload }) => {
                if (state.notifications.some(n => n.id === 'performMerge')) {
                    notificationsSlice.caseReducers.deleteNotification(state, {
                        payload: { id: 'performMerge' },
                        type: 'notificationsSlice/deleteNotification',
                    });
                }
                state.notifications.push({
                    id: 'performMergeFailed',
                    type: 'error',
                    content: 'Unable to perform operation',
                });
            })
            .addMatcher(matchesApiSlice.endpoints.performMerge.matchFulfilled, (state, { payload }) => {
                if (state.notifications.some(n => n.id === 'performMerge')) {
                    notificationsSlice.caseReducers.deleteNotification(state, {
                        payload: { id: 'performMerge' },
                        type: 'notificationsSlice/deleteNotification',
                    });
                }
                state.notifications.push({
                    id: 'performMergeFulfilled',
                    type: 'success',
                    content: 'Operation performed successfully',
                });
            });
    },
});

export default notificationsSlice.reducer;
export const selectNotifications = (state: RootState) => state.notifications.notifications;
export const { addNotification, deleteNotification, clearNotifications } = notificationsSlice.actions;
