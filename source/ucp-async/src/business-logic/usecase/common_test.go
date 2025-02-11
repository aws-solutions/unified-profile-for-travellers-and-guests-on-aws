// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"log"
	"testing"

	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
	model "tah/upt/source/ucp-common/src/model/async"
	"time"
)

func TestCreateEventSourceMapping(t *testing.T) {
	errorTableName := "error-table-event-source-create-tests" + time.Now().Format("20060102150405")
	errorDbClient := db.Init(errorTableName, "pk", "sk", "", "")
	err := errorDbClient.CreateTableWithOptions(errorTableName, "pk", "sk", db.TableOptions{
		StreamEnabled: true,
	})
	if err != nil {
		t.Errorf("Could not create test DB: %v", err)

	}
	errorDbClient.WaitForTableCreation() // wait if table is still being created
	lambdaConfig := lambda.InitMock("ucpRetryEnvName")

	services := model.Services{
		ErrorDB:           errorDbClient,
		RetryLambdaConfig: lambdaConfig,
	}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	hasCreated, err := createRetryLambdaEventMapping(services, tx)
	if err != nil {
		t.Errorf("Could not create retry lambda event source mapping: %v", err)
	}
	/////////////////////
	// Test EventSourceMapping Creation
	streamArn, err := errorDbClient.GetStreamArn()
	if err != nil || streamArn == "" {
		t.Errorf("Error getting stream arn %v", err)
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
	if !hasCreated {
		t.Errorf("hasCreated flag shouldbe set to true is mapping have been created")
	}

	log.Printf("Running the function again")
	hasCreated, err = createRetryLambdaEventMapping(services, tx)
	if err != nil {
		t.Errorf("Could not create retry lambda event source mapping: %v", err)
	}
	if hasCreated {
		t.Errorf("hasCreated flag shouldbe set to false is mapping already existsd")
	}

	err = errorDbClient.DeleteTable(errorTableName)
	if err != nil {
		t.Errorf("Error deleting table %v", err)
	}
}

func TestIsPermissionSystemEnabled(t *testing.T) {
	enabled := isPermissionSystemEnabled(map[string]string{
		"USE_PERMISSION_SYSTEM": "true",
	})
	if enabled != true {
		t.Errorf("isPermissionSystemEnabled should return true when USE_PERMISSION_SYSTEM is set to true")
	}
	enabled = isPermissionSystemEnabled(map[string]string{})
	if enabled != true {
		t.Errorf("isPermissionSystemEnabled should return true when USE_PERMISSION_SYSTEM is not set")
	}
	enabled = isPermissionSystemEnabled(map[string]string{
		"USE_PERMISSION_SYSTEM": "false",
	})
	if enabled != false {
		t.Errorf("isPermissionSystemEnabled should return false when USE_PERMISSION_SYSTEM is set to false")
	}
}
