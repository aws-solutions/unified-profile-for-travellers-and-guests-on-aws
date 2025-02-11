// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as cdk from 'aws-cdk-lib';
import * as apiGatewayV2 from 'aws-cdk-lib/aws-apigatewayv2';
import * as cloudwatch from 'aws-cdk-lib/aws-cloudwatch';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as kinesis from 'aws-cdk-lib/aws-kinesis';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import { CdkBase, CdkBaseProps } from '../cdk-base';

type DashboardProps = {
    // Real-time ingestion
    readonly entryKinesisStream: kinesis.Stream;
    readonly entryLambdaFunction: lambda.Function;
    readonly ingestorKinesisStream: kinesis.Stream;
    readonly ingestorLambdaFunction: lambda.Function;
    // Change processor
    readonly changeProcessorKinesisStream: kinesis.Stream;
    readonly changeProcessorLambdaFunction: lambda.Function;
    // Error queues, Lambda function, and DynamoDB table
    readonly airBookingBatchErrorQueue: sqs.Queue;
    readonly clickStreamBatchErrorQueue: sqs.Queue;
    readonly customerServiceInteractionBatchErrorQueue: sqs.Queue;
    readonly guestProfileBatchErrorQueue: sqs.Queue;
    readonly hotelBookingBatchErrorQueue: sqs.Queue;
    readonly hotelStayBatchErrorQueue: sqs.Queue;
    readonly paxProfileBatchErrorQueue: sqs.Queue;
    readonly dlqChangeProcessor: sqs.Queue;
    readonly dlqGo: sqs.Queue;
    readonly dlqPython: sqs.Queue;
    readonly dlqBatchProcessor: sqs.Queue;
    readonly connectProfileDomainErrorQueue: sqs.Queue;
    readonly errorRetryLambdaFunction: lambda.Function;
    readonly errorTable: dynamodb.Table;
    // API
    readonly api: apiGatewayV2.HttpApi;
    // Sync Lambda function
    readonly syncLambdaFunction: lambda.Function;
    // CP Writer queue
    readonly cpWriterQueue: sqs.Queue;
};

type MetricProps = {
    readonly metricName: string;
    readonly namespace: string;
    readonly dimensionsMap: cloudwatch.DimensionsMap;
    readonly label: string;
    readonly statistic: string;
    readonly period?: cdk.Duration;
};

enum MetricNamespace {
    API_GATEWAY = 'AWS/ApiGateway',
    DYNAMODB = 'AWS/DynamoDB',
    KINESIS = 'AWS/Kinesis',
    LAMBDA = 'AWS/Lambda',
    SQS = 'AWS/SQS'
}

export class Dashboard extends CdkBase {
    public readonly dashboard: cloudwatch.Dashboard;
    private readonly props: DashboardProps;

    constructor(cdkProps: CdkBaseProps, props: DashboardProps) {
        super(cdkProps);

        this.props = props;

        // Dashboard
        this.dashboard = this.createDashboard();

        // Real-time ingestion
        this.createRealTimeIngestionWidgets();

        // Change processor
        this.createChangeProcessorWidgets();

        // Error
        this.createErrorWidgets();

        // API Gateway
        this.createApiGatewayWidgets();

        // Sync Lambda function
        this.createSyncLambdaFunctionWidgets();

        // Customer Profiles
        this.createCustomerProfilesWidgets();
    }

    /**
     * Creates a CloudWatch dashboard.
     * @returns CloudWatch dashboard
     */
    private createDashboard(): cloudwatch.Dashboard {
        return new cloudwatch.Dashboard(this.stack, 'ucp-cloudwatch-dash', {
            dashboardName: cdk.Fn.join('-', ['unified-traveler-profile', this.envName, cdk.Aws.REGION])
        });
    }

