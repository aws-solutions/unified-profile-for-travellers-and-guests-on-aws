// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"encoding/json"
	"reflect"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"testing"
	"time"
)

type AuroraToProfileTestCase struct {
	Data              map[string]interface{}
	Expected          profilemodel.Profile
	ExpectedObjType   string
	ExpectedTimeStamp time.Time
}

func TestGetLaterDate(t *testing.T) {
	t.Parallel()
	testName := t.Name()
	u := LCSUtils{}
	utcNow := time.Now().UTC()
	oneYearBefore := utcNow.AddDate(-1, 0, 0)
	oneHourBefore := utcNow.Add(-1 * time.Hour)
	oneNsBefore := utcNow.Add(-1 * time.Nanosecond)
	oneNsAfter := utcNow.Add(1 * time.Nanosecond)
	oneHourAfter := utcNow.Add(1 * time.Hour)
	oneYearAfter := utcNow.AddDate(1, 0, 0)

	//	utcNow and the testDate are called in alternating order
	//	to the func to ensure that the func is commutative
	test := u.getLaterDate(oneYearBefore, utcNow)
	if test != utcNow {
		t.Fatalf("[%s] Expected %s but got %s", testName, utcNow, test)
	}

	test = u.getLaterDate(utcNow, oneHourBefore)
	if test != utcNow {
		t.Fatalf("[%s] Expected %s but got %s", testName, utcNow, test)
	}

	test = u.getLaterDate(oneNsBefore, utcNow)
	if test != utcNow {
		t.Fatalf("[%s] Expected %s but got %s", testName, utcNow, test)
	}

	test = u.getLaterDate(utcNow, oneNsAfter)
	if test != oneNsAfter {
		t.Fatalf("[%s] Expected %s but got %s", testName, oneNsAfter, test)
	}
	test = u.getLaterDate(oneHourAfter, utcNow)
	if test != oneHourAfter {
		t.Fatalf("[%s] Expected %s but got %s", testName, oneHourAfter, test)
	}
	test = u.getLaterDate(utcNow, oneYearAfter)
	if test != oneYearAfter {
		t.Fatalf("[%s] Expected %s but got %s", testName, oneYearAfter, test)
	}
}

func TestUtilsAuroraToProfile(t *testing.T) {
	t.Parallel()
	u := LCSUtils{}
	now := time.Now()
	tests := []AuroraToProfileTestCase{
		{
			Data: map[string]interface{}{
				"connect_id":                 "abcde",
				"FirstName_ts":               now,
				"timestamp":                  now,
				"FirstName_obj_type":         "air_booking",
				"FirstName":                  "John",
				"LastName":                   "Doe",
				"EmailAddress":               "john.doe@example.com",
				"PhoneNumber":                "1234567890",
				"Address.Address1":           "3 main street",
				"Attributes.customer_attr_1": "test_custom_attribute",
			},
			Expected: profilemodel.Profile{
				LastUpdated:  now,
				ProfileId:    "abcde",
				FirstName:    "John",
				LastName:     "Doe",
				EmailAddress: "john.doe@example.com",
				PhoneNumber:  "1234567890",
				Address: profilemodel.Address{
					Address1: "3 main street",
				},
				Attributes: map[string]string{
					"customer_attr_1": "test_custom_attribute",
				},
			},
			ExpectedObjType:   "air_booking",
			ExpectedTimeStamp: now,
		},
	}

	for _, test := range tests {
		profile, objTypeMap, tsMap := u.auroraToProfile(test.Data)
		actual, _ := json.Marshal(profile)
		excepted, _ := json.Marshal(test.Expected)
		if string(actual) != string(excepted) {
			t.Fatalf("Expected %s but got %s", excepted, actual)
		}
		if objTypeMap["FirstName"] != test.ExpectedObjType {
			t.Fatalf("Expected %s but got %s", test.ExpectedObjType, objTypeMap["FirstName"])
		}
		if tsMap["FirstName"] != test.ExpectedTimeStamp {
			t.Fatalf("Expected %s but got %s", test.ExpectedTimeStamp, tsMap["FirstName"])
		}
	}
}

func TestUtilsParseFielNamesFromTemplate(t *testing.T) {
	t.Parallel()
	tests := map[string][]string{
		"{{.FirstName}} {{.LastName}}": {"FirstName", "LastName"},
	}
	for template, expected := range tests {
		fields := parseFieldNamesFromTemplate(template)
		if !core.IsEqual(fields, expected) {
			t.Errorf("Fields in template %s should be %v and NOT %v", template, expected, fields)
		}
	}
}

