// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { expect, test } from 'vitest';
import { PrivacyProfileLocationsTableItem, PrivacySearchResponse, PrivacySearchesTableItem } from '../../models/privacy';
import reducer, { PrivacySliceState, clearSearchResult } from '../../store/privacySlice';

const InitialState: PrivacySliceState = {
    searchResult: [],
    searchResults: [],
};

test('Should Return Initial State', () => {
    expect(reducer(undefined, { type: undefined })).toEqual(InitialState);
});

test('Should handle clearing search results', () => {
    expect(
        reducer(
            {
                searchResult: [
                    {
                        path: 'S3path',
                        source: 'S3source',
                    },
                ],
                searchResults: [],
            },
            clearSearchResult(),
        ),
    ).toEqual(InitialState);
});

test('Should Handle listSearches matchFulfilled', () => {
    const expectedSearchResults: PrivacySearchesTableItem = {
        connectId: 'a',
        searchDate: new Date(),
        status: 'success',
        errorMessage: '',
        totalResultsFound: 0,
    };
    const searchResponse: PrivacySearchResponse[] = [
        {
            ...expectedSearchResults,
            domainName: 'a',
            locationResults: {},
        },
    ];
    expect(
        reducer(InitialState, {
            type: 'solution-api/executeQuery/fulfilled',
            meta: {
                arg: {
                    endpointName: 'listSearches',
                },
            },
            payload: searchResponse,
        }),
    ).toEqual({
        ...InitialState,
        searchResults: [expectedSearchResults],
    });
});

test('Should handle empty getProfileSearchResult matchFulfilled', () => {
    const searchResponse: PrivacySearchResponse = {
        connectId: 'a',
        domainName: 'a',
        searchDate: new Date(),
        status: 'success',
        locationResults: {},
        totalResultsFound: 0,
        errorMessage: '',
    };

    expect(
        reducer(InitialState, {
            type: 'solution-api/executeQuery/fulfilled',
            meta: { arg: { endpointName: 'getProfileSearchResult' } },
            payload: searchResponse,
        }),
    ).toEqual({ ...InitialState, searchResult: [] });
});

test('Should handle getProfileSearchResult matchFulfilled', () => {
    const s3SearchResults: PrivacyProfileLocationsTableItem = {
        path: 's3path',
        source: 'S3',
    };
    const profileSearchResults: PrivacyProfileLocationsTableItem = {
        path: 'accpPath',
        source: 'Customer Profile',
    };
    const matchSearchResults: PrivacyProfileLocationsTableItem = {
        path: 'ddbPath',
        source: 'Entity Resolution Matches',
    };
    const expectedSearchResults: PrivacyProfileLocationsTableItem[] = [s3SearchResults, profileSearchResults, matchSearchResults];

    const searchResponse: PrivacySearchResponse = {
        connectId: 'a',
        domainName: 'a',
        searchDate: new Date(),
        status: 'success',
        locationResults: {
            S3: [s3SearchResults.path],
            'Customer Profile': [profileSearchResults.path],
            'Entity Resolution Matches': [matchSearchResults.path],
        },
        totalResultsFound: 3,
        errorMessage: '',
    };

    expect(
        reducer(InitialState, {
            type: 'solution-api/executeQuery/fulfilled',
            meta: {
                arg: {
                    endpointName: 'getProfileSearchResult',
                },
            },
            payload: searchResponse,
        }),
    ).toEqual({
        ...InitialState,
        searchResult: expectedSearchResults,
    });
});
