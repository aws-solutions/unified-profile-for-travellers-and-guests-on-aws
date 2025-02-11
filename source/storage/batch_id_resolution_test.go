// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	"context"
	"strconv"
	"testing"
	"time"

	"tah/upt/source/tah-core/ecs"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecs "github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/google/uuid"
)

func TestStartRuleBasedJob(t *testing.T) {
	t.Parallel()

	taskDefinition := "my-task-definition"
	containerName := "my-container-name"
	domainName := "my-domain-name"
	objectTypeName := "my-object-type"
	clusterArn := "my-cluster-arn"
	subnets := []string{"my-subnet-a", "my-subnet-b"}
	sgs := []string{"my-sg"}
	taskConfig := ecs.TaskConfig{
		TaskDefinitionArn: taskDefinition,
		ContainerName:     containerName,
		ClusterArn:        clusterArn,
		SubnetIds:         subnets,
		SecurityGroupIds:  sgs,
	}
	eventId := "my-event-id"
	objectCounts := map[string]int64{
		objectTypeName:          2e6,
		countTableProfileObject: 9e16,
	}
	ruleSet := RuleSet{
		Name:                    "my-rule-set",
		Rules:                   []Rule{{Index: 0, Name: "my-rule", Description: "my rule description", Conditions: []Condition{}}},
		LatestVersion:           4,
		LastActivationTimestamp: time.Now(),
	}
	expectedPartitions := []UuidPartition{
		{
			LowerBound: uuid.MustParse("00000000-0000-4000-8000-000000000000"),
			UpperBound: uuid.MustParse("7FFFFFFF-FFFF-4FFF-BFFF-FFFFFFFFFFFF"),
		},
		{
			LowerBound: uuid.MustParse("80000000-0000-4000-8000-000000000000"),
			UpperBound: uuid.MustParse("FFFFFFFF-FFFF-4FFF-BFFF-FFFFFFFFFFFF"),
		},
	}
	cfg, mockEcs := ecs.InitMockEcs()
	for _, partition := range expectedPartitions {
		mockEcs.On("RunTask", context.Background(), &awsecs.RunTaskInput{
			Cluster:        &clusterArn,
			TaskDefinition: &taskDefinition,
			LaunchType:     types.LaunchTypeFargate,
			NetworkConfiguration: &types.NetworkConfiguration{
				AwsvpcConfiguration: &types.AwsVpcConfiguration{
					Subnets:        subnets,
					SecurityGroups: sgs,
				},
			},
			Overrides: &types.TaskOverride{ContainerOverrides: []types.ContainerOverride{{
				Name: &containerName,
				Environment: []types.KeyValuePair{
					{Name: aws.String("DOMAIN_NAME"), Value: &domainName},
					{Name: aws.String("OBJECT_TYPE_NAME"), Value: &objectTypeName},
					{Name: aws.String("RULE_SET_VERSION"), Value: aws.String(strconv.Itoa(ruleSet.LatestVersion))},
					{Name: aws.String("RULE_ID"), Value: aws.String(strconv.Itoa(0))},
					{Name: aws.String("PARTITION_LOWER_BOUND"), Value: aws.String(partition.LowerBound.String())},
					{Name: aws.String("PARTITION_UPPER_BOUND"), Value: aws.String(partition.UpperBound.String())},
				}}},
			},
			StartedBy: &eventId}).Return(&awsecs.RunTaskOutput{}, nil)
	}

	err := StartRuleBasedJob(context.Background(), cfg, taskConfig, domainName, objectCounts, ruleSet, eventId)
	if err != nil {
		t.Errorf("expect no error, got %v", err)
	}

	mockEcs.AssertExpectations(t)
	mockEcs.AssertNumberOfCalls(t, "RunTask", 2)
}

func TestPartitionUuidSpace(t *testing.T) {
	t.Parallel()

	it, err := PartitionUuid4Space(0)
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}
	var partitions []UuidPartition
	for part, ok, err := it(); ok && err == nil; part, ok, err = it() {
		partitions = append(partitions, part)
	}
	if err != nil {
		t.Fatalf("expect no error generating partitions, got %v", err)
	}
	expected := []UuidPartition{
		{
			LowerBound: uuid.MustParse("00000000-0000-4000-8000-000000000000"),
			UpperBound: uuid.MustParse("FFFFFFFF-FFFF-4FFF-BFFF-FFFFFFFFFFFF"),
		},
	}
	assertMatch(t, expected, partitions)

	it, err = PartitionUuid4Space(1)
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}
	partitions = []UuidPartition{}
	for part, ok, err := it(); ok && err == nil; part, ok, err = it() {
		partitions = append(partitions, part)
	}
	if err != nil {
		t.Fatalf("expect no error generating partitions, got %v", err)
	}
	expected = []UuidPartition{
		{
			LowerBound: uuid.MustParse("00000000-0000-4000-8000-000000000000"),
			UpperBound: uuid.MustParse("7FFFFFFF-FFFF-4FFF-BFFF-FFFFFFFFFFFF"),
		},
		{
			LowerBound: uuid.MustParse("80000000-0000-4000-8000-000000000000"),
			UpperBound: uuid.MustParse("FFFFFFFF-FFFF-4FFF-BFFF-FFFFFFFFFFFF"),
		},
	}
	assertMatch(t, expected, partitions)

	bits := 8
	it, err = PartitionUuid4Space(bits)
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}
	expectedPartitions := 1 << bits
	var numPartitions int
	for _, ok, err := it(); ok && err == nil; _, ok, err = it() {
		numPartitions++
	}
	if numPartitions != expectedPartitions {
		t.Errorf("expect %v partitions, got %v", expectedPartitions, numPartitions)
	}
}

func assertMatch(t *testing.T, expected, actual []UuidPartition) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expect %v partition, got %v", len(expected), len(actual))
	}
	for i := 0; i < len(actual); i++ {
		if actual[i].LowerBound != expected[i].LowerBound {
			t.Errorf("expect lower bound to be %v, got %v", expected[i].LowerBound, actual[i].LowerBound)
		}
		if actual[i].UpperBound != expected[i].UpperBound {
			t.Errorf("expect upper bound to be %v, got %v", expected[i].UpperBound, actual[i].UpperBound)
		}
	}
}
