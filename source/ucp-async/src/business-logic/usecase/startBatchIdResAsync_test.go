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
	"time"

	awsecs "github.com/aws/aws-sdk-go-v2/service/ecs"
	awsecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/stretchr/testify/mock"
)

type MockSsm struct {
	mock.Mock
}

func (m *MockSsm) GetParametersByPath(ctx context.Context, namespace string) (map[string]string, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(map[string]string), args.Error(1)
}

type MockEcs struct {
	mock.Mock
}

func (m *MockEcs) RunTask(
	ctx context.Context,
	taskConfig ecs.TaskConfig,
	envOverrides []awsecstypes.KeyValuePair,
	startedBy string,
) (*awsecs.RunTaskOutput, error) {
	args := m.Called(ctx, taskConfig, envOverrides, startedBy)
	return args.Get(0).(*awsecs.RunTaskOutput), args.Error(1)
}

func TestStartBatchIdRes(t *testing.T) {
	t.Parallel()

	uc := InitStartBatchIdRes()
	if uc.Name() != "StartBatchIdRes" {
		t.Errorf("expect usecase name to be %v, got %v", "StartBatchIdRes", uc.Name())
	}

	expectedTaskConfig := ecs.TaskConfig{
		TaskDefinitionArn: "my-task-arn",
		ContainerName:     "my-container-name",
		ClusterArn:        "my-cluster-arn",
		SubnetIds:         []string{"my-subnet-a", "my-subnet-b"},
		SecurityGroupIds:  []string{"my-sg"},
	}

	mockSsm := MockSsm{}
	ssmParamNamespace := "/my-namespace/"
	mockSsm.On("GetParametersByPath", context.TODO(), ssmParamNamespace).Return(map[string]string{
		"VpcSubnets":             strings.Join(expectedTaskConfig.SubnetIds, ","),
		"VpcSecurityGroup":       strings.Join(expectedTaskConfig.SecurityGroupIds, ","),
		"IdResClusterArn":        expectedTaskConfig.ClusterArn,
		"IdResTaskDefinitionArn": expectedTaskConfig.TaskDefinitionArn,
		"IdResContainerName":     expectedTaskConfig.ContainerName,
	}, nil)

	eventId := "test-event-id"

	mockEcs := MockEcs{}
	mockEcs.On("RunTask", context.TODO(), expectedTaskConfig, mock.Anything, eventId).Return(&awsecs.RunTaskOutput{}, nil)

	domainName := "my_test_domain_" + time.Now().Format("20060102150405")
	accp := customerprofileslcs.InitMockV2()
	accp.On("SetDomain", domainName).Return(nil)
	accp.On("GetDomain").Return(customerprofileslcs.Domain{
		Name:      domainName,
		AllCounts: map[string]int64{"my-obj-type": 1000},
	}, nil)
	accp.On("GetActiveRuleSet").Return(customerprofileslcs.RuleSet{
		Name:          "my-rule-set",
		Rules:         []customerprofileslcs.Rule{{Index: 0, Name: "my-rule"}},
		LatestVersion: 3,
	}, nil)
	accp.On("InvalidateStitchingRuleSetsCache").Return(nil)
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
		Usecase:       async.USECASE_START_BATCH_ID_RES,
		TransactionID: "test-transaction-id",
		Body:          usecasemodel.StartBatchIdResBody{Domain: domainName},
	}

	err := uc.Execute(payload, services, core.NewTransaction(t.Name(), "", core.LogLevelDebug))
	if err != nil {
		t.Errorf("Error invoking usecase: %v", err)
	}

	mockSsm.AssertExpectations(t)
	mockEcs.AssertExpectations(t)
}
