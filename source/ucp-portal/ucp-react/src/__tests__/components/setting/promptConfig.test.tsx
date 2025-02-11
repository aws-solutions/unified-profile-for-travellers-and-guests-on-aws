// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { fireEvent, screen, act } from '@testing-library/react';
import { renderAppContent } from '../../test-utils.tsx';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { http } from 'msw';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { ok } from '../../../mocks/handlers.ts';
import { ConfigResponse } from '../../../models/config.ts';

const summaryGenActive: Partial<ConfigResponse> = {
    domainSetting: {
        promptConfig: {
            isActive: true,
            value: 'dummy prompt',
        },
        matchConfig: {
            isActive: true,
            value: '0.8',
        },
    },
};

const newDomain = {
    customerProfileDomain: 'test_domain',
    numberOfObjects: 10,
    numberOfProfiles: 30,
    matchingEnabled: false,
    cacheMode: 0,
    created: '',
    lastUpdated: '',
    mappings: null,
    integrations: null,
    needsMappingUpdate: false,
};

describe('Tests for Prompt configutation', () => {
    it('renders PromptConfig section', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROMPT_CONFIG, async () => await ok(summaryGenActive)));

        // act
        renderAppContent({
            initialRoute: '/settings',
            preloadedState: {
                domains: {
                    selectedDomain: newDomain,
                    allDomains: [],
                    promptConfig: { isActive: true, value: '' },
                    matchConfig: { isActive: false, value: '' },
                } as any,
            },
        });

        // assert
        expect(screen.getByLabelText('summaryGenToggle')).toBeInTheDocument();
        expect(screen.getByLabelText('summaryGenTextbox')).toBeInTheDocument();
    });

    it('clicks summary generation toggle', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROMPT_CONFIG, async () => await ok(summaryGenActive)));

        // act
        renderAppContent({
            initialRoute: '/settings',
            preloadedState: {
                domains: {
                    selectedDomain: newDomain,
                    allDomains: [],
                    promptConfig: { isActive: true, value: '' },
                    matchConfig: { isActive: false, value: '' },
                } as any,
                user: {
                    appAccessPermission: 2 ** 6,
                },
            },
        });

        // assert
        const summaryGenToggle = await screen.getByLabelText('summaryGenToggle');

        await act(async () => {
            fireEvent.click(summaryGenToggle);
        });

        expect(summaryGenToggle).not.toBeChecked();
    });

    it('changes prompt text value', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROMPT_CONFIG, async () => await ok(summaryGenActive)));

        // act
        renderAppContent({
            initialRoute: '/settings',
            preloadedState: {
                domains: {
                    selectedDomain: newDomain,
                    allDomains: [],
                    promptConfig: { isActive: true, value: '' },
                    matchConfig: { isActive: false, value: '' },
                } as any,
            },
        });

        // assert
        const summaryGenTextbox = await screen.getByLabelText('summaryGenTextbox');
        const textValue = 'textValue';
        await act(async () => {
            fireEvent.change(summaryGenTextbox, { target: { value: textValue } });
        });
        expect(summaryGenTextbox).toHaveValue(textValue);
    });

    it('saves prompt config', async () => {
        // arrange
        server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PROMPT_CONFIG, async () => await ok(summaryGenActive)));

        // act
        renderAppContent({
            initialRoute: '/settings',
            preloadedState: {
                domains: {
                    selectedDomain: newDomain,
                    allDomains: [],
                    promptConfig: { isActive: true, value: '' },
                    matchConfig: { isActive: false, value: '' },
                } as any,
                user: {
                    appAccessPermission: 2 ** 6,
                },
            },
        });

        const summaryGenTextbox = await screen.getByLabelText('summaryGenTextbox');
        const textValue = 'textValue';
        await act(async () => {
            fireEvent.change(summaryGenTextbox, { target: { value: textValue } });
        });

        // assert
        const savePromptConfig = await screen.getByLabelText('savePromptConfig');

        await act(async () => {
            fireEvent.click(savePromptConfig);
        });
    });
});
