import * as s3 from 'aws-cdk-lib/aws-s3';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import { Column } from '@aws-cdk/aws-glue-alpha';

export class BusinessObjectPipelineOutput {
    connectorJobName: string
    customerJobName: string
    bucket: s3.Bucket
    tableName: string
    errorQueue: sqs.Queue
}


export class GlueSchema {
    columns: Column[]
}