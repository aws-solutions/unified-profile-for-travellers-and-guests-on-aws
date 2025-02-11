// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"log"
	"testing"

	"tah/upt/source/tah-core/appregistry"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

func TestListConnectors(t *testing.T) {
	t.Parallel()
	log.Printf("TestListConnectors: initialize test resources")
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
	appReg := appregistry.Init("eu-west-2", "", "")
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	reg := registry.Registry{ConfigDB: &cfg, AppRegistry: &appReg}
	uc := NewListConnectors()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)
	rq := events.APIGatewayProxyRequest{}
	tx.Debug("Api Gateway request", rq)
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "ListConnectors", err0)
	}
	err0 = uc.ValidateRequest(wrapper)
	if err0 != nil {
		t.Errorf("[%s] Error validating request request %v", "ListConnectors", err0)
	}
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] Error running use case: %v", "ListConnectors", err)
	}

	connectors := uc.applicationsToConnectors([]appregistry.ApplicationSummary{
		{Name: "travel-and-hospitality-connector-app1"},
		{Name: "travel-and-hospitality-connector-app2"},
		{Name: "app3"},
	}, []string{"dom_1", "dom_2"})
	if len(connectors) != 2 {
		t.Errorf("[%s] Error converting applications to connectors. Shoudl have 2 connectors, not %v", "ListConnectors", len(connectors))
	}

	apiRes, err2 := uc.CreateResponse(rs)
	if err2 != nil {
		t.Errorf("[%s] Error creating response %v", "ListConnectors", err2)
	}
	tx.Debug("Api Gateway response", apiRes)
}
