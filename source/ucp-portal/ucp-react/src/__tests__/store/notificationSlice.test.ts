// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { expect, test } from 'vitest';
import reducer, {
    NotificationPayload,
    NotificationState,
    addNotification,
    clearNotifications,
    deleteNotification,
} from '../../store/notificationsSlice';
import React from 'react';

const initialState: NotificationState = {
    notifications: [],
};

const TEST_ID = 'test_id';

const testNotification: NotificationPayload = {
    id: TEST_ID,
    type: 'success',
    content: 'test_message',
};

const asyncRunningNotification: NotificationPayload = {
    id: TEST_ID,
    type: 'in-progress',
    content: 'test_in_progress_message',
};

test('Should Return Initial State', () => {
    expect(reducer(undefined, { type: undefined })).toEqual(initialState);
});

test('Should handle adding notification', () => {
    expect(reducer(initialState, addNotification(testNotification))).toEqual({ notifications: [testNotification] });
});

test('Should handle deleting notification', () => {
    expect(reducer({ notifications: [testNotification] }, deleteNotification({ id: TEST_ID }))).toEqual(initialState);
});

test('POST create domain Api match fulfilled', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/fulfilled',
            meta: { arg: { endpointName: 'createDomain' } },
            payload: {
                item_type: TEST_ID,
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'in-progress',
                content: 'Creating Domain',
                loading: true,
            },
        ],
    });
});

test('DELETE domain Api match fulfilled', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/fulfilled',
            meta: { arg: { endpointName: 'deleteDomain' } },
            payload: {
                item_type: TEST_ID,
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'in-progress',
                content: 'Deleting Domain',
                loading: true,
            },
        ],
    });
});

test('DELETE all errors Api match fulfilled', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/fulfilled',
            meta: { arg: { endpointName: 'deleteAllErrors' } },
            payload: {
                item_type: TEST_ID,
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'in-progress',
                content: 'Deleting All Errors',
                loading: true,
            },
        ],
    });
});

test('GET createDomain async status Api match fulfilled (successful async case)', () => {
    expect(
        reducer(
            { notifications: [asyncRunningNotification] },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getAsyncStatus' } },
                payload: {
                    item_id: 'createDomain',
                    item_type: TEST_ID,
                    status: 'success',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'success',
                content: 'Successfully Created Domain',
            },
        ],
    });
});

test('GET createDomain async status Api match fulfilled (running async case)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeQuery/fulfilled',
            meta: { arg: { endpointName: 'getAsyncStatus' } },
            payload: {
                item_id: 'createDomain',
                status: 'running',
            },
        }),
    ).toEqual(initialState);
});

test('GET createDomain async status Api match fulfilled (failed async case)', () => {
    expect(
        reducer(
            { notifications: [asyncRunningNotification] },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getAsyncStatus' } },
                payload: {
                    item_id: 'createDomain',
                    item_type: TEST_ID,
                    status: 'failed',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'error',
                content: 'Unable to Create Domain',
            },
        ],
    });
});

test('GET deleteDomain async status Api match fulfilled (successful async case)', () => {
    expect(
        reducer(
            { notifications: [asyncRunningNotification] },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getAsyncStatus' } },
                payload: {
                    item_id: 'deleteDomain',
                    item_type: TEST_ID,
                    status: 'success',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'success',
                content: 'Successfully Deleted Domain',
            },
        ],
    });
});

test('GET deleteDomain async status Api match fulfilled (running async case)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeQuery/fulfilled',
            meta: { arg: { endpointName: 'getAsyncStatus' } },
            payload: {
                item_id: 'deleteDomain',
                status: 'running',
            },
        }),
    ).toEqual(initialState);
});

test('GET deleteDomain async status Api match fulfilled (failed async case)', () => {
    expect(
        reducer(
            { notifications: [asyncRunningNotification] },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getAsyncStatus' } },
                payload: {
                    item_id: 'deleteDomain',
                    item_type: TEST_ID,
                    status: 'failed',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'error',
                content: 'Unable to Delete Domain',
            },
        ],
    });
});

