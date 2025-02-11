// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package datasync

import (
	"tah/upt/source/tah-core/core"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/datasync"
)

type IConfig interface {
	SetTx(tx core.Transaction) error
	CreateS3Location(bucketName, bucketPrefix, accessRoleArn string) (string, error)
	DeleteLocation(locationArn string) error
	CreateTask(taskName, sourceArn, destinationArn string, cloudwatchGroupArn *string) (string, error)
	ListTasks() ([]Task, error)
	DeleteTask(taskArn string) error
	StartTaskExecution(taskArn string) (string, error)
	ListTaskExecutions() ([]TaskExecutionStatus, error)
}

type DataSyncConfig struct {
	Client *datasync.DataSync
	Tx     core.Transaction
}

type Task struct {
	Name   string
	Status string
	ARN    string
}

type TaskExecutionStatus struct {
	ExecutionArn    string
	ExecutionStatus string
}

// Task status options
const (
	TaskExecutionStatusQueued       string = "QUEUED"
	TaskExecutionStatusLaunching    string = "LAUNCHING"
	TaskExecutionStatusPreparing    string = "PREPARING"
	TaskExecutionStatusTransferring string = "TRANSFERRING"
	TaskExecutionStatusVerifying    string = "VERIFYING"
	TaskExecutionStatusSuccess      string = "SUCCESS"
	TaskExecutionStatusError        string = "ERROR"
)

func InitRegion(region string, solutionId, solutionVersion string, logLevel core.LogLevel) *DataSyncConfig {
	session := session.Must(session.NewSession())
	client := core.CreateClient(solutionId, solutionVersion)
	svc := datasync.New(session, aws.NewConfig().WithRegion(region).WithHTTPClient(client))
	return &DataSyncConfig{
		Client: svc,
		Tx:     core.NewTransaction("datasync", "", logLevel),
	}
}

func (c *DataSyncConfig) SetTx(tx core.Transaction) error {
	tx.LogPrefix = "datasync"
	c.Tx = tx
	return nil
}

func (c *DataSyncConfig) CreateS3Location(bucketName, bucketPrefix, accessRoleArn string) (string, error) {
	input := &datasync.CreateLocationS3Input{
		S3BucketArn: aws.String(core.S3BucketNameToARN(bucketName)),
		S3Config: &datasync.S3Config{
			BucketAccessRoleArn: aws.String(accessRoleArn),
		},
		S3StorageClass: aws.String(datasync.S3StorageClassStandard),
		Subdirectory:   aws.String(bucketPrefix),
	}

	output, err := c.Client.CreateLocationS3(input)
	if err != nil {
		c.Tx.Error("[DataSync][CreateS3Location] Error: %v", err)
		return "", err
	}

	return aws.StringValue(output.LocationArn), nil
}

func (c *DataSyncConfig) ListLocations() ([]string, error) {
	locations := []string{}
	output, err := c.Client.ListLocations(&datasync.ListLocationsInput{})
	if err != nil {
		c.Tx.Error("[DataSync][ListLocations] Error: %v", err)
		return locations, err
	}
	for _, location := range output.Locations {
		locations = append(locations, aws.StringValue(location.LocationArn))
	}
	return locations, nil
}

func (c *DataSyncConfig) DeleteLocation(locationArn string) error {
	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(locationArn),
	}
	_, err := c.Client.DeleteLocation(input)
	return err
}

func (c *DataSyncConfig) CreateTask(taskName, sourceArn, destinationArn string, cloudwatchGroupArn *string) (string, error) {
	input := &datasync.CreateTaskInput{
		CloudWatchLogGroupArn:  cloudwatchGroupArn,
		DestinationLocationArn: aws.String(destinationArn),
		Name:                   aws.String(taskName),
		Options: &datasync.Options{
			TransferMode: aws.String(datasync.TransferModeAll),
		},
		SourceLocationArn: aws.String(sourceArn),
	}

	output, err := c.Client.CreateTask(input)
	if err != nil {
		c.Tx.Error("[DataSync][CreateTask] Error: %v", err)
	}

	return aws.StringValue(output.TaskArn), err
}

func (c *DataSyncConfig) ListTasks() ([]Task, error) {
	tasks := []Task{}
	output, err := c.Client.ListTasks(&datasync.ListTasksInput{})
	if err != nil {
		c.Tx.Error("[DataSync][ListTasks] Error: %v", err)
		return tasks, err
	}

	for _, t := range output.Tasks {
		task := Task{
			Name:   aws.StringValue(t.Name),
			Status: aws.StringValue(t.Status),
			ARN:    aws.StringValue(t.TaskArn),
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (c *DataSyncConfig) DeleteTask(taskArn string) error {
	input := &datasync.DeleteTaskInput{
		TaskArn: aws.String(taskArn),
	}
	_, err := c.Client.DeleteTask(input)
	if err != nil {
		c.Tx.Error("[DataSync][DeleteTask] Error: %v", err)
	}
	return err
}

func (c *DataSyncConfig) StartTaskExecution(taskArn string) (string, error) {
	input := &datasync.StartTaskExecutionInput{
		TaskArn: aws.String(taskArn),
	}

	output, err := c.Client.StartTaskExecution(input)
	if err != nil {
		c.Tx.Error("[DataSync][StartTaskExecution] Error: %v", err)
		return "", err
	}

	return aws.StringValue(output.TaskExecutionArn), nil
}

func (c *DataSyncConfig) ListTaskExecutions() ([]TaskExecutionStatus, error) {
	taskExecutionStatuses := []TaskExecutionStatus{}
	output, err := c.Client.ListTaskExecutions(&datasync.ListTaskExecutionsInput{})
	if err != nil {
		c.Tx.Error("[DataSync][ListTaskExecutions] Error: %v", err)
		return taskExecutionStatuses, err
	}

	for _, task := range output.TaskExecutions {
		status := TaskExecutionStatus{
			ExecutionArn:    aws.StringValue(task.TaskExecutionArn),
			ExecutionStatus: aws.StringValue(task.Status),
		}
		taskExecutionStatuses = append(taskExecutionStatuses, status)
	}

	return taskExecutionStatuses, nil
}
