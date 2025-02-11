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

func TestListErrors(t *testing.T) {
	t.Parallel()
	log.Printf("TestListErrors: initialize test resources")
	tName := "ucp-test-list-errors-" + core.GenerateUniqueId()
	dynamo_pk := "error_type"
	dynamo_sk := "error_id"
	cfg, err := db.InitWithNewTable(tName, dynamo_pk, dynamo_sk, "", "")
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
	reg := registry.Registry{ErrorDB: &cfg}
	uc := NewListErrors()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)
	rq := events.APIGatewayProxyRequest{}
	tx.Debug("Api Gateway request", rq)
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "ListErrors", err0)
	}
	err0 = uc.ValidateRequest(wrapper)
	if err0 != nil {
		t.Errorf("[%s] Error validating request request %v", "ListErrors", err0)
	}
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] Error running use case: %v", "ListErrors", err)
	}
	apiRes, err2 := uc.CreateResponse(rs)
	if err2 != nil {
		t.Errorf("[%s] Error creating response %v", "ListErrors", err2)
	}
	tx.Debug("Api Gateway response", apiRes)
}
