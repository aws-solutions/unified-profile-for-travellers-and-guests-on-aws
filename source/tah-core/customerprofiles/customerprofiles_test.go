// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofiles

import (
	"errors"
	"log"
	"strconv"
	core "tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	kinesis "tah/upt/source/tah-core/kinesis"
	kms "tah/upt/source/tah-core/kms"
	s3 "tah/upt/source/tah-core/s3"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

var fieldMappings = FieldMappings{
	{
		Type:        "STRING",
		Source:      "_source.firstName",
		Target:      "_profile.FirstName",
		Indexes:     []string{},
		Searcheable: true,
	},
	{
		Type:        "STRING",
		Source:      "_source.lastName",
		Target:      "_profile.LastName",
		Indexes:     []string{},
		Searcheable: true,
	},
	{
		Type:        "STRING",
		Source:      "_source.profile_id",
		Target:      "_profile.Attributes.profile_id",
		Indexes:     []string{"PROFILE"},
		Searcheable: true,
	},
	{
		Type:    "STRING",
		Source:  "_source.booking_id",
		Target:  "booking_id",
		Indexes: []string{"UNIQUE"},
		KeyOnly: true,
	},
}

type ExpectedProfile struct {
	FirstName         string
	LastName          string
	EmailAddress      string
	BusinessName      string
	HomePhoneNumber   string
	MobilePhoneNumber string
	Gender            string
	Address1          string
	Address2          string
	Address3          string
	Address4          string
	Province          string
	State             string
	Country           string
	Attributes        map[string]string
	PostalCode        string
}

var profile_id = "id-123456"
var profileObject = `{"firstName": "First", "lastName": "Last", "profile_id": "` + profile_id + `", "booking_id": "booking123", "testInt": 123, "testFloat": 123.456, "testBool": true}`

var profile_id_merge = "id-123457"
var profileObject_merge = `{"firstName": "Firsty", "lastName": "Last", "profile_id": "` + profile_id_merge + `", "booking_id": "booking1234", "testInt": 123, "testFloat": 123.456, "testBool": false}`

// object to test case when an object arrives for a profile that has been merged
var profileObject_merge2 = `{"firstName": "Firsty", "lastName": "Last", "profile_id": "` + profile_id_merge + `", "booking_id": "booking12345", "testInt": 123, "testFloat": 123.456, "testBool": false}`

func TestMergeMany(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	s3c := s3.InitRegion(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	kmsc := kms.Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	log.Printf("[TestCustomerProfiles] Setup test environment")
	bucketName, err0 := s3c.CreateRandomBucket("tah-customerprofiles-unit-test")
	if err0 != nil {
		t.Errorf("Could not create bucket to unit test TAH Core Customer Profile %v", err0)
	}
	resource := []string{"arn:aws:s3:::" + bucketName, "arn:aws:s3:::" + bucketName + "/*"}
	actions := []string{"s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy"}
	principal := map[string][]string{"Service": {"appflow.amazonaws.com"}}
	err0 = s3c.AddPolicy(bucketName, resource, actions, principal)
	if err0 != nil {
		t.Errorf("error adding bucket policy %+v", err0)
	}
	keyArn, err := kmsc.CreateKey("tah-unit-test-key")
	if err != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err)
	}

	// Admin Creation Tasks
	profilesSvc2 := Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	domainName := "test-domain_2" + time.Now().Format("2006-01-02-15-04-05")
	domExists, _ := profilesSvc2.DomainExists(domainName)
	if domExists {
		t.Errorf("[TestCustomerProfiles] Domain should not exist before creation")
	}
	err = profilesSvc2.CreateDomain(domainName, keyArn, map[string]string{"key": "value"}, "", DomainOptions{})
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error creating domain %v", err)
	}

	objectName := "sample_object"
	err8 := profilesSvc2.CreateMapping(objectName, "description of the test mapping", fieldMappings)
	if err8 != nil {
		t.Errorf("[TestCustomerProfiles] Error creating mapping %v", err8)
	}
	err = profilesSvc2.WaitForMappingCreation(objectName)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error creating mapping %v", err)
	}

	testSize := 20

	// Profile Management
	log.Printf("[TestCustomerProfiles] Creating %v Profiles", testSize)
	for i := 0; i < testSize; i++ {
		var profileObject1 = `{"firstName": "First", "lastName": "Last", "profile_id": "source_id_` + strconv.Itoa(i) + `", "booking_id": "booking123", "testInt": 123, "testFloat": 123.456, "testBool": true}`
		var profileObject2 = `{"firstName": "First", "lastName": "Last", "profile_id": "target_id_` + strconv.Itoa(i) + `", "booking_id": "booking123", "testInt": 123, "testFloat": 123.456, "testBool": true}`
		profilesSvc2.PutProfileObject(profileObject1, objectName)
		profilesSvc2.PutProfileObject(profileObject2, objectName)
	}

	for i := 0; i < testSize; i++ {
		_, err := WaitForProfileCreationByProfileId("source_id_"+strconv.Itoa(i), "profile_id", profilesSvc2)
		if err != nil {
			t.Errorf("[TestCustomerProfiles] Error waiting for source profiles creation %v", err)
		}
		_, err = WaitForProfileCreationByProfileId("target_id_"+strconv.Itoa(i), "profile_id", profilesSvc2)
		if err != nil {
			t.Errorf("[TestCustomerProfiles] Error waiting for target profiles creation %v", err)
		}
	}

	profilePairs := []ProfilePair{}
	profileIdMapping := map[string]string{}
	for i := 0; i < testSize; i++ {
		accpProfileId1, err := profilesSvc2.GetProfileId("profile_id", "source_id_"+strconv.Itoa(i))
		if err != nil {
			t.Errorf("[TestCustomerProfiles] Error getting profile id %v", err)
		}
		accpProfileId2, err := profilesSvc2.GetProfileId("profile_id", "target_id_"+strconv.Itoa(i))
		if err != nil {
			t.Errorf("[TestCustomerProfiles] Error getting profile id %v", err)
		}
		log.Printf("[TestCustomerProfiles] Adding  profile pair (%s,%s) to merge", accpProfileId1, accpProfileId2)
		profilePairs = append(profilePairs, ProfilePair{SourceID: accpProfileId1, TargetID: accpProfileId2})
		profileIdMapping["source_id_"+strconv.Itoa(i)] = accpProfileId1
		profileIdMapping["target_id_"+strconv.Itoa(i)] = accpProfileId2
	}
	log.Printf("[TestCustomerProfiles] Merging %v profile pairs", testSize)
	out, err := profilesSvc2.MergeMany(profilePairs)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error merging %v profile pairs %v", testSize, err)
	}
	if out != "SUCCESS" {
		t.Errorf("[TestCustomerProfiles] Error merging %v profile pairs, expected success", testSize)
	}

	for i := 0; i < testSize; i++ {
		err := WaitForMerge(profileIdMapping["source_id_"+strconv.Itoa(i)], "target_id_"+strconv.Itoa(i), profilesSvc2)
		if err != nil {
			t.Errorf("[TestCustomerProfiles] Failed to merge profile %v and %v pairs. Error: %v", "source_id_"+strconv.Itoa(i), "target_id_"+strconv.Itoa(i), err)
		} else {
			log.Printf("[TestCustomerProfiles] Successfully merged profile %v and %v  pairs", "source_id_"+strconv.Itoa(i), "target_id_"+strconv.Itoa(i))
		}
		/*log.Printf("Wait for profile to be searchable by %v", "source_id_"+strconv.Itoa(i))
		  err = WaitForMerge(profileIdMapping["source_id_"+strconv.Itoa(i)], "source_id_"+strconv.Itoa(i), profilesSvc2)
		  if err != nil {
		      t.Errorf("[TestCustomerProfiles] profile not serchable by %v (error: %v)", "source_id_"+strconv.Itoa(i), err)
		  }*/

	}

	// Admin Deletion Tasks
	err9 := profilesSvc2.DeleteMapping(objectName)
	if err9 != nil {
		t.Errorf("[TestCustomerProfiles] Error deleting mapping %v", err9)
	}
	err = profilesSvc2.DeleteDomainByName(domainName)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error deleting domain %v", err0)
	}
	err = s3c.DeleteBucket(bucketName)
	if err != nil {
		t.Errorf("Error deleting bucket %v", err)
	}
	err = kmsc.DeleteKey(keyArn)
	if err != nil {
		t.Errorf("Error deleting key %v", err)
	}
}

