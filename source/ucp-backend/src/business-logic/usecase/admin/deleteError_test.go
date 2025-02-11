// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"encoding/json"
	"log"
	"testing"

	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	async "tah/upt/source/ucp-common/src/model/async"

	"github.com/aws/aws-lambda-go/events"
)

func TestDeleteError(t *testing.T) {
	t.Parallel()
	log.Printf("TestDeleteError: initialize test resources")
	tName := "ucp-test-delete-errors-" + core.GenerateUniqueId()
	dynamo_pk := "error_type"
	dynamo_sk := "error_id"
	cfg, err := db.InitWithNewTable(tName, dynamo_pk, dynamo_sk, "", "")
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = cfg.DeleteTable(tName)
		if err != nil {
			t.Errorf("[%s] Error deleting table %v", t.Name(), err)
		}
	})
	cfg.WaitForTableCreation()
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	reg := registry.Registry{ErrorDB: &cfg}
	uc := NewDeleteError()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)

	log.Printf("TestDeleteError: create test resources")
	testID := "error_2023-07-31-14-19-31fa1213d7-a680-4110-97b1-a654e9b96ab0"
	_, err = cfg.Save(model.UcpIngestionErrorToDelete{
		Type: ERROR_PK,
		ID:   testID,
	})
	if err != nil {
		t.Errorf("[TestDeleteError] error adding test data to dynamo: %v", err)
	}
	errList := []model.UcpIngestionErrorToDelete{}
	err = cfg.FindStartingWith(ERROR_PK, "error_", &errList)
	if err != nil {
		t.Errorf("[TestDeleteError] error listing errors %v", err)
	}
	if len(errList) != 1 {
		t.Errorf("[TestDeleteError] error table should have 1 test error but has %+v", len(errList))
	}
	rq := events.APIGatewayProxyRequest{PathParameters: map[string]string{"id": testID}}
	tx.Debug("Api Gateway request", rq)
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "DeleteError", err0)
	}
	uc.reg.SetAppAccessPermission(constant.ClearAllErrorsPermission)
	err0 = uc.ValidateRequest(wrapper)
	if err0 != nil {
		t.Errorf("[%s] Error validating request request %v", "DeleteError", err0)
	}
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] Error running use case: %v", "DeleteError", err)
	}
	apiRes, err2 := uc.CreateResponse(rs)
	if err2 != nil {
		t.Errorf("[%s] Error creating response %v", "DeleteError", err2)
	}
	tx.Debug("Api Gateway response", apiRes)

	res := model.UcpIngestionErrorToDelete{}
	err = cfg.Get(ERROR_PK, testID, &res)
	if err != nil {
		t.Errorf("[TestDeleteError] error getting test data from dynamo: %v", err)
	}
	if res.ID != "" {
		t.Errorf("[TestDeleteError] error data not deleted from dynamo: %v", err)
	}
	errList = []model.UcpIngestionErrorToDelete{}
	err = cfg.FindStartingWith(ERROR_PK, "error_", &errList)
	if err != nil {
		t.Errorf("[TestDeleteError] error listing errors %v", err)
	}
	if len(errList) != 0 {
		t.Errorf("[TestDeleteError] error table should be empty but is %+v", errList)
	}
}

func TestDeleteErrorsAsync(t *testing.T) {
	// Setup
	tn := "ucp-test-delete-errors-2-" + core.GenerateUniqueId()
	configDb, err := db.InitWithNewTable(tn, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("[TestUpdatePortalConfig] error creating test config table: %v", err)
	}
	t.Cleanup(func() {
		err = configDb.DeleteTable(tn)
		if err != nil {
			t.Errorf("[TestUpdatePortalConfig] error deleting test config table: %v", err)
		}
	})
	configDb.WaitForTableCreation()
	lambdaMock := lambda.InitMock("ucpAsyncEnv")
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	reg := registry.Registry{
		ConfigDB:    &configDb,
		ErrorDB:     &configDb, // must be set but is not used, no need to create additional resource
		AsyncLambda: lambdaMock,
	}
	uc := NewDeleteError()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)

	// Test
	rq := events.APIGatewayProxyRequest{PathParameters: map[string]string{"id": "*"}}
	wrapper, err := uc.CreateRequest(rq)
	if err != nil {
		t.Errorf("[%s] Error creating request %v", "DeleteError", err)
	}
	uc.reg.SetAppAccessPermission(constant.ClearAllErrorsPermission)
	err = uc.ValidateRequest(wrapper)
	if err != nil {
		t.Errorf("[%s] Error validating request request %v", "DeleteError", err)
	}
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] Error running use case: %v", "DeleteError", err)
	}
	apiRes, err := uc.CreateResponse(rs)
	if err != nil {
		t.Errorf("[%s] Error creating response %v", "DeleteError", err)
	}
	if apiRes.StatusCode != 200 {
		t.Errorf("[%s] Error status code should be 200 but is %d", "DeleteError", apiRes.StatusCode)
	}
	// Convert payload and verify
	payloadStr := string(lambdaMock.Payload.([]uint8))
	payload := async.AsyncInvokePayload{}
	err = json.Unmarshal([]byte(payloadStr), &payload)
	if err != nil {
		t.Errorf("[%s] Error parsing payload %v", "DeleteError", err)
	}
	if payload.Usecase != "emptyTable" || payload.Body.(map[string]interface{})["tableName"] != tn ||
		payload.EventID == "" || payload.TransactionID == "" {
		t.Errorf("[%s] Error payload usecase should be emptyTable but is %s", "DeleteError", payload.Usecase)
	}
}
