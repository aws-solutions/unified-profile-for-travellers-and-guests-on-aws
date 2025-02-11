// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { CloudFormationCustomResourceEvent, Context } from 'aws-lambda';
import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';
import * as crypto from 'crypto';
import { CompletionStatus, CustomResourceActions, CustomResourceRequestTypes, StatusTypes, deployFrontEnd } from './lib';

/**
 * Custom resource Lambda handler.
 * @param event The CloudFormation custom resource event
 * @param context The Lambda context
 */
export async function handler(event: CloudFormationCustomResourceEvent, context: Context): Promise<void> {
    console.info('Received event:', JSON.stringify(event, null, 2));

    const { RequestType, ResourceProperties } = event;
    const response: CompletionStatus = {
        Status: StatusTypes.SUCCESS,
        Data: {}
    };

    try {
        if (ResourceProperties.CustomAction === CustomResourceActions.CREATE_UUID && RequestType === CustomResourceRequestTypes.CREATE) {
            response.Data = { UUID: crypto.randomUUID() };
        } else if (
            ResourceProperties.CustomAction === CustomResourceActions.DEPLOY_FRONT_END &&
            [CustomResourceRequestTypes.CREATE.toString(), CustomResourceRequestTypes.UPDATE.toString()].includes(RequestType)
        ) {
            console.info('Deploying front end with params: ', ResourceProperties);
            await deployFrontEnd(ResourceProperties.FrontEndConfig);
        }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
        console.error(`Error occurred at ${event.RequestType}::${ResourceProperties.CustomAction}`, error);

        response.Status = StatusTypes.FAILED;
        response.Data.Error = {
            Code: error.code ?? 'CustomResourceError',
            Message: error.message ?? 'Custom resource error occurred.'
        };
    } finally {
        await sendCloudFormationResponse(event, context.logStreamName, response);
    }
}

/**
 * Send custom resource response.
 * @param event Custom resource event.
 * @param logStreamName Custom resource log stream name.
 * @param response Response completion status.
 * @returns The promise of the sent request.
 */
async function sendCloudFormationResponse(
    event: CloudFormationCustomResourceEvent,
    logStreamName: string,
    response: CompletionStatus
): Promise<AxiosResponse> {
    const responseBody = JSON.stringify({
        Status: response.Status,
        Reason: `See the details in CloudWatch Log Stream: ${logStreamName}`,
        PhysicalResourceId: event.LogicalResourceId,
        StackId: event.StackId,
        RequestId: event.RequestId,
        LogicalResourceId: event.LogicalResourceId,
        Data: response.Data
    });

    const config: AxiosRequestConfig = {
        headers: {
            'Content-Type': '',
            'Content-Length': responseBody.length
        }
    };

    return axios.put(event.ResponseURL, responseBody, config);
}
