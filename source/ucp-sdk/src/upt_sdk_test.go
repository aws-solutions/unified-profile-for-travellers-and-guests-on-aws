// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package upt_sdk

import (
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
)

func TestSdk(t *testing.T) {
	_, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Errorf("Error loading configs: %v", err)
	}

	domainName := "sdk_tst_" + strings.ToLower(core.GenerateUniqueId())

	uptSdk, err := Init(infraConfig)
	if err != nil {
		t.Errorf("Error initializing SDK: %v", err)
	}

	err = uptSdk.SetupTestAuth(domainName)
	if err != nil {
		t.Errorf("Error setting up auth: %v", err)
	}
	t.Cleanup(func() {
		err := uptSdk.CleanupAuth(domainName)
		if err != nil {
			t.Errorf("Error cleaning up auth: %v", err)
		}
	})

	err = uptSdk.CreateDomainAndWait(domainName, 300)
	if err != nil {
		t.Errorf("Error creating domain: %v", err)
	}
	t.Cleanup(func() {
		err := uptSdk.DeleteDomainAndWait(domainName)
		if err != nil {
			t.Errorf("Error deleting domain: %v", err)
		}
	})

	//Retrieve Config
	domainCfg, err := uptSdk.GetDomainConfig(domainName)
	if err != nil {
		t.Errorf("Error getting domain config: %v", err)
	}
	if len(domainCfg.UCPConfig.Domains) != 1 {
		t.Errorf("Expected one domain, got %d", len(domainCfg.UCPConfig.Domains))
		return
	}
	if domainCfg.UCPConfig.Domains[0].Name != domainName {
		t.Errorf("Domain name should be %s and not %s", domainName, domainCfg.UCPConfig.Domains[0].Name)
	}

	// For now we are just testing that the SDK makes the right API call (these function are only used for testing at the moment)
	_, err = uptSdk.SearchProfile(domainName, "profile_id", []string{"test_id"})
	if err != nil {
		t.Errorf("Error searching profile: %v", err)
	}

	res, err := uptSdk.RetreiveProfile(domainName, "invalid_id", []string{})
	if err == nil {
		t.Errorf("Retreiving a profile with an invalid ID should return an error")
		return
	}
	expectedErr := "Use case execution failed: no profile exist in domain " + domainName + " for upt_id invalid_id"
	if res.Error.Msg != expectedErr {
		t.Errorf("Retreiving a profile with an invalid ID should return error %s' and NOT '%s'", expectedErr, err.Error())
	}

	//Retrieve Job Status
	res, err = uptSdk.GetJobsStatus(domainName)
	if err != nil {
		t.Errorf("Error getting job status: %v", err)
	}
	if len(res.AwsResources.Jobs) != 7 {
		t.Errorf("Job status response should have %d jobs and not %d", 6, len(res.AwsResources.Jobs))
	}
}
