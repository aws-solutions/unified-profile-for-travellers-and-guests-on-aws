// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package ecs

import (
	"context"

	awsecs "github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/stretchr/testify/mock"
)

type MockEcs struct {
	mock.Mock
}

// static type check
var _ EcsApi = &MockEcs{}

func initMock() *MockEcs {
	return &MockEcs{}
}

func InitMockEcs() (*EcsConfig, *MockEcs) {
	mock := initMock()
	return &EcsConfig{svc: mock}, mock
}

func (m *MockEcs) RunTask(
	ctx context.Context,
	params *awsecs.RunTaskInput,
	optFns ...func(*awsecs.Options),
) (*awsecs.RunTaskOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*awsecs.RunTaskOutput), args.Error(1)
}

func (m *MockEcs) ListTasks(
	ctx context.Context,
	params *awsecs.ListTasksInput,
	optFns ...func(*awsecs.Options),
) (*awsecs.ListTasksOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*awsecs.ListTasksOutput), args.Error(1)
}

func (m *MockEcs) DescribeTasks(
	ctx context.Context,
	params *awsecs.DescribeTasksInput,
	optFns ...func(*awsecs.Options),
) (*awsecs.DescribeTasksOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*awsecs.DescribeTasksOutput), args.Error(1)
}
