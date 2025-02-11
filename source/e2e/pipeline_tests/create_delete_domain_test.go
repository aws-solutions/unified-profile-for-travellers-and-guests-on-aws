// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package e2epipeline

import (
	util "tah/upt/source/e2e"
	"tah/upt/source/ucp-common/src/utils/config"
	uptSdk "tah/upt/source/ucp-sdk/src"
	"testing"
)

func TestCreateDeleteDomain(t *testing.T) {
	t.Parallel()

	_, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("error loading configs: %v", err)
	}

	domainName := util.RandomDomain("t")

	uptHandler, err := uptSdk.Init(infraConfig)
	if err != nil {
		t.Fatalf("error initializing SDK: %v", err)
	}

	err = uptHandler.SetupTestAuth(domainName)
	if err != nil {
		t.Fatalf("error setting up auth: %v", err)
	}
	t.Cleanup(func() {
		err := uptHandler.CleanupAuth(domainName)
		if err != nil {
			t.Errorf("error cleaning up auth: %v", err)
		}
	})

	iterations := 3
	for i := 0; i < iterations; i++ {
		err = uptHandler.CreateDomainAndWait(domainName, 300)
		if err != nil {
			t.Fatalf("error creating domain on iteration %d: %v", i, err)
		}

		err := uptHandler.DeleteDomainAndWait(domainName)
		if err != nil {
			t.Fatalf("Error deleting domain on iteration %d: %v", i, err)
		}
	}
}
