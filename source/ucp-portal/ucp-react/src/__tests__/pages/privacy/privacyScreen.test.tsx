// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import createWrapper from '@cloudscape-design/components/test-utils/dom';
import { act, screen, waitFor } from '@testing-library/react';
import { http } from 'msw';
import { ok } from '../../../mocks/handlers';
import { ConfigResponse } from '../../../models/config';
import { ROUTES } from '../../../models/constants';
import { SearchContainerTestId, SearchResultsContainerTestId } from '../../../models/privacy';
import { ApiEndpoints } from '../../../store/solutionApi';
import { MOCK_SERVER_URL, server } from '../../server';
import { renderAppContent } from '../../test-utils';

const initialRoute: string = `/${ROUTES.PRIVACY}`;

const emptyResponse: Partial<ConfigResponse> = {
    privacySearchResults: [],
};

const filledResponse: Partial<ConfigResponse> = {
    privacySearchResults: [
        {
            domainName: 'test',
            connectId: 'testConnectId',
            status: 'success',
            searchDate: new Date(),
            locationResults: {},
            errorMessage: '',
            totalResultsFound: 0,
        },
        {
            domainName: 'test',
            connectId: 'testConnectId2',
            status: 'running',
            searchDate: new Date(),
            locationResults: {},
            errorMessage: '',
            totalResultsFound: 0,
        },
    ],
};

describe('Tests for Privacy Screen page', () => {
    it('Renders all sections', async () => {
        //  arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PRIVACY, async () => await ok(filledResponse)));

        //  act
        await act(async () => {
            renderAppContent({ initialRoute });
        });

        //  assert
        expect(screen.getByText('Search for Profile Data')).toBeInTheDocument();
        expect(screen.getByText('Create Profile Search')).toBeInTheDocument();
        const searchResultsContainer = await screen.findByTestId(SearchResultsContainerTestId);
        expect(searchResultsContainer).toBeInTheDocument();
        const element = createWrapper(searchResultsContainer).findTable();
        expect(element).toBeTruthy();

        waitFor(
            () => {
                const rows = element?.findRows();
                expect(rows?.length).toBe(filledResponse.privacySearchResults?.length);
            },
            {
                timeout: 8000,
                interval: 1000,
            },
        );
    });

    it('Updates Search Text', async () => {
        //  arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PRIVACY, async () => await ok(emptyResponse)));

        //  act
        await act(async () => {
            renderAppContent({ initialRoute });
        });

        //  assert
        const searchContainer = await screen.findByTestId(SearchContainerTestId);
        expect(searchContainer).toBeInTheDocument();
        const element = createWrapper(searchContainer);
        const inputField = element.findFormField()?.findControl()?.findInput();
        expect(inputField).toBeTruthy();
        await act(async () => {
            inputField?.setInputValue('connect_id_to_search_for');
        });
        expect(inputField?.getInputValue()).toBe('connect_id_to_search_for');
    });

    it('Searches for a profile', async () => {
        //  arrange
        server.use(
            http.get(MOCK_SERVER_URL + ApiEndpoints.PRIVACY, async () => await ok(emptyResponse), {
                once: true,
            }),

            http.get(MOCK_SERVER_URL + ApiEndpoints.PRIVACY, async () => await ok(filledResponse)),
        );

        //  act
        await act(async () => {
            renderAppContent({ initialRoute });
        });

        //  assert
        const resultsContainer = await screen.findByTestId(SearchResultsContainerTestId);
        const resultsTable = createWrapper(resultsContainer).findTable();
        expect(resultsTable).toBeTruthy();
        const rows = resultsTable?.findRows();
        expect(rows?.length).toBe(0);

        const searchContainer = await screen.findByTestId(SearchContainerTestId);
        const searchContainerWrapper = createWrapper(searchContainer);
        const searchInput = searchContainerWrapper.findFormField()?.findControl()?.findInput();
        expect(searchInput).toBeTruthy();
        searchInput?.setInputValue('connect_id_to_search_for');

        const searchButton = searchContainerWrapper.findFormField()?.findSecondaryControl()?.findButton();
        expect(searchButton).toBeTruthy();
        await act(async () => {
            searchButton?.click();
        });

        waitFor(
            () => {
                const updatedRows = resultsTable?.findRows();
                expect(updatedRows?.length).toBe(0);
            },
            {
                interval: 1000,
                timeout: 5000,
            },
        );
    });
});
