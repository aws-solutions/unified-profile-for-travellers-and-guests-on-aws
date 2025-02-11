// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { act, fireEvent, screen } from '@testing-library/react';
import { ok } from 'assert';
import { http } from 'msw';
import { ConfigResponse } from '../../../models/config.ts';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { renderAppContent } from '../../test-utils.tsx';

const partialResponse: Partial<ConfigResponse> = {
    portalConfig: {
        hyperlinkMappings: [
            {
                accpObject: 'accpObject1',
                fieldName: 'fieldName1',
                hyperlinkTemplate: 'https://www.example.com/',
            },
        ],
    },
};

it('opens portal config page and adds a config', async () => {
    server.use(http.get(MOCK_SERVER_URL + ApiEndpoints.PORTAL_CONFIG, async () => await ok(partialResponse)));

    await act(async () => {
        renderAppContent({
            initialRoute: '/config',
            preloadedState: {
                user: {
                    appAccessPermission: 2 ** 7,
                },
                config: {
                    accpRecords: [
                        {
                            name: 'testObject',
                            objectFields: ['one', 'two'],
                        },
                    ],
                    dataFieldInput: {
                        label: 'one',
                        value: 'one',
                    },
                    dataRecordInput: {
                        label: 'testObject',
                        value: 'testObject',
                    },
                    hyperlinkMappings: [
                        {
                            accpObject: 'accpObject1',
                            fieldName: 'fieldName1',
                            hyperlinkTemplate: 'https://www.example.com/',
                        },
                    ],
                    urlTemplateInput: 'https://www.example.com/',
                },
            },
        });
    });

    const addConfigButton = screen.getByTestId('addConfigButton');
    expect(addConfigButton).toBeInTheDocument();
    expect(addConfigButton).toBeEnabled();

    await act(async () => {
        fireEvent.click(addConfigButton);
    });
});
