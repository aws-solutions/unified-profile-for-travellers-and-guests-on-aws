// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { act, fireEvent, screen } from '@testing-library/react';
import { http } from 'msw';
import { ok } from '../../../mocks/handlers.ts';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { renderAppContent } from '../../test-utils.tsx';

describe('Tests for Email Input', () => {
    it('renders email input', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE, async () => await ok([])));

        // act
        renderAppContent({
            initialRoute: '/profile',
        });

        // assert
        expect(screen.getByLabelText('Email')).toBeInTheDocument();
    });

    it('updates the redux variable on value change', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE, async () => await ok([])));

        // act
        const { renderResult, store } = renderAppContent({
            initialRoute: '/profile',
            preloadedState: {
                profiles: {
                    email: '',
                    searchResults: { profiles: [] },
                } as any,
            },
        });

        // assert
        const emailIdInput = screen.getByLabelText('Email');
        const targetValue = 'emailamazoncom';

        await act(async () => {
            fireEvent.change(emailIdInput, { target: { value: targetValue } });
        });

        const profileStore = store.getState().profiles;
        expect(profileStore.email).toBe(targetValue);
    });
});
