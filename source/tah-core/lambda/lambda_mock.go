// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lambda

import (
	"errors"
	core "tah/upt/source/tah-core/core"
)

type LambdaMock struct {
	Payload             interface{}
	EventSourceMappings []EventSourceMapping
	FunctionName        string
}

func InitMock(fName string) *LambdaMock {
	return &LambdaMock{
		EventSourceMappings: []EventSourceMapping{},
		FunctionName:        fName,
	}
}

func (l *LambdaMock) Invoke(payload interface{}, res core.JSONObject, async bool) (interface{}, error) {
	l.Payload = payload
	return nil, nil
}

func (l *LambdaMock) InvokeAsync(payload interface{}) (interface{}, error) {
	l.Payload = payload
	return nil, nil
}

func (l *LambdaMock) CreateFunction(name string, roleArn string, s3Bucket string, s3Key string, options LambdaOptions) error {
	return nil
}
func (l *LambdaMock) DeleteFunction(name string) error {
	return nil
}
func (l *LambdaMock) CreateDynamoEventSourceMapping(dynamoDbStreamArn string, functionName string) error {
	l.EventSourceMappings = append(l.EventSourceMappings, EventSourceMapping{
		StreamArn:    dynamoDbStreamArn,
		FunctionName: functionName,
		UUID:         core.UUID(),
	})
	return nil
}
func (l *LambdaMock) SearchDynamoEventSourceMappings(dynamoDbStreamArn string, functionName string) ([]EventSourceMapping, error) {
	return l.EventSourceMappings, nil
}
func (l *LambdaMock) DeleteDynamoEventSourceMapping(uuid string) error {
	newMappings := []EventSourceMapping{}
	hasDeletedSomething := false
	for _, mapping := range l.EventSourceMappings {
		if mapping.UUID != uuid {
			newMappings = append(newMappings, mapping)
		} else {
			hasDeletedSomething = true
		}
	}
	l.EventSourceMappings = newMappings
	if !hasDeletedSomething {
		return errors.New("mapping not found")
	}
	return nil
}

func (l *LambdaMock) Get(str string) string {
	if str == "FunctionName" {
		return l.FunctionName
	}
	return ""
}

func (l *LambdaMock) ListTags() (map[string]string, error) {
	return make(map[string]string), nil
}
