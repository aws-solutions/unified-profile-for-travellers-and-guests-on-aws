// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { fireEvent, screen, act } from '@testing-library/react';
import { renderAppContent } from '../../test-utils.tsx';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { http } from 'msw';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { ok } from '../../../mocks/handlers.ts';

describe('Tests for Phone Input', () => {
    it('renders Phone input', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE, async () => await ok([])));

        // act
        renderAppContent({
            initialRoute: '/profile',
        });

        // assert
        expect(screen.getByLabelText('Phone')).toBeInTheDocument();
    });

    it('updates the redux variable on value change', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE, async () => await ok([])));

        // act
        const { renderResult, store } = renderAppContent({
            initialRoute: '/profile',
            preloadedState: {
                profiles: {
                    phone: '',
                    searchResults: { profiles: [] },
                } as any,
            },
        });

        // assert
        const phoneInput = screen.getByLabelText('Phone');
        const targetValue = '9876543210';

        await act(async () => {
            fireEvent.change(phoneInput, { target: { value: targetValue } });
        });

        const profileStore = store.getState().profiles;
        expect(profileStore.phone).toBe(targetValue);
    });
});
