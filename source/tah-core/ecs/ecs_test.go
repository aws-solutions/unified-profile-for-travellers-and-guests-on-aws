// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package ecs

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecs "github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	awsecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/stretchr/testify/mock"
)

func TestRunTask(t *testing.T) {
	t.Parallel()

	taskDefinition := "my-task-definition"
	containerName := "my-container-name"
	clusterArn := "my-cluster-arn"
	eventId := "my-event-id"
	subnetIds := []string{"my-subnet"}
	securityGroupIds := []string{"my-sg"}

	taskConfig := TaskConfig{
		TaskDefinitionArn: taskDefinition,
		ContainerName:     containerName,
		ClusterArn:        clusterArn,
		SubnetIds:         subnetIds,
		SecurityGroupIds:  securityGroupIds,
	}

	cfg, mockEcs := InitMockEcs()

	domainName := "my-domain-name"
	objectTypeName := "my-object-type"
	envOverrides := []types.KeyValuePair{
		{Name: aws.String("DOMAIN_NAME"), Value: &domainName},
		{Name: aws.String("OBJECT_TYPE_NAME"), Value: &objectTypeName},
	}

	input := awsecs.RunTaskInput{
		Cluster:        &clusterArn,
		TaskDefinition: &taskDefinition,
		LaunchType:     awsecstypes.LaunchTypeFargate,
		NetworkConfiguration: &awsecstypes.NetworkConfiguration{
			AwsvpcConfiguration: &awsecstypes.AwsVpcConfiguration{
				Subnets:        subnetIds,
				SecurityGroups: securityGroupIds,
			},
		},
		Overrides: &awsecstypes.TaskOverride{
			ContainerOverrides: []awsecstypes.ContainerOverride{{
				Name:        &containerName,
				Environment: envOverrides,
			}},
		},
		StartedBy: &eventId,
	}

	mockEcs.On("RunTask", context.TODO(), &input).Return(&awsecs.RunTaskOutput{}, nil)

	_, err := cfg.RunTask(context.TODO(), taskConfig, envOverrides, eventId)
	if err != nil {
		t.Errorf("expect no error, got %v", err)
	}

	mockEcs.AssertExpectations(t)
}

func TestListTasks(t *testing.T) {
	t.Parallel()

	taskDefinition := "my-task-definition"
	containerName := "my-container-name"
	clusterArn := "my-cluster-arn"
	eventId := "my-event-id"
	subnetIds := []string{"my-subnet"}
	securityGroupIds := []string{"my-sg"}

	taskConfig := TaskConfig{
		TaskDefinitionArn: taskDefinition,
		ContainerName:     containerName,
		ClusterArn:        clusterArn,
		SubnetIds:         subnetIds,
		SecurityGroupIds:  securityGroupIds,
	}

	cfg, mockEcs := InitMockEcs()

	input := awsecs.ListTasksInput{
		Cluster:   &taskConfig.ClusterArn,
		StartedBy: &eventId,
	}
	mockEcs.On("ListTasks", context.TODO(), &input).Return(&awsecs.ListTasksOutput{}, nil)

	_, err := cfg.ListTasks(context.TODO(), taskConfig, &eventId, nil)
	if err != nil {
		t.Errorf("expect no error, got %v", err)
	}

	mockEcs.AssertExpectations(t)
}

func TestWaitForTasks(t *testing.T) {
	t.Parallel()

	taskDefinition := "my-task-definition"
	containerName := "my-container-name"
	clusterArn := "my-cluster-arn"
	subnetIds := []string{"my-subnet"}
	securityGroupIds := []string{"my-sg"}

	taskConfig := TaskConfig{
		TaskDefinitionArn: taskDefinition,
		ContainerName:     containerName,
		ClusterArn:        clusterArn,
		SubnetIds:         subnetIds,
		SecurityGroupIds:  securityGroupIds,
	}

	cfg, mockEcs := InitMockEcs()

	stopped := "STOPPED"
	task1 := types.Task{
		LastStatus: &stopped,
	}
	task2 := types.Task{
		LastStatus: &stopped,
	}

	taskArn1 := "arn1"
	taskArn2 := "arn2"
	taskArns := []string{taskArn1, taskArn2}
	maxTimeout := 10 * time.Second

	input := awsecs.DescribeTasksInput{
		Tasks:   taskArns,
		Cluster: &taskConfig.ClusterArn,
	}
	mockEcs.On("DescribeTasks", mock.Anything, &input).Return(&awsecs.DescribeTasksOutput{Tasks: []types.Task{task1, task2}}, nil)

	_, err := cfg.WaitForTasks(context.TODO(), taskConfig, taskArns, maxTimeout)
	if err != nil {
		t.Errorf("expect no error, got %v", err)
	}

	mockEcs.AssertExpectations(t)
}
