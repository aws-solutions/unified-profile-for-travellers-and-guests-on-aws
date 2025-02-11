// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package firehose

type MockConfig struct {
	DeliveryStream  *DeliveryStream
	Status          *string
	Records         *[]Record
	IngestionErrors *[]IngestionError
}

func InitMock(stream *DeliveryStream, status *string, records *[]Record, ingestionError *[]IngestionError) *MockConfig {
	return &MockConfig{
		DeliveryStream:  stream,
		Status:          status,
		Records:         records,
		IngestionErrors: ingestionError,
	}
}

func (c *MockConfig) CreateS3Stream(streamName string, bucketArn string, roleArn string) error {
	return nil
}

func (c *MockConfig) Delete(streamName string) error {
	return nil
}

func (c *MockConfig) WaitForStreamCreation(streamName string, timeoutSeconds int) (string, error) {
	if c.Status != nil {
		return *c.Status, nil
	}
	return "", nil
}

func (c *MockConfig) Describe(streamName string) (DeliveryStream, error) {
	if c.DeliveryStream != nil {
		return *c.DeliveryStream, nil
	}
	return DeliveryStream{}, nil
}

func (c *MockConfig) PutRecords(streamName string, recs []Record) ([]IngestionError, error) {
	c.Records = &recs
	if c.IngestionErrors != nil {
		return *c.IngestionErrors, nil
	}
	return []IngestionError{}, nil
}
