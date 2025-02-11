// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"strings"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	kms "tah/upt/source/tah-core/kms"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/ucp-common/src/utils/config"
	testutils "tah/upt/source/ucp-common/src/utils/test"
	"testing"
)

func TestSearchDomains(t *testing.T) {
	// Set up resources
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("unable to load configs: %v", err)
	}
	kmsc := kms.Init(envCfg.Region, "", "")
	name := "lcs-test-search-domains-" + core.GenerateUniqueId()
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
	_, err = mergeQueueClient.CreateRandom("listDomainTest")
	if err != nil {
		t.Errorf("error creating merge queue: %v", err)
	}
	t.Cleanup(func() {
		err = mergeQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting merge queue: %v", err)
		}
	})
	initParams := testutils.ProfileStorageParams{
		LcsConfigTable:   &lcsConfigTable,
		LcsKinesisStream: lcsKinesisStream,
		MergeQueueClient: &mergeQueueClient,
	}
	profileStorage, err := testutils.InitProfileStorage(initParams)
	if err != nil {
		t.Fatalf("unable to initialize profile storage")
	}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)

	keyArn, err1 := kmsc.CreateKey("tah-unit-test-key")
	if err1 != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err1)
	}
	domainName := "test_domain" + strings.ToLower(core.GenerateUniqueId())
	domainName2 := "test_domain2" + strings.ToLower(core.GenerateUniqueId())
	options := customerprofiles.DomainOptions{
		AiIdResolutionOn:       false,
		RuleBaseIdResolutionOn: true,
	}
	err7 := profileStorage.CreateDomainWithQueue(domainName, keyArn, map[string]string{"envName": "ucp-common-env"}, "", "", options)
	if err7 != nil {
		t.Errorf("[TestSearchDomains] Error creating domain %v", err7)
	}
	err7 = profileStorage.CreateDomainWithQueue(domainName2, keyArn, map[string]string{"envName": "ucp-common-env-2"}, "", "", options)
	if err7 != nil {
		t.Errorf("[TestSearchDomains] Error creating domain %v", err7)
	}
	domains, err8 := SearchDomain(tx, profileStorage, "ucp-common-env")
	if err8 != nil {
		t.Errorf("[TestSearchDomains] Error searching domain %v", err8)
	}
	if len(domains) != 1 {
		t.Errorf("[TestSearchDomains] Error searching domain: should have 1 domain exactly")
	}
	if domains[0].Name != domainName {
		t.Errorf("[TestSearchDomains] Error searching domain: should return domain %v", domainName)
	}

	err10 := profileStorage.DeleteDomainByName(domainName)
	if err10 != nil {
		t.Errorf("[TestSearchDomains] Error deleting domain %v", err10)
	}
	err10 = profileStorage.DeleteDomainByName(domainName2)
	if err10 != nil {
		t.Errorf("[TestSearchDomains] Error deleting domain %v", err10)
	}

	err = kmsc.DeleteKey(keyArn)
	if err != nil {
		t.Errorf("Error deleting key %v", err)
	}

}
