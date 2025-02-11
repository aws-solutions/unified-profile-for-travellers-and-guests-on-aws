// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import reducer, { PermissionSystemState, PermissionSystemType } from '../../store/permissionSystemSlice';

const initialState: PermissionSystemState = {
    permissionSystem: {
        type: PermissionSystemType.Cognito,
    },
};

test('Should return initial state', () => {
    expect(reducer(undefined, { type: undefined })).toEqual(initialState);
});

test('Should handle setting permission state', () => {
    expect(
        reducer(initialState, {
            type: 'permissionSystem/setPermissionSystem',
            payload: {
                type: PermissionSystemType.Cognito,
            },
        }),
    ).toEqual({
        permissionSystem: {
            type: PermissionSystemType.Cognito,
        },
    });
});
