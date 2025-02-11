// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
)

///////////////////////////////////////////////////////////////////////////
// Thes test validate that when emptying the error table, the lambda event-source mapping
// is recreated correctly for the rety lambda to receive dynamo stream
//////////////////////////////////////////////////////////////////////////////

func TestErrorRefresh(t *testing.T) {
	t.SkipNow()
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}
	services, err := InitServices()
	if err != nil {
		t.Fatalf("Error initializing services: %v", err)
	}
	rootDir, err := config.AbsolutePathToProjectRootDir()
	if err != nil {
		t.Fatalf("Error getting absolute path to project root dir: %v", err)
	}
	postmanDir := rootDir + "source/ucp-backend/e2e/"
	adminConfig, err := Setup(t, services, envConfig, infraConfig)
	if err != nil {
		t.Errorf("Could not Setup test environment: %v", err)
	} else {
		streamArnBeforeRefresh, err := services.ErrorTable.GetStreamArn()
		if err != nil {
			t.Errorf("[%s] Error getting streamArn: %v", t.Name(), err)
		}
		if streamArnBeforeRefresh == "" {
			t.Errorf("[TestRealTime] Error table streamArn is empty before error table refresh")
		}
		mappings, err := services.RetryLambda.SearchDynamoEventSourceMappings(streamArnBeforeRefresh, services.RetryLambda.Get("FunctionName"))
		if err != nil {
			t.Errorf("[TestRealTime] Error searching for event source mapping before error table refresh: %s", err)
		}
		if len(mappings) == 0 {
			t.Errorf("[TestRealTime] Could not find event source mapping before error table refresh")
		}

		folders := []string{AUTH, ERROR_MANAGEMENT, POLL_ASYNC_EVENT}
		deleteErrorCmdRun := buildCommand(postmanDir, folders, adminConfig, []string{})
		err = deleteErrorCmdRun.Run()
		if err != nil {
			t.Errorf("[TestRealTime] Error running delete error command: %s", err)
		}

		streamArnAfterRefresh, err := services.ErrorTable.GetStreamArn()
		if err != nil {
			t.Errorf("[%s] Error getting streamArn: %v", t.Name(), err)
		}
		if streamArnAfterRefresh == "" {
			t.Errorf("[%s] Error table streamArn is empty after error table refresh", t.Name())
		}
		if streamArnAfterRefresh == streamArnBeforeRefresh {
			t.Errorf("[TestRealTime] Error table streamArn did not change after error table refresh")
		}

		mappings, err = services.RetryLambda.SearchDynamoEventSourceMappings(streamArnAfterRefresh, services.RetryLambda.Get("FunctionName"))
		if err != nil {
			t.Errorf("[TestRealTime] Error searching for event source mapping after error table refresh: %s", err)
		}
		if len(mappings) == 0 {
			t.Errorf("[TestRealTime] Could not find event source mapping with new event stream after error table refresh")
		}
	}
	err = Cleanup(t, adminConfig, services, envConfig, infraConfig)
	if err != nil {
		t.Errorf("Could not Cleanup test environment: %v", err)
	}
}
