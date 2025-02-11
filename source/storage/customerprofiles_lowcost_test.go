// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	aurora "tah/upt/source/tah-core/aurora"
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	secret "tah/upt/source/tah-core/secret"
	changeEvent "tah/upt/source/ucp-common/src/model/change-event"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"

	"github.com/google/uuid"
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

func areSlicesSame(slice1, slice2 []string) bool {
	// Check if lengths of the slices are different
	if len(slice1) != len(slice2) {
		return false
	}

	countMap1 := make(map[string]int)
	countMap2 := make(map[string]int)

	for _, item := range slice1 {
		countMap1[item]++
	}

	for _, item := range slice2 {
		countMap2[item]++
	}

	// Check if the maps are equal
	return reflect.DeepEqual(countMap1, countMap2)
}

func SetupAurora(t *testing.T) (*aurora.PostgresDBConfig, error) {
	////////////
	//TEST SETUP
	////////////
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	if envCfg.AuroraClusterName == "" {
		// use local cluster name if set in env.json
		t.Fatalf("error AuroraClusterName is not set in env.json")
	}
	clusterID := envCfg.AuroraClusterName

	if envCfg.AuroraDbName == "" {
		t.Fatalf("error AuroraDbName is not set in env.json")
	}
	dbName := envCfg.AuroraDbName
	auroraCfg := aurora.InitAurora(envCfg.Region, "test_solution_id", "test_solution_version", core.LogLevelDebug)
	auroraCfg.SetTx(core.NewTransaction(t.Name(), "", core.LogLevelDebug))
	secretCfg := secret.InitWithRegion("", envCfg.Region, "", "")
	var posgres *aurora.PostgresDBConfig

	secretArn, err := secretCfg.GetSecretArn("aurora-" + clusterID)
	if err != nil {
		log.Printf("[Aurora] Error getting secret arn: %+v ", err)
		return &aurora.PostgresDBConfig{}, err
	}
	pwd := secret.InitWithRegion(secretArn, envCfg.Region, "", "").Get("password")
	uname := secret.InitWithRegion(secretArn, envCfg.Region, "", "").Get("username")
	if envCfg.AuroraEndpoint == "localhost" {
		posgres = aurora.InitPostgresql(envCfg.AuroraEndpoint, dbName, uname, pwd, 50, core.LogLevelDebug)
	} else {
		cluster, err := auroraCfg.GetCluster(clusterID)
		if err != nil {
			t.Logf("Error: Test cluster does not exists. Run the aurora test first top get the test cluster created")
			return &aurora.PostgresDBConfig{}, fmt.Errorf(
				"Test cluster does not exists. Run the aurora test first to get the test cluster created",
			)
		} else {
			log.Printf("Test cluster already exists. Running tests.")
			auroraCfg.SecretStoreArn, err = secretCfg.GetSecretArn("aurora-" + clusterID)
			if err != nil {
				log.Printf("[customerprofiles] Error getting secret arn: %+v ", err)
				return &aurora.PostgresDBConfig{}, err
			}
			log.Printf("Secret arn for accessing aurora: %s", auroraCfg.SecretStoreArn)
			log.Printf("Cluster arn for accessing aurora: %s", cluster.Arn)
			auroraCfg.ClusterArn = cluster.Arn
			auroraCfg.DbName = dbName
			log.Printf("Cluster arn for accessing aurora: %s", cluster.Arn)
			posgres = aurora.InitPostgresql(cluster.WriterEndpoint, dbName, uname, pwd, 50, core.LogLevelDebug)
		}
	}
	if posgres == nil {
		log.Printf("[Aurora] Error initializing postgresql")
		return &aurora.PostgresDBConfig{}, errors.New("Error initializing postgresql")
	}
	if posgres.ConErr() != nil {
		log.Printf("[Aurora] Error connecting to postgresql: %+v ", posgres.ConErr())
		return &aurora.PostgresDBConfig{}, fmt.Errorf("error connecting to aurora: %v", posgres.ConErr())
	}
	// pinging in unit tests since it requires specific setup to connect locally.
	err = posgres.Ping()
	if err != nil {
		log.Printf("[Aurora] Unable to connect to PostgreSQL database: %v", err)
		log.Printf("Validate the cluster has been set up and you are connected to the Corp network.")
		return &aurora.PostgresDBConfig{}, fmt.Errorf("error connecting to aurora: %v", err)
	}
	return posgres, nil
}

func TestLowCostMergeInteractionRecords(t *testing.T) {
	t.Parallel()
	now := time.Now()

	testSet1 := [][]map[string]interface{}{
		{
			{"id": "1", "timestamp": now.Add(0 * time.Second)},
			{"id": "2", "timestamp": now.Add(2 * time.Second)},
		},
		{
			{"id": "3", "timestamp": now.Add(1 * time.Second)},
			{"id": "4", "timestamp": now.Add(3 * time.Second)},
		},
	}
	expected1 := "1-3-2-4"
	testSet2 := [][]map[string]interface{}{
		{
			{"id": "1", "timestamp": now.Add(0 * time.Second)},
			{"id": "2", "timestamp": now.Add(2 * time.Second)},
		},
		{
			{"id": "3", "timestamp": now.Add(1 * time.Second)},
		},
	}
	expected2 := "1-3-2"
	testSets := [][][]map[string]interface{}{testSet1, testSet2}
	expected := []string{expected1, expected2}
	for i, testSet := range testSets {
		count := 0
		for _, set := range testSet {
			count += len(set)
		}
		res1 := MergeInteractionRecords(testSet)
		if len(res1) != count {
			t.Errorf("Expected %v records, got %d", count, len(res1))
			return
		}

		strs := []string{}
		for _, r := range res1 {
			strs = append(strs, r["id"].(string))
		}
		if strings.Join(strs, "-") != expected[i] {
			t.Errorf("Expected %v, got %v", expected[i], strings.Join(strs, "-"))
		}
	}
}

func TestLowCostBasicPutProfile(t *testing.T) {
	t.Parallel()
	testName := t.Name()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("[%s] error loading env config: %v", testName, err)
	}
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[%s] Error setting up aurora: %+v ", testName, err)
	}
	dynamoTableName := randomTable("customer-profiles-low-cost")
	cfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", testName, err)
	}
	t.Cleanup(func() {
		err = cfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] error deleting dynamo table: %v", testName, err)
		}
	})
	err = cfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] error waiting for table creation: %v", testName, err)
	}

	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)

	mergeQueue := buildTestMergeQueue(t, envCfg.Region)
	cpWriterQueue := buildTestCPWriterQueue(t, envCfg.Region)
	initOptions := CustomerProfileInitOptions{MergeQueueClient: mergeQueue}
	lc := InitLowCost(
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

	////////////
	//TEST RUN
	////////////
	testDomainNameValid := randomDomain("tdn_basic_put")
	err = lc.CreateDomainWithQueue(testDomainNameValid, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "test"}, "", "", DomainOptions{
		DynamoMode: func(b bool) *bool { return &b }(true), // CreateDomainWithQueue defaults DynamoMode to true; adding here to be explicit
	})
	if err != nil {
		t.Fatalf("[%s] error creating domain: %v", testName, err)
	}
	t.Cleanup(func() {
		err = lc.DeleteDomainByName(testDomainNameValid)
		if err != nil {
			t.Errorf("[%s] error deleting domain: %v", testName, err)
		}
	})

	//	Create Dynamo CFG Object to directly interface with cache table
	cache := db.Init("ucp_domain_"+testDomainNameValid, "connect_id", "record_type", "", "")

	err = lc.CreateMapping("air_booking", "air booking mapping", []FieldMapping{
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
			Type:    "STRING",
			Source:  "_source.price",
			Target:  "price",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.extra_mapped_field",
			Target:  "extra_mapped_field",
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        MappingTypeString,
			Source:      "_source." + LAST_UPDATED_FIELD,
			Target:      LAST_UPDATED_FIELD,
			Searcheable: true,
		},
	})
	if err != nil {
		t.Fatalf("[%s] error creating mapping: %v", testName, err)
	}

	err = lc.CreateMapping("hotel_booking", "hotel booking mapping", []FieldMapping{
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
			Source:  "_source.startDate",
			Target:  "startDate",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.firstName",
			Target:  "firstName",
			KeyOnly: true,
		},
		{
			Type:   "STRING",
			Source: "_source.lastName",
			Target: "_profile.LastName",
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        MappingTypeString,
			Source:      "_source." + LAST_UPDATED_FIELD,
			Target:      LAST_UPDATED_FIELD,
			Searcheable: true,
		},
	})
	if err != nil {
		t.Fatalf("[%s] error creating second mapping: %v", testName, err)
	}

	var cid1, cid2 string

	obj, _ := json.Marshal(map[string]interface{}{
		"traveler_id":    "test_traveler_id1",
		"accp_object_id": "test_accp_object_id",
		"segment_id":     "test_segment_id",
		"from":           "test_from",
		"arrival_date":   "test_arrival_date",
		"price":          100,
		"email":          "test_email@example.com",
		"last_updated":   time.Now().UTC().AddDate(0, 0, -1).Format(INGEST_TIMESTAMP_FORMAT),
	})
	err = lc.PutProfileObject(string(obj), "air_booking")
	if err != nil {
		t.Fatalf("[%s] error creating profile object: %v", testName, err)
	}

	res, err := lc.SearchProfiles("profile_id", []string{"test_traveler_id1"})
	if err != nil {
		t.Fatalf("[%s] error searching profiles: %v", testName, err)
	}
	if len(res) != 1 {
		t.Fatalf("[%s] search should have 1 record and not %v", testName, len(res))
	} else {
		log.Printf("Search profile Result: %+v", res)
		cid1 = res[0].ProfileId
		if res[0].Attributes["traveler_id"] != "test_traveler_id1" {
			t.Fatalf("[%s] profile id should be test_traveler_id1 and not %v", testName, res[0].ProfileId)
		}
	}
	log.Printf("cid1: %v", cid1)

	_, err = lc.GetProfileByProfileId("invalid_id", []string{}, []PaginationOptions{})
	if err == nil {
		t.Fatalf("[TestBasicPutProfile] retrieve invalid profile id should return an error")

	}

	obj, _ = json.Marshal(map[string]interface{}{
		"traveler_id":    "test_traveler_id2",
		"accp_object_id": "test_accp_object_id2",
		"segment_id":     "test_segment_id",
		"from":           "test_from",
		"arrival_date":   time.Now().UTC().Format(time.RFC3339Nano),
		"price":          100.01,
		"email":          "test_email2@example.com",
		"last_updated":   time.Now().UTC().AddDate(-1, 0, 0).Format(INGEST_TIMESTAMP_FORMAT),
	})
	err = lc.PutProfileObject(string(obj), "air_booking")
	if err != nil {
		t.Fatalf("[%s] error creating profile object: %v", testName, err)
	}

	res, err = lc.SearchProfiles("profile_id", []string{"test_traveler_id2"})
	if err != nil {
		t.Fatalf("[%s] error searching profiles: %v", testName, err)
	}
	if len(res) != 1 {
		t.Fatalf("[%s] search should have 1 record and not %v", testName, len(res))
	} else {
		log.Printf("Search profile Result: %+v", res)
		cid2 = res[0].ProfileId
		if res[0].Attributes["traveler_id"] != "test_traveler_id2" {
			t.Fatalf("[%s] profile id should be test_traveler_id2 and not %v", testName, res[0].Attributes["traveler_id"])
		}
	}
	log.Printf("cid2: %v", cid2)

	_, err = lc.MergeProfiles(cid1, cid2, ProfileMergeContext{})
	if err != nil {
		t.Fatalf("[%s] error merging profiles: %v", testName, err)
	}

	res1, err := lc.SearchProfiles("profile_id", []string{"test_traveler_id1"})
	if err != nil {
		t.Fatalf("[%s] error searching profiles: %v", testName, err)
	}
	if len(res1) != 1 {
		t.Fatalf("[%s] search should have 1 record and not %v", testName, len(res1))
	}
	log.Printf("Search profile 1 Result after merge: %+v", res1)

	res2, err := lc.SearchProfiles("profile_id", []string{"test_traveler_id2"})
	if err != nil {
		t.Fatalf("[%s] error searching profiles: %v", testName, err)
	}
	if len(res2) != 1 {
		t.Fatalf("[%s] search should have 1 record and not %v", testName, len(res2))
	} else {
		log.Printf("Search profile 2 Result after merge: %+v", res2)
		if res2[0].ProfileId != res1[0].ProfileId {
			t.Fatalf("[%s] connect ID should be the same for porifle 1 and 2 after merge operation", testName)
		}
	}

	profile, err := lc.GetProfile(cid1, []string{"air_booking"}, []PaginationOptions{})
	if err != nil {
		t.Fatalf("[%s] error getting profile: %v", testName, err)
	}
	if profile.ProfileId != cid1 {
		t.Fatalf("[%s] connect id should be %v and not %v", testName, cid1, profile.ProfileId)
	}
	if len(profile.ProfileObjects) != 2 {
		t.Fatalf("[%s] profile should have 2 objects and not %v", testName, len(profile.ProfileObjects))
	}

	profile, err = lc.GetProfile(cid2, []string{"air_booking"}, []PaginationOptions{})
	if err != nil {
		t.Fatalf("[%s] connect_id %s should now point to %s", testName, cid2, cid1)
	}
	if len(profile.ProfileObjects) != 2 {
		t.Fatalf("[%s] profile should have 2 objects and not %v", testName, len(profile.ProfileObjects))
	}
	//	check cached profile
	var cachedProfile DynamoProfileRecord
	err = cache.Get(cid1, "profile_main", &cachedProfile)
	if err != nil {
		t.Fatalf("[%s] error retreiving profile from dynamo cache: %v", testName, err)
	}
	if cachedProfile.PersonalEmailAddress != "test_email@example.com" {
		t.Fatalf("[%s] expected: %s got: %s", testName, "test_email@example.com", cachedProfile.PersonalEmailAddress)
	}

	//inserting new object with travelr_id2
	hotelBookingObject, _ := json.Marshal(map[string]string{
		"traveler_id":    "test_traveler_id2",
		"accp_object_id": "hotel_booking_1",
		"booking_id":     "IG73987KJH",
		"startDate":      "20MAR2025",
		"firstName":      "john",
		"lastName":       "doe",
		"email":          "test_email3@example.com",
		"last_updated":   time.Now().UTC().AddDate(0, 0, -2).Format(INGEST_TIMESTAMP_FORMAT),
	})
	err = lc.PutProfileObject(string(hotelBookingObject), "hotel_booking")
	if err != nil {
		t.Fatalf("[%s] error creating profile object: %v", testName, err)
	}
	profile, err = lc.GetProfile(cid1, []string{"hotel_booking"}, []PaginationOptions{})
	if err != nil {
		t.Fatalf("[%s] error retreiving profile with cid %v: %v", testName, cid1, err)
	}
	//we check that the profile object has been added to the profile with cid1 (since test_traveler_id2 has ben merged into it)
	found := true
	for _, o := range profile.ProfileObjects {
		if o.ID == "hotel_booking_1" && o.Type == "hotel_booking" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("[%s] profile with %v should have the hotel_booking_1 object", testName, cid1)
	}

	err = cache.Get(cid1, "profile_main", &cachedProfile)
	if err != nil {
		t.Fatalf("[%s] error retreiving profile from dynamo cache: %v", t.Name(), err)
	}
	//tests added for test coverage purpose
	err = lc.ClearDynamoCache()
	if err != nil {
		t.Fatalf("[TestBasicPutProfile] error clearing dynamo cache: %v", err)
	}

	err = lc.ClearCustomerProfileCache(map[string]string{"envName": "dev"}, "")
	if err == nil {
		t.Fatalf("[TestBasicPutProfile] clearing CP cache should return an error when CP not enabled")
	}
	if err.Error() != "CP cache is not enabled" {
		t.Fatalf("[TestBasicPutProfile] clearing CP cache should return 'CP cache is not enabled' and not '%s'", err.Error())
	}
	err = lc.CacheProfile(cid1, 1)
	if err != nil {
		t.Fatalf("[TestBasicPutProfile] error clearing dynamo cache: %v", err)
	}
}

