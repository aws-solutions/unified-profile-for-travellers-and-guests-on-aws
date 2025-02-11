// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package firehose

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/s3"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsFirehose "github.com/aws/aws-sdk-go/service/firehose"
)

var DEFAULT_BATCH_SIZE = 100
var WAIT_DELAY = 5
var ITERATOR_TYPE_AT_SEQUENCE_NUMBER = "AT_SEQUENCE_NUMBER"
var ITERATOR_TYPE_AFTER_SEQUENCE_NUMBER = "AFTER_SEQUENCE_NUMBER"
var ITERATOR_TYPE_TRIM_HORIZON = "TRIM_HORIZON"
var ITERATOR_TYPE_LATEST = "LATEST"
var ITERATOR_TYPE_AT_TIMESTAMP = "AT_TIMESTAMP"

type IConfig interface {
	CreateS3Stream(streamName string, bucketArn string, roleArn string) error
	Delete(streamName string) error
	WaitForStreamCreation(streamName string, timeoutSeconds int) (string, error)
	Describe(streamName string) (DeliveryStream, error)
	PutRecords(streamName string, recs []Record) ([]IngestionError, error)
}

type Config struct {
	Client             *awsFirehose.Firehose
	Region             string
	DeliveryStreamName string
}

var _ IConfig = (*Config)(nil)

type Record struct {
	Data string
}

type DeliveryStream struct {
	Status string
}

type IngestionError struct {
	ErrorCore    string
	ErrorMessage string
	RecordID     string
}

type PartitionKeyVal struct {
	Key string
	Val string
}

func InitAndCreate(streamName, bucketArn, roleArn, region, solutionId, solutionVersion string) (*Config, error) {
	cfg := Init(streamName, region, solutionId, solutionVersion)
	err := cfg.CreateS3Stream(streamName, bucketArn, roleArn)
	return cfg, err
}
func InitAndCreateWithPartitioning(streamName, bucketArn, roleArn string, partitions []PartitionKeyVal, region, solutionId, solutionVersion string) (*Config, error) {
	cfg := Init(streamName, region, solutionId, solutionVersion)
	err := cfg.CreateS3StreamWithDynamicPartitioning(streamName, bucketArn, roleArn, partitions)
	return cfg, err
}

func InitAndCreateWithPartitioningAndDataFormatConversion(streamName, bucketArn, roleArn string, partitions []PartitionKeyVal, region, solutionId, solutionVersion string, dataFormatConversionConfiguration *awsFirehose.DataFormatConversionConfiguration) (*Config, error) {
	cfg := Init(streamName, region, solutionId, solutionVersion)
	err := cfg.CreateS3StreamWithDynamicPartitioningDataFormatConversion(streamName, bucketArn, roleArn, partitions, dataFormatConversionConfiguration)
	return cfg, err
}

func Init(streamName, region, solutionId, solutionVersion string) *Config {
	mySession := session.Must(session.NewSession())
	httpClient := core.CreateClient(solutionId, solutionVersion)
	cfg := aws.NewConfig().WithRegion(region).WithMaxRetries(0).WithHTTPClient(httpClient)
	svc := awsFirehose.New(mySession, cfg)
	return &Config{
		Client:             svc,
		Region:             region,
		DeliveryStreamName: streamName,
	}
}

// utility function to wait for data in S3
func WaitForData(s3c s3.S3Config, timeout int) error {
	return s3c.WaitForData(timeout)
}

func WaitForObject(s3c s3.S3Config, prefix string, timeout int) error {
	return s3c.WaitForObject(prefix, timeout)
}

func (c *Config) CreateS3Stream(streamName string, bucketArn string, roleArn string) error {
	in := &awsFirehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(c.DeliveryStreamName),
		ExtendedS3DestinationConfiguration: &awsFirehose.ExtendedS3DestinationConfiguration{
			BucketARN:      aws.String(bucketArn),
			BufferingHints: &awsFirehose.BufferingHints{IntervalInSeconds: aws.Int64(60)},
			RoleARN:        aws.String(roleArn),
			ProcessingConfiguration: &awsFirehose.ProcessingConfiguration{
				Enabled: aws.Bool(true),
				Processors: []*awsFirehose.Processor{
					{
						Type: aws.String("AppendDelimiterToRecord"),
						Parameters: []*awsFirehose.ProcessorParameter{
							{
								ParameterName:  aws.String("Delimiter"),
								ParameterValue: aws.String("\\n"),
							},
						},
					},
				},
			},
		},
	}
	_, err := c.Client.CreateDeliveryStream(in)
	return err
}

func (c *Config) CreateS3StreamWithDynamicPartitioning(streamName string, bucketArn string, roleArn string, partitions []PartitionKeyVal) error {
	return c.CreateS3StreamWithDynamicPartitioningDataFormatConversion(streamName, bucketArn, roleArn, partitions, nil)
}

