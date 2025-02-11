// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package privacy

import (
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

func TestGetPrivacySearch(t *testing.T) {
	testName := testutils.GetTestName(TestGetPrivacySearch)
	testPostFix := strings.ToLower(core.GenerateUniqueId())
	t.Logf("[%s] Initialize Test Resources", testName)

	uc := NewGetPrivacySearch()
	if uc.Name() != "GetPrivacySearch" {
		t.Fatalf("[%s] UseCase Name does not match", testName)
	}

	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	uc.SetTx(&tx)
	if uc.Tx().LogPrefix != tx.LogPrefix {
		t.Fatalf("[%s] Transaction does not match", testName)
	}
	tableName := "ucp-test-privacy-search-" + testPostFix
	privacyCfg, err := db.InitWithNewTable(tableName, PRIVACY_DB_PK, PRIVACY_DB_SK, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error initializing dynamo table: %v", testName, err)
	}
	privacyCfg.WaitForTableCreation()
	t.Cleanup(func() { privacyCfg.DeleteTable(tableName) })

	domainName := "testDomain"
	connectId := "550e8400-e29b-41d4-a716-446655440000"
	var s3Results []string = []string{"search1", "search2"}
	var matchDbResults []string = []string{"search3", "search4"}
	var profileStorageResults []string = []string{"search5", "search6"}
	privacySearchSeedData := usecase.PrivacySearchResult{
		DomainName: domainName,
		ConnectId:  connectId,
		LocationResults: map[usecase.LocationType][]string{
			usecase.S3LocationType:             s3Results,
			usecase.MatchDbLocationType:        matchDbResults,
			usecase.ProfileStorageLocationType: profileStorageResults,
		},
	}
	privacyCfg.Save(privacySearchSeedData)

	reg := registry.Registry{
		PrivacyDB: &privacyCfg,
		Env: map[string]string{
			constants.ACCP_DOMAIN_NAME_ENV_VAR: domainName,
		},
	}

	uc.SetRegistry(&reg)
	if uc.Registry().PrivacyDB.TableName != reg.PrivacyDB.TableName {
		t.Fatalf("[%s] privacyDB TableName does not match", testName)
	}
	if uc.Registry().Env[constants.ACCP_DOMAIN_NAME_ENV_VAR] != reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR] {
		t.Fatalf("[%s] Env does not match", testName)
	}

	rq := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			ID_PATH_PARAMETER: connectId,
		},
		Headers: map[string]string{
			DOMAIN_HEADER: domainName,
		},
	}
	tx.Debug("[%s] API Gateway Request: %v", testName, rq)
	wrapper, err := uc.CreateRequest(rq)
	if err != nil {
		t.Fatalf("[%s] Error creating request: %v", testName, err)
	}
	uc.reg.SetAppAccessPermission(admin.GetPrivacySearchPermission)
	err = uc.ValidateRequest(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error Validating request: %v", testName, err)
	}
	res, err := uc.Run(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error running use case: %v", testName, err)
	}

	if res.PrivacySearchResult.DomainName != domainName {
		t.Fatalf("[%s] domain name does not match; expected: %s, got: %s", testName, domainName, res.PrivacySearchResult.DomainName)
	}
	if res.PrivacySearchResult.ConnectId != connectId {
		t.Fatalf("[%s] connect id does not match; expected: %s, got: %s", testName, connectId, res.PrivacySearchResult.ConnectId)
	}
	if len(res.PrivacySearchResult.LocationResults) != 3 {
		t.Fatalf("[%s] length of location results should be 3", testName)
	}
	if len(res.PrivacySearchResult.LocationResults[usecase.S3LocationType]) != 2 {
		t.Fatalf("[%s] length of s3 results should be 2", testName)
	}
	if len(res.PrivacySearchResult.LocationResults[usecase.MatchDbLocationType]) != 2 {
		t.Fatalf("[%s] length of match db results should be 2", testName)
	}
	if len(res.PrivacySearchResult.LocationResults[usecase.ProfileStorageLocationType]) != 2 {
		t.Fatalf("[%s] length of profile storage results should be 2", testName)
	}

	apiResponse, err := uc.CreateResponse(res)
	if err != nil {
		t.Fatalf("[%s] Error creating response: %v", testName, err)
	}
	tx.Debug("[%s] API Response: %v", testName, apiResponse)
}
