// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/ecs"
	"tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"

	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/aws/aws-sdk-go/aws"
)

type RebuildCacheUsecase struct {
	name string
}

// Static Interface Check
var _ Usecase = (*RebuildCacheUsecase)(nil)

func InitRebuildCache() *RebuildCacheUsecase {
	return &RebuildCacheUsecase{
		name: "RebuildCacheAsync",
	}
}

// Name implements Usecase.
func (u *RebuildCacheUsecase) Name() string {
	return u.name
}

// Execute implements Usecase.
func (u *RebuildCacheUsecase) Execute(payload async.AsyncInvokePayload, services async.Services, tx core.Transaction) error {
	payloadJson, err := json.Marshal(payload.Body)
	if err != nil {
		tx.Error("[%s] Error marshalling payload %v", u.name, err)
	}

	// Convert json to RebuildCacheBody
	data := uc.RebuildCacheBody{}
	err = json.Unmarshal(payloadJson, &data)
	if err != nil {
		tx.Error("[%s] Error unmarshalling payload body %v", u.name, err)
		return err
	}
	tx.Debug("[%s] Payload body: %+v", u.name, data)

	namespace := services.Env["SSM_PARAM_NAMESPACE"]
	tx.Debug("[%s] Retrieving SSM Parameters under namespace: %v", u.name, namespace)
	params, err := services.SsmConfig.GetParametersByPath(context.TODO(), namespace)
	if err != nil {
		tx.Error("error querying SSM parameters: %v", err)
		return err
	}
	tx.Debug("[%s] Retrieved SSM Parameters: %v", u.name, params)
	taskConfig, err := getTaskConfigForRebuildCache(params)
	if err != nil {
		tx.Error("[%s] Error loading ECS task config for rebuild cache: %v", u.name, err)
		return err
	}

	domainName := data.DomainName
	services.AccpConfig.SetDomain(domainName)
	services.AccpConfig.SetTx(tx)

	cacheMode := data.CacheMode
	refreshCp := cacheMode&customerprofileslcs.CUSTOMER_PROFILES_MODE == customerprofileslcs.CUSTOMER_PROFILES_MODE
	refreshDynamo := cacheMode&customerprofileslcs.DYNAMO_MODE == customerprofileslcs.DYNAMO_MODE

	// Get profile count
	domain, err := services.AccpConfig.GetDomain()
	if err != nil {
		return fmt.Errorf("error retrieving domain")
	}
	tx.Debug("[%s] Domain Config: %v", u.name, domain)
	profileCount := domain.NProfiles
	tx.Info("%v profiles found. Calculating partitioning", profileCount)

	// Partitioning based on count
	partitionBits := int(min(4, max(0, profileCount/1e4-1)))
	tx.Info("[%s] Partition Bits: %v", u.name, partitionBits)

	var allErrors error
	if refreshCp {
		tx.Info("[%s] Refreshing Customer Profiles Cache", u.name)
		// Clear CP Cache
		tags := map[string]string{"envName": data.Env, "aws_solution": "SO0244"}
		err = services.AccpConfig.ClearCustomerProfileCache(tags, data.MatchBucketName)
		if err != nil {
			return fmt.Errorf("error clearing customer profiles cache: %v", err)
		}

		getPartition, err := customerprofileslcs.PartitionUuid4Space(partitionBits)
		if err != nil {
			return fmt.Errorf("error partitioning for customer profiles cache: %v", err)
		}

		err = startEcsTask(context.TODO(), getPartition, "CUSTOMER_PROFILES", domainName, payload.EventID, services.EcsConfig, taskConfig)
		if err != nil {
			allErrors = errors.Join(allErrors, err)
		}
	}

	if refreshDynamo {
		tx.Info("[%s] Refreshing DynamoDB Cache", u.name)
		// Clear Dynamo Cache
		err := services.AccpConfig.ClearDynamoCache()
		if err != nil {
			return fmt.Errorf("error clearing dynamo cache: %v", err)
		}

		getPartition, err := customerprofileslcs.PartitionUuid4Space(partitionBits)
		if err != nil {
			return fmt.Errorf("error partitioning for dynamo cache: %v", err)
		}
		err = startEcsTask(context.TODO(), getPartition, "DYNAMO", domainName, payload.EventID, services.EcsConfig, taskConfig)
		if err != nil {
			allErrors = errors.Join(allErrors, err)
		}
	}
	return allErrors
}

func getTaskConfigForRebuildCache(params map[string]string) (ecs.TaskConfig, error) {
	param, ok := params["VpcSubnets"]
	if !ok || param == "" {
		err := errors.New("no value for VpcSubnets")
		return ecs.TaskConfig{}, err
	}
	subnetIds := strings.Split(param, ",")

	param, ok = params["VpcSecurityGroup"]
	if !ok || param == "" {
		err := errors.New("no value for VpcSecurityGroup")
		return ecs.TaskConfig{}, err
	}
	securityGroups := []string{param}

	param, ok = params["RebuildCacheClusterArn"]
	if !ok || param == "" {
		err := errors.New("no value for RebuildCacheClusterArn")
		return ecs.TaskConfig{}, err
	}
	clusterArn := param

	param, ok = params["RebuildCacheTaskDefinitionArn"]
	if !ok || param == "" {
		err := errors.New("no value for RebuildCacheTaskDefinitionArn")
		return ecs.TaskConfig{}, err
	}
	taskDefinitionArn := param

	param, ok = params["RebuildCacheContainerName"]
	if !ok || param == "" {
		err := errors.New("no value for RebuildCacheContainerName")
		return ecs.TaskConfig{}, err
	}
	containerName := param

	return ecs.TaskConfig{
		ClusterArn:        clusterArn,
		TaskDefinitionArn: taskDefinitionArn,
		ContainerName:     containerName,
		SubnetIds:         subnetIds,
		SecurityGroupIds:  securityGroups,
	}, nil
}

func startEcsTask(
	ctx context.Context,
	getPartition func() (customerprofileslcs.UuidPartition, bool, error),
	cacheType, domainName, eventId string,
	ecsCfg customerprofileslcs.EcsApi,
	taskConfig ecs.TaskConfig,
) error {
	var allErrors error
	fmt.Printf("Attempting to start ECS task for domain: %s, cacheType: %s, eventId: %s \n", domainName, cacheType, eventId)
	for partition, ok, err := getPartition(); ok && err == nil; partition, ok, err = getPartition() {
		envOverrides := []ecsTypes.KeyValuePair{
			{Name: aws.String("DOMAIN_NAME"), Value: &domainName},
			{Name: aws.String("CACHE_TYPE"), Value: aws.String(cacheType)},
			{Name: aws.String("PARTITION_LOWER_BOUND"), Value: aws.String(partition.LowerBound.String())},
			{Name: aws.String("PARTITION_UPPER_BOUND"), Value: aws.String(partition.UpperBound.String())},
		}
		response, err := ecsCfg.RunTask(ctx, taskConfig, envOverrides, eventId)
		if err != nil {
			fmt.Printf("Error running task: %s \n", err)
			allErrors = errors.Join(allErrors, err)
		}
		if response != nil {
			for _, failure := range response.Failures {
				arn := "Unknown ARN"
				if failure.Arn != nil {
					arn = *failure.Arn
				}
				reason := "Unknown reason"
				if failure.Reason != nil {
					reason = *failure.Reason
				}
				detail := "No detail"
				if failure.Detail != nil {
					detail = *failure.Detail
				}
				allErrors = errors.Join(
					allErrors,
					fmt.Errorf("error running task on %v (%v): %v", arn, reason, detail),
				)
			}
		}
	}
	return allErrors
}
