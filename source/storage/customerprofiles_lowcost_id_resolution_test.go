// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

type genericTestProfileObject struct {
	ObjectID    string    `json:"accp_object_id"`
	TravelerID  string    `json:"traveler_id"`
	LoyaltyID   string    `json:"loyalty_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	BookingID   string    `json:"booking_id"`
	BookerID    string    `json:"booker_id"`
	Amount      string    `json:"amount"`
	LastUpdated time.Time `json:"last_updated"`
}

type idrTestExpectedOutput struct {
	SearchKey    string
	SearchValue  string
	ProfileCount int
}

const genericTestObjectType_Booking string = "test_booking"

// Test Case: basic match on email
var profiles_EmailMatch = []genericTestProfileObject{
	{"obj_id_1", "traveler_1", "loyalty_1", "John", "Doe", "john@example.com", "booking_1", "traveler_1", "100.00", utcNow},
	{"obj_id_2", "traveler_2", "", "John", "Doe", "john@example.com", "booking_2", "", "250.00", utcNow},
}
var rules_EmailMatch = []Rule{
	{
		Index: 0,
		Conditions: []Condition{
			{
				Index:               0,
				ConditionType:       CONDITION_TYPE_MATCH,
				IncomingObjectType:  genericTestObjectType_Booking,
				IncomingObjectField: "email",
				Op:                  RULE_OP_EQUALS,
				ExistingObjectType:  genericTestObjectType_Booking,
				ExistingObjectField: "email",
			},
		},
	},
}
var expected_EmailMatch = []idrTestExpectedOutput{
	{
		SearchKey:    "profile_id",
		SearchValue:  "traveler_1",
		ProfileCount: 2,
	},
	{
		SearchKey:    "profile_id",
		SearchValue:  "traveler_2",
		ProfileCount: 2,
	},
}

// Test Case: multiple matches on email
var profiles_MultipleEmailMatches = []genericTestProfileObject{
	{"obj_id_1", "traveler_1", "", "", "", "john@example.com", "", "", "", utcNow},
	{"obj_id_2", "traveler_2", "", "", "", "john@example.com", "", "", "", utcNow},
	{"obj_id_3", "traveler_3", "", "", "", "john@example.com", "", "", "", utcNow},
	{"obj_id_4", "traveler_4", "", "John", "Doe", "john@example.com", "", "", "", utcNow},
	{"obj_id_5", "traveler_5", "", "", "", "john@example.com", "", "", "", utcNow},
	{"obj_id_6", "traveler_6", "", "", "", "john@example.com", "", "", "", utcNow},
}
var expected_MultipleEmailMatches = []idrTestExpectedOutput{
	{
		SearchKey:    "profile_id",
		SearchValue:  "traveler_1",
		ProfileCount: 6,
	},
}

// Test Case: complex
// skip empty name, match on email and loyalty
var profiles_Complex = []genericTestProfileObject{
	// john@example.com should not match, empty name is skipped
	{"obj_id_1", "traveler_1", "", "John", "Doe", "john@example.com", "", "traveler_1", "100.00", utcNow},
	{"obj_id_7", "traveler_7", "", "", "", "john@example.com", "booking_7", "traveler_7", "350.00", utcNow},
	// loyalty_2 should match
	{"obj_id_2", "traveler_2", "loyalty_2", "Martha", "Rivera", "martha@example.com", "booking_2", "traveler_2", "249.99", utcNow},
	{"obj_id_8", "traveler_8", "loyalty_2", "Martha", "Rivera", "", "booking_8", "traveler_8", "249.99", utcNow},
	// juan@example.com should match
	{"obj_id_3", "traveler_3", "", "Juan", "Perez", "juan@example.com", "booking_3", "traveler_2", "200.00", utcNow},
	{"obj_id_4", "traveler_4", "", "Juan", "Perez", "juan@example.com", "booking_4", "traveler_4", "150.00", utcNow},
	// no matches
	{"obj_id_5", "traveler_5", "", "Jorge", "Gonzalez", "jorge@example.com", "booking_5", "traveler_5", "250.00", utcNow},
	{"obj_id_6", "traveler_6", "", "Ana", "Martinez", "ana@example.com", "booking_6", "traveler_6", "300.00", utcNow},
}
var rules_Complex = []Rule{
	{
		// skip empty name, match on email
		Index: 0,
		Conditions: []Condition{
			{
				Index:               0,
				ConditionType:       CONDITION_TYPE_SKIP,
				IncomingObjectType:  genericTestObjectType_Booking,
				IncomingObjectField: "first_name",
				Op:                  RULE_OP_EQUALS_VALUE,
				SkipConditionValue:  StringVal(""),
			},
			{
				Index:               1,
				ConditionType:       CONDITION_TYPE_MATCH,
				IncomingObjectType:  genericTestObjectType_Booking,
				IncomingObjectField: "email",
				Op:                  RULE_OP_EQUALS,
				ExistingObjectType:  genericTestObjectType_Booking,
				ExistingObjectField: "email",
			},
		},
	},
	{
		// skip empty name, match on loyalty_id
		Index: 1,
		Conditions: []Condition{
			{
				Index:               0,
				ConditionType:       CONDITION_TYPE_SKIP,
				IncomingObjectType:  genericTestObjectType_Booking,
				IncomingObjectField: "first_name",
				Op:                  RULE_OP_EQUALS_VALUE,
				SkipConditionValue:  StringVal(""),
			},
			{
				Index:               2,
				ConditionType:       CONDITION_TYPE_MATCH,
				IncomingObjectType:  genericTestObjectType_Booking,
				IncomingObjectField: "loyalty_id",
				Op:                  RULE_OP_EQUALS,
				ExistingObjectType:  genericTestObjectType_Booking,
				ExistingObjectField: "loyalty_id",
			},
		},
	},
}
var expected_Complex = []idrTestExpectedOutput{
	{
		SearchKey:    "profile_id",
		SearchValue:  "traveler_1",
		ProfileCount: 1,
	},
	{
		SearchKey:    "profile_id",
		SearchValue:  "traveler_7",
		ProfileCount: 1,
	},
	{
		SearchKey:    "profile_id",
		SearchValue:  "traveler_2",
		ProfileCount: 2,
	},
	{
		SearchKey:    "profile_id",
		SearchValue:  "traveler_3",
		ProfileCount: 2,
	},
	{
		SearchKey:    "profile_id",
		SearchValue:  "traveler_5",
		ProfileCount: 1,
	},
	{
		SearchKey:    "profile_id",
		SearchValue:  "traveler_6",
		ProfileCount: 1,
	},
}

// Test Case: match after merge
// Two profiles initially match on email address, then merge occurs.
// After the merge, the new combined profile matches on FirstName + LastName.
var profiles_MatchAfterMerge = []genericTestProfileObject{
	{"obj_id_1", "traveler_1", "", "John", "", "john@example.com", "booking_1", "", "100.00", utcNow},
	{"obj_id_2", "traveler_2", "", "", "Doe", "john@example.com", "booking_2", "", "250.00", utcNow},
	{"obj_id_3", "traveler_3", "", "John", "Doe", "", "booking_3", "", "200.00", utcNow},
}
var rules_MatchAfterMerge = []Rule{
	{
		Index: 0,
		Conditions: []Condition{
			{
				Index:               0,
				ConditionType:       CONDITION_TYPE_MATCH,
				IncomingObjectType:  genericTestObjectType_Booking,
				IncomingObjectField: "email",
				ExistingObjectType:  genericTestObjectType_Booking,
				ExistingObjectField: "email",
			},
		},
	},
	{
		Index: 1,
		Conditions: []Condition{
			{
				Index:               0,
				ConditionType:       CONDITION_TYPE_MATCH,
				IncomingObjectType:  PROFILE_OBJECT_TYPE_NAME,
				IncomingObjectField: "FirstName",
				Op:                  RULE_OP_EQUALS,
				ExistingObjectType:  PROFILE_OBJECT_TYPE_NAME,
				ExistingObjectField: "FirstName",
			},
			{
				Index:               1,
				ConditionType:       CONDITION_TYPE_MATCH,
				IncomingObjectType:  PROFILE_OBJECT_TYPE_NAME,
				IncomingObjectField: "LastName",
				Op:                  RULE_OP_EQUALS,
				ExistingObjectType:  PROFILE_OBJECT_TYPE_NAME,
				ExistingObjectField: "LastName",
			},
		},
	},
}
var expected_MatchAfterMerge = []idrTestExpectedOutput{
	{
		SearchKey:    "profile_id",
		SearchValue:  "traveler_1",
		ProfileCount: 3,
	},
}

var testCases = []struct {
	name     string
	objects  []genericTestProfileObject
	rules    []Rule
	expected []idrTestExpectedOutput
}{
	{"basic email match", profiles_EmailMatch, rules_EmailMatch, expected_EmailMatch},
	{"email match with multiple merges", profiles_MultipleEmailMatches, rules_EmailMatch, expected_MultipleEmailMatches},
	{"complex", profiles_Complex, rules_Complex, expected_Complex},
	{"second match after merge", profiles_MatchAfterMerge, rules_MatchAfterMerge, expected_MatchAfterMerge},
}

func TestIRRuleBased(t *testing.T) {
	for i, tc := range testCases {
		tc := tc // quirk we need to deal with until we use Go 1.22 https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/loopclosure
		i := i
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Test Setup
			testPostFix := fmt.Sprintf(
				"%v_%d",
				strings.ToLower(core.GenerateUniqueId()),
				i,
			) // avoid potential clash in unique id (due to same underlying timestamp)
			domainName := "rule_based_" + testPostFix
			lc := ruleBasedTestSetup(t, testPostFix, domainName)

			err := lc.SaveIdResRuleSet(tc.rules)
			if err != nil {
				t.Fatalf("[%s] Error saving rule set: %v", tc.name, err)
			}

			err = lc.ActivateIdResRuleSet()
			if err != nil {
				t.Fatalf("[%s] Error activating rule set: %v", tc.name, err)
			}

			jsonRecs := []string{}
			for _, obj := range tc.objects {
				jsonRec, err := json.Marshal(obj)
				if err != nil {
					t.Fatalf("[%s] Error marshaling object: %v", tc.name, err)
				}
				jsonRecs = append(jsonRecs, string(jsonRec))
			}
			ingestTestRecords(t, lc, jsonRecs)
			processMerges(t, lc)

			lcs := lc.(*CustomerProfileConfigLowCost) // ideally we stick to interface, but for testing the record count we need access to the LCS db
			countByConnectId, err := testOnlyGetCount(
				lcs,
				domainName,
			) // domainName is also the name of master table, which we use to get the count of profiles
			if err != nil {
				t.Errorf("[%s] Error getting profile count: %v", tc.name, err)
			}
			for _, expected := range tc.expected {
				profiles, err := lc.SearchProfiles(expected.SearchKey, []string{expected.SearchValue})
				if err != nil {
					t.Errorf("[%s] Error searching profiles: %v", tc.name, err)
					continue
				}
				if len(profiles) == 0 {
					t.Errorf("[%s] No profiles found for %s", tc.name, expected.SearchValue)
					continue
				}

				connectId := profiles[0].ProfileId
				if countByConnectId[connectId] != expected.ProfileCount {
					t.Errorf(
						"[%s] Expected profile count: %d, actual profile count: %d",
						tc.name,
						expected.ProfileCount,
						countByConnectId[connectId],
					)
				}
			}
		})
	}
}

// Set up LCS and associated resources. t.Cleanup handles deleting resources
// after the test has finished running.
//
// Returns ICustomerProfileConfig (or fails the test immediately in the event of an error)
func ruleBasedTestSetup(t *testing.T, testPostFix string, domainName string) ICustomerProfileLowCostConfig {
	// SQS merge queue
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	sqsCfg := sqs.Init(envCfg.Region, "", "")
	_, err = sqsCfg.Create("rule-based-id-res-" + testPostFix)
	if err != nil {
		t.Fatalf("[%s] error creating queue: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = sqsCfg.Delete()
		if err != nil {
			t.Errorf("[%s] error deleting queue: %v", t.Name(), err)
		}
	})
	// Aurora
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[%s] Error setting up aurora: %+v ", t.Name(), err)
	}
	dynamoTableName := "rule-based-id-res-" + testPostFix
	// DynamoDB config table
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] error deleting table: %v", t.Name(), err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] error waiting for table creation: %v", t.Name(), err)
	}
	// Kinesis profile change event stream
	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)
	// CP Writer queue
	cpWriterQueue := buildTestCPWriterQueue(t, envCfg.Region)
	// Init low cost config
	options := CustomerProfileInitOptions{
		MergeQueueClient: &sqsCfg,
	}
	lc := InitLowCost(
		envCfg.Region,
		auroraCfg,
		&dynamoCfg,
		kinesisMockCfg,
		cpWriterQueue,
		"tah/upt/source/tah-core/customerprofiles",
		core.TEST_SOLUTION_ID,
		core.TEST_SOLUTION_VERSION,
		core.LogLevelDebug,
		&options,
		InitLcsCache(),
	)

	// Create domain
	err = lc.CreateDomainWithQueue(
		domainName,
		"",
		map[string]string{DOMAIN_CONFIG_ENV_KEY: "ruleBased"},
		"",
		"",
		DomainOptions{RuleBaseIdResolutionOn: true},
	)
	if err != nil {
		t.Fatalf("[%s] error creating domain: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = lc.DeleteDomain()
		if err != nil {
			t.Errorf("[%s] error deleting domain: %v", t.Name(), err)
		}
	})

	// Create mapping
	err = lc.CreateMapping(genericTestObjectType_Booking, "test booking", generateTestMappings())
	if err != nil {
		t.Fatalf("[%s] error creating mapping: %v", t.Name(), err)
	}

	return lc
}

func ingestTestRecords(t *testing.T, lc ICustomerProfileLowCostConfig, recs []string) {
	// Send ingest profile
	for _, rec := range recs {
		err := lc.PutProfileObject(rec, genericTestObjectType_Booking)
		if err != nil {
			t.Errorf("[%s] error putting rec: %v", t.Name(), err)
		}
	}
}

func processMerges(t *testing.T, cfg ICustomerProfileLowCostConfig) {
	// We need LCS specific functionality, so we must use the specific config
	lc := cfg.(*CustomerProfileConfigLowCost)
	// Poll SQS for two minutes or until we get repeated empty message responses
	timeout := 120
	max := timeout / 5
	emptyMessageCounter := 0
	emptyMessageMax := 3
	for i := 0; i < max; i++ {
		messages, err := lc.MergeQueueClient.Get(sqs.GetMessageOptions{VisibilityTimeout: 1800})
		// Exit early if error occurs
		if err != nil {
			t.Errorf("[%s] error getting messages: %v", t.Name(), err)
			break
		}
		// Check for empty message response, return early if we can
		if len(messages.Peek) == 0 {
			emptyMessageCounter++
		} else {
			emptyMessageCounter = 0
		}
		if emptyMessageCounter == emptyMessageMax {
			break
		}
		// Process messages
		for _, msg := range messages.Peek {
			var mergeRq MergeRequest
			err = json.Unmarshal([]byte(msg.Body), &mergeRq)
			if err != nil {
				t.Errorf("[%s] error unmarshalling merge request: %v", t.Name(), err)
				continue
			}
			_, err = lc.MergeProfiles(mergeRq.TargetID, mergeRq.SourceID, ProfileMergeContext{})
			if err != nil {
				t.Errorf("[%s] error merging profiles from request %+v: %v", t.Name(), mergeRq, err)
			}
		}
		log.Printf("Waiting 5 seconds to poll merge queue. Empty messages in a row: %d", emptyMessageCounter)
		time.Sleep(5 * time.Second)
	}
}

// Utility function to generate basic mappings for tests.
//
// Can be extended with additional mappings for a specific use case
// (i.e. need to validate rule based match on "ip_address")
func generateTestMappings(additionalMappings ...FieldMapping) []FieldMapping {
	fieldMappings := []FieldMapping{
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
			Target:  "_profile.Attributes.profile_id",
			Indexes: []string{"PROFILE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.loyalty_id",
			Target:  "_profile.Attributes.loyalty_id",
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.first_name",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.last_name",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.booking_id",
			Target:  "booking_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.booker_id",
			Target:  "booker_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.amount",
			Target:  "amount",
			KeyOnly: true,
		},
		{
			Type:    MappingTypeString,
			Source:  "_source." + LAST_UPDATED_FIELD,
			Target:  LAST_UPDATED_FIELD,
			KeyOnly: true,
		},
	}

	fieldMappings = append(fieldMappings, additionalMappings...)
	return fieldMappings
}

// Utility function to get exact count of each connect_id. This is NOT efficient for
// counting items in a large table and should only be used for testing purposes.
func testOnlyGetCount(lc *CustomerProfileConfigLowCost, tableName string) (map[string]int, error) {
	sql := fmt.Sprintf(`
		SELECT connect_id, COUNT(*) AS count
		FROM %s
		GROUP BY connect_id;
	`, tableName)
	res, err := lc.Data.AurSvc.Query(sql)
	if err != nil {
		return nil, err
	}

	countById := make(map[string]int)
	for _, row := range res {
		connectId, ok := row["connect_id"].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected response for connect_id")
		}
		count, ok := row["count"].(int64)
		if !ok {
			return nil, fmt.Errorf("unexpected response for count")
		}
		countById[connectId] = int(count)
	}

	return countById, nil
}
