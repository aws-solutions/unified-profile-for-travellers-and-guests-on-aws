// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

declare namespace Cypress {
    interface Chainable<Subject = any> {
        /**
         * Custom command to login
         * @example cy.login()
         */
        login(): Chainable<void>;
        loginWithPasswordChange(): Chainable<void>;
        setTokens(): Chainable<void>;
    }
}
