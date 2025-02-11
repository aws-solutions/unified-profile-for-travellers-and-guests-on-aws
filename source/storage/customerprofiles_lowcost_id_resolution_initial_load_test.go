// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"encoding/json"
	"fmt"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
)

////////////////////////////////////////////////////////////////////////
// This test file is used to validate the SQL request created to identify all the impacted
// interaction records during the initial data load.
// the test validate the SQL query generated, ingest some interaction records and validate, run the query abd validate the records founds
// functional use cases can be added easily by adding data to the various constants: IR_FIRST_LOAD_TEST_CASES, MAPPINGS, EXPECTED_TIDS...
////////////////////////////////////////////////////////////////////////

var MAPPINGS = []ObjectMapping{
	{
		Name: "booking",
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
				Source:  "_source.loyalty_id",
				Target:  "loyalty_id",
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.booking_id",
				Target:  "booking_id",
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.test_booker_id",
				Target:  "test_booker_id",
				KeyOnly: true,
			},
			{
				Type:   "STRING",
				Source: "_source.email",
				Target: "_profile.PersonalEmailAddress",
			},
			{
				Type:   "STRING",
				Source: "_source.first_name",
				Target: "_profile.FirstName",
			},
			{
				Type:   "STRING",
				Source: "_source.last_name",
				Target: "_profile.LastName",
			},
		},
	},
	{
		Name: "clickstream",
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
				Source:  "_source.booking_id",
				Target:  "booking_id",
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
		Name: "loyalty",
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
				Source:  "_source.loyalty_id",
				Target:  "loyalty_id",
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
				Type:   "STRING",
				Source: "_source.email",
				Target: "_profile.PersonalEmailAddress",
			},
		},
	},
	{
		Name: "purchase",
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
				Source:  "_source.cc_name",
				Target:  "cc_name",
				KeyOnly: true,
			},
			{
				Type:   "STRING",
				Source: "_source.email",
				Target: "_profile.PersonalEmailAddress",
			},
		},
	},
}

var NORMALIZATION_SETTINGS = IndexNormalizationSettings{
	Lowercase: true,
	Trim:      true,
}

var IR_INITIAL_RULE_SET_1 = RuleSet{Rules: []Rule{
	//clickstream and booking record with same booking id in same time period
	{
		Index: 0,
		Conditions: []Condition{
			{
				Index:                0,
				ConditionType:        CONDITION_TYPE_SKIP,
				IncomingObjectType:   "booking",
				IncomingObjectField:  "test_booker_id",
				Op:                   RULE_OP_NOT_EQUALS,
				IncomingObjectType2:  "booking",
				IncomingObjectField2: "traveler_id",
				IndexNormalization:   NORMALIZATION_SETTINGS,
			},
			{
				Index:               1,
				ConditionType:       CONDITION_TYPE_MATCH,
				IncomingObjectType:  "booking",
				IncomingObjectField: "booking_id",
				ExistingObjectType:  "clickstream",
				ExistingObjectField: "booking_id",
				IndexNormalization:  NORMALIZATION_SETTINGS,
			},
			{
				Index:                2,
				ConditionType:        CONDITION_TYPE_FILTER,
				IncomingObjectType:   "booking",
				IncomingObjectField:  "timestamp",
				ExistingObjectType:   "clickstream",
				ExistingObjectField:  "timestamp",
				Op:                   RULE_OP_WITHIN_SECONDS,
				FilterConditionValue: IntValue(5),
				IndexNormalization:   NORMALIZATION_SETTINGS,
			},
		},
	},
	//loyalty id matches booking and profile
	{
		Index: 1,
		Conditions: []Condition{
			{
				Index:               0,
				ConditionType:       CONDITION_TYPE_MATCH,
				IncomingObjectType:  "booking",
				IncomingObjectField: "loyalty_id",
				Op:                  RULE_OP_NOT_EQUALS,
				ExistingObjectType:  "loyalty",
				ExistingObjectField: "loyalty_id",
				IndexNormalization:  NORMALIZATION_SETTINGS,
			},
		},
	},
	//2 bookings with same loyalty ID
	{
		Index: 2,
		Conditions: []Condition{
			{
				Index:               0,
				ConditionType:       CONDITION_TYPE_MATCH,
				IncomingObjectType:  "booking",
				IncomingObjectField: "loyalty_id",
				Op:                  RULE_OP_NOT_EQUALS,
				ExistingObjectType:  "booking",
				ExistingObjectField: "loyalty_id",
				IndexNormalization:  NORMALIZATION_SETTINGS,
			},
		},
	},
	//2 bookings with same loyalty ID where the guest is the booker
	{
		Index: 3,
		Conditions: []Condition{
			{
				Index:                0,
				ConditionType:        CONDITION_TYPE_SKIP,
				IncomingObjectType:   "booking",
				IncomingObjectField:  "test_booker_id",
				Op:                   RULE_OP_NOT_EQUALS,
				IncomingObjectType2:  "booking",
				IncomingObjectField2: "traveler_id",
				IndexNormalization:   NORMALIZATION_SETTINGS,
			},
			{
				Index:               0,
				ConditionType:       CONDITION_TYPE_MATCH,
				IncomingObjectType:  "booking",
				IncomingObjectField: "loyalty_id",
				Op:                  RULE_OP_NOT_EQUALS,
				ExistingObjectType:  "booking",
				ExistingObjectField: "loyalty_id",
				IndexNormalization:  NORMALIZATION_SETTINGS,
			},
		},
	},
},
}

