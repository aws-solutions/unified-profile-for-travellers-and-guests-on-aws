// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"encoding/json"
	"log"
	"strconv"
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

func TestSearchMultiKeySearch1(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[customerprofiles] Error setting up aurora: %+v ", err)
		return
	}
	dynamoTableName := "customer-profiles-low-cost-multi-key-test-" + core.GenerateUniqueId()
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[TestMultipleSearch] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[TestMultipleSearch] error deleting table: %v", err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[TestMultipleSearch] error waiting for table creation: %v", err)
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
	domainName := randomDomain("multi_key")

	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[TestMultipleSearch] Error creating domain: %v", err)
	}
	t.Cleanup(func() {
		err = lc.DeleteDomainByName(domainName)
		if err != nil {
			t.Errorf("[TestMultipleSearch] error deleting domain: %v", err)
		}
	})

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
		{
			Type:        "STRING",
			Source:      "_source.first_name",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
	}
	err = lc.CreateMapping(airBookingType, "Mapping for "+airBookingType, testMappings)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error creating mapping: %v", err)
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
		{
			Type:        "STRING",
			Source:      "_source.last_name",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
	}
	err = lc.CreateMapping(clickstreamType, "Mapping for "+clickstreamType, clickstreamMappings)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error creating mapping: %v", err)
	}

	profileOneTravelerID := "test_traveler_id"
	obj1, _ := json.Marshal(map[string]string{
		"traveler_id":    profileOneTravelerID,
		"accp_object_id": "test_accp_object_id",
		"segment_id":     "test_segment_id",
		"from":           "test_from",
		"arrival_date":   "test_arrival_date",
		"email":          "airBookingEmail@example.com",
		"first_name":     "Michael",
	})
	err = lc.PutProfileObject(string(obj1), airBookingType)
	if err != nil {
		t.Errorf("[TestHistory] error creating profile object 1: %v", err)
	}

	obj2, _ := json.Marshal(map[string]string{
		"traveler_id":      profileOneTravelerID,
		"accp_object_id":   "test_accp_object_id",
		"session_id":       "test_session_id",
		"num_pax_children": "test_num_pax_children",
		"email":            "clickstreamEmail@example.com",
		"last_name":        "Smith",
	})
	err = lc.PutProfileObject(string(obj2), clickstreamType)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error creating profile object 1: %v", err)
	}
	profile, err := lc.SearchProfiles("FirstName", []string{"Michael"})
	log.Printf("profile: %+v", profile)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	if len(profile) != 1 {
		t.Errorf("[TestMultipleSearch] error searching profiles: profile should return %d results and not %d", 1, len(profile))
	} else {
		if profile[0].FirstName != "Michael" {
			t.Errorf("[TestMultipleSearch] error searching profiles. FirstName should be %s and not %s", "Michael", profile[0].FirstName)
		}
		if profile[0].LastName != "Smith" {
			t.Errorf("[TestMultipleSearch] error searching profiles. last name should be %s and not %s", "Smith", profile[0].LastName)
		}
	}

	err = lc.PutProfileObject(string(obj2), clickstreamType)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error creating profile object 1: %v", err)
	}
	profile, err = lc.SearchProfiles("LastName", []string{"Smith"})
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	if len(profile) != 1 {
		t.Errorf("[TestMultipleSearch] error searching profiles: profile should return %d results and not %d", 1, len(profile))
	} else {
		if profile[0].FirstName != "Michael" {
			t.Errorf("[TestMultipleSearch] error searching profiles. FirstName should be %s and not %s", "Michael", profile[0].FirstName)
		}
		if profile[0].LastName != "Smith" {
			t.Errorf("[TestMultipleSearch] error searching profiles. last name should be %s and not %s", "Smith", profile[0].LastName)
		}
	}

	//Now, send in multiple of Michaels, and a guy not named Michael
	profileTwoTravelerID := "test_traveler_John"
	obj3, _ := json.Marshal(map[string]string{
		"traveler_id":    profileTwoTravelerID,
		"accp_object_id": "test_accp_object_id",
		"segment_id":     "test_segment_id",
		"from":           "test_from",
		"arrival_date":   "test_arrival_date",
		"email":          "XXXXXXXXXXXXXXXXXXXXXXXXX",
		"first_name":     "John",
	})
	err = lc.PutProfileObject(string(obj3), airBookingType)
	if err != nil {
		t.Errorf("[TestHistory] error creating profile object 1: %v", err)
	}
	for i := 0; i < 10; i++ {
		obj1, _ := json.Marshal(map[string]string{
			"traveler_id":    profileOneTravelerID + strconv.Itoa(i),
			"accp_object_id": "test_accp_object_id",
			"segment_id":     "test_segment_id",
			"from":           "test_from" + strconv.Itoa(i),
			"arrival_date":   "test_arrival_date",
			"email":          "XXXXXXXXXXXXXXXXXXXXXXXXX",
			"first_name":     "Michael",
		})
		err = lc.PutProfileObject(string(obj1), airBookingType)
		if err != nil {
			t.Errorf("[TestHistory] error creating profile object 1: %v", err)
		}
	}
	profile, err = lc.SearchProfiles("FirstName", []string{"Michael"})
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	if len(profile) != 11 {
		t.Errorf("[TestMultipleSearch] error searching profiles: profile should return %d results and not %d", 11, len(profile))
	}

	//PrepareMultiSearch
	firstNameValues := []string{"Michael"}
	fromValues := []string{"test_from"}

	multiValueSearch := make(map[string][]string)
	multiValueSearch["FirstName"] = firstNameValues
	multiValueSearch["air_booking.from"] = fromValues

	cids, err := lc.SearchOnMultipleKeysProfile(multiValueSearch)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	profile, err = lc.Data.FetchProfiles(domainName, cids)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	if len(profile) != 1 {
		t.Errorf("[TestMultipleSearch] error searching profiles: profile should return %d results and not %d", 1, len(profile))
	} else {
		if profile[0].LastName != "Smith" {
			t.Errorf("[TestMultipleSearch] error searching profiles. last name should be %s and not %s", "Smith", profile[0].LastName)
		}
	}

	//adding another Michael with test_from
	profileMichaelTestFrom := "test_traveler_Michael"
	obj4, _ := json.Marshal(map[string]string{
		"traveler_id":    profileMichaelTestFrom,
		"accp_object_id": "test_accp_object_id",
		"segment_id":     "test_segment_id",
		"from":           "test_from",
		"arrival_date":   "new_test_arrival_date",
		"email":          "XXXXXXXXXXXXXXXXXXXXXXXXX",
		"first_name":     "Michael",
	})
	err = lc.PutProfileObject(string(obj4), airBookingType)
	if err != nil {
		t.Errorf("[TestHistory] error creating profile object 1: %v", err)
	}
	//Now there are two Michaels with test_from
	cids, err = lc.SearchOnMultipleKeysProfile(multiValueSearch)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	profile, err = lc.Data.FetchProfiles(domainName, cids)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	if len(profile) != 2 {
		t.Errorf("[TestMultipleSearch] error searching profiles: profile should return %d results and not %d", 2, len(profile))
	}

	delete(multiValueSearch, "FirstName")
	//now try with email
	multiValueSearch["clickstream.session_id"] = []string{"test_session_id"}
	cids, err = lc.SearchOnMultipleKeysProfile(multiValueSearch)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	profile, err = lc.Data.FetchProfiles(domainName, cids)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	if len(profile) != 1 {
		t.Errorf("[TestMultipleSearch] error searching profiles: profile should return %d results and not %d", 1, len(profile))
	} else {
		if profile[0].LastName != "Smith" {
			t.Errorf("[TestMultipleSearch] error searching profiles. last name should be %s and not %s", "Smith", profile[0].LastName)
		}
	}

	//now add clickstream to John
	obj5, _ := json.Marshal(map[string]string{
		"traveler_id":      profileTwoTravelerID,
		"accp_object_id":   "test_accp_object_id",
		"session_id":       "test_session_id",
		"num_pax_children": "test_num_pax_children",
		"email":            "clickstreamEmail@example.com",
		"last_name":        "Smith",
	})
	err = lc.PutProfileObject(string(obj5), clickstreamType)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error creating profile object 1: %v", err)
	}

	//the previous search no uniquely determines a value
	cids, err = lc.SearchOnMultipleKeysProfile(multiValueSearch)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	profile, err = lc.Data.FetchProfiles(domainName, cids)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	if len(profile) != 2 {
		t.Errorf("[TestMultipleSearch] wrong number of profiles: %v", len(profile))
	}

	//adding back first name to get unique determination
	multiValueSearch["FirstName"] = firstNameValues
	cids, err = lc.SearchOnMultipleKeysProfile(multiValueSearch)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	profile, err = lc.Data.FetchProfiles(domainName, cids)
	if len(profile) != 1 {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	} else {
		if profile[0].LastName != "Smith" {
			t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
		}
	}
}

