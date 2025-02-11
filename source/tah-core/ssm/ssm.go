// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package ssm

import (
	"context"
	"strings"

	mw "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go/middleware"
)

type SsmApi interface {
	GetParametersByPath(
		ctx context.Context,
		params *awsssm.GetParametersByPathInput,
		optFns ...func(*awsssm.Options),
	) (*awsssm.GetParametersByPathOutput, error)
}

// static type check
var _ SsmApi = &awsssm.Client{}

type SsmConfig struct {
	svc SsmApi
}

func InitSsm(ctx context.Context, region, solutionId, solutionVersion string) (*SsmConfig, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithAPIOptions([]func(*middleware.Stack) error{mw.AddUserAgentKey("AWSSOLUTION/" + solutionId + "/" + solutionVersion)}),
	)
	if err != nil {
		return &SsmConfig{}, err
	}
	svc := awsssm.NewFromConfig(cfg)
	return &SsmConfig{svc: svc}, nil
}

func (c *SsmConfig) GetParametersByPath(ctx context.Context, namespace string) (map[string]string, error) {
	paginator := awsssm.NewGetParametersByPathPaginator(
		c.svc,
		&awsssm.GetParametersByPathInput{Path: &namespace, Recursive: aws.Bool(true)},
	)
	results := map[string]string{}
	var allErrors error
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if page != nil {
			for _, parameter := range page.Parameters {
				if parameter.Name != nil && parameter.Value != nil {
					results[strings.TrimPrefix(*parameter.Name, namespace)] = *parameter.Value
				}
			}
		}
		if err != nil {
			return map[string]string{}, err
		}
	}
	return results, allErrors
}