func TestCustomerProfiles(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	s3c := s3.InitRegion(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	kmsc := kms.Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	log.Printf("[TestCustomerProfiles] Setup test environment")
	bucketName, err0 := s3c.CreateRandomBucket("tah-customerprofiles-unit-test")
	if err0 != nil {
		t.Errorf("Could not create bucket to unit test TAH Core Customer Profile %v", err0)
	}
	t.Cleanup(func() {
		err := s3c.DeleteBucket(bucketName)
		if err != nil {
			t.Errorf("Error deleting bucket %v", err)
		}
	})
	resource := []string{"arn:aws:s3:::" + bucketName, "arn:aws:s3:::" + bucketName + "/*"}
	actions := []string{"s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy"}
	principal := map[string][]string{"Service": {"appflow.amazonaws.com"}}
	err0 = s3c.AddPolicy(bucketName, resource, actions, principal)
	if err0 != nil {
		t.Errorf("error adding bucket policy %+v", err0)
	}
	keyArn, err := kmsc.CreateKey("tah-unit-test-key")
	if err != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err)
	}
	t.Cleanup(func() {
		err := kmsc.DeleteKey(keyArn)
		if err != nil {
			t.Errorf("Error deleting key %v", err)
		}
	})

	// Admin Creation Tasks
	profilesSvc2 := Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	domainName := "test-domain" + time.Now().Format("2006-01-02-15-04-05")
	domExists, _ := profilesSvc2.DomainExists(domainName)
	if domExists {
		t.Errorf("[TestCustomerProfiles] Domain should not exist before creation")
	}
	err7 := profilesSvc2.CreateDomain(domainName, keyArn, map[string]string{"key": "value"}, "", DomainOptions{})
	if err7 != nil {
		t.Errorf("[TestCustomerProfiles] Error creating domain %v", err7)
	}
	t.Cleanup(func() {
		err := profilesSvc2.DeleteDomainByName(domainName)
		if err != nil {
			t.Errorf("[TestCustomerProfiles] Error deleting domain %v", err)
		}
	})
	dom, err5 := profilesSvc2.GetDomain()
	if err5 != nil {
		t.Errorf("[TestCustomerProfiles] Error retrieving domain %v", err5)
	}
	if dom.Tags["key"] != "value" {
		t.Errorf("[TestCustomerProfiles] Error retrieving domain: missing tags %v", dom)
	}
	domExists, _ = profilesSvc2.DomainExists(domainName)
	if !domExists {
		t.Errorf("[TestCustomerProfiles] Domain should exist after creation")
	}
	profilesSvc2.DomainName = dom.Name
	objectName := "sample_object"
	err8 := profilesSvc2.CreateMapping(objectName, "description of the test mapping", fieldMappings)
	if err8 != nil {
		t.Errorf("[TestCustomerProfiles] Error creating mapping %v", err8)
	}
	err = profilesSvc2.WaitForMappingCreation(objectName)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error creating mapping %v", err)
	}
	mappings, err := profilesSvc2.GetMappings()
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting mapping %v", err)
	}
	log.Printf("[TestCustomerProfiles] Mapping: %+v", mappings)

	flowName := "integration_test" + time.Now().Format("2006-01-02-15-04-05")
	err = profilesSvc2.PutIntegration(flowName, objectName, bucketName, "prefix", fieldMappings, time.Now())
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error putting integration %v", err)
	}
	err = profilesSvc2.WaitForIntegrationCreation(flowName + "_OnDemand")
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error waiting for integration %v", err)
	}
	integrations, err := profilesSvc2.GetIntegrations()
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting integrations %v", err)
	}
	if len(integrations) != 2 {
		t.Errorf("[TestCustomerProfiles] Error with PutIntegraton, expected integrations not found")
	}

	// Profile Management
	log.Printf("[TestCustomerProfiles] Create profile1")
	err = profilesSvc2.PutProfileObject(profileObject, objectName)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error putting test profile %v", err)
	}
	i, max, duration, accpProfileId := 0, 12, 5*time.Second, ""
	for ; i < max; i++ {
		accpProfileId, err = profilesSvc2.GetProfileId("profile_id", profile_id)
		if accpProfileId != "" && err == nil {
			log.Printf("Customer's Profile ID: %v", accpProfileId)
			break
		}
		log.Printf("[TestCustomerProfiles] Profile is not available yet, waiting 5 seconds")
		time.Sleep(duration)
	}
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting test profile ID %v", err)
	}
	if accpProfileId == "" {
		t.Errorf("[TestCustomerProfiles] Error getting test profile ID")
	}
	log.Printf("[TestCustomerProfiles] Profile1 has accp_id: %v", accpProfileId)
	profile, err := profilesSvc2.GetProfile(accpProfileId, []string{objectName}, []PaginationOptions{})
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting profile ID %v", err)
	}
	if len(profile.ProfileObjects) == 0 {
		t.Errorf("[TestCustomerProfiles] Error: Unable to fetch custom profile objects")
	}
	_, _, err = profilesSvc2.GetErrors()
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting ingestion errors %v", err)
	}

	//Second Profile
	log.Printf("[TestCustomerProfiles] Create profile2")
	err = profilesSvc2.PutProfileObject(profileObject_merge, objectName)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error putting second test profile %v", err)
	}
	i, max, duration, accpProfileId_merge := 0, 12, 5*time.Second, ""
	for ; i < max; i++ {
		accpProfileId_merge, err = profilesSvc2.GetProfileId("profile_id", profile_id_merge)
		if accpProfileId_merge != "" && err == nil {
			log.Printf("Customer's Profile ID: %v", accpProfileId_merge)
			break
		}
		log.Printf("[TestCustomerProfiles] Second Profile is not available yet, waiting 5 seconds")
		time.Sleep(duration)
	}
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting second test profile ID %v", err)
	}
	if accpProfileId_merge == "" {
		t.Errorf("[TestCustomerProfiles] Error getting second test profile ID")
	}
	log.Printf("[TestCustomerProfiles] Profile2 has accp_id: %v", accpProfileId_merge)
	profile, err = profilesSvc2.GetProfile(accpProfileId_merge, []string{objectName}, []PaginationOptions{})
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting second profile ID %v", err)
	}
	if len(profile.ProfileObjects) == 0 {
		t.Errorf("[TestCustomerProfiles] Error: Unable to fetch custom profile objects")
	}
	_, _, err = profilesSvc2.GetErrors()
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting ingestion errors %v", err)
	}

	//Merge
	log.Printf("[TestCustomerProfiles] Merge profiles 1 and 2")
	out, err := profilesSvc2.MergeProfiles(accpProfileId, accpProfileId_merge)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error merging profiles %v", err)
	}
	log.Printf("[TestCustomerProfiles] Merge result: %v", out)
	if out != "Success!" {
		t.Errorf("[TestCustomerProfiles] Error merging profiles, expected success")
	}

	log.Printf("[TestCustomerProfiles] Wait for merge to take effect")
	profile, err = profilesSvc2.GetProfile(accpProfileId, []string{objectName}, []PaginationOptions{})
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting profile ID %v", err)
	}
	if len(profile.ProfileObjects) != 2 {
		for i := 0; i < 10; i++ {
			profile, err = profilesSvc2.GetProfile(accpProfileId, []string{objectName}, []PaginationOptions{})
			if err != nil {
				t.Errorf("[TestCustomerProfiles] Error getting profile ID %v", err)
			}
			log.Printf("[TestCustomerProfiles] profile1 %+v", profile)
			if len(profile.ProfileObjects) == 2 {
				log.Printf("[TestCustomerProfiles] Merge Success! profile1 has 2 objects after merge!")
				break
			}
		}
		if i == 9 {
			t.Errorf("[TestCustomerProfiles] Error: Merge failed! profile1 should have 2 objects after merge")
		}
	} else {
		log.Printf("[TestCustomerProfiles] Merge Success! profile1 has 2 objects after merge!")
	}

	log.Printf("[TestCustomerProfiles] Checking that profile2 has been deleted")
	// profile, err = profilesSvc2.GetProfile(accpProfileId_merge, []string{objectName}, []PaginationOptions{})
	// if err != nil {
	//  t.Errorf("[TestCustomerProfiles] Error getting profile ID %v", err)
	// }
	log.Printf("[TestCustomerProfiles] profile2 %+v", profile)
	exists, _ := profilesSvc2.Exists(accpProfileId_merge)
	if exists {
		t.Errorf("[TestCustomerProfiles] Error: Profile 2 should no longer exist")
	}
	// if len(profile.ProfileObjects) != 0 {
	//  t.Errorf("[TestCustomerProfiles] Error: Profile 2 should have 0 objects after merge")
	// }

	log.Printf("[TestCustomerProfiles] retreive ACCP ID for profile2 using profile2 old id")
	accpProfileId_new, err := profilesSvc2.GetProfileId("profile_id", profile_id_merge)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting profile ID %v", err)
	}
	if accpProfileId_new != accpProfileId {
		t.Errorf("[TestCustomerProfiles] Error: Profile 2 ID should be associated with profile 1 accpID after merge")
	}
	log.Printf("[TestCustomerProfiles] Profile2 now has accp_id: %v", accpProfileId_new)

	log.Printf("[TestCustomerProfiles] Putting profile object 3 with profile2 old id:  %v", profileObject_merge2)
	err = profilesSvc2.PutProfileObject(profileObject_merge2, objectName)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] failed to put 3rd profile object: %v", err)
	}
	profObject, err := profilesSvc2.GetSpecificProfileObject("booking_id", "booking123", accpProfileId, objectName)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting profile object %v", err)
	}
	log.Printf("[TestCustomerProfiles] profile object: %+v", profObject)
	if profObject.Attributes["testBool"] != "true" {
		t.Errorf("[TestCustomerProfiles] Error: booking_id should be booking123")
	}
	for i := 0; i < 10; i++ {
		profile, err = profilesSvc2.GetProfile(accpProfileId, []string{objectName}, []PaginationOptions{})
		if err != nil {
			t.Errorf("[TestCustomerProfiles] Error getting profile ID %v", err)
		}
		log.Printf("[TestCustomerProfiles] profile1 %+v", profile)
		if len(profile.ProfileObjects) == 3 {
			log.Printf("[TestCustomerProfiles] Success! profile 1 has now %v objects:  %+v", len(profile.ProfileObjects), profile.ProfileObjects)
			break
		}
		log.Printf("[TestCustomerProfiles] profile 1 has %v objects, sleeping for %v", len(profile.ProfileObjects), duration)
		time.Sleep(duration)
		if i == 9 {
			t.Errorf("[TestCustomerProfiles] Error: Wrong number of objects! profile1 should have 3 objects after put profile object")
		}
	}
	exists, _ = profilesSvc2.Exists(accpProfileId_merge)
	if exists {
		t.Errorf("[TestCustomerProfiles] profile 2 shoudl not have been recreated")
	}

	msg, err := profilesSvc2.DeleteProfileObject("booking_id", "booking123", accpProfileId, objectName)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error deleting profile object %v", err)
	}
	log.Printf("[TestCustomerProfiles] Delete profile object result: %v", msg)
	if msg != "OK" {
		t.Errorf("[TestCustomerProfiles] Error deleting profile object, expected success")
	}
	profile, err = profilesSvc2.GetProfile(accpProfileId, []string{objectName}, []PaginationOptions{})
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting profile ID %v", err)
	}
	log.Printf("[TestCustomerProfiles] profile1 %+v", profile)
	if len(profile.ProfileObjects) == 2 {
		log.Printf("[TestCustomerProfiles] Success! profile 1 has now %v objects:  %+v", len(profile.ProfileObjects), profile.ProfileObjects)
	}

	// Create Event Stream
	accpStreamName := "accp-test-stream-" + core.GenerateUniqueId()
	destinationStreamName := "destination-test-stream-" + core.GenerateUniqueId()
	kinesisConfig, err := kinesis.InitAndCreate(destinationStreamName, envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION, core.LogLevelDebug)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error creating kinesis stream %v", err)
	}
	defer func() {
		err = kinesisConfig.Delete(destinationStreamName)
		if err != nil {
			t.Errorf("[TestCustomerProfiles] Error deleting kinesis stream %v", err)
		}
	}()
	_, err = kinesisConfig.WaitForStreamCreation(300)
	stream, err := kinesisConfig.Describe()
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error describing kinesis stream %v", err)
	}
	err = profilesSvc2.CreateEventStream(domainName, accpStreamName, stream.Arn)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error creating event stream %v", err)
	}
	err = profilesSvc2.DeleteEventStream(domainName, accpStreamName)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error deleting event stream %v", err)
	}

	// Admin Deletion Tasks
	err = profilesSvc2.DeleteMapping(objectName)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error deleting mapping %v", err)
	}
}