var OBJ_MAPPINGS_1 = []ObjectMapping{
	{
		Name: "accp_object_id1",
		Fields: []FieldMapping{
			{
				Type:    "STRING",
				Source:  "_source.accp_object_id",
				Target:  "accp_object_id",
				Indexes: []string{"UNIQUE"},
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.traveler_id",
				Target:  "_profile.Attributes.traveler_id",
				Indexes: []string{"PROFILE"},
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.segment_id",
				Target:  "segment_id",
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.from",
				Target:  "from",
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.arrival_date",
				Target:  "arrival_date",
				KeyOnly: true,
			},
			{
				Type:        "STRING",
				Source:      "_source.email",
				Target:      "_profile.PersonalEmailAddress",
				Searcheable: true,
			},
		},
	},
	{
		Name: "accp_object_id2",
		Fields: []FieldMapping{
			{
				Type:    "STRING",
				Source:  "_source.accp_object_id",
				Target:  "accp_object_id",
				Indexes: []string{"UNIQUE"},
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.traveler_id",
				Target:  "_profile.Attributes.traveler_id",
				Indexes: []string{"PROFILE"},
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.item",
				Target:  "item",
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.price",
				Target:  "price",
				KeyOnly: true,
			},
			{
				Type:   "STRING",
				Source: "_source.cc_number",
				Target: "_profile.Attributes.cc_number",
			},
			{
				Type:   "STRING",
				Source: "_source.cc_name",
				Target: "_profile.Attributes.full_name",
			},
			{
				Type:   "STRING",
				Source: "_source.firstName",
				Target: "_profile.FirstName",
			},
		},
	},
}

type ObjeMappingToColumnTest struct {
	Mappings               []ObjectMapping
	ExpectedObjectColumns  [][]string
	ExpectedProfileColumns []string
}

var TESTS = []ObjeMappingToColumnTest{
	{
		Mappings: OBJ_MAPPINGS_1,
		ExpectedObjectColumns: [][]string{
			{"accp_object_id", "traveler_id", "segment_id", "from", "arrival_date", "email"},
			{"accp_object_id", "traveler_id", "item", "price", "cc_number", "cc_name", "firstName"},
		},
		ExpectedProfileColumns: append(EXPECTED_STANDARD_ATTR, []string{"Attributes.traveler_id", "Attributes.cc_number", "Attributes.full_name"}...),
	},
}

func TestUtilsMappingToDbColumn(t *testing.T) {
	t.Parallel()
	u := LCSUtils{}
	for _, test := range TESTS {
		profileColumns := u.mappingsToProfileDbColumns(test.Mappings)
		if len(profileColumns) != len(test.ExpectedProfileColumns) {
			t.Errorf("profile columns length mismatch, expected %+v, got %+v", len(test.ExpectedProfileColumns), len(profileColumns))
		}
		if !core.IsEqual(profileColumns, test.ExpectedProfileColumns) {
			t.Errorf("profile columns mismatch, expected %+v, got %+v", test.ExpectedProfileColumns, profileColumns)
		}

		for i, mapping := range test.Mappings {
			objectColumns := u.mappingToInteractionDbColumns(mapping)
			if len(objectColumns) != len(test.ExpectedObjectColumns[i]) {
				t.Errorf("Object columns length mismatch, expected %+v, got %+v", len(test.ExpectedObjectColumns[i]), len(objectColumns))
			}
			if !core.IsEqual(objectColumns, test.ExpectedObjectColumns[i]) {
				t.Errorf("Object columns mismatch, expected %+v, got %+v", test.ExpectedObjectColumns[i], objectColumns)
			}

		}

	}
}

func TestGoodConditionValidation(t *testing.T) {
	t.Parallel()
	t.Logf("[%s] initialize test resources", t.Name())
	objectFields := map[string][]string{
		"air_booking": {"to", "timestamp"},
		"email_history": {
			"address",
			"timestamp",
		},
	}
	profileFields := []string{"FirstName", "LastName"}
	objectFields["_profile"] = profileFields

	NORMALIZATION_SETTINGS_DEFAULT := IndexNormalizationSettings{
		Lowercase: true,
		Trim:      true,
	}

	goodSkipCondition := Condition{
		Index:                0,
		ConditionType:        CONDITION_TYPE_SKIP,
		IncomingObjectType:   "_profile",
		IncomingObjectField:  "FirstName",
		Op:                   RULE_OP_EQUALS,
		IncomingObjectType2:  "_profile",
		IncomingObjectField2: "LastName",
		IndexNormalization:   NORMALIZATION_SETTINGS_DEFAULT,
	}

	goodMatchCondition := Condition{
		Index:               1,
		ConditionType:       CONDITION_TYPE_MATCH,
		IncomingObjectType:  "air_booking",
		IncomingObjectField: "to",
		ExistingObjectType:  "email_history",
		ExistingObjectField: "address",
		IndexNormalization:  NORMALIZATION_SETTINGS_DEFAULT,
	}

	goodFilterCondition := Condition{
		Index:                2,
		ConditionType:        CONDITION_TYPE_FILTER,
		IncomingObjectType:   "air_booking",
		IncomingObjectField:  "timestamp",
		Op:                   RULE_OP_WITHIN_SECONDS,
		ExistingObjectType:   "email_history",
		ExistingObjectField:  "timestamp",
		FilterConditionValue: Value{ValueType: CONDITION_VALUE_TYPE_INT, IntValue: 5},
		IndexNormalization:   NORMALIZATION_SETTINGS_DEFAULT,
	}

	testRules := []Rule{
		{
			Index:       1,
			Name:        "test_rule_1",
			Description: "test_rule_1",
			Conditions:  []Condition{goodSkipCondition, goodMatchCondition, goodFilterCondition},
		},
	}

	for _, testRule := range testRules {
		err := validateConditions(testRule.Conditions, objectFields)
		t.Logf("errArr: %v", err)
		if err != nil {
			t.Errorf("Expected error, got %v", err)
		}
	}
}