// Happy path only, more tests to be added
func TestLowCostHappyPath(t *testing.T) {
	t.Parallel()
	testName := t.Name()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("[%s] error loading env config: %v", testName, err)
	}
	// Create required resources
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[%s] Error setting up aurora: %+v ", testName, err)
	}
	dynamoTableName := randomTable("profiles-low-cost")
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", testName, err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] error deleting dynamo table: %v", testName, err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] error waiting for table creation: %v", testName, err)
	}
	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)

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
	domainName := randomDomain("lcs_happy_path")
	envNameTest := "testenv"

	// Validate domain is not cached before GetDomain()
	_, found := lc.Config.retrieveFromCache(domainName, controlPlaneCacheType_DomainConfig)
	if found {
		t.Fatalf("[%s] domain config found in cache", testName)
	}

	// Create domain
	err = lc.CreateDomainWithQueue(
		domainName,
		"",
		map[string]string{DOMAIN_CONFIG_ENV_KEY: envNameTest},
		"",
		"",
		DomainOptions{RuleBaseIdResolutionOn: true},
	)
	if err != nil {
		t.Errorf("[%s] error creating domain: %v", testName, err)
	}

	//////////////
	// Cleanup logic to be executed even if we return early
	//////////////////
	t.Cleanup(func() {
		log.Printf("")
		log.Printf("**************************************")
		log.Printf("******* CLEANING UP TEST %s **********", t.Name())
		log.Printf("**************************************")
		log.Printf("")
		err = lc.DeleteDomain()
		if err != nil {
			t.Errorf("[%s] error deleting domain: %v", testName, err)
		}
	})

	// Get domain
	domain, err := lc.GetDomain()
	if err != nil {
		t.Fatalf("[%s] error getting domain: %v", testName, err)
	}
	if domain.Name != domainName || domain.Tags[DOMAIN_CONFIG_ENV_KEY] != envNameTest {
		t.Fatalf("[%s] unexpected GetDomain result", testName)
	}
	// Validate domain is cached
	_, found = lc.Config.retrieveFromCache(domainName, controlPlaneCacheType_DomainConfig)
	if !found {
		t.Fatalf("[%s] domain config not found in cache", testName)
	}

	// List domains
	domains, err := lc.ListDomains()
	if err != nil {
		t.Fatalf("[%s] error listing domains: %v", testName, err)
	}
	if len(domains) != 1 || domains[0].Name != domainName || domains[0].Tags[DOMAIN_CONFIG_ENV_KEY] != envNameTest {
		t.Fatalf("[%s] unexpected ListDomains result", testName)
	}

	// Create mapping
	objectTypeName := "air_booking"
	testMappings := []FieldMapping{
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
			Type:    "STRING",
			Source:  "_source.price",
			Target:  "price",
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        MappingTypeString,
			Source:      "_source." + LAST_UPDATED_FIELD,
			Target:      LAST_UPDATED_FIELD,
			Searcheable: true,
		},
	}
	err = lc.CreateMapping(objectTypeName, "Mapping for "+objectTypeName, testMappings)
	if err != nil {
		t.Errorf("[%s] error creating mapping: %v", testName, err)
	}

	// Validate no mappings are cached
	_, found = lc.Config.retrieveFromCache(domainName, controlPlaneCacheType_Mappings)
	if found {
		t.Fatalf("[%s] mappings found in cache", testName)
	}
	// Get mappings
	mappings, err := lc.GetMappings()
	if err != nil {
		t.Fatalf("[%s] error getting mappings: %v", testName, err)
	}
	if len(mappings) != 1 {
		t.Fatalf("[%s] unexpected GetMappings result", testName)
		return
	}
	for _, mapping := range mappings {
		m, err := lc.GetMapping(mapping.Name)
		if err != nil {
			t.Fatalf("[%s] error getting mapping: %v", testName, err)
		}
		if m.Name != mapping.Name {
			t.Fatalf("[%s] unexpected GetMapping result", testName)
		}
	}
	// Validate mappings are cached
	cachedMappings, found := lc.Config.retrieveFromCache(domainName, controlPlaneCacheType_Mappings)
	if !found {
		t.Fatalf("[%s] mappings not found in cache", testName)
	}
	if cachedMappingsCast, ok := cachedMappings.([]ObjectMapping); ok {
		if len(cachedMappingsCast) != len(mappings) {
			t.Fatalf("[%s] unexpected GetMappings result", testName)
		}
	} else {
		t.Fatalf("[%s] unexpected GetMappings result", testName)
	}

	// Enable/Disable ID Res
	err = lc.DisableRuleBasedIdRes()
	if err != nil {
		t.Fatalf("[%s] error disabling rule based id res: %v", testName, err)
	}

	err = lc.EnableRuleBasedIdRes()
	if err != nil {
		t.Fatalf("[%s] error enabling rule based id res: %v", testName, err)
	}

	// Testing Rule Sets
	testRules := []Rule{
		{
			Index:       1,
			Name:        "test_rule_1",
			Description: "test_rule_1",
			Conditions: []Condition{
				{
					ConditionType:        CONDITION_TYPE_MATCH,
					IncomingObjectType:   objectTypeName,
					IncomingObjectField:  "arrival_date",
					Op:                   RULE_OP_EQUALS,
					IncomingObjectType2:  "_profile",
					IncomingObjectField2: "PersonalEmailAddress",
				},
			},
		},
		{
			Index:       2,
			Name:        "test_rule_2",
			Description: "test_rule_2",
			Conditions: []Condition{
				{
					ConditionType:        CONDITION_TYPE_MATCH,
					IncomingObjectType:   objectTypeName,
					IncomingObjectField:  "arrival_date",
					Op:                   RULE_OP_EQUALS,
					IncomingObjectType2:  "_profile",
					IncomingObjectField2: "PersonalEmailAddress",
				},
			},
		},
		{
			Index:       3,
			Name:        "test_rule_3",
			Description: "test_rule_3",
			Conditions: []Condition{
				{
					ConditionType:        CONDITION_TYPE_MATCH,
					IncomingObjectType:   objectTypeName,
					IncomingObjectField:  "arrival_date",
					Op:                   RULE_OP_EQUALS,
					IncomingObjectType2:  "_profile",
					IncomingObjectField2: "PersonalEmailAddress",
				},
			},
		},
	}
	// Save is res rule set
	err = lc.SaveIdResRuleSet(testRules)
	if err != nil {
		t.Fatalf("[%s] error creating saving new rule set: %v", testName, err)
	}

	listRuleSetResult0, err := lc.ListIdResRuleSets(false)
	if err != nil {
		t.Fatalf("[%s] error listing rule sets: %v", testName, err)
	}
	if len(listRuleSetResult0) == 0 {
		t.Fatalf("[%s] should have at least one ruleset", testName)
		return
	}
	// Validate empty id res rule set is cached
	cachedIdResRuleSet, found := lc.Config.retrieveFromCache(domainName, controlPlaneCacheType_ActiveStitchingRuleSet)
	if !found {
		t.Fatalf("[%s] active rule set not found in cache", testName)
	}
	if cachedIdResRuleSetCast, ok := cachedIdResRuleSet.(RuleSet); ok {
		if len(cachedIdResRuleSetCast.Rules) != 0 {
			t.Fatalf("[%s] expected 0 rules in active cache until rules are activated", testName)
		}
	} else {
		t.Fatalf("[%s] unexpected ListIdResRuleSets result", testName)
	}

	if listRuleSetResult0[0].Name != "active" {
		t.Fatalf("[%s] wrong rule set name returned", testName)
	}
	if len(listRuleSetResult0[0].Rules) != 0 {
		t.Fatalf("[%s] wrong number of rules returned for active rules", testName)
	}
	if listRuleSetResult0[0].LatestVersion != 0 {
		t.Fatalf("[%s] wrong latest version returned for active rules", testName)
	}
	if !listRuleSetResult0[0].LastActivationTimestamp.IsZero() {
		t.Fatalf("[%s] no activated rule set, response should have zero value for last activated timestamp", testName)
	}
	if listRuleSetResult0[1].Name != "draft" {
		t.Fatalf("[%s] wrong rule set name returned", testName)
	}
	if len(listRuleSetResult0[1].Rules) != 3 {
		t.Fatalf("[%s] wrong number of rules returned for draft rules", testName)
	}
	if listRuleSetResult0[1].LatestVersion != 0 {
		t.Fatalf("[%s] wrong latest version returned for draft rules", testName)
	}
	if !listRuleSetResult0[1].LastActivationTimestamp.IsZero() {
		t.Fatalf("[%s] draft rule set should have zero value for last activated timestamp", testName)
	}
	for index, rule := range listRuleSetResult0[1].Rules {
		if rule.Index != testRules[index].Index {
			t.Fatalf("[%s] index returned in incorrect order", testName)
		}
	}

	err = lc.ActivateIdResRuleSet()
	if err != nil {
		t.Fatalf("[%s] error activating rule set: %v", testName, err)
	}

	listRuleSetResult1, err := lc.ListIdResRuleSets(false)
	if err != nil {
		t.Fatalf("[%s] error listing rule sets: %v", testName, err)
	}
	if len(listRuleSetResult1) == 0 {
		t.Fatalf("[%s] should have at least one ruleset", testName)
	}
	if listRuleSetResult1[0].Name != "active" {
		t.Fatalf("[%s] wrong rule set name returned", testName)
	}
	if len(listRuleSetResult1[0].Rules) != 3 {
		t.Fatalf("[%s] wrong number of rules returned for active rules", testName)
	}
	if listRuleSetResult1[0].LatestVersion != 1 {
		t.Fatalf("[%s] wrong latest version returned for active rules", testName)
	}
	if listRuleSetResult1[0].LastActivationTimestamp.IsZero() {
		t.Fatalf("[%s] Activated rule set should not have zero value for last activated timestamp", testName)
	}
	for index, rule := range listRuleSetResult1[0].Rules {
		if rule.Index != testRules[index].Index {
			t.Fatalf("[%s] index returned in incorrect order", testName)
		}
	}
	if listRuleSetResult1[1].Name != "draft" {
		t.Fatalf("[%s] wrong rule set name returned", testName)
	}
	if len(listRuleSetResult1[1].Rules) != 0 {
		t.Fatalf("[%s] wrong number of rules returned for draft rules", testName)
	}
	if listRuleSetResult1[1].LatestVersion != 0 {
		t.Fatalf("[%s] wrong latest version returned for draft rules", testName)
	}
	if !listRuleSetResult1[1].LastActivationTimestamp.IsZero() {
		t.Fatalf("[%s] draft rule set should have zero value for last activated timestamp", testName)
	}
	// Validate id res rule set has been updated
	cachedIdResRuleSet, found = lc.Config.retrieveFromCache(domainName, controlPlaneCacheType_ActiveStitchingRuleSet)
	if !found {
		t.Fatalf("[%s] active rule set not found in cache", testName)
	}
	if cachedIdResRuleSetCast, ok := cachedIdResRuleSet.(RuleSet); ok {
		if len(cachedIdResRuleSetCast.Rules) != len(testRules) {
			t.Fatalf("[%s] expected cache to match active", testName)
		}
	} else {
		t.Fatalf("[%s] unexpected ListIdResRuleSets result", testName)
	}

	v1ActivationTime := listRuleSetResult1[0].LastActivationTimestamp

	testRules = append(testRules, Rule{
		Index:       4,
		Name:        "test_rule_4",
		Description: "test_rule_4",
		Conditions: []Condition{
			{
				ConditionType:        CONDITION_TYPE_MATCH,
				IncomingObjectType:   objectTypeName,
				IncomingObjectField:  "arrival_date",
				Op:                   RULE_OP_EQUALS,
				IncomingObjectType2:  "_profile",
				IncomingObjectField2: "PersonalEmailAddress",
			},
		},
	})

	err = lc.SaveIdResRuleSet(testRules)
	if err != nil {
		t.Fatalf("[%s] error creating saving new rule set: %v", testName, err)
	}
	listRuleSetResult2, err := lc.ListIdResRuleSets(false)
	if err != nil {
		t.Fatalf("[%s] error listing rule sets: %v", testName, err)
	}
	if len(listRuleSetResult2) == 0 {
		t.Fatalf("[%s] should have at least one ruleset", testName)
	}
	if listRuleSetResult2[0].Name != "active" {
		t.Fatalf("[%s] wrong rule set name returned", testName)
	}
	if len(listRuleSetResult2[0].Rules) != 3 {
		t.Fatalf("[%s] wrong number of rules returned for active rules", testName)
	}
	if listRuleSetResult2[0].LatestVersion != 1 {
		t.Fatalf("[%s] wrong latest version returned for active rules", testName)
	}
	if listRuleSetResult2[0].LastActivationTimestamp.IsZero() {
		t.Fatalf("[%s] Activated rule set should not have zero value for last activated timestamp", testName)
	}
	for index, rule := range listRuleSetResult2[0].Rules {
		if rule.Index != testRules[index].Index {
			t.Fatalf("[%s] index returned in incorrect order", testName)
		}
	}
	if listRuleSetResult2[1].Name != "draft" {
		t.Fatalf("[%s] wrong rule set name returned", testName)
	}
	if len(listRuleSetResult2[1].Rules) != 4 {
		t.Fatalf("[%s] wrong number of rules returned for draft rules", testName)
	}
	if listRuleSetResult2[1].LatestVersion != 0 {
		t.Fatalf("[%s] wrong latest version returned for draft rules", testName)
	}
	if !listRuleSetResult2[1].LastActivationTimestamp.IsZero() {
		t.Fatalf("[%s] Draft rule set should have zero value for last activated timestamp", testName)
	}
	for index, rule := range listRuleSetResult2[1].Rules {
		if rule.Index != testRules[index].Index {
			t.Fatalf("[%s] index returned in incorrect order", testName)
		}
	}

	err = lc.ActivateIdResRuleSet()
	if err != nil {
		t.Fatalf("[%s] error activating rule set: %v", testName, err)
	}

	listRuleSetResult3, err := lc.ListIdResRuleSets(true)
	if err != nil {
		t.Fatalf("[%s] error listing rule sets: %v", testName, err)
	}
	if len(listRuleSetResult3) != 2 {
		lc.Tx.Info("[%s] listRuleSetResult3 %v", listRuleSetResult3)
		t.Fatalf("[%s] wrong number of rule sets returned", testName)
		return
	}
	if listRuleSetResult3[0].Name != "v1" {
		t.Fatalf("[%s] wrong rule set name returned", testName)
	}
	if len(listRuleSetResult3[0].Rules) != 3 {
		t.Fatalf("[%s] wrong number of rules returned for historical rules", testName)
	}
	if listRuleSetResult3[0].LatestVersion != 0 {
		t.Fatalf("[%s] wrong latest version returned for historical rules", testName)
	}
	if listRuleSetResult3[0].LastActivationTimestamp != v1ActivationTime {
		t.Fatalf("[%s] Historical rule set should have same timestamp as when v1 was activated", testName)
	}
	for index, rule := range listRuleSetResult3[0].Rules {
		if rule.Index != testRules[index].Index {
			t.Fatalf("[%s] index returned in incorrect order", testName)
		}
	}
	if listRuleSetResult3[1].Name != "active" {
		t.Fatalf("[%s] wrong rule set name returned", testName)
	}
	if len(listRuleSetResult3[1].Rules) != 4 {
		t.Fatalf("[%s] wrong number of rules returned for draft rules", testName)
	}
	if listRuleSetResult3[1].LatestVersion != 2 {
		t.Fatalf("[%s] wrong latest version returned for draft rules", testName)
	}
	if listRuleSetResult3[1].LastActivationTimestamp.IsZero() {
		t.Fatalf("[%s] Active rule set should not have zero value for last activated timestamp", testName)
	}
	for index, rule := range listRuleSetResult3[1].Rules {
		if rule.Index != testRules[index].Index {
			t.Fatalf("[%s] index returned in incorrect order", testName)
		}
	}
	err = lc.DisableRuleBasedIdRes()
	if err != nil {
		t.Fatalf("[%s] error disabling rule based id res: %v", testName, err)
	}

	// DATA PLANE TESTING

	// Data plane variables
	profileOneTravelerID := "test_traveler_id1"
	profileTwoTravelerID := "test_traveler_id2"
	obj1, _ := json.Marshal(map[string]interface{}{
		"traveler_id":      profileOneTravelerID,
		"accp_object_id":   "test_accp_object_id",
		"segment_id":       "test_segment_id",
		"from":             "test_from",
		"arrival_date":     "test_arrival_date",
		"price":            100,
		"email":            "test_email@example.com",
		LAST_UPDATED_FIELD: utcNow.Format(INGEST_TIMESTAMP_FORMAT),
	})
	obj2, _ := json.Marshal(map[string]interface{}{
		"traveler_id":      profileTwoTravelerID,
		"accp_object_id":   "test_accp_object_id2",
		"segment_id":       "test_segment_id",
		"from":             "test_from",
		"arrival_date":     time.Now().Format(time.DateOnly),
		"price":            100.01,
		"email":            "test_email2@example.com",
		LAST_UPDATED_FIELD: utcNow.Format(INGEST_TIMESTAMP_FORMAT),
	})
	obj3, _ := json.Marshal(map[string]interface{}{
		"traveler_id":      profileTwoTravelerID,
		"accp_object_id":   "test_accp_object_id3",
		"segment_id":       "test_segment_id2",
		"from":             "test_from2",
		"arrival_date":     "test_arrival_date2",
		"email":            "test_email2@example.com",
		LAST_UPDATED_FIELD: utcNow.Format(INGEST_TIMESTAMP_FORMAT),
	})
	unmappedObj, _ := json.Marshal(map[string]interface{}{
		"traveler_id":      profileTwoTravelerID,
		"accp_object_id":   "test_accp_object_id4",
		"segment_id":       "test_segment_id3",
		"from":             "test_from3",
		"arrival_date":     "test_arrival_date3",
		"email":            "test_email2@example.com",
		"unmapped_field":   "unmapped_value",
		LAST_UPDATED_FIELD: utcNow.Format(INGEST_TIMESTAMP_FORMAT),
	})

	objects := []string{string(obj1), string(obj2), string(obj3), string(unmappedObj)}

	expectedProfileCounts := []int64{1, 2, 2, 2}
	expectedObjectCounts := []int64{1, 2, 3, 3}
	expectedSuccess := []bool{true, true, true, false}

	if len(objects) != len(expectedProfileCounts) || len(objects) != len(expectedObjectCounts) || len(objects) != len(expectedSuccess) {
		t.Errorf("[TestHappyPath] wrong number of test objects. should have %d", len(objects))
	}

	for i, obj := range objects {
		err = lc.PutProfileObject(obj, objectTypeName)
		if err != nil && expectedSuccess[i] {
			t.Errorf("[TestHappyPath] error creating profile object: %v", err)
		}
		dom, err := lc.GetDomain()
		if err != nil {
			t.Errorf("[TestHappyPath] error getting domain to validate count: %v", err)
		}
		if dom.NProfiles != expectedProfileCounts[i] {
			t.Errorf("[TestHappyPath] wrong number of profiles in domain: %v", dom.NProfiles)
		}
		if dom.NObjects != expectedObjectCounts[i] {
			t.Errorf("[TestHappyPath] wrong number of profile objects in domain: %v", dom.NObjects)
		}
	}
	partition := UuidPartition{
		LowerBound: uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000000")),
		UpperBound: uuid.Must(uuid.Parse("FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF")),
	}
	cids, err := lc.GetProfilePartition(partition)
	if err != nil {
		t.Fatalf("[TestHappyPath] error getting profile partition: %v", err)
	}
	if len(cids) != 2 {
		t.Fatalf("[TestHappyPath] wrong number of profile in partition %v. should have %d and have %d", partition, 2, len(cids))
	}

	// Search profiles
	profileOneConnectID := ""
	res, err := lc.SearchProfiles("profile_id", []string{profileOneTravelerID})
	if err != nil {
		t.Errorf("[%s] error searching profiles: %v", testName, err)
	}
	if len(res) != 1 {
		t.Errorf("[%s] search should have 1 record and not %v", testName, len(res))
	} else {
		log.Printf("Search profile Result: %+v", res)
		profileOneConnectID = res[0].ProfileId
		if res[0].Attributes["traveler_id"] != profileOneTravelerID {
			t.Errorf("[%s] profile id should be %s and not %v", testName, profileOneTravelerID, res[0].Attributes["traveler_id"])
		}
	}

	res, err = lc.SearchProfiles("air_booking.from", []string{"test_from"})
	if err != nil {
		t.Errorf("[%s] error searching profiles: %v", testName, err)
	}
	if len(res) != 2 {
		t.Errorf("[%s] search should have 2 record and not %v", testName, len(res))
	} else {
		log.Printf("Search profile Result: %+v", res)
		eqalityTest := areSlicesSame([]string{profileOneTravelerID, profileTwoTravelerID}, []string{res[0].Attributes["traveler_id"], res[1].Attributes["traveler_id"]})
		if !eqalityTest {
			t.Errorf("[%s] profile id should be %s and not %v", testName, profileOneTravelerID, res[0].Attributes["traveler_id"])
		}
	}

	profileTwoConnectID := ""
	res, err = lc.SearchProfiles("profile_id", []string{profileTwoTravelerID})
	if err != nil {
		t.Errorf("[%s] error searching profiles: %v", testName, err)
	}
	if len(res) != 1 {
		t.Errorf("[%s] search should have 1 record and not %v", testName, len(res))
	} else {
		log.Printf("Search profile Result: %+v", res)
		profileTwoConnectID = res[0].ProfileId
		if res[0].Attributes["traveler_id"] != profileTwoTravelerID {
			t.Errorf("[%s] profile id should be %s and not %v", testName, profileTwoTravelerID, res[0].Attributes["traveler_id"])
		}
	}

	// Get profile
	profileOne, err := lc.GetProfile(profileOneConnectID, []string{"air_booking"}, []PaginationOptions{})
	if err != nil {
		t.Errorf("[%s] error getting profile: %v", testName, err)
	}
	log.Printf("Profile 1: %v", profileOne)
	if profileOne.Attributes["traveler_id"] != profileOneTravelerID {
		t.Fatalf("[%s] Incorrect traveler id returned", t.Name())
	}
	if profileOne.PersonalEmailAddress != "test_email@example.com" {
		t.Fatalf("[%s] Incorrect email address returned", t.Name())
	}
	profileTwo, err := lc.GetProfile(profileTwoConnectID, []string{"air_booking"}, []PaginationOptions{})
	if err != nil {
		t.Errorf("[%s] error getting profile: %v", testName, err)
	}
	log.Printf("Profile 2: %v", profileTwo)
	if profileTwo.Attributes["traveler_id"] != profileTwoTravelerID {
		t.Fatalf("[%s] Incorrect traveler id returned", t.Name())
	}
	if profileTwo.PersonalEmailAddress != "test_email2@example.com" {
		t.Fatalf("[%s] Incorrect email address returned", t.Name())
	}

	// Get profile by traveler id by Dynamo GSI
	connectId1, err := lc.Data.GetConnectIdByTravelerIdDynamo(profileOneTravelerID)
	if err != nil {
		t.Fatalf("[%s] Failed to retrieve Profile by traveler id %s", t.Name(), profileOneTravelerID)
	}
	connectId2, err := lc.Data.GetConnectIdByTravelerIdDynamo(profileTwoTravelerID)
	if err != nil {
		t.Fatalf("[%s] Failed to retrieve Profile by traveler id %s", t.Name(), profileTwoTravelerID)
	}
	if connectId1 != profileOne.ProfileId {
		t.Fatalf("[%s] Incorrect connect id returned for profile 1, expected %v, found %v", t.Name(), profileOne.ProfileId, connectId1)
	}
	if connectId2 != profileTwo.ProfileId {
		t.Fatalf("[%s] Incorrect connect id returned for profile 2, expected %v, found %v", t.Name(), profileTwo.ProfileId, connectId2)
	}

	// Get specific profile object
	profileObj, err := lc.GetSpecificProfileObject("accp_object_id", "test_accp_object_id", profileOneConnectID, "air_booking")
	if err != nil {
		t.Errorf("[%s] error getting profile object: %v", testName, err)
	}
	if profileObj.ID != "test_accp_object_id" || profileObj.Attributes["traveler_id"] != "test_traveler_id1" ||
		profileObj.Attributes["price"] != "100" {
		t.Errorf("[%s] unexpected profile object returned", testName)
	}

	// Merge profiles
	_, err = lc.MergeProfiles(profileOneConnectID, profileTwoConnectID, ProfileMergeContext{})
	if err != nil {
		t.Errorf("[%s] error merging profiles: %v", testName, err)
	}
	//Get Profile History
	history, err := lc.GetProfileHistory(profileOneConnectID, domainName, 100)
	log.Printf("History for profile %s: %v", profileOneConnectID, history)
	if err != nil {
		t.Errorf("[%s] error getting profile object: %v", testName, err)
	}
	if len(history) == 0 {
		t.Errorf("[%s] profile history should have at least one item", testName)
	}

	originalConnectId, err := lc.FindCurrentParentOfProfile(domainName, profileTwoConnectID)
	if err != nil {
		t.Errorf("[%s] error finding current parent of profile: %v", testName, err)
	}
	if originalConnectId != profileOneConnectID {
		t.Errorf("[%s] wrong parent profile returned. Have %s, should be %s", testName, originalConnectId, profileOneConnectID)
	}

	dom, err := lc.GetDomain()
	if err != nil {
		t.Errorf("[%s] error getting domain to validate count: %v", testName, err)
	}
	if dom.NProfiles != 1 {
		t.Errorf("[%s] wrong number of profiles in domain after merge: %v", testName, dom.NProfiles)
	}
	if dom.NObjects != 3 {
		t.Errorf("[%s] wrong number of profile objects in domain after merge: %v", testName, dom.NObjects)
	}

	//Test Idempotency
	_, err = lc.MergeProfiles(profileOneConnectID, profileTwoConnectID, ProfileMergeContext{})
	if err != nil {
		data, _ := lc.Data.AurSvc.Query(fmt.Sprintf(`SELECT * FROM %s`, dom.Name+"_profile_history"))
		log.Printf("Merge history table dump: %+v", data)
		t.Errorf("[TestHappyPath] error (%s) merging profiles a second time. Merge should be idempotent:", err)
	}

	// Validate change events
	log.Printf("Events sent to Kinesis: %v", kinesisMockCfg.Records)
	if kinesisMockCfg.Records == nil {
		t.Error("Kinesis records list is nil")
	} else {
		// Continue testing change events if we can safely dereference pointer
		if len(*kinesisMockCfg.Records) != 8 {
			t.Errorf("Expected 8 events, got %d", len(*kinesisMockCfg.Records))
		}
		var profileEvents []changeEvent.ProfileChangeRecordWithProfile
		var objectEvents []changeEvent.ProfileChangeRecordWithProfileObject
		for _, record := range *kinesisMockCfg.Records {
			profileEvent, objectEvent, err := kinesisRecToChangeEvent(record)
			if err != nil {
				t.Errorf("[%s] Error getting event: %v", t.Name(), err)
				continue
			}
			if !reflect.DeepEqual(profileEvent, changeEvent.ProfileChangeRecordWithProfile{}) {
				profileEvents = append(profileEvents, profileEvent)
			} else if !reflect.DeepEqual(objectEvent, changeEvent.ProfileChangeRecordWithProfileObject{}) {
				objectEvents = append(objectEvents, objectEvent)
			} else {
				t.Errorf("[%s] Expected a profileEvent or objectEvent to be parsed", t.Name())
			}
		}
		// index 0 is object and profile level create events for first record of profile one
		if objectEvents[0].EventType != EventTypeCreated || objectEvents[0].Object["profile_id"] != profileOneTravelerID || objectEvents[0].ObjectTypeName != "air_booking" {
			t.Fatalf("[%s] Unexpected event at index 0 of object events", t.Name())
		}
		if profileEvents[0].EventType != EventTypeCreated || profileEvents[0].Object.ProfileId != profileOneConnectID || len(profileEvents[0].Object.ProfileObjects) != 1 || profileEvents[0].ObjectTypeName != "_profile" {
			t.Fatalf("[%s] Unexpected event at index 0 of profile events", t.Name())
		}
		// index 1 is object and profile level create events for first record of profile two
		if objectEvents[1].EventType != EventTypeCreated || objectEvents[1].Object["profile_id"] != profileTwoTravelerID || objectEvents[1].ObjectTypeName != "air_booking" {
			t.Fatalf("[%s] Unexpected event at index 1 of object events", t.Name())
		}
		if profileEvents[1].EventType != EventTypeCreated || profileEvents[1].Object.ProfileId != profileTwoConnectID || len(profileEvents[1].Object.ProfileObjects) != 1 || profileEvents[1].ObjectTypeName != "_profile" {
			t.Fatalf("[%s] Unexpected event at index 1 of profile events", t.Name())
		}
		// index 2 is object and profile level update events for second record of profile one
		if objectEvents[2].EventType != EventTypeUpdated || objectEvents[2].Object["profile_id"] != profileTwoTravelerID || objectEvents[2].ObjectTypeName != "air_booking" {
			t.Fatalf("[%s] Unexpected event at index 2 of object events", t.Name())
		}
		if profileEvents[2].EventType != EventTypeUpdated || profileEvents[2].Object.ProfileId != profileTwoConnectID || len(profileEvents[2].Object.ProfileObjects) != 1 || profileEvents[2].ObjectTypeName != "_profile" {
			t.Fatalf("[%s] Unexpected event at index 2 of profile events", t.Name())
		}
		// merge event for target profile
		if profileEvents[3].EventType != EventTypeMerged || profileEvents[3].Object.ProfileId != profileOneConnectID || len(profileEvents[3].Object.ProfileObjects) != 0 || profileEvents[3].ObjectTypeName != "_profile" {
			t.Fatalf("[%s] Unexpected event at index 3 of profile events", t.Name())
		}
		// delete event for source profile
		if profileEvents[4].EventType != EventTypeDeleted || profileEvents[4].Object.ProfileId != profileTwoConnectID || len(profileEvents[4].Object.ProfileObjects) != 0 || profileEvents[4].ObjectTypeName != "_profile" {
			t.Fatalf("[%s] Unexpected event at index 4 of profile events", t.Name())
		}
	}

	//	Delete Profiles
	preDeleteObjectsCount, err := lc.Data.GetAllObjectCounts(domainName)
	if err != nil {
		t.Errorf("[%s] error getting profile objects counts: %v", t.Name(), err)
	}

	err = lc.DeleteProfile(profileOneConnectID, DeleteProfileOptions{DeleteHistory: true})
	if err != nil {
		t.Errorf("[%s] error deleting profile: %v", t.Name(), err)
	}

	profile, err := lc.GetProfile(profileOneConnectID, []string{}, []PaginationOptions{})
	log.Printf("GetProfile expected error after delete: %v", err)
	if err == nil {
		t.Errorf("[%s] profile delete failed: should get an error after deleting profile and not %v", t.Name(), profile)
	}

	_, err = lc.Data.GetProfileByConnectIDDynamo(domainName, profileOneConnectID)
	if err == nil {
		t.Errorf("[%s] should have an error being unable to find the profile: %v", t.Name(), err)
	}

	postDeleteObjectCounts, err := lc.Data.GetAllObjectCounts(domainName)
	if err != nil {
		t.Errorf("[%s] error getting profile objects counts: %v", t.Name(), err)
	}
	// Validate counts
	for key, count := range preDeleteObjectsCount {
		if preDeleteObjectsCount[key]-1 != postDeleteObjectCounts[key] {
			t.Errorf("[%s] wrong count for objectType: %s; expected: %v, got: %v", t.Name(), key, count-1, postDeleteObjectCounts[key])
		}
	}
}

