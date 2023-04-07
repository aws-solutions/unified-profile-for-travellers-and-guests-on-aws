package admin

import (
	"os"
	"tah/core/core"
	customerprofiles "tah/core/customerprofiles"
	kms "tah/core/kms"
	"testing"
	"time"
)

var UCP_REGION = os.Getenv("UCP_REGION")

func TestSearchDomains(t *testing.T) {
	kmsc := kms.Init(UCP_REGION)
	profilesSvc2 := customerprofiles.Init(UCP_REGION)
	tx := core.NewTransaction("ucp_common_test", "")

	keyArn, err1 := kmsc.CreateKey("tah-unit-test-key")
	if err1 != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err1)
	}
	domainName := "test-domain" + time.Now().Format("2006-01-02-15-04-05")
	domainName2 := "test-domain2" + time.Now().Format("2006-01-02-15-04-05")
	err7 := profilesSvc2.CreateDomainWithQueue(domainName, true, keyArn, map[string]string{"envName": "ucp-common-env"}, "")
	if err7 != nil {
		t.Errorf("[TestSearchDomains] Error creating domain %v", err7)
	}
	err7 = profilesSvc2.CreateDomainWithQueue(domainName2, true, keyArn, map[string]string{"envName": "ucp-common-env-2"}, "")
	if err7 != nil {
		t.Errorf("[TestSearchDomains] Error creating domain %v", err7)
	}
	domains, err8 := SearchDomain(tx, profilesSvc2, "ucp-common-env")
	if err8 != nil {
		t.Errorf("[TestSearchDomains] Error searching domain %v", err8)
	}
	if len(domains) != 1 {
		t.Errorf("[TestSearchDomains] Error searching domain: shooudl have 1 domain exactly")
	}
	if domains[0].Name != domainName {
		t.Errorf("[TestSearchDomains] Error searching domain: should return domain %v", domainName)
	}

	err10 := profilesSvc2.DeleteDomainByName(domainName)
	if err10 != nil {
		t.Errorf("[TestSearchDomains] Error deleting domain %v", err10)
	}
	err10 = profilesSvc2.DeleteDomainByName(domainName2)
	if err10 != nil {
		t.Errorf("[TestSearchDomains] Error deleting domain %v", err10)
	}

	err := kmsc.DeleteKey(keyArn)
	if err != nil {
		t.Errorf("Error deleting key %v", err)
	}

}