    /**
     * Creates real-time ingestion CloudWatch widgets which include
     * entry Kinesis data stream, entry Lambda function, ingestor Kinesis data stream, and ingestor Lambda function.
     */
    private createRealTimeIngestionWidgets(): void {
        const realTimeStreamWidget = new cloudwatch.GraphWidget({ title: 'Real Time stream (Kinesis)' });
        realTimeStreamWidget.addLeftMetric(
            this.createKinesisMetric(
                this.props.entryKinesisStream,
                '# Records fetched (transformer)',
                'GetRecords.Records',
                cloudwatch.Stats.SUM,
                1
            )
        );
        realTimeStreamWidget.addLeftMetric(
            this.createKinesisMetric(
                this.props.ingestorKinesisStream,
                '# Records fetched (ingestor)',
                'GetRecords.Records',
                cloudwatch.Stats.SUM,
                1
            )
        );

        const realTimeIteratorWidget = new cloudwatch.GraphWidget({ title: 'Real Time stream iterator(Kinesis)' });
        realTimeIteratorWidget.addLeftMetric(
            this.createKinesisMetric(
                this.props.entryKinesisStream,
                '# Records fetched (transformer)',
                'GetRecords.IteratorAgeMilliseconds',
                cloudwatch.Stats.MAXIMUM,
                5
            )
        );
        realTimeIteratorWidget.addLeftMetric(
            this.createKinesisMetric(
                this.props.ingestorKinesisStream,
                '# Records fetched (ingestor)',
                'GetRecords.IteratorAgeMilliseconds',
                cloudwatch.Stats.MAXIMUM,
                5
            )
        );

        const ingestorLambdaFunctionWidget = new cloudwatch.GraphWidget({ title: 'Real time ingestion (Lambda)' });
        ingestorLambdaFunctionWidget.addLeftMetric(
            this.createLambdaMetric(
                this.props.entryLambdaFunction,
                '# concurrent executions',
                'ConcurrentExecutions',
                cloudwatch.Stats.MAXIMUM
            )
        );

        this.dashboard.addWidgets(realTimeStreamWidget, realTimeIteratorWidget, ingestorLambdaFunctionWidget);
    }

    /**
     * Creates change processor CloudWatch widgets which include
     * change processor Kinesis data stream and change processor Lambda function.
     */
    private createChangeProcessorWidgets(): void {
        const changeProcessorStreamWidget = new cloudwatch.GraphWidget({ title: 'Change Processor (Kinesis)' });
        changeProcessorStreamWidget.addLeftMetric(
            this.createKinesisMetric(
                this.props.changeProcessorKinesisStream,
                '# Records fetched',
                'GetRecords.Records',
                cloudwatch.Stats.MAXIMUM,
                1
            )
        );

        const changeProcessorLambdaFunctionWidget = new cloudwatch.GraphWidget({ title: 'Change Processor (Lambda)' });
        changeProcessorLambdaFunctionWidget.addLeftMetric(
            this.createLambdaMetric(
                this.props.changeProcessorLambdaFunction,
                '# concurrent executions',
                'ConcurrentExecutions',
                cloudwatch.Stats.MAXIMUM
            )
        );

        this.dashboard.addWidgets(changeProcessorStreamWidget, changeProcessorLambdaFunctionWidget);
    }