// Test creating domains with invalid names
func TestLowCostInvalidDomainName(t *testing.T) {
	t.Parallel()
	testName := t.Name()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("[%s] error loading env config: %v", testName, err)
	}
	// Create required resources
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[%s] Error setting up aurora: %+v ", testName, err)
	}
	dynamoTableName := randomTable("customer-profiles-low-cost-invalid-domain-test")
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", testName, err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] error deleting dynamo table: %v", testName, err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] error waiting for table creation: %v", testName, err)
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

	log.Printf("Creating domain with invalid name")
	domainName := "a"
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err == nil {
		t.Errorf("[%s] error creating domain: should not have been created", testName)
	} else {
		log.Printf("[%s] did not create invalid domain: %v", testName, err)
	}

	domainName = "This name has spaces"
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err == nil {
		t.Errorf("[%s] error creating domain: should not have been created", testName)
	} else {
		log.Printf("[%s] did not create invalid domain: %v", testName, err)
	}

	domainName = "CAPITAL_LETTERS"
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err == nil {
		t.Errorf("[%s] error creating domain: should not have been created", testName)
	} else {
		log.Printf("[%s] did not create invalid domain: %v", testName, err)
	}

	domainName = "thisisaverylongdomainwithlotsandlotsandlotsofwordsandletterswhichshouldnotwork"
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err == nil {
		t.Errorf("[%s] error creating domain: should not have been created", testName)
	} else {
		log.Printf("[%s] did not create invalid domain: %v", testName, err)
	}

	domainName = "dashes-are-bad"
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err == nil {
		t.Errorf("[%s] error creating domain: should not have been created", testName)
	} else {
		log.Printf("[%s] did not create invalid domain: %v", testName, err)
	}

	domainName = "missing_env_tag"
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{}, "", "", DomainOptions{})
	if err == nil {
		t.Errorf("[%s] error creating domain: should not have been created", testName)
	}

	//Sending multiple create requests
	log.Printf("Sending multiple create requests")
	domainName = randomDomain("happy_domain_1")
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[%s] error creating domain: %v", testName, err)
	}

	secondDomainName := randomDomain("happy_domain_2")
	err = lc.CreateDomainWithQueue(secondDomainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[%s] error creating domain: %v", testName, err)
	}

	thirdDomainName := randomDomain("happy_domain_3")
	err = lc.CreateDomainWithQueue(thirdDomainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[%s] error creating domain: %v", testName, err)
	}

	//Sending multiple delete requests
	log.Printf("Sending multiple delete requests")
	err = lc.DeleteDomainByName(domainName)
	if err != nil {
		t.Errorf("[%s] error deleting domain: %v", testName, err)
	}

	err = lc.DeleteDomainByName(secondDomainName)
	if err != nil {
		t.Errorf("[%s] error deleting domain: %v", testName, err)
	}

	err = lc.DeleteDomainByName(thirdDomainName)
	if err != nil {
		t.Errorf("[%s] error deleting domain: %v", testName, err)
	}
}

