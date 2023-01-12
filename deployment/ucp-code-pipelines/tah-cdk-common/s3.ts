import { RemovalPolicy } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { aws_s3 as s3 } from 'aws-cdk-lib';
import { NagSuppressions } from 'cdk-nag';

export class Bucket extends s3.Bucket {
    //this constructor creates a default bucket that complies with all cdk_nag roles for AWS solutions
    constructor(scope: Construct, id: string, logsLocation: s3.Bucket) {
        super(scope, id, {
            removalPolicy: RemovalPolicy.DESTROY,
            versioned: true,
            serverAccessLogsBucket: logsLocation,
            serverAccessLogsPrefix: "bucket-" + id,
            encryption: s3.BucketEncryption.S3_MANAGED,
            blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
            enforceSSL: true,
        })
    }

}

//bucket to store access logs created
export class AccessLogBucket extends s3.Bucket {
    //this constructor creates a default bucket that complies with all cdk_nag roles for AWS solutions
    constructor(scope: Construct, id: string) {
        super(scope, id, {
            removalPolicy: RemovalPolicy.DESTROY,
            versioned: true,
            encryption: s3.BucketEncryption.S3_MANAGED,
            blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
            enforceSSL: true,
        })
        NagSuppressions.addResourceSuppressions(this, [
            {
                id: 'AwsSolutions-S1',
                reason: 'bucket used to store access logs'
            },
        ]);
    }
}