func (c *Config) CreateS3StreamWithDynamicPartitioningDataFormatConversion(streamName string, bucketArn string, roleArn string, partitions []PartitionKeyVal, dataFormatConversionConfiguration *awsFirehose.DataFormatConversionConfiguration) error {
	in := &awsFirehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(c.DeliveryStreamName),
		DeliveryStreamEncryptionConfigurationInput: &awsFirehose.DeliveryStreamEncryptionConfigurationInput{
			KeyType: aws.String("AWS_OWNED_CMK"),
		},
		ExtendedS3DestinationConfiguration: &awsFirehose.ExtendedS3DestinationConfiguration{
			BucketARN:      aws.String(bucketArn),
			BufferingHints: &awsFirehose.BufferingHints{IntervalInSeconds: aws.Int64(60)},
			RoleARN:        aws.String(roleArn),
			DynamicPartitioningConfiguration: &awsFirehose.DynamicPartitioningConfiguration{
				Enabled: aws.Bool(true),
			},
			DataFormatConversionConfiguration: dataFormatConversionConfiguration,
			ErrorOutputPrefix:                 aws.String("error/!{firehose:error-output-type}/"),
			Prefix:                            aws.String(buildPrefix(partitions)),
			ProcessingConfiguration: &awsFirehose.ProcessingConfiguration{
				Enabled: aws.Bool(true),
				Processors: []*awsFirehose.Processor{
					{
						Type: aws.String("MetadataExtraction"),
						Parameters: []*awsFirehose.ProcessorParameter{
							{
								ParameterName:  aws.String("MetadataExtractionQuery"),
								ParameterValue: aws.String(buildProcessorParameterValue(partitions)),
							},
							{
								ParameterName:  aws.String("JsonParsingEngine"),
								ParameterValue: aws.String("JQ-1.6"),
							},
						},
					},
					{
						Type: aws.String("AppendDelimiterToRecord"),
						Parameters: []*awsFirehose.ProcessorParameter{
							{
								ParameterName:  aws.String("Delimiter"),
								ParameterValue: aws.String("\\n"),
							},
						},
					},
				},
			},
		},
	}
	log.Printf("Stream creation input: %+v", in)
	_, err := c.Client.CreateDeliveryStream(in)
	return err
}

func buildProcessorParameterValue(partitions []PartitionKeyVal) string {
	parts := []string{}
	for _, partition := range partitions {
		parts = append(parts, partition.Key+": "+partition.Val)
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func buildPrefix(partitions []PartitionKeyVal) string {
	parts := []string{"data"}
	for _, partition := range partitions {
		parts = append(parts, partition.Key+"=!{partitionKeyFromQuery:"+partition.Key+"}")
	}
	return strings.Join(parts, "/") + "/"
}

func (c *Config) WaitForStreamCreation(streamName string, timeoutSeconds int) (string, error) {
	log.Printf("Waiting for crawler run to complete")
	stream, err := c.Describe(streamName)
	if err != nil {
		return "", err
	}
	it := 0
	for stream.Status != awsFirehose.DeliveryStreamStatusActive {
		log.Printf("Stream State: %v Waiting %v seconds before checking again", stream.Status, WAIT_DELAY)
		time.Sleep(time.Duration(WAIT_DELAY) * time.Second)
		stream, err = c.Describe(streamName)
		if err != nil {
			return "", err
		}
		it += 1
		if it*WAIT_DELAY >= timeoutSeconds {
			return "", fmt.Errorf("stream wait timed out after %v seconds", it*WAIT_DELAY)
		}
	}
	log.Printf("Crawler State: %v. Completed", stream.Status)
	return stream.Status, err
}

func (c *Config) Describe(streamName string) (DeliveryStream, error) {
	in := &awsFirehose.DescribeDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
	}
	out, err := c.Client.DescribeDeliveryStream(in)
	log.Printf("[Describe] DescribeStream output: %+v", out)
	if err != nil {
		return DeliveryStream{}, err
	}
	return DeliveryStream{
		Status: *out.DeliveryStreamDescription.DeliveryStreamStatus,
	}, err
}

func (c *Config) Delete(streamName string) error {
	in := &awsFirehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(streamName),
	}
	_, err := c.Client.DeleteDeliveryStream(in)
	return err
}

func (c *Config) PutRecords(streamName string, recs []Record) ([]IngestionError, error) {
	log.Printf("[kinesis] putting %v records to Kinesis data stream %v by batch of %v", len(recs), streamName, DEFAULT_BATCH_SIZE)
	batchedRecs := core.Chunk(core.InterfaceSlice(recs), DEFAULT_BATCH_SIZE)
	errs := []IngestionError{}
	for _, batch := range batchedRecs {
		in := &awsFirehose.PutRecordBatchInput{
			DeliveryStreamName: aws.String(streamName),
			Records:            []*awsFirehose.Record{},
		}
		for _, record := range batch {
			in.Records = append(in.Records, &awsFirehose.Record{
				Data: []byte(record.(Record).Data),
			})
		}
		out, err := c.Client.PutRecordBatch(in)
		if err != nil {
			log.Printf("[kinesis] error putting records to Kinesis data stream %v: %v", streamName, err)
		}
		for _, failedRec := range out.RequestResponses {
			if failedRec.ErrorCode != nil && failedRec.ErrorMessage != nil {
				recId := ""
				if failedRec.RecordId != nil {
					recId = *failedRec.RecordId
				}
				errs = append(errs, IngestionError{
					ErrorCore:    *failedRec.ErrorCode,
					ErrorMessage: *failedRec.ErrorMessage,
					RecordID:     recId,
				})
			}
		}
	}
	if len(errs) > 0 {
		log.Printf("[kinesis] failed to ingest %v: records", len(errs))
		return errs, errors.New("Failed to ingests " + strconv.Itoa(len(errs)) + " records")
	}
	return []IngestionError{}, nil
}
