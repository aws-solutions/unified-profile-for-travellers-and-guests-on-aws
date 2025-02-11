// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/ucp-common/src/utils/config"
	"tah/upt/source/ucp-common/src/utils/test"
	"testing"
)

func TestRetrieveTraveller(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	envConfig, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("Error loading env config: %v", err)
	}
	name := "lcs-test-retrieve-traveler-" + core.GenerateUniqueId()
	lcsConfigTable, err := db.InitWithNewTable(name, "pk", "sk", "", "")
	if err != nil {
		t.Fatalf("error creating lcs config table: %v", err)
	}
	t.Cleanup(func() {
		err = lcsConfigTable.DeleteTable(lcsConfigTable.TableName)
		if err != nil {
			t.Errorf("error deleting lcs config table: %v", err)
		}
	})
	err = lcsConfigTable.WaitForTableCreation()
	if err != nil {
		t.Fatalf("error waiting for lcs config table: %v", err)
	}
	lcsKinesisStream, err := kinesis.InitAndCreate(name, envConfig.Region, "", "", core.LogLevelDebug)
	if err != nil {
		t.Fatalf("error creating lcs kinesis stream: %v", err)
	}
	t.Cleanup(func() {
		err = lcsKinesisStream.Delete(lcsKinesisStream.Stream)
		if err != nil {
			t.Errorf("error deleting lcs kinesis stream: %v", err)
		}
	})
	_, err = lcsKinesisStream.WaitForStreamCreation(300)
	if err != nil {
		t.Fatalf("error waiting for lcs kinesis stream: %v", err)
	}
	mergeQueueClient := sqs.Init(envConfig.Region, "", "")
	_, err = mergeQueueClient.CreateRandom("mergerRetrieveTravellerTest")
	if err != nil {
		t.Errorf("error creating merge queue: %v", err)
	}
	t.Cleanup(func() {
		err = mergeQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting merge queue: %v", err)
		}
	})
	cpWriterQueueClient := sqs.Init(envConfig.Region, "", "")
	_, err = cpWriterQueueClient.CreateRandom("cpWriterRetrieveTravellerTest")
	if err != nil {
		t.Errorf("error creating cp writer queue: %v", err)
	}
	t.Cleanup(func() {
		err = cpWriterQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting cp writer queue: %v", err)
		}
	})
	initParams := test.ProfileStorageParams{
		LcsConfigTable:      &lcsConfigTable,
		LcsKinesisStream:    lcsKinesisStream,
		MergeQueueClient:    &mergeQueueClient,
		CPWriterQueueClient: &cpWriterQueueClient,
	}
	accp, err := test.InitProfileStorage(initParams)
	if err != nil {
		t.Fatalf("Error initializing profile storage: %v", err)
	}
	testID := "test_id"
	_, err = RetreiveTraveller(tx, accp, testID, []customerprofiles.PaginationOptions{})
	if err == nil {
		t.Errorf("RetreiveTraveller should have failed with an invalid ID")
	}
}
