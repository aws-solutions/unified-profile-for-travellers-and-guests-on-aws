// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"reflect"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"testing"
)

////////////////////////////////////////////////////////////////////////
// This test file is used to validate the SQL request created to index interactions into the IR table
// and to search for matches
// both th SQL query structure and their execution are validated
// for each test case, user can configure each ingestion step, the expected SQL, the expected IR table content and the expected matches
////////////////////////////////////////////////////////////////////////

var IR_RULE_SET_1 = RuleSet{Rules: []Rule{
	//clickstream and booking record with same booking id in same time period
	{
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "booking", IncomingObjectField: "booker_id", Op: RULE_OP_NOT_EQUALS, IncomingObjectType2: "booking", IncomingObjectField2: "traveler_id", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 1, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "booking", IncomingObjectField: "booking_id", ExistingObjectType: "clickstream", ExistingObjectField: "booking_id", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 2, ConditionType: CONDITION_TYPE_FILTER, IncomingObjectType: "booking", IncomingObjectField: "timestamp", ExistingObjectType: "clickstream", ExistingObjectField: "timestamp", Op: RULE_OP_WITHIN_SECONDS, FilterConditionValue: IntValue(5), IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	},
	//loyalty id matches booking and profile
	{
		Index: 1,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "booking", IncomingObjectField: "loyalty_id", Op: RULE_OP_NOT_EQUALS, ExistingObjectType: "loyalty", ExistingObjectField: "loyalty_id", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	},
	//2 bookings with same loyalty ID
	{
		Index: 2,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "booking", IncomingObjectField: "loyalty_id", Op: RULE_OP_NOT_EQUALS, ExistingObjectType: "booking", ExistingObjectField: "loyalty_id", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	},
	//2 bookings with same loyalty ID where the guest is the booker
	{
		Index: 3,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_SKIP, IncomingObjectType: "booking", IncomingObjectField: "booker_id", Op: RULE_OP_NOT_EQUALS, IncomingObjectType2: "booking", IncomingObjectField2: "traveler_id", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: "booking", IncomingObjectField: "loyalty_id", Op: RULE_OP_NOT_EQUALS, ExistingObjectType: "booking", ExistingObjectField: "loyalty_id", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	},
},
}

var IR_RULE_SET_2 = RuleSet{Rules: []Rule{
	{
		//testing template logic
		Index: 0,
		Conditions: []Condition{
			{Index: 0, ConditionType: CONDITION_TYPE_MATCH, IncomingObjectType: PROFILE_OBJECT_TYPE_NAME, IncomingObjectField: "{{.FirstName}} {{.LastName}}", Op: RULE_OP_NOT_EQUALS, ExistingObjectType: "purchase", ExistingObjectField: "cc_name", IndexNormalization: NORMALIZATION_SETTINGS_DEFAULT},
		},
	},
},
}

type IRTestCase struct {
	RuleSets   RuleSet
	TestEvents []TestEvents
}

type TestEvents struct {
	Profile              profilemodel.Profile
	ObjectType           string
	ObjectData           map[string]string
	ExpectedMatches      map[string][]string
	ExpectedTableColumns []string
	ExpectedTableContent [][]string
	ExpectedInsert       string
	ExpectedInsertArgs   []interface{}
	ExpectedSearch       string
	ExpectedSearchArgs   []interface{}
}

