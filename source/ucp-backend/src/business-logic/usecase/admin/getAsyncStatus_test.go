// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"log"
	"testing"

	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

func TestGetAsyncStatus(t *testing.T) {
	t.Parallel()
	log.Printf("TestGetAsyncStatus: initialize test resources")
	tName := "ucp-test-Async-status-" + core.GenerateUniqueId()
	dynamo_portal_config_pk := "item_id"
	dynamo_portal_config_sk := "item_type"
	cfg, err := db.InitWithNewTable(tName, dynamo_portal_config_pk, dynamo_portal_config_sk, "", "")
	if err != nil {
		t.Fatalf("[TestUpdatePortalConfig] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = cfg.DeleteTable(tName)
		if err != nil {
			t.Errorf("Error deleting table %v", err)
		}
	})
	cfg.WaitForTableCreation()

	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	reg := registry.Registry{ConfigDB: &cfg}
	uc := NewGetAsyncStatus()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)
	rq := events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{
			"id":      "test_id",
			"usecase": "createDomain",
		},
	}
	tx.Debug("Api Gateway request", rq)
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "GetAsyncStatus", err0)
	}
	err0 = uc.ValidateRequest(wrapper)
	if err0 != nil {
		t.Errorf("[%s] Error validating request request %v", "GetAsyncStatus", err0)
	}
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] Error running use case: %v", "GetAsyncStatus", err)
	}
	apiRes, err2 := uc.CreateResponse(rs)
	if err2 != nil {
		t.Errorf("[%s] Error creating response %v", "GetAsyncStatus", err2)
	}
	tx.Debug("Api Gateway response", apiRes)
}
