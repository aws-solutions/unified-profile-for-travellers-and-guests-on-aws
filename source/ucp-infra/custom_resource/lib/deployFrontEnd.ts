// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { PutObjectCommandInput, S3 } from '@aws-sdk/client-s3';
import * as mime from 'mime-types';
import * as unzipper from 'unzipper';
import { FrontEndConfig } from './types';

const s3 = new S3({});

/**
 * Deploy Front End
 * "FrontEndConfig": {
 *   "artifactBucketName": "artifact-bucket-name",
 *   "artifactBucketPath": "artifact-bucket-path",
 *   "cognitoClientId": "user-pool-client-id",
 *   "cognitoUserPoolId": "user-pool-id",
 *   "staticContentBucketName": "static-content-bucket-name",
 *   "ucpApiUrl": "https://example.com/"
 * }
 */
export async function deployFrontEnd(frontEndConfig: FrontEndConfig): Promise<void> {
    const bucketName = frontEndConfig.staticContentBucketName;
    const artifactBucketPath = frontEndConfig.artifactBucketPath;
    const artifactBucket = frontEndConfig.artifactBucketName;

    try {
        // Get stream of the file to be extracted from the zip
        console.info(`Downloading s3://${artifactBucket}/${artifactBucketPath}`);
        const item = await s3.getObject({ Bucket: artifactBucket, Key: artifactBucketPath });
        const body = <NodeJS.ReadableStream>item.Body;
        const zip = body.pipe(unzipper.Parse({ forceStream: true }));

        // Upload files to S3 bucket
        for await (const entry of zip) {
            const fileName = entry.path;
            const entryMimeType = mime.lookup(fileName);
            const content = await entry.buffer();
            const key = fileName;
            console.info(`Uploading s3://${bucketName}/${key}`);

            const input: PutObjectCommandInput = { Bucket: bucketName, Key: key, Body: content };

            if (entry.type === 'File' && entryMimeType) {
                console.info('ContentType: ', entryMimeType);
                input.ContentType = entryMimeType;
            }

            await s3.putObject(input);
        }
    } catch (error) {
        console.error('Error: during archive extraction', error);
        throw error;
    }

    console.info('Successfully copied all objects');

    // Upload UI config
    const config = {
        cognitoClientId: frontEndConfig.cognitoClientId,
        cognitoUserPoolId: frontEndConfig.cognitoUserPoolId,
        region: process.env.AWS_REGION,
        ucpApiUrl: frontEndConfig.ucpApiUrl,
        usePermissionSystem: frontEndConfig.usePermissionSystem
    };

    const configKey = 'assets/ucp-config.json';
    console.info('Uploading config file', config, ' to ', bucketName, configKey);

    const configInput: PutObjectCommandInput = { Bucket: bucketName, Key: configKey, Body: JSON.stringify(config) };
    await s3.putObject(configInput);
    console.info('Successfully uploaded config file');
}
