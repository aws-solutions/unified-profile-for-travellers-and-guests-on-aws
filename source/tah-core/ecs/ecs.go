// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package ecs

import (
	"context"
	"time"

	mw "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	awsecs "github.com/aws/aws-sdk-go-v2/service/ecs"
	awsecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/smithy-go/middleware"
)

type EcsApi interface {
	RunTask(ctx context.Context, params *awsecs.RunTaskInput, optFns ...func(*awsecs.Options)) (*awsecs.RunTaskOutput, error)
	ListTasks(ctx context.Context, params *awsecs.ListTasksInput, optFns ...func(*awsecs.Options)) (*awsecs.ListTasksOutput, error)
	DescribeTasks(
		ctx context.Context,
		params *awsecs.DescribeTasksInput,
		optFns ...func(*awsecs.Options),
	) (*awsecs.DescribeTasksOutput, error)
}

// static type check
var _ EcsApi = &awsecs.Client{}

type EcsConfig struct {
	svc EcsApi
}

type TaskConfig struct {
	TaskDefinitionArn string
	ContainerName     string
	ClusterArn        string
	SubnetIds         []string
	SecurityGroupIds  []string
}

func InitEcs(ctx context.Context, region, solutionId, solutionVersion string) (*EcsConfig, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithAPIOptions([]func(*middleware.Stack) error{mw.AddUserAgentKey("AWSSOLUTION/" + solutionId + "/" + solutionVersion)}),
	)
	if err != nil {
		return &EcsConfig{}, err
	}
	svc := awsecs.NewFromConfig(cfg)
	return &EcsConfig{svc: svc}, nil
}

func (cfg *EcsConfig) RunTask(
	ctx context.Context,
	taskConfig TaskConfig,
	envOverrides []awsecstypes.KeyValuePair,
	startedBy string,
) (*awsecs.RunTaskOutput, error) {
	input := awsecs.RunTaskInput{
		Cluster:        &taskConfig.ClusterArn,
		TaskDefinition: &taskConfig.TaskDefinitionArn,
		LaunchType:     awsecstypes.LaunchTypeFargate,
		NetworkConfiguration: &awsecstypes.NetworkConfiguration{
			AwsvpcConfiguration: &awsecstypes.AwsVpcConfiguration{
				Subnets:        taskConfig.SubnetIds,
				SecurityGroups: taskConfig.SecurityGroupIds,
			},
		},
		Overrides: &awsecstypes.TaskOverride{
			ContainerOverrides: []awsecstypes.ContainerOverride{{
				Name:        &taskConfig.ContainerName,
				Environment: envOverrides,
			}},
		},
		StartedBy: &startedBy,
	}
	return cfg.svc.RunTask(ctx, &input)
}

func (cfg *EcsConfig) ListTasks(ctx context.Context, taskConfig TaskConfig, startedBy, desiredStatus *string) ([]string, error) {
	input := awsecs.ListTasksInput{
		Cluster:   &taskConfig.ClusterArn,
		StartedBy: startedBy,
	}
	if desiredStatus != nil {
		input.DesiredStatus = awsecstypes.DesiredStatus(*desiredStatus)
	}
	res, err := cfg.svc.ListTasks(ctx, &input)
	if err != nil {
		return []string{}, err
	}
	return res.TaskArns, nil
}

func (cfg *EcsConfig) WaitForTasks(
	ctx context.Context,
	taskConfig TaskConfig,
	taskArns []string,
	maxWaitDur time.Duration,
) (*awsecs.DescribeTasksOutput, error) {
	waiter := awsecs.NewTasksStoppedWaiter(cfg.svc)
	input := awsecs.DescribeTasksInput{
		Tasks:   taskArns,
		Cluster: &taskConfig.ClusterArn,
	}
	return waiter.WaitForOutput(ctx, &input, maxWaitDur)
}