func TestSearchMultiKeySearchSameTable(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[TestMultipleSearchSameTable] Error setting up aurora: %+v ", err)
	}
	dynamoTableName := "customer-profiles-low-cost-multi-key-same-table-test-" + core.GenerateUniqueId()
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[TestMultipleSearchSameTable] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[TestMultipleSearchSameTable] error deleting table: %v", err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[TestMultipleSearchSameTable] error waiting for table creation: %v", err)
	}
	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)
	cpWriterQueue := buildTestCPWriterQueue(t, envCfg.Region)

	// Init low cost config
	mergeQueue := buildTestMergeQueue(t, envCfg.Region)
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
	domainName := randomDomain("multiKeySameTbl")

	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[TestMultipleSearchSameTable] Error creating domain: %v", err)
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
		t.Errorf("[TestMultipleSearchSameTable] error creating mapping: %v", err)
	}

	for i := 0; i < 6; i++ {
		profileOneTravelerID := "test_traveler"
		obj1, _ := json.Marshal(map[string]string{
			"traveler_id":    profileOneTravelerID + strconv.Itoa(i%3),
			"accp_object_id": "test_accp_object_id" + strconv.Itoa(i),
			"first_name":     "William" + strconv.Itoa(i%3),
			"middle_name":    "Jennings" + strconv.Itoa(i%2),
			"last_name":      "Bryan",
			"created_on":     "test_created_on" + strconv.Itoa(i%3),
			"created_by":     "test_created_by" + strconv.Itoa(i%2),
		})
		err = lc.PutProfileObject(string(obj1), guestProfileType)
		if err != nil {
			t.Errorf("[TestMultipleSearchSameTable] error creating profile object: %v", err)
		}
	}

	firstNameValues := []string{"William2"}
	fromValues := []string{"Jennings0"}

	multiValueSearch := make(map[string][]string)
	multiValueSearch["guest_profile.first_name"] = firstNameValues
	multiValueSearch["guest_profile.middle_name"] = fromValues
	cids, err := lc.SearchOnMultipleKeysProfile(multiValueSearch)
	if err != nil {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	}
	profile, err := lc.Data.FetchProfiles(domainName, cids)
	if len(profile) != 1 {
		t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
	} else {
		if profile[0].LastName != "Bryan" {
			t.Errorf("[TestMultipleSearch] error searching profiles: %v", err)
		}
	}

	err = lc.DeleteDomainByName(domainName)
	if err != nil {
		t.Errorf("[TestMultipleSearchSameTable] error deleting domain: %v", err)
	}
}

