package usecase

import (
	"log"
	"os"
	customerprofiles "tah/core/customerprofiles"
	kms "tah/core/kms"
	s3 "tah/core/s3"
	model "tah/ucp/src/business-logic/model"
	"testing"
)

var UCP_REGION = getRegion()

func TestDomainCreationDeletion(t *testing.T) {
	s3c := s3.InitRegion(UCP_REGION)
	kmsc := kms.Init(UCP_REGION)
	bucketName, err0 := s3c.CreateRandomBucket("ucp-unit-test")
	if err0 != nil {
		t.Errorf("Ccound not crerate bucket to unit test UCP %v", err0)
	}
	resource := []string{"arn:aws:s3:::" + bucketName, "arn:aws:s3:::" + bucketName + "/*"}
	actions := []string{"s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy"}
	principal := map[string][]string{"Service": []string{"appflow.amazonaws.com"}}
	err0 = s3c.AddPolicy(bucketName, resource, actions, principal)
	if err0 != nil {
		t.Errorf("error adding bucket policy %+v", err0)
	}
	keyArn, err1 := kmsc.CreateKey("ucp-unit-test-key")
	if err1 != nil {
		t.Errorf("Ccound not create KMS key to unit test UCP %v", err1)
	}
	log.Printf("Testing domain creation and deletion")
	testDomain := "ucp-component-test-domain"
	var profileClient = customerprofiles.InitWithDomain("", UCP_REGION)
	req := model.UCPRequest{
		Domain: model.Domain{Name: testDomain},
	}
	log.Printf("Testing domain creation")
	_, err := CreateUcpDomain(req, profileClient, keyArn, bucketName)
	if err != nil {
		t.Errorf("Error creating UCP domain: %v", err)
	}
	profileClient = customerprofiles.InitWithDomain(testDomain, UCP_REGION)
	log.Printf("Testing domain deletion")
	_, err = DeleteUcpDomain(req, profileClient)
	if err != nil {
		t.Errorf("Error deleting UCP domain: %v", err)
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

// TODO: moov this somewhere centralized
func getRegion() string {
	//getting region for local testing
	region := os.Getenv("UCP_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}