func TestBadConditionValidation(t *testing.T) {
	t.Logf("[%s] initialize test resources", t.Name())
	objectFields := map[string][]string{
		"air_booking": {"to", "timestamp"},
		"email_history": {
			"address",
			"timestamp",
		},
	}
	profileFields := []string{"FirstName", "LastName"}
	objectFields["_profile"] = profileFields

	NORMALIZATION_SETTINGS_DEFAULT := IndexNormalizationSettings{
		Lowercase: true,
		Trim:      true,
	}

	emptyIncomingRules := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "", IncomingObjectField: "", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidIncomingTypeRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "fake_type", IncomingObjectField: "fake_field", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidIncomingSkipFieldRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "_profile", IncomingObjectField: "fake_field", Op: RULE_OP_EQUALS, IncomingObjectType2: "_profile", IncomingObjectField2: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	emptySkipOpRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "_profile", IncomingObjectField: "FirstName", Op: "", IncomingObjectType2: "_profile", IncomingObjectField2: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidSkipValueRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "_profile", IncomingObjectField: "FirstName", Op: RULE_OP_MATCHES_REGEXP, SkipConditionValue: Value{}, IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	emptyIncoming2Rules := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "_profile", IncomingObjectField: "FirstName", Op: RULE_OP_EQUALS, IncomingObjectType2: "", IncomingObjectField2: "", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidIncomingType2Rule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "_profile", IncomingObjectField: "FirstName", Op: RULE_OP_EQUALS, IncomingObjectType2: "fake_type", IncomingObjectField2: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidIncomingSkip2FieldRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "_profile", IncomingObjectField: "FirstName", Op: RULE_OP_EQUALS, IncomingObjectType2: "_profile", IncomingObjectField2: "fake_field", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	noMatchConditionRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "_profile", IncomingObjectField: "FirstName", Op: RULE_OP_EQUALS, IncomingObjectType2: "_profile", IncomingObjectField2: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	emptyExistingTypeMatchRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "_profile", IncomingObjectField: "FirstName", ExistingObjectType: "", ExistingObjectField: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	emptyExistingFieldMatchRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "_profile", IncomingObjectField: "FirstName", ExistingObjectType: "_profile", ExistingObjectField: "", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidTemplatingIncomingMatchRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "_profile", IncomingObjectField: "{{.InvalidName}}_name", ExistingObjectType: "_profile", ExistingObjectField: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidTemplatingExistingMatchRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "_profile", IncomingObjectField: "FirstName", ExistingObjectType: "_profile", ExistingObjectField: "{{.InvalidName}}_name", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidFilterOpRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_FILTER, IncomingObjectType: "air_booking", IncomingObjectField: "timestamp", Op: "", ExistingObjectType: "email_history", ExistingObjectField: "timestamp", FilterConditionValue: Value{ValueType: CONDITION_VALUE_TYPE_INT, IntValue: 5}, IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidFilterField1Rule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_FILTER, IncomingObjectType: "air_booking", IncomingObjectField: "invalid", Op: RULE_OP_WITHIN_SECONDS, ExistingObjectType: "email_history", ExistingObjectField: "timestamp", FilterConditionValue: Value{ValueType: CONDITION_VALUE_TYPE_INT, IntValue: 5}, IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidFilterField2Rule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_FILTER, IncomingObjectType: "air_booking", IncomingObjectField: "timestamp", Op: RULE_OP_WITHIN_SECONDS, ExistingObjectType: "email_history", ExistingObjectField: "invalid", FilterConditionValue: Value{ValueType: CONDITION_VALUE_TYPE_INT, IntValue: 5}, IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidExistingTypeRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_FILTER, IncomingObjectType: "air_booking", IncomingObjectField: "timestamp", Op: RULE_OP_WITHIN_SECONDS, ExistingObjectType: "invalid_type", ExistingObjectField: "timestamp", FilterConditionValue: Value{ValueType: CONDITION_VALUE_TYPE_INT, IntValue: 5}, IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidFilterValueRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_FILTER, IncomingObjectType: "air_booking", IncomingObjectField: "timestamp", Op: RULE_OP_WITHIN_SECONDS, ExistingObjectType: "email_history", ExistingObjectField: "timestamp", FilterConditionValue: Value{}, IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	invalidFilterMatchRule := Rule{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_FILTER, IncomingObjectType: "email_history", IncomingObjectField: "timestamp", Op: RULE_OP_WITHIN_SECONDS, ExistingObjectType: "_profile", ExistingObjectField: "timestamp", FilterConditionValue: Value{ValueType: CONDITION_VALUE_TYPE_INT, IntValue: 5}, IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	}

	testRules := []Rule{emptyIncomingRules, invalidIncomingTypeRule, invalidIncomingSkipFieldRule, emptySkipOpRule, invalidSkipValueRule, emptyIncoming2Rules, invalidIncomingType2Rule, invalidIncomingSkip2FieldRule, noMatchConditionRule, emptyExistingTypeMatchRule, emptyExistingFieldMatchRule, invalidTemplatingIncomingMatchRule, invalidTemplatingExistingMatchRule, invalidFilterOpRule, invalidFilterField1Rule, invalidFilterField2Rule, invalidExistingTypeRule, invalidFilterValueRule, invalidFilterMatchRule}
	for _, testRule := range testRules {
		err := validateConditions(testRule.Conditions, objectFields)
		if err == nil {
			t.Errorf("error should not be nil %v", testRule.Conditions)
		}
	}
}

