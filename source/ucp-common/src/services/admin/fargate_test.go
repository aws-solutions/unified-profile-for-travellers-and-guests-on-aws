// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"testing"
)

func TestFargate(t *testing.T) {
	taskConfig, err := GetTaskConfigForIdRes(map[string]string{
		"VpcSubnets":             "test11,test12",
		"VpcSecurityGroup":       "test2",
		"IdResClusterArn":        "test3",
		"IdResTaskDefinitionArn": "test4",
		"IdResContainerName":     "test5",
	})

	if err != nil {
		t.Fatalf("error parsing task config map: %v", err)
	}
	if taskConfig.TaskDefinitionArn != "test4" {
		t.Fatalf("expected TaskDefinitionArn to be %s, got %s", "test1", taskConfig.TaskDefinitionArn)
	}
	if taskConfig.ContainerName != "test5" {
		t.Fatalf("expected ContainerName to be %s, got %s", "test5", taskConfig.ContainerName)
	}
	if taskConfig.ClusterArn != "test3" {
		t.Fatalf("expected ClusterArn to be %s, got %s", "test3", taskConfig.ClusterArn)
	}
	if len(taskConfig.SubnetIds) != 2 {
		t.Fatalf("expected 2 SubnetIds, got %d", len(taskConfig.SubnetIds))
	}
	if taskConfig.SubnetIds[0] != "test11" {
		t.Fatalf("expected SubnetIds to be %s, got %s", "test11", taskConfig.SubnetIds[0])
	}
	if taskConfig.SubnetIds[1] != "test12" {
		t.Fatalf("expected SubnetIds to be %s, got %s", "test12", taskConfig.SubnetIds[1])
	}
	if len(taskConfig.SecurityGroupIds) != 1 {
		t.Fatalf("expected 2 SecurityGroupIds, got %d", len(taskConfig.SecurityGroupIds))
	}
	if taskConfig.SecurityGroupIds[0] != "test2" {
		t.Fatalf("expected SecurityGroupIds to be %s, got %s", "test21", taskConfig.SecurityGroupIds[0])
	}
}
