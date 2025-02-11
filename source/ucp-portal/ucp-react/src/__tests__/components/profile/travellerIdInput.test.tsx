// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { fireEvent, screen, act } from '@testing-library/react';
import { renderAppContent } from '../../test-utils.tsx';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { http } from 'msw';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { ok } from '../../../mocks/handlers.ts';

describe('Tests for TravellerId Input', () => {
    it('renders TravellerId input', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE, async () => await ok([])));

        // act
        renderAppContent({
            initialRoute: '/profile',
        });

        // assert
        expect(screen.getByLabelText('Traveller Id')).toBeInTheDocument();
    });

    it('updates the redux variable on value change', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE, async () => await ok([])));

        // act
        const { renderResult, store } = renderAppContent({
            initialRoute: '/profile',
            preloadedState: {
                profiles: {
                    travellerId: '',
                    searchResults: { profiles: [] },
                } as any,
            },
        });

        // assert
        const travellerIdInput = screen.getByLabelText('Traveller Id');
        const targetValue = 'TravellerIdOne';

        await act(async () => {
            fireEvent.change(travellerIdInput, { target: { value: targetValue } });
        });

        const profileStore = store.getState().profiles;
        expect(profileStore.travellerId).toBe(targetValue);
    });
});