test('GET emptyErrorTable async status Api match fulfilled (successful async case)', () => {
    expect(
        reducer(
            { notifications: [asyncRunningNotification] },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getAsyncStatus' } },
                payload: {
                    item_id: 'emptyTable',
                    item_type: TEST_ID,
                    status: 'success',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'success',
                content: 'Successfully Emptied Error Table',
            },
        ],
    });
});

test('GET emptyErrorTable async status Api match fulfilled (failed async case)', () => {
    expect(
        reducer(
            { notifications: [asyncRunningNotification] },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getAsyncStatus' } },
                payload: {
                    item_id: 'emptyTable',
                    item_type: TEST_ID,
                    status: 'failed',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'error',
                content: 'Unable to Empty Error Table',
            },
        ],
    });
});

test('GET unknownRequest async status Api match fulfilled (successful async case)', () => {
    expect(
        reducer(
            { notifications: [asyncRunningNotification] },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getAsyncStatus' } },
                payload: {
                    item_id: 'unknownRequest',
                    item_type: TEST_ID,
                    status: 'success',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'warning',
                content: 'Unknown Event Type Occurred',
            },
        ],
    });
});

test('POST runJob endpoint match fulfilled)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/fulfilled',
            meta: { arg: { endpointName: 'runJob' } },
            payload: {
                TxID: TEST_ID,
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'success',
                content: 'Successfully Ran Glue Job',
            },
        ],
    });
});

test('POST runAllJobs endpoint match fulfilled)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/fulfilled',
            meta: { arg: { endpointName: 'runAllJobs' } },
            payload: {
                TxID: TEST_ID,
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'success',
                content: 'Successfully Ran All Glue Jobs',
            },
        ],
    });
});

test('POST runAllJobs endpoint match rejected)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/rejected',
            meta: { arg: { endpointName: 'runAllJobs' } },
            payload: {},
        }),
    ).toEqual({ notifications: [{ id: 'FAILED RUN ALL', type: 'error', content: 'Run All Jobs Request Failed' }] });
});

test('POST createSearch endpoint match fulfilled)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/fulfilled',
            meta: { arg: { endpointName: 'createSearch' } },
            payload: {
                item_type: TEST_ID,
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'in-progress',
                loading: true,
                content: 'Searching for Profile Data. Click the Refresh icon above the results grid to get the most up to date status.',
            },
        ],
    });
});

test('POST createSearch endpoint match rejected)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/rejected',
            meta: { arg: { endpointName: 'createSearch' }, requestId: TEST_ID },
            payload: {
                status: 'CUSTOM_ERROR',
                error: 'custom',
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'error',
                content: [`Failed to create search: custom`, React.createElement('br', { key: `${TEST_ID}br1` }), ''],
            },
        ],
    });
});

test('POST deleteSearches endpoint match fulfilled)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/fulfilled',
            meta: { arg: { endpointName: 'deleteSearches' } },
            payload: {
                TxID: TEST_ID,
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'success',
                loading: false,
                content: 'Successfully deleted privacy searches',
            },
        ],
    });
});

test('POST deleteSearches endpoint match pending)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/pending',
            meta: { arg: { endpointName: 'deleteSearches' }, requestId: TEST_ID },
            payload: {
                TxID: TEST_ID,
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'in-progress',
                loading: true,
                content: 'Deleting privacy searches',
            },
        ],
    });
});

test('POST deleteSearches endpoint match rejected)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/rejected',
            meta: { arg: { endpointName: 'deleteSearches' }, requestId: TEST_ID },
            payload: {
                status: 'CUSTOM_ERROR',
                error: 'custom',
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'error',
                loading: false,
                content: `Failed to delete privacy searches: custom `,
            },
        ],
    });
});

