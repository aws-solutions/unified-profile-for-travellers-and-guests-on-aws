package usecase

import (
	"log"
	"os"
	customerprofiles "tah/core/customerprofiles"
	model "tah/ucp/src/business-logic/model"
	"testing"
)

var UCP_REGION = os.Getenv("UCP_REGION")

func TestDomainCreationDeletion(t *testing.T) {
	log.Printf("Testing domain creation and deletion")
	testDomain := "ucp-component-test-domain"
	var profileClient = customerprofiles.InitWithDomain("", UCP_REGION)
	req := model.UCPRequest{
		Domain: model.Domain{Name: testDomain},
	}
	log.Printf("Testing domain creation")
	_, err := CreateUcpDomain(req, profileClient)
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