var IR_INITIAL_RULE_SET_2 = RuleSet{Rules: []Rule{
	{
		//testing template logic
		Index: 0,
		Conditions: []Condition{
			{
				Index:               0,
				ConditionType:       CONDITION_TYPE_MATCH,
				IncomingObjectType:  PROFILE_OBJECT_TYPE_NAME,
				IncomingObjectField: "{{.FirstName}} {{.LastName}}",
				Op:                  RULE_OP_NOT_EQUALS,
				ExistingObjectType:  "purchase",
				ExistingObjectField: "cc_name",
				IndexNormalization:  NORMALIZATION_SETTINGS,
			},
		},
	},
},
}

type IRInitialLoadTestCase struct {
	RuleSet         RuleSet
	ExpectedQueries map[string]string
}

var IR_FIRST_LOAD_TEST_CASES = []IRInitialLoadTestCase{
	{
		RuleSet: IR_INITIAL_RULE_SET_1,
		ExpectedQueries: map[string]string{
			"3-booking":     `SELECT i."accp_object_id", i."traveler_id", i."loyalty_id", i."booking_id", i."test_booker_id", i."email", i."first_name", i."last_name", m."connect_id", m."ProfileId", m."AccountNumber", m."AdditionalInformation", m."PartyType", m."BusinessName", m."FirstName", m."MiddleName", m."LastName", m."BirthDate", m."Gender", m."PhoneNumber", m."MobilePhoneNumber", m."HomePhoneNumber", m."BusinessPhoneNumber", m."EmailAddress", m."BusinessEmailAddress", m."PersonalEmailAddress", m."Address.Address1", m."Address.Address2", m."Address.Address3", m."Address.Address4", m."Address.City", m."Address.State", m."Address.Country", m."Address.PostalCode", m."Address.Province", m."ShippingAddress.Address1", m."ShippingAddress.Address2", m."ShippingAddress.Address3", m."ShippingAddress.Address4", m."ShippingAddress.City", m."ShippingAddress.State", m."ShippingAddress.Country", m."ShippingAddress.PostalCode", m."ShippingAddress.Province", m."MailingAddress.Address1", m."MailingAddress.Address2", m."MailingAddress.Address3", m."MailingAddress.Address4", m."MailingAddress.City", m."MailingAddress.State", m."MailingAddress.Country", m."MailingAddress.PostalCode", m."MailingAddress.Province", m."BillingAddress.Address1", m."BillingAddress.Address2", m."BillingAddress.Address3", m."BillingAddress.Address4", m."BillingAddress.City", m."BillingAddress.State", m."BillingAddress.Country", m."BillingAddress.PostalCode", m."BillingAddress.Province", m."timestamp", m."Attributes.traveler_id" FROM ir_test_domain_booking i LEFT JOIN ir_test_domain_search m ON i.connect_id = m.connect_id WHERE i.test_booker_id = i.traveler_id AND i."loyalty_id" != ''`,
			"0-booking":     `SELECT i."accp_object_id", i."traveler_id", i."loyalty_id", i."booking_id", i."test_booker_id", i."email", i."first_name", i."last_name", m."connect_id", m."ProfileId", m."AccountNumber", m."AdditionalInformation", m."PartyType", m."BusinessName", m."FirstName", m."MiddleName", m."LastName", m."BirthDate", m."Gender", m."PhoneNumber", m."MobilePhoneNumber", m."HomePhoneNumber", m."BusinessPhoneNumber", m."EmailAddress", m."BusinessEmailAddress", m."PersonalEmailAddress", m."Address.Address1", m."Address.Address2", m."Address.Address3", m."Address.Address4", m."Address.City", m."Address.State", m."Address.Country", m."Address.PostalCode", m."Address.Province", m."ShippingAddress.Address1", m."ShippingAddress.Address2", m."ShippingAddress.Address3", m."ShippingAddress.Address4", m."ShippingAddress.City", m."ShippingAddress.State", m."ShippingAddress.Country", m."ShippingAddress.PostalCode", m."ShippingAddress.Province", m."MailingAddress.Address1", m."MailingAddress.Address2", m."MailingAddress.Address3", m."MailingAddress.Address4", m."MailingAddress.City", m."MailingAddress.State", m."MailingAddress.Country", m."MailingAddress.PostalCode", m."MailingAddress.Province", m."BillingAddress.Address1", m."BillingAddress.Address2", m."BillingAddress.Address3", m."BillingAddress.Address4", m."BillingAddress.City", m."BillingAddress.State", m."BillingAddress.Country", m."BillingAddress.PostalCode", m."BillingAddress.Province", m."timestamp", m."Attributes.traveler_id" FROM ir_test_domain_booking i LEFT JOIN ir_test_domain_search m ON i.connect_id = m.connect_id WHERE i.test_booker_id = i.traveler_id AND i."booking_id" != ''`,
			"1-booking":     `SELECT i."accp_object_id", i."traveler_id", i."loyalty_id", i."booking_id", i."test_booker_id", i."email", i."first_name", i."last_name", m."connect_id", m."ProfileId", m."AccountNumber", m."AdditionalInformation", m."PartyType", m."BusinessName", m."FirstName", m."MiddleName", m."LastName", m."BirthDate", m."Gender", m."PhoneNumber", m."MobilePhoneNumber", m."HomePhoneNumber", m."BusinessPhoneNumber", m."EmailAddress", m."BusinessEmailAddress", m."PersonalEmailAddress", m."Address.Address1", m."Address.Address2", m."Address.Address3", m."Address.Address4", m."Address.City", m."Address.State", m."Address.Country", m."Address.PostalCode", m."Address.Province", m."ShippingAddress.Address1", m."ShippingAddress.Address2", m."ShippingAddress.Address3", m."ShippingAddress.Address4", m."ShippingAddress.City", m."ShippingAddress.State", m."ShippingAddress.Country", m."ShippingAddress.PostalCode", m."ShippingAddress.Province", m."MailingAddress.Address1", m."MailingAddress.Address2", m."MailingAddress.Address3", m."MailingAddress.Address4", m."MailingAddress.City", m."MailingAddress.State", m."MailingAddress.Country", m."MailingAddress.PostalCode", m."MailingAddress.Province", m."BillingAddress.Address1", m."BillingAddress.Address2", m."BillingAddress.Address3", m."BillingAddress.Address4", m."BillingAddress.City", m."BillingAddress.State", m."BillingAddress.Country", m."BillingAddress.PostalCode", m."BillingAddress.Province", m."timestamp", m."Attributes.traveler_id" FROM ir_test_domain_booking i LEFT JOIN ir_test_domain_search m ON i.connect_id = m.connect_id WHERE i."loyalty_id" != ''`,
			"1-loyalty":     `SELECT i."accp_object_id", i."traveler_id", i."loyalty_id", i."from", i."arrival_date", i."email", m."connect_id", m."ProfileId", m."AccountNumber", m."AdditionalInformation", m."PartyType", m."BusinessName", m."FirstName", m."MiddleName", m."LastName", m."BirthDate", m."Gender", m."PhoneNumber", m."MobilePhoneNumber", m."HomePhoneNumber", m."BusinessPhoneNumber", m."EmailAddress", m."BusinessEmailAddress", m."PersonalEmailAddress", m."Address.Address1", m."Address.Address2", m."Address.Address3", m."Address.Address4", m."Address.City", m."Address.State", m."Address.Country", m."Address.PostalCode", m."Address.Province", m."ShippingAddress.Address1", m."ShippingAddress.Address2", m."ShippingAddress.Address3", m."ShippingAddress.Address4", m."ShippingAddress.City", m."ShippingAddress.State", m."ShippingAddress.Country", m."ShippingAddress.PostalCode", m."ShippingAddress.Province", m."MailingAddress.Address1", m."MailingAddress.Address2", m."MailingAddress.Address3", m."MailingAddress.Address4", m."MailingAddress.City", m."MailingAddress.State", m."MailingAddress.Country", m."MailingAddress.PostalCode", m."MailingAddress.Province", m."BillingAddress.Address1", m."BillingAddress.Address2", m."BillingAddress.Address3", m."BillingAddress.Address4", m."BillingAddress.City", m."BillingAddress.State", m."BillingAddress.Country", m."BillingAddress.PostalCode", m."BillingAddress.Province", m."timestamp", m."Attributes.traveler_id" FROM ir_test_domain_loyalty i LEFT JOIN ir_test_domain_search m ON i.connect_id = m.connect_id WHERE i."loyalty_id" != ''`,
			"2-booking":     `SELECT i."accp_object_id", i."traveler_id", i."loyalty_id", i."booking_id", i."test_booker_id", i."email", i."first_name", i."last_name", m."connect_id", m."ProfileId", m."AccountNumber", m."AdditionalInformation", m."PartyType", m."BusinessName", m."FirstName", m."MiddleName", m."LastName", m."BirthDate", m."Gender", m."PhoneNumber", m."MobilePhoneNumber", m."HomePhoneNumber", m."BusinessPhoneNumber", m."EmailAddress", m."BusinessEmailAddress", m."PersonalEmailAddress", m."Address.Address1", m."Address.Address2", m."Address.Address3", m."Address.Address4", m."Address.City", m."Address.State", m."Address.Country", m."Address.PostalCode", m."Address.Province", m."ShippingAddress.Address1", m."ShippingAddress.Address2", m."ShippingAddress.Address3", m."ShippingAddress.Address4", m."ShippingAddress.City", m."ShippingAddress.State", m."ShippingAddress.Country", m."ShippingAddress.PostalCode", m."ShippingAddress.Province", m."MailingAddress.Address1", m."MailingAddress.Address2", m."MailingAddress.Address3", m."MailingAddress.Address4", m."MailingAddress.City", m."MailingAddress.State", m."MailingAddress.Country", m."MailingAddress.PostalCode", m."MailingAddress.Province", m."BillingAddress.Address1", m."BillingAddress.Address2", m."BillingAddress.Address3", m."BillingAddress.Address4", m."BillingAddress.City", m."BillingAddress.State", m."BillingAddress.Country", m."BillingAddress.PostalCode", m."BillingAddress.Province", m."timestamp", m."Attributes.traveler_id" FROM ir_test_domain_booking i LEFT JOIN ir_test_domain_search m ON i.connect_id = m.connect_id WHERE i."loyalty_id" != ''`,
			"0-clickstream": `SELECT i."accp_object_id", i."traveler_id", i."booking_id", i."from", i."arrival_date", i."email", m."connect_id", m."ProfileId", m."AccountNumber", m."AdditionalInformation", m."PartyType", m."BusinessName", m."FirstName", m."MiddleName", m."LastName", m."BirthDate", m."Gender", m."PhoneNumber", m."MobilePhoneNumber", m."HomePhoneNumber", m."BusinessPhoneNumber", m."EmailAddress", m."BusinessEmailAddress", m."PersonalEmailAddress", m."Address.Address1", m."Address.Address2", m."Address.Address3", m."Address.Address4", m."Address.City", m."Address.State", m."Address.Country", m."Address.PostalCode", m."Address.Province", m."ShippingAddress.Address1", m."ShippingAddress.Address2", m."ShippingAddress.Address3", m."ShippingAddress.Address4", m."ShippingAddress.City", m."ShippingAddress.State", m."ShippingAddress.Country", m."ShippingAddress.PostalCode", m."ShippingAddress.Province", m."MailingAddress.Address1", m."MailingAddress.Address2", m."MailingAddress.Address3", m."MailingAddress.Address4", m."MailingAddress.City", m."MailingAddress.State", m."MailingAddress.Country", m."MailingAddress.PostalCode", m."MailingAddress.Province", m."BillingAddress.Address1", m."BillingAddress.Address2", m."BillingAddress.Address3", m."BillingAddress.Address4", m."BillingAddress.City", m."BillingAddress.State", m."BillingAddress.Country", m."BillingAddress.PostalCode", m."BillingAddress.Province", m."timestamp", m."Attributes.traveler_id" FROM ir_test_domain_clickstream i LEFT JOIN ir_test_domain_search m ON i.connect_id = m.connect_id WHERE i."booking_id" != ''`,
		},
	},
	{
		RuleSet: IR_INITIAL_RULE_SET_2,
		ExpectedQueries: map[string]string{
			"0-_profile": `SELECT m."connect_id", m."ProfileId", m."AccountNumber", m."AdditionalInformation", m."PartyType", m."BusinessName", m."FirstName", m."MiddleName", m."LastName", m."BirthDate", m."Gender", m."PhoneNumber", m."MobilePhoneNumber", m."HomePhoneNumber", m."BusinessPhoneNumber", m."EmailAddress", m."BusinessEmailAddress", m."PersonalEmailAddress", m."Address.Address1", m."Address.Address2", m."Address.Address3", m."Address.Address4", m."Address.City", m."Address.State", m."Address.Country", m."Address.PostalCode", m."Address.Province", m."ShippingAddress.Address1", m."ShippingAddress.Address2", m."ShippingAddress.Address3", m."ShippingAddress.Address4", m."ShippingAddress.City", m."ShippingAddress.State", m."ShippingAddress.Country", m."ShippingAddress.PostalCode", m."ShippingAddress.Province", m."MailingAddress.Address1", m."MailingAddress.Address2", m."MailingAddress.Address3", m."MailingAddress.Address4", m."MailingAddress.City", m."MailingAddress.State", m."MailingAddress.Country", m."MailingAddress.PostalCode", m."MailingAddress.Province", m."BillingAddress.Address1", m."BillingAddress.Address2", m."BillingAddress.Address3", m."BillingAddress.Address4", m."BillingAddress.City", m."BillingAddress.State", m."BillingAddress.Country", m."BillingAddress.PostalCode", m."BillingAddress.Province", m."timestamp", m."Attributes.traveler_id" FROM ir_test_domain_search m WHERE m."FirstName" != '' AND m."LastName" != ''`,
			"0-purchase": `SELECT i."accp_object_id", i."traveler_id", i."segment_id", i."from", i."cc_name", i."email", m."connect_id", m."ProfileId", m."AccountNumber", m."AdditionalInformation", m."PartyType", m."BusinessName", m."FirstName", m."MiddleName", m."LastName", m."BirthDate", m."Gender", m."PhoneNumber", m."MobilePhoneNumber", m."HomePhoneNumber", m."BusinessPhoneNumber", m."EmailAddress", m."BusinessEmailAddress", m."PersonalEmailAddress", m."Address.Address1", m."Address.Address2", m."Address.Address3", m."Address.Address4", m."Address.City", m."Address.State", m."Address.Country", m."Address.PostalCode", m."Address.Province", m."ShippingAddress.Address1", m."ShippingAddress.Address2", m."ShippingAddress.Address3", m."ShippingAddress.Address4", m."ShippingAddress.City", m."ShippingAddress.State", m."ShippingAddress.Country", m."ShippingAddress.PostalCode", m."ShippingAddress.Province", m."MailingAddress.Address1", m."MailingAddress.Address2", m."MailingAddress.Address3", m."MailingAddress.Address4", m."MailingAddress.City", m."MailingAddress.State", m."MailingAddress.Country", m."MailingAddress.PostalCode", m."MailingAddress.Province", m."BillingAddress.Address1", m."BillingAddress.Address2", m."BillingAddress.Address3", m."BillingAddress.Address4", m."BillingAddress.City", m."BillingAddress.State", m."BillingAddress.Country", m."BillingAddress.PostalCode", m."BillingAddress.Province", m."timestamp", m."Attributes.traveler_id" FROM ir_test_domain_purchase i LEFT JOIN ir_test_domain_search m ON i.connect_id = m.connect_id WHERE i."cc_name" != ''`,
		},
	},
}

