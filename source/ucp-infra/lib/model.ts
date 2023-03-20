import * as s3 from 'aws-cdk-lib/aws-s3';
import { Column } from '@aws-cdk/aws-glue-alpha';

export class BusinessObjectPipelineOutput {
    connectorJobName: string
    customerJobName: string
    bucket: s3.Bucket
    tableName: string
}


export class GlueSchema {
    columns: Column[]
}