test('POST purgeProfile endpoint match pending)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/pending',
            meta: { arg: { endpointName: 'purgeProfile' } },
        }),
    ).toEqual({
        notifications: [
            {
                id: 'purgeRequestInitiated',
                type: 'info',
                content: 'Profile Data Purge Request Received',
            },
        ],
    });
});

test('POST portalConfig endpoint match rejected)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/rejected',
            meta: { arg: { endpointName: 'postPortalConfig' } },
            error: 'Invalid Input',
        }),
    ).toEqual({
        notifications: [{ id: 'Failed to Update Config', type: 'error', content: 'Failed to Update Config: ' + 'Invalid Input' }],
    });
});

test('POST getPurgeIsRunning endpoint match fulfilled)', () => {
    expect(
        reducer(
            {
                notifications: [
                    {
                        id: 'purgeRequestInitiated',
                        type: 'info',
                        content: 'Profile Data Purge Request Received',
                    },
                ],
            },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getPurgeIsRunning' } },
                payload: {
                    TxID: TEST_ID,
                    isPurgeRunning: true,
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: 'purgeInProgressNotification',
                type: 'in-progress',
                loading: true,
                content: 'Purge in Progress. Click the Refresh icon above the results grid to get the most up to date status.',
            },
        ],
    });

    expect(
        reducer(
            {
                notifications: [
                    {
                        id: 'purgeRequestInitiated',
                        type: 'info',
                        content: 'Profile Data Purge Request Received',
                    },
                    {
                        id: 'purgeInProgressNotification',
                        type: 'in-progress',
                        loading: true,
                        content: 'Purge in Progress. Click the Refresh icon above the results grid to get the most up to date status.',
                    },
                ],
            },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getPurgeIsRunning' } },
                payload: {
                    TxID: TEST_ID,
                    isPurgeRunning: false,
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: 'purgeInProgressNotification',
                type: 'info',
                content: 'Purge Run Completed. See Grid for Details.',
            },
        ],
    });
});

test('POST saveRuleSet endpoint match fulfilled)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/fulfilled',
            meta: { arg: { endpointName: 'saveRuleSet' }, requestId: TEST_ID },
            payload: {
                TxID: TEST_ID,
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'success',
                content: 'Successfully Saved Rule Set to Draft',
            },
        ],
    });
});

test('POST saveRuleSet endpoint match rejected)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/rejected',
            meta: { arg: { endpointName: 'saveRuleSet' }, requestId: TEST_ID },
            error: 'Invalid Input',
        }),
    ).toEqual({
        notifications: [
            {
                id: TEST_ID,
                type: 'error',
                content: `Failed to save rule set: Unknown Error`,
            },
        ],
    });
});

test('POST getProfilesSummary Api match pending', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeQuery/pending',
            meta: { arg: { endpointName: 'getProfilesSummary' } },
            payload: {
                id: 'getProfilesSummaryInitiated',
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: 'getProfilesSummaryInitiated',
                type: 'in-progress',
                loading: true,
                content: 'Generating profile summary',
            },
        ],
    });
});

test('POST getProfilesSummary Api match fulfilled', () => {
    expect(
        reducer(
            {
                notifications: [
                    {
                        id: 'getProfilesSummaryInitiated',
                        type: 'in-progress',
                        loading: true,
                        content: 'Generating profile summary',
                    },
                ],
            },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getProfilesSummary' } },
                payload: {
                    id: 'getProfilesSummaryInitiated',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: 'getProfilesSummaryFulfilled',
                type: 'success',
                content: 'Profile summary generated',
            },
        ],
    });
});

test('POST getProfilesById Api match pending', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeQuery/pending',
            meta: { arg: { endpointName: 'getProfilesById' } },
            payload: {
                id: 'getProfilesById',
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: 'getProfilesByIdInitiated',
                type: 'in-progress',
                loading: true,
                content: 'Searching for profiles',
            },
        ],
    });
});

