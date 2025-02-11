// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ir

import (
	"encoding/json"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIRMain(t *testing.T) {
	// Set up mocks
	cpMock := customerprofileslcs.InitMockV2()

	// Set Env Variables
	testDomainName := "test_domain_name"
	objectTypeName := "hotel_booking"
	ruleSetVersion := "2"
	ruleId := "0"
	t.Setenv("DOMAIN_NAME", testDomainName)
	t.Setenv("OBJECT_TYPE_NAME", objectTypeName)
	t.Setenv("RULE_SET_VERSION", ruleSetVersion)
	t.Setenv("RULE_ID", ruleId)
	partition := customerprofileslcs.UuidPartition{LowerBound: uuid.Nil, UpperBound: uuid.MustParse("FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF")}
	t.Setenv("PARTITION_LOWER_BOUND", partition.LowerBound.String())
	t.Setenv("PARTITION_UPPER_BOUND", partition.UpperBound.String())

	// Create Test RuleSets
	testCondition1 := customerprofileslcs.Condition{
		Index:               0,
		ConditionType:       customerprofileslcs.CONDITION_TYPE_MATCH,
		IncomingObjectType:  "hotel_booking",
		IncomingObjectField: "bookingId",
	}

	testCondition2 := customerprofileslcs.Condition{
		Index:               0,
		ConditionType:       customerprofileslcs.CONDITION_TYPE_MATCH,
		IncomingObjectType:  "hotel_booking",
		IncomingObjectField: "productId",
	}

	testRule1 := customerprofileslcs.Rule{
		Index:       0,
		Name:        "test_rule",
		Description: "test_rule",
		Conditions:  []customerprofileslcs.Condition{testCondition1},
	}

	testRule2 := customerprofileslcs.Rule{
		Index:       0,
		Name:        "test_rule",
		Description: "test_rule",
		Conditions:  []customerprofileslcs.Condition{testCondition2},
	}

	testActiveRuleSet := customerprofileslcs.RuleSet{
		Rules:                   []customerprofileslcs.Rule{testRule1},
		Name:                    "active",
		LatestVersion:           2,
		LastActivationTimestamp: time.Now(),
	}

	testVersionedRuleSet := customerprofileslcs.RuleSet{
		Rules:                   []customerprofileslcs.Rule{testRule2},
		Name:                    "v1",
		LatestVersion:           0,
		LastActivationTimestamp: time.Now(),
	}

	testRuleSets := []customerprofileslcs.RuleSet{testVersionedRuleSet, testActiveRuleSet}

	// Create Test Objects
	object := make(map[string]interface{})
	var objects []map[string]interface{}
	testConnectId := uuid.New()
	dummyField := "dummy_value"
	object["dummy_field"] = dummyField

	objectStr, err := json.Marshal(object)
	if err != nil {
		t.Logf("error marshalling object: %v", object)
	}

	object["connect_id"] = testConnectId.String()
	objects = append(objects, object)

	cpMock.On("SetDomain", testDomainName).Return(nil)
	cpMock.On("ListIdResRuleSets", true).Return(testRuleSets, nil)
	cpMock.On("GetInteractionTable", testDomainName, objectTypeName, partition, uuid.Nil, 1000).Return(objects, nil)
	cpMock.On("RunRuleBasedIdentityResolution", string(objectStr), objectTypeName, testConnectId.String(), testActiveRuleSet).
		Return(true, nil)
	cpMock.On("GetInteractionTable", testDomainName, objectTypeName, partition, testConnectId, 1000).Return([]map[string]interface{}{}, nil)

	// Call Handler
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	err = HandleWithServices(cpMock, tx)
	if err != nil {
		t.Errorf("expect no error, got: %v", err)
	}
}

func TestUnsetVariables(t *testing.T) {
	// Set up mocks
	cpMock := customerprofileslcs.InitMockV2()

	// Env Variables
	testDomainName := "test_domain_name"
	objectTypeName := "hotel_booking"
	ruleSetVersion := "2"

	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	// Call Handler
	err := HandleWithServices(cpMock, tx)
	if err.Error() != "domain name should be provided" {
		t.Errorf("expect to fail, got: %v", err)
	}
	t.Setenv("DOMAIN_NAME", testDomainName)
	err = HandleWithServices(cpMock, tx)
	if err.Error() != "object type name should be provided" {
		t.Errorf("expect to fail, got: %v", err)
	}
	t.Setenv("OBJECT_TYPE_NAME", objectTypeName)
	err = HandleWithServices(cpMock, tx)
	if err.Error() != "rule set version should be provided" {
		t.Errorf("expect to fail, got: %v", err)
	}
	t.Setenv("RULE_SET_VERSION", ruleSetVersion)
	err = HandleWithServices(cpMock, tx)
	if err.Error() != "rule id should be provided" {
		t.Errorf("expect to fail, got: %v", err)
	}
}

func TestNoncurrentActive(t *testing.T) {
	// Set up mocks
	cpMock := customerprofileslcs.InitMockV2()

	// Env Variables
	testDomainName := "test_domain_name"
	objectTypeName := "hotel_booking"
	ruleSetVersion := "2"
	ruleId := "0"

	t.Setenv("DOMAIN_NAME", testDomainName)
	t.Setenv("OBJECT_TYPE_NAME", objectTypeName)
	t.Setenv("RULE_SET_VERSION", ruleSetVersion)
	t.Setenv("RULE_ID", ruleId)

	// Create Test RuleSets
	testCondition1 := customerprofileslcs.Condition{
		Index:               0,
		ConditionType:       customerprofileslcs.CONDITION_TYPE_MATCH,
		IncomingObjectType:  "hotel_booking",
		IncomingObjectField: "bookingId",
	}

	testRule1 := customerprofileslcs.Rule{
		Index:       0,
		Name:        "test_rule",
		Description: "test_rule",
		Conditions:  []customerprofileslcs.Condition{testCondition1},
	}

	testActiveRuleSet := customerprofileslcs.RuleSet{
		Rules:                   []customerprofileslcs.Rule{testRule1},
		Name:                    "active",
		LatestVersion:           1,
		LastActivationTimestamp: time.Now(),
	}

	testRuleSets := []customerprofileslcs.RuleSet{testActiveRuleSet}

	cpMock.On("SetDomain", testDomainName).Return(nil)
	cpMock.On("ListIdResRuleSets", true).Return(testRuleSets, nil)

	// Call Handler
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	err := HandleWithServices(cpMock, tx)
	if err.Error() != "rule set version mismatch" {
		t.Errorf("expect to fail, got: %v", err)
	}
}