func TestSearchAddressFields(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Errorf("[TestSearchAddressFields] Error setting up aurora: %+v ", err)
		return
	}
	// DynamoDB
	dynamoTableName := "customer-profiles-low-cost-upt-search-address-" + core.GenerateUniqueId()
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("[TestSearchAddressFields] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[TestMultipleSearchSameTable] error deleting table: %v", err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Errorf("[TestSearchAddressFields] error waiting for table creation: %v", err)
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

	domainName := "upt_search_address"
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[TestSearchAddressFields] Error creating domain: %v", err)
	}

	objectTypeName := "upt_addresses_attributes"
	mapping := []FieldMapping{
		{
			Type:   "STRING",
			Source: "_source.first_name",
			Target: "_profile.FirstName",
		},
		{
			Type:   "STRING",
			Source: "_source.middle_name",
			Target: "_profile.MiddleName",
		},
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
		t.Errorf("[TestSearchAddressFields] CreateMapping failed: %s", err)
	}

	profileOneTravelerID := "test_profile_one_traveler_id"
	obj1, _ := json.Marshal(map[string]string{
		"traveller_id":           profileOneTravelerID,
		"accp_object_id":         "test_accp_object_id",
		"first_name":             "Carl",
		"middle_name":            "Marks",
		"address_business_state": "MA",
		"address_mailing_state":  "TX",
	})
	profileTwoTravelerID := "test_profile_one_traveler_id_two"
	obj2, _ := json.Marshal(map[string]string{
		"traveller_id":           profileTwoTravelerID,
		"accp_object_id":         "test_accp_object_id_2",
		"first_name":             "Bob",
		"middle_name":            "Marks",
		"address_business_state": "CA",
		"address_mailing_state":  "TX",
	})
	err = lc.PutProfileObject(string(obj1), objectTypeName)
	if err != nil {
		t.Errorf("[TestSearchAddressFields] error creating profile object: %v", err)
	}
	err = lc.PutProfileObject(string(obj2), objectTypeName)
	if err != nil {
		t.Errorf("[TestSearchAddressFields] error creating profile object: %v", err)
	}
	multiValueSearch := make(map[string][]string)
	multiValueSearch["ShippingAddress.State"] = []string{"MA"}
	profile, err := lc.SearchProfiles("ShippingAddress.State", []string{"MA"})
	if err != nil {
		t.Errorf("[TestSearchAddressFields] error searching profiles: %v", err)
	}
	if len(profile) == 0 {
		t.Logf("[TestOverrideWithEmptyString] waiting 5 seconds")
		time.Sleep(5 * time.Second)
		profile, err = lc.SearchProfiles("ShippingAddress.State", []string{"MA"})
		if err != nil {
			t.Errorf("[TestOverrideWithEmptyString] error searching profiles: %v", err)
		}
	}
	if len(profile) == 0 {
		t.Errorf("[TestSearchAddressFields] wrong number of profiles: %v", len(profile))
	} else {
		if profile[0].FirstName != "Carl" {
			t.Errorf("[TestSearchAddressFields] wrong profile returned: %v", profile[0])
		}
	}

	_, err = lc.SearchProfiles("invalid_search_key", []string{"value"})
	if err == nil {
		t.Errorf("[TestSearchAddressFields] searching on invalid key should return an error")
	}

	// Cleanup
	err = lc.DeleteDomain()
	if err != nil {
		t.Errorf("[TestSearchAddressFields] DeleteDomain failed: %s", err)
	}
}

