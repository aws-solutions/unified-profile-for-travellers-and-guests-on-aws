// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Domain } from '../../models/config';
import reducer, { setSelectedDomain } from '../../store/domainSlice';
import { expect, test } from 'vitest';

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

const newDomain: Domain = {
    customerProfileDomain: 'test_domain',
    numberOfObjects: 10,
    numberOfProfiles: 30,
    matchingEnabled: false,
    cacheMode: 0,
    created: '',
    lastUpdated: '',
    mappings: null,
    integrations: null,
    needsMappingUpdate: false,
};

const initialState = {
    selectedDomain: initDomain,
    allDomains: [],
    promptConfig: { isActive: false, value: '' },
    matchConfig: { isActive: false, value: '' },
};

test('Should Return Initial State', () => {
    expect(reducer(undefined, { type: undefined })).toEqual(initialState);
});

test('Should handle selecting a domain', () => {
    expect(
        reducer(
            {
                selectedDomain: initDomain,
                allDomains: [newDomain],
                promptConfig: { isActive: false, value: '' },
                matchConfig: { isActive: false, value: '' },
            },
            setSelectedDomain(newDomain),
        ),
    ).toEqual({
        selectedDomain: newDomain,
        allDomains: [newDomain],
        promptConfig: { isActive: false, value: '' },
        matchConfig: { isActive: false, value: '' },
    });
});

test('GET domain Api match fulfilled', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeQuery/fulfilled',
            meta: { arg: { endpointName: 'getDomain' } },
            payload: {
                config: {
                    domains: [newDomain],
                },
            },
        }),
    ).toEqual({ ...initialState, selectedDomain: newDomain });
});

test('GET prompt config match fulfilled', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeQuery/fulfilled',
            meta: { arg: { endpointName: 'getPromptConfig' } },
            payload: {
                domainSetting: {
                    promptConfig: { isActive: false, value: '' },
                    matchConfig: { isActive: false, value: '' },
                },
            },
        }),
    ).toEqual({ ...initialState, matchConfig: { isActive: false, value: '0' }, promptConfig: initialState.promptConfig });
});

test('GET all domain Api match fulfilled', () => {
    expect(
        reducer(initialState, {
            type: 'solution-api/executeQuery/fulfilled',
            meta: { arg: { endpointName: 'getAllDomains' } },
            payload: {
                config: {
                    domains: [newDomain, initDomain],
                },
            },
        }),
    ).toEqual({ ...initialState, allDomains: [newDomain, initDomain] });
});

test('GET async status Api match fulfilled (successful async case)', () => {
    expect(
        reducer(
            {
                selectedDomain: newDomain,
                allDomains: [],
                promptConfig: { isActive: false, value: '' },
                matchConfig: { isActive: false, value: '' },
            },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getAsyncStatus' } },
                payload: {
                    item_id: 'deleteDomain',
                    status: 'success',
                },
            },
        ),
    ).toEqual({ ...initialState });
});

test('GET async status Api match fulfilled (running async case)', () => {
    expect(
        reducer(
            {
                selectedDomain: newDomain,
                allDomains: [],
                promptConfig: { isActive: false, value: '' },
                matchConfig: { isActive: false, value: '' },
            },
            {
                type: 'solution-api/executeQuery/fulfilled',
                meta: { arg: { endpointName: 'getAsyncStatus' } },
                payload: {
                    item_id: 'deleteDomain',
                    status: 'running',
                },
            },
        ),
    ).toEqual({ ...initialState, selectedDomain: newDomain });
});
