// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { delay, http, HttpResponse } from 'msw';
import { returnConfig } from '../__tests__/utils/test-data-factory.ts';
import { ApiEndpoints } from '../store/solutionApi.ts';

/**
 * Return a 200 OK http response with the given payload.
 * Delays the response by 200ms to simulate realistic latency and allow
 * to test a loading spinner etc on the UI.
 */
export const ok = async (payload: object | object[], delayMilliseconds: number = 200) => {
    await delay(delayMilliseconds);
    return HttpResponse.json(payload, {
        status: 200,
        headers: [['Access-Control-Allow-Origin', '*']],
    });
};

const badRequest = async (payload: object | object[], delayMilliseconds: number = 200) => {
    await delay(delayMilliseconds);
    return HttpResponse.json(payload, {
        status: 400,
        headers: [['Access-Control-Allow-Origin', '*']],
    });
};

export const getDomainsHandler = (apiUrl: string) =>
    http.get(apiUrl + ApiEndpoints.ADMIN, () => {
        return ok(config);
    });

export const getSingleDomainHandler = (apiUrl: string) =>
    http.get(apiUrl + ApiEndpoints.ADMIN + '/initial_domain', () => {
        return ok(singleConfig);
    });

export const createDomainHandler = (apiUrl: string) =>
    http.delete(apiUrl + ApiEndpoints.ADMIN, () => {
        return ok(singleConfig);
    });

export const deleteDomainHandler = (apiUrl: string) =>
    http.delete(apiUrl + ApiEndpoints.ADMIN + '/initial_domain', () => {
        return ok(singleConfig);
    });

export const getAsyncStatusHandler = (apiUrl: string) =>
    http.get(apiUrl + ApiEndpoints.ASYNC, ({ request }) => {
        const url = new URL(request.url);
        const test_id = url.searchParams.get('id');
        const useCase = url.searchParams.get('usecase');
        return ok(singleConfig);
    });

export const getPortalConfigHandler = (apiUrl: string) =>
    http.get(apiUrl + ApiEndpoints.PORTAL_CONFIG, () => {
        return ok(singleConfig);
    });

export const postPortalConfigHandler = (apiUrl: string) =>
    http.post(apiUrl + ApiEndpoints.PORTAL_CONFIG, () => {
        return ok(singleConfig);
    });

export const getAllErrorsHandler = (apiUrl: string) =>
    http.get(apiUrl + ApiEndpoints.ERROR, ({ request }) => {
        const url = new URL(request.url);
        const page = url.searchParams.get('page');
        const pageSize = url.searchParams.get('pageSize');
        return ok(singleConfig);
    });

export const deleteErrorHandler = (apiUrl: string) =>
    http.delete(apiUrl + ApiEndpoints.ERROR + '/test_error_id', () => {
        return ok(singleConfig);
    });

export const deleteAllErrorHandler = (apiUrl: string) =>
    http.delete(apiUrl + ApiEndpoints.ERROR + '/*', () => {
        return ok(singleConfig);
    });

export const getJobsHandler = (apiUrl: string) =>
    http.get(apiUrl + ApiEndpoints.JOB, () => {
        return ok(singleConfig);
    });

export const runJobHandler = (apiUrl: string) =>
    http.post(apiUrl + ApiEndpoints.PORTAL_CONFIG, () => {
        return ok(singleConfig);
    });

export const createPrivacySearchHandler = (apiUrl: string) =>
    http.post(apiUrl + ApiEndpoints.PRIVACY, () => {
        return ok(singleConfig);
    });

/**
 * @param apiUrl the base url for http requests. only requests to this base url will be intercepted and handled by mock-service-worker.
 */
export const handlers = (apiUrl: string) => [
    getDomainsHandler(apiUrl),
    getSingleDomainHandler(apiUrl),
    createDomainHandler(apiUrl),
    deleteDomainHandler(apiUrl),
    getAsyncStatusHandler(apiUrl),
    getPortalConfigHandler(apiUrl),
    postPortalConfigHandler(apiUrl),
    getAllErrorsHandler(apiUrl),
    deleteErrorHandler(apiUrl),
    deleteAllErrorHandler(apiUrl),
    getJobsHandler(apiUrl),
    runJobHandler(apiUrl),
    createPrivacySearchHandler(apiUrl),
];

export const config = returnConfig('tx_1', 'initial_domain', 0, true);
export const singleConfig = returnConfig('tx_1', 'initial_domain', 6, false);
