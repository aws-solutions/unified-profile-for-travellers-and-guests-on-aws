// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

/**
 * This unit tests are run with Jest mock.
 * The purpose of this one is to check if there's any business logic error and functional errors.
 */
import * as JSZip from 'jszip';
import { CustomError, mockAxios, mockCrypto, mockS3 } from './mock';

// application code must be loaded after mocks
import { handler } from '..';
import { CompletionStatus, CustomResourceActions, CustomResourceRequestTypes, StatusTypes } from '../lib';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const context: any = {
    logStreamName: 'log-stream-name'
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const createResponseBody = (event: any, response: CompletionStatus): string =>
    JSON.stringify({
        Status: response.Status,
        Reason: `See the details in CloudWatch Log Stream: ${context.logStreamName}`,
        PhysicalResourceId: event.LogicalResourceId,
        StackId: event.StackId,
        RequestId: event.RequestId,
        LogicalResourceId: event.LogicalResourceId,
        Data: response.Data
    });

beforeEach(() => mockAxios.mockReset());

describe('Create UUID', () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const event: any = {
        LogicalResourceId: 'CustomResource',
        StackId: 'UnifiedProfileForTravelersAndGuests',
        RequestId: 'CustomResource',
        ResponseURL: 'https://example.com',
        RequestType: CustomResourceRequestTypes.CREATE,
        ResourceProperties: {
            CustomAction: CustomResourceActions.CREATE_UUID
        }
    };

    beforeEach(() => mockCrypto.mockReset());

    test('UUID should be created when creating the solution', async () => {
        // Prepare
        mockCrypto.mockReturnValueOnce('uuid');
        mockAxios.mockResolvedValueOnce(undefined);

        // Call
        await handler(event, context);

        // Verify
        const responseBody = createResponseBody(event, {
            Status: StatusTypes.SUCCESS,
            Data: {
                UUID: 'uuid'
            }
        });
        expect(mockCrypto).toHaveBeenCalledTimes(1);
        expect(mockAxios).toHaveBeenCalledTimes(1);
        expect(mockAxios).toHaveBeenNthCalledWith(1, event.ResponseURL, responseBody, {
            headers: { 'Content-Type': '', 'Content-Length': responseBody.length }
        });
    });

    test('UUID should not be created when updating the solution', async () => {
        // Prepare
        event.RequestType = CustomResourceRequestTypes.UPDATE;
        mockAxios.mockResolvedValueOnce(undefined);

        // Call
        await handler(event, context);

        // Verify
        const responseBody = createResponseBody(event, {
            Status: StatusTypes.SUCCESS,
            Data: {}
        });
        expect(mockCrypto).not.toHaveBeenCalled();
        expect(mockAxios).toHaveBeenCalledTimes(1);
        expect(mockAxios).toHaveBeenNthCalledWith(1, event.ResponseURL, responseBody, {
            headers: { 'Content-Type': '', 'Content-Length': responseBody.length }
        });
    });
});