func TestFieldMappings(t *testing.T) {
	sources := fieldMappings.GetSourceNames()
	expectedSources := []string{"firstName", "lastName", "profile_id", "booking_id"}
	if core.IsEqual(sources, expectedSources) == false {
		t.Errorf("[GetSourceNames] not working as expected. %v shoudl equal %v", sources, expectedSources)
	}
}

func TestSetDomain(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	svc := Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	svc.SetDomain("testDomain")
	if svc.DomainName != "testDomain" {
		t.Errorf("Error setting domain %v", svc.DomainName)
	}
}

func TestProfileFieldMapping(t *testing.T) {
	firstName := "Michael"
	lastName := "Mac"
	var (
		PROFILE_ID          = "12345"
		EXTERNAL_CONNECT_ID = "connect12345"
		FIRST_NAME          = firstName
		LAST_NAME           = lastName
		EMAIL_ADDRESS       = "email@example.com"
		HOME_PHONE_NUMBER   = "XXX-YYY-ZZZZ"
		MOBILE_PHONE_NUMBER = "YYY-XXX-ZZZZ"
		BUSINESS_NAME       = "Some cloud service provider"
		ADDRESS1            = "100 Mathway Ave"
		STATE               = "Massachusetts"
		COUNTRY             = "India"
		POSTAL_CODE         = "XXXXX"
		ADDRESS2            = "Apt -1"
		ADDRESS3            = "Some Or Ville"
		ADDRESS4            = "Fort Med"
		PROVINCE            = "Alberta"
		GENDER              = "Male"
	)
	ATTRIBUTES := map[string]string{
		"profile_id": PROFILE_ID,
	}

	prof := profilemodel.Profile{
		LastUpdated:       time.Now(),
		ProfileId:         EXTERNAL_CONNECT_ID,
		FirstName:         FIRST_NAME,
		LastName:          LAST_NAME,
		EmailAddress:      EMAIL_ADDRESS,
		HomePhoneNumber:   HOME_PHONE_NUMBER,
		MobilePhoneNumber: MOBILE_PHONE_NUMBER,
		BusinessName:      BUSINESS_NAME,
		Address: profilemodel.Address{
			Address1: ADDRESS1,
			State:    STATE,
			Country:  COUNTRY,
		},
		ShippingAddress: profilemodel.Address{
			PostalCode: POSTAL_CODE,
		},
		BillingAddress: profilemodel.Address{
			Address2: ADDRESS2,
		},
		MailingAddress: profilemodel.Address{
			Address3: ADDRESS3,
			Address4: ADDRESS4,
			Province: PROVINCE,
		},
		Gender:     GENDER,
		Attributes: ATTRIBUTES,
	}

	expectedProf := ExpectedProfile{
		FirstName:         FIRST_NAME,
		LastName:          LAST_NAME,
		EmailAddress:      EMAIL_ADDRESS,
		HomePhoneNumber:   HOME_PHONE_NUMBER,
		MobilePhoneNumber: MOBILE_PHONE_NUMBER,
		BusinessName:      BUSINESS_NAME,
		Address1:          ADDRESS1,
		State:             STATE,
		Country:           COUNTRY,
		PostalCode:        POSTAL_CODE,
		Address2:          ADDRESS2,
		Address3:          ADDRESS3,
		Address4:          ADDRESS4,
		Province:          PROVINCE,
		Gender:            GENDER,
		Attributes:        ATTRIBUTES,
	}

	//setup environment
	domainName := "test-domain-profile" + time.Now().Format("2006-01-02-15-04-05")
	profilesSvc2 := SetupCpTestingEnvironment(t, domainName)

	//Adding Profile Level Mapping
	profMappings := BuildProfileFieldMapping()
	err := profilesSvc2.CreateMapping(PROFILE_FIELD_OBJECT_TYPE, "some description", profMappings)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Failed to create profile level mappings %v", err)
	}
	//Serializing profile object
	err = profilesSvc2.PutProfileAsObject(prof)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Failed to insert profile object for profile")
	}
	id, err := WaitForProfileCreationByProfileId(EXTERNAL_CONNECT_ID, CONNECT_ID_ATTRIBUTE, profilesSvc2)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error waiting for profile to be created %v", err)
	}
	profile, err := profilesSvc2.SearchProfiles("_fullName", []string{firstName + " " + lastName})
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error searching for the profile")
	}
	if len(profile) == 0 {
		t.Fatalf("[TestCustomerProfiles] No profiles were returned")
	}
	profile, err = profilesSvc2.SearchProfiles(CONNECT_ID_ATTRIBUTE, []string{EXTERNAL_CONNECT_ID})
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error searching for the profile %v", err)
	}
	if len(profile) == 0 {
		t.Fatalf("[TestCustomerProfiles] No profiles were returned")
	}
	getProfile, err := profilesSvc2.GetProfile(id, []string{}, []PaginationOptions{})
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error getting the profile %v", err)
	}
	validateProfile(t, profile[0], expectedProf)
	validateProfile(t, getProfile, expectedProf)
	objectMappings := BuildGenericIntegrationFieldMappingWithTravelerId()
	err = profilesSvc2.CreateMapping("air_booking", "some description", objectMappings)
	if err != nil {
		t.Fatalf("[TestCustomerProfiles] Failed to create profile level mappings %v", err)
	}
	airObjectAttributes := map[string]interface{}{
		"booking_id":   "testId",
		"traveller_id": "testProfileId",
	}
	airObject := profilemodel.ProfileObject{
		ID:                  "testAccpId",
		Type:                "air_booking",
		AttributesInterface: airObjectAttributes,
		Attributes: map[string]string{
			"booking_id":   "testId",
			"traveller_id": "testProfileId",
		},
	}
	airObjectAttributesNew := map[string]interface{}{
		"booking_id":   "testIdWrong",
		"traveller_id": "testProfileIdWrong",
	}
	airObjectNew := profilemodel.ProfileObject{
		ID:                  "testAccpId",
		Type:                "air_booking",
		AttributesInterface: airObjectAttributesNew,
		Attributes: map[string]string{
			"booking_id":   "testIdWrong",
			"traveller_id": "testProfileIdWrong",
		},
	}
	//send 2nd object with same id, different booking, connect id, original object should be protected
	err = profilesSvc2.PutProfileObjectFromLcs(airObject, EXTERNAL_CONNECT_ID)
	if err != nil {
		t.Fatalf("[TestCustomerProfiles] Failed to insert profile object for air booking")
	}
	err = profilesSvc2.PutProfileObjectFromLcs(airObjectNew, EXTERNAL_CONNECT_ID+"dummy")
	if err != nil {
		t.Fatalf("[TestCustomerProfiles] Failed to insert profile object for air booking")
	}
	err = WaitForExpectedCondition(func() bool {
		getProfile, err = profilesSvc2.GetProfile(id, []string{"air_booking"}, []PaginationOptions{})
		if err != nil {
			t.Logf("[TestCustomerProfiles] Error getting the profile %v", err)
			return false
		}
		if len(getProfile.ProfileObjects) != 1 {
			t.Logf("[TestCustomerProfiles] Profile object not found yet")
			return false
		}
		profObj := getProfile.ProfileObjects[0]
		if profObj.Type != "air_booking" {
			t.Logf("[TestCustomerProfiles] Profile object type is not air_booking")
			return false
		}
		if profObj.Attributes["booking_id"] != "testId" {
			t.Logf("[TestCustomerProfiles] Profile object attribute booking_id is not testId")
			return false
		}
		if profObj.Attributes[CONNECT_ID_ATTRIBUTE] != EXTERNAL_CONNECT_ID {
			t.Logf("[TestCustomerProfiles] Profile object connect id attribute is wrong")
			return false
		}
		if profObj.Attributes[UNIQUE_OBJECT_ATTRIBUTE] != "testAccpId"+EXTERNAL_CONNECT_ID {
			t.Logf("[TestCustomerProfiles] Profile object unique attribute is wrong")
			return false
		}
		testProf, err := profilesSvc2.SearchProfiles("traveller_id", []string{"testProfileId"})
		if err != nil {
			t.Logf("[TestCustomerProfiles] Error searching for the profile %v", err)
			return false
		}
		if len(testProf) != 1 {
			t.Logf("[TestCustomerProfiles] Expected one profile, got %v", len(testProf))
			return false
		}
		if testProf[0].FirstName != FIRST_NAME {
			t.Fatalf("[TestCustomerProfiles] Expected first name %v, got %v", FIRST_NAME, testProf[0].FirstName)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Errorf("[TestProfileInsertionCP] error waiting for profile creation: %v", err)
	}
}

func TestProfileFieldAttributesUpdate(t *testing.T) {
	firstName := "Michael"
	lastName := "Mac"
	var (
		LCS_CONNECT_ID = "12345"
		PROFILE_ID     = "XXXXX"
		FIRST_NAME     = firstName
		LAST_NAME      = lastName
		EMAIL_ADDRESS  = "email@example.com"
		BUSINESS_NAME  = "Some cloud service provider"
		ADDRESS1       = "100 Mathway Ave"
		STATE          = "Massachusetts"
		COUNTRY        = "India"
	)
	ATTRIBUTES := map[string]string{
		"creditcard":  "something",
		"destination": "California",
		"hometown":    "Austin",
		"Purchase":    "snacks",
		"profile_id":  PROFILE_ID,
	}

	prof := profilemodel.Profile{
		LastUpdated:  time.Now(),
		ProfileId:    LCS_CONNECT_ID,
		FirstName:    FIRST_NAME,
		LastName:     LAST_NAME,
		EmailAddress: EMAIL_ADDRESS,
		BusinessName: BUSINESS_NAME,
		Address: profilemodel.Address{
			Address1: ADDRESS1,
			State:    STATE,
			Country:  COUNTRY,
		},
		Attributes: ATTRIBUTES,
	}

	expectedProf := ExpectedProfile{
		FirstName:    FIRST_NAME,
		LastName:     LAST_NAME,
		EmailAddress: EMAIL_ADDRESS,
		BusinessName: BUSINESS_NAME,
		Address1:     ADDRESS1,
		State:        STATE,
		Country:      COUNTRY,
		Attributes:   ATTRIBUTES,
	}

	domainName := "test-domain-profile" + time.Now().Format("2006-01-02-15-04-05")
	profilesSvc2 := SetupCpTestingEnvironment(t, domainName)

	//Adding Profile Level Mapping
	profMappings := BuildProfileFieldMapping()
	err := profilesSvc2.CreateMapping(PROFILE_FIELD_OBJECT_TYPE, "some description", profMappings)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Failed to create profile level mappings %v", err)
	}
	err = profilesSvc2.UpdateProfileAttributesFieldMapping(PROFILE_FIELD_OBJECT_TYPE, "some description", []string{"Attributes.creditcard", "Attributes.destination", "Attributes.hometown", "Attributes.Purchase", "Attributes.profile_id"})
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Failed to update profile level mappings %v", err)
	}
	//Serializing profile object
	err = profilesSvc2.PutProfileAsObject(prof)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Failed to insert profile object for profile")
	}
	_, err = WaitForProfileCreationByProfileId(PROFILE_ID, "profile_id", profilesSvc2)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error waiting for profile to be created %v", err)
	}
	profile, err := profilesSvc2.SearchProfiles("_fullName", []string{firstName + " " + lastName})
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error searching for the profile")
	}
	if len(profile) == 0 {
		t.Fatalf("[TestCustomerProfiles] No profiles were returned")
	}
	profile, err = profilesSvc2.SearchProfiles(CONNECT_ID_ATTRIBUTE, []string{LCS_CONNECT_ID})
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error searching for the profile %v", err)
	}
	validateProfile(t, profile[0], expectedProf)
}

