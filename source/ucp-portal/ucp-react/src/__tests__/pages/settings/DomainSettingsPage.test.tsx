// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { act, fireEvent, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderAppContent } from '../../test-utils.tsx';
import { ConfigResponse } from '../../../models/config.ts';
import { AsyncEventType } from '../../../models/async.ts';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { http } from 'msw';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { ok } from 'assert';

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

const deleteDomainResponse: Partial<ConfigResponse> = {
    asyncEvent: {
        item_id: '',
        item_type: AsyncEventType.DELETE_DOMAIN,
        lastUpdated: new Date(),
        progress: 0,
        status: '',
    },
};

it('opens welcome page', async () => {
    await act(async () => {
        renderAppContent({
            initialRoute: '/',
        });
    });
});

it('opens settings page and deletes domain', async () => {
    server.use(
        http.delete(
            MOCK_SERVER_URL + ApiEndpoints.ADMIN + `/${newDomain.customerProfileDomain}`,
            async () => await ok(deleteDomainResponse),
        ),
    );

    await act(async () => {
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
                    appAccessPermission: 2 ** 5,
                },
            },
        });
    });

    // assert
    expect(screen.getByTestId('deleteDomainButton')).toBeInTheDocument();
    const deleteButton = await screen.getByTestId('deleteDomainButton');

    await act(async () => {
        fireEvent.click(deleteButton);
    });

    expect(screen.getByTestId('confirmDeleteDomainButton')).toBeInTheDocument();
    const confirmDeleteButton = await screen.getByTestId('confirmDeleteDomainButton');

    await act(async () => {
        fireEvent.click(confirmDeleteButton);
    });
});

// SHOULD BE ADDED TO E2E TESTS
// it('opens settings page and selects a domain', async () => {
//   // WHEN rendering the /settings route
//   renderAppContent({
//     initialRoute: '/settings',
//   });

//   //wait for the store to be set up
//   await new Promise((r) => setTimeout(r, 5000));

//   // THEN
//   const withinMain = within(screen.getByTestId('topnav'));
//   const topNavButton = withinMain.getByRole('button', {name: "Select a domain to get started"})
//   await userEvent.click(topNavButton)
//   const domainMenu = withinMain.getByRole('menu', {name: "Select a domain to get started"})
//   expect(domainMenu).toBeVisible();

//   const withinSelector = within(domainMenu)
//   const singleDomainButton = withinSelector.getByText("initial_domain")
//   expect(singleDomainButton).toBeVisible();
//   await userEvent.click(singleDomainButton)
//   expect(singleDomainButton).not.toBeVisible()
//   expect(topNavButton).toHaveTextContent("initial_domain")

//   const withinSettings = within(screen.getByTestId('settings'));

//   await new Promise((r) => setTimeout(r, 8000));

//   const domainName = withinSettings.getByRole("heading", {name: "Domain Name"})
//   expect(domainName).toBeVisible()

//   expect(await withinSettings.findByText("initial_domain")).toBeInTheDocument();
// });