    /**
     * Creates error widgets which include
     * error queues, an error retry Lambda function, and an error DynamoDB table.
     */
    private createErrorWidgets(): void {
        // Batch ingestion error
        const batchIngestionErrorWidget = new cloudwatch.GraphWidget({ title: 'Batch Ingestion errors (SQS)' });
        batchIngestionErrorWidget.addLeftMetric(
            this.createSqsMetric(this.props.hotelBookingBatchErrorQueue, 'Hotel booking', 'NumberOfMessagesSent', cloudwatch.Stats.SUM)
        );
        batchIngestionErrorWidget.addLeftMetric(
            this.createSqsMetric(this.props.airBookingBatchErrorQueue, 'Air booking', 'NumberOfMessagesSent', cloudwatch.Stats.SUM)
        );
        batchIngestionErrorWidget.addLeftMetric(
            this.createSqsMetric(this.props.guestProfileBatchErrorQueue, 'Guest Profiles', 'NumberOfMessagesSent', cloudwatch.Stats.SUM)
        );
        batchIngestionErrorWidget.addLeftMetric(
            this.createSqsMetric(this.props.paxProfileBatchErrorQueue, 'Pax Profile', 'NumberOfMessagesSent', cloudwatch.Stats.SUM)
        );
        batchIngestionErrorWidget.addLeftMetric(
            this.createSqsMetric(this.props.clickStreamBatchErrorQueue, 'Clickstream', 'NumberOfMessagesSent', cloudwatch.Stats.SUM)
        );
        batchIngestionErrorWidget.addLeftMetric(
            this.createSqsMetric(this.props.hotelStayBatchErrorQueue, 'Hotel Stay', 'NumberOfMessagesSent', cloudwatch.Stats.SUM)
        );
        batchIngestionErrorWidget.addLeftMetric(
            this.createSqsMetric(
                this.props.customerServiceInteractionBatchErrorQueue,
                'Customer Service Interaction',
                'NumberOfMessagesSent',
                cloudwatch.Stats.MAXIMUM
            )
        );
        batchIngestionErrorWidget.addLeftMetric(
            this.createSqsMetric(this.props.dlqBatchProcessor, 'Transformer errors', 'NumberOfMessagesSent', cloudwatch.Stats.SUM)
        );

        // Real-time ingestion error
        const realTimeIngestionErrorWidget = new cloudwatch.GraphWidget({ title: 'Real Time Ingestion Errors (SQS)' });
        realTimeIngestionErrorWidget.addLeftMetric(
            this.createSqsMetric(this.props.dlqPython, 'Transformer errors', 'NumberOfMessagesSent', cloudwatch.Stats.SUM)
        );
        realTimeIngestionErrorWidget.addLeftMetric(
            this.createSqsMetric(this.props.dlqGo, 'Ingestor errors', 'NumberOfMessagesSent', cloudwatch.Stats.SUM)
        );

        // Amazon Connect Customer Profiles error
        const accpErrorWidget = new cloudwatch.GraphWidget({ title: 'ACCP Error queues (SQS)' });
        accpErrorWidget.addLeftMetric(
            this.createSqsMetric(
                this.props.connectProfileDomainErrorQueue,
                '# errors from accp',
                'NumberOfMessagesSent',
                cloudwatch.Stats.SUM
            )
        );

        // Change processor error queue
        const changeProcessorErrorQueueWidget = new cloudwatch.GraphWidget({ title: 'Change Processor errors (SQS)' });
        changeProcessorErrorQueueWidget.addLeftMetric(
            this.createSqsMetric(
                this.props.dlqChangeProcessor,
                '# errors from change processor',
                'NumberOfMessagesSent',
                cloudwatch.Stats.MAXIMUM
            )
        );

        // Error retries DynamoDB stream
        const errorRetriesDynamoDbWidget = new cloudwatch.GraphWidget({ title: 'Error Retries (DDB streams)' });
        errorRetriesDynamoDbWidget.addLeftMetric(
            this.createDynamoDbMetric(
                this.props.errorTable,
                '# msg fetched',
                'ReturnedRecordsCount',
                cloudwatch.Stats.MAXIMUM,
                'GetRecords'
            )
        );

        // Error retries Lambda function
        const errorRetriesLambdaFunctionWidget = new cloudwatch.GraphWidget({ title: 'Error Retries (Lambda)' });
        errorRetriesLambdaFunctionWidget.addLeftMetric(
            this.createLambdaMetric(
                this.props.errorRetryLambdaFunction,
                '# concurrent executions',
                'ConcurrentExecutions',
                cloudwatch.Stats.MAXIMUM
            )
        );

        this.dashboard.addWidgets(
            batchIngestionErrorWidget,
            realTimeIngestionErrorWidget,
            accpErrorWidget,
            changeProcessorErrorQueueWidget,
            errorRetriesDynamoDbWidget,
            errorRetriesLambdaFunctionWidget
        );
    }

    /**
     * Creates API Gateway CloudWatch widgets which include
     * API Gateway traffic, errors, and response time.
     */
    private createApiGatewayWidgets(): void {
        const trafficWidget = new cloudwatch.GraphWidget({ title: 'UCP Backend traffic' });
        trafficWidget.addLeftMetric(this.createApiGatewayMetric(this.props.api, 'Api Gateway errors', '4xx', cloudwatch.Stats.MAXIMUM));
        trafficWidget.addLeftMetric(this.createApiGatewayMetric(this.props.api, 'Api Gateway errors', '5xx', cloudwatch.Stats.MAXIMUM));
        trafficWidget.addLeftMetric(this.createApiGatewayMetric(this.props.api, 'Api Gateway count', '	Count', cloudwatch.Stats.MAXIMUM));

        const responseTimeWidget = new cloudwatch.GraphWidget({ title: 'UCP Backend response Time' });
        responseTimeWidget.addLeftMetric(
            this.createApiGatewayMetric(this.props.api, 'responseTime', 'IntegrationLatency', cloudwatch.Stats.MAXIMUM)
        );

        this.dashboard.addWidgets(trafficWidget, responseTimeWidget);
    }

    /**
     * Creates sync Lambda function CloudWatch widgets which include
     * Lambda duration and errors.
     */
    private createSyncLambdaFunctionWidgets(): void {
        const durationWidget = new cloudwatch.GraphWidget({ title: 'Partition Sync Duration (Lambda)' });
        durationWidget.addLeftMetric(
            this.createLambdaMetric(this.props.syncLambdaFunction, 'Duration', 'Duration', cloudwatch.Stats.MAXIMUM)
        );

        const errorsWidget = new cloudwatch.GraphWidget({ title: 'Partition Sync Errors (Lambda)' });
        errorsWidget.addLeftMetric(this.createLambdaMetric(this.props.syncLambdaFunction, 'Exec Errors', 'Errors', cloudwatch.Stats.SUM));

        this.dashboard.addWidgets(durationWidget, errorsWidget);
    }

