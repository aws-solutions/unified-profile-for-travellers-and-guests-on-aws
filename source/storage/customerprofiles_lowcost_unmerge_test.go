// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"tah/upt/source/tah-core/aurora"
	core "tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

/*
READ BEFORE RUNNING TESTS

Typically, all tests in this repository can be run directly without any special
configuration. However, the low cost storage solution requires connection to an
Amazon Aurora Serverless cluster. The cluster can take several minutes to create
and requires unique setup for secure local connections.

For these reasons, we have a guide for one-time setup, and then assume the
cluster is available when running tests.
*/

type SearchFromProfileTest struct {
	Profile               profilemodel.Profile
	ExpectedSql           string
	ExpectedArgs          []interface{}
	LastUpdatedObjectType map[string]string
	LastUpdatedTimestamp  map[string]time.Time
}

func CreateSearchFromProfileTests(domain string) []SearchFromProfileTest {
	var serachFromProfileNow = time.Now()
	return []SearchFromProfileTest{
		{
			Profile: profilemodel.Profile{
				ProfileId:            "abcd-efgh-ijklmnop-uvwxyz",
				FirstName:            "John",
				PersonalEmailAddress: "test@example.com",
				Attributes: map[string]string{
					"traveler_id": "test_custom_attr",
				},
				LastUpdated: serachFromProfileNow,
			},
			LastUpdatedObjectType: map[string]string{
				"PersonalEmailAddress":   "air_booking",
				"FirstName":              "hotel_booking",
				"Attributes.traveler_id": "hotel_booking",
			},
			LastUpdatedTimestamp: map[string]time.Time{
				"PersonalEmailAddress":   serachFromProfileNow.AddDate(0, 0, 1),
				"FirstName":              serachFromProfileNow.AddDate(0, 0, 2),
				"Attributes.traveler_id": serachFromProfileNow.AddDate(0, 0, 3),
			},
			ExpectedSql: `INSERT INTO ` + domain + `_search ("connect_id","timestamp","FirstName","FirstName_ts","FirstName_obj_type","PersonalEmailAddress","PersonalEmailAddress_ts","PersonalEmailAddress_obj_type","Attributes.traveler_id","Attributes.traveler_id_ts","Attributes.traveler_id_obj_type") VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
			ExpectedArgs: []interface{}{"abcd-efgh-ijklmnop-uvwxyz",
				serachFromProfileNow,
				"John",
				serachFromProfileNow.AddDate(0, 0, 2),
				"hotel_booking",
				"test@example.com",
				serachFromProfileNow.AddDate(0, 0, 1),
				"air_booking",
				"test_custom_attr",
				serachFromProfileNow.AddDate(0, 0, 3),
				"hotel_booking"},
		},
	}

}

func TestUnmergeRebuildSearchRecordFromProfile(t *testing.T) {
	t.Parallel()
	dynamoTableName := "lcs" + core.GenerateUniqueId()
	domainName := randomDomain("searchFromProf")
	lc := SetupUnmergeTest(t, domainName, dynamoTableName)

	mappings, err := lc.GetMappings()
	if err != nil {
		t.Fatalf("Error getting mappings: %s", err)
	}
	for _, test := range CreateSearchFromProfileTests(domainName) {
		sql, args := lc.Data.buildSearchProfileInsertRowFromProfileSql(
			domainName,
			test.Profile,
			test.LastUpdatedObjectType,
			test.LastUpdatedTimestamp,
			mappings,
		)
		log.Printf("Test %d", 1)
		log.Printf("sql: %s", sql)
		log.Printf("args: %+v", args)
		if sql != test.ExpectedSql {
			t.Fatalf("Invalid SQL Expected sql: '%s'\n got\n '%s'\n", test.ExpectedSql, sql)
		}
		if len(args) != len(test.ExpectedArgs) {
			t.Fatalf("Expected %d args, got: %d", len(test.ExpectedArgs), len(args))
		}
		for i, exp := range test.ExpectedArgs {
			if exp != args[i] {
				t.Fatalf("Expected arg %d: %s, got: %s", i, exp, args[i])
			}
		}
		_, err = lc.Data.AurSvc.Query(sql, args...)
		if err != nil {
			t.Fatalf("Error executing sql: %s", err)
		}
	}
}

var (
	ruleMergeObjOne = ProfileMergeContext{
		MergeType:              MergeTypeRule,
		ConfidenceUpdateFactor: 0.9,
		RuleID:                 "rule_id1",
		RuleSetVersion:         "v1",
	}
	ruleMergeObjTwo = ProfileMergeContext{
		MergeType:              MergeTypeRule,
		ConfidenceUpdateFactor: 0.8,
		RuleID:                 "rule_id2",
		RuleSetVersion:         "v1",
	}
	ruleMergeObjThree = ProfileMergeContext{
		MergeType:              MergeTypeRule,
		ConfidenceUpdateFactor: 0.7,
		RuleID:                 "rule_id3",
		RuleSetVersion:         "v1",
	}
	aiMergeObj = ProfileMergeContext{
		MergeType:              MergeTypeAI,
		ConfidenceUpdateFactor: 0.6,
	}
)

var interactionPrefix = "interaction0_for_profile_"

// Sets up LCS, domain, creates dynamo table
func SetupUnmergeTest(
	t *testing.T,
	testDomainNameValid string,
	dynamoTableName string,
) *CustomerProfileConfigLowCost {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	var lc *CustomerProfileConfigLowCost
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[SetupUnmergeTest] Error setting up aurora: %+v ", err)
	}

	cfg, err := db.InitWithNewTable(dynamoTableName, "pk", "sk", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[SetupUnmergeTest] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = cfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] error deleting dynamo table: %v", t.Name(), err)
		}
	})
	err = cfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[SetupUnmergeTest] error waiting for table creation: %v", err)
	}
	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)

	mergeQueue := buildTestMergeQueue(t, envCfg.Region)
	cpWriterQueue := buildTestCPWriterQueue(t, envCfg.Region)
	initOptions := CustomerProfileInitOptions{MergeQueueClient: mergeQueue}
	lc = InitLowCost(
		envCfg.Region,
		auroraCfg,
		&cfg,
		kinesisMockCfg,
		cpWriterQueue,
		"tah/upt/source/tah-core/customerprofile",
		core.TEST_SOLUTION_ID,
		core.TEST_SOLUTION_VERSION,
		core.LogLevelDebug,
		&initOptions,
		InitLcsCache(),
	)

	envNameTest := "testenv"
	err = lc.CreateDomainWithQueue(testDomainNameValid, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: envNameTest}, "", "", DomainOptions{})
	if err != nil {
		t.Fatalf("[SetupUnmergeTest] error creating domain: %v", err)
	}
	t.Cleanup(func() { lc.DeleteDomainByName(testDomainNameValid) })

	lc.CreateMapping("air_booking", "air booking mapping", []FieldMapping{
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
		{
			Type:   MappingTypeString,
			Source: "_source." + LAST_UPDATED_FIELD,
			Target: LAST_UPDATED_FIELD,
		},
	})

	return lc
}

// Sends a request to create the profile object in LCS and ensure profile has been created
func CreateProfiles(obj []byte, tid string, lc *CustomerProfileConfigLowCost, t *testing.T) (string, error) {
	var uptId string
	err := lc.PutProfileObject(string(obj), "air_booking")
	if err != nil {
		t.Errorf("[CreateProfiles] error creating profile object: %v", err)
		return "", err
	}

	res, err := lc.SearchProfiles("profile_id", []string{tid})
	if err != nil {
		t.Errorf("[CreateProfiles] error searching profiles: %v", err)
		return "", err
	}
	if len(res) != 1 {
		t.Errorf("[CreateProfiles] search should have 1 record and not %v", len(res))
		return "", err
	} else {
		uptId = res[0].ProfileId
		if res[0].Attributes["traveler_id"] != tid {
			t.Errorf("[CreateProfiles] profile id should be profile_id1 and not %v", res[0].ProfileId)
			return "", err
		}
	}
	return uptId, nil
}

// Runs a loop to create N profiles, invokes `CreateProfiles` which handles profile creation
func CreateNProfiles(lc *CustomerProfileConfigLowCost, t *testing.T, N int) ([]string, error) {
	var uptIdList [PROFILE_HISTORY_DEPTH + 1]string
	var err error

	for i := 0; i < N; i++ {
		obj, _ := json.Marshal(map[string]string{
			"traveler_id":      "profile_id" + fmt.Sprint(i),
			"accp_object_id":   interactionPrefix + fmt.Sprint(i),
			"segment_id":       "test_segment_id",
			"from":             "test_from",
			"arrival_date":     "test_arrival_date",
			"email":            "test_email" + fmt.Sprint(i) + "@example.com",
			LAST_UPDATED_FIELD: time.Now().UTC().Format(INGEST_TIMESTAMP_FORMAT),
		})
		uptIdList[i], err = CreateProfiles(obj, "profile_id"+fmt.Sprint(i), lc, t)
		if err != nil {
			t.Errorf("[CreateNProfiles] error creating profile object: %v", err)
			return uptIdList[:], err
		}
	}

	return uptIdList[:], nil
}

// Ingests N interactions to a profile
func sendNInteractionsToProfile(t *testing.T, lc *CustomerProfileConfigLowCost, N int, profileNum int) {
	for i := 0; i < N; i++ {
		obj1, _ := json.Marshal(map[string]string{
			"traveler_id":      "profile_id" + fmt.Sprint(profileNum),
			"accp_object_id":   "new_interaction" + fmt.Sprint(i) + "_for_profile_" + fmt.Sprint(profileNum),
			"segment_id":       "test_segment_id",
			"from":             "test_from" + fmt.Sprint(i),
			"arrival_date":     "test_arrival_date",
			"email":            "airBookingEmail" + fmt.Sprint(profileNum) + "@example.com",
			LAST_UPDATED_FIELD: time.Now().UTC().Format(INGEST_TIMESTAMP_FORMAT),
		})
		err := lc.PutProfileObject(string(obj1), "air_booking")
		if err != nil {
			t.Errorf("[sendNInteractionsToProfile] error creating profile object 1: %v", err)
		}
	}
}

// Unmerges a profile after a series of merges
func TestUnmergeSimple(t *testing.T) {
	t.Parallel()
	testDomainNameValid := randomDomain("unmergeSimple")
	dynamoTableName := "customer-profiles-low-cost" + core.GenerateUniqueId()
	lc := SetupUnmergeTest(t, testDomainNameValid, dynamoTableName)

	nProfiles := 5
	uptIdList, err := CreateNProfiles(lc, t, nProfiles)
	if err != nil {
		t.Fatalf("[%s] error creating profile obect: %v", t.Name(), err)
	}

	// Merge 0 -> 1
	// {1, {0}} {2} {3} {4}
	_, err = lc.MergeProfiles(uptIdList[1], uptIdList[0], ruleMergeObjOne)
	//  {1, {0}} {2} {3} {4}
	if err != nil {
		t.Fatalf("[%s] error merging profiles: %v", t.Name(), err)
	}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)

	// Merge 1 -> 3
	_, err = lc.MergeProfiles(uptIdList[3], uptIdList[1], aiMergeObj)
	// {2} {3, {1, {0}}} {4}
	if err != nil {
		t.Fatalf("[%s] error merging profiles: %v", t.Name(), err)
	}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor*aiMergeObj.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"1",
		aiMergeObj.ConfidenceUpdateFactor,
	)

	// Merge 3 -> 4
	_, err = lc.MergeProfiles(uptIdList[4], uptIdList[3], ruleMergeObjTwo)
	// {2} {4, {3, {1, {0}}}}
	if err != nil {
		t.Fatalf("[%s] error merging profiles: %v", t.Name(), err)
	}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor*aiMergeObj.ConfidenceUpdateFactor*ruleMergeObjTwo.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"1",
		aiMergeObj.ConfidenceUpdateFactor*ruleMergeObjTwo.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"3",
		ruleMergeObjTwo.ConfidenceUpdateFactor,
	)

	_, err = lc.UnmergeProfiles(uptIdList[1], uptIdList[4], "", "air_booking")
	// {2} {4, {3}} {1, {0}}
	if err != nil {
		t.Fatalf("[%s] error unmerging profiles: %v", t.Name(), err)
	}
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(t, lc.Data.AurSvc, testDomainNameValid, "air_booking", interactionPrefix+"1", 1)

	validateDynamoProfileCreationAfterUnmerge(t, lc, uptIdList[1])
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[1], "profile_id1 profile_id0")
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[4], "profile_id4 profile_id3")
}

// Validates the merge history for an interaction after a series of merges
func TestUnmergeContextHistory(t *testing.T) {
	t.Parallel()
	testDomainNameValid := randomDomain("context_history")
	dynamoTableName := "customer-profiles-low-cost" + core.GenerateUniqueId()
	lc := SetupUnmergeTest(t, testDomainNameValid, dynamoTableName)

	nProfiles := 5
	uptIdList, err := CreateNProfiles(lc, t, nProfiles)
	if err != nil {
		t.Fatalf("[%s] error creating profile object: %v", t.Name(), err)
	}

	// Merge 0 -> 1
	// {1, {0}} {2} {3} {4}
	lc.MergeProfiles(uptIdList[1], uptIdList[0], ruleMergeObjOne)
	//  {1, {0}} {2} {3} {4}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)

	// Merge 2 -> 3
	lc.MergeProfiles(uptIdList[3], uptIdList[2], aiMergeObj)
	// {2} {3, {1, {0}}} {4}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"2",
		aiMergeObj.ConfidenceUpdateFactor,
	)

	// Merge 1 -> 3
	lc.MergeProfiles(uptIdList[3], uptIdList[1], ruleMergeObjTwo)
	// {2} {3, {1, {0}}} {4}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor*ruleMergeObjTwo.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"1",
		ruleMergeObjTwo.ConfidenceUpdateFactor,
	)

	// Merge 3 -> 4
	lc.MergeProfiles(uptIdList[4], uptIdList[3], ruleMergeObjThree)
	// {2} {4, {3, {1, {0}}}}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor*ruleMergeObjTwo.ConfidenceUpdateFactor*ruleMergeObjThree.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"1",
		ruleMergeObjTwo.ConfidenceUpdateFactor*ruleMergeObjThree.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"2",
		aiMergeObj.ConfidenceUpdateFactor*ruleMergeObjThree.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"3",
		ruleMergeObjThree.ConfidenceUpdateFactor,
	)

	i0History, err := lc.BuildInteractionHistory(testDomainNameValid, interactionPrefix+"0", "air_booking", "")
	if err != nil {
		t.Fatalf("[%s] error fetching interaction history: %v", t.Name(), err)
	}
	i2History, err := lc.BuildInteractionHistory(testDomainNameValid, interactionPrefix+"2", "air_booking", "")
	if err != nil {
		t.Fatalf("[%s] error fetching interaction history: %v", t.Name(), err)
	}

	if len(i0History) != 4 {
		t.Fatalf("[%s] Incorrect number of interactions", t.Name())
	}

	if len(i2History) != 3 {
		t.Fatalf("[%s] Incorrect number of interactions", t.Name())
	}
}

// Performs 2 unmerges after a series of merge events
// Purpose is to ensure unmerge logic takes into account previous unmerge events
func TestUnmergeMultipleUnmerges(t *testing.T) {
	t.Parallel()
	testDomainNameValid := randomDomain("multipleUnmerge")
	dynamoTableName := "customer-profiles-low-cost" + core.GenerateUniqueId()
	lc := SetupUnmergeTest(t, testDomainNameValid, dynamoTableName)

	nProfiles := 5
	uptIdList, err := CreateNProfiles(lc, t, nProfiles)
	if err != nil {
		t.Fatalf("[%s] error creating profile object: %v", t.Name(), err)
	}

	// Merge 0 -> 1
	// {1}, {0}, {2} {3} {4}
	_, err = lc.MergeProfiles(uptIdList[1], uptIdList[0], ruleMergeObjOne)
	//  {1, {0}} {2} {3} {4}
	if err != nil {
		t.Fatalf("[%s] error merging profiles 0 into 1: %v", t.Name(), err)
	}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)

	// Merge 1 -> 3
	_, err = lc.MergeProfiles(uptIdList[3], uptIdList[1], ruleMergeObjTwo)
	// {2} {3, {1, {0}}} {4}
	if err != nil {
		t.Fatalf("[%s] error merging profiles 1 into 3: %v", t.Name(), err)
	}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor*ruleMergeObjTwo.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"1",
		ruleMergeObjTwo.ConfidenceUpdateFactor,
	)

	// Merge 3 -> 4
	_, err = lc.MergeProfiles(uptIdList[4], uptIdList[3], ruleMergeObjThree)
	// {2} {4, {3, {1, {0}}}}
	if err != nil {
		t.Fatalf("[%s] error merging profiles 3 into 4: %v", t.Name(), err)
	}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor*ruleMergeObjTwo.ConfidenceUpdateFactor*ruleMergeObjThree.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"1",
		ruleMergeObjTwo.ConfidenceUpdateFactor*ruleMergeObjThree.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"3",
		ruleMergeObjThree.ConfidenceUpdateFactor,
	)

	// Unmerge 1 from 4
	_, err = lc.UnmergeProfiles(uptIdList[1], uptIdList[4], "", "air_booking")
	// {2} {4, {3}} {1, {0}}
	if err != nil {
		t.Fatalf("[%s] error merging profiles 1 from 4: %v", t.Name(), err)
	}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(t, lc.Data.AurSvc, testDomainNameValid, "air_booking", interactionPrefix+"1", 1)

	validateDynamoProfileCreationAfterUnmerge(t, lc, uptIdList[1])
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[1], "profile_id1 profile_id0")
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[4], "profile_id4 profile_id3")

	// Unmerge 3 from 4
	_, err = lc.UnmergeProfiles(uptIdList[3], uptIdList[4], "", "air_booking")
	// {2} {4} {3} {1, {0}}
	if err != nil {
		t.Fatalf("[%s] error unmerging profiles 3 from 4: %v", t.Name(), err)
	}
	validateConfidenceScoreCorrectness(t, lc.Data.AurSvc, testDomainNameValid, "air_booking", interactionPrefix+"3", 1)

	validateDynamoProfileCreationAfterUnmerge(t, lc, uptIdList[3])
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[3], "profile_id3")
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[4], "profile_id4")
}

// Merges and unmerges a particular profile in different "parents"
// Purpose is to ensure unmerge logic takes into account previous merge/unmerge events
func TestUnmergeSequence(t *testing.T) {
	t.Parallel()
	testDomainNameValid := randomDomain("unmerge_seq")
	dynamoTableName := "customer-profiles-low-cost" + core.GenerateUniqueId()
	lc := SetupUnmergeTest(t, testDomainNameValid, dynamoTableName)

	nProfiles := 4
	uptIdList, err := CreateNProfiles(lc, t, nProfiles)
	if err != nil {
		t.Fatalf("[%s] error creating profile object: %v", t.Name(), err)
	}

	// Merge 0 -> 1
	// {1, {0}} {2} {3}
	lc.MergeProfiles(uptIdList[1], uptIdList[0], ruleMergeObjOne)
	//  {1, {0}} {2} {3} {4}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)

	// Merge 1 -> 2
	lc.MergeProfiles(uptIdList[2], uptIdList[1], ruleMergeObjTwo)
	// {3} {2, {1, {0}}}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor*ruleMergeObjTwo.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"1",
		ruleMergeObjTwo.ConfidenceUpdateFactor,
	)

	// Merge 2 -> 3
	lc.MergeProfiles(uptIdList[3], uptIdList[2], ruleMergeObjThree)
	// {3, {2, {1, {0}}}}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor*ruleMergeObjTwo.ConfidenceUpdateFactor*ruleMergeObjThree.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"1",
		ruleMergeObjTwo.ConfidenceUpdateFactor*ruleMergeObjThree.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"2",
		ruleMergeObjThree.ConfidenceUpdateFactor,
	)

	// Unmerge 0 from 1
	lc.UnmergeProfiles(uptIdList[0], uptIdList[1], "", "air_booking")
	// {3, {2, {1}}} {0}

	validateConfidenceScoreCorrectness(t, lc.Data.AurSvc, testDomainNameValid, "air_booking", interactionPrefix+"0", 1)

	validateDynamoProfileCreationAfterUnmerge(t, lc, uptIdList[0])
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[3], "profile_id3 profile_id2 profile_id1")

	// Merge 0 -> 3
	lc.MergeProfiles(uptIdList[3], uptIdList[0], ruleMergeObjOne)
	// {3, {2, {1}}, {0}}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)

	// Unmerge 1 from 3
	lc.UnmergeProfiles(uptIdList[1], uptIdList[3], "", "air_booking")
	// {3, {2}, {0}} {1}

	validateConfidenceScoreCorrectness(t, lc.Data.AurSvc, testDomainNameValid, "air_booking", interactionPrefix+"1", 1)

	validateDynamoProfileCreationAfterUnmerge(t, lc, uptIdList[1])
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[3], "profile_id3 profile_id2 profile_id0")
}

// Merges and unmerges the same profiles twice
// Purpose is to ensure that after an unmerge, the same profiles can be again merged and unmerged
// i.e. a duplicate merge event
/*
var (
	ruleMergeObjOne = ProfileMergeContext{
		MergeType:              MergeTypeRule,
		ConfidenceUpdateFactor: 0.9,
		RuleID:                 "rule_id1",
		RuleSetVersion:         "v1",
	}
	ruleMergeObjTwo = ProfileMergeContext{
		MergeType:              MergeTypeRule,
		ConfidenceUpdateFactor: 0.8,
		RuleID:                 "rule_id2",
		RuleSetVersion:         "v1",
	}
	ruleMergeObjThree = ProfileMergeContext{
		MergeType:              MergeTypeRule,
		ConfidenceUpdateFactor: 0.7,
		RuleID:                 "rule_id3",
		RuleSetVersion:         "v1",
	}
	aiMergeObj = ProfileMergeContext{
		MergeType:              MergeTypeAI,
		ConfidenceUpdateFactor: 0.6,
	}
)*/
func TestUnmergeDuplicateMerges1(t *testing.T) {
	t.Parallel()
	testDomainNameValid := randomDomain("dup_merge")
	dynamoTableName := "customer-profiles-low-cost" + core.GenerateUniqueId()
	lc := SetupUnmergeTest(t, testDomainNameValid, dynamoTableName)

	nProfiles := 4
	uptIdList, err := CreateNProfiles(lc, t, nProfiles)
	if err != nil {
		t.Fatalf("[%s] error creating profile object: %v", t.Name(), err)
	}

	// Step 1: Merge 0 -> 1
	// {1, {0}} {2} {3}
	_, err = lc.MergeProfiles(uptIdList[1], uptIdList[0], ruleMergeObjOne)
	if err != nil {
		t.Fatalf("[%s] error merging profiles 1 into 0: %v", t.Name(), err)
	}
	//  {1, {0}} {2} {3} {4}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)

	// Step 2: Merge 1 -> 2
	_, err = lc.MergeProfiles(uptIdList[2], uptIdList[1], ruleMergeObjTwo)
	if err != nil {
		t.Fatalf("[%s] error merging profiles 2 into 1: %v", t.Name(), err)
	}
	// {3} {2, {1, {0}}}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor*ruleMergeObjTwo.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"1",
		ruleMergeObjTwo.ConfidenceUpdateFactor,
	)

	// Step 3: Merge 2 -> 3
	_, err = lc.MergeProfiles(uptIdList[3], uptIdList[2], ruleMergeObjThree)
	if err != nil {
		t.Fatalf("[%s] error merging profiles 3 into 2: %v", t.Name(), err)
	}
	// {3, {2, {1, {0}}}}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor*ruleMergeObjTwo.ConfidenceUpdateFactor*ruleMergeObjThree.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"1",
		ruleMergeObjTwo.ConfidenceUpdateFactor*ruleMergeObjThree.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"2",
		ruleMergeObjThree.ConfidenceUpdateFactor,
	)

	// Step 4: Unmerge 0 from 1
	_, err = lc.UnmergeProfiles(uptIdList[0], uptIdList[1], "", "air_booking")
	if err != nil {
		t.Fatalf("[%s] error unmerging profiles 1 from 0: %v", t.Name(), err)
	}
	// {3, {2}} {1, {0}}

	validateConfidenceScoreCorrectness(t, lc.Data.AurSvc, testDomainNameValid, "air_booking", interactionPrefix+"0", 1)
	validateDynamoProfileCreationAfterUnmerge(t, lc, uptIdList[0])

	// Step 5: Merge 0 -> 1
	// (Duplicate of step 1)
	_, err = lc.MergeProfiles(uptIdList[1], uptIdList[0], ruleMergeObjOne)
	if err != nil {
		t.Fatalf("[%s] error merging profiles 1 into 0: %v", t.Name(), err)
	}
	// {3, {2, {1, {0}}}}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)

	// Step 6: Unmerge 1 from 3
	_, err = lc.UnmergeProfiles(uptIdList[1], uptIdList[3], "", "air_booking")
	if err != nil {
		t.Fatalf("[%s] error unmerging profiles 3 from 1: %v", t.Name(), err)
	}
	// {3, {2}} {1, {0}}

	validateConfidenceScoreCorrectness(t, lc.Data.AurSvc, testDomainNameValid, "air_booking", interactionPrefix+"1", 1)
	validateDynamoProfileCreationAfterUnmerge(t, lc, uptIdList[1])
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[1], "profile_id0 profile_id1")
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[3], "profile_id3 profile_id2")
}

func TestUnmergeDuplicateMerges2(t *testing.T) {
	t.Parallel()
	testDomainNameValid := randomDomain("dup_merge_2")
	dynamoTableName := "customer-profiles-low-cost" + core.GenerateUniqueId()
	lc := SetupUnmergeTest(t, testDomainNameValid, dynamoTableName)

	nProfiles := 4
	uptIdList, err := CreateNProfiles(lc, t, nProfiles)
	if err != nil {
		t.Fatalf("[%s] error creating profile object: %v", t.Name(), err)
	}

	// {0} {1} {2} {3}
	// Step 1: Merge 0 -> 1
	_, err = lc.MergeProfiles(uptIdList[1], uptIdList[0], ruleMergeObjOne)
	if err != nil {
		t.Fatalf("[%s] error unmerging profiles 3 from 1: %v", t.Name(), err)
	}
	// {1, {0}} {2} {3}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)

	// Step 2: Unmerge 0 from 1
	_, err = lc.UnmergeProfiles(uptIdList[0], uptIdList[1], "", "air_booking")
	if err != nil {
		t.Fatalf("[%s] error unmerging profiles 1 from 0: %v", t.Name(), err)
	}
	// {0} {1} {2} {3}

	validateConfidenceScoreCorrectness(t, lc.Data.AurSvc, testDomainNameValid, "air_booking", interactionPrefix+"0", 1)

	// Step 3: Merge 0 -> 2
	_, err = lc.MergeProfiles(uptIdList[2], uptIdList[0], ruleMergeObjOne)
	if err != nil {
		t.Fatalf("[%s] error unmerging profiles 3 from 1: %v", t.Name(), err)
	}
	// {2, {0}} {1} {3}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)

	// Step 4: Merge 1 -> 3
	_, err = lc.MergeProfiles(uptIdList[3], uptIdList[1], ruleMergeObjOne)
	if err != nil {
		t.Fatalf("[%s] error unmerging profiles 3 from 1: %v", t.Name(), err)
	}
	// {2, {0}} {3, {1}}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"1",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)

	// Step 5: Unmerge 1 from 3
	_, err = lc.UnmergeProfiles(uptIdList[1], uptIdList[3], "", "air_booking")
	if err != nil {
		t.Fatalf("[%s] error unmerging profiles 3 from 1: %v", t.Name(), err)
	}
	// {2, {0}} {1} {3}

	validateConfidenceScoreCorrectness(
		t,
		lc.Data.AurSvc,
		testDomainNameValid,
		"air_booking",
		interactionPrefix+"0",
		ruleMergeObjOne.ConfidenceUpdateFactor,
	)
	validateConfidenceScoreCorrectness(t, lc.Data.AurSvc, testDomainNameValid, "air_booking", interactionPrefix+"1", 1)

	validateDynamoProfileCreationAfterUnmerge(t, lc, uptIdList[1])
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[1], "profile_id1")

}

// Purpose is to ensure that an unmerge should also "pull out" the interactions belonging to the unmerged profile
func TestUnmergeHasCorrectInteractions(t *testing.T) {
	t.Parallel()
	testDomainNameValid := randomDomain("unmergeInteract")
	dynamoTableName := "customer-profiles-low-cost" + core.GenerateUniqueId()
	lc := SetupUnmergeTest(t, testDomainNameValid, dynamoTableName)

	nProfiles := 5
	uptIdList, err := CreateNProfiles(lc, t, nProfiles)
	if err != nil {
		t.Fatalf("[%s] error creating profile object: %v", t.Name(), err)
	}

	// {0} {1} {2} {3} {4}
	// Merge 0 -> 1
	lc.MergeProfiles(uptIdList[1], uptIdList[0], ruleMergeObjOne)
	//  {1, {0}} {2} {3} {4}

	// Merge 1 -> 3
	lc.MergeProfiles(uptIdList[3], uptIdList[1], ruleMergeObjTwo)
	// {2} {3, {1, {0}}} {4}

	// Merge 3 -> 4
	lc.MergeProfiles(uptIdList[4], uptIdList[3], ruleMergeObjThree)
	// {2} {4, {3, {1, {0}}}}

	// Initially each profile has one interaction
	// send a few more interactions to profiles 1 and 0
	sendNInteractionsToProfile(t, lc, 3, 1)
	sendNInteractionsToProfile(t, lc, 2, 0)

	lc.UnmergeProfiles(uptIdList[1], uptIdList[4], "", "air_booking")
	// {2} {4, {3}} {1, {0}}

	conn, _ := lc.Data.AurSvc.AcquireConnection(lc.Tx)
	defer conn.Release()
	objectTypeNameToId := make(map[string]string, 1)
	objectTypeNameToId["air_booking"] = ""
	interactionsForOne, _ := lc.Data.FindInteractionRecordsByConnectID(conn, lc.DomainName, uptIdList[1], []ObjectMapping{}, objectTypeNameToId, []PaginationOptions{})
	if len(interactionsForOne) != 7 {
		t.Fatalf("[%s] Incorrect interactions for profile one", t.Name())
	}

	interactionsForFour, _ := lc.Data.FindInteractionRecordsByConnectID(conn, lc.DomainName, uptIdList[4], []ObjectMapping{}, objectTypeNameToId, []PaginationOptions{})
	if len(interactionsForFour) != 2 {
		t.Fatalf("[%s] Incorrect interactions for profile four", t.Name())
	}

	interactionsForTwo, _ := lc.Data.FindInteractionRecordsByConnectID(conn, lc.DomainName, uptIdList[2], []ObjectMapping{}, objectTypeNameToId, []PaginationOptions{})
	if len(interactionsForTwo) != 1 {
		t.Fatalf("[%s] Incorrect interactions for profile two", t.Name())
	}
}

// This test validates unmerge api's ability to unmerge by interactionId if the connectID for profile to unmerge is not provided
func TestUnmergeByInteractionId(t *testing.T) {
	t.Parallel()
	testDomainNameValid := randomDomain("unmerge_id")
	dynamoTableName := "customer-profiles-low-cost" + core.GenerateUniqueId()
	lc := SetupUnmergeTest(t, testDomainNameValid, dynamoTableName)

	nProfiles := 5
	uptIdList, err := CreateNProfiles(lc, t, nProfiles)
	if err != nil {
		t.Fatalf("[%s] error creating profile object: %v", t.Name(), err)
	}

	// Merge 0 -> 1
	// {1, {0}} {2} {3} {4}
	lc.MergeProfiles(uptIdList[1], uptIdList[0], ruleMergeObjOne)
	//  {1, {0}} {2} {3} {4}

	// Merge 1 -> 3
	lc.MergeProfiles(uptIdList[3], uptIdList[1], aiMergeObj)
	// {2} {3, {1, {0}}} {4}

	// Merge 3 -> 4
	lc.MergeProfiles(uptIdList[4], uptIdList[3], ruleMergeObjTwo)
	// {2} {4, {3, {1, {0}}}}

	lc.UnmergeProfiles("", uptIdList[4], interactionPrefix+"1", "air_booking")
	// {2} {4, {3}} {1, {0}}

	validateDynamoProfileCreationAfterUnmerge(t, lc, uptIdList[1])
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[1], "profile_id1 profile_id0")
	validateProfileHasCorrectChildrenProfilesAfterUnmerge(t, lc, testDomainNameValid, uptIdList[4], "profile_id4 profile_id3")
}

// Unmerge request should be denied if both connectId and interactionId are not given
func TestUnmergeInvalidUnmergeRequest(t *testing.T) {
	t.Parallel()
	testDomainNameValid := randomDomain("unmerge_invalid")
	dynamoTableName := "customer-profiles-low-cost" + core.GenerateUniqueId()
	lc := SetupUnmergeTest(t, testDomainNameValid, dynamoTableName)

	nProfiles := 2
	uptIdList, err := CreateNProfiles(lc, t, nProfiles)
	if err != nil {
		t.Fatalf("[%s] error creating profile object: %v", t.Name(), err)
	}

	// Merge 0 -> 1
	// {1} {0}
	lc.MergeProfiles(uptIdList[1], uptIdList[0], ruleMergeObjOne)
	//  {1, {0}}

	_, err = lc.UnmergeProfiles("", uptIdList[0], "", "air_booking")
	if err == nil || err.Error() != "[UnmergeProfiles] unable to perform unmerge with the information given" {
		t.Fatalf("[%s] Invalid request was not denied", t.Name())
	}

}

func TestUnmergeBuildDescendants(t *testing.T) {
	t.Parallel()
	mergeHistoryMap := make(map[string]map[string]int)

	// {1} {2} {3} {4} {5}

	// Step 1: Merge 1 -> 2
	// Step 2: Merge 2 -> 3
	// Step 3: Merge 3 -> 4
	// Step 4: Merge 4 -> 5

	// Result {5, {4, {3, {2, {1}}}}}
	mergeHistoryMap["upt_id_5"] = make(map[string]int)
	mergeHistoryMap["upt_id_5"]["upt_id_4"] = 1
	mergeHistoryMap["upt_id_4"] = make(map[string]int)
	mergeHistoryMap["upt_id_4"]["upt_id_3"] = 1
	mergeHistoryMap["upt_id_3"] = make(map[string]int)
	mergeHistoryMap["upt_id_3"]["upt_id_2"] = 1
	mergeHistoryMap["upt_id_2"] = make(map[string]int)
	mergeHistoryMap["upt_id_2"]["upt_id_1"] = 1

	for i := 5; i >= 1; i-- {
		lineageListForProfile := buildDescendantsFromMap("upt_id_"+strconv.Itoa(i), mergeHistoryMap)
		if len(
			lineageListForProfile,
		) != i { // Lineage for a profile includes self, Profile 5 should have 5 profiles in its lineage (5,4,3,2,1)
			t.Fatalf("[%s] Incorrect number of descendants in lineage list for profile %v", t.Name(), i)
		}
	}
}

// Helper function to validate the outcome of an unmerge - i.e. profiles being unmerged from the parent
func validateProfileHasCorrectChildrenProfilesAfterUnmerge(
	t *testing.T,
	lc *CustomerProfileConfigLowCost,
	testDomainNameValid string,
	uptId string,
	expectedValue string,
) {
	conn, err := lc.Data.AurSvc.AcquireConnection(core.NewTransaction(t.Name(), "", core.LogLevelDebug))
	if err != nil {
		t.Errorf("Error aquiring connection: %v", err)
	}
	defer conn.Release()
	children, _ := lc.Data.GetProfileIdRecords(conn, testDomainNameValid, uptId)
	var lineageList []string
	for _, child := range children {
		for _, value := range child {
			lineageList = append(lineageList, value.(string))
		}
	}

	if strings.Join(lineageList, " ") != expectedValue {
		t.Fatalf(
			"[validateProfileHasCorrectChildrenProfilesAfterUnmerge] Expected %s, got %v",
			expectedValue,
			strings.Join(lineageList, " "),
		)
	}
}

func validateDynamoProfileCreationAfterUnmerge(t *testing.T, lc *CustomerProfileConfigLowCost, uptId string) {
	profile, err := lc.GetProfile(uptId, []string{}, []PaginationOptions{})
	if err != nil {
		t.Fatalf("[validateDynamoProfileCreationAfterUnmerge] Profile %s was not created. Error: %v", uptId, err)
	}
	if profile.ProfileId != uptId {
		t.Fatalf("[validateDynamoProfileCreationAfterUnmerge] Profile %s was not created correctly: %+v", uptId, profile)
	}
}

// Fetches the interaction and compares the confidence score computed after merge/unmerge with expected value
func validateConfidenceScoreCorrectness(
	t *testing.T,
	auroraCfg *aurora.PostgresDBConfig,
	domain, objectTypeName, interactionId string,
	expectedConfidence float64,
) {
	dh := DataHandler{
		AurSvc: auroraCfg,
	}

	tableName := createInteractionTableName(domain, objectTypeName)
	interaction, err := dh.AurSvc.Query(
		fmt.Sprintf(`SELECT overall_confidence_score FROM %s WHERE accp_object_id = '%s'`, tableName, interactionId),
	)
	if err != nil {
		t.Fatalf("[validateConfidenceScoreCorrectness] Profile was not created in dynamo")
	}
	if len(interaction) != 1 {
		t.Fatalf("[validateConfidenceScoreCorrectness] incorrect number of interactions found. should be 1. have %d", len(interaction))
	}

	actualConfidence, ok := interaction[0]["overall_confidence_score"].(float64)
	if !ok {
		t.Fatalf("[validateConfidenceScoreCorrectness] error casting confidence score into string")
	}
	expectedRounded := fmt.Sprintf("%.2f", expectedConfidence)
	actualRounded := fmt.Sprintf("%.2f", actualConfidence)
	if expectedRounded != actualRounded {
		t.Fatalf("[validateConfidenceScoreCorrectness] incorrect score, expected %v found %v", expectedRounded, actualRounded)
	}
}
