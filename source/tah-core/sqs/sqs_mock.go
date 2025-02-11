// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package sqs

import "github.com/stretchr/testify/mock"

type MockConfig struct {
	mock.Mock
}

var _ IConfig = &MockConfig{}

func InitMock() *MockConfig {
	return &MockConfig{}
}

// Create implements IConfig.
func (m *MockConfig) Create(queueName string) (string, error) {
	args := m.Called(queueName)
	return args.String(0), args.Error(1)
}

// CreateRandom implements IConfig.
func (m *MockConfig) CreateRandom(prefix string) (string, error) {
	args := m.Called(prefix)
	return args.String(0), args.Error(1)
}

// Delete implements IConfig.
func (m *MockConfig) Delete() error {
	args := m.Called()
	return args.Error(0)
}

// DeleteByName implements IConfig.
func (m *MockConfig) DeleteByName(queueName string) error {
	args := m.Called(queueName)
	return args.Error(0)
}

// Get implements IConfig.
func (m *MockConfig) Get(options GetMessageOptions) (QueuePeekContent, error) {
	args := m.Called(options)
	return args.Get(0).(QueuePeekContent), args.Error(1)
}

// GetAttributes implements IConfig.
func (m *MockConfig) GetAttributes() (map[string]string, error) {
	args := m.Called()
	return args.Get(0).(map[string]string), args.Error(1)
}

// GetQueueArn implements IConfig.
func (m *MockConfig) GetQueueArn() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// GetQueueUrl implements IConfig.
func (m *MockConfig) GetQueueUrl(name string) (string, error) {
	args := m.Called(name)
	return args.String(0), args.Error(1)
}

// SendWithStringAttributes implements IConfig.
func (m *MockConfig) SendWithStringAttributes(msg string, msgAttr map[string]string) error {
	args := m.Called(msg, msgAttr)
	return args.Error(0)
}

func (m *MockConfig) SendWithStringAttributesAndDelay(msg string, msgAttr map[string]string, delaySeconds int) error {
	args := m.Called(msg, msgAttr)
	return args.Error(0)
}

// SetAttribute implements IConfig.
func (m *MockConfig) SetAttribute(key string, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

// SetPolicy implements IConfig.
func (m *MockConfig) SetPolicy(principalType string, principal string, actions []string) error {
	args := m.Called(principalType, principal, actions)
	return args.Error(0)
}
