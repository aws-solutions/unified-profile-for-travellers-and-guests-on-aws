// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import createWrapper from '@cloudscape-design/components/test-utils/dom';
import { act, screen, waitFor } from '@testing-library/react';
import { http } from 'msw';
import { ok } from '../../../mocks/handlers';
import { ConfigResponse } from '../../../models/config';
import { ROUTES } from '../../../models/constants';
import { PrivacyResultsLocationsTableTestId } from '../../../models/privacy';
import { ApiEndpoints } from '../../../store/solutionApi';
import { MOCK_SERVER_URL, server } from '../../server';
import { renderAppContent } from '../../test-utils';

const initialRoute: string = `/${ROUTES.PRIVACY_RESULTS}`;
const connectId: string = '12345';

const emptyResponse: Partial<ConfigResponse> = {
    privacySearchResult: {
        domainName: 'test',
        connectId,
        status: 'success',
        searchDate: new Date(),
        locationResults: {},
        errorMessage: '',
        totalResultsFound: 0,
    },
};

const filledResponse: Partial<ConfigResponse> = {
    privacySearchResult: {
        domainName: 'test',
        connectId,
        status: 'success',
        searchDate: new Date(),
        locationResults: {
            S3: ['s3://bucket/file.parquet'],
            'Customer Profile': ['profileId: 54321'],
            'Entity Resolution Matches': ['ddb://table-name/pk'],
        },
        errorMessage: '',
        totalResultsFound: 1,
    },
};

describe('Tests for Privacy Results Screen page', () => {
    it('Renders Privacy Locations results', async () => {
        //  arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PRIVACY + '/:connectId', async () => await ok(filledResponse)));

        //  act
        await act(async () => {
            renderAppContent({ initialRoute });
        });

        //  assert
        const locationsTable = await screen.findByTestId(PrivacyResultsLocationsTableTestId);
        expect(locationsTable).toBeInTheDocument();

        const locationsTableWrapper = createWrapper().findTable();
        expect(locationsTableWrapper).toBeTruthy();

        waitFor(
            () => {
                const rows = locationsTableWrapper?.findRows();
                expect(rows?.length).toBe(3);
            },
            {
                timeout: 8000,
                interval: 1000,
            },
        );
    });
});
