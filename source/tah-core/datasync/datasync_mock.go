// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package datasync

import (
	"tah/upt/source/tah-core/core"
)

type MockConfig struct {
	Tasks          *[]Task
	TaskExecutions *[]TaskExecutionStatus
}

func InitMock(tasks *[]Task, taskExecutions *[]TaskExecutionStatus) *MockConfig {
	config := &MockConfig{
		Tasks:          tasks,
		TaskExecutions: taskExecutions,
	}
	return config
}

func (c *MockConfig) SetTx(tx core.Transaction) error {
	return nil
}

func (c *MockConfig) CreateS3Location(bucketName, bucketPrefix, accessRoleArn string) (string, error) {
	return "", nil
}

func (c *MockConfig) DeleteLocation(locationArn string) error {
	return nil
}

func (c *MockConfig) CreateTask(taskName, sourceArn, destinationArn string, cloudwatchGroupArn *string) (string, error) {
	return "", nil
}

func (c *MockConfig) ListTasks() ([]Task, error) {
	if c.Tasks != nil {
		return *c.Tasks, nil
	}
	return []Task{}, nil
}

func (c *MockConfig) DeleteTask(taskArn string) error {
	return nil
}

func (c *MockConfig) StartTaskExecution(taskArn string) (string, error) {
	return "", nil
}

func (c *MockConfig) ListTaskExecutions() ([]TaskExecutionStatus, error) {
	if c.TaskExecutions != nil {
		return *c.TaskExecutions, nil
	}
	return []TaskExecutionStatus{}, nil
}
