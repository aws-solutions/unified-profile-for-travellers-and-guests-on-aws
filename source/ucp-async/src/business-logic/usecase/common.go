// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"errors"
	"fmt"
	"tah/upt/source/tah-core/core"
	common "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/async"
)

const INITIAL_CLAUDE_PROMPT = "The following json records contain information about a traveler including name, job, date of birth, address, loyalty profiles, Air and hotel bookings, web and mobile interactions and customer service conversation. Summarize this traveler profile in 150 words while keeping their name accurate and highlighting the largest purchases, far destinations, high loyalty points or status or anything else that would be outstanding for an average traveler."

var appAccessRoleMap = map[string]common.AppPermission{
	"domainSettingsRole/": common.ConfigureGenAiPermission | common.RebuildCachePermission | common.IndustryConnectorPermission,
	"profileManagerRole/": common.SearchProfilePermission | common.DeleteProfilePermission | common.MergeProfilePermission | common.UnmergeProfilePermission,
	"ruleSetManagerRole/": common.ActivateRuleSetPermission | common.SaveRuleSetPermission | common.ListRuleSetPermission,
	"gdprManagerRole/":    common.ListPrivacySearchPermission | common.GetPrivacySearchPermission | common.CreatePrivacySearchPermission | common.DeletePrivacySearchPermission | common.PrivacyDataPurgePermission,
	"adminRole/":          common.AdminPermission,
}

func buildAppAccessGroupPrefix(domainName string) string {
	return "app-" + domainName + "-"
}

func buildAppAccessGroupName(groupNamePrefix string, roleName string, rolePermission common.AppPermission) string {
	return groupNamePrefix + roleName + fmt.Sprintf("%x", common.AppPermission(rolePermission))
}

func buildDataAccessGroupName(domainName string) string {
	return "ucp-" + domainName + "-" + common.DataAccessPrefix
}

func createRetryLambdaEventMapping(services model.Services, tx core.Transaction) (bool, error) {
	retryLambdaName := services.RetryLambdaConfig.Get("FunctionName")
	errorTableName := services.ErrorDB.TableName
	tx.Info("[CreateUcpDomain][eventSourceMapping] Creating Retry Lambda Event Mapping (if not exists) for lambda %v and table %v", retryLambdaName, errorTableName)
	streamArn, err := services.ErrorDB.GetStreamArn()
	if err != nil {
		tx.Error("[CreateUcpDomain][eventSourceMapping] error getting stream arn: %v", err)
		return false, errors.New("retry Lambda Event Mapping not created because stream arn could not be fetched")
	}
	if streamArn == "" {
		tx.Error("[CreateUcpDomain][eventSourceMapping] Retry Lambda Event Mapping not created because Error table streamArn is empty")
		return false, errors.New("retry Lambda Event Mapping not created because Error table streamArn is empty")
	}

	mappings, err := services.RetryLambdaConfig.SearchDynamoEventSourceMappings(streamArn, retryLambdaName)
	if err != nil {
		tx.Error("[CreateUcpDomain][eventSourceMapping] Error looking for existing eventSource mapping for retry lambda %v", err)
		return false, err
	}
	if len(mappings) > 0 {
		tx.Info("[CreateUcpDomain][eventSourceMapping] Event source mapping already exists for lambda %v and stream %v", retryLambdaName, streamArn)
		return false, nil
	}
	tx.Debug("[CreateUcpDomain][eventSourceMapping] Creating Retry Lambda Event Mapping for lambda %v and table %v", retryLambdaName, errorTableName)
	err = services.RetryLambdaConfig.CreateDynamoEventSourceMapping(streamArn, retryLambdaName)
	if err != nil {
		tx.Error("[CreateUcpDomain][eventSourceMapping] Error creating Retry Lambda Event Mapping for lambda %v and table %v", retryLambdaName, errorTableName)
		return false, err
	}
	tx.Info("[CreateUcpDomain][eventSourceMapping] Successfully created Retry Lambda Event Mapping for lambda %v and table %v", retryLambdaName, errorTableName)
	return true, nil
}

// Check if the customer is using the granular permission system. Default is true unless customer explicitly deploys with granular permissions disabled.
func isPermissionSystemEnabled(env map[string]string) bool {
	if env["USE_PERMISSION_SYSTEM"] == "" {
		return true // case for customers who deployed infra before the CloudFormation Parameter was added and passed as a Lambda env var
	}

	return env["USE_PERMISSION_SYSTEM"] == "true"
}
