// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"testing"
)

func TestAll(t *testing.T) {
	// Initial setup
	tableName := "connector-link-service-test-" + core.GenerateUniqueId()
	dbClient, err := db.InitWithNewTable(tableName, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("[ConnectorLinkService] Error initializing test db: %v", err)
	}
	t.Cleanup(func() {
		err = dbClient.DeleteTable(tableName)
		if err != nil {
			t.Errorf("[ConnectorLinkService] Error deleting test table: %v", err)
		}
	})
	err = dbClient.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[ConnectorLinkService] Error waiting for table creation: %v", err)
	}

	// Check initial state
	initialDomains, err := GetLinkedDomains(dbClient)
	if err != nil {
		t.Errorf("[ConnectorLinkService] Error getting initial domains: %v", err)
	}
	if len(initialDomains.DomainList) != 0 {
		t.Errorf("[ConnectorLinkService] Expected 0 domains, got %v", len(initialDomains.DomainList))
	}
	// Add domains
	err = AddLinkedDomain(dbClient, "domain1")
	if err != nil {
		t.Errorf("[ConnectorLinkService] Error adding domain: %v", err)
	}
	err = AddLinkedDomain(dbClient, "domain2")
	if err != nil {
		t.Errorf("[ConnectorLinkService] Error adding domain: %v", err)
	}
	// Remove domain
	err = RemoveLinkedDomain(dbClient, "domain1")
	if err != nil {
		t.Errorf("[ConnectorLinkService] Error removing domain: %v", err)
	}
	// Validate domains match expected results
	domains, err := GetLinkedDomains(dbClient)
	if err != nil {
		t.Errorf("[ConnectorLinkService] Error getting domains: %v", err)
	}
	if len(domains.DomainList) != 1 || domains.DomainList[0] != "domain2" {
		t.Errorf("[ConnectorLinkService] Unexpected domain list")
	}
}