    /**
     * Creates a CloudWatch metric.
     * @param props `MetricProps` to create CloudWatch metric
     * @returns
     */
    private createMetric(props: MetricProps): cloudwatch.Metric {
        return new cloudwatch.Metric({
            ...props,
            period: props.period ?? cdk.Duration.minutes(1)
        });
    }

    /**
     * Creates CloudWatch widgets for Customer Profiles integration
     */
    createCustomerProfilesWidgets() {
        const customerProfilesIngestionWidget = new cloudwatch.GraphWidget({ title: 'Customer Profiles integration' });
        customerProfilesIngestionWidget.addLeftMetric(
            this.createSqsMetric(this.props.cpWriterQueue, 'Objects sent', 'NumberOfMessagesSent', cloudwatch.Stats.SUM)
        );
        customerProfilesIngestionWidget.addLeftMetric(
            this.createSqsMetric(
                this.props.cpWriterQueue,
                'Approximate age of oldest message',
                'ApproximateAgeOfOldestMessage',
                cloudwatch.Stats.MAXIMUM
            )
        );

        this.dashboard.addWidgets(customerProfilesIngestionWidget);
    }

    /**
     * Creates a Lambda function CloudWatch metric.
     * @param lambdaFunction Lambda function
     * @param label Metric label
     * @param metricName Metric name
     * @param statistic Metric statistic
     * @returns Lambda function CloudWatch metric
     */
    private createLambdaMetric(lambdaFunction: lambda.Function, label: string, metricName: string, statistic: string): cloudwatch.Metric {
        return this.createMetric({
            metricName,
            namespace: MetricNamespace.LAMBDA,
            dimensionsMap: { FunctionName: lambdaFunction.functionName },
            label,
            statistic
        });
    }

    /**
     * Creates an API Gateway CloudWatch metric.
     * @param api API Gateway HTTP API
     * @param label Metric label
     * @param metricName Metric name
     * @param statistic Metric statistic
     * @returns API CloudWatch metric
     */
    private createApiGatewayMetric(api: apiGatewayV2.HttpApi, label: string, metricName: string, statistic: string): cloudwatch.Metric {
        return this.createMetric({
            metricName: metricName,
            namespace: 'AWS/ApiGateway',
            dimensionsMap: { ApiId: api.apiId },
            label,
            statistic
        });
    }

    /**
     * Creates a DynamoDB CloudWatch metric.
     * @param table DynamoDB table
     * @param label Metric label
     * @param metricName Metric name
     * @param statistic Metric statistic
     * @param operation Metric operation
     * @returns DynamoDB CloudWatch metric
     */
    private createDynamoDbMetric(
        table: dynamodb.Table,
        label: string,
        metricName: string,
        statistic: string,
        operation: string
    ): cloudwatch.Metric {
        return this.createMetric({
            metricName: metricName,
            namespace: 'AWS/DynamoDB',
            dimensionsMap: {
                Operation: operation,
                StreamLabel: cdk.Fn.select(3, cdk.Fn.split('/', <string>table.tableStreamArn)),
                TableName: table.tableName
            },
            label,
            statistic
        });
    }

    /**
     * Creates a Kinesis stream CloudWatch metric.
     * @param stream Kinesis stream
     * @param label Metric label
     * @param metricName Metric name
     * @param statistic Metric statistic
     * @param period Metric period
     * @returns Kinesis stream CloudWatch metric
     */
    private createKinesisMetric(
        stream: kinesis.Stream,
        label: string,
        metricName: string,
        statistic: string,
        period: number
    ): cloudwatch.Metric {
        return this.createMetric({
            metricName,
            namespace: MetricNamespace.KINESIS,
            dimensionsMap: { StreamName: stream.streamName },
            label,
            statistic,
            period: cdk.Duration.minutes(period)
        });
    }

    /**
     * Creates a SQS queue CloudWatch metric.
     * @param queue SQS queue
     * @param label Metric label
     * @param metricName Metric name
     * @param statistic Metric statistic
     * @returns SQS queue CloudWatch metric
     */
    private createSqsMetric(queue: sqs.Queue, label: string, metricName: string, statistic: string): cloudwatch.Metric {
        return this.createMetric({
            metricName: metricName,
            namespace: 'AWS/SQS',
            dimensionsMap: { QueueName: queue.queueName },
            label,
            statistic
        });
    }
}
