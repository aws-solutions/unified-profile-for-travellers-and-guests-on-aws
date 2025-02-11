// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	core "tah/upt/source/tah-core/core"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type MockDBConfig struct {
}

func (c MockDBConfig) SetTx(tx core.Transaction) error {
	return nil
}
func (c MockDBConfig) Save(prop interface{}) (interface{}, error) {
	return prop, nil
}
func (c MockDBConfig) FindStartingWith(pk string, value string, data interface{}) error {
	return nil
}
func (c MockDBConfig) DeleteMany(data interface{}) error {
	return nil
}
func (c MockDBConfig) SaveMany(data interface{}) error {
	return nil
}
func (c MockDBConfig) FindAll(data interface{}, lastEvaluatedKey map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	return nil, nil
}
func (c MockDBConfig) Get(pk string, sk string, data interface{}) error {
	return nil
}