func TestLowCostSeparateProfileObjectsIntoProfile(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	//creation of domain
	// Create required resources
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[%s] Error setting up aurora: %+v ", t.Name(), err)
	}
	dynamoTableName := randomTable("customer-profiles-low-cost-separate-profile")
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] error deleting dynamo table: %v", t.Name(), err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] error waiting for table creation: %v", t.Name(), err)
	}

	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)

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

	log.Printf("Creating domain for multiple object types test")
	domainName := randomDomain("multi_objects")
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[TestMultipleObjects] Error creating domain: %v", err)
	}

	// Create mapping
	airBookingType := "air_booking"
	testMappings := []FieldMapping{
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
	}
	err = lc.CreateMapping(airBookingType, "Mapping for "+airBookingType, testMappings)
	if err != nil {
		t.Errorf("[TestHistory] error creating mapping: %v", err)
	}

	clickstreamType := "clickstream"
	clickstreamMappings := []FieldMapping{
		// Profile Object Unique Key
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.session_id",
			Target:  "session_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.num_pax_children",
			Target:  "num_pax_children",
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
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
	}
	err = lc.CreateMapping(clickstreamType, "Mapping for "+clickstreamType, clickstreamMappings)
	if err != nil {
		t.Errorf("[TestHistory] error creating mapping: %v", err)
	}
	//data plane testing
	for i := 0; i < 10; i++ {
		profileOneTravelerID := "test_traveler_id" + strconv.Itoa(i)
		obj1, _ := json.Marshal(map[string]string{
			"traveler_id":    profileOneTravelerID,
			"accp_object_id": "test_accp_object_id" + strconv.Itoa(i),
			"segment_id":     "test_segment_id",
			"from":           "test_from" + strconv.Itoa(i),
			"arrival_date":   "test_arrival_date",
			"email":          "airBookingEmail" + strconv.Itoa(i) + "@example.com",
		})
		err = lc.PutProfileObject(string(obj1), airBookingType)
		if err != nil {
			t.Errorf("[TestHistory] error creating profile object 1: %v", err)
		}
	}

	for i := 0; i < 10; i++ {
		profileOneTravelerID := "test_traveler_id" + strconv.Itoa(i)
		obj1, _ := json.Marshal(map[string]string{
			"traveler_id":      profileOneTravelerID,
			"accp_object_id":   "test_accp_object_id" + strconv.Itoa(i),
			"session_id":       "test_session_id" + strconv.Itoa(i),
			"num_pax_children": "test_num_pax_children",
			"email":            "clickstreamEmail" + strconv.Itoa(i) + "@example.com",
		})
		err = lc.PutProfileObject(string(obj1), clickstreamType)
		if err != nil {
			t.Errorf("[TestHistory] error creating profile object 1: %v", err)
		}
	}

	//Place new clickstream objects into the same traveller id
	for i := 10; i < 20; i++ {
		profileOneTravelerID := "test_traveler_id" + strconv.Itoa(i-10)
		obj1, _ := json.Marshal(map[string]string{
			"traveler_id":      profileOneTravelerID,
			"accp_object_id":   "test_accp_object_id" + strconv.Itoa(i),
			"session_id":       "test_session_id" + strconv.Itoa(i),
			"num_pax_children": "test_num_pax_children",
			"email":            "clickstreamEmail" + strconv.Itoa(i-10) + "@example.com",
		})
		err = lc.PutProfileObject(string(obj1), clickstreamType)
		if err != nil {
			t.Errorf("[TestHistory] error creating profile object 1: %v", err)
		}
	}

	profile, err := lc.SearchProfiles("clickstream.num_pax_children", []string{"test_num_pax_children"})
	if err != nil {
		t.Errorf("[TestHistory] error searching profiles: %v", err)
	}
	log.Printf("Profile length: %v", len(profile))
	if len(profile) != 10 {
		t.Errorf("[TestHistory] wrong number of profiles returned: %v", err)
	}

	//Individual Object level search
	for i := 0; i < 10; i++ {
		profileOneTravelerID := "test_traveler_id" + strconv.Itoa(i)
		profile, err := lc.SearchProfiles("air_booking.traveler_id", []string{profileOneTravelerID})
		if err != nil {
			t.Errorf("[TestHistory] error searching profiles: %v", err)
		}
		if len(profile) != 1 {
			t.Errorf("[TestHistory] wrong number of profiles returned: %v", err)
		}

		profile, err = lc.SearchProfiles("profile_id", []string{profileOneTravelerID})
		if err != nil {
			t.Errorf("[TestHistory] error searching profiles: %v", err)
		}
		if len(profile) != 1 {
			t.Errorf("[TestHistory] wrong number of profiles returned: %v", err)
		}

		testFrom := "test_from" + strconv.Itoa(i)
		profile, err = lc.SearchProfiles("air_booking.from", []string{testFrom})
		if err != nil {
			t.Errorf("[TestHistory] error searching profiles: %v", err)
		}
		if len(profile) != 1 {
			t.Errorf("[TestHistory] wrong number of profiles returned: %v", err)
		}

		testEmail := "clickstreamEmail" + strconv.Itoa(i) + "@example.com"
		profile, err = lc.SearchProfiles("clickstream.email", []string{testEmail})
		if err != nil {
			t.Errorf("[TestHistory] error searching profiles: %v", err)
		}
		if len(profile) != 1 {
			t.Errorf("[TestHistory] wrong number of profiles returned: %v", err)
		}

		profile, err = lc.SearchProfiles("PersonalEmailAddress", []string{testEmail})
		if err != nil {
			t.Errorf("[TestHistory] error searching profiles: %v", err)
		}
		if len(profile) != 1 {
			t.Errorf("[TestHistory] wrong number of profiles returned: %v", err)
		}

		testSessionId := "test_session_id" + strconv.Itoa(i)
		profile, err = lc.SearchProfiles("clickstream.session_id", []string{testSessionId})
		if err != nil {
			t.Errorf("[TestHistory] error searching profiles: %v", err)
		}
		if len(profile) != 1 {
			t.Errorf("[TestHistory] wrong number of profiles returned: %v", err)
		}
		log.Printf("Profile Id: %v", profile[0].Attributes["traveler_id"])
		log.Printf("Profile Id: %v", profile[0].ProfileId)

		newTestSessionId := "test_session_id" + strconv.Itoa(i+10)
		profile, err = lc.SearchProfiles("clickstream.session_id", []string{newTestSessionId})
		if err != nil {
			t.Errorf("[TestHistory] error searching profiles: %v", err)
		}
		if len(profile) != 1 {
			t.Errorf("[TestHistory] wrong number of profiles returned: %v", err)
		}
	}

	err = lc.DeleteDomainByName(domainName)
	if err != nil {
		t.Errorf("[TestHistory] error deleting domain: %v", err)
	}
}

