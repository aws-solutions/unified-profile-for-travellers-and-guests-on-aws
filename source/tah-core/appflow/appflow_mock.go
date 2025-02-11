// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package appflow

import "github.com/stretchr/testify/mock"

type MockConfig struct {
	mock.Mock
}

var _ IConfig = (*MockConfig)(nil)

func InitMock() *MockConfig {
	return &MockConfig{}
}

// DeleteFlow implements IConfig.
func (m *MockConfig) DeleteFlow(name string, forceDelete bool) error {
	args := m.Called(name, forceDelete)
	return args.Error(0)
}

// GetFlow implements IConfig.
func (m *MockConfig) GetFlow(flowName string) (Flow, error) {
	args := m.Called(flowName)
	return args.Get(0).(Flow), args.Error(1)
}

// GetFlows implements IConfig.
func (m *MockConfig) GetFlows(names []string) ([]Flow, error) {
	args := m.Called(names)
	return args.Get(0).([]Flow), args.Error(1)
}

// StartFlow implements IConfig.
func (m *MockConfig) StartFlow(name string) (FlowStatusOutput, error) {
	args := m.Called(name)
	return args.Get(0).(FlowStatusOutput), args.Error(1)
}

// StopFlow implements IConfig.
func (m *MockConfig) StopFlow(name string) (FlowStatusOutput, error) {
	args := m.Called(name)
	return args.Get(0).(FlowStatusOutput), args.Error(1)
}
