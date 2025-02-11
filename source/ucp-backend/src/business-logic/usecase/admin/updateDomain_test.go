// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func TestUpdateDomainV2(t *testing.T) {
	t.Parallel()
	// Setup
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	mockDomain := customerprofiles.Domain{}
	mockMappings := []customerprofiles.ObjectMapping{}
	profileStorage := customerprofiles.InitMockV2()
	profileStorage.On("GetDomain").Return(mockDomain, nil)
	profileStorage.On("GetMappings").Return(mockMappings, nil)
	for _, accpRecord := range admin.ACCP_RECORDS {
		profileStorage.On("CreateMapping", accpRecord.Name, "mock.Anything", "mock.Anything").Return(nil)
	}
	errorTableName := "update-domain-test-" + time.Now().Format("2006-01-02-15-04-05")
	dbClient := db.Init(errorTableName, "pk", "sk", "", "")
	err := dbClient.CreateTableWithOptions(errorTableName, "pk", "sk", db.TableOptions{
		StreamEnabled: true,
	})
	if err != nil {
		t.Errorf("Could not create config db client %v", err)
	}
	err = dbClient.WaitForTableCreation()
	if err != nil {
		t.Errorf("Error waiting for config table creation: %v", err)
	}
	registry := registry.NewRegistry("region", core.LogLevelDebug, registry.ServiceHandlers{
		Accp:        profileStorage,
		ErrorDB:     &dbClient,
		RetryLambda: lambda.InitMock("retryMock"),
		Cognito:     cognito.InitMock(nil, nil),
	})
	req := events.APIGatewayProxyRequest{}

	uc := NewUpdateDomain()

	// Test Name()
	if uc.Name() != "UpdateDomain" {
		t.Errorf("Expected 'UpdateDomain', got '%s'", uc.Name())
	}
	// Test SetTx()
	uc.SetTx(&tx)
	// Test Tx()
	if uc.Tx().TransactionID != tx.TransactionID {
		t.Errorf("Expected transaction ID '%s, got '%s'", tx.TransactionID, uc.Tx().TransactionID)
	}
	// Test SetRegistry()
	uc.SetRegistry(&registry)
	// Test Registry()
	if uc.Registry() != &registry {
		t.Errorf("Actual registry does not match expected")
	}
	// Test CreateRequest()
	reqWrapper, err := uc.CreateRequest(req)
	if err != nil {
		t.Errorf("Error creating request: %s", err)
	}
	// Test ValidateRequest()
	err = uc.ValidateRequest(reqWrapper)
	if err != nil {
		t.Errorf("Error validating request: %s", err)
	}
	// Test Run()
	resWrapper, err := uc.Run(reqWrapper)
	if err != nil {
		t.Errorf("Error running use case: %s", err)
	}
	// Test CreateResponse()
	res, err := uc.CreateResponse(resWrapper)
	if err != nil {
		t.Errorf("Error creating response: %s", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", res.StatusCode)
	}

	// Cleanup
	err = dbClient.DeleteTable(dbClient.TableName)
	if err != nil {
		t.Errorf("Error deleting table: %s", err)
	}
}