func TestSearchAdvancedAddressFields(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[TestSearchAdvancedAddressFields] Error setting up aurora: %+v ", err)
	}
	// DynamoDB
	dynamoTableName := "customer-profiles-low-cost-upt-search-advanced-" + core.GenerateUniqueId()
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[TestSearchAdvancedAddressFields] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[TestSearchAdvancedAddressFields] DeleteTable failed: %s", err)
		}
	})
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[TestSearchAdvancedAddressFields] error waiting for table creation: %v", err)
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

	domainName := randomDomain("upt_adv_search")
	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", "", DomainOptions{})
	if err != nil {
		t.Errorf("[TestSearchAddressFields] Error creating domain: %v", err)
	}
	t.Cleanup(func() {
		err = lc.DeleteDomain()
		if err != nil {
			t.Errorf("[TestUpdateProfile] DeleteDomain failed: %s", err)
		}
	})

	objectTypeName := "upt_addresses_attributes"
	mapping := []FieldMapping{
		{
			Type:   "STRING",
			Source: "_source.first_name",
			Target: "_profile.FirstName",
		},
		{
			Type:   "STRING",
			Source: "_source.middle_name",
			Target: "_profile.MiddleName",
		},
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
		t.Errorf("[TestSearchAddressFields] CreateMapping failed: %s", err)
	}

	secondObjectTypeName := "upt_addresses_attributes_2"
	secondMapping := []FieldMapping{
		{
			Type:   "STRING",
			Source: "_source.first_name",
			Target: "_profile.FirstName",
		},
		{
			Type:   "STRING",
			Source: "_source.middle_name",
			Target: "_profile.MiddleName",
		},
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
			Source: "_source.address_business_state_copy",
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

	err = lc.CreateMapping(secondObjectTypeName, domainName, secondMapping)
	if err != nil {
		t.Errorf("[TestSearchAddressFields] CreateMapping failed: %s", err)
	}

	profileOneTravelerID := "test_profile_one_traveler_id"
	obj1, _ := json.Marshal(map[string]string{
		"traveller_id":           profileOneTravelerID,
		"accp_object_id":         "test_accp_object_id",
		"first_name":             "Carl",
		"middle_name":            "Marks",
		"address_business_state": "MA",
		"address_mailing_state":  "TX",
	})
	profileTwoTravelerID := "test_profile_one_traveler_id_two"
	obj2, _ := json.Marshal(map[string]string{
		"traveller_id":           profileTwoTravelerID,
		"accp_object_id":         "test_accp_object_id_2",
		"first_name":             "Bob",
		"middle_name":            "Marks",
		"address_business_state": "CA",
		"address_mailing_state":  "TX",
	})
	err = lc.PutProfileObject(string(obj1), objectTypeName)
	if err != nil {
		t.Errorf("[TestOverrideWithEmptyString] error creating profile object: %v", err)
	}
	err = lc.PutProfileObject(string(obj2), objectTypeName)
	if err != nil {
		t.Errorf("[TestOverrideWithEmptyString] error creating profile object: %v", err)
	}
	multiValueSearch := make(map[string][]string)
	multiValueSearch["ShippingAddress.State"] = []string{"MA"}
	profile, err := lc.SearchProfiles("ShippingAddress.State", []string{"MA"})
	if err != nil {
		t.Errorf("[TestOverrideWithEmptyString] error searching profiles: %v", err)
	}
	if len(profile) == 0 {
		t.Logf("[TestOverrideWithEmptyString] waiting 5 seconds")
		time.Sleep(5 * time.Second)
		profile, err = lc.SearchProfiles("ShippingAddress.State", []string{"MA"})
		if err != nil {
			t.Errorf("[TestOverrideWithEmptyString] error searching profiles: %v", err)
		}
	}
	if len(profile) == 0 {
		t.Errorf("[TestOverrideWithEmptyString] wrong number of profiles: %v", len(profile))
	} else {
		if profile[0].FirstName != "Carl" {
			t.Errorf("[TestOverrideWithEmptyString] wrong profile returned: %v", profile[0])
		}
	}
}