func TestLowCostTwoProfileObjectsSameObjectId(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	//creation of domain
	// Create required resources
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[customerprofiles] Error setting up aurora: %+v ", err)
	}
	dynamoTableName := randomTable("lcs-two-profile-objects-same-id")
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[TestInvalidDomainName] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] error deleting dynamo table: %v", t.Name(), err)
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

	log.Printf("Creating domain for multiple object types test")
	domainName := randomDomain("same_id_objects")
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[TestMultipleObjects] Error creating domain: %v", err)
		return
	}

	// Create mapping
	airBookingType := "air_booking"
	testMappings := []FieldMapping{
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
	}
	err = lc.CreateMapping(airBookingType, "Mapping for "+airBookingType, testMappings)
	if err != nil {
		t.Errorf("[TestHistory] error creating mapping: %v", err)
	}

	profileOneTravelerID := "test_traveler_id"
	obj1, _ := json.Marshal(map[string]string{
		"traveler_id":    profileOneTravelerID,
		"accp_object_id": "test_accp_object_id",
		"segment_id":     "test_segment_id",
		"from":           "test_from",
		"arrival_date":   "test_arrival_date",
		"email":          "airBookingEmail@example.com",
	})
	err = lc.PutProfileObject(string(obj1), airBookingType)
	if err != nil {
		t.Errorf("[TestHistory] error creating profile object 1: %v", err)
	}
	//Placing a second object with the same object id
	profileOneTravelerID = "test_traveler_id"
	obj2, _ := json.Marshal(map[string]string{
		"traveler_id":    profileOneTravelerID,
		"accp_object_id": "test_accp_object_id",
		"segment_id":     "test_segment_id",
		"from":           "new_from",
		"arrival_date":   "test_arrival_date",
		"email":          "airBookingEmailNew@example.com",
	})
	err = lc.PutProfileObject(string(obj2), airBookingType)
	if err != nil {
		t.Errorf("[TestHistory] error creating profile object 2: %v", err)
	}

	profile, err := lc.SearchProfiles("air_booking.from", []string{"new_from"})
	if err != nil {
		t.Errorf("[TestHistory] error searching profiles: %v", err)
	}
	log.Printf("Profile length: %v", len(profile))
	log.Printf("Profile: %v", profile[0].PersonalEmailAddress)
	log.Printf("Profile: %v", profile[0].ProfileObjects)

	profile, err = lc.SearchProfiles("air_booking.from", []string{"test_from"})
	if err != nil {
		t.Errorf("[TestHistory] error searching profiles: %v", err)
	}
	log.Printf("Profile length: %v", len(profile))
	// log.Printf("Profile: %v", profile[0].PersonalEmailAddress)
	// log.Printf("Profile: %v", profile[0].ProfileObjects)

	err = lc.DeleteDomainByName(domainName)
	if err != nil {
		t.Errorf("[TestHistory] error deleting domain: %v", err)
	}
}

func TestLowCostOverrideWithEmptyString(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[TestOverrideWithEmptyString] Error setting up aurora: %+v ", err)
	}
	dynamoTableName := randomTable("customer-profiles-low-cost-override-empty-string")
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[TestOverrideWithEmptyString] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] error deleting dynamo table: %v", t.Name(), err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[TestOverrideWithEmptyString] error waiting for table creation: %v", err)
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

	log.Printf("Creating domain")
	domainName := randomDomain("empty_string")

	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[TestOverrideWithEmptyString] Error creating domain: %v", err)
	}

	guestProfileType := "guest_profile"
	guestProfileMapping := []FieldMapping{
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.first_name",
			Target:  "_profile.FirstName",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.middle_name",
			Target:  "_profile.MiddleName",
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.last_name",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.created_on",
			Target:  "created_on",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.created_by",
			Target:  "created_by",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.traveler_id",
			Target:  "_profile.Attributes.traveler_id",
			Indexes: []string{"PROFILE"},
			KeyOnly: true,
		},
	}
	err = lc.CreateMapping(guestProfileType, "Mapping for "+guestProfileType, guestProfileMapping)
	if err != nil {
		t.Errorf("[TestOverrideWithEmptyString] error creating mapping: %v", err)
	}

	profileOneTravelerID := "test_traveler"
	obj1, _ := json.Marshal(map[string]string{
		"traveler_id":    profileOneTravelerID,
		"accp_object_id": "test_accp_object_id",
		"first_name":     "William",
		"middle_name":    "Jennings",
		"last_name":      "Bryan",
		"created_on":     "test_created_on",
		"created_by":     "test_created_by",
	})
	err = lc.PutProfileObject(string(obj1), guestProfileType)
	if err != nil {
		t.Errorf("[TestOverrideWithEmptyString] error creating profile object: %v", err)
	}

	obj2, _ := json.Marshal(map[string]string{
		"traveler_id":    profileOneTravelerID,
		"accp_object_id": "test_accp_object_id",
		"first_name":     "",
		"middle_name":    "Marks",
		"last_name":      "",
		"created_on":     "test_created_on",
		"created_by":     "test_created_by",
	})

	err = lc.PutProfileObject(string(obj2), guestProfileType)
	if err != nil {
		t.Errorf("[TestOverrideWithEmptyString] error creating profile object: %v", err)
	}

	profile, err := lc.SearchProfiles("FirstName", []string{"William"})
	if err != nil {
		t.Errorf("[TestOverrideWithEmptyString] error searching profiles: %v", err)
	}
	if len(profile) != 1 {
		t.Errorf("[TestOverrideWithEmptyString] wrong number of profiles: %v", len(profile))
	}
	profile, err = lc.SearchProfiles("MiddleName", []string{"Jennings"})
	if err != nil {
		t.Errorf("[TestOverrideWithEmptyString] error searching profiles: %v", err)
	}
	if len(profile) != 0 {
		t.Errorf("[TestOverrideWithEmptyString] wrong number of profiles: %v", len(profile))
	}
	profile, err = lc.SearchProfiles("MiddleName", []string{"Marks"})
	if err != nil {
		t.Errorf("[TestOverrideWithEmptyString] error searching profiles: %v", err)
	}
	if len(profile) != 1 {
		t.Errorf("[TestOverrideWithEmptyString] wrong number of profiles: %v", len(profile))
	}

	err = lc.DeleteDomainByName(domainName)
	if err != nil {
		t.Errorf("[TestOverrideWithEmptyString] error deleting domain: %v", err)
	}
}