func TestTravellerIdSearchability(t *testing.T) {
	const (
		LCS_CONNECT_ID = "12345"
		PROFILE_ID     = "XXXXX"
		FIRST_NAME     = "John"
		LAST_NAME      = "Calvin"
		EMAIL_ADDRESS  = "email@example.com"
	)
	ATTRIBUTES := map[string]string{
		"profile_id": PROFILE_ID,
	}

	prof := profilemodel.Profile{
		LastUpdated:  time.Now(),
		ProfileId:    LCS_CONNECT_ID,
		FirstName:    FIRST_NAME,
		LastName:     LAST_NAME,
		EmailAddress: EMAIL_ADDRESS,
		Attributes:   ATTRIBUTES,
	}

	expectedProf := ExpectedProfile{
		FirstName:    FIRST_NAME,
		LastName:     LAST_NAME,
		EmailAddress: EMAIL_ADDRESS,
		Attributes:   ATTRIBUTES,
	}

	domainName := "test-traveller-id-search" + time.Now().Format("2006-01-02-15-04-05")
	profilesSvc2 := SetupCpTestingEnvironment(t, domainName)

	//Adding Profile Level Mapping
	profMappings := BuildProfileFieldMapping()
	err := profilesSvc2.CreateMapping(PROFILE_FIELD_OBJECT_TYPE, "some description", profMappings)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Failed to create profile level mappings %v", err)
	}
	//add object level mapping
	objectMappings := BuildGenericIntegrationFieldMappingWithTravelerId()
	err = profilesSvc2.CreateMapping("air_booking", "some description", objectMappings)
	if err != nil {
		t.Fatalf("[TestCustomerProfiles] Failed to create profile level mappings %v", err)
	}
	//add 10 objects
	for i := 0; i < 10; i++ {
		airObjectAttributes := map[string]interface{}{
			"booking_id":   "testId",
			"traveller_id": "testProfileId" + strconv.Itoa(i),
		}
		airObject := profilemodel.ProfileObject{
			ID:                  "testAccpId" + strconv.Itoa(i),
			Type:                "air_booking",
			AttributesInterface: airObjectAttributes,
			Attributes: map[string]string{
				"booking_id":   "testId",
				"traveller_id": "testProfileId" + strconv.Itoa(i),
			},
		}
		err = profilesSvc2.PutProfileObjectFromLcs(airObject, LCS_CONNECT_ID)
		if err != nil {
			t.Fatalf("[TestCustomerProfiles] Failed to insert profile object for air booking")
		}
	}
	//Serializing profile object
	err = profilesSvc2.PutProfileAsObject(prof)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Failed to insert profile object for profile")
	}
	id, err := WaitForProfileCreationByProfileId(PROFILE_ID, "profile_id", profilesSvc2)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error waiting for profile to be created %v", err)
	}

	err = WaitForExpectedCondition(func() bool {
		getProfile, err := profilesSvc2.GetProfile(id, []string{"air_booking"}, []PaginationOptions{})
		if err != nil {
			t.Logf("[TestCustomerProfiles] Error getting the profile %v", err)
			return false
		}
		if len(getProfile.ProfileObjects) != 10 {
			t.Logf("[TestCustomerProfiles] Profile object not found yet")
			return false
		}
		for i := 0; i < 10; i++ {
			profile, err := profilesSvc2.SearchProfiles("traveller_id", []string{"testProfileId" + strconv.Itoa(i)})
			if err != nil {
				t.Logf("[TestCustomerProfiles] Error searching for the profile %v", err)
				return false
			}
			if len(profile) != 1 {
				t.Logf("[TestCustomerProfiles] Expected one profile, got %v", len(profile))
				return false
			}
			validateProfile(t, profile[0], expectedProf)
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Errorf("[TestProfileInsertionCP] error waiting for profile creation: %v", err)
	}
}

func SetupCpTestingEnvironment(t *testing.T, domainName string) (profilesSvc2 *CustomerProfileConfig) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	s3c := s3.InitRegion(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	kmsc := kms.Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	log.Printf("[TestCustomerProfiles] Setup test environment")
	bucketName, err0 := s3c.CreateRandomBucket("tah-customerprofiles-unit-test")
	if err0 != nil {
		t.Errorf("Could not create bucket to unit test TAH Core Customer Profile %v", err0)
	}
	t.Cleanup(func() {
		err := s3c.DeleteBucket(bucketName)
		if err != nil {
			t.Errorf("Error deleting bucket %v", err)
		}
	},
	)
	resource := []string{"arn:aws:s3:::" + bucketName, "arn:aws:s3:::" + bucketName + "/*"}
	actions := []string{"s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy"}
	principal := map[string][]string{"Service": {"appflow.amazonaws.com"}}
	err0 = s3c.AddPolicy(bucketName, resource, actions, principal)
	if err0 != nil {
		t.Errorf("error adding bucket policy %+v", err0)
	}
	keyArn, err := kmsc.CreateKey("tah-unit-test-attributes-key")
	if err != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err)
	}
	t.Cleanup(func() {
		err := kmsc.DeleteKey(keyArn)
		if err != nil {
			t.Errorf("Error deleting key %v", err)
		}
	})
	// Admin Creation Tasks
	profilesSvc2 = Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	err7 := profilesSvc2.CreateDomain(domainName, keyArn, map[string]string{"key": "value"}, "", DomainOptions{})
	if err7 != nil {
		t.Errorf("[TestCustomerProfiles] Error creating domain %v", err7)
	}
	t.Cleanup(func() {
		err := profilesSvc2.DeleteDomainByName(domainName)
		if err != nil {
			t.Errorf("[TestCustomerProfiles] Error deleting domain %v", err)
		}
	})
	return profilesSvc2
}

