// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"testing"

	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/datasync"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constants "tah/upt/source/ucp-common/src/constant/admin"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	services "tah/upt/source/ucp-common/src/services/admin"

	"github.com/aws/aws-lambda-go/events"
)

func TestLinkIndustryConnector(t *testing.T) {
	t.Parallel()
	// Create test table
	domainList := commonModel.IndustryConnectorDomainList{
		ItemID:     constants.CONFIG_DB_LINKED_CONNECTORS_PK,
		ItemType:   constants.CONFIG_DB_LINKED_CONNECTORS_SK,
		DomainList: []string{"domain1"},
	}
	tableName := "config-link-connector-test-" + core.GenerateUniqueId()
	configDb, err := db.InitWithNewTable(tableName, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("Error while initializing db: %v", err)
	}
	t.Cleanup(func() {
		err = configDb.DeleteTable(tableName)
		if err != nil {
			t.Errorf("Error deleting table %v", err)
		}
	})
	configDb.WaitForTableCreation()
	_, err = configDb.Save(domainList)
	if err != nil {
		t.Errorf("Error while creating test table records: %v", err)
	}
	// Create mock DataSync client
	dsc := datasync.InitMock(nil, nil)

	// Run tests
	addDomainReqBody := "{\"linkIndustryConnectorRq\":{\"domainName\":\"domain2\",\"removeLink\":false}}"
	addDomainExpectedResults := []string{"domain1", "domain2"}
	runTest(t, &configDb, dsc, addDomainReqBody, addDomainExpectedResults)

	removeDomainReqBody := "{\"linkIndustryConnectorRq\":{\"domainName\":\"domain1\",\"removeLink\":true}}"
	removeDomainExpectedResults := []string{"domain2"}
	runTest(t, &configDb, dsc, removeDomainReqBody, removeDomainExpectedResults)
}

// Helper function to run test usecase
func runTest(t *testing.T, configDb *db.DBConfig, dsc datasync.IConfig, reqBody string, expectedDomains []string) {
	// Create usecase
	usecase := NewLinkIndustryConnector()
	if usecase.Name() != "LinkIndustryConnector" {
		t.Errorf("Expected name LinkIndustryConnector, got %s", usecase.Name())
	}

	// Set tx
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	usecase.SetTx(&tx)
	if usecase.Tx().LogPrefix != t.Name() {
		t.Error("Expected tx to be set")
	}

	// Set registry
	reg := registry.Registry{
		ConfigDB: configDb,
		DataSync: dsc,
	}
	usecase.SetRegistry(&reg)
	if usecase.reg == nil || usecase.reg.ConfigDB == nil {
		t.Errorf("Expected registry to be set")
	}

	// Create request
	apigwReq := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Body:       reqBody,
	}
	req, err := usecase.CreateRequest(apigwReq)
	if err != nil {
		t.Errorf("Error while creating request: %v", err)
	}
	usecase.reg.SetAppAccessPermission(constants.IndustryConnectorPermission)

	// Validate request
	err = usecase.ValidateRequest(req)
	if err != nil {
		t.Errorf("Error while validating request: %v", err)
	}

	// Run
	res, err := usecase.Run(req)
	if err != nil {
		t.Errorf("Error while running usecase: %v", err)
	}

	// Check actual domain string saved to table
	actualDomains, err := services.GetLinkedDomains(*usecase.reg.ConfigDB)
	if err != nil {
		t.Errorf("Error while getting domain list: %v", err)
	}
	if !core.IsEqual(actualDomains.DomainList, expectedDomains) {
		t.Errorf("Expected domains %s, got %s", expectedDomains, actualDomains.DomainList)
	}

	// Create response
	apigwRes, err := usecase.CreateResponse(res)
	if err != nil {
		t.Errorf("Error while creating response: %v", err)
	}
	if apigwRes.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", apigwRes.StatusCode)
	}
}
