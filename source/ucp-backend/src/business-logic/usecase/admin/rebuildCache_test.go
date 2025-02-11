// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"log"
	"strings"
	"testing"

	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/model/async"
	"tah/upt/source/ucp-common/src/utils/config"

	"github.com/aws/aws-lambda-go/events"
)

func TestRebuildCache(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	log.Printf("[%s] initialize test resources", t.Name())
	domainName := "rebuild_cache_uc_" + strings.ToLower(core.GenerateUniqueId())

	// Config DB
	configTableName := "rebuild-cache-config-" + strings.ToLower(core.GenerateUniqueId())
	configDbClient, err := db.InitWithNewTable(configTableName, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = configDbClient.DeleteTable(configTableName)
		if err != nil {
			t.Errorf("[%s] Error deleting table %v", t.Name(), err)
		}
	})
	configDbClient.WaitForTableCreation()

	// Merge Queue
	mergeQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = mergeQueueClient.CreateRandom("rebuildCacheMQ")
	if err != nil {
		t.Fatalf("[%s] error creating merge queue: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = mergeQueueClient.Delete()
		if err != nil {
			t.Errorf("[%s] error deleting merge queue: %v", t.Name(), err)
		}
	})

	// Profile Storage
	domains := []customerprofiles.Domain{}
	accpConfig := customerprofiles.InitMock(nil, &domains, nil, nil, nil)
	asyncLambdaConfig := lambda.InitMock("ucpAsync")

	// Create Registry
	services := registry.ServiceHandlers{
		AsyncLambda: asyncLambdaConfig,
		Accp:        accpConfig,
		ConfigDB:    &configDbClient,
	}
	reg := registry.NewRegistry(envCfg.Region, core.LogLevelDebug, services)
	reg.Env["ACCP_DOMAIN_NAME"] = domainName
	reg.Env["LAMBDA_ENV"] = "e2e-env"
	reg.Env["MATCH_BUCKET_NAME"] = "match-bucket-name"

	rebuildCacheUc := NewRebuildCache()
	rebuildCacheUc.SetTx(&tx)
	rebuildCacheUc.SetRegistry(&reg)

	// Creating Success Request
	params := make(map[string]string)
	params["type"] = "3"

	req := events.APIGatewayProxyRequest{
		QueryStringParameters: params,
	}

	wrapper, err := rebuildCacheUc.CreateRequest(req)
	if err != nil {
		t.Fatalf("[%v] create request failed, expected success: %v", t.Name(), err)
	}

	// Creating Validation Fail Request
	largeParams := make(map[string]string)
	largeParams["type"] = "250"

	badReqLarge := events.APIGatewayProxyRequest{
		QueryStringParameters: largeParams,
	}

	badWrapper, err := rebuildCacheUc.CreateRequest(badReqLarge)
	if err != nil {
		t.Fatalf("[%v] large create request failed, expected success: %v", t.Name(), err)
	}

	// Out of Range Request
	outOfRangeParams := make(map[string]string)
	outOfRangeParams["type"] = "999"

	oorReq := events.APIGatewayProxyRequest{
		QueryStringParameters: outOfRangeParams,
	}

	_, err = rebuildCacheUc.CreateRequest(oorReq)
	if err == nil {
		t.Fatalf("[%v] oor create request succeeded, expected failure", t.Name())
	}

	// No Param Request
	_, err = rebuildCacheUc.CreateRequest(events.APIGatewayProxyRequest{})
	if err == nil {
		t.Fatalf("[%v] bad create request succeeded, expected failure", t.Name())
	}

	// Good Wrapper Validation
	rebuildCacheUc.reg.SetAppAccessPermission(admin.RebuildCachePermission)
	err = rebuildCacheUc.ValidateRequest(wrapper)
	if err != nil {
		t.Fatalf("[%v] validation failed, expected success: %v", t.Name(), err)
	}
	// Bad Wrapper Validation
	err = rebuildCacheUc.ValidateRequest(badWrapper)
	if err == nil {
		t.Fatalf("[%v] bad validation succeeded, expected failure", t.Name())
	}

	// Good Wrapper Run
	response, err := rebuildCacheUc.Run(wrapper)
	if err != nil {
		t.Fatalf("[%v] rebuild cache execution failed: %v", t.Name(), err)
	}
	asyncEvent := response.AsyncEvent
	if asyncEvent.Usecase != async.USECASE_REBUILD_CACHE {
		t.Fatalf("[%v] incorrect response, expected rebuildCache usecase: %v", t.Name(), err)
	}
	if asyncEvent.Status != async.EVENT_STATUS_INVOKED {
		t.Fatalf("[%v] incorrect response, expected invoked status: %v", t.Name(), err)
	}
}
