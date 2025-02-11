// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package privacy

import (
	"fmt"
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/model/async/usecase"
	testutils "tah/upt/source/ucp-common/src/utils/test"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestDeletePrivacySearch(t *testing.T) {
	testName := testutils.GetTestName(TestSearchForData)
	testPostFix := strings.ToLower(core.GenerateUniqueId())
	t.Logf("[%s] Initialize Test Resources", testName)

	uc := NewDeletePrivacySearch()
	if uc.Name() != "DeletePrivacySearch" {
		t.Fatalf("[%s] Expected DeletePrivacySearch, got %s", testName, uc.Name())
	}

	tx := core.NewTransaction(testName, testPostFix, core.LogLevelDebug)
	uc.SetTx(&tx)
	if uc.Tx().LogPrefix != tx.LogPrefix && uc.Tx().TransactionID != tx.TransactionID {
		t.Fatalf("[%s] Expected %v, got %v", testName, tx, uc.Tx())
	}

	privacyTableName := "ucp-test-privacy-search-" + testPostFix
	privacyCfg, err := db.InitWithNewTable(privacyTableName, PRIVACY_DB_PK, PRIVACY_DB_SK, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] Error initializing privacy table: %v", testName, err)
	}
	t.Cleanup(func() { privacyCfg.DeleteTable(privacyTableName) })
	privacyCfg.WaitForTableCreation()

	domainName := "test_delete_privacy_search_result"
	connectId := core.GenerateUniqueId()
	privacySearchResult := usecase.PrivacySearchResult{
		DomainName: domainName,
		ConnectId:  connectId,
	}
	_, err = privacyCfg.Save(privacySearchResult)
	if err != nil {
		t.Fatalf("[%s] Error saving privacy table: %v", testName, err)
	}
	connectId2 := core.GenerateUniqueId()
	privacySearchResult2 := usecase.PrivacySearchResult{
		DomainName: domainName,
		ConnectId:  connectId2,
	}
	_, err = privacyCfg.Save(privacySearchResult2)
	if err != nil {
		t.Fatalf("[%s] Error saving privacy table: %v", testName, err)
	}

	reg := registry.Registry{
		PrivacyDB: &privacyCfg,
		Env:       map[string]string{},
	}
	uc.SetRegistry(&reg)
	if uc.Registry().PrivacyDB.TableName != privacyCfg.TableName {
		t.Fatalf("[%s] Expected %v, got %v", testName, privacyCfg.TableName, uc.Registry().PrivacyDB.TableName)
	}

	badApiGwReq := events.APIGatewayProxyRequest{}
	badWrapper, err := uc.CreateRequest(badApiGwReq)
	if err != nil {
		t.Fatalf("[%s] error creating request: %v", testName, err)
	}
	err = uc.ValidateRequest(badWrapper)
	if err == nil {
		t.Fatalf("[%s] Expected error, got nil", testName)
	}
	badApiGwReq = events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			ID_PATH_PARAMETER: connectId,
		},
	}
	badWrapper, err = uc.CreateRequest(badApiGwReq)
	if err != nil {
		t.Fatalf("[%s] error creating request: %v", testName, err)
	}
	err = uc.ValidateRequest(badWrapper)
	if err == nil {
		t.Fatalf("[%s] Expected error, got nil", testName)
	}

	//	Add domain name to env var for success path
	uc.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR] = domainName

	apiGwRq := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			ID_PATH_PARAMETER: connectId,
		},
		Body: fmt.Sprintf(`["%s","%s"]`, connectId, connectId2),
	}
	tx.Debug("[%s] API Gateway Request: %v", testName, apiGwRq)

	wrapper, err := uc.CreateRequest(apiGwRq)
	if err != nil {
		t.Fatalf("[%s] Error creating request: %v", testName, err)
	}
	uc.reg.SetAppAccessPermission(admin.DeletePrivacySearchPermission)
	err = uc.ValidateRequest(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error validating request: %v", testName, err)
	}

	res, err := uc.Run(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error running use case: %v", testName, err)
	}

	apiReponse, err := uc.CreateResponse(res)
	if err != nil {
		t.Fatalf("[%s] Error creating response: %v", testName, err)
	}
	tx.Debug("[%s] API Gateway Response: %v", testName, apiReponse)

	privacySearchResult = usecase.PrivacySearchResult{}
	err = privacyCfg.Get(domainName, connectId, &privacySearchResult)
	if err != nil {
		t.Fatalf("[%s] Error getting privacy table: %v", testName, err)
	}
	if privacySearchResult.ConnectId != "" && privacySearchResult.DomainName != "" {
		t.Fatalf("[%s] Expected privacySearchResult to be empty, got %v", testName, privacySearchResult)
	}

	privacySearchResult2 = usecase.PrivacySearchResult{}
	err = privacyCfg.Get(domainName, connectId2, &privacySearchResult)
	if err != nil {
		t.Fatalf("[%s] Error getting privacy table: %v", testName, err)
	}
	if privacySearchResult2.ConnectId != "" && privacySearchResult2.DomainName != "" {
		t.Fatalf("[%s] Expected privacySearchResult to be empty, got %v", testName, privacySearchResult2)
	}
}