func TestLowCostUptHotelBooking(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	// Aurora
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[TestUptHotelBooking] Error setting up aurora: %+v ", err)
	}
	// DynamoDB
	dynamoTableName := randomTable("customer-profiles-low-cost-upt-hotel-booking")
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[TestUptHotelBooking] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] error deleting dynamo table: %v", t.Name(), err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[TestUptHotelBooking] error waiting for table creation: %v", err)
	}
	// Kinesis
	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)

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

	domainName := randomDomain("hotel_booking")
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[TestUptHotelBooking] Error creating domain: %v", err)
	}

	objectTypeName := "upt_hotel_booking"
	mapping := []FieldMapping{
		// Profile Object Unique Key
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
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
			Source:  "_source.totalAfterTax",
			Target:  "totalAfterTax",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.email_type",
			Target:  "email_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_updated_by",
			Target:  "last_updated_by",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.n_nights",
			Target:  "n_nights",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.product_id",
			Target:  "product_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_updated",
			Target:  "last_updated",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.n_guests",
			Target:  "n_guests",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.room_type_description",
			Target:  "room_type_description",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.room_type_name",
			Target:  "room_type_name",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.totalBeforeTax",
			Target:  "totalBeforeTax",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.attribute_names",
			Target:  "attribute_names",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.attribute_codes",
			Target:  "attribute_codes",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.attribute_descriptions",
			Target:  "attribute_descriptions",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.room_type_code",
			Target:  "room_type_code",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.check_in_date",
			Target:  "check_in_date",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.object_type",
			Target:  "object_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.hotel_code",
			Target:  "hotel_code",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.status",
			Target:  "status",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.model_version",
			Target:  "model_version",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.address_type",
			Target:  "address_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.phone_type",
			Target:  "phone_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.creation_channel_id",
			Target:  "creation_channel_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_update_channel_id",
			Target:  "last_update_channel_id",
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
			Source:  "_source.tx_id",
			Target:  "tx_id",
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
			Source:  "_source.address_is_primary",
			Target:  "address_is_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.address_business_is_primary",
			Target:  "address_business_is_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.address_billing_is_primary",
			Target:  "address_billing_is_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.address_mailing_is_primary",
			Target:  "address_mailing_is_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.phone_primary",
			Target:  "phone_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.email_primary",
			Target:  "email_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.total_segment_before_tax",
			Target:  "total_segment_before_tax",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.total_segment_after_tax",
			Target:  "total_segment_after_tax",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.rate_plan_name",
			Target:  "rate_plan_name",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.rate_plan_code",
			Target:  "rate_plan_code",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.rate_plan_description",
			Target:  "rate_plan_description",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.add_on_names",
			Target:  "add_on_names",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.add_on_codes",
			Target:  "add_on_codes",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.add_on_descriptions",
			Target:  "add_on_descriptions",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_updated_partition",
			Target:  "last_updated_partition",
			KeyOnly: true,
		},
		// Profile Data
		{
			Type:        "STRING",
			Source:      "_source.traveller_id",
			Target:      "_profile.Attributes.profile_id",
			Searcheable: true,
			Indexes:     []string{"PROFILE"},
		},
		{
			Type:   "STRING",
			Source: "_source.company",
			Target: "_profile.BusinessName",
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.phone",
			Target:      "_profile.PhoneNumber",
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.honorific",
			Target: "_profile.Attributes.honorific",
		},
		{
			Type:        "STRING",
			Source:      "_source.first_name",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.middle_name",
			Target:      "_profile.MiddleName",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.last_name",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.gender",
			Target: "_profile.Gender",
		},
		{
			Type:   "STRING",
			Source: "_source.pronoun",
			Target: "_profile.Attributes.pronoun",
		},
		{
			Type:   "STRING",
			Source: "_source.date_of_birth",
			Target: "_profile.BirthDate",
		},
		{
			Type:   "STRING",
			Source: "_source.job_title",
			Target: "_profile.Attributes.job_title",
		},
		{
			Type:   "STRING",
			Source: "_source.nationality_code",
			Target: "_profile.Attributes.nationality_code",
		},
		{
			Type:   "STRING",
			Source: "_source.nationality_name",
			Target: "_profile.Attributes.nationality_name",
		},
		{
			Type:   "STRING",
			Source: "_source.language_code",
			Target: "_profile.Attributes.language_code",
		},
		{
			Type:   "STRING",
			Source: "_source.language_name",
			Target: "_profile.Attributes.language_name",
		},
		{
			Type:   "STRING",
			Source: "_source.pms_id",
			Target: "_profile.Attributes.pms_id",
		},
		{
			Type:   "STRING",
			Source: "_source.crs_id",
			Target: "_profile.Attributes.crs_id",
		},
		{
			Type:   "STRING",
			Source: "_source.gds_id",
			Target: "_profile.Attributes.gds_id",
		},
		{
			Type:   "STRING",
			Source: "_source.payment_type",
			Target: "_profile.Attributes.payment_type",
		},
		{
			Type:   "STRING",
			Source: "_source.cc_token",
			Target: "_profile.Attributes.cc_token",
		},
		{
			Type:   "STRING",
			Source: "_source.cc_type",
			Target: "_profile.Attributes.cc_type",
		},
		{
			Type:   "STRING",
			Source: "_source.cc_exp",
			Target: "_profile.Attributes.cc_exp",
		},
		{
			Type:   "STRING",
			Source: "_source.cc_cvv",
			Target: "_profile.Attributes.cc_cvv",
		},
		{
			Type:   "STRING",
			Source: "_source.cc_name",
			Target: "_profile.Attributes.cc_name",
		},
		// Home Address
		{
			Type:   "STRING",
			Source: "_source.address_line1",
			Target: "_profile.Address.Address1",
		},
		{
			Type:   "STRING",
			Source: "_source.address_line2",
			Target: "_profile.Address.Address2",
		},
		{
			Type:   "STRING",
			Source: "_source.address_line3",
			Target: "_profile.Address.Address3",
		},
		{
			Type:   "STRING",
			Source: "_source.address_line4",
			Target: "_profile.Address.Address4",
		},
		{
			Type:   "STRING",
			Source: "_source.address_city",
			Target: "_profile.Address.City",
		},
		{
			Type:   "STRING",
			Source: "_source.address_state",
			Target: "_profile.Address.State",
		},
		{
			Type:   "STRING",
			Source: "_source.address_province",
			Target: "_profile.Address.Province",
		},
		{
			Type:   "STRING",
			Source: "_source.address_postal_code",
			Target: "_profile.Address.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.address_country",
			Target: "_profile.Address.Country",
		},
		// Business Address
		{
			Type:   "STRING",
			Source: "_source.address_billing_line1",
			Target: "_profile.BillingAddress.Address1",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_line2",
			Target: "_profile.BillingAddress.Address2",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_line3",
			Target: "_profile.BillingAddress.Address3",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_line4",
			Target: "_profile.BillingAddress.Address4",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_city",
			Target: "_profile.BillingAddress.City",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_state",
			Target: "_profile.BillingAddress.State",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_province",
			Target: "_profile.BillingAddress.Province",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_postal_code",
			Target: "_profile.BillingAddress.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_country",
			Target: "_profile.BillingAddress.Country",
		},
		// Mailing Address
		{
			Type:   "STRING",
			Source: "_source.address_mailing_line1",
			Target: "_profile.MailingAddress.Address1",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_line2",
			Target: "_profile.MailingAddress.Address2",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_line3",
			Target: "_profile.MailingAddress.Address3",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_line4",
			Target: "_profile.MailingAddress.Address4",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_city",
			Target: "_profile.MailingAddress.City",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_state",
			Target: "_profile.MailingAddress.State",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_province",
			Target: "_profile.MailingAddress.Province",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_postal_code",
			Target: "_profile.MailingAddress.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_country",
			Target: "_profile.MailingAddress.Country",
		},
		// Business
		{
			Type:   "STRING",
			Source: "_source.address_business_line1",
			Target: "_profile.ShippingAddress.Address1",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line2",
			Target: "_profile.ShippingAddress.Address2",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line3",
			Target: "_profile.ShippingAddress.Address3",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line4",
			Target: "_profile.ShippingAddress.Address4",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_city",
			Target: "_profile.ShippingAddress.City",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_state",
			Target: "_profile.ShippingAddress.State",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_province",
			Target: "_profile.ShippingAddress.Province",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_postal_code",
			Target: "_profile.ShippingAddress.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_country",
			Target: "_profile.ShippingAddress.Country",
		},
		{
			Type:   "STRING",
			Source: "_source.email_business",
			Target: "_profile.BusinessEmailAddress",
		},
		{
			Type:   "STRING",
			Source: "_source.phone_home",
			Target: "_profile.HomePhoneNumber",
		},
		{
			Type:   "STRING",
			Source: "_source.phone_mobile",
			Target: "_profile.MobilePhoneNumber",
		},
		{
			Type:   "STRING",
			Source: "_source.phone_business",
			Target: "_profile.BusinessPhoneNumber",
		},
		{
			Type:   "STRING",
			Source: "_source.last_booking_id",
			Target: "_profile.Attributes.last_booking_id",
		},
	}

	err = lc.CreateMapping(objectTypeName, "", mapping)
	if err != nil {
		t.Errorf("[TestUptHotelBooking] CreateMapping failed: %s", err)
	}

	obj, err := json.Marshal(map[string]interface{}{
		"accp_object_id": "ROM05|MEA0GZX5SC|G0WBL6G1JL|2023-12-15|QAHSGEEHQD", "add_on_codes": "", "add_on_descriptions": "", "add_on_names": "", "address_billing_city": "", "address_billing_country": "", "address_billing_is_primary": false, "address_billing_line1": "", "address_billing_line2": "", "address_billing_line3": "", "address_billing_line4": "", "address_billing_postal_code": "", "address_billing_province": "", "address_billing_state": "", "address_business_city": "", "address_business_country": "", "address_business_is_primary": false, "address_business_line1": "", "address_business_line2": "", "address_business_line3": "", "address_business_line4": "", "address_business_postal_code": "", "address_business_province": "", "address_business_state": "", "address_city": "", "address_country": "", "address_is_primary": false, "address_line1": "", "address_line2": "", "address_line3": "", "address_line4": "", "address_mailing_city": "Indianapolis", "address_mailing_country": "IQ", "address_mailing_is_primary": false, "address_mailing_line1": "1518 Lake Mountland", "address_mailing_line2": "more content line 2", "address_mailing_line3": "more content line 3", "address_mailing_line4": "more content line 4", "address_mailing_postal_code": "17231", "address_mailing_province": "AB", "address_mailing_state": "", "address_postal_code": "", "address_province": "", "address_state": "", "address_type": "mailing", "attribute_codes": "MINI_BAR|BALCONY", "attribute_descriptions": "Mini bar with local snacks and beverages|Large balcony with chairs and a table", "attribute_names": "Mini bar|Balcony", "booker_id": "1354804801", "booking_id": "MEA0GZX5SC", "cc_cvv": "", "cc_exp": "", "cc_name": "", "cc_token": "", "cc_type": "", "check_in_date": "2023-12-15", "company": "Panjiva", "creation_channel_id": "gds", "crs_id": "", "date_of_birth": "1985-05-31", "email": "", "email_business": "", "email_primary": "", "first_name": "Flossie", "gds_id": "", "gender": "other", "honorific": "", "hotel_code": "ROM05", "job_title": "Executive", "language_code": "", "language_name": "", "last_booking_id": "MEA0GZX5SC", "last_name": "Wolf", "last_update_channel_id": "", "last_updated": "2024-04-09T09:39:49.341150Z", "last_updated_by": "Lucinda Lindgren", "last_updated_partition": "2024-04-09-09", "middle_name": "Noe", "model_version": "1.1.0", "n_guests": 1, "n_nights": 1, "nationality_code": "", "nationality_name": "", "object_type": "hotel_booking", "payment_type": "cash", "phone": "", "phone_business": "", "phone_home": "+1234567890", "phone_mobile": "", "phone_primary": "", "pms_id": "", "product_id": "QAHSGEEHQD", "pronoun": "he", "rate_plan_code": "XMAS_WEEK", "rate_plan_description": "Special rate for the chrismas season", "rate_plan_name": "Xmas special rate", "room_type_code": "DBL", "room_type_description": "Room with Double bed", "room_type_name": "Double room", "segment_id": "G0WBL6G1JL", "status": "confirmed", "total_segment_after_tax": 289.18982, "total_segment_before_tax": 251.4694, "traveller_id": "1354804801", "tx_id": "34cc8864-5484-4e6f-bca7-711b8856ee5f",
	})
	if err != nil {
		t.Errorf("[TestUptHotelBooking] Marshaling failed: %s", err)
	}

	err = lc.PutProfileObject(string(obj), objectTypeName)
	if err != nil {
		t.Errorf("[TestUptHotelBooking] PutProfileObject failed: %s", err)
	}
	profiles, err := lc.SearchProfiles("profile_id", []string{"1354804801"})
	if err != nil {
		t.Errorf("[TestUptHotelBooking] SearchProfiles failed: %s", err)
	}
	if len(profiles) != 1 {
		t.Errorf("[TestUptHotelBooking] SearchProfiles returned %d profiles, expected 1", len(profiles))
	} else {
		profile, err := lc.GetProfile(profiles[0].ProfileId, []string{objectTypeName}, []PaginationOptions{})
		if err != nil {
			t.Errorf("[TestUptHotelBooking] GetProfile failed: %s", err)
		}
		if profile.HomePhoneNumber != "+1234567890" {
			t.Errorf("[TestUptHotelBooking] HomePhoneNumber is not correct: %s", profile.HomePhoneNumber)
		}
		if profile.BusinessName != "Panjiva" {
			t.Errorf("[TestUptHotelBooking] BusinessName is not correct: %s", profile.BusinessName)
		}
	}

	// Cleanup
	err = lc.DeleteDomain()
	if err != nil {
		t.Errorf("[TestUptHotelBooking] DeleteDomain failed: %s", err)
	}
}

