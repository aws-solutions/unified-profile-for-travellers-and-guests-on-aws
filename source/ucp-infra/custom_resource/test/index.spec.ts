// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

/**
 * This unit tests run actual AWS SDK.
 * The purpose of this one is to check if the function actually work as expected on the real environment.
 */
import { S3 } from '@aws-sdk/client-s3';
import * as JSZip from 'jszip';
import { deployFrontEnd } from '../lib';

const s3 = new S3({});
// Generate random bucket name for testing purpose
const bucketName = `test-bucket-${Math.ceil(Math.random() * 10e10)}`;
const artifactBucketName = `test-bucket-artifact-${Math.ceil(Math.random() * 10e10)}`;

const fileMap = new Map<string, string>([
    ['index.html', '<html></html>'],
    ['index.js', 'console.log("Hello World!");']
]);
const zip = new JSZip();

for (const [key, val] of fileMap) {
    zip.file(key, val);
}

beforeAll(async () => {
    console.log('Creating bucket for testing');
    await s3.createBucket({ Bucket: bucketName });
    await s3.createBucket({ Bucket: artifactBucketName });

    const zipFile = await zip.generateAsync({ type: 'uint8array' });
    await s3.putObject({ Bucket: artifactBucketName, Key: 'ui/dist.zip', Body: zipFile });

    const objects = await s3.listObjects({ Bucket: artifactBucketName });
    console.log('Successfully uploaded test archive. Objects in bucket: ', objects.Contents);
}, 30000);

afterAll(async () => {
    console.log('Deleting bucket for testing');

    const deleteObjectPromises = [
        s3.deleteObject({ Bucket: artifactBucketName, Key: 'ui/dist.zip' }),
        s3.deleteObject({ Bucket: bucketName, Key: 'assets/ucp-config.json' })
    ];

    for (const key of fileMap.keys()) {
        deleteObjectPromises.push(s3.deleteObject({ Bucket: artifactBucketName, Key: key }));
    }

    await Promise.allSettled(deleteObjectPromises);
    await Promise.allSettled([s3.deleteBucket({ Bucket: bucketName }), s3.deleteBucket({ Bucket: artifactBucketName })]);
}, 30000);

describe('Custom resource tests', () => {
    test('Test S3 Upload', async () => {
        // Prepare
        const expected = ['assets/ucp-config.json'];

        for (const key of fileMap.keys()) {
            expected.push(key);
        }

        // Call
        await deployFrontEnd({
            artifactBucketPath: 'ui/dist.zip',
            artifactBucketName: artifactBucketName,
            cognitoClientId: 'userpool-client',
            cognitoUserPoolId: 'userpool',
            deployed: new Date().toISOString(),
            staticContentBucketName: bucketName,
            ucpApiUrl: 'https://example.com',
            usePermissionSystem: 'true'
        });

        // Verify
        // Check that objects are uploaded to bucket and that they have correct key
        const objects = await s3.listObjects({ Bucket: bucketName });
        const keys = objects.Contents?.map(content => content.Key);

        console.info('Contents', objects.Contents);

        expect(objects.Contents?.length).toBe(fileMap.size + 1);
        expect(keys).toEqual(expected);
    }, 60000);
});
