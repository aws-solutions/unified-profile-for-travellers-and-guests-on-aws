// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package privacy

import (
	"encoding/json"
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	testutils "tah/upt/source/ucp-backend/src/business-logic/testutils"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	async "tah/upt/source/ucp-common/src/model/async"
	usecase "tah/upt/source/ucp-common/src/model/async/usecase"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

func TestSearchForData(t *testing.T) {
	testName := t.Name()
	t.Logf("[%s] Initialize Test Resources", testName)

	testPostFix := strings.ToLower(core.GenerateUniqueId())
	asyncLambdaConfig := lambda.InitMock("ucpAsyncEnv")

	uc := NewCreatePrivacySearch()
	if uc.Name() != "SearchData" {
		t.Errorf("[%s] Incorrect Name", testName)
	}

	tx := core.NewTransaction(testName, testPostFix, core.LogLevelDebug)
	uc.SetTx(&tx)
	if uc.Tx().LogPrefix != tx.LogPrefix && uc.Tx().TransactionID != tx.TransactionID {
		t.Errorf("[%s] Incorrect Tx", testName)
	}

	configTableName := "create-delete-test-" + testPostFix
	configDbClient, err := db.InitWithNewTable(configTableName, "item_id", "item_type", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("[%s] Error initializing db: %v", testName, err)
	}
	t.Cleanup(func() { configDbClient.DeleteTable(configTableName) })
	configDbClient.WaitForTableCreation()

	privacyTableName := "privacy-table-" + testPostFix
	privacyDbClient, err := db.InitWithNewTable(
		privacyTableName,
		"domainName",
		"connectId",
		core.TEST_SOLUTION_ID,
		core.TEST_SOLUTION_VERSION,
	)
	if err != nil {
		t.Errorf("[%s] Error initializing privacy db: %v", testName, err)
	}
	t.Cleanup(func() { privacyDbClient.DeleteTable(privacyTableName) })
	privacyDbClient.WaitForTableCreation()

	services := registry.ServiceHandlers{
		AsyncLambda: asyncLambdaConfig,
		PrivacyDB:   &privacyDbClient,
		ConfigDB:    &configDbClient,
	}
	reg := registry.NewRegistry(testutils.GetTestRegion(), core.LogLevelDebug, services)
	domainName := "test-domain"
	reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR] = domainName
	uc.SetRegistry(&reg)
	if uc.Registry() != &reg {
		t.Errorf("[%s] Incorrect Registry", testName)
	}

	connectIds := []string{uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()}
	createPrivacySearchRqs := make([]model.CreatePrivacySearchRq, len(connectIds))
	for i, connectId := range connectIds {
		createPrivacySearchRqs[i] = model.CreatePrivacySearchRq{
			ConnectId: connectId,
		}
	}
	reqWrapper := model.RequestWrapper{
		CreatePrivacySearchRq: createPrivacySearchRqs,
	}
	jsonBody, err := json.Marshal(reqWrapper)
	if err != nil {
		t.Fatalf("[%s] Error marshaling request: %v", testName, err)
	}
	req := events.APIGatewayProxyRequest{
		Body: string(jsonBody),
	}
	tx.Debug("[%s] API Gateway Request: %v", testName, req)
	wrapper, err := uc.CreateRequest(req)
	if err != nil {
		t.Fatalf("[%s] Error creating request: %v", testName, err)
	}
	uc.reg.SetAppAccessPermission(admin.CreatePrivacySearchPermission)
	err = uc.ValidateRequest(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error Validating request: %v", testName, err)
	}

	badWrapper, err := uc.CreateRequest(events.APIGatewayProxyRequest{})
	if err != nil {
		t.Fatalf("[%s] Error creating request: %v", testName, err)
	}
	err = uc.ValidateRequest(badWrapper)
	if err == nil {
		t.Fatalf("[%s] Validation should have failed: %v", testName, badWrapper)
	}

	err = configDbClient.WaitForTableCreation()
	if err != nil {
		t.Fatalf("error waiting for config db: %v", err)
	}
	err = privacyDbClient.WaitForTableCreation()
	if err != nil {
		t.Fatalf("error waiting for privacy db: %v", err)
	}

	res, err := uc.Run(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error running use case: %v", testName, err)
	}
	apiResponse, err := uc.CreateResponse(res)
	if err != nil {
		t.Fatalf("[%s] Error creating response: %v", testName, err)
	}
	tx.Debug("[%s] API Response: %v", testName, apiResponse)
	if apiResponse.StatusCode != 200 {
		t.Fatalf("[%s] Incorrect Status Code", testName)
	}
	jsonRes, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("[%s] Error marshaling response: %v", testName, err)
	}
	if apiResponse.Body != string(jsonRes) {
		t.Fatalf("[%s] Incorrect Body", testName)
	}
	if res.AsyncEvent.Status != async.EVENT_STATUS_INVOKED {
		t.Fatalf("[%s] Incorrect Status", testName)
	}

	asyncSearchResult := async.AsyncEvent{}
	err = configDbClient.Get(res.AsyncEvent.Usecase, res.AsyncEvent.EventID, &asyncSearchResult)
	if err != nil {
		t.Fatalf("[%s] Error getting Async Status result: %v", testName, err)
	}
	if asyncSearchResult.EventID == "" {
		t.Log("no async event found, waiting for dynamo to propagate")
		time.Sleep(time.Second)
		err = configDbClient.Get(res.AsyncEvent.Usecase, res.AsyncEvent.EventID, &asyncSearchResult)
		if err != nil {
			t.Fatalf("[%s] error getting async event: %v", testName, err)
		}
	}
	if asyncSearchResult.EventID != res.AsyncEvent.EventID {
		t.Fatalf("[%s] Incorrect Async Event Result, expected: %s; got: %s", testName, res.AsyncEvent.EventID, asyncSearchResult.EventID)
	}
	if asyncSearchResult.Usecase != res.AsyncEvent.Usecase {
		t.Fatalf("[%s] Incorrect Async Event Result, expected: %s; got: %s", testName, domainName, asyncSearchResult.Usecase)
	}
	if asyncSearchResult.Status != async.EVENT_STATUS_INVOKED {
		t.Fatalf("[%s] Incorrect Async Event Result, expected: %s; got: %s", testName, async.EVENT_STATUS_INVOKED, asyncSearchResult.Status)
	}

	privacySearchResult := usecase.PrivacySearchResult{}
	err = privacyDbClient.Get(domainName, connectIds[0], &privacySearchResult)
	if err != nil {
		t.Fatalf("[%s] Error getting privacy search result: %v", testName, err)
	}
	if privacySearchResult.ConnectId != connectIds[0] {
		t.Fatalf("[%s] Incorrect privacy search result, expected: %s; got: %s", testName, connectIds[0], privacySearchResult.ConnectId)
	}
	if privacySearchResult.DomainName != domainName {
		t.Fatalf("[%s] Incorrect privacy search result, expected: %s; got: %s", testName, domainName, privacySearchResult.DomainName)
	}
	if privacySearchResult.Status != usecase.PRIVACY_STATUS_SEARCH_INVOKED {
		t.Fatalf(
			"[%s] Incorrect privacy search result, expected: %s; got: %s",
			testName,
			usecase.PRIVACY_STATUS_SEARCH_INVOKED,
			privacySearchResult.Status,
		)
	}
}