func TestLowCostMergeManyLC(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	// Set up resources
	// Aurora
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[TestMergeManyLC] Error setting up aurora: %+v ", err)
	}
	// DynamoDB
	dynamoTableName := randomTable("lcs-merge-many")
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[TestMergeManyLC] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[TestMergeManyLC] DeleteTable failed: %s", err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[TestMergeManyLC] error waiting for table creation: %v", err)
	}
	// Kinesis
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

	// Create domain
	domainName := randomDomain("lcs_merge_many")
	envNameTest := "testenv"
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: envNameTest}, "", "", DomainOptions{})
	if err != nil {
		t.Fatalf("[TestMergeManyLC] error creating domain: %v", err)
	}
	t.Cleanup(func() {
		err = lc.DeleteDomain()
		if err != nil {
			t.Errorf("[TestMergeManyLC] DeleteDomain failed: %s", err)
		}
	})

	// Create mapping
	err = lc.CreateMapping("air_booking", "air booking mapping", []FieldMapping{
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
			Source:  "_source.test_field",
			Target:  "test_field",
			KeyOnly: true,
		},
		{
			Type:        MappingTypeString,
			Source:      "_source." + LAST_UPDATED_FIELD,
			Target:      "LAST_UPDATED_FIELD",
			Searcheable: true,
		},
	})
	if err != nil {
		t.Errorf("[TestMergeManyLC] error creating mapping: %v", err)
	}
	nProfilesToTest := 30
	profileObjects := make([]string, nProfilesToTest)
	for i := 0; i < nProfilesToTest; i++ {
		obj, _ := json.Marshal(map[string]interface{}{
			"traveler_id":    fmt.Sprintf("id_%d", i),
			"accp_object_id": fmt.Sprintf("obj_%d", i),
			"test_field":     "test_val",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		})
		profileObjects[i] = string(obj)
	}

	upIDs := []string{}
	for i, obj := range profileObjects {
		err = lc.PutProfileObject(obj, "air_booking")
		if err != nil {
			t.Errorf("[TestMergeManyLC] error putting profile object: %v", err)
		}
		upID, err := lc.GetUptID(fmt.Sprintf("id_%d", i))
		if err != nil {
			t.Errorf("[TestMergeManyLC] error getting upt id: %v", err)
		}
		upIDs = append(upIDs, upID)
		profile, err := lc.GetProfile(upID, []string{"air_booking"}, []PaginationOptions{})
		if err != nil {
			t.Errorf("[TestMergeManyLC][%d]  GetProfile failed: %s", i, err)
		}
		if len(profile.ProfileObjects) != 1 {
			t.Fatalf("[TestMergeManyLC][%d]  profile %+v should have 1 objects and not %d", i, profile, len(profile.ProfileObjects))
		}
		if profile.ProfileObjects[0].ID != fmt.Sprintf("obj_%d", i) {
			t.Fatalf("[TestMergeManyLC][%d]  profile %+v should have object with id obj_%d", i, profile, i)
		}
	}

	mergePairs := make([]ProfilePair, len(profileObjects)/2)
	for i := 0; i < len(profileObjects)/2; i++ {
		mergePairs[i] = ProfilePair{
			SourceID: upIDs[i],
			TargetID: upIDs[i+len(profileObjects)/2],
		}
	}
	_, err = lc.MergeMany(mergePairs)
	if err != nil {
		t.Errorf("[TestMergeManyLC] MergeMany failed: %s", err)
	}

	for i := 0; i < len(profileObjects)/2; i++ {
		profile, err := lc.GetProfile(upIDs[i], []string{"air_booking"}, []PaginationOptions{})
		if err != nil {
			t.Errorf("[TestMergeManyLC][%d]  GetProfile failed: %s", i, err)
		}
		if len(profile.ProfileObjects) != 2 {
			t.Fatalf("[TestMergeManyLC][%d]  profile %+v should have 2 objects and not %d", i, profile, len(profile.ProfileObjects))
		}
		if profile.ProfileObjects[1].ID != fmt.Sprintf("obj_%d", i) {
			t.Fatalf("[TestMergeManyLC][%d]  profile %+v should have object with id obj_%d", i, profile, i)
		}
		if profile.ProfileObjects[0].ID != fmt.Sprintf("obj_%d", i+len(profileObjects)/2) {
			t.Fatalf("[TestMergeManyLC][%d] profile %+v should have object with id obj_%d", i, profile, i+len(profileObjects)/2)
		}
	}
}

func TestLowCostUpdateProfile(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	// Set up resources
	// Aurora
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[TestUpdateProfile] Error setting up aurora: %+v ", err)
	}
	// DynamoDB
	dynamoTableName := randomTable("lcs-update-profile")
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[TestUpdateProfile] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[TestUpdateProfile] DeleteTable failed: %s", err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[TestUpdateProfile] error waiting for table creation: %v", err)
	}
	// Kinesis
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

	// Create domain
	domainName := randomDomain("update_prof")
	envNameTest := "testenv"
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: envNameTest}, "", "", DomainOptions{})
	if err != nil {
		t.Fatalf("[TestUpdateProfile] error creating domain: %v", err)
	}
	t.Cleanup(func() {
		err = lc.DeleteDomain()
		if err != nil {
			t.Errorf("[TestUpdateProfile] DeleteDomain failed: %s", err)
		}
	})

	// Create mapping
	err = lc.CreateMapping("air_booking", "air booking mapping", []FieldMapping{
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
			Source:  "_source.loyalty",
			Target:  "loyalty",
			KeyOnly: true,
		},
	})
	if err != nil {
		t.Fatalf("[TestUpdateProfile] error creating mapping: %v", err)
	}

	originalObject, _ := json.Marshal(map[string]interface{}{
		"traveler_id":    "trav_1",
		"accp_object_id": "obj_1",
		"loyalty":        "gold",
	})

	err = lc.PutProfileObject(string(originalObject), "air_booking")
	if err != nil {
		t.Fatalf("[TestUpdateProfile] error putting profile object: %v", err)
	}
	profile, err := lc.GetProfileByProfileId("trav_1", []string{"air_booking"}, []PaginationOptions{})
	if err != nil {
		t.Fatalf("[TestUpdateProfile] error retreiving profile by profile id %v", err)
	}
	if len(profile.ProfileObjects) != 1 {
		t.Fatalf("[TestUpdateProfile] ProfileObjects length is %d, expected 1", len(profile.ProfileObjects))
	}
	if profile.ProfileObjects[0].Attributes["loyalty"] != "gold" {
		t.Errorf("[TestUpdateProfile] Expected profile attribute %s, got %s", "gold", profile.Attributes["loyalty"])
	}

	updatedObject, _ := json.Marshal(map[string]interface{}{
		"traveler_id":    "trav_1",
		"accp_object_id": "obj_1",
		"loyalty":        "platinum",
	})
	err = lc.PutProfileObject(string(updatedObject), "air_booking")
	if err != nil {
		t.Fatalf("[TestUpdateProfile] error putting profile object: %v", err)
	}
	profile, err = lc.GetProfileByProfileId("trav_1", []string{"air_booking"}, []PaginationOptions{})
	if err != nil {
		t.Fatalf("[TestUpdateProfile] error retreiving profile by profile id %v", err)
	}
	if profile.ProfileObjects[0].AttributesInterface["loyalty"] != "platinum" {
		t.Errorf(
			"[TestUpdateProfile] Expected profile attribute %s, got %s",
			"platinum",
			profile.ProfileObjects[0].AttributesInterface["loyalty"],
		)
	}

}

func TestLowCostCreateProfileTableAndRetrieveColumnNames(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[TestRetrieveColumnNames] Error setting up aurora: %+v ", err)
	}
	// DynamoDB
	dynamoTableName := randomTable("customer-profiles-low-cost-upt-column-names")
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[TestRetrieveColumnNames] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] error deleting dynamo table: %v", t.Name(), err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[TestRetrieveColumnNames] error waiting for table creation: %v", err)
	}

	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)

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

	domainName := randomDomain("lcs_get_columns")
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[TestRetrieveColumnNames] Error creating domain: %v", err)
	}

	objectTypeName := "upt_addresses_attributes"
	mapping := []FieldMapping{
		// Profile Object Unique Key
		{
			Type:   "STRING",
			Source: "_source.last_booking_id",
			Target: "_profile.Attributes.last_booking_id",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_state",
			Target: "_profile.MailingAddress.State",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_province",
			Target: "_profile.MailingAddress.Province",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_postal_code",
			Target: "_profile.MailingAddress.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_country",
			Target: "_profile.MailingAddress.Country",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line1",
			Target: "_profile.ShippingAddress.Address1",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line2",
			Target: "_profile.ShippingAddress.Address2",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line3",
			Target: "_profile.ShippingAddress.Address3",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line4",
			Target: "_profile.ShippingAddress.Address4",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_city",
			Target: "_profile.ShippingAddress.City",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_state",
			Target: "_profile.ShippingAddress.State",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_province",
			Target: "_profile.ShippingAddress.Province",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_postal_code",
			Target: "_profile.ShippingAddress.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_country",
			Target: "_profile.ShippingAddress.Country",
		},
		// Profile Data
		{
			Type:        "STRING",
			Source:      "_source.traveller_id",
			Target:      "_profile.Attributes.profile_id",
			Searcheable: true,
			Indexes:     []string{"PROFILE"},
		},
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
	}

	err = lc.CreateMapping(objectTypeName, domainName, mapping)
	if err != nil {
		t.Errorf("[TestRetrieveColumnNames] CreateMapping failed: %s", err)
	}

	resp, err := lc.Data.GetProfileMasterColumns(domainName)
	if err != nil {
		t.Errorf("Error")
	}

	//checking if contains addresses
	fieldFound := false
	for _, column := range resp {
		if column == "ShippingAddress.Country" {
			fieldFound = true
		}
	}
	if !fieldFound {
		t.Errorf("Error: ShippingAddress.Country not found")
	}
	// Cleanup
	err = lc.DeleteDomain()
	if err != nil {
		t.Errorf("[TestRetrieveColumnNames] DeleteDomain failed: %s", err)
	}
}

