// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package privacy

import (
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	testutils "tah/upt/source/ucp-backend/src/business-logic/testutils"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/model/async"
	test "tah/upt/source/ucp-common/src/utils/test"
	"testing"
	"time"

	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

func TestPurgePrivacyData(t *testing.T) {
	testName := t.Name()
	testPostFix := core.GenerateUniqueId()

	t.Logf("[%s] Initialize Test Resources", testName)
	asyncLambdaConfig := lambda.InitMock("ucpAsyncEnv")

	uc := NewCreatePrivacyDataPurge()
	if uc.Name() != "PurgePrivacyData" {
		t.Fatalf("[%s] %s", testName, test.FormatErrorMessage("incorrect usecase name", "PurgePrivacyData", uc.Name()))
	}

	tx := core.NewTransaction(testName, testPostFix, core.LogLevelDebug)
	uc.SetTx(&tx)
	if uc.Tx().LogPrefix != tx.LogPrefix && uc.Tx().TransactionID != tx.TransactionID {
		t.Fatalf("[%s] %s", testName, test.FormatErrorMessage("incorrect transaction details", "", ""))
	}

	configTableName := "test-config-table-" + testPostFix
	configDbClient, err := db.InitWithNewTable(configTableName, "item_id", "item_type", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error initialize test-config-table: %v", testName, err)
	}
	configDbClient.WaitForTableCreation()
	t.Cleanup(func() { configDbClient.DeleteTable(configTableName) })

	privacyDbTableName := "test-privacy-table-" + testPostFix
	privacyDbCfg, err := db.InitWithNewTable(
		privacyDbTableName,
		PRIVACY_DB_PK,
		PRIVACY_DB_SK,
		core.TEST_SOLUTION_ID,
		core.TEST_SOLUTION_VERSION,
	)
	if err != nil {
		t.Fatalf("[%s] error initialize test-privacy-table: %v", testName, err)
	}
	err = privacyDbCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] error waiting for test-privacy-table: %v", testName, err)
	}
	t.Cleanup(func() { privacyDbCfg.DeleteTable(privacyDbTableName) })

	domainName := "test-domain-" + testPostFix
	groupName := "ucp-" + domainName + "-admin-" + testPostFix
	cognitoClient := cognito.InitMock(nil, &[]cognito.Group{{Name: groupName, Description: "*/*"}})

	services := registry.ServiceHandlers{
		AsyncLambda: asyncLambdaConfig,
		ConfigDB:    &configDbClient,
		Cognito:     cognitoClient,
		PrivacyDB:   &privacyDbCfg,
	}
	reg := registry.NewRegistry(testutils.GetTestRegion(), core.LogLevelDebug, services)
	reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR] = domainName
	uc.SetRegistry(&reg)
	if uc.Registry() != &reg {
		t.Fatalf("[%s] incorrect registry", testName)
	}

	connectIds := []string{core.GenerateUniqueId(), core.GenerateUniqueId(), core.GenerateUniqueId()}
	reqWrapper := model.RequestWrapper{
		PrivacyPurge: model.CreatePrivacyPurgeRq{
			ConnectIds: connectIds,
		},
	}
	jsonBody, err := json.Marshal(reqWrapper)
	if err != nil {
		t.Fatalf("[%s] error marshaling request: %v", testName, err)
	}
	req := events.APIGatewayProxyRequest{
		Body: string(jsonBody),
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"claims": map[string]interface{}{
					"cognito:groups": []string{groupName},
					"username":       "test_user",
				},
			},
		},
	}
	tx.Debug("[%s] API Gateway Request: %v", testName, req)
	wrapper, err := uc.CreateRequest(req)
	if err != nil {
		t.Fatalf("[%s] error creating request: %v", testName, err)
	}
	uc.reg.SetAppAccessPermission(admin.PrivacyDataPurgePermission)
	err = uc.ValidateRequest(wrapper)
	if err != nil {
		t.Fatalf("[%s] error validating request: %v", testName, err)
	}

	res, err := uc.Run(wrapper)
	if err != nil {
		t.Fatalf("[%s] error running usecase: %v", testName, err)
	}
	apiResponse, err := uc.CreateResponse(res)
	if err != nil {
		t.Fatalf("[%s] error creating response: %v", testName, err)
	}
	tx.Debug("[%s] API Gateway Response: %v", testName, apiResponse)
	if apiResponse.StatusCode != 200 {
		t.Fatalf("[%s] %s", testName, test.FormatErrorMessage("incorrect status code", "200", apiResponse.StatusCode))
	}
	jsonRes, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("[%s] error marshaling response: %v", testName, err)
	}
	if apiResponse.Body != string(jsonRes) {
		t.Fatalf("[%s] %s", testName, test.FormatErrorMessage("incorrect response", string(jsonRes), apiResponse.Body))
	}
	if res.AsyncEvent.Status != async.EVENT_STATUS_INVOKED {
		t.Fatalf(
			"[%s] %s",
			testName,
			test.FormatErrorMessage("incorrect async event status", async.EVENT_STATUS_INVOKED, res.AsyncEvent.Status),
		)
	}

	asyncSearchResult := async.AsyncEvent{}
	err = configDbClient.Get(res.AsyncEvent.Usecase, res.AsyncEvent.EventID, &asyncSearchResult)
	if err != nil {
		t.Fatalf("[%s] error getting async event: %v", testName, err)
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
		t.Fatalf(
			"[%s] %s",
			testName,
			test.FormatErrorMessage("incorrect async event id", res.AsyncEvent.EventID, asyncSearchResult.EventID),
		)
	}
	if asyncSearchResult.Usecase != res.AsyncEvent.Usecase {
		t.Fatalf(
			"[%s] %s",
			testName,
			test.FormatErrorMessage("incorrect async event usecase", res.AsyncEvent.Usecase, asyncSearchResult.Usecase),
		)
	}
	if asyncSearchResult.Status != async.EVENT_STATUS_INVOKED {
		t.Fatalf(
			"[%s] %s",
			testName,
			test.FormatErrorMessage("incorrect async event status", async.EVENT_STATUS_INVOKED, asyncSearchResult.Status),
		)
	}
}