func TestIRSelectQueries(t *testing.T) {
	t.Parallel()
	domain := "ir_test_domain"
	irh := IdentityResolutionHandler{
		Tx: core.NewTransaction(t.Name(), "", core.LogLevelDebug),
	}

	for tcindex, testCase := range IR_FIRST_LOAD_TEST_CASES {
		irh.Tx.Debug("*****************************")
		irh.Tx.Debug("* TEST CASE %d ", tcindex)
		irh.Tx.Debug("*****************************")
		irh.Tx.Debug("")

		ruleMap := map[string]Rule{}
		for _, rule := range testCase.RuleSet.Rules {
			ruleMap[fmt.Sprintf("%d", rule.Index)] = rule
		}
		ruleIdObjectPairs := createRuleObjectPairs(testCase.RuleSet)
		for _, pair := range ruleIdObjectPairs {
			sql := ""
			var err error
			if pair.ObjectType == PROFILE_OBJECT_TYPE_NAME {
				sql, err = irh.selectProfilesToSql(domain, ruleMap[pair.RuleID], MAPPINGS)
				if err != nil {
					t.Errorf("[test-%d] Error generating sql: %v", tcindex, err)
				}
			} else {
				sql, err = irh.selectInteractionProfilesToSql(domain, pair.ObjectType, ruleMap[pair.RuleID], MAPPINGS)
				if err != nil {
					t.Errorf("[test-%d] Error generating sql: %v", tcindex, err)
				}
			}
			expected := testCase.ExpectedQueries[pair.RuleID+"-"+pair.ObjectType]
			if expected == "" {
				t.Errorf("[test-%d] unexpected SQL query for rule %s and %s: %s", tcindex, pair.RuleID, pair.ObjectType, sql)
				continue
			}
			if sql != expected {
				t.Errorf(
					"[test-%d] invalid SQL query for rule %s and %s: %s, Got: %s",
					tcindex,
					pair.RuleID,
					pair.ObjectType,
					expected,
					sql,
				)
			}

		}
	}
}

