import * as s3 from 'aws-cdk-lib/aws-s3';

export class BusinessObjectPipelineOutput {
    connectorJobName: string
    bucket: s3.Bucket
}
