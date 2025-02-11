// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

export type PrivacySearchRequest = {
    connectId: string;
};

export type PrivacyPurgeRequest = {
    connectIds: string[];
};

export type PrivacySearchResponse = {
    domainName: string;
    connectId: string;
    status: string;
    errorMessage: string;
    searchDate: Date;
    locationResults: { [key: string]: string[] };
    totalResultsFound: number;
};

export type PrivacySearchesTableItem = {
    connectId: string;
    status: string;
    searchDate: Date;
    totalResultsFound: number;
    errorMessage: string;
};

export type PrivacyProfileLocationsTableItem = {
    source: string;
    path: string;
};

export type PrivacyPurgeStatus = {
    isPurgeRunning: boolean;
};

export const PrivacySearchStatus = {
    PRIVACY_STATUS_SEARCH_INVOKED: 'search_invoked',
    PRIVACY_STATUS_SEARCH_RUNNING: 'search_running',
    PRIVACY_STATUS_SEARCH_SUCCESS: 'search_success',
    PRIVACY_STATUS_SEARCH_FAILED: 'search_failed',
    PRIVACY_STATUS_PURGE_INVOKED: 'purge_invoked',
    PRIVACY_STATUS_PURGE_RUNNING: 'purge_running',
    PRIVACY_STATUS_PURGE_SUCCESS: 'purge_success',
    PRIVACY_STATUS_PURGE_FAILED: 'purge_failed',
} as const;

export const SearchContainerTestId = 'privacy-searches-search-container';
export const SearchResultsTestId = 'profile-searches-table';
export const SearchResultsContainerTestId = 'privacy-searches-table-container';
export const PrivacyResultsLocationsTableTestId = 'privacy-results-locations-table';
