// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAccpRecordsNames(t *testing.T) {
	names := AccpRecordsNames()
	for i, name := range names {
		if name != ACCP_RECORDS[i].Name {
			t.Errorf("Expected %s, got %s", name, ACCP_RECORDS[i].Name)
		}

	}
}

func TestAccpRecordsMap(t *testing.T) {
	rceMap := AccpRecordsMap()
	for key, val := range rceMap {
		if key != val.Name {
			t.Errorf("Invalid map. Expected %s, got %s", key, val.Name)
		}
	}
}

func TestIsValidAccpRecord(t *testing.T) {
	expected := []bool{false, true, true}
	for i, testString := range []string{"invalid", ACCP_RECORD_AIR_BOOKING, ACCP_RECORD_CLICKSTREAM} {
		if IsValidAccpRecord(testString) != expected[i] {
			t.Errorf("%v => expected %v, %v", testString, expected[i], IsValidAccpRecord(testString))
		}
	}
}

func getExpectedValue(key string) AppPermission {
	switch key {
	case "SearchProfilePermission":
		return SearchProfilePermission
	case "DeleteProfilePermission":
		return DeleteProfilePermission
	case "MergeProfilePermission":
		return MergeProfilePermission
	case "UnmergeProfilePermission":
		return UnmergeProfilePermission
	case "CreateDomainPermission":
		return CreateDomainPermission
	case "DeleteDomainPermission":
		return DeleteDomainPermission
	case "ConfigureGenAiPermission":
		return ConfigureGenAiPermission
	case "SaveHyperlinkPermission":
		return SaveHyperlinkPermission
	case "SaveRuleSetPermission":
		return SaveRuleSetPermission
	case "ActivateRuleSetPermission":
		return ActivateRuleSetPermission
	case "RunGlueJobsPermission":
		return RunGlueJobsPermission
	case "ClearAllErrorsPermission":
		return ClearAllErrorsPermission
	case "RebuildCachePermission":
		return RebuildCachePermission
	case "IndustryConnectorPermission":
		return IndustryConnectorPermission
	case "ListPrivacySearchPermission":
		return ListPrivacySearchPermission
	case "GetPrivacySearchPermission":
		return GetPrivacySearchPermission
	case "CreatePrivacySearchPermission":
		return CreatePrivacySearchPermission
	case "DeletePrivacySearchPermission":
		return DeletePrivacySearchPermission
	case "PrivacyDataPurgePermission":
		return PrivacyDataPurgePermission
	case "ListRuleSetPermission":
		return ListRuleSetPermission
	default:
		return 0
	}
}

func TestPermissionValuesAreConsistent(t *testing.T) {
	permissionsFile, err := os.ReadFile("../../../../appPermissions.json")
	if err != nil {
		t.Fatalf("[readPermissions] cannot read permission json; error: %s", err)
	}

	var permissions map[string]int
	if err := json.Unmarshal(permissionsFile, &permissions); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	for key, value := range permissions {
		expectedValue := getExpectedValue(key)
		if expectedValue == 0 {
			t.Errorf("Unexpected permission %s found in appPermission.json", key)
			continue
		}
		if AppPermission(1<<value) != expectedValue {
			t.Errorf("Mismatch for key %s: expected %v, got %v", key, expectedValue, 1<<value)
		}
	}
}
