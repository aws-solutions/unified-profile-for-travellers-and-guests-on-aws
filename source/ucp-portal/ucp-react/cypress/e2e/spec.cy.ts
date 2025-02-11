// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import createWrapper from '@cloudscape-design/components/test-utils/selectors';

it('First Login', () => {
    if (Cypress.env('runPasswordChangeLogin')) {
        cy.loginWithPasswordChange();
    }
});

describe('React UI End To End Testing', () => {
    const createdDomainNames = [];

    before(() => {
        cy.setTokens();
    });

    after(() => {
        cy.clearLocalStorageSnapshot();
        cy.clearLocalStorage();
    });

    beforeEach(() => {
        cy.restoreLocalStorage();
        cy.visit(Cypress.env('cloudfrontDomainName'));
    });

    afterEach(() => {
        cy.saveLocalStorage();
        for (const domainName of createdDomainNames) {
            cy.window().then(win => {
                const idToken = win.localStorage.getItem('idToken');
                const requestObject = {
                    url: Cypress.env('ucpApiUrl') + '/api/ucp/admin/' + domainName,
                    method: 'DELETE',
                    headers: {
                        Authorization: 'Bearer ' + idToken,
                        'customer-profiles-domain': domainName,
                    },
                };
                cy.request(requestObject);
            });
        }
    });

    it('Login and Logout', () => {
        cy.get('[data-testid="sideNav"]').should('exist');
        cy.get('[data-testid="sideNav"] a[href="/logout"]').click();
    });

    it('Creates and Deletes a Domain', () => {
        const baseUrl = Cypress.env('ucpApiUrl');
        const topNav = createWrapper().findTopNavigation();
        const sideNav = createWrapper().findSideNavigation();
        cy.get(sideNav.toSelector()).should('exist').should('be.visible');
        cy.get(topNav.toSelector()).should('exist');

        console.log('Going To Domain Create Page');
        cy.get(topNav.findUtilities().find('[aria-label="createDomainTopNav"]').toSelector()).click();

        console.log('Should See Domain Create Page');
        const createDomainForm = createWrapper().findForm('[data-testid="domainCreateForm"]');
        cy.get(createDomainForm.toSelector()).should('exist');

        console.log('Inputting Test Domain Name');
        const testDomainName = 'cypress_e2e_' + (Math.random() + 1).toString(36).substring(7);
        cy.get(createDomainForm.findContent().findInput('[data-testid="domainCreateInput"]').toSelector()).type(testDomainName);

        // Click Domain Create Button
        console.log('Create Domain');
        cy.intercept('POST', baseUrl + '/api/ucp/admin').as('createDomainRequest');
        cy.get(createDomainForm.findActions().findButton().toSelector()).last().contains('Create').click();
        cy.wait('@createDomainRequest').then(({ request, response }) => {
            cy.wrap(request).its('body').should('exist');
            cy.wrap(request).its('body.domain.customerProfileDomain').should('eq', testDomainName);
            cy.wrap(response).its('body').should('exist');
            cy.wrap(response).its('statusCode').should('eq', 200);
            cy.wrap(response).its('body.asyncEvent').should('exist');
            cy.wrap(response).its('body.asyncEvent.item_id').should('eq', 'createDomain');
            cy.wrap(response).its('body.asyncEvent.item_type').should('exist');
            cy.wrap(response).its('body.asyncEvent.status').should('eq', 'invoked');
        });

        // Wait for Async
        console.log('Waiting for async');
        const notifications = createWrapper().findFlashbar();
        cy.get(notifications.findItems().toSelector()).should('be.visible').contains('Creating Domain');
        cy.get(notifications.findItems().toSelector()).should('be.visible').contains('Successfully Created Domain', { timeout: 60000 });
        cy.get(notifications.find('button').toSelector()).click();
        createdDomainNames.push(testDomainName);

        // Should See Domain Settings Page
        console.log('Should See Domain Settings Page with new domain selected');
        cy.get('[data-testid="settingsPageHeader"]').contains(testDomainName, { timeout: 20000 });
        cy.get(sideNav.findItemByIndex(4).findSectionGroupTitle().toSelector()).should('be.visible').contains(testDomainName);
        cy.get(topNav.findUtilities().findButtonDropdown().toSelector())
            .find('button[aria-label="domainSelector"]')
            .first()
            .contains(testDomainName);

        // Delete Domain Button Click
        console.log('Delete Domain');
        const settingsPage = createWrapper().findContentLayout('[data-testid="settingsContent"]');
        cy.get(settingsPage.findHeader().findButton('[data-testid="deleteDomainButton"]').toSelector(), { timeout: 10000 })
            .should('exist')
            .should('not.be.disabled')
            .click();

        // Should see Modal
        const deleteDomainModal = createWrapper().findModal('[data-testid="deleteDomainModal"]');
        cy.get(deleteDomainModal.toSelector())
            .should('be.visible')
            .contains('Selected Domain: ' + testDomainName);

        // Confirm Delete
        cy.intercept('DELETE', baseUrl + '/api/ucp/admin/' + testDomainName).as('deleteDomainRequest');
        cy.get(deleteDomainModal.findFooter().findButton('[data-testid="confirmDeleteDomainButton"]').toSelector()).should('exist').click();
        cy.wait('@deleteDomainRequest').then(({ request, response }) => {
            cy.wrap(request).its('headers').should('exist');
            cy.wrap(request).its('headers.customer-profiles-domain').should('eq', testDomainName);
            cy.wrap(response).its('body').should('exist');
            cy.wrap(response).its('statusCode').should('eq', 200);
            cy.wrap(response).its('body.asyncEvent').should('exist');
            cy.wrap(response).its('body.asyncEvent.item_id').should('eq', 'deleteDomain');
            cy.wrap(response).its('body.asyncEvent.item_type').should('exist');
            cy.wrap(response).its('body.asyncEvent.status').should('eq', 'invoked');
        });

        // Wait for Async
        console.log('Waiting for async');
        cy.get(notifications.findItems().toSelector()).should('be.visible').contains('Deleting Domain');
        cy.get(notifications.findItems().toSelector()).should('be.visible').contains('Successfully Deleted Domain', { timeout: 60000 });
        cy.get(notifications.find('button').toSelector()).click();
        createdDomainNames.pop();

        // // Should see Welcome Page
        cy.get('[data-id="Solution-header"]').contains('Unified profiles for travelers and guests on AWS');
        cy.get(sideNav.findItemByIndex(4).findSectionGroupTitle().toSelector()).should('not.exist');
        cy.get(topNav.findUtilities().findButtonDropdown().toSelector())
            .find('button[aria-label="domainSelector"]')
            .first()
            .contains('Select a domain to get started');
    });
});
