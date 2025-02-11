// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import axios from 'axios';

export const mockAxios = jest.fn();
jest.mock('axios');
axios.put = mockAxios;

export const mockS3 = {
    getObject: jest.fn(),
    putObject: jest.fn()
};
jest.mock('@aws-sdk/client-s3', () => ({
    ...jest.requireActual('@aws-sdk/client-s3'),
    S3: jest.fn(() => mockS3)
}));

export const mockCrypto = jest.fn();
jest.mock('crypto', () => ({
    randomUUID: mockCrypto
}));

export class CustomError extends Error {
    public code: string;

    constructor(code: string, message: string) {
        super(message);
        this.code = code;
    }
}
