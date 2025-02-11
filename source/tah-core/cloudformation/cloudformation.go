// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go/aws"
)

type CfConfig struct {
	cfnClient *cloudformation.Client
}

func Init(ctx context.Context, region string) (*CfConfig, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	return &CfConfig{
		cfnClient: cloudformation.NewFromConfig(cfg),
	}, nil
}

func (op *CfConfig) Deploy(ctx context.Context, stackName string, templateUrl string, parameters map[string]string) (string, error) {
	stack, err := op.DescribeStack(ctx, stackName)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return op.CreateStack(ctx, stackName, templateUrl, parameters)
		}
		return "", fmt.Errorf("failed to describe stack: %v", err)
	}

	if stack.StackStatus == types.StackStatusCreateComplete ||
		stack.StackStatus == types.StackStatusUpdateComplete ||
		stack.StackStatus == types.StackStatusUpdateRollbackComplete {
		return op.UpdateStack(ctx, stackName, templateUrl, parameters)
	}

	return "", fmt.Errorf("stack %s is in %s state", stackName, stack.StackStatus)
}

func (op *CfConfig) DescribeStack(ctx context.Context, stackName string) (*types.Stack, error) {
	resp, err := op.cfnClient.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: &stackName,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Stacks) == 0 {
		return nil, fmt.Errorf("no stack found with name %s", stackName)
	}
	return &resp.Stacks[0], nil
}

func (op *CfConfig) CreateStack(ctx context.Context, stackName string, templateUrl string, parameters map[string]string) (string, error) {
	resp, err := op.cfnClient.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:   &stackName,
		TemplateURL: &templateUrl,
		Parameters:  op.ConvertToCloudFormationParameters(parameters),
		Capabilities: []types.Capability{
			types.CapabilityCapabilityIam,
			types.CapabilityCapabilityNamedIam,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create stack: %v", err)
	}

	waiter := cloudformation.NewStackCreateCompleteWaiter(op.cfnClient)
	err = waiter.Wait(ctx, &cloudformation.DescribeStacksInput{
		StackName: &stackName,
	}, time.Duration(30)*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed waiting for stack creation: %v", err)
	}

	return *resp.StackId, nil
}

func (op *CfConfig) UpdateStack(ctx context.Context, stackName string, templateUrl string, parameters map[string]string) (string, error) {
	log.Printf("Updating stack %s with template %s and parameters %v", stackName, templateUrl, parameters)
	resp, err := op.cfnClient.UpdateStack(ctx, &cloudformation.UpdateStackInput{
		StackName:   &stackName,
		TemplateURL: &templateUrl,
		Parameters:  op.ConvertToCloudFormationParameters(parameters),
		Capabilities: []types.Capability{
			types.CapabilityCapabilityIam,
			types.CapabilityCapabilityNamedIam,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "No updates are to be performed") {
			return stackName, nil
		}
		return "", fmt.Errorf("failed to update stack: %v", err)
	}

	waiter := cloudformation.NewStackUpdateCompleteWaiter(op.cfnClient)
	err = waiter.Wait(ctx, &cloudformation.DescribeStacksInput{
		StackName: &stackName,
	}, time.Duration(30)*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed waiting for stack update: %v", err)
	}

	return *resp.StackId, nil
}

func (op *CfConfig) DeleteStack(ctx context.Context, stackName string) error {
	_, err := op.cfnClient.DeleteStack(ctx, &cloudformation.DeleteStackInput{
		StackName: &stackName,
	})
	if err != nil {
		return fmt.Errorf("failed to initiate stack deletion: %v", err)
	}

	waiter := cloudformation.NewStackDeleteCompleteWaiter(op.cfnClient)
	err = waiter.Wait(ctx, &cloudformation.DescribeStacksInput{
		StackName: &stackName,
	}, time.Duration(30)*time.Minute)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return nil
		}
		return fmt.Errorf("failed waiting for stack deletion: %v", err)
	}

	return nil
}

func (op *CfConfig) GetStackOutputs(ctx context.Context, stackName string) (map[string]string, error) {
	stack, err := op.DescribeStack(ctx, stackName)
	if err != nil {
		return nil, err
	}

	outputs := make(map[string]string)
	for _, output := range stack.Outputs {
		outputs[*output.OutputKey] = *output.OutputValue
	}
	return outputs, nil
}

func (op *CfConfig) ConvertToCloudFormationParameters(params map[string]string) []types.Parameter {
	var cfnParams []types.Parameter
	for key, value := range params {
		cfnParams = append(cfnParams, types.Parameter{
			ParameterKey:   aws.String(key),
			ParameterValue: aws.String(value),
		})
	}
	return cfnParams
}
