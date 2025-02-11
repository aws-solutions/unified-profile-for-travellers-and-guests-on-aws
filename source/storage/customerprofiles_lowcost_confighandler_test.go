// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/ucp-common/src/feature"
	"testing"
)

func TestConfigHandler(t *testing.T) {
	t.Parallel()
	////////////
	//TEST SETUP
	////////////
	dynamoTableName := "customer-profiles-low-cost-config-test-" + core.GenerateUniqueId()
	cfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = cfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] Error deleting table %v", t.Name(), err)
		}
	})
	err = cfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] error waiting for table creation: %v", t.Name(), err)
	}

	////////////
	//TEST RUN
	////////////
	domainName := randomDomain("tdn_config_handler")
	idResBucket := "test_bucket"
	idResEnabled := true

	h := ConfigHandler{Svc: &cfg, cache: InitLcsCache()}

	_, exists := h.DomainConfigExists(domainName)
	if exists {
		t.Errorf("[%s] domain exists but shouldn't", t.Name())
	}

	featureSetVersion := feature.LATEST_FEATURE_SET_VERSION
	err = h.CreateDomainConfig(LcDomainConfig{
		Name:                 domainName,
		AiIdResolutionBucket: idResBucket,
		Options:              DomainOptions{AiIdResolutionOn: idResEnabled, FeatureSetVersion: featureSetVersion},
	})
	if err != nil {
		t.Errorf("[%s] error creating config: %v", t.Name(), err)
	}
	_, exists = h.DomainConfigExists(domainName)
	if !exists {
		t.Errorf("[%s] domain doesn't exist but should", t.Name())
	}

	domain, err := h.GetDomainConfig(domainName)
	if err != nil {
		t.Errorf("[%s] error getting config: %v", t.Name(), err)
	}
	if domain.Name != domainName {
		t.Errorf("[%s] wrong domain name", t.Name())
	}
	if domain.AiIdResolutionBucket != idResBucket {
		t.Errorf("[%s] wrong idResBucket. should be %v", t.Name(), idResBucket)
	}
	if !domain.Options.AiIdResolutionOn {
		t.Errorf("[%s] ai ID res should be enabled", t.Name())
	}
	if domain.DomainVersion != featureSetVersion {
		t.Errorf("[%s] Domain version does not use the set value", t.Name())
	}
	if domain.CompatibleVersion != featureSetVersion {
		t.Errorf("[%s] Compatible version does not use the set value", t.Name())
	}

	err = h.DeleteDomainConfig(domainName)
	if err != nil {
		t.Errorf("[%s] error deleting config: %v", t.Name(), err)
	}
	_, exists = h.DomainConfigExists(domainName)
	if exists {
		t.Errorf("[%s] domain exists but shouldn't", t.Name())
	}
}

// This is a test to simulate domains which were created before versioning was introduced
// It manually saves an entry in the config table which does not include domainVersion & compatibleVersions
// Upon retrieval, the versions should be same as the initial feature set version
func TestConfigForDomainWithoutDomainVersion(t *testing.T) {
	t.Parallel()
	dynamoTableName := "customer-profiles-low-cost-config-test-" + core.GenerateUniqueId()
	cfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = cfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] Error deleting table %v", t.Name(), err)
		}
	})
	err = cfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] error waiting for table creation: %v", t.Name(), err)
	}

	domainName := randomDomain("tdn_config_handler")
	h := ConfigHandler{Svc: &cfg, cache: InitLcsCache()}

	_, exists := h.DomainConfigExists(domainName)
	if exists {
		t.Errorf("[%s] domain exists but shouldn't", t.Name())
	}

	// Save config without version attributes
	dyn := DynamoLcDomainConfig{
		Pk:                   domainName,
		Sk:                   DOMAIN_CONFIG_SK,
		Options:              DomainOptions{},
		AiIdResolutionBucket: "",
		QueueUrl:             "",
		KmsArn:               "",
		EnvName:              "",
		CustomerProfileMode:  false,
		DynamoMode:           false,
	}
	_, err = h.Svc.Save(dyn)
	if err != nil {
		t.Errorf("[%s] error creating config: %v", t.Name(), err)
	}
	_, exists = h.DomainConfigExists(domainName)
	if !exists {
		t.Errorf("[%s] domain doesn't exist but should", t.Name())
	}

	domain, err := h.GetDomainConfig(domainName)
	if err != nil {
		t.Errorf("[%s] error getting config: %v", t.Name(), err)
	}
	if domain.Name != domainName {
		t.Errorf("[%s] wrong domain name", t.Name())
	}

	if domain.DomainVersion != feature.INITIAL_FEATURE_SET_VERSION {
		t.Errorf("[%s] Domain version does not use the set value", t.Name())
	}
	if domain.CompatibleVersion != feature.INITIAL_FEATURE_SET_VERSION {
		t.Errorf("[%s] Compatible version does not use the set value", t.Name())
	}

	err = h.DeleteDomainConfig(domainName)
	if err != nil {
		t.Errorf("[%s] error deleting config: %v", t.Name(), err)
	}
	_, exists = h.DomainConfigExists(domainName)
	if exists {
		t.Errorf("[%s] domain exists but shouldn't", t.Name())
	}
}

func TestNonExistentDomainConfig(t *testing.T) {
	t.Parallel()
	dynamoTableName := randomTable("nonexistent-domain-get")
	cfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%v] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err := cfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%v] Error deleting dynamo table", t.Name())
		}
	})
	err = cfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%v] error waiting for table creation: %v", t.Name(), err)
	}
	domainName := randomDomain("nonexistent-domain-get")
	h := ConfigHandler{Svc: &cfg, cache: InitLcsCache()}
	domCfg, err := h.GetDomainConfig(domainName)
	if err != ErrDomainNotFound {
		t.Fatalf("[%v] Should error on nonexistent domain get", t.Name())
	}
	if domCfg.Name != "" {
		t.Fatalf("[%v] Return shouldn't have domain name value", t.Name())
	}
}