describe('Deploy frontend', () => {
    const fileMap = new Map<string, string>([
        ['index.html', '<html></html>'],
        ['index.js', 'console.log("Hello World!");']
    ]);
    const zip = new JSZip();

    for (const [key, val] of fileMap) {
        zip.file(key, val);
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const event: any = {
        LogicalResourceId: 'CustomResource',
        StackId: 'UnifiedProfileForTravelersAndGuests',
        RequestId: 'CustomResource',
        ResponseURL: 'https://example.com',
        RequestType: CustomResourceRequestTypes.CREATE,
        ResourceProperties: {
            CustomAction: CustomResourceActions.DEPLOY_FRONT_END,
            FrontEndConfig: {
                cognitoClientId: 'client-xyz',
                cognitoUserPoolId: 'userpool-id',
                deployed: '2024-02-05T00:00:00.000Z',
                ucpApiUrl: 'https://example.com',
                staticContentBucketName: 'content-bucket',
                artifactBucketName: 'artifact-bucket',
                artifactBucketPath: 'solution/ui/dist.zip',
                usePermissionSystem: 'true'
            }
        }
    };

    beforeEach(() => {
        mockAxios.mockReset();
        mockS3.getObject.mockReset();
        mockS3.putObject.mockReset();
    });

    test('Frontend should be deployed when creating the solution', async () => {
        // Prepare
        const zipStream = zip.generateNodeStream();
        mockAxios.mockResolvedValueOnce(undefined);
        mockS3.getObject.mockResolvedValueOnce({ Body: zipStream });
        mockS3.putObject.mockResolvedValue(undefined);

        // Call
        await handler(event, context);

        // Verify
        const responseBody = createResponseBody(event, {
            Status: StatusTypes.SUCCESS,
            Data: {}
        });
        expect(mockAxios).toHaveBeenCalledTimes(1);
        expect(mockAxios).toHaveBeenNthCalledWith(1, event.ResponseURL, responseBody, {
            headers: { 'Content-Type': '', 'Content-Length': responseBody.length }
        });
        expect(mockS3.getObject).toHaveBeenCalledTimes(1);
        expect(mockS3.putObject).toHaveBeenCalledTimes(fileMap.size + 1); // Zip files and a config file
    });

    test('When getting objects fails, custom resource should report failure', async () => {
        // Prepare
        const error = new CustomError('ErrorCode', 'Failure'); // To cover `Code` and `Message`, CustomError is used.
        mockS3.getObject.mockRejectedValueOnce(error);

        // Call
        await handler(event, context);

        // Verify
        const responseBody = createResponseBody(event, {
            Status: StatusTypes.FAILED,
            Data: {
                Error: {
                    Code: error.code,
                    Message: error.message
                }
            }
        });
        expect(mockAxios).toHaveBeenCalledTimes(1);
        expect(mockAxios).toHaveBeenNthCalledWith(1, event.ResponseURL, responseBody, {
            headers: { 'Content-Type': '', 'Content-Length': responseBody.length }
        });
        expect(mockS3.getObject).toHaveBeenCalledTimes(1);
        expect(mockS3.putObject).not.toHaveBeenCalled();
    });

    test('When putting objects fails, custom resource should report failure', async () => {
        // Prepare
        const error = 'Failure'; // To cover missing `Code` and `Message`, general string is used.
        const zipStream = zip.generateNodeStream();
        mockAxios.mockResolvedValueOnce(undefined);
        mockS3.getObject.mockResolvedValueOnce({ Body: zipStream });
        mockS3.putObject.mockRejectedValueOnce(error);

        // Call
        await handler(event, context);

        // Verify
        const responseBody = createResponseBody(event, {
            Status: StatusTypes.FAILED,
            Data: {
                Error: {
                    Code: 'CustomResourceError',
                    Message: 'Custom resource error occurred.'
                }
            }
        });
        expect(mockAxios).toHaveBeenCalledTimes(1);
        expect(mockAxios).toHaveBeenNthCalledWith(1, event.ResponseURL, responseBody, {
            headers: { 'Content-Type': '', 'Content-Length': responseBody.length }
        });
        expect(mockS3.getObject).toHaveBeenCalledTimes(1);
        expect(mockS3.putObject).toHaveBeenCalledTimes(1);
    });

    test('Frontend should be deployed when updating the solution', async () => {
        // Prepare
        event.RequestType = CustomResourceRequestTypes.UPDATE;
        const zipStream = zip.generateNodeStream();
        mockAxios.mockResolvedValueOnce(undefined);
        mockS3.getObject.mockResolvedValueOnce({ Body: zipStream });
        mockS3.putObject.mockResolvedValue(undefined);

        // Call
        await handler(event, context);

        // Verify
        const responseBody = createResponseBody(event, {
            Status: StatusTypes.SUCCESS,
            Data: {}
        });
        expect(mockAxios).toHaveBeenCalledTimes(1);
        expect(mockAxios).toHaveBeenNthCalledWith(1, event.ResponseURL, responseBody, {
            headers: { 'Content-Type': '', 'Content-Length': responseBody.length }
        });
        expect(mockS3.getObject).toHaveBeenCalledTimes(1);
        expect(mockS3.putObject).toHaveBeenCalledTimes(fileMap.size + 1); // Zip files and a config file
    });

    test('Frontend should not be deployed when deleting the solution', async () => {
        // Prepare
        event.RequestType = CustomResourceRequestTypes.DELETE;

        // Call
        await handler(event, context);

        // Verify
        const responseBody = createResponseBody(event, {
            Status: StatusTypes.SUCCESS,
            Data: {}
        });
        expect(mockAxios).toHaveBeenCalledTimes(1);
        expect(mockAxios).toHaveBeenNthCalledWith(1, event.ResponseURL, responseBody, {
            headers: { 'Content-Type': '', 'Content-Length': responseBody.length }
        });
        expect(mockS3.getObject).not.toHaveBeenCalled();
    });
});
