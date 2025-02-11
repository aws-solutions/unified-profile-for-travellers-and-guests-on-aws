// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package glue

type JobRunRq struct {
	JobName   string
	Arguments map[string]string
}

type MockConfig struct {
	JobRunRequested []JobRunRq
	Partitions      []Partition
}

var _ IConfig = &MockConfig{}

func InitMock(options ...func(*MockConfig)) *MockConfig {
	return &MockConfig{}
}

func WithJobRunsRequested(jobRunRqs []JobRunRq) func(*MockConfig) {
	return func(m *MockConfig) {
		m.JobRunRequested = jobRunRqs
	}
}

func WithPartitions(partitions []Partition) func(*MockConfig) {
	return func(m *MockConfig) {
		m.Partitions = partitions
	}
}

func (m *MockConfig) CreateTable(name string, bucketName string, partitionKeys []PartitionKey, schema Schema) error {
	return nil
}

func (m *MockConfig) DeleteTable(name string) error {
	return nil
}

func (m *MockConfig) RunJob(name string, args map[string]string) error {
	m.JobRunRequested = append(m.JobRunRequested, JobRunRq{
		JobName:   name,
		Arguments: args,
	})
	return nil
}

func (m *MockConfig) AddPartitionsToTable(tableName string, partitions []Partition) []error {
	m.Partitions = append(m.Partitions, partitions...)
	return []error{}
}

func (m *MockConfig) AddParquetPartitionsToTable(tableName string, partitions []Partition) []error {
	m.Partitions = append(m.Partitions, partitions...)
	return []error{}
}

func (m *MockConfig) GetPartitionsFromTable(tableName string) ([]Partition, error) {
	return m.Partitions, nil
}

func (m *MockConfig) RemovePartitionsFromTable(tableName string, partitions []Partition) []error {
	// this is not a pattern we want to adopt, effective for this mock use case
	updated := []Partition{}
	for _, p := range m.Partitions {
		for _, toDelete := range partitions {
			if p.Location == toDelete.Location && slicesEqual(p.Values, toDelete.Values) {
				continue
			}
			updated = append(updated, p)
		}
	}
	m.Partitions = updated

	return []error{}
}

// Taken from slices package introduced in Go 1.21
// https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/slices/slices.go;l=18
func slicesEqual[S ~[]E, E comparable](s1, s2 S) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