// Note: here we make sure we have unique traveler ID to ensure unique profile_id/connect_id mappings
var TEST_BOOKINGS = []map[string]string{
	{"accp_object_id": "1", "traveler_id": "b1", "loyalty_id": "loy1", "booking_id": "bid1", "test_booker_id": "b1"},
	{"accp_object_id": "2", "traveler_id": "b2", "loyalty_id": "loy9", "booking_id": "bid2", "test_booker_id": "b1"},
	{"accp_object_id": "3", "traveler_id": "b3", "loyalty_id": "loy10", "booking_id": "bid3"},
	{"accp_object_id": "4", "traveler_id": "b4", "loyalty_id": "loy10", "booking_id": "bid4", "test_booker_id": "b4"},
	{"accp_object_id": "5", "traveler_id": "b5", "loyalty_id": "loy10", "booking_id": "bid5", "test_booker_id": "b5"},
	{"accp_object_id": "6", "traveler_id": "b6", "first_name": "John", "last_name": "doe"},
	{"accp_object_id": "7", "traveler_id": "b7", "first_name": "John", "test_booker_id": "b7"},
}
var TEST_LOYALTY = []map[string]string{
	{"accp_object_id": "1", "traveler_id": "l1", "loyalty_id": "loy1"},
	{"accp_object_id": "2", "traveler_id": "l2", "loyalty_id": "loy2"},
}
var TEST_CLICKSTREAM = []map[string]string{
	{"accp_object_id": "1", "traveler_id": "c1", "booking_id": "bid1"},
	{"accp_object_id": "2", "traveler_id": "c2", "booking_id": "bid9"},
	{"accp_object_id": "3", "traveler_id": "c3"},
}