func TestConditionValidationCache(t *testing.T) {
	t.Logf("[%s] initialize test resources", t.Name())
	objectFields := map[string][]string{
		"air_booking": {"to", "timestamp"},
		"email_history": {
			"address",
			"timestamp",
		},
	}
	profileFields := []string{"FirstName", "LastName"}
	objectFields["_profile"] = profileFields

	rules := []Rule{
		{
			Index: 0,
			Conditions: []Condition{
				{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "", IncomingObjectField: "", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
				{Index: 1, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			},
		},
		{
			Index: 2,
			Conditions: []Condition{
				{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "fake_type", IncomingObjectField: "fake_field", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
				{Index: 1, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			},
		},
		{
			Index: 3,
			Conditions: []Condition{
				{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "_profile", IncomingObjectField: "fake_field", Op: RULE_OP_EQUALS, IncomingObjectType2: "_profile", IncomingObjectField2: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
				{Index: 1, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "air_booking", IncomingObjectField: "to", ExistingObjectType: "email_history", ExistingObjectField: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			},
		},
		{
			Index: 4,
			Conditions: []Condition{
				{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "_profile", IncomingObjectField: "FirstName", Op: "", IncomingObjectType2: "_profile", IncomingObjectField2: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
				{Index: 1, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "air_booking", IncomingObjectField: "to", IncomingObjectType2: "email_history", IncomingObjectField2: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			},
		},
	}

	for _, testRule := range rules {
		err := validateCacheConditions(testRule.Conditions, objectFields)
		if err == nil {
			t.Errorf("error should not be nil %v", testRule.Conditions)
		}
	}

	goodRules := []Rule{
		{
			Index: 0,
			Conditions: []Condition{
				{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "_profile", IncomingObjectField: "FirstName", Op: RULE_OP_EQUALS, IncomingObjectType2: "_profile", IncomingObjectField2: "LastName", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
				{Index: 1, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "air_booking", IncomingObjectField: "to", Op: RULE_OP_EQUALS, IncomingObjectType2: "email_history", IncomingObjectField2: "address", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			},
		},
	}

	for _, testRule := range goodRules {
		err := validateCacheConditions(testRule.Conditions, objectFields)
		if err != nil {
			t.Errorf("error %v", err)
		}
	}
}

func TestFiltering(t *testing.T) {
	rows := []map[string]interface{}{
		{
			OBJECT_TYPE_NAME_FIELD: "clickstream",
			"accp_object_id":       "test1",
			"timestamp":            time.Now(),
			"field_value":          "right_value",
		},
		{
			OBJECT_TYPE_NAME_FIELD: "air_booking",
			"accp_object_id":       "test2",
			"timestamp":            time.Now().Add(5 * time.Hour),
			"field_value":          "right_value",
		},
		{
			OBJECT_TYPE_NAME_FIELD: "air_booking",
			"accp_object_id":       "test2",
			"timestamp":            time.Now(),
			"field_value":          "wrong_value",
		},
		{
			OBJECT_TYPE_NAME_FIELD: "air_booking",
			"accp_object_id":       "test3",
			"timestamp":            time.Now(),
			"field_value":          "right_value",
		},
		{
			OBJECT_TYPE_NAME_FIELD: "air_booking",
			"accp_object_id":       "test4",
			"timestamp":            time.Now(),
			"field_value":          "right_value",
		},
		{
			OBJECT_TYPE_NAME_FIELD: "purchase",
			"accp_object_id":       "placeholder",
			"actual_object_id":     "test1",
			"timestamp":            time.Now(),
			"field_value":          "right_value",
		},
		{
			OBJECT_TYPE_NAME_FIELD: "purchase",
			"accp_object_id":       "placeholder",
			"actual_object_id":     "test2",
			"timestamp":            time.Now().Add(5 * time.Hour),
			"field_value":          "right_value",
		},
		{
			OBJECT_TYPE_NAME_FIELD: "purchase",
			"accp_object_id":       "placeholder",
			"actual_object_id":     "test3",
			"timestamp":            time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC),
			"field_value":          "right_value",
		},
		{
			OBJECT_TYPE_NAME_FIELD: "purchase",
			"accp_object_id":       "placeholder",
			"actual_object_id":     "test3",
			"timestamp":            time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC),
			"field_value":          "wrong_value",
		},
		{
			OBJECT_TYPE_NAME_FIELD: "hotel_booking",
			"accp_object_id":       "test1",
			"timestamp":            time.Now(),
			"field_value":          "right_value",
		},
		{
			OBJECT_TYPE_NAME_FIELD: "hotel_booking",
			"accp_object_id":       "test2",
			"timestamp":            time.Now(),
			"field_value":          "right_value",
		},
		{
			OBJECT_TYPE_NAME_FIELD: "hotel_booking",
			"accp_object_id":       "test3",
			"timestamp":            time.Now(),
			"field_value":          "right_value",
		},
	}

	objectKeyMap := map[string]string{
		"clickstream":   "accp_object_id",
		"air_booking":   "accp_object_id",
		"purchase":      "actual_object_id",
		"hotel_booking": "accp_object_id",
	}

	filteredRows := filterDuplicatesKeepLatest(rows, objectKeyMap)
	if len(filteredRows) != len(rows)-2 {
		t.Errorf("Expected all right value rows, got %d", len(filteredRows))
	}

	for _, row := range filteredRows {
		if row["field_value"] != "right_value" {
			t.Errorf("Expected right value, got %s", row["field_value"])
		}
	}
}

func TestFilterDuplicatesKeepLatest(t *testing.T) {
	typeKeyMap := map[string]string{
		"User": "user_id",
		"Post": "post_id",
	}

	tests := []struct {
		name     string
		input    []map[string]interface{}
		expected []map[string]interface{}
	}{
		{
			name: "No duplicates",
			input: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "Post", "post_id": "A001", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
			},
			expected: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "Post", "post_id": "A001", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "Duplicate with later timestamp",
			input: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
			},
			expected: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "Duplicate with earlier timestamp",
			input: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
			},
			expected: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "Mixed types and duplicates",
			input: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "Post", "post_id": "A001", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 12, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "Post", "post_id": "B002", "timestamp": time.Date(2023, 6, 25, 13, 0, 0, 0, time.UTC)},
			},
			expected: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 12, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "Post", "post_id": "A001", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "Post", "post_id": "B002", "timestamp": time.Date(2023, 6, 25, 13, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "Invalid object type",
			input: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "Invalid", "invalid_id": "001", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
			},
			expected: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "Missing key field",
			input: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "User", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
			},
			expected: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "Missing timestamp",
			input: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "002"},
			},
			expected: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "Duplicate with identical timestamp",
			input: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "name": "Alice", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "name": "Bob", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
			},
			expected: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "name": "Alice", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "Multiple duplicates with mixed timestamps",
			input: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "name": "Alice", "timestamp": time.Date(2023, 6, 25, 10, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "name": "Bob", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "name": "Charlie", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "name": "David", "timestamp": time.Date(2023, 6, 25, 10, 30, 0, 0, time.UTC)},
			},
			expected: []map[string]interface{}{
				{OBJECT_TYPE_NAME_FIELD: "User", "user_id": "001", "name": "Bob", "timestamp": time.Date(2023, 6, 25, 11, 0, 0, 0, time.UTC)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterDuplicatesKeepLatest(tt.input, typeKeyMap)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("filterDuplicatesKeepLatest() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// [profile_id ProfileId AccountNumber AdditionalInformation PartyType BusinessName FirstName MiddleName LastName BirthDate Gender PhoneNumber MobilePhoneNumber HomePhoneNumber BusinessPhoneNumber EmailAddress BusinessEmailAddress PersonalEmailAddress Address.Address1 Address.Address2 Address.Address3 Address.Address4 Address.City Address.State Address.Country Address.PostalCode Address.Province ShippingAddress.Address1 ShippingAddress.Address2 ShippingAddress.Address3 ShippingAddress.Address4 ShippingAddress.City ShippingAddress.State ShippingAddress.Country ShippingAddress.PostalCode ShippingAddress.Province MailingAddress.Address1 MailingAddress.Address2 MailingAddress.Address3 MailingAddress.Address4 MailingAddress.City MailingAddress.State MailingAddress.Country MailingAddress.PostalCode MailingAddress.Province BillingAddress.Address1 BillingAddress.Address2 BillingAddress.Address3 BillingAddress.Address4 BillingAddress.City BillingAddress.State BillingAddress.Country BillingAddress.PostalCode BillingAddress.Province timestamp Attributes.traveler_id]
var EXPECTED_STANDARD_ATTR = []string{
	"connect_id",
	"ProfileId",
	"AccountNumber",
	"AdditionalInformation",
	"PartyType",
	"BusinessName",
	"FirstName",
	"MiddleName",
	"LastName",
	"BirthDate",
	"Gender",
	"PhoneNumber",
	"MobilePhoneNumber",
	"HomePhoneNumber",
	"BusinessPhoneNumber",
	"EmailAddress",
	"BusinessEmailAddress",
	"PersonalEmailAddress",
	"Address.Address1",
	"Address.Address2",
	"Address.Address3",
	"Address.Address4",
	"Address.City",
	"Address.State",
	"Address.Country",
	"Address.PostalCode",
	"Address.Province",
	"ShippingAddress.Address1",
	"ShippingAddress.Address2",
	"ShippingAddress.Address3",
	"ShippingAddress.Address4",
	"ShippingAddress.City",
	"ShippingAddress.State",
	"ShippingAddress.Country",
	"ShippingAddress.PostalCode",
	"ShippingAddress.Province",
	"MailingAddress.Address1",
	"MailingAddress.Address2",
	"MailingAddress.Address3",
	"MailingAddress.Address4",
	"MailingAddress.City",
	"MailingAddress.State",
	"MailingAddress.Country",
	"MailingAddress.PostalCode",
	"MailingAddress.Province",
	"BillingAddress.Address1",
	"BillingAddress.Address2",
	"BillingAddress.Address3",
	"BillingAddress.Address4",
	"BillingAddress.City",
	"BillingAddress.State",
	"BillingAddress.Country",
	"BillingAddress.PostalCode",
	"BillingAddress.Province",
	"timestamp",
}

var BuildWhereClauseTestData = []struct {
	name     string
	input    []BaseCriteria
	expected string
}{
	{
		name: "simple eq",
		input: []BaseCriteria{
			SearchCriterion{
				Column:   "name",
				Operator: EQ,
				Value:    "John",
			},
		},
		expected: `"name" = 'John'`,
	},
	{
		name: "multiple eq",
		input: []BaseCriteria{
			SearchCriterion{
				Column:   "name",
				Operator: EQ,
				Value:    "John",
			},
			SearchOperator{Operator: AND},
			SearchCriterion{
				Column:   "age",
				Operator: EQ,
				Value:    30,
			},
		},
		expected: `"name" = 'John' AND "age" = 30`,
	},
	{
		name: "grouped clause",
		input: []BaseCriteria{
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "name",
						Operator: EQ,
						Value:    "John",
					},
					SearchOperator{
						Operator: AND,
					},
					SearchCriterion{
						Column:   "age",
						Operator: NEQ,
						Value:    30,
					},
				},
			},
			SearchOperator{Operator: OR},
			SearchCriterion{
				Column:   "hobby",
				Operator: LIKE,
				Value:    "%gaming%",
			},
		},
		expected: `("name" = 'John' AND "age" != 30) OR "hobby" LIKE '%gaming%'`,
	},
	{
		name: "two grouped clause",
		input: []BaseCriteria{
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "name",
						Operator: EQ,
						Value:    "John",
					},
					SearchOperator{
						Operator: AND,
					},
					SearchCriterion{
						Column:   "age",
						Operator: EQ,
						Value:    30,
					},
				},
			},
			SearchOperator{Operator: OR},
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "hobby",
						Operator: LIKE,
						Value:    "%gaming%",
					},
					SearchOperator{
						Operator: AND,
					},
					SearchCriterion{
						Column:   "testScore",
						Operator: GT,
						Value:    80,
					},
				},
			},
		},
		expected: `("name" = 'John' AND "age" = 30) OR ("hobby" LIKE '%gaming%' AND "testScore" > 80)`,
	},
	{
		name: "nested clause beginning",
		input: []BaseCriteria{
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchGroup{
						Criteria: []BaseCriteria{
							SearchCriterion{
								Column:   "name",
								Operator: EQ,
								Value:    "John",
							},
							SearchOperator{Operator: AND},
							SearchCriterion{
								Column:   "age",
								Operator: EQ,
								Value:    30,
							},
						},
					},
					SearchOperator{Operator: OR},
					SearchCriterion{
						Column:   "hobby",
						Operator: LIKE,
						Value:    "%gaming%",
					},
				},
			},
			SearchOperator{Operator: AND},
			SearchCriterion{
				Column:   "testScore",
				Operator: GTE,
				Value:    80,
			},
		},
		expected: `(("name" = 'John' AND "age" = 30) OR "hobby" LIKE '%gaming%') AND "testScore" >= 80`,
	},
	{
		name: "nested clause middle",
		input: []BaseCriteria{
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "name",
						Operator: EQ,
						Value:    "John",
					},
					SearchOperator{Operator: AND},
					SearchGroup{
						Criteria: []BaseCriteria{
							SearchCriterion{
								Column:   "age",
								Operator: EQ,
								Value:    30,
							},
							SearchOperator{Operator: OR},
							SearchCriterion{
								Column:   "hobby",
								Operator: LIKE,
								Value:    "%gaming%",
							},
						},
					},
				},
			},
			SearchOperator{Operator: AND},
			SearchCriterion{
				Column:   "testScore",
				Operator: LT,
				Value:    80,
			},
		},
		expected: `("name" = 'John' AND ("age" = 30 OR "hobby" LIKE '%gaming%')) AND "testScore" < 80`,
	},
	{
		name: "nested clause end",
		input: []BaseCriteria{
			SearchCriterion{
				Column:   "name",
				Operator: EQ,
				Value:    "John",
			},
			SearchOperator{Operator: AND},
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "age",
						Operator: EQ,
						Value:    30,
					},
					SearchOperator{Operator: OR},
					SearchGroup{
						Criteria: []BaseCriteria{
							SearchCriterion{
								Column:   "hobby",
								Operator: LIKE,
								Value:    "%gaming%",
							},
							SearchOperator{Operator: AND},
							SearchCriterion{
								Column:   "testScore",
								Operator: LTE,
								Value:    80,
							},
						},
					},
				},
			},
		},
		expected: `"name" = 'John' AND ("age" = 30 OR ("hobby" LIKE '%gaming%' AND "testScore" <= 80))`,
	},
	{
		name: "basic where in",
		input: []BaseCriteria{
			SearchCriterion{
				Column:   "hobby",
				Operator: IN,
				Value:    []any{"videos", "golf"},
			},
		},
		expected: `"hobby" IN ('videos','golf')`,
	},
	{
		name: "nested clause with in",
		input: []BaseCriteria{
			SearchCriterion{
				Column:   "name",
				Operator: NEQ,
				Value:    "John",
			},
			SearchOperator{Operator: AND},
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "age",
						Operator: EQ,
						Value:    30,
					},
					SearchOperator{Operator: OR},
					SearchCriterion{
						Column:   "testScore",
						Operator: GTE,
						Value:    80,
					},
				},
			},
			SearchOperator{Operator: AND},
			SearchCriterion{
				Column:   "hobby",
				Operator: NOTIN,
				Value:    []any{"video gaming", "golf"},
			},
		},
		expected: `"name" != 'John' AND ("age" = 30 OR "testScore" >= 80) AND "hobby" NOT IN ('video gaming','golf')`,
	},
}

