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
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestGetPurgeStatus(t *testing.T) {
	testPostFix := strings.ToLower(core.GenerateUniqueId())
	testName := t.Name()
	t.Logf("[%s] Initialize Test Resources", testName)

	uc := NewGetPurgeStatus()
	if uc.Name() != "GetPurgeStatus" {
		t.Fatalf("[%s] Expected GetPurgeStatus, got %s", testName, uc.Name())
	}

	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	uc.SetTx(&tx)
	if uc.Tx().LogPrefix != tx.LogPrefix {
		t.Fatalf("[%s] Expected %s, got %s", testName, tx.LogPrefix, uc.Tx().LogPrefix)
	}
	tableName := "ucp-test-privacy-search-" + strings.ToLower(testName) + "-" + testPostFix
	privacyCfg, err := db.InitWithNewTable(tableName, PRIVACY_DB_PK, PRIVACY_DB_SK, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] Error initializing db: %v", testName, err)
	}
	t.Cleanup(func() { privacyCfg.DeleteTable(tableName) })
	err = privacyCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] Error waiting for table creation: %v", testName, err)
	}

	domainName := "testDomain" + testPostFix
	purgeSuccessConnectId := strings.ToLower(core.GenerateUniqueId())
	purgeInvokedConnectId := strings.ToLower(core.GenerateUniqueId())
	purgeRunningConnectId := strings.ToLower(core.GenerateUniqueId())
	purgeFailedConnectId := strings.ToLower(core.GenerateUniqueId())

	//	Save SUCCESS status
	err = privacyCfg.SaveMany([]usecase.PrivacySearchResult{
		{
			DomainName: domainName,
			ConnectId:  purgeSuccessConnectId,
			Status:     usecase.PRIVACY_STATUS_SEARCH_SUCCESS,
			TxID:       tx.TransactionID,
		},
		{
			DomainName: domainName,
			ConnectId:  purgeSuccessConnectId + "|s3://somewhere.parquet",
			Status:     usecase.PRIVACY_STATUS_PURGE_SUCCESS,
			TxID:       tx.TransactionID,
		},
		{
			DomainName: domainName,
			ConnectId:  purgeSuccessConnectId + "alt1",
			Status:     usecase.PRIVACY_STATUS_SEARCH_SUCCESS,
			TxID:       tx.TransactionID,
		},
		{
			DomainName: domainName,
			ConnectId:  purgeSuccessConnectId + "alt1" + "|s3://somewhere.parquet",
			Status:     usecase.PRIVACY_STATUS_PURGE_SUCCESS,
			TxID:       tx.TransactionID,
		},
	})
	if err != nil {
		t.Fatalf("[%s] Error seeding data: %v", testName, err)
	}

	//	Test Purge Success Status; should return true
	reg := registry.Registry{
		PrivacyDB: &privacyCfg,
		Env: map[string]string{
			constants.ACCP_DOMAIN_NAME_ENV_VAR: domainName,
		},
	}

	uc.SetRegistry(&reg)
	if uc.Registry().PrivacyDB.TableName != privacyCfg.TableName {
		t.Fatalf("[%s] privacyDB TableName does not match", testName)
	}
	if uc.Registry().Env[constants.ACCP_DOMAIN_NAME_ENV_VAR] != reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR] {
		t.Fatalf("[%s] Env[ACCP_DOMAIN_NAME_ENV_VAR] does not match", testName)
	}

	rq := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			DOMAIN_HEADER: domainName,
		},
	}
	tx.Debug("[%s] API Gateway Request: %v", testName, rq)
	wrapper, err := uc.CreateRequest(rq)
	if err != nil {
		t.Fatalf("[%s] Error creating request: %v", testName, err)
	}
	uc.reg.SetAppAccessPermission(admin.PrivacyDataPurgePermission)
	err = uc.ValidateRequest(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error validating request: %v", testName, err)
	}
	res, err := uc.Run(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error running request: %v", testName, err)
	}
	if res.PrivacyPurgeStatus.IsPurgeRunning != false {
		t.Fatalf("[%s] Expected %t, got %t", testName, false, true)
	}

	//	TEST SUCCESS + FAILED; should return false
	err = privacyCfg.SaveMany([]usecase.PrivacySearchResult{
		{
			DomainName: domainName,
			ConnectId:  purgeFailedConnectId,
			Status:     usecase.PRIVACY_STATUS_SEARCH_SUCCESS,
			TxID:       tx.TransactionID,
		},
		{
			DomainName: domainName,
			ConnectId:  purgeFailedConnectId + "|s3://somewhere.parquet",
			Status:     usecase.PRIVACY_STATUS_PURGE_FAILED,
			TxID:       tx.TransactionID,
		},
	})
	if err != nil {
		t.Fatalf("[%s] Error seeding data: %v", testName, err)
	}
	res, err = uc.Run(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error running request: %v", testName, err)
	}
	if res.PrivacyPurgeStatus.DomainName != domainName {
		t.Fatalf("[%s] Expected %s, got %s", testName, domainName, res.PrivacyPurgeStatus.DomainName)
	}
	if res.PrivacyPurgeStatus.IsPurgeRunning != false {
		t.Fatalf("[%s] Expected %t, got %t", testName, false, true)
	}

	//	Test SUCCESS + INVOKED; should return true
	err = privacyCfg.SaveMany([]usecase.PrivacySearchResult{
		{
			DomainName: domainName,
			ConnectId:  purgeInvokedConnectId,
			Status:     usecase.PRIVACY_STATUS_SEARCH_SUCCESS,
			TxID:       tx.TransactionID,
		},
		{
			DomainName: domainName,
			ConnectId:  purgeInvokedConnectId + "|s3://somewhere.parquet",
			Status:     usecase.PRIVACY_STATUS_PURGE_INVOKED,
			TxID:       tx.TransactionID,
		},
	})
	if err != nil {
		t.Fatalf("[%s] Error seeding data: %v", testName, err)
	}
	res, err = uc.Run(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error running request: %v", testName, err)
	}
	if res.PrivacyPurgeStatus.IsPurgeRunning != true {
		t.Fatalf("[%s] Expected %t, got %t", testName, true, false)
	}

	//	TEST SUCCESS + INVOKED + RUNNING + FAILED
	err = privacyCfg.SaveMany([]usecase.PrivacySearchResult{
		{
			DomainName: domainName,
			ConnectId:  purgeRunningConnectId,
			Status:     usecase.PRIVACY_STATUS_SEARCH_SUCCESS,
			TxID:       tx.TransactionID,
		},
		{
			DomainName: domainName,
			ConnectId:  purgeRunningConnectId + "|s3://somewhere.parquet",
			Status:     usecase.PRIVACY_STATUS_PURGE_RUNNING,
			TxID:       tx.TransactionID,
		},
	})
	if err != nil {
		t.Fatalf("[%s] Error seeding data: %v", testName, err)
	}
	res, err = uc.Run(wrapper)
	if err != nil {
		t.Fatalf("[%s] Error running request: %v", testName, err)
	}
	if res.PrivacyPurgeStatus.IsPurgeRunning != true {
		t.Fatalf("[%s] Expected %t, got %t", testName, true, false)
	}
}
