package customerprofiles

import (
	"os"
	kms "tah/core/kms"
	s3 "tah/core/s3"
	"testing"
)

// var env = os.Getenv("TAH_CORE_ENV")
var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")

func TestCustomerProfiles(t *testing.T) {
	/*profilesSvc := InitWithDomain("cloudrack-profiles-dev-2", "eu-central-1")
	_, err := profilesSvc.SearchProfiles("LastName", []string{"Rollat"})
	if err != nil {
		t.Errorf("[TestCustomerProfiles]error searching profile: %v", err)
	}
	_, _, err2 := profilesSvc.GetErrors()
	if err2 != nil {
		t.Errorf("[TestCustomerProfiles] Error getting ingestion errors %v", err2)
	}
	_, err3 := profilesSvc.GetProfile("3ca5c40566ba48db8db2a0b6f5c1936d")
	if err3 != nil {
		t.Errorf("[TestCustomerProfiles] Error rettreiving profile %v", err3)
	}
	_, err4 := profilesSvc.GetProfileObject("_order", "3ca5c40566ba48db8db2a0b6f5c1936d")
	if err4 != nil {
		t.Errorf("[TestCustomerProfiles] Error retrieving profile objects %v", err4)
	}
	_, err5 := profilesSvc.GetDomain()
	if err5 != nil {
		t.Errorf("[TestCustomerProfiles] Error retrieving domain %v", err5)
	}
	_, err6 := profilesSvc.IndexMatchesById()
	if err6 != nil {
		t.Errorf("[TestCustomerProfiles] Error indexing matches %v", err6)
	}
		_, err11 := profilesSvc.GetMappings()
	if err11 != nil {
		t.Errorf("[TestCustomerProfiles] Error getting mapping %v", err11)
	}

	_, err13 := profilesSvc.GetIntegrations()
	if err13 != nil {
		t.Errorf("[TestCustomerProfiles] Error getting integrations %v", err13)
	}
	*/

	s3c := s3.InitRegion(TAH_CORE_REGION)
	kmsc := kms.Init(TAH_CORE_REGION)
	bucketName, err0 := s3c.CreateRandomBucket("tah-customerprofile-unit-test")
	if err0 != nil {
		t.Errorf("Ccound not crerate bucket to unit test TAH Core Customer Profile %v", err0)
	}
	resource := []string{"arn:aws:s3:::" + bucketName, "arn:aws:s3:::" + bucketName + "/*"}
	actions := []string{"s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy"}
	principal := map[string][]string{"Service": []string{"appflow.amazonaws.com"}}
	err0 = s3c.AddPolicy(bucketName, resource, actions, principal)
	if err0 != nil {
		t.Errorf("error adding bucket policy %+v", err0)
	}
	keyArn, err1 := kmsc.CreateKey("tah-unit-test-key")
	if err1 != nil {
		t.Errorf("Ccound not create KMS key to unit test UCP %v", err1)
	}

	profilesSvc2 := Init(TAH_CORE_REGION)
	err7 := profilesSvc2.CreateDomain("test-domain", true, keyArn, map[string]string{"key": "value"})
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
	err8 := profilesSvc2.CreateMapping("testMapping", "description of the test mapping", []FieldMapping{
		{
			Type:        "STRING",
			Source:      "firstName",
			Target:      "_profile.FirstName",
			Indexes:     []string{},
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "lastName",
			Target:      "_profile.LastName",
			Indexes:     []string{},
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "accountId",
			Target:      "_profile.AccountNumber",
			Indexes:     []string{"UNIQUE", "PROFILE"},
			Searcheable: true,
		},
	})
	if err8 != nil {
		t.Errorf("[TestCustomerProfiles] Error creating mapping %v", err8)
	}
	// err14 := profilesSvc2.WaitForMappingCreation("testMapping")
	// if err14 != nil {
	// 	t.Errorf("[TestCustomerProfiles] Error creating mapping %v", err14)
	// }
	_, err12 := profilesSvc2.GetMappings()
	if err12 != nil {
		t.Errorf("[TestCustomerProfiles] Error getting mapping %v", err12)
	}
	_, err13 := profilesSvc2.PutIntegration("testMapping", bucketName)
	if err12 != nil {
		t.Errorf("[TestCustomerProfiles] Error putting integration %v", err13)
	}
	_, err14 := profilesSvc2.GetIntegrations()
	if err14 != nil {
		t.Errorf("[TestCustomerProfiles] Error getting integrations %v", err14)
	}
	_, _, err15 := profilesSvc2.GetErrors()
	if err15 != nil {
		t.Errorf("[TestCustomerProfiles] Error getting ingestion errors %v", err15)
	}
	err9 := profilesSvc2.DeleteMapping("testMapping")
	if err9 != nil {
		t.Errorf("[TestCustomerProfiles] Error deleting mapping %v", err9)
	}

	err10 := profilesSvc2.DeleteDomainByName("test-domain")
	if err10 != nil {
		t.Errorf("[TestCustomerProfiles] Error deleting domain %v", err10)
	}
	err := s3c.DeleteBucket(bucketName)
	if err != nil {
		t.Errorf("Error deleting bucket %v", err)
	}
	err = kmsc.DeleteKey(keyArn)
	if err != nil {
		t.Errorf("Error deleting key %v", err)
	}
}