var IR_TEST_CASES = []IRTestCase{
	{
		RuleSets: IR_RULE_SET_1,
		TestEvents: []TestEvents{
			{
				Profile:    profilemodel.Profile{ProfileId: "cid1"},
				ObjectType: "clickstream",
				ObjectData: map[string]string{
					"booking_id":  "B123456789",
					"traveler_id": "s1",
					"timestamp":   "2018-01-01 00:00:40.000",
				},
				ExpectedMatches:      map[string][]string{},
				ExpectedTableColumns: []string{"connect_id", "rule", "index_value_1", "index_value_2", "timestamp"},
				ExpectedTableContent: [][]string{
					{"cid1", "0", "", "b123456789", "2018-01-01 00:00:40.000"},
				},
				ExpectedInsert:     `INSERT INTO ir_test_domain_rule_index (connect_id, rule, index_value_1, index_value_2, timestamp) VALUES ('cid1', '0', ?, ?, '2018-01-01 00:00:40.000') ON CONFLICT (connect_id, rule, index_value_1, index_value_2) DO NOTHING;`,
				ExpectedInsertArgs: []interface{}{"", "b123456789"},
				ExpectedSearch:     `SELECT connect_id, rule FROM ir_test_domain_rule_index WHERE ((rule = '0' AND (index_value_1 = ?) AND (("timestamp" >= timestamp '2018-01-01 00:00:35.000' AND "timestamp" <= timestamp '2018-01-01 00:00:45.000')))) AND NOT connect_id = 'cid1';`,
				ExpectedSearchArgs: []interface{}{"b123456789"},
			},
			//a booker
			{
				Profile:    profilemodel.Profile{ProfileId: "cid2"},
				ObjectType: "booking",
				ObjectData: map[string]string{
					"booking_id":  "B123456789",
					"traveler_id": "t1",
					"booker_id":   "t1",
					"timestamp":   "2018-01-01 00:00:43.000",
					"loyalty_id":  "123456789",
				},
				ExpectedMatches:      map[string][]string{"0": {"cid1"}},
				ExpectedTableColumns: []string{"connect_id", "rule", "index_value_1", "index_value_2", "timestamp"},
				ExpectedTableContent: [][]string{
					{"cid1", "0", "", "b123456789", "2018-01-01 00:00:40.000"},
					{"cid2", "0", "b123456789", "", "2018-01-01 00:00:43.000"},
					{"cid2", "1", "123456789", "", "2018-01-01 00:00:43.000"},
					{"cid2", "2", "123456789", "123456789", "2018-01-01 00:00:43.000"},
					{"cid2", "3", "123456789", "123456789", "2018-01-01 00:00:43.000"},
				},
				ExpectedInsert:     `INSERT INTO ir_test_domain_rule_index (connect_id, rule, index_value_1, index_value_2, timestamp) VALUES ('cid2', '0', ?, ?, '2018-01-01 00:00:43.000'), ('cid2', '1', ?, ?, '2018-01-01 00:00:43.000'), ('cid2', '2', ?, ?, '2018-01-01 00:00:43.000'), ('cid2', '3', ?, ?, '2018-01-01 00:00:43.000') ON CONFLICT (connect_id, rule, index_value_1, index_value_2) DO NOTHING;`,
				ExpectedInsertArgs: []interface{}{"b123456789", "", "123456789", "", "123456789", "123456789", "123456789", "123456789"},
				ExpectedSearch:     `SELECT connect_id, rule FROM ir_test_domain_rule_index WHERE ((rule = '0' AND (index_value_2 = ?) AND (("timestamp" >= timestamp '2018-01-01 00:00:38.000' AND "timestamp" <= timestamp '2018-01-01 00:00:48.000'))) OR (rule = '1' AND (index_value_2 = ?)) OR (rule = '2' AND (index_value_1 = ?)) OR (rule = '3' AND (index_value_1 = ?))) AND NOT connect_id = 'cid2';`,
				ExpectedSearchArgs: []interface{}{"b123456789", "123456789", "123456789", "123456789"},
			},
			//not a booker
			{
				Profile:    profilemodel.Profile{ProfileId: "cid3"},
				ObjectType: "booking",
				ObjectData: map[string]string{
					"booking_id":  "B123456789",
					"traveler_id": "t2",
					"booker_id":   "t1",
					"timestamp":   "2018-01-01 00:00:44.000",
					"loyalty_id":  "123456789",
				},
				ExpectedMatches:      map[string][]string{"2": {"cid2"}},
				ExpectedTableColumns: []string{"connect_id", "rule", "index_value_1", "index_value_2", "timestamp"},
				ExpectedTableContent: [][]string{
					{"cid1", "0", "", "b123456789", "2018-01-01 00:00:40.000"},
					{"cid2", "0", "b123456789", "", "2018-01-01 00:00:43.000"},
					{"cid2", "1", "123456789", "", "2018-01-01 00:00:43.000"},
					{"cid2", "2", "123456789", "123456789", "2018-01-01 00:00:43.000"},
					{"cid2", "3", "123456789", "123456789", "2018-01-01 00:00:43.000"},
					{"cid3", "1", "123456789", "", "2018-01-01 00:00:44.000"},
					{"cid3", "2", "123456789", "123456789", "2018-01-01 00:00:44.000"},
				},
				ExpectedInsert:     `INSERT INTO ir_test_domain_rule_index (connect_id, rule, index_value_1, index_value_2, timestamp) VALUES ('cid3', '1', ?, ?, '2018-01-01 00:00:44.000'), ('cid3', '2', ?, ?, '2018-01-01 00:00:44.000') ON CONFLICT (connect_id, rule, index_value_1, index_value_2) DO NOTHING;`,
				ExpectedInsertArgs: []interface{}{"123456789", "", "123456789", "123456789"},
				ExpectedSearch:     `SELECT connect_id, rule FROM ir_test_domain_rule_index WHERE ((rule = '1' AND (index_value_2 = ?)) OR (rule = '2' AND (index_value_1 = ?))) AND NOT connect_id = 'cid3';`,
				ExpectedSearchArgs: []interface{}{"123456789", "123456789"},
			},
			//loyalty matches
			{
				Profile:    profilemodel.Profile{ProfileId: "cid4"},
				ObjectType: "loyalty",
				ObjectData: map[string]string{
					"loyalty_id":  "123456789",
					"traveler_id": "t3",
					"points":      "120",
					"timestamp":   "2018-01-01 00:00:45.000",
				},
				ExpectedMatches:      map[string][]string{"1": {"cid2", "cid3"}},
				ExpectedTableColumns: []string{"connect_id", "rule", "index_value_1", "index_value_2", "timestamp"},
				ExpectedTableContent: [][]string{
					{"cid1", "0", "", "b123456789", "2018-01-01 00:00:40.000"},
					{"cid2", "0", "b123456789", "", "2018-01-01 00:00:43.000"},
					{"cid2", "1", "123456789", "", "2018-01-01 00:00:43.000"},
					{"cid2", "2", "123456789", "123456789", "2018-01-01 00:00:43.000"},
					{"cid2", "3", "123456789", "123456789", "2018-01-01 00:00:43.000"},
					{"cid3", "1", "123456789", "", "2018-01-01 00:00:44.000"},
					{"cid3", "2", "123456789", "123456789", "2018-01-01 00:00:44.000"},
					{"cid4", "1", "", "123456789", "2018-01-01 00:00:45.000"},
				},
				ExpectedInsert:     `INSERT INTO ir_test_domain_rule_index (connect_id, rule, index_value_1, index_value_2, timestamp) VALUES ('cid4', '1', ?, ?, '2018-01-01 00:00:45.000') ON CONFLICT (connect_id, rule, index_value_1, index_value_2) DO NOTHING;`,
				ExpectedInsertArgs: []interface{}{"", "123456789"},
				ExpectedSearch:     `SELECT connect_id, rule FROM ir_test_domain_rule_index WHERE ((rule = '1' AND (index_value_1 = ?))) AND NOT connect_id = 'cid4';`,
				ExpectedSearchArgs: []interface{}{"123456789"},
			},
		},
	},
	{
		RuleSets: IR_RULE_SET_2,
		TestEvents: []TestEvents{
			{
				Profile:    profilemodel.Profile{ProfileId: "cid1", FirstName: "John", LastName: "Doe"},
				ObjectType: "booking",
				ObjectData: map[string]string{
					"booking_id":  "B123456789",
					"traveler_id": "s1",
					"timestamp":   "2018-01-01 00:00:40.000",
				},
				ExpectedMatches:      map[string][]string{},
				ExpectedTableColumns: []string{"connect_id", "rule", "index_value_1", "index_value_2", "timestamp"},
				ExpectedTableContent: [][]string{
					{"cid1", "0", "john doe", "", "2018-01-01 00:00:40.000"},
				},
				ExpectedInsert:     `INSERT INTO ir_test_domain_rule_index (connect_id, rule, index_value_1, index_value_2, timestamp) VALUES ('cid1', '0', ?, ?, '2018-01-01 00:00:40.000') ON CONFLICT (connect_id, rule, index_value_1, index_value_2) DO NOTHING;`,
				ExpectedInsertArgs: []interface{}{"john doe", ""},
				ExpectedSearch:     `SELECT connect_id, rule FROM ir_test_domain_rule_index WHERE ((rule = '0' AND (index_value_2 = ?))) AND NOT connect_id = 'cid1';`,
				ExpectedSearchArgs: []interface{}{"john doe"},
			},
			{
				Profile:    profilemodel.Profile{ProfileId: "cid2"},
				ObjectType: "purchase",
				ObjectData: map[string]string{
					"cc_name":     "John Doe",
					"traveler_id": "s1",
					"timestamp":   "2018-01-01 00:00:41.000",
				},
				ExpectedMatches:      map[string][]string{"0": {"cid1"}},
				ExpectedTableColumns: []string{"connect_id", "rule", "index_value_1", "index_value_2", "timestamp"},
				ExpectedTableContent: [][]string{
					{"cid1", "0", "john doe", "", "2018-01-01 00:00:40.000"},
					{"cid2", "0", "", "john doe", "2018-01-01 00:00:41.000"},
				},
				ExpectedInsert:     `INSERT INTO ir_test_domain_rule_index (connect_id, rule, index_value_1, index_value_2, timestamp) VALUES ('cid2', '0', ?, ?, '2018-01-01 00:00:41.000') ON CONFLICT (connect_id, rule, index_value_1, index_value_2) DO NOTHING;`,
				ExpectedInsertArgs: []interface{}{"", "john doe"},
				ExpectedSearch:     `SELECT connect_id, rule FROM ir_test_domain_rule_index WHERE ((rule = '0' AND (index_value_1 = ?))) AND NOT connect_id = 'cid2';`,
				ExpectedSearchArgs: []interface{}{"john doe"},
			},
		},
	},
}