func WaitForProfileCreationByConnectID(profileId string, profilesSvc2 *CustomerProfileConfig) error {
	i, max, duration := 0, 12, 5*time.Second
	for ; i < max; i++ {
		exists, err := profilesSvc2.Exists(profileId)
		if err != nil {
			return err
		}
		if exists {
			log.Printf("Customer's Profile ID: %v foudn", profileId)
			return nil
		}
		log.Printf("[TestCustomerProfiles] Profile is not available yet, waiting 5 seconds")
		time.Sleep(duration)
	}
	return errors.New("Profile not found after timeout")
}

func WaitForProfileCreationByProfileId(profileId string, keyName string, profilesSvc2 *CustomerProfileConfig) (string, error) {
	i, max, duration := 0, 12, 5*time.Second
	for ; i < max; i++ {
		accpProfileId, _ := profilesSvc2.GetProfileId(keyName, profileId)
		if accpProfileId != "" {
			log.Printf("Successfully found Customer's Profile ID: %v", accpProfileId)
			return accpProfileId, nil
		}
		log.Printf("[TestCustomerProfiles] Profile is not available yet, waiting 5 seconds")
		time.Sleep(duration)
	}
	return "", errors.New("Profile not found after timeout")
}

func WaitingForProfileDeletion(connectID string, profilesSvc2 *CustomerProfileConfig) error {
	log.Printf("[TestCustomerProfiles] Waiting for profile to be deleted")
	i, max, duration := 0, 12, 5*time.Second
	for ; i < max; i++ {
		exists, _ := profilesSvc2.Exists(connectID)
		if !exists {
			log.Printf("Customer's Profile ID: %v has disapeared after merge. returning success", connectID)
			return nil
		}
		log.Printf("[TestCustomerProfiles] Profile %v is not deleted yet, waiting 5 seconds", connectID)
		time.Sleep(duration)
	}
	return errors.New("Profile not deleted after timeout")
}

