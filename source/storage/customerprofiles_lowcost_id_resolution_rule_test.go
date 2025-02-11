// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"testing"
)

////////////////////////////////////////////////////////////////////////
// This test file is used to validate the rules against their corresponding IndexRecord objects
// you can add add additional test cases to the PROFILE_TEST_CASES constant
////////////////////////////////////////////////////////////////////////

var NORMALIZATION_SETTINGS_DEFAULT = IndexNormalizationSettings{
	Lowercase: true,
	Trim:      true,
}

var EMPTY_RULE_SET = []Rule{}
var FIRST_NAME_LAST_NAME = []Rule{
	{
		Index: 0,
		Conditions: []Condition{
			{ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField: "FirstName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	},
}
var FIRST_NAME_LAST_NAME_WITH_SKIP_EMPTY_EMAIL = []Rule{
	{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField: "EmailAddress", Op: RULE_OP_EQUALS_VALUE, SkipConditionValue: StringVal(""), IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField: "FirstName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 2, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	},
}
var FIRST_NAME_LAST_NAME_BOOKING_ID = []Rule{
	{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField: "FirstName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 2, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "booking", IncomingObjectField: "booking_id", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	},
}

var FIRST_NAME_LAST_NAME_INVERSION = []Rule{
	{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField: "FirstName", ExistingObjectType: PROFILE_OBJECT_TYPE_NAME, ExistingObjectField: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField: "LastName", ExistingObjectType: PROFILE_OBJECT_TYPE_NAME, ExistingObjectField: "FirstName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	},
}

var CLICKSTREAM_BOOKING_MATCHING = []Rule{
	{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField: "Attribute.booker_id", Op: RULE_OP_NOT_EQUALS, IncomingObjectType2: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField2: "traveler_id", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "booking", IncomingObjectField: "booking_id", ExistingObjectType: "clickstream", ExistingObjectField: "booking_id", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 2, ConditionType: CONDITION_TYPE_FILTER, IncomingObjectType: "booking", IncomingObjectField: "timestamp", ExistingObjectType: "clickstream", ExistingObjectField: "timestamp", Op: RULE_OP_WITHIN_SECONDS, FilterConditionValue: IntValue(5), IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	},
}

var RULE_SETS = map[string]RuleSet{
	"empty":                      {Rules: EMPTY_RULE_SET},
	"first_name_last_name_match": {Rules: FIRST_NAME_LAST_NAME},
	"first_name_last_name_match_with_skip_empty_email": {Rules: FIRST_NAME_LAST_NAME_WITH_SKIP_EMPTY_EMAIL},
	"first_name_last_name_booking_id_match":            {Rules: FIRST_NAME_LAST_NAME_BOOKING_ID},
	"clickstream_booking_matching":                     {Rules: CLICKSTREAM_BOOKING_MATCHING},
	"first_name_last_name_inversion":                   {Rules: FIRST_NAME_LAST_NAME_INVERSION},
}

type ProfileTestCase struct {
	Name                 string
	Profile              profilemodel.Profile
	ObjectType           string
	ObjectData           map[string]string
	ExpectedIndexRecords map[string][]RuleIndexRecord
}

var PROFILE_TEST_CASES = []ProfileTestCase{
	{
		Name:       "Clickstream matching booking",
		Profile:    profilemodel.Profile{ProfileId: "cid2"},
		ObjectType: "clickstream",
		ObjectData: map[string]string{
			"booking_id":  "B123456789",
			"traveler_id": "s1",
			"timestamp":   "2018-01-01T00:00:40",
		},
		ExpectedIndexRecords: map[string][]RuleIndexRecord{
			"clickstream_booking_matching": {
				//no filter files since it is a timestamp condition
				{ConnectID: "cid2", RuleIndex: 0, FilterFields: []FilterField{}, IndexValue1: "", IndexValue2: "b123456789"},
			},
		},
	},
	{
		Name:       "Profile with all fields and booking matching clickstream",
		Profile:    profilemodel.Profile{ProfileId: "cid1", FirstName: "John ", LastName: "Doe", EmailAddress: ""},
		ObjectType: "booking",
		ObjectData: map[string]string{
			"FirstName":   "John",
			"LastName":    "Doe",
			"booking_id":  "B123456789",
			"traveler_id": "t1",
			"booker_id":   "t1",
			"timestamp":   "2018-01-01T00:00:00",
		},
		ExpectedIndexRecords: map[string][]RuleIndexRecord{
			"empty": {},
			"first_name_last_name_match": {
				{ConnectID: "cid1", RuleIndex: 0, FilterFields: []FilterField{}, IndexValue1: "john_doe", IndexValue2: "john_doe"},
			},
			"first_name_last_name_match_with_skip_empty_email": {},
			"first_name_last_name_booking_id_match": {
				{ConnectID: "cid1", RuleIndex: 0, FilterFields: []FilterField{}, IndexValue1: "john_doe_b123456789", IndexValue2: "john_doe_b123456789"},
			},
			"clickstream_booking_matching": {
				//no filter files since it is a timestamp condition
				{ConnectID: "cid1", RuleIndex: 0, FilterFields: []FilterField{}, IndexValue1: "b123456789", IndexValue2: ""},
			},
			"first_name_last_name_inversion": {
				//no filter files since it is a timestamp condition
				{ConnectID: "cid1", RuleIndex: 0, FilterFields: []FilterField{}, IndexValue1: "john_doe", IndexValue2: "doe_john"},
			},
		},
	},
	{
		Name:       "Profile with missing fields",
		Profile:    profilemodel.Profile{ProfileId: "cid1", FirstName: " ", LastName: "Doe"},
		ObjectType: "booking",
		ObjectData: map[string]string{
			"FirstName": "",
			"LastName":  "Doe",
		},
		ExpectedIndexRecords: map[string][]RuleIndexRecord{
			"empty": {},
			//We should have no index here since the profile does not have a first name
			"first_name_last_name_match": {},
		},
	},
}

func TestIRIdentifyImpactedObjectTypes(t *testing.T) {
	t.Parallel()
	i := IdentityResolutionHandler{
		Tx: core.NewTransaction(t.Name(), "", core.LogLevelDebug),
	}
	for tIndex, profileTestCase := range PROFILE_TEST_CASES {
		for ruleSetName := range profileTestCase.ExpectedIndexRecords {
			ruleSet := RULE_SETS[ruleSetName]
			i.Tx.Info("Test case-%d (%s) with RuleSet %s", tIndex, profileTestCase.Name, ruleSetName)
			indexRecords := i.createRuleSetIndex(profileTestCase.Profile, profileTestCase.ObjectType, profileTestCase.ObjectData, ruleSet)
			expectedIndexRecords := profileTestCase.ExpectedIndexRecords[ruleSetName]
			if len(expectedIndexRecords) > 0 {
				if len(indexRecords) != len(expectedIndexRecords) {
					t.Errorf("Test case-%d (%s) with RuleSet %s failed: expected %d index records, got %d", tIndex, profileTestCase.Name, ruleSetName, len(expectedIndexRecords), len(indexRecords))
					return
				}
				for i, indexRecord := range indexRecords {
					expected := profileTestCase.ExpectedIndexRecords[ruleSetName][i]
					if expected.ConnectID != indexRecord.ConnectID {
						t.Errorf("Test case-%d (%s) with RuleSet %s failed: expected ConnectID %s, got %s", tIndex, profileTestCase.Name, ruleSetName, expected.ConnectID, indexRecord.ConnectID)
						return
					}
					if len(expected.FilterFields) != len(indexRecord.FilterFields) {
						t.Errorf("Test case-%d (%s)with RuleSet %s failed: expected %d filter fields, got %d", tIndex, profileTestCase.Name, ruleSetName, len(expected.FilterFields), len(indexRecord.FilterFields))
						return
					}
					for i, filterField := range indexRecord.FilterFields {
						expFilterField := expected.FilterFields[i]
						if expFilterField.Name != filterField.Name {
							t.Errorf("Test case-%d (%s)with RuleSet %s failed: expected filter field %s, got %s", tIndex, profileTestCase.Name, ruleSetName, expFilterField.Name, filterField.Name)
						}
						if expFilterField.Value != filterField.Value {
							t.Errorf("Test case-%d (%s)with RuleSet %s failed: expected filter field value %s, got %s", tIndex, profileTestCase.Name, ruleSetName, expFilterField.Value, filterField.Value)
							return
						}
					}
					if expected.RuleIndex != indexRecord.RuleIndex {
						t.Errorf("Test case-%d (%s)with RuleSet %s failed: expected RuleIndex %d, got %d", tIndex, profileTestCase.Name, ruleSetName, expected.RuleIndex, indexRecord.RuleIndex)
						return
					}
					if expected.RuleName != indexRecord.RuleName {
						t.Errorf("Test case-%d (%s) with RuleSet %s failed: expected RuleName %s, got %s", tIndex, profileTestCase.Name, ruleSetName, expected.RuleName, indexRecord.RuleName)
						return
					}
					if expected.IndexValue1 != indexRecord.IndexValue1 {
						t.Errorf("Test case-%d (%s) with RuleSet %s failed: expected IndexValue1 %s, got %s", tIndex, profileTestCase.Name, ruleSetName, expected.IndexValue1, indexRecord.IndexValue1)
						return
					}
					if expected.IndexValue2 != indexRecord.IndexValue2 {
						t.Errorf("Test case-%d (%s)with RuleSet %s failed: expected IndexValue2 %s, got %s", tIndex, profileTestCase.Name, ruleSetName, expected.IndexValue2, indexRecord.IndexValue2)
						return
					}
				}
			} else {
				i.Tx.Info("No expected index provided for Test case-%d (%s) and rule %s", tIndex, profileTestCase.Name, ruleSetName)
			}

			i.Tx.Info("Test case-%d (%s) with RuleSet %s passed", tIndex, profileTestCase.Name, ruleSetName)
		}
	}
}