func TestBuildWhereClause(t *testing.T) {
	for _, tt := range BuildWhereClauseTestData {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildWhereClause(tt.input)
			if err != nil {
				t.Errorf("[%s][%s] unexpected error: %s", t.Name(), tt.name, err)
			}
			if result != tt.expected {
				t.Errorf("[%s][%s] expected %s, got %s", t.Name(), tt.name, tt.expected, result)
			}
		})
	}
}

var BuildParameterizedWhereClauseTestData = []struct {
	name           string
	input          []BaseCriteria
	expectedClause string
	expectedParams []interface{}
}{
	{
		name: "simple eq",
		input: []BaseCriteria{
			SearchCriterion{
				Column:   "name",
				Operator: EQ,
				Value:    "John",
			},
		},
		expectedClause: `"name" = ?`,
		expectedParams: []interface{}{"John"},
	},
	{
		name: "multiple eq",
		input: []BaseCriteria{
			SearchCriterion{
				Column:   "name",
				Operator: EQ,
				Value:    "John",
			},
			SearchOperator{Operator: AND},
			SearchCriterion{
				Column:   "age",
				Operator: EQ,
				Value:    30,
			},
		},
		expectedClause: `"name" = ? AND "age" = ?`,
		expectedParams: []interface{}{"John", 30},
	},
	{
		name: "grouped clause",
		input: []BaseCriteria{
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "name",
						Operator: EQ,
						Value:    "John",
					},
					SearchOperator{
						Operator: AND,
					},
					SearchCriterion{
						Column:   "age",
						Operator: NEQ,
						Value:    30,
					},
				},
			},
			SearchOperator{Operator: OR},
			SearchCriterion{
				Column:   "hobby",
				Operator: LIKE,
				Value:    "%gaming%",
			},
		},
		expectedClause: `("name" = ? AND "age" != ?) OR "hobby" LIKE ?`,
		expectedParams: []interface{}{"John", 30, "%gaming%"},
	},
	{
		name: "two grouped clause",
		input: []BaseCriteria{
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "name",
						Operator: EQ,
						Value:    "John",
					},
					SearchOperator{
						Operator: AND,
					},
					SearchCriterion{
						Column:   "age",
						Operator: EQ,
						Value:    30,
					},
				},
			},
			SearchOperator{Operator: OR},
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "hobby",
						Operator: LIKE,
						Value:    "%gaming%",
					},
					SearchOperator{
						Operator: AND,
					},
					SearchCriterion{
						Column:   "testScore",
						Operator: GT,
						Value:    80,
					},
				},
			},
		},
		expectedClause: `("name" = ? AND "age" = ?) OR ("hobby" LIKE ? AND "testScore" > ?)`,
		expectedParams: []interface{}{"John", 30, "%gaming%", 80},
	},
	{
		name: "nested clause beginning",
		input: []BaseCriteria{
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchGroup{
						Criteria: []BaseCriteria{
							SearchCriterion{
								Column:   "name",
								Operator: EQ,
								Value:    "John",
							},
							SearchOperator{Operator: AND},
							SearchCriterion{
								Column:   "age",
								Operator: EQ,
								Value:    30,
							},
						},
					},
					SearchOperator{Operator: OR},
					SearchCriterion{
						Column:   "hobby",
						Operator: LIKE,
						Value:    "%gaming%",
					},
				},
			},
			SearchOperator{Operator: AND},
			SearchCriterion{
				Column:   "testScore",
				Operator: GTE,
				Value:    80,
			},
		},
		expectedClause: `(("name" = ? AND "age" = ?) OR "hobby" LIKE ?) AND "testScore" >= ?`,
		expectedParams: []interface{}{"John", 30, "%gaming%", 80},
	},
	{
		name: "nested clause middle",
		input: []BaseCriteria{
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "name",
						Operator: EQ,
						Value:    "John",
					},
					SearchOperator{Operator: AND},
					SearchGroup{
						Criteria: []BaseCriteria{
							SearchCriterion{
								Column:   "age",
								Operator: EQ,
								Value:    30,
							},
							SearchOperator{Operator: OR},
							SearchCriterion{
								Column:   "hobby",
								Operator: LIKE,
								Value:    "%gaming%",
							},
						},
					},
				},
			},
			SearchOperator{Operator: AND},
			SearchCriterion{
				Column:   "testScore",
				Operator: LT,
				Value:    80,
			},
		},
		expectedClause: `("name" = ? AND ("age" = ? OR "hobby" LIKE ?)) AND "testScore" < ?`,
		expectedParams: []interface{}{"John", 30, "%gaming%", 80},
	},
	{
		name: "nested clause end",
		input: []BaseCriteria{
			SearchCriterion{
				Column:   "name",
				Operator: EQ,
				Value:    "John",
			},
			SearchOperator{Operator: AND},
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "age",
						Operator: EQ,
						Value:    30,
					},
					SearchOperator{Operator: OR},
					SearchGroup{
						Criteria: []BaseCriteria{
							SearchCriterion{
								Column:   "hobby",
								Operator: LIKE,
								Value:    "%gaming%",
							},
							SearchOperator{Operator: AND},
							SearchCriterion{
								Column:   "testScore",
								Operator: LTE,
								Value:    80,
							},
						},
					},
				},
			},
		},
		expectedClause: `"name" = ? AND ("age" = ? OR ("hobby" LIKE ? AND "testScore" <= ?))`,
		expectedParams: []interface{}{"John", 30, "%gaming%", 80},
	},
	{
		name: "basic where in",
		input: []BaseCriteria{
			SearchCriterion{
				Column:   "hobby",
				Operator: IN,
				Value:    []any{"videos", "golf"},
			},
		},
		expectedClause: `"hobby" IN (?,?)`,
		expectedParams: []interface{}{"videos", "golf"},
	},
	{
		name: "nested clause with in",
		input: []BaseCriteria{
			SearchCriterion{
				Column:   "name",
				Operator: NEQ,
				Value:    "John",
			},
			SearchOperator{Operator: AND},
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{
						Column:   "age",
						Operator: EQ,
						Value:    30,
					},
					SearchOperator{Operator: OR},
					SearchCriterion{
						Column:   "testScore",
						Operator: GTE,
						Value:    80,
					},
				},
			},
			SearchOperator{Operator: AND},
			SearchCriterion{
				Column:   "hobby",
				Operator: NOTIN,
				Value:    []any{"video gaming", "golf"},
			},
		},
		expectedClause: `"name" != ? AND ("age" = ? OR "testScore" >= ?) AND "hobby" NOT IN (?,?)`,
		expectedParams: []interface{}{"John", 30, 80, "video gaming", "golf"},
	},
}

func TestBuildParameterizedWhereClause(t *testing.T) {
	for _, tt := range BuildParameterizedWhereClauseTestData {
		t.Run(tt.name, func(t *testing.T) {
			result, params, err := buildParameterizedWhereClause(tt.input)
			if err != nil {
				t.Errorf("[%s][%s] unexpected error: %s", t.Name(), tt.name, err)
			}
			if result != tt.expectedClause {
				t.Errorf("[%s][%s] expected %s, got %s", t.Name(), tt.name, tt.expectedClause, result)
			}
			if !reflect.DeepEqual(params, tt.expectedParams) {
				t.Errorf("[%s][%s] expected %v, got %v", t.Name(), tt.name, tt.expectedParams, params)
			}
		})
	}
}
