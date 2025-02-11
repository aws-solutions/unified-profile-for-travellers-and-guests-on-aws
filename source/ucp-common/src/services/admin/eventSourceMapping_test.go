// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"log"
	"testing"

	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
)

func TestCreateEventSourceMapping(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	errorTableName := "error-table-event-source-create-tests" + core.GenerateUniqueId()
	errorDbClient := db.Init(errorTableName, "pk", "sk", "", "")
	err := errorDbClient.CreateTableWithOptions(errorTableName, "pk", "sk", db.TableOptions{
		StreamEnabled: true,
	})
	if err != nil {
		t.Errorf("Could not create test DB: %v", err)

	}
	errorDbClient.WaitForTableCreation() // wait if table is still being created
	lambdaConfig := lambda.InitMock("ucpRetryEnvName")

	hasEsm, err := HasEventSourceMapping(lambdaConfig, errorDbClient, tx)
	if err != nil {
		t.Errorf("Error looking for event source mapping %v", err)
	}
	if hasEsm {
		t.Errorf("Event source mapping should not exist")
	}

	hasCreated, err := CreateRetryLambdaEventSourceMapping(lambdaConfig, errorDbClient, tx)
	if err != nil {
		t.Errorf("Could not create retry lambda event source mapping: %v", err)
	}
	/////////////////////
	// Test EventSourceMapping Creation
	streamArn, err := errorDbClient.GetStreamArn()
	if err != nil || streamArn == "" {
		t.Errorf("Error getting stream arn %v", err)
	}
	mock := lambdaConfig
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

	hasEsm, err = HasEventSourceMapping(lambdaConfig, errorDbClient, tx)
	if err != nil {
		t.Errorf("Error looking for event source mapping %v", err)
	}
	if !hasEsm {
		t.Errorf("Event source mapping should exist")
	}

	log.Printf("Running the function again")
	hasCreated, err = CreateRetryLambdaEventSourceMapping(lambdaConfig, errorDbClient, tx)
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

func TestCreateEventSourceMappingError(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	errorTableName := "error-table-event-source-create-tests" + core.GenerateUniqueId()
	errorDbClient := db.Init(errorTableName, "pk", "sk", "", "")
	//create table without streaming activated
	err := errorDbClient.CreateTableWithOptions(errorTableName, "pk", "sk", db.TableOptions{})
	if err != nil {
		t.Errorf("Could not create test DB: %v", err)

	}
	errorDbClient.WaitForTableCreation() // wait if table is still being created
	lambdaConfig := lambda.InitMock("ucpRetryEnvName")

	hasEsm, err := HasEventSourceMapping(lambdaConfig, errorDbClient, tx)
	if err == nil {
		t.Errorf("HasEventSourceMapping should retur and error when no stream exists in dynamo")
	}
	if hasEsm {
		t.Errorf("Event source mapping should not exist")
	}

	hasCreated, err := CreateRetryLambdaEventSourceMapping(lambdaConfig, errorDbClient, tx)
	if err == nil {
		t.Errorf("Could not create retry lambda event source mapping")
	}
	if hasCreated {
		t.Errorf("Event source mapping should not have been created is stream does not exist")
	}

	err = errorDbClient.DeleteTable(errorTableName)
	if err != nil {
		t.Errorf("Error deleting table %v", err)
	}
}
