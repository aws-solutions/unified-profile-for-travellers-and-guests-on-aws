// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package appflow

import (
	"log"
	"sync"
	core "tah/upt/source/tah-core/core"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	appflowSdk "github.com/aws/aws-sdk-go/service/appflow"
)

type Config struct {
	Client *appflowSdk.Appflow
}

type IConfig interface {
	GetFlow(flowName string) (Flow, error)
	GetFlows(names []string) ([]Flow, error)
	StartFlow(name string) (FlowStatusOutput, error)
	StopFlow(name string) (FlowStatusOutput, error)
	DeleteFlow(name string, forceDelete bool) error
}

type Flow struct {
	Name           string
	Description    string
	Status         string
	StatusMessage  string
	LastRun        time.Time
	LastRunStatus  string
	LastRunMessage string
	SourceType     string
	SourceDetails  string
	TargetType     string
	TargetDetails  string
	Trigger        string
}

type FlowStatusOutput struct {
	ExecutionId string
	FlowArn     string
	FlowStatus  string
}

const (
	FLOW_STATUS_ACTIVE     string = "Active"
	FLOW_STATUS_DEPRECATED string = "Deprecated"
	FLOW_STATUS_DELETED    string = "Deleted"
	FLOW_STATUS_DRAFT      string = "Draft"
	FLOW_STATUS_ERRORED    string = "Errored"
	FLOW_STATUS_SUSPENDED  string = "Suspended"
)

func Init(solutionId, solutionVersion string) Config {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := core.CreateClient(solutionId, solutionVersion)

	return Config{
		Client: appflowSdk.New(session, &aws.Config{HTTPClient: client}),
	}
}

func (c Config) GetFlow(flowName string) (Flow, error) {
	input := &appflowSdk.DescribeFlowInput{
		FlowName: aws.String(flowName),
	}
	out, err := c.Client.DescribeFlow(input)
	if err != nil {
		return Flow{}, err
	}
	flow := Flow{
		Name:          aws.StringValue(out.FlowName),
		Description:   aws.StringValue(out.Description),
		Status:        aws.StringValue(out.FlowStatus),
		SourceType:    aws.StringValue(out.SourceFlowConfig.ConnectorType),
		SourceDetails: buildSourceDetails(*out.SourceFlowConfig),
		TargetType:    aws.StringValue(out.DestinationFlowConfigList[0].ConnectorType),
		TargetDetails: buildTargetDetails(*out.DestinationFlowConfigList[0]),
		Trigger:       buildTriggerDescription(*out.TriggerConfig),
	}
	if out.FlowStatusMessage != nil {
		flow.StatusMessage = *out.FlowStatusMessage
	}
	if out.LastRunExecutionDetails != nil {
		flow.LastRun = *out.LastRunExecutionDetails.MostRecentExecutionTime
		flow.LastRunStatus = *out.LastRunExecutionDetails.MostRecentExecutionStatus
		if out.LastRunExecutionDetails.MostRecentExecutionMessage != nil {
			flow.LastRunMessage = *out.LastRunExecutionDetails.MostRecentExecutionMessage
		}
	}
	return flow, nil
}

func buildTriggerDescription(cfg appflowSdk.TriggerConfig) string {
	if aws.StringValue(cfg.TriggerType) == "Scheduled" {
		return "Scheduled on " + aws.StringValue(
			cfg.TriggerProperties.Scheduled.ScheduleExpression,
		) + " starting " + aws.TimeValue(cfg.TriggerProperties.Scheduled.ScheduleStartTime).
			Format("Mon Jan 2 15:04:05")
	}
	return *cfg.TriggerType
}

func buildSourceDetails(sfc appflowSdk.SourceFlowConfig) string {
	if *sfc.ConnectorType == "S3" {
		props := sfc.SourceConnectorProperties.S3
		if props != nil {
			return *props.BucketName + "/" + *props.BucketPrefix
		}
	}
	return ""
}

func buildTargetDetails(dfc appflowSdk.DestinationFlowConfig) string {
	if *dfc.ConnectorType == "CustomerProfiles" {
		props := (*dfc.DestinationConnectorProperties).CustomerProfiles
		if props != nil {
			return *props.DomainName + "/" + *props.ObjectTypeName
		}
	}
	return ""
}

