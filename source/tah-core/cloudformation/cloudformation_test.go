// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import (
	"context"
	"fmt"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/s3"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testTemplate = `{
    "AWSTemplateFormatVersion": "2010-09-09",
    "Description": "Minimal stack for testing",
    "Resources": {
        "TestBucket": {
            "Type": "AWS::S3::Bucket",
            "Properties": {
                "VersioningConfiguration": {
                    "Status": "Suspended"
                }
            }
        }
    },
    "Outputs": {
        "BucketName": {
            "Description": "Name of test bucket",
            "Value": {"Ref": "TestBucket"}
        }
    }
}`

func TestInit(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	ctx := context.Background()
	cfg, err := Init(ctx, envCfg.EnvName)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.cfnClient)
}

func TestDeploy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}

	ctx := context.Background()
	cfg, err := Init(ctx, envCfg.Region)
	assert.NoError(t, err)

	stackName := "test-stack-" + time.Now().Format("20060102-150405")
	s3c, err := s3.InitWithRandBucket("tah-core-cf", "test", envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("Could not initialize random bucket %+v", err)
	}
	t.Cleanup(func() {
		s3c.EmptyAndDelete()
	})
	errs := s3c.SaveMany([]string{"testCf.template"}, []string{testTemplate}, "application/json", 1)
	if errs != nil {
		t.Errorf("Could not save template %+v", errs)
	}

	params := map[string]string{}

	templateURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s/testCf.template",
		s3c.Bucket, envCfg.Region, s3c.Path)

	stackID, err := cfg.Deploy(ctx, stackName, templateURL, params)
	t.Cleanup(func() {
		cfg.DeleteStack(ctx, stackName)
	})
	assert.NoError(t, err)
	assert.Contains(t, stackID, stackName)

	// Test GetStackOutputs
	outputs, err := cfg.GetStackOutputs(ctx, stackName)
	assert.NoError(t, err)
	assert.NotEmpty(t, outputs)
}

func TestConvertToCloudFormationParameters(t *testing.T) {
	cfg := &CfConfig{}
	params := map[string]string{
		"Param1": "Value1",
		"Param2": "Value2",
	}

	cfParams := cfg.ConvertToCloudFormationParameters(params)
	assert.Len(t, cfParams, 2)

	paramMap := make(map[string]string)
	for _, param := range cfParams {
		paramMap[*param.ParameterKey] = *param.ParameterValue
	}

	assert.Equal(t, "Value1", paramMap["Param1"])
	assert.Equal(t, "Value2", paramMap["Param2"])
}

func TestDescribeStackNonexistent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}

	ctx := context.Background()
	cfg, err := Init(ctx, envCfg.Region)
	assert.NoError(t, err)

	stack, err := cfg.DescribeStack(ctx, "nonexistent-stack")
	assert.Error(t, err)
	assert.Nil(t, stack)
	assert.Contains(t, err.Error(), "does not exist")
}