func WaitForMerge(accpProfileId string, profileId2 string, profilesSvc2 *CustomerProfileConfig) error {
	log.Printf("[TestCustomerProfiles] Waiting for profile to be deleted")
	i, max, duration := 0, 12, 5*time.Second
	for ; i < max; i++ {
		accpProfileId2, _ := profilesSvc2.GetProfileId("profile_id", profileId2)

		if accpProfileId == accpProfileId2 {
			log.Printf("Target profile %s returns connectID %s: Merge successfull", profileId2, accpProfileId)
			return nil
		}
		log.Printf("[TestCustomerProfiles] Both profiles %s and %s have different connect IDs. Merge not yet completed", accpProfileId, accpProfileId2)
		time.Sleep(duration)
	}
	return errors.New("Profile merge failed after timeout")
}

func validateProfile(t *testing.T, prof profilemodel.Profile, expected ExpectedProfile) {
	if prof.FirstName != expected.FirstName {
		t.Errorf("First name does not match the expected value. Expected %s, got %s", expected.FirstName, prof.FirstName)
	}
	if prof.LastName != expected.LastName {
		t.Errorf("Last name does not match the expected value. Expected %s, got %s", expected.LastName, prof.LastName)
	}
	if prof.EmailAddress != expected.EmailAddress {
		t.Errorf("Email address does not match the expected value. Expected %s, got %s", expected.EmailAddress, prof.EmailAddress)
	}
	if prof.HomePhoneNumber != expected.HomePhoneNumber {
		t.Errorf("Home phone number does not match the expected value. Expected %s, got %s", expected.HomePhoneNumber, prof.HomePhoneNumber)
	}
	if prof.MobilePhoneNumber != expected.MobilePhoneNumber {
		t.Errorf("Mobile phone number does not match the expected value. Expected %s, got %s", expected.MobilePhoneNumber, prof.MobilePhoneNumber)
	}
	if prof.BusinessName != expected.BusinessName {
		t.Errorf("Business name does not match the expected value. Expected %s, got %s", expected.BusinessName, prof.BusinessName)
	}

	// Validate primary address
	if prof.Address.Address1 != expected.Address1 {
		t.Errorf("Address1 does not match the expected value. Expected %s, got %s", expected.Address1, prof.Address.Address1)
	}
	if prof.Address.State != expected.State {
		t.Errorf("State does not match the expected value. Expected %s, got %s", expected.State, prof.Address.State)
	}
	if prof.Address.Country != expected.Country {
		t.Errorf("Country does not match the expected value. Expected %s, got %s", expected.Country, prof.Address.Country)
	}

	// Validate shipping address
	if prof.ShippingAddress.PostalCode != expected.PostalCode {
		t.Errorf("Postal code does not match the expected value. Expected %s, got %s", expected.PostalCode, prof.ShippingAddress.PostalCode)
	}

	// Validate billing address
	if prof.BillingAddress.Address2 != expected.Address2 {
		t.Errorf("Address2 does not match the expected value. Expected %s, got %s", expected.Address2, prof.BillingAddress.Address2)
	}

	// Validate mailing address
	if prof.MailingAddress.Address3 != expected.Address3 {
		t.Errorf("Address3 does not match the expected value. Expected %s, got %s", expected.Address3, prof.MailingAddress.Address3)
	}
	if prof.MailingAddress.Address4 != expected.Address4 {
		t.Errorf("Address4 does not match the expected value. Expected %s, got %s", expected.Address4, prof.MailingAddress.Address4)
	}
	if prof.MailingAddress.Province != expected.Province {
		t.Errorf("Province does not match the expected value. Expected %s, got %s", expected.Province, prof.MailingAddress.Province)
	}

	// Validate gender
	if prof.Gender != expected.Gender {
		t.Errorf("Gender does not match the expected value. Expected %s, got %s", expected.Gender, prof.Gender)
	}

	// Checking attributes
	for key, expectedValue := range expected.Attributes {
		actualValue, ok := prof.Attributes[key]
		if !ok {
			t.Errorf("%s not found in attributes", key)
		} else if actualValue != expectedValue {
			t.Errorf("%s does not match the expected value. Expected %s, got %s", key, expectedValue, actualValue)
		}
	}

	// Ensure no unexpected attributes are present
	for key := range prof.Attributes {
		if key == CONNECT_ID_ATTRIBUTE {
			continue
		}
		if _, ok := expected.Attributes[key]; !ok {
			t.Errorf("%s found in attributes, but it is not expected", key)
		}
	}
}

func WaitForExpectedCondition(action func() bool, maxAttempts int, duration time.Duration) error {
	for i := 0; i < maxAttempts; i++ {
		if action() {
			return nil
		}
		log.Printf("[WaitForExpectedCondition] condition not met, waiting %v", duration)
		time.Sleep(duration)
	}
	return errors.New("[WaitForExpectedCondition] condition not met after maximum attempts")
}