func TestIRSqlRequestGeneration(t *testing.T) {
	t.Parallel()
	domain := "ir_test_domain"
	irh := IdentityResolutionHandler{
		Tx: core.NewTransaction(t.Name(), "", core.LogLevelDebug),
	}

	for tcindex, testCase := range IR_TEST_CASES {
		irh.Tx.Debug("*****************************")
		irh.Tx.Debug("* TEST CASE %d ", tcindex)
		irh.Tx.Debug("*****************************")
		irh.Tx.Debug("")

		for _, evt := range testCase.TestEvents {
			indexRecords := irh.createRuleSetIndex(evt.Profile, evt.ObjectType, evt.ObjectData, testCase.RuleSets)
			sql, args, err := irh.createFindMatchSql(domain, indexRecords, evt.Profile.ProfileId)
			if err != nil {
				t.Errorf("Error creating find match SQL %+v ", err)
			}
			if evt.ExpectedSearch != "" && sql != evt.ExpectedSearch {
				t.Errorf("Invalid Search SQL! have '%s' and should be: '%s' ", sql, evt.ExpectedSearch)
			}
			if evt.ExpectedSearchArgs != nil && !reflect.DeepEqual(args, evt.ExpectedSearchArgs) {
				t.Errorf("Invalid Args! have '%s' and should be: '%s' ", args, evt.ExpectedSearchArgs)
			}
			insertSql, args := irh.createIndexSql(domain, indexRecords)
			if evt.ExpectedInsert != "" && insertSql != evt.ExpectedInsert {
				t.Errorf("Invalid Insert SQL! have '%s' and should be: '%s' ", insertSql, evt.ExpectedInsert)
			}
			if evt.ExpectedInsertArgs != nil && !reflect.DeepEqual(args, evt.ExpectedInsertArgs) {
				t.Errorf("Invalid Args! have '%s' and should be: '%s' ", args, evt.ExpectedInsertArgs)
			}
		}
	}
}

