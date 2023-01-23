package usecase

import (
	"log"
	"os"
	customerprofiles "tah/core/customerprofiles"
	model "tah/ucp/src/business-logic/model"
	"testing"
)

var UCP_REGION = getRegion()

//check test.sh script for definition oof these environement variables
var KMS_KEY_PROFILE_DOMAIN = os.Getenv("KMS_KEY_PROFILE_DOMAIN")
var CONNECT_PROFILE_SOURCE_BUCKET = os.Getenv("CONNECT_PROFILE_SOURCE_BUCKET")

func TestDomainCreationDeletion(t *testing.T) {
	log.Printf("Testing domain creation and deletion")
	testDomain := "ucp-component-test-domain"
	var profileClient = customerprofiles.InitWithDomain("", UCP_REGION)
	req := model.UCPRequest{
		Domain: model.Domain{Name: testDomain},
	}
	log.Printf("Testing domain creation")
	_, err := CreateUcpDomain(req, profileClient, KMS_KEY_PROFILE_DOMAIN, CONNECT_PROFILE_SOURCE_BUCKET)
	if err != nil {
		t.Errorf("Error creating UCP domain: %v", err)
	}
	profileClient = customerprofiles.InitWithDomain(testDomain, UCP_REGION)
	log.Printf("Testing domain deletion")
	_, err = DeleteUcpDomain(req, profileClient)
	if err != nil {
		t.Errorf("Error deleting UCP domain: %v", err)
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
