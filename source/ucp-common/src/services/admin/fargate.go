// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package admin

import (
	"errors"
	"strings"
	"tah/upt/source/tah-core/ecs"
)

func GetTaskConfigForIdRes(params map[string]string) (ecs.TaskConfig, error) {
	param, ok := params["VpcSubnets"]
	if !ok {
		err := errors.New("no value for VpcSubnets")
		return ecs.TaskConfig{}, err
	}
	subnetIds := strings.Split(param, ",")

	param, ok = params["VpcSecurityGroup"]
	if !ok {
		err := errors.New("no value for VpcSecurityGroup")
		return ecs.TaskConfig{}, err
	}
	securityGroups := []string{param}

	param, ok = params["IdResClusterArn"]
	if !ok {
		err := errors.New("no value for IdResClusterArn")
		return ecs.TaskConfig{}, err
	}
	clusterArn := param

	param, ok = params["IdResTaskDefinitionArn"]
	if !ok {
		err := errors.New("no value for IdResTaskDefinitionArn")
		return ecs.TaskConfig{}, err
	}
	taskDefinitionArn := param

	param, ok = params["IdResContainerName"]
	if !ok {
		err := errors.New("no value for IdResContainerName")
		return ecs.TaskConfig{}, err
	}
	containerName := param

	return ecs.TaskConfig{
		TaskDefinitionArn: taskDefinitionArn,
		ContainerName:     containerName,
		ClusterArn:        clusterArn,
		SubnetIds:         subnetIds,
		SecurityGroupIds:  securityGroups,
	}, nil
}
