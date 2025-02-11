// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { act, screen, within } from '@testing-library/react';
import { renderAppContent } from '../../test-utils.tsx';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { http } from 'msw';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { ok } from '../../../mocks/handlers.ts';
import { describe } from 'node:test';

describe('Tests for Profiles page', () => {
    it('renders the profile page', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE, async () => await ok([])));

        // act
        await act(async () => {
            renderAppContent({
                initialRoute: '/profile',
            });
        });

        // assert
        const withinMain = within(screen.getByTestId('main-content'));
        expect(withinMain.getByText('Search Traveller')).toBeInTheDocument();
    });
});
