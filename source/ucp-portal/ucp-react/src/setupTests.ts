// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { Amplify } from 'aws-amplify';
import nodeFetch, { Request, Response } from 'node-fetch';
import { afterAll, afterEach, beforeAll } from 'vitest';
import { MOCK_SERVER_URL, server } from './__tests__/server';

process.env.TZ = 'UTC'; // fix environment timezone for tests to UTC

// avoid a problem of RTK Query in vitest, see https://github.com/reduxjs/redux-toolkit/issues/3254
Object.assign(global, { fetch: nodeFetch, Request, Response });

beforeAll(() => {
    Amplify.configure({
        Auth: {
            region: '',
            userPoolId: '',
            userPoolWebClientId: '',
        },
        API: {
            endpoints: [
                {
                    name: 'solution-api',
                    endpoint: MOCK_SERVER_URL,
                },
            ],
        },
    });
    server.listen({ onUnhandledRequest: 'warn' });
});
afterAll(() => server.close());
afterEach(() => {
    cleanup();
    server.resetHandlers();
});