func TestIRTable(t *testing.T) {
	t.Parallel()
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Errorf("[lc_cp_ir_table] Error setting up aurora: %+v ", err)
		return
	}
	domain := "ir_test_domain"
	irh := IdentityResolutionHandler{
		AurSvc: auroraCfg,
		Tx:     core.NewTransaction(t.Name(), "", core.LogLevelDebug),
	}
	conn, err := auroraCfg.AcquireConnection(irh.Tx)
	if err != nil {
		t.Errorf("[lc_cp_ir_table] Error acquiring connection: %+v ", err)
		return
	}
	defer conn.Release()

	for i, testCase := range IR_TEST_CASES {
		irh.Tx.Debug("*****************************")
		irh.Tx.Debug("* TEST CASE %d ", i)
		irh.Tx.Debug("*****************************")
		irh.Tx.Debug("")
		err = irh.CreateIRTable(domain)
		if err != nil {
			t.Errorf("[lc_cp_ir_table] Error creating IR table: %+v ", err)
		}
		for j, evt := range testCase.TestEvents {
			matches, err := irh.FindMatches(conn, domain, evt.ObjectType, evt.ObjectData, evt.Profile, testCase.RuleSets)
			if err != nil {
				t.Errorf("[test_%d] Error finding match for profile %s and object type %s from event %d", i, evt.Profile.ProfileId, evt.ObjectType, j)
				break
			}
			matchMap := map[string][]string{}
			for _, match := range matches {
				matchMap[match.RuleIndex] = match.MatchIDs
			}
			if len(evt.ExpectedMatches) == 0 && len(matchMap) > 0 {
				t.Errorf("[test_%d] Invalid match for profile %s and object type %s for event %d have %+v and should have none", i, evt.Profile.ProfileId, evt.ObjectType, j, matchMap)
			}
			for rule, connectIds := range evt.ExpectedMatches {
				if !core.IsEqual(matchMap[rule], connectIds) {
					t.Errorf("[test_%d] Invalid match for profile %s and object type %s for event %d and rule %s have %s and should be %s", i, evt.Profile.ProfileId, evt.ObjectType, j, rule, matchMap[rule], connectIds)
				}
			}
			err = irh.IndexInIRTable(conn, domain, evt.ObjectType, evt.ObjectData, evt.Profile, testCase.RuleSets)
			if err != nil {
				t.Errorf("[test_%d] Error indexing profile %s and object type %s from event %d: %v", i, evt.Profile.ProfileId, evt.ObjectType, j, err)
				break
			}
			items, err := irh.ShowIRTable(domain, 10, 100)
			if err != nil {
				t.Errorf("[test_%d] Error showing IR table: %v", i, err)
			}
			rows := auroraResToRowArray(items, evt.ExpectedTableColumns)
			if len(rows) != len(evt.ExpectedTableContent) {
				t.Errorf("[test_%d] Invalid IR table content for profile %s and object type %s for event %d. have %d  rows and should be %d", i, evt.Profile.ProfileId, evt.ObjectType, j, len(rows), len(evt.ExpectedTableContent))
				break
			} else {
				for k, row := range rows {
					expected := evt.ExpectedTableContent[k]
					if !core.IsEqual(row, expected) {
						t.Errorf("[test_%d] Invalid IR table content for profile %s and object type %s for event %d. have %+v and should be %+v", i, evt.Profile.ProfileId, evt.ObjectType, j, row, expected)
						break
					}
				}
			}

		}

		err := irh.EmptyIRTable(domain)
		if err != nil {
			t.Errorf("[lc_cp_ir_table] Error emptying IR table: %+v ", err)
		}
		items, err := irh.ShowIRTable(domain, 10, 100)
		if err != nil {
			t.Errorf("[lc_cp_ir_table] Error showing IR table: %+v ", err)
		}
		if len(items) > 0 {
			t.Errorf("[lc_cp_ir_table] IR table should be empty after TRUNCATE!")
		}

		err = irh.DeleteIRTable(domain)
		if err != nil {
			t.Errorf("[lc_cp_ir_table] Error deleting IR table: %+v ", err)
		}
	}
}