test('POST getProfilesById Api match rejected', () => {
    expect(
        reducer(
            {
                notifications: [
                    {
                        id: 'getProfilesByIdInitiated',
                        type: 'in-progress',
                        loading: true,
                        content: 'Searching for profiles',
                    },
                ],
            },
            {
                type: 'solution-api/executeQuery/rejected',
                meta: { arg: { endpointName: 'getProfilesById' } },
                payload: {
                    id: 'getProfilesById',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: 'getProfilesByIdFulfilled',
                type: 'error',
                content: 'Profile not found',
            },
        ],
    });
});

test('POST getProfilesById Api match fulfilled', () => {
    expect(
        reducer(
            {
                notifications: [
                    {
                        id: 'getProfilesByIdInitiated',
                        type: 'in-progress',
                        loading: true,
                        content: 'Searching for profiles',
                    },
                ],
            },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getProfilesById' } },
                payload: {
                    id: 'getProfilesById',
                    profiles: [{}],
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: 'getProfilesByIdFulfilled',
                type: 'success',
                content: '1 profile found',
            },
        ],
    });
});

test('POST rebuildCache endpoint match fulfilled)', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/fulfilled',
            meta: { arg: { endpointName: 'rebuildCache' } },
            payload: {},
        }),
    ).toEqual({
        notifications: [
            {
                id: 'rebuildCacheInitiated',
                type: 'in-progress',
                loading: true,
                content: 'Rebuild Cache Initiated',
            },
        ],
    });
});

test('POST savePromptConfig Api match pending', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/pending',
            meta: { arg: { endpointName: 'savePromptConfig' } },
            payload: {
                id: 'savePromptConfig',
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: 'savePromptConfigInitiated',
                type: 'in-progress',
                loading: true,
                content: 'Saving prompt configuration',
            },
        ],
    });
});

test('POST savePromptConfig Api match fulfilled', () => {
    expect(
        reducer(
            {
                notifications: [
                    {
                        id: 'savePromptConfigInitiated',
                        type: 'in-progress',
                        loading: true,
                        content: 'Operation initiated',
                    },
                ],
            },
            {
                type: 'solution-api/executeMutation/fulfilled',
                meta: { arg: { endpointName: 'savePromptConfig' } },
                payload: {
                    id: 'savePromptConfig',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: 'savePromptConfigFulfilled',
                type: 'success',
                content: 'Prompt configuration saved',
            },
        ],
    });
});

test('POST performMerge Api match pending', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/pending',
            meta: { arg: { endpointName: 'performMerge' } },
            payload: {
                id: 'performMerge',
            },
        }),
    ).toEqual({
        notifications: [
            {
                id: 'performMerge',
                type: 'in-progress',
                loading: true,
                content: 'Operation initiated',
            },
        ],
    });
});

test('POST performMerge Api match rejected', () => {
    expect(
        reducer(
            {
                notifications: [
                    {
                        id: 'performMerge',
                        type: 'in-progress',
                        loading: true,
                        content: 'Operation initiated',
                    },
                ],
            },
            {
                type: 'solution-api/executeMutation/rejected',
                meta: { arg: { endpointName: 'performMerge' } },
                payload: {
                    id: 'performMerge',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: 'performMergeFailed',
                type: 'error',
                content: 'Unable to perform operation',
            },
        ],
    });
});

test('POST performMerge Api match fulfilled', () => {
    expect(
        reducer(
            {
                notifications: [
                    {
                        id: 'performMerge',
                        type: 'in-progress',
                        loading: true,
                        content: 'Operation initiated',
                    },
                ],
            },
            {
                type: 'solution-api/executeMutation/fulfilled',
                meta: { arg: { endpointName: 'performMerge' } },
                payload: {
                    id: 'performMerge',
                },
            },
        ),
    ).toEqual({
        notifications: [
            {
                id: 'performMergeFulfilled',
                type: 'success',
                content: 'Operation performed successfully',
            },
        ],
    });
});
