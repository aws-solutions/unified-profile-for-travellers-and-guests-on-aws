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
	"tah/core/sqs"

	common "tah/ucp-common/src/constant/admin"
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
		t.Errorf("Could not create bucket to unit test UCP %v", err0)
	}
	sqsClient := sqs.Init(UCP_REGION)
	qUrl, err5 := sqsClient.CreateRandom("ucp-unit-test-queue")
	if err5 != nil {
		t.Errorf("Could not create queue to unit test UCP %v", err5)
	}
	err5 = sqsClient.SetPolicy("Service", "profile.amazonaws.com", []string{"SQS:*"})
	if err5 != nil {
		t.Errorf("Could not Set policy on the queue %v", err5)
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
	glueClient := glue.Init(UCP_REGION, "glue_db_create_delete_domain")
	err0 = glueClient.CreateDatabase(glueClient.DbName)
	if err0 != nil {
		t.Errorf("Error creating database: %v", err0)
	}

	profiles := customerprofiles.InitWithDomain(testDomain, UCP_REGION)
	dbConfig := db.Init("TEST_TABLE", "TEST_PK", "TEST_SK")

	reg := registry.NewRegistry(UCP_REGION, registry.ServiceHandlers{AppRegistry: &appregistryClient, Iam: &iamClient, Glue: &glueClient, Accp: &profiles, ErrorDB: &dbConfig})
	reg.AddEnv("KMS_KEY_PROFILE_DOMAIN", keyArn)
	reg.AddEnv("LAMBDA_ENV", "dev_test")
	reg.AddEnv("CONNECT_PROFILE_SOURCE_BUCKET", bucketName)
	reg.AddEnv("ACCP_DOMAIN_DLQ", qUrl)

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
	log.Printf("Verifying all Glue Tables created")

	tables, err := glueClient.ListTables()
	if err != nil {
		t.Errorf("Error retrieving tables after creation")
	}
	if len(tables) != len(common.BUSINESS_OBJECTS) {
		t.Errorf("Wrong number of tables in list, %v tables in database, creation", len(tables))
	}
	log.Printf("Testing domain deletions")
	//deleting teh doomain will delete the SQS as well
	_, err = deleteUc.Run(req)
	if err != nil {
		t.Errorf("Error deleting UCP domain: %v", err)
	}
	log.Printf("Verifying all tables are gone")
	tablesAfterDeletion, err := glueClient.ListTables()
	if err != nil {
		t.Errorf("Error retrieving tables after deletion")
	}
	if len(tablesAfterDeletion) != 0 {
		t.Errorf("Wrong number of tables in list, %v tables in database, deletion", len(tablesAfterDeletion))
	}
	err = s3c.DeleteBucket(bucketName)
	if err != nil {
		t.Errorf("Error deleting bucket %v", err)
	}
	log.Printf("Deleting Databases")
	err = glueClient.DeleteDatabase(glueClient.DbName)
	if err != nil {
		t.Errorf("Error deleting database %v", err)
	}
	err = kmsc.DeleteKey(keyArn)
	if err != nil {
		t.Errorf("Error deleting key %v", err)
	}
}
