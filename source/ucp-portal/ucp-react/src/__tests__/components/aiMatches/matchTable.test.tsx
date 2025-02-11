// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { act, fireEvent, screen } from '@testing-library/react';
import { http } from 'msw';
import { ok } from '../../../mocks/handlers.ts';
import { ConfigResponse } from '../../../models/config.ts';
import { ROUTES } from '../../../models/constants.ts';
import { MatchPair } from '../../../models/match.ts';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { renderAppContent } from '../../test-utils.tsx';

const matchPair: MatchPair = {
    domain_sourceProfileId: 'domain_qb1a3db7827e495ehw7365e7e29da0c9',
    match_targetProfileId: 'match_pff36408ddg243ff93d60f6b46807445',
    targetProfileID: 'pff36408ddg243ff93d60f6b46807445',
    score: '0.9904272410936669',
    domain: 'test_domain',
    profileDataDiff: [
        {
            fieldName: 'FirstName',
            sourceValue: 'Carlos',
            targetValue: 'Carlos',
        },
        {
            fieldName: 'LastName',
            sourceValue: 'Sainz Jr',
            targetValue: 'Sainz',
        },
        {
            fieldName: 'BirthDate',
            sourceValue: '09-01-1994',
            targetValue: '09-01-1994',
        },
        {
            fieldName: 'Gender',
            sourceValue: '',
            targetValue: 'M',
        },
        {
            fieldName: 'PhoneNumber',
            sourceValue: '7777777777',
            targetValue: '32984238',
        },
        {
            fieldName: 'MobilePhoneNumber',
            sourceValue: '7777777771',
            targetValue: '7777777774',
        },
        {
            fieldName: 'HomePhoneNumber',
            sourceValue: '7777777772',
            targetValue: '7777777773',
        },
        {
            fieldName: 'EmailAddress',
            sourceValue: '',
            targetValue: 'cs55@example.com',
        },
        {
            fieldName: 'PersonalEmailAddress',
            sourceValue: 'cs55@example.com',
            targetValue: '',
        },
        {
            fieldName: 'BusinessEmailAddress',
            sourceValue: '',
            targetValue: '',
        },
        {
            fieldName: 'Address.City',
            sourceValue: 'Seattle',
            targetValue: 'Redmond',
        },
        {
            fieldName: 'ShippingAddress.City',
            sourceValue: 'Seattle',
            targetValue: 'Redmond',
        },
        {
            fieldName: 'MailingAddress.City',
            sourceValue: 'Seattle',
            targetValue: 'Redmond',
        },
        {
            fieldName: 'BillingAddress.City',
            sourceValue: 'Seattle',
            targetValue: 'Redmond',
        },
    ],
    runId: 'latest',
    scoreTargetId: 'match_0.9904272410936669_pff36408ddg243ff93d60f6b46807445_qb1a3db7827e495ehw7365e7e29da0c9',
    mergeInProgress: false,
};

const filledResponse: Partial<ConfigResponse> = {
    matchPairs: [matchPair],
};

describe('Tests for Match table display', () => {
    it('renders match table', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE + '?matches=true', async () => await ok(filledResponse)));

        // act
        renderAppContent({
            initialRoute: '/' + ROUTES.AI_MATCHES,
        });

        // assert
        expect(screen.getByLabelText('viewProfiles')).toBeInTheDocument();
        expect(screen.getByLabelText('aiConfidenceScore')).toBeInTheDocument();
        expect(screen.getByLabelText('firstName')).toBeInTheDocument();
        expect(screen.getByLabelText('lastName')).toBeInTheDocument();
        expect(screen.getByLabelText('birthDate')).toBeInTheDocument();
    });

    it('clicks refetch button', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROFILE + '?matches=true', async () => await ok(filledResponse)));

        // act
        renderAppContent({
            initialRoute: '/' + ROUTES.AI_MATCHES,
        });

        await act(async () => {
            fireEvent.click(screen.getByLabelText('refetchMatchesButton'));
        });

        await act(async () => {
            fireEvent.click(screen.getByLabelText('markFalseButton'));
        });

        await act(async () => {
            fireEvent.click(screen.getByLabelText('mergeButton'));
        });

        const diffButton = await screen.findByLabelText('openAttrDiffButton');
        expect(diffButton).toBeInTheDocument();

        await act(async () => {
            fireEvent.click(diffButton);
        });
    });
});
