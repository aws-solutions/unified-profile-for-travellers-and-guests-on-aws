// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Amplify, Auth } from 'aws-amplify';
import 'cypress-localstorage-commands';

const username = Cypress.env('email');
const password = Cypress.env('pwd');
const userPoolId = Cypress.env('cognitoUserPoolId');
const clientId = Cypress.env('cognitoClientId');

Amplify.configure({
    Auth: {
        userPoolId: userPoolId,
        userPoolWebClientId: clientId,
    },
});

Cypress.Commands.add('setTokens', () => {
    cy.then(() => Auth.signIn(username, password)).then(cognitoUser => {
        console.log(cognitoUser);
        const idToken = cognitoUser.signInUserSession.idToken.jwtToken;
        const accessToken = cognitoUser.signInUserSession.accessToken.jwtToken;

        cy.setLocalStorage('accessToken', accessToken);
        cy.setLocalStorage('idToken', idToken);
    });
    cy.saveLocalStorage();
});

Cypress.Commands.add('login', () => {
    console.log('Go to page');
    cy.visit(Cypress.env('cloudfrontDomainName'));

    console.log('Enter Credentials');
    cy.get('input[name=username]').should('exist').type(username);
    cy.get('input[name=password]').should('exist').type(password, { log: false });
    cy.get('button').contains('Sign in').should('exist').click();

    console.log('See Solution Landing Page');
    cy.get('[data-testid="sideNav"]').should('exist');
    cy.get('[data-testid="sideNav"] a[href="/logout"]').should('exist');
});

Cypress.Commands.add('loginWithPasswordChange', () => {
    console.log('Go to page');
    cy.visit(Cypress.env('cloudfrontDomainName'));

    console.log('Enter Credentials');
    cy.get('input[name=username]').should('exist').type(username);
    cy.get('input[name=password]').should('exist').type(password);
    cy.get('button').contains('Sign in').should('exist').click();

    console.log('See Password Update Screen');
    cy.get('button').contains('Change Password').should('exist');
    cy.get('input[name=password]').should('exist').type(password);
    cy.get('input[name=confirm_password]').should('exist').type(password);
    cy.get('button').contains('Change Password').click();

    console.log('See Solution Landing Page');
    cy.get('[data-testid="sideNav"]').should('exist');
    cy.get('[data-testid="sideNav"] a[href="/logout"]').click();
});

// The react e2e test fails intermittently on an exception thrown by the ResizeObserver after the test for creating and deleting a domain
// It could occur if a resize event triggers an update cycle that doesn’t complete within the same frame. 
// The ResizeObserver, which monitors changes to an element’s size, could then fall into an infinite loop.
// This might be a sideffect of Cypress because it clicks things too fast.
// Manually creating and deleting a domain from UI did not throw the exception

// https://stackoverflow.com/a/50387233
// https://github.com/quasarframework/quasar/issues/2233#issuecomment-414070235

// Suppressing the error for cypress 
Cypress.on('uncaught:exception', err => !err.message.includes('ResizeObserver loop completed with undelivered notifications'));
