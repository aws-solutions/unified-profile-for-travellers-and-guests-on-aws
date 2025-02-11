// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package usecase

import (
	"context"
	"strings"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/ecs"
	"tah/upt/source/ucp-common/src/model/async"
	usecasemodel "tah/upt/source/ucp-common/src/model/async/usecase"
	"testing"

	awsecs "github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/stretchr/testify/mock"
)

func TestRebuildCache(t *testing.T) {
	t.Parallel()

	uc := InitRebuildCache()
	if uc.Name() != "RebuildCacheAsync" {
		t.Errorf("expect usecase name to be %v, got %v", "RebuildCacheAsync", uc.Name())
	}

	expectedTaskConfig := ecs.TaskConfig{
		TaskDefinitionArn: "my-task-arn",
		ContainerName:     "my-container-name",
		ClusterArn:        "my-cluster-arn",
		SubnetIds:         []string{"my-subnet-a", "my-subnet-b"},
		SecurityGroupIds:  []string{"my-sg"},
	}

	mockSsm := MockSsm{}
	ssmParamNamespace := "/rebuild-cache-namespace/"
	mockSsm.On("GetParametersByPath", context.TODO(), ssmParamNamespace).Return(map[string]string{
		"VpcSubnets":                    strings.Join(expectedTaskConfig.SubnetIds, ","),
		"VpcSecurityGroup":              strings.Join(expectedTaskConfig.SecurityGroupIds, ","),
		"RebuildCacheClusterArn":        expectedTaskConfig.ClusterArn,
		"RebuildCacheTaskDefinitionArn": expectedTaskConfig.TaskDefinitionArn,
		"RebuildCacheContainerName":     expectedTaskConfig.ContainerName,
	}, nil)

	eventId := "test-event-id"

	mockEcs := MockEcs{}
	mockEcs.On("RunTask", context.TODO(), expectedTaskConfig, mock.Anything, eventId).Return(&awsecs.RunTaskOutput{}, nil)

	domainName := "rebuild_cache_async_" + strings.ToLower(core.GenerateUniqueId())
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	env := "test-env"
	tags := map[string]string{"envName": env, "aws_solution": "SO0244"}
	matchBucketName := "test-bucket-name"
	accp := customerprofileslcs.InitMockV2()
	accp.On("SetDomain", domainName).Return(nil)
	accp.On("SetTx", tx).Return()
	accp.On("GetDomain").Return(customerprofileslcs.Domain{
		Name:      domainName,
		NProfiles: 2000,
	}, nil)
	accp.On("ClearCustomerProfileCache", tags, matchBucketName).Return(nil)
	accp.On("ClearDynamoCache").Return(nil)
	solutionsConfig := awssolutions.InitMock()
	services := async.Services{
		AccpConfig:      accp,
		SolutionsConfig: solutionsConfig,
		SsmConfig:       &mockSsm,
		EcsConfig:       &mockEcs,
		Env:             map[string]string{"SSM_PARAM_NAMESPACE": ssmParamNamespace},
	}

	payload := async.AsyncInvokePayload{
		EventID:       eventId,
		Usecase:       async.USECASE_REBUILD_CACHE,
		TransactionID: "test-transaction-id",
		Body: usecasemodel.RebuildCacheBody{
			MatchBucketName: matchBucketName,
			DomainName:      domainName,
			Env:             "test-env",
			CacheMode:       customerprofileslcs.CacheMode(3),
		},
	}

	err := uc.Execute(payload, services, tx)
	if err != nil {
		t.Errorf("Error invoking usecase: %v", err)
	}

	mockSsm.AssertExpectations(t)
	mockEcs.AssertExpectations(t)
}
