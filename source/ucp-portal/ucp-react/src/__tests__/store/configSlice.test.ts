// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { configApiSlice } from '../../store/configApiSlice';
import reducer, { setDataRecordInput, setDataFieldInput, setUrlTemplateInput, clearFormFields } from '../../store/configSlice';
import { expect, test } from 'vitest';

const initialState = {
    dataRecordInput: null,
    dataFieldInput: null,
    urlTemplateInput: '',
    hyperlinkMappings: [],
    accpRecords: [],
};

const testHyperlinkMapping = {
    accpObject: 'test_accpObject',
    fieldName: 'test_fieldName',
    hyperlinkTemplate: 'testHyperlinkTemplate',
};

const testAccpRecord = {
    name: 'test_accpRecord',
    struct: { test_key: 'test_value' },
};

test('Should Return Initial State', () => {
    expect(reducer(undefined, { type: undefined })).toEqual(initialState);
});

test('Should handle data record input value change', () => {
    expect(reducer(initialState, setDataRecordInput({ value: 'testValue', label: 'testLabel' }))).toEqual({
        ...initialState,
        dataRecordInput: { value: 'testValue', label: 'testLabel' },
    });
});

test('Should handle data field input value change', () => {
    expect(reducer(initialState, setDataFieldInput({ value: 'testValue', label: 'testLabel' }))).toEqual({
        ...initialState,
        dataFieldInput: { value: 'testValue', label: 'testLabel' },
    });
});

test('Should handle url template value change', () => {
    expect(reducer(initialState, setUrlTemplateInput('testUrlTemplate'))).toEqual({
        ...initialState,
        urlTemplateInput: 'testUrlTemplate',
    });
});

test('Should clear all form data', () => {
    expect(
        reducer(
            {
                ...initialState,
                dataRecordInput: { value: 'testValue1', label: 'testLabel1' },
                dataFieldInput: { value: 'testValue2', label: 'testLabel2' },
                urlTemplateInput: 'testUrlTemplate',
            },
            clearFormFields(),
        ),
    ).toEqual(initialState);
});

test('GET portalConfig Api match fulfilled', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeQuery/fulfilled',
            meta: { arg: { endpointName: 'getPortalConfig' } },
            payload: {
                portalConfig: {
                    hyperlinkMappings: [testHyperlinkMapping],
                },
                accpRecords: [testAccpRecord],
            },
        }),
    ).toEqual({ ...initialState, hyperlinkMappings: [testHyperlinkMapping], accpRecords: [testAccpRecord] });
});

test('POST portalConfig Api match fulfilled', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeMutation/fulfilled',
            meta: { arg: { endpointName: 'postPortalConfig' } },
            payload: {
                portalConfig: {
                    hyperlinkMappings: [testHyperlinkMapping],
                },
            },
        }),
    ).toEqual({ ...initialState, hyperlinkMappings: [testHyperlinkMapping] });
});
