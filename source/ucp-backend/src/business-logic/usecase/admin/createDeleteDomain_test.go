package admin

import (
	"log"
	"tah/core/appregistry"
	"tah/core/core"
	"tah/core/customerprofiles"
	"tah/core/db"
	"tah/core/glue"
	"tah/core/iam"
	"tah/core/kms"
	"tah/core/s3"
	model "tah/ucp/src/business-logic/model/common"
	testutils "tah/ucp/src/business-logic/testutils"
	"tah/ucp/src/business-logic/usecase/registry"
	"testing"
	"time"
)

var UCP_REGION = testutils.GetTestRegion()

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
	//adding timestamp to avoid propagation issue
	keyArn, err1 := kmsc.CreateKey("ucp-unit-test-key")
	if err1 != nil {
		t.Errorf("Cound not create KMS key to unit test UCP %v", err1)
	}
	log.Printf("Testing domain creation and deletion")
	testDomain := "ucp-component-test-domain-" + time.Now().Format("15-04-05")
	req := model.RequestWrapper{
		Domain: model.Domain{Name: testDomain},
	}
	//init registry
	tx := core.NewTransaction("ucp_test", "")
	appregistryClient := appregistry.Init(UCP_REGION)
	iamClient := iam.Init()
	glueClient := glue.Init(UCP_REGION, "test_glue_db")
	profiles := customerprofiles.InitWithDomain(testDomain, UCP_REGION)
	dbConfig := db.Init("TEST_TABLE", "TEST_PK", "TEST_SK")

	reg := registry.NewRegistry(UCP_REGION, &appregistryClient, &iamClient, &glueClient, &profiles, &dbConfig)
	reg.AddEnv("KMS_KEY_PROFILE_DOMAIN", keyArn)
	reg.AddEnv("LAMBDA_ENV", "dev_test")
	reg.AddEnv("CONNECT_PROFILE_SOURCE_BUCKET", bucketName)

	createUc := NewCreateDomain()
	createUc.SetRegistry(&reg)
	createUc.SetTx(&tx)
	deleteUc := NewDeleteDomain()
	deleteUc.SetRegistry(&reg)
	deleteUc.SetTx(&tx)
	log.Printf("Testing domain creation")
	_, err := createUc.Run(req)
	if err != nil {
		t.Errorf("Error creating UCP domain: %v", err)
	}
	log.Printf("Testing domain deletion")
	_, err = deleteUc.Run(req)
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