func (c Config) GetFlows(names []string) ([]Flow, error) {
	log.Printf("[core][appflow] Getting all Flows")
	var err error
	var flowMutex sync.Mutex
	flows := []Flow{}
	var wg sync.WaitGroup
	wg.Add(len(names))
	for _, name := range names {
		go func(flowName string) {
			flow, err2 := c.GetFlow(flowName)
			if err2 != nil {
				err = err2
			}
			{
				flowMutex.Lock()
				defer flowMutex.Unlock()
				flows = append(flows, flow)
			}
			wg.Done()
		}(name)
	}
	wg.Wait()
	log.Printf("[core][appflow] Flows: %+v", flows)
	return flows, err
}

// Create a simple test flow to use in unit tests, to test other functions
// like starting, stopping, and deleting flows.
//
// This should not be used in production and therefore is not exported.
func (c Config) createTestFlow(flowName, bucketName string) error {
	connectorType := "S3"
	sourceConfig := appflowSdk.SourceFlowConfig{
		ConnectorType: &connectorType,
		SourceConnectorProperties: &appflowSdk.SourceConnectorProperties{
			S3: &appflowSdk.S3SourceProperties{
				BucketName: &bucketName,
			},
		},
	}
	destinationConfig := appflowSdk.DestinationFlowConfig{
		ConnectorType: &connectorType,
		DestinationConnectorProperties: &appflowSdk.DestinationConnectorProperties{
			S3: &appflowSdk.S3DestinationProperties{
				BucketName: &bucketName,
			},
		},
	}
	destinationConfigList := []*appflowSdk.DestinationFlowConfig{
		&destinationConfig,
	}
	field := "test-field"
	taskType := "Filter"
	task := appflowSdk.Task{
		SourceFields: []*string{&field},
		TaskType:     &taskType,
	}
	tasks := []*appflowSdk.Task{&task}
	triggerType := "Scheduled"
	pullMode := "Incremental"
	scheduleExpression := "rate(5minutes)"
	triggerSchedule := appflowSdk.ScheduledTriggerProperties{
		DataPullMode:       &pullMode,
		ScheduleExpression: &scheduleExpression,
	}
	input := appflowSdk.CreateFlowInput{
		FlowName:                  &flowName,
		SourceFlowConfig:          &sourceConfig,
		DestinationFlowConfigList: destinationConfigList,
		Tasks:                     tasks,
		TriggerConfig: &appflowSdk.TriggerConfig{
			TriggerType: &triggerType,
			TriggerProperties: &appflowSdk.TriggerProperties{
				Scheduled: &triggerSchedule,
			},
		},
	}

	_, err := c.Client.CreateFlow(&input)
	return err
}

func (c Config) StartFlow(name string) (FlowStatusOutput, error) {
	in := appflowSdk.StartFlowInput{
		FlowName: &name,
	}
	out, err := c.Client.StartFlow(&in)
	if err != nil {
		return FlowStatusOutput{}, err
	}
	status := FlowStatusOutput{
		ExecutionId: aws.StringValue(out.ExecutionId),
		FlowArn:     aws.StringValue(out.FlowArn),
		FlowStatus:  aws.StringValue(out.FlowStatus),
	}
	return status, nil
}

func (c Config) StopFlow(name string) (FlowStatusOutput, error) {
	in := appflowSdk.StopFlowInput{
		FlowName: &name,
	}
	out, err := c.Client.StopFlow(&in)
	if err != nil {
		return FlowStatusOutput{}, err
	}
	status := FlowStatusOutput{
		FlowArn:    aws.StringValue(out.FlowArn),
		FlowStatus: aws.StringValue(out.FlowStatus),
	}
	return status, nil
}

func (c Config) DeleteFlow(name string, forceDelete bool) error {
	input := appflowSdk.DeleteFlowInput{
		FlowName:    &name,
		ForceDelete: &forceDelete,
	}
	_, err := c.Client.DeleteFlow(&input)
	return err
}
