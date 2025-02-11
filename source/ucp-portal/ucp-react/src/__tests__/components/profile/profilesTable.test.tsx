// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { screen } from '@testing-library/react';
import { renderAppContent } from '../../test-utils.tsx';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { http } from 'msw';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { ok } from '../../../mocks/handlers.ts';

describe('Tests for Profile table display', () => {
    it('renders profile table', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE, async () => await ok([])));

        // act
        renderAppContent({
            initialRoute: '/profile',
        });

        // assert
        expect(screen.getByLabelText('connectId')).toBeInTheDocument();
        expect(screen.getByLabelText('travellerId')).toBeInTheDocument();
        expect(screen.getByLabelText('firstName')).toBeInTheDocument();
        expect(screen.getByLabelText('lastName')).toBeInTheDocument();
        expect(screen.getByLabelText('emailId')).toBeInTheDocument();
    });
});