var TEST_PURCHASE = []map[string]string{
	{"accp_object_id": "1", "traveler_id": "p1", "cc_name": "John Doe"},
	{"accp_object_id": "2", "traveler_id": "p2", "cc_name": "Jane Doe"},
	{"accp_object_id": "3", "traveler_id": "p3"},
}

var TEST_DATA = map[string][]map[string]string{
	"booking":     TEST_BOOKINGS,
	"loyalty":     TEST_LOYALTY,
	"clickstream": TEST_CLICKSTREAM,
	"purchase":    TEST_PURCHASE,
}

var EXPECTED_TIDS = map[string][]string{
	"0-3-booking":     {"b1", "b4", "b5"},
	"0-0-booking":     {"b1", "b4", "b5"},
	"0-1-booking":     {"b1", "b2", "b3", "b4", "b5"},
	"0-1-loyalty":     {"l1", "l2"},
	"0-2-booking":     {"b1", "b2", "b3", "b4", "b5"},
	"0-0-clickstream": {"c1", "c2"},
	"1-0-_profile":    {"b6"},
	"1-0-purchase":    {"p1", "p2"},
}

func TestIRTableInitLoad(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	irh := IdentityResolutionHandler{
		Tx: core.NewTransaction(t.Name(), "", core.LogLevelDebug),
	}
	//creation of domain
	// Create required resources
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[customerprofiles] Error setting up aurora: %+v ", err)
		return
	}
	dynamoTableName := "customer-profiles-ir-table-load-" + core.GenerateUniqueId()
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, "pk", "sk", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[TestInvalidDomainName] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[TestMultipleSearch] error deleting table: %v", err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[TestInvalidDomainName] error waiting for table creation: %v", err)
	}
	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)

	// Init low cost config
	mergeQueue := buildTestMergeQueue(t, envCfg.Region)
	cpWriterQueue := buildTestCPWriterQueue(t, envCfg.Region)
	initOptions := CustomerProfileInitOptions{MergeQueueClient: mergeQueue}
	lc := InitLowCost(
		envCfg.Region,
		auroraCfg,
		&dynamoCfg,
		kinesisMockCfg,
		cpWriterQueue,
		"tah/upt/source/tah-core/customerprofile",
		core.TEST_SOLUTION_ID,
		core.TEST_SOLUTION_VERSION,
		core.LogLevelDebug,
		&initOptions,
		InitLcsCache(),
	)

	t.Logf("Creating domain for multiple object types test")
	domainName := randomDomain("ir_table_test")
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{"envName": "dev"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[TestMultipleObjects] Error creating domain: %v", err)
	}
	t.Cleanup(func() {
		err = lc.DeleteDomainByName(domainName)
		if err != nil {
			t.Errorf("[TestMultipleObjects] error deleting domain: %v", err)
		}
	})

	for _, mapping := range MAPPINGS {
		err = lc.CreateMapping(mapping.Name, "Mapping for "+mapping.Name, mapping.Fields)
		if err != nil {
			t.Errorf("[TestHistory] error creating mapping: %v", err)
		}
	}

	for objType, objects := range TEST_DATA {
		for _, object := range objects {
			obj, _ := json.Marshal(object)
			err = lc.PutProfileObject(string(obj), objType)
			if err != nil {
				t.Errorf("[TestHistory] error creating profile object: %v", err)
			}
		}
	}

	/********************************
	*  Run all SQL queries
	***********************************/
	for tcindex, testCase := range IR_FIRST_LOAD_TEST_CASES {
		lc.Tx.Debug("*****************************")
		lc.Tx.Debug("* TEST CASE %d ", tcindex)
		lc.Tx.Debug("*****************************")
		lc.Tx.Debug("")

		ruleMap := map[string]Rule{}
		for _, rule := range testCase.RuleSet.Rules {
			ruleMap[fmt.Sprintf("%d", rule.Index)] = rule
		}
		ruleIdObjectPairs := createRuleObjectPairs(testCase.RuleSet)
		for _, pair := range ruleIdObjectPairs {
			sql := ""
			var err error
			if pair.ObjectType == PROFILE_OBJECT_TYPE_NAME {
				sql, err = irh.selectProfilesToSql(domainName, ruleMap[pair.RuleID], MAPPINGS)
				if err != nil {
					t.Errorf("[test-%d] Error generating sql: %v", tcindex, err)
				}
			} else {
				sql, err = irh.selectInteractionProfilesToSql(domainName, pair.ObjectType, ruleMap[pair.RuleID], MAPPINGS)
				if err != nil {
					t.Errorf("[test-%d] Error generating sql: %v", tcindex, err)
				}
			}
			data, err := lc.Data.AurSvc.Query(sql)
			if err != nil {
				t.Errorf("[test-%d] Error executing sql: %v", tcindex, err)
			}
			lc.Tx.Info("Result for sql query: %+v", data)
			expected := EXPECTED_TIDS[fmt.Sprint(tcindex)+"-"+pair.RuleID+"-"+pair.ObjectType]
			if len(expected) != len(data) {
				t.Errorf("[test-%d] Expected %d results, got %d", tcindex, len(expected), len(data))
			}
			tids := []string{}
			for _, rec := range data {
				cid, ok := rec["connect_id"].(string)
				if !ok {
					t.Errorf("[test-%d] Invalid connect_id type: %v", tcindex, rec["connect_id"])
				}
				//getting profile_id form connect ID
				data, err := lc.Data.AurSvc.Query(fmt.Sprintf(`SELECT "profile_id" FROM %s WHERE connect_id='%s'`, domainName, cid))
				if err != nil {
					t.Errorf("[test-%d] could not get profile id from connect id: %v", tcindex, err)
				}
				if len(data) != 1 {
					t.Errorf(
						"[test-%d] there should only be 1 profile_id for each connect_id for the purpose of this test: Have: %v",
						tcindex,
						data,
					)
				}
				tid, ok := data[0]["profile_id"].(string)
				if !ok {
					t.Errorf("[test-%d] Invalid profile_id type: %v", tcindex, rec["profile_id"])
				}
				tids = append(tids, tid)
			}
			if !core.IsEqual(expected, tids) {
				t.Errorf("[test-%d][rule-%s][%s] Expected %v, got %v", tcindex, pair.RuleID, pair.ObjectType, expected, tids)
			}
		}
	}
}
