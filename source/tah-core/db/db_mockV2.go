// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"tah/upt/source/tah-core/core"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/mock"
)

type MockV2 struct {
	mock.Mock
}

var _ IDBConfig = &MockV2{}

func InitMockV2() *MockV2 {
	return &MockV2{}
}

// DeleteMany implements IDBConfig.
func (m *MockV2) DeleteMany(data interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

// FindStartingWith implements IDBConfig.
func (m *MockV2) FindStartingWith(pk string, value string, data interface{}) error {
	args := m.Called(pk, value, data)
	return args.Error(0)
}

// FindStartingWithAndFilter implements IDBConfig.
func (m *MockV2) FindStartingWithAndFilter(pk string, value string, data interface{}, filter DynamoFilterExpression) error {
	args := m.Called(pk, value, data, filter)
	return args.Error(0)
}

// FindStartingWithAndFilterWithIndex implements IDBConfig.
func (m *MockV2) FindStartingWithAndFilterWithIndex(pk string, value string, data interface{}, queryOptions QueryOptions) error {
	args := m.Called(pk, value, data, queryOptions)
	return args.Error(0)
}

// Get implements IDBConfig.
func (m *MockV2) Get(pk string, sk string, data interface{}) error {
	args := m.Called(pk, sk, data)
	return args.Error(0)
}

// GetItemCount implements IDBConfig.
func (m *MockV2) GetItemCount() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// Save implements IDBConfig.
func (m *MockV2) Save(prop interface{}) (interface{}, error) {
	args := m.Called(prop)
	return args.Get(0), args.Error(1)
}

// SaveMany implements IDBConfig.
func (m *MockV2) SaveMany(data interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

// SetTx implements IDBConfig.
func (m *MockV2) SetTx(tx core.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

// FindAll implements IDBConfig.
func (m *MockV2) FindAll(data interface{}, lastEvaluatedKey map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	args := m.Called(data, lastEvaluatedKey)
	return args.Get(0).(map[string]types.AttributeValue), args.Error(1)
}

// TableName implements IDBConfig.
func (m *MockV2) TableName() string {
	args := m.Called()
	return args.String(0)
}
