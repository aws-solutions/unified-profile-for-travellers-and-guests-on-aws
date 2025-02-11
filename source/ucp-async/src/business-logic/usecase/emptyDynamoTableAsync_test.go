// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"testing"
	"time"

	"tah/upt/source/ucp-common/src/model/async"
	ucModel "tah/upt/source/ucp-common/src/model/async/usecase"

	"strings"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
)

func TestEmptyDynamoTable(t *testing.T) {
	// Table setup
	ts := time.Now().Format("20060102150405")
	testTableName := "empty-table-test-" + ts
	testDbClient := db.Init(testTableName, "pk", "sk", "", "")
	err := testDbClient.CreateTableWithOptions(testTableName, "pk", "sk", db.TableOptions{
		StreamEnabled: true,
	})
	if err != nil {
		t.Fatalf("Error initializing db: %v", err)
	}
	t.Cleanup(func() {
		err = testDbClient.DeleteTable(testTableName)
		if err != nil {
			t.Errorf("[%s] error deleting table: %v", t.Name(), err)
		}
	})
	err = testDbClient.WaitForTableCreation()
	if err != nil {
		t.Fatalf("Error waiting for test table creation: %v", err)
	}

	// Async services setup
	configTableName := "empty-table-config-test-" + ts
	configDbClient, err := db.InitWithNewTable(configTableName, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("Could not create config db client %v", err)
	}
	t.Cleanup(func() {
		err = configDbClient.DeleteTable(configTableName)
		if err != nil {
			t.Errorf("[%s] error deleting table: %v", t.Name(), err)
		}
	})
	err = configDbClient.WaitForTableCreation()
	if err != nil {
		t.Fatalf("Error waiting for config table creation: %v", err)
	}
	solutionsConfig := awssolutions.InitMock()
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	lambdaConfig := lambda.InitMock("ucpRetryEnvName")

	services := async.Services{
		ConfigDB:          configDbClient,
		ErrorDB:           testDbClient,
		SolutionsConfig:   solutionsConfig,
		RetryLambdaConfig: lambdaConfig,
	}

	// Execute usecase
	uc := InitEmptyDynamoTable()
	if uc.Name() != "EmptyDynamoTable" {
		t.Errorf("Expected EmptyDynamoTable, got %v", uc.Name())
	}
	payload := async.AsyncInvokePayload{
		EventID:       "test-event-id",
		Usecase:       async.USECASE_EMPTY_TABLE,
		TransactionID: "test-transaction-id",
		Body: ucModel.EmptyTableBody{
			TableName: testTableName,
		},
	}
	err = uc.Execute(payload, services, tx)
	if err != nil {
		t.Errorf("Error invoking usecase: %v", err)
	}

	/////////////////////
	// Test EventSourceMapping Creation
	streamArn, err := testDbClient.GetStreamArn()
	if err != nil || streamArn == "" {
		t.Errorf("Error getting stream arn: %v", err)
	}
	mock := services.RetryLambdaConfig.(*lambda.LambdaMock)
	if len(mock.EventSourceMappings) != 1 {
		t.Errorf("Lambda config Mock should have 1 event source mapping but has %v", len(mock.EventSourceMappings))
	} else {
		if mock.EventSourceMappings[0].StreamArn != streamArn {
			t.Errorf("Event source mapping arn does not match, %v, %v", mock.EventSourceMappings[0].StreamArn, streamArn)
		}
		if !strings.HasPrefix(mock.EventSourceMappings[0].FunctionName, "ucpRetry") {
			t.Errorf("Event source mapping function name %v does not start with ucpRetry", mock.EventSourceMappings[0].FunctionName)
		}
	}
}