func TestCreateDomainWithWrongOptions(t *testing.T) {
	lc := CustomerProfileConfigLowCost{}
	dynMode := false
	cpMode := false
	err := lc.CreateDomainWithQueue("valid_domain_name", "test_kms_arn", map[string]string{
		"envName": "dev",
	}, "test_queue", "test_match_bucket", DomainOptions{
		DynamoMode:          &dynMode,
		CustomerProfileMode: &cpMode,
	})
	if err == nil {
		t.Fatalf("Create domain with both cache option set to false should return an error")
	}
	expectedErr := "customer profile mode and dynamo mode cannot be both false, at least one downstream cache option must be chosen"
	if err.Error() != expectedErr {
		t.Fatalf("Create domain with both cache option set to false should return %s and not %s", expectedErr, err.Error())

	}
}

// this test is used to for test coverage. can't validate since no return value
func TestCleanupDomain(t *testing.T) {
	lc := CustomerProfileConfigLowCost{
		Config: ConfigHandler{Svc: &db.DBConfig{}, cache: InitLcsCache()},
	}
	lc.cleanupDomainData("invalid_domain")
}

// this test is used to for test coverage.
func TestActivateSaveListIdResRuleSet(t *testing.T) {
	lc := CustomerProfileConfigLowCost{
		Config:       ConfigHandler{Svc: &db.DBConfig{}, cache: InitLcsCache()},
		DomainConfig: LcDomainConfig{Options: DomainOptions{RuleBaseIdResolutionOn: false}},
	}
	err := lc.ActivateIdResRuleSet()
	if err == nil {
		t.Fatalf("Should get an error when domain is not set")
	}
	err = lc.SaveIdResRuleSet([]Rule{})
	if err == nil {
		t.Fatalf("Should get an error when domain is not set")
	}
	_, err = lc.ListIdResRuleSets(false)
	if err == nil {
		t.Fatalf("Should get an error when domain is not set")
	}
	lc.DomainName = "test_domain"
	err = lc.ActivateIdResRuleSet()
	if err == nil {
		t.Fatalf("Should get an error when rule based ID resolution is off")
	}
	err = lc.SaveIdResRuleSet([]Rule{})
	if err == nil {
		t.Fatalf("Should get an error when rule based ID resolution is off")
	}
	_, err = lc.ListIdResRuleSets(false)
	if err == nil {
		t.Fatalf("Should get an error when rule based ID resolution is off")
	}
}

func TestInOrderDynamoDBIngestion(t *testing.T) {
	t.Parallel()
	testName := t.Name()[:15]
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("[%s] error loading env config: %v", testName, err)
	}
	// Set up resources for LCS
	// Aurora
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[%s] Error setting up aurora: %+v ", testName, err)
	}

	//	Set up Dynamo
	dynamoTableName := randomTable(testName)
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", testName, err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[%s] DeleteTable failed: %s", testName, err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] error waiting for table creation: %v", testName, err)
	}

	//	Setup kinesis, merge SQS, CP writer SQS
	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)
	mergeQueue := buildTestMergeQueue(t, envCfg.Region)
	cpWriterQueue := buildTestCPWriterQueue(t, envCfg.Region)
	initOptions := CustomerProfileInitOptions{MergeQueueClient: mergeQueue}

	lc := InitLowCost(envCfg.Region,
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

	//	Create Domain
	domainName := randomDomain(testName)
	envNameTest := "testenv"
	truePtr := true
	falsePtr := false
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: envNameTest}, "", "", DomainOptions{DynamoMode: &truePtr, CustomerProfileMode: &falsePtr})
	if err != nil {
		t.Fatalf("[%s] error creating domain: %v", testName, err)
	}
	t.Cleanup(func() {
		err = lc.DeleteDomain()
		if err != nil {
			t.Errorf("[%s] DeleteDomain failed: %s", testName, err)
		}
	})

	//	Create Dynamo CFG Object to directly interface with cache table
	cache := db.Init("ucp_domain_"+domainName, "connect_id", "record_type", "", "")

	//	Setup Mappings
	airBookingObjectTypeName := "air_booking"
	testAirBookingMappings := []FieldMapping{
		{
			Type:    MappingTypeString,
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{INDEX_UNIQUE},
			KeyOnly: true,
		},
		{
			Type:    MappingTypeString,
			Source:  "_source.traveler_id",
			Target:  "_profile.Attributes.traveler_id",
			Indexes: []string{INDEX_PROFILE},
			KeyOnly: true,
		},
		{
			Type:    MappingTypeString,
			Source:  "_source.from",
			Target:  "from",
			KeyOnly: true,
		},
		{
			Type:    MappingTypeString,
			Source:  "_source.arrival_date",
			Target:  "arrival_date",
			KeyOnly: true,
		},
		{
			Type:        MappingTypeString,
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        MappingTypeString,
			Source:      "_source." + LAST_UPDATED_FIELD,
			Target:      LAST_UPDATED_FIELD,
			Searcheable: true,
		},
	}
	hotelBookingObjectTypeName := "hotel_booking"
	testHotelBookingMappings := []FieldMapping{
		{
			Type:    MappingTypeString,
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{INDEX_UNIQUE},
			KeyOnly: true,
		},
		{
			Type:    MappingTypeString,
			Source:  "_source.traveler_id",
			Target:  "_profile.Attributes.traveler_id",
			Indexes: []string{INDEX_PROFILE},
			KeyOnly: true,
		},
		{
			Type:    MappingTypeString,
			Source:  "_source.hotel_name",
			Target:  "hotel_name",
			KeyOnly: true,
		},
		{
			Type:    MappingTypeString,
			Source:  "_source.check_in_date",
			Target:  "check_in_date",
			KeyOnly: true,
		},
		{
			Type:        MappingTypeString,
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        MappingTypeString,
			Source:      "_source." + LAST_UPDATED_FIELD,
			Target:      LAST_UPDATED_FIELD,
			Searcheable: true,
		},
	}
	err = lc.CreateMapping(airBookingObjectTypeName, "Mapping for "+airBookingObjectTypeName, testAirBookingMappings)
	if err != nil {
		t.Fatalf("[%s] error creating mapping: %v", testName, err)
	}
	err = lc.CreateMapping(hotelBookingObjectTypeName, "Mapping for "+hotelBookingObjectTypeName, testHotelBookingMappings)
	if err != nil {
		t.Fatalf("[%s] error creating mapping: %v", testName, err)
	}
	_, err = lc.GetMappings()
	if err != nil {
		t.Fatalf("[%s] error retrieving mappings: %v", testName, err)
	}

	// Setup Test Data
	type ProfileIngest []struct {
		objectType string
		data       map[string]string
	}
	// utcNow := time.Now().UTC()
	var inOrderIngestionTestData = []struct {
		name              string
		skipTest          bool
		profilesToIngest  ProfileIngest
		expectedCacheData DynamoProfileRecord
	}{
		{
			name:     "Basic",
			skipTest: false,
			profilesToIngest: ProfileIngest{
				{
					objectType: airBookingObjectTypeName,
					data: map[string]string{

						"traveler_id":    "basicTrav1",
						"accp_object_id": randomTable("basicTrav1"),
						"from":           "ORD",
						"arrival_date":   "05/07/2000",
						"email":          "johnny@example.com",
						"last_updated":   utcNow.AddDate(0, 0, -1).Format(INGEST_TIMESTAMP_FORMAT),
					},
				},
			},
			expectedCacheData: DynamoProfileRecord{
				PersonalEmailAddress: "johnny@example.com",
			},
		},
		{
			name:     "Sequential Ingestion",
			skipTest: false,
			profilesToIngest: ProfileIngest{
				{
					objectType: airBookingObjectTypeName,
					data: map[string]string{
						"traveler_id":    "seqTrav1",
						"accp_object_id": randomTable("seqTrav1"),
						"from":           "IAD",
						"arrival_date":   "05/06/2008",
						"email":          "seqTrav1@example.com",
						"last_updated":   utcNow.AddDate(-1, 0, 0).Format(INGEST_TIMESTAMP_FORMAT),
					},
				},
				{
					objectType: airBookingObjectTypeName,
					data: map[string]string{
						"traveler_id":    "seqTrav1",
						"accp_object_id": randomTable("seqTrav1"),
						"from":           "ORD",
						"arrival_date":   "01/12/2018",
						"email":          "billy@example.com",
						"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
					},
				},
			},
			expectedCacheData: DynamoProfileRecord{
				PersonalEmailAddress: "billy@example.com",
			},
		},
		{
			name:     "Out of order ingestion",
			skipTest: false,
			profilesToIngest: ProfileIngest{
				{
					objectType: airBookingObjectTypeName,
					data: map[string]string{
						"traveler_id":    "ooiTrav1",
						"accp_object_id": randomTable("ooiTrav1"),
						"from":           "ORD",
						"arrival_date":   "01/12/2018",
						"email":          "donnie@example.com",
						"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
					},
				},
				{
					objectType: airBookingObjectTypeName,
					data: map[string]string{
						"traveler_id":    "ooiTrav1",
						"accp_object_id": randomTable("ooiTrav1"),
						"from":           "IAD",
						"arrival_date":   "05/06/2008",
						"email":          "ooiTrav1@example.com",
						"last_updated":   utcNow.AddDate(-1, 0, 0).Format(INGEST_TIMESTAMP_FORMAT),
					},
				},
			},
			expectedCacheData: DynamoProfileRecord{
				PersonalEmailAddress: "donnie@example.com",
			},
		},
		{
			name:     "Out of Order ingestion; multiple obj types",
			skipTest: false,
			profilesToIngest: ProfileIngest{
				{
					objectType: airBookingObjectTypeName,
					data: map[string]string{
						"traveler_id":    "ooimotTrav1",
						"accp_object_id": randomTable("ooimotTrav1"),
						"from":           "IAD",
						"arrival_date":   "05/06/2008",
						"email":          "ooimotTrav1@example.com",
						"last_updated":   utcNow.AddDate(-1, 0, 0).Format(INGEST_TIMESTAMP_FORMAT)},
				},
				{
					objectType: airBookingObjectTypeName,
					data: map[string]string{
						"traveler_id":    "ooimotTrav1",
						"accp_object_id": randomTable("ooimotTrav1"),
						"from":           "ORD",
						"arrival_date":   "01/12/2018",
						"email":          "jd@example.com",
						"last_updated":   utcNow.AddDate(0, 0, -1).Format(INGEST_TIMESTAMP_FORMAT),
					},
				},
				{
					objectType: hotelBookingObjectTypeName,
					data: map[string]string{
						"traveler_id":    "ooimotTrav1",
						"accp_object_id": randomTable("ooimotTrav1"),
						"hotel_name":     "Marriott",
						"check_in_date":  "05/06/2008",
						"email":          "jd56@example.com",
						"last_updated":   utcNow.AddDate(0, 0, -2).Format(INGEST_TIMESTAMP_FORMAT),
					},
				},
			},
			expectedCacheData: DynamoProfileRecord{
				PersonalEmailAddress: "jd@example.com",
			},
		},
	}

	for _, testData := range inOrderIngestionTestData {
		if testData.skipTest {
			continue
		}
		t.Run(testData.name, func(t *testing.T) {
			var connectId string
			for _, profile := range testData.profilesToIngest {
				obj, err := json.Marshal(profile.data)
				if err != nil {
					t.Fatalf("[%s] [%s] error marshaling profile: %v", testName, testData.name, err)
				}
				err = lc.PutProfileObject(string(obj), profile.objectType)
				if err != nil {
					t.Fatalf("[%s] [%s] error creating profile object: %v", testName, testData.name, err)
				}
				connectId, err = WaitForProfileCreationByProfileId(profile.data["traveler_id"], RESERVED_FIELD_PROFILE_ID, lc)
				if err != nil {
					t.Fatalf("[%s] [%s] error waiting for profile from aurora: %v", testName, testData.name, err)
				}
			}

			var updatedCachedProfile DynamoProfileRecord
			err = cache.Get(connectId, "profile_main", &updatedCachedProfile)
			if err != nil {
				t.Fatalf("[%s] [%s] error retrieving profiles from dynamo cache: %v", testName, testData.name, err)
			}

			if updatedCachedProfile.PersonalEmailAddress != testData.expectedCacheData.PersonalEmailAddress {
				t.Fatalf("[%s] [%s] cache email mismatch; expected: '%s', got '%s'", testName, testData.name, testData.expectedCacheData.PersonalEmailAddress, updatedCachedProfile.PersonalEmailAddress)
			}
			// cacheExpiresAt := updatedCachedProfile.ExpiresAt.Format(TIMESTAMP_FORMAT)
			// expectedExpiresAt := testData.expectedCacheData.ExpiresAt.Format(TIMESTAMP_FORMAT)
			// if cacheExpiresAt != expectedExpiresAt {
			// 	t.Fatalf("[%s] [%s] cache expiry mismatch; expected: '%s', got '%s'", testName, testData.name, expectedExpiresAt, cacheExpiresAt)
			// }
		})
	}
}
