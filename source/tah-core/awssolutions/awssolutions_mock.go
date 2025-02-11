// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package awssolutions

type MockConfig struct {
	Metrics map[string]interface{}
}

func InitMock() MockConfig {
	return MockConfig{
		Metrics: make(map[string]interface{}),
	}
}

func (c MockConfig) SendMetrics(data map[string]interface{}) (bool, error) {
	for k, v := range data {
		c.Metrics[k] = v
	}
	return true, nil
}
