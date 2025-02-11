// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package ssm

import (
	"context"
	"testing"

	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go/aws"
)

func TestGetParametersByPath(t *testing.T) {
	t.Parallel()

	ssmConfig, mock := InitMockSsm()

	namespace := "/my-namespace/"
	input := awsssm.GetParametersByPathInput{Path: &namespace, Recursive: aws.Bool(true)}
	mock.On("GetParametersByPath", context.TODO(), &input).
		Return(&awsssm.GetParametersByPathOutput{Parameters: []types.Parameter{
			{Name: aws.String(namespace + "foo"), Value: aws.String("foo-value")},
			{Name: aws.String(namespace + "bar"), Value: aws.String("bar-value")}}},
			nil)

	results, err := ssmConfig.GetParametersByPath(context.TODO(), namespace)
	if err != nil {
		t.Fatalf("expect no error, got: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expect 2 keys, got: %d", len(results))
	}
	if results["foo"] != "foo-value" {
		t.Errorf("expect foo-value, got: %s", results["foo"])
	}
	if results["bar"] != "bar-value" {
		t.Errorf("expect bar-value, got: %s", results["bar"])
	}

	mock.AssertExpectations(t)
}
