package customerprofiles

import (
	"log"
	"os"
	core "tah/core/core"
	kms "tah/core/kms"
	s3 "tah/core/s3"
	"testing"
	"time"
)

// var env = os.Getenv("TAH_CORE_ENV")
var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")
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
		Indexes:     []string{"UNIQUE", "PROFILE"},
		Searcheable: true,
	},
}
var profile_id = "id-123456"
var profileObject = `{"firstName": "First", "lastName": "Last", "profile_id": "` + profile_id + `"}`

func TestCustomerProfiles(t *testing.T) {
	s3c := s3.InitRegion(TAH_CORE_REGION)
	kmsc := kms.Init(TAH_CORE_REGION)
	bucketName, err0 := s3c.CreateRandomBucket("tah-customerprofile-unit-test")
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
	keyArn, err1 := kmsc.CreateKey("tah-unit-test-key")
	if err1 != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err1)
	}

	// Admin Creation Tasks
	profilesSvc2 := Init(TAH_CORE_REGION)
	domainName := "test-domain" + time.Now().Format("2006-01-02-15-04-05")
	err7 := profilesSvc2.CreateDomain(domainName, true, keyArn, map[string]string{"key": "value"})
	if err7 != nil {
		t.Errorf("[TestCustomerProfiles] Error creating domain %v", err7)
	}
	dom, err5 := profilesSvc2.GetDomain()
	if err5 != nil {
		t.Errorf("[TestCustomerProfiles] Error retrieving domain %v", err5)
	}
	if dom.Tags["key"] != "value" {
		t.Errorf("[TestCustomerProfiles] Error retrieving domain: missing tags %v", dom)
	}
	profilesSvc2.DomainName = dom.Name
	objectName := "sample_object"
	err8 := profilesSvc2.CreateMapping(objectName, "description of the test mapping", fieldMappings)
	if err8 != nil {
		t.Errorf("[TestCustomerProfiles] Error creating mapping %v", err8)
	}
	err14 := profilesSvc2.WaitForMappingCreation(objectName)
	if err14 != nil {
		t.Errorf("[TestCustomerProfiles] Error creating mapping %v", err14)
	}
	_, err12 := profilesSvc2.GetMappings()
	if err12 != nil {
		t.Errorf("[TestCustomerProfiles] Error getting mapping %v", err12)
	}
	flowName := "integration_test"
	err13 := profilesSvc2.PutIntegration(flowName, objectName, bucketName, fieldMappings, time.Now())
	if err13 != nil {
		t.Errorf("[TestCustomerProfiles] Error putting integration %v", err13)
	}
	err := profilesSvc2.WaitForIntegrationCreation(flowName + "_OnDemand")
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error waiting for integration %v", err13)
	}
	integrations, err14 := profilesSvc2.GetIntegrations()
	if err14 != nil {
		t.Errorf("[TestCustomerProfiles] Error getting integrations %v", err14)
	}
	if len(integrations) != 2 {
		t.Errorf("[TestCustomerProfiles] Error with PutIntegraton, expected integrations not found")
	}

	// Profile Management
	err = profilesSvc2.PutProfileObject(profileObject, objectName)
	if err != nil {
		t.Errorf("[TestCustomerProfiles] Error putting test profile %v", err)
	}
	i, max, duration, accpProfileId := 0, 12, 5*time.Second, ""
	for ; i < max; i++ {
		accpProfileId, err = profilesSvc2.GetProfileId(profile_id)
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
	log.Printf("Customer's Profile ID: %v", accpProfileId)
	_, _, err15 := profilesSvc2.GetErrors()
	if err15 != nil {
		t.Errorf("[TestCustomerProfiles] Error getting ingestion errors %v", err15)
	}

	// Admin Deletion Tasks
	err9 := profilesSvc2.DeleteMapping(objectName)
	if err9 != nil {
		t.Errorf("[TestCustomerProfiles] Error deleting mapping %v", err9)
	}
	err10 := profilesSvc2.DeleteDomainByName(domainName)
	if err10 != nil {
		t.Errorf("[TestCustomerProfiles] Error deleting domain %v", err10)
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

func TestFieldMappings(t *testing.T) {
	sources := fieldMappings.GetSourceNames()
	expectedSources := []string{"firstName", "lastName", "profile_id"}
	if core.IsEqual(sources, expectedSources) == false {
		t.Errorf("[GetSourceNames] not working as expected")
	}
}
