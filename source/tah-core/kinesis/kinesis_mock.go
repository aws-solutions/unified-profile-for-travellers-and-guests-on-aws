// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kinesis

type MockConfig struct {
	Stream          *Stream
	Status          *string
	Records         *[]Record
	IngestionErrors *[]IngestionError
}

func InitMock(stream *Stream, status *string, records *[]Record, ingestionError *[]IngestionError) *MockConfig {
	return &MockConfig{
		Stream:          stream,
		Status:          status,
		Records:         records,
		IngestionErrors: ingestionError,
	}
}

func (c *MockConfig) Create(streamName string) error {
	return nil
}

func (c *MockConfig) Delete(streamName string) error {
	return nil
}

func (c *MockConfig) WaitForStreamCreation(timeoutSeconds int) (string, error) {
	if c.Status != nil {
		return *c.Status, nil
	}
	return "", nil
}

func (c *MockConfig) Describe() (Stream, error) {
	if c.Stream != nil {
		return *c.Stream, nil
	}
	return Stream{}, nil
}

func (c *MockConfig) PutRecords(recs []Record) (error, []IngestionError) {
	if c.Records != nil {
		// appending new records to existing slice
		combinedRecs := append(*c.Records, recs...)
		c.Records = &combinedRecs
	} else {
		// creating slice with new records
		c.Records = &recs
	}
	if c.IngestionErrors != nil {
		return nil, *c.IngestionErrors
	}
	return nil, []IngestionError{}
}

func (c *MockConfig) InitConsumer(iteratorType string) error {
	return nil
}

func (c *MockConfig) FetchRecords(timeoutSeconds int) ([]Record, error) {
	if c.Records != nil {
		return *c.Records, nil
	}
	return []Record{}, nil
}
