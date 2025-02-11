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
	"tah/upt/source/ucp-common/src/utils/utils"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestListPrivacySearches(t *testing.T) {
	testName := testutils.GetTestName(TestListPrivacySearches)
	testPostFix := strings.ToLower(core.GenerateUniqueId())
	t.Logf("[%s] Initialize Test Resources", testName)

	uc := NewListPrivacySearches()
	if uc.Name() != "ListPrivacySearches" {
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
	var s3Results []string = []string{"search1", "search2"}
	var matchDbResults []string = []string{"search3", "search4"}
	var profileStorageResults []string = []string{"search5", "search6"}
	privacySearchSeedData := usecase.PrivacySearchResult{
		DomainName: domainName,
		ConnectId:  "12345",
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
	if uc.Registry() != &reg {
		t.Fatalf("[%s] registry set improperly", testName)
	}

	apiReq := events.APIGatewayProxyRequest{}
	tx.Debug("[%s] API Gateway Request: %v", testName, apiReq)
	wrapper, err := uc.CreateRequest(apiReq)
	if err != nil {
		t.Fatalf("[%s] Error creating request: %v", testName, err)
	}
	uc.reg.SetAppAccessPermission(admin.ListPrivacySearchPermission)
	err = uc.ValidateRequest(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error validating request: %v", testName, err)
	}
	res, err := uc.Run(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error running use case: %v", testName, err)
	}

	if res.PrivacySearchResults == nil {
		t.Fatalf("[%s] privacy search results not set", testName)
	}
	privacySearchResults := utils.SafeDereference(res.PrivacySearchResults)
	if len(privacySearchResults) != 1 {
		t.Fatalf("[%s] length of privacy search results should be 1", testName)
	}
	if privacySearchResults[0].DomainName != domainName {
		t.Fatalf("[%s] domain name does not match; expected: %s, got: %s", testName, domainName, privacySearchResults[0].DomainName)
	}
	if privacySearchResults[0].ConnectId != "12345" {
		t.Fatalf("[%s] connect id does not match; expected: %s, got: %s", testName, "12345", privacySearchResults[0].ConnectId)
	}
	if len(privacySearchResults[0].LocationResults) != 0 {
		t.Fatalf("[%s] length of location results should be 0", testName)
	}

	apiResponse, err := uc.CreateResponse(res)
	if err != nil {
		t.Fatalf("[%s] Error creating response: %v", testName, err)
	}
	tx.Debug("[%s] API Gateway Response: %v", testName, apiResponse)

}
