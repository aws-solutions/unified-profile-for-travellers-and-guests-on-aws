// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"log"
	"testing"

	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/utils/config"
	"tah/upt/source/ucp-common/src/utils/test"

	"github.com/aws/aws-lambda-go/events"
)

func TestListUcpDomains(t *testing.T) {
	t.Parallel()
	// Set up resources
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("unable to load configs: %v", err)
	}
	log.Printf("TestListUcpDomains: initialize test resources")
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	name := "lcs-test-list-domains-" + core.GenerateUniqueId()
	lcsConfigTable, err := db.InitWithNewTable(name, "pk", "sk", "", "")
	if err != nil {
		t.Errorf("error creating lcs config table: %v", err)
	}
	t.Cleanup(func() {
		err = lcsConfigTable.DeleteTable(lcsConfigTable.TableName)
		if err != nil {
			t.Errorf("error deleting lcs config table: %v", err)
		}
	})
	err = lcsConfigTable.WaitForTableCreation()
	if err != nil {
		t.Errorf("error waiting for lcs config table: %v", err)
	}
	lcsKinesisStream, err := kinesis.InitAndCreate(name, envCfg.Region, "", "", core.LogLevelDebug)
	if err != nil {
		t.Errorf("error creating lcs kinesis stream: %v", err)
	}
	t.Cleanup(func() {
		err = lcsKinesisStream.Delete(lcsKinesisStream.Stream)
		if err != nil {
			t.Errorf("error deleting lcs kinesis stream: %v", err)
		}
	})
	_, err = lcsKinesisStream.WaitForStreamCreation(300)
	if err != nil {
		t.Errorf("error waiting for lcs kinesis stream: %v", err)
	}
	mergeQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = mergeQueueClient.CreateRandom("mergeQueueListDomainsTest")
	if err != nil {
		t.Errorf("error creating merge queue: %v", err)
	}
	t.Cleanup(func() {
		err = mergeQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting merge queue: %v", err)
		}
	})
	cpWriterQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = cpWriterQueueClient.CreateRandom("cpWriterListDomainsTest")
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
	profiles, err := test.InitProfileStorage(initParams)
	if err != nil {
		t.Fatalf("Error initializing profile storage: %v", err)
	}
	reg := registry.Registry{Accp: profiles}
	uc := NewListUcpDomains()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)
	rq := events.APIGatewayProxyRequest{}
	tx.Debug("Api Gateway request", rq)
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "ListUcpDomains", err0)
	}
	err0 = uc.ValidateRequest(wrapper)
	if err0 != nil {
		t.Errorf("[%s] Error validating request request %v", "ListUcpDomains", err0)
	}
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] Error running use case: %v", "ListUcpDomains", err)
	}
	apiRes, err2 := uc.CreateResponse(rs)
	if err2 != nil {
		t.Errorf("[%s] Error creating response %v", "ListUcpDomains", err2)
	}
	tx.Debug("Api Gateway response", apiRes)

	log.Printf("TestListUcpDomains: deleting test resources")
}
