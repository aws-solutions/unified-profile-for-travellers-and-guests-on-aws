// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { fireEvent, screen, act } from '@testing-library/react';
import { renderAppContent } from '../../test-utils.tsx';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { http } from 'msw';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { ok } from '../../../mocks/handlers.ts';

describe('Tests for Last Name Input', () => {
    it('renders last name input', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE, async () => await ok([])));

        // act
        renderAppContent({
            initialRoute: '/profile',
        });

        // assert
        expect(screen.getByLabelText('Last Name')).toBeInTheDocument();
    });

    it('updates the redux variable on value change', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE, async () => await ok([])));

        // act
        const { renderResult, store } = renderAppContent({
            initialRoute: '/profile',
            preloadedState: {
                profiles: {
                    lastName: '',
                    searchResults: { profiles: [] },
                } as any,
            },
        });

        // assert
        const lastNameInput = screen.getByLabelText('Last Name');
        const targetValue = 'Frost';

        await act(async () => {
            fireEvent.change(lastNameInput, { target: { value: targetValue } });
        });

        const profileStore = store.getState().profiles;
        expect(profileStore.lastName).toBe(targetValue);
    });
});
