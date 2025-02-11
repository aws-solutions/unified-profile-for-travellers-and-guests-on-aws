// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package ssm

import (
	"context"

	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/stretchr/testify/mock"
)

type MockSsm struct {
	mock.Mock
}

// static type check
var _ SsmApi = &MockSsm{}

func InitMockSsm() (*SsmConfig, *MockSsm) {
	mock := &MockSsm{}
	return &SsmConfig{svc: mock}, mock
}

func (m *MockSsm) GetParametersByPath(
	ctx context.Context,
	params *awsssm.GetParametersByPathInput,
	optFns ...func(*awsssm.Options),
) (*awsssm.GetParametersByPathOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*awsssm.GetParametersByPathOutput), args.Error(1)
}
