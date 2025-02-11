// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"fmt"
	"log"
	"strings"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/kms"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/tah-core/s3"
	"tah/upt/source/tah-core/sqs"
	common "tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/feature"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	assetsSchema "tah/upt/source/ucp-common/src/model/assets-schema"
	model "tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	"tah/upt/source/ucp-common/src/utils/config"
	testutils "tah/upt/source/ucp-common/src/utils/test"
	"testing"

	"github.com/stretchr/testify/mock"
)

type setupResources struct {
	lambdaConfig         *lambda.LambdaMock
	qUrl                 string
	cognitoClient        *cognito.MockCognitoConfig
	destinationStream    kinesis.Stream
	profiles             customerprofileslcs.ICustomerProfileLowCostConfig
	solutionsConfig      awssolutions.MockConfig
	configDbClient       db.DBConfig
	errorDbClient        db.DBConfig
	glueClient           glue.Config
	privacyDbClient      db.DBConfig
	portalConfigDbClient db.DBConfig
	lcsConfigDbClient    db.DBConfig
	asyncLambdaConfig    *lambda.LambdaMock
	keyArn               string
	bucketName           string
	bucketNameMatch      string
	glueSchemaPath       string
	testPostFix          string
	testDomain           string
	s3ExportBucketName   string
	travelerTableName    string
	tx                   core.Transaction
}

func testSetup(t *testing.T, testName string) setupResources {
	testPostFix := strings.ToLower(core.GenerateUniqueId())
	testDomain := "dom_" + testPostFix
	s3ExportBucketName := "test_s3_export_bucket_" + testPostFix
	travelerTableName := "ucp_traveler_test_" + testPostFix
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	// Set up resources
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("unable to load env config: %v", err)
	}
	rootDir, err := config.AbsolutePathToProjectRootDir()
	if err != nil {
		t.Fatalf("unable to determine root dir")
	}
	glueSchemaPath := rootDir + "source/tah-common/"
	s3c := s3.InitRegion(envCfg.Region, "", "")
	kmsc := kms.Init(envCfg.Region, "", "")

	configTableName := "config-table-create-delete-async-test-" + testPostFix
	configDbClient, err := db.InitWithNewTable(configTableName, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("[%s] Could not create config db client %v", testName, err)
	}
	t.Cleanup(func() {
		err = configDbClient.DeleteTable(configTableName)
		if err != nil {
			t.Errorf("[%s] Error deleting table %v", testName, err)
		}
	})
	configDbClient.WaitForTableCreation()

	errorTableName := "error-table-create-delete-async-test-" + testPostFix
	errorDbClient := db.Init(errorTableName, "pk", "sk", "", "")
	err = errorDbClient.CreateTableWithOptions(errorTableName, "pk", "sk", db.TableOptions{
		StreamEnabled: true,
	})
	if err != nil {
		t.Fatalf("[%s] Could not create error db client %v", testName, err)
	}
	t.Cleanup(func() {
		err = errorDbClient.DeleteTable(errorTableName)
		if err != nil {
			t.Errorf("[%s] Error deleting table %v", testName, err)
		}
	})

	portalConfigTableName := "portal-table-create-delete-async-test-" + testPostFix
	portalConfigDbClient := db.Init(portalConfigTableName, "config_item", "config_item_category", "", "")
	err = portalConfigDbClient.CreateTableWithOptions(portalConfigTableName, "config_item", "config_item_category", db.TableOptions{
		StreamEnabled: true,
	})
	if err != nil {
		t.Fatalf("[%s] Could not create portal config db client %v", testName, err)
	}
	t.Cleanup(func() {
		err = portalConfigDbClient.DeleteTable(portalConfigTableName)
		if err != nil {
			t.Errorf("[%s] Error deleting table %v", testName, err)
		}
	})

	privacyResultsTableName := "privacy-results-table-create-delete-async-test-" + testPostFix
	privacyDbClient := db.Init(privacyResultsTableName, "domainName", "connectId", "", "")
	err = privacyDbClient.CreateTable(privacyResultsTableName, "domainName", "connectId")
	if err != nil {
		t.Fatalf("[%s] Could not create privacy db client %v", testName, err)
	}
	t.Cleanup(func() {
		err = privacyDbClient.DeleteTable(privacyDbClient.TableName)
		if err != nil {
			t.Errorf("[%s] error deleting table: %v", testName, err)
		}
	})

	bucketName, err := s3c.CreateRandomBucket("ucp-unit-test")
	if err != nil {
		t.Fatalf("[%s] Could not create bucket to unit test UCP %v", testName, err)
	}
	t.Cleanup(func() {
		err = s3c.DeleteBucket(bucketName)
		if err != nil {
			t.Errorf("[%s] Error deleting bucket %v", testName, err)
		}
	})

	bucketNameMatch, err := s3c.CreateRandomBucket("ucp-unit-match")
	if err != nil {
		t.Fatalf("Could not create bucket to unit test UCP Match%v", err)
	}
	t.Cleanup(func() {
		err = s3c.EmptySpecificBucket(bucketNameMatch)
		if err != nil {
			t.Errorf("[%s] Error emptying match bucket %v", testName, err)
		}
		err = s3c.DeleteBucket(bucketNameMatch)
		if err != nil {
			t.Errorf("[%s] Error deleting match bucket %v", testName, err)
		}
	})

	lambdaConfig := lambda.InitMock("ucpRetryEnvName")
	asyncLambdaConfig := lambda.InitMock("ucpAsyncEnvName")
	sqsClient := sqs.Init(envCfg.Region, "", "")
	qUrl, err := sqsClient.CreateRandom("ucp-unit-test-queue")
	if err != nil {
		t.Fatalf("[%s] Could not create queue to unit test UCP %v", testName, err)
	}
	t.Cleanup(func() { sqsClient.Delete() })
	err = sqsClient.SetPolicy("Service", "profile.amazonaws.com", []string{"SQS:*"})
	if err != nil {
		t.Fatalf("[%s] Could not Set policy on the queue %v", testName, err)
	}
	resource := []string{"arn:aws:s3:::" + bucketName, "arn:aws:s3:::" + bucketName + "/*"}
	actions := []string{"s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy"}
	principal := map[string][]string{"Service": {"appflow.amazonaws.com"}}
	err = s3c.AddPolicy(bucketName, resource, actions, principal)
	if err != nil {
		t.Fatalf("[%s] error adding bucket policy %+v", testName, err)
	}
	resource = []string{"arn:aws:s3:::" + bucketNameMatch, "arn:aws:s3:::" + bucketNameMatch + "/*"}
	principal = map[string][]string{"Service": {"profile.amazonaws.com"}}
	err = s3c.AddPolicy(bucketNameMatch, resource, actions, principal)
	if err != nil {
		t.Fatalf("[%s] error adding bucket policy %+v", testName, err)
	}
	//adding postfix to avoid propagation issue
	keyArn, err := kmsc.CreateKey("ucp-unit-test-key-" + testPostFix)
	if err != nil {
		t.Fatalf("[%s] Cound not create KMS key to unit test UCP %v", testName, err)
	}
	t.Cleanup(func() {
		err = kmsc.DeleteKey(keyArn)
		if err != nil {
			t.Errorf("[%s] Error deleting key %v", testName, err)
		}
	})

	log.Printf("Testing domain creation and deletion")
	cognitoClient := cognito.InitMock(nil, nil)
	cognitoClient.Mock.On("CreateGroup", "mock.Anything", "mock.Anything")
	cognitoClient.Mock.On("IsGroupExistsException", "mock.Anything")
	cognitoClient.Mock.On("DeleteGroup", "mock.Anything")
	glueClient := glue.Init(envCfg.Region, "glue_db_create_delete_domain"+testPostFix, "", "")
	err = glueClient.CreateDatabase(glueClient.DbName)
	if err != nil {
		t.Fatalf("[%s] Error creating database: %v", testName, err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteDatabase(glueClient.DbName)
		if err != nil {
			t.Errorf("[%s] Error deleting database %v", testName, err)
		}
	})

	schema, err := assetsSchema.LoadSchema(&tx, glueSchemaPath, commonModel.BusinessObject{Name: "traveller"})
	if err != nil {
		t.Fatalf("[%s] Error loading traveler schema: %v", testName, err)
	}
	err = glueClient.CreateTable(
		travelerTableName,
		s3ExportBucketName+"/data",
		[]glue.PartitionKey{{Name: "domainname", Type: "string"}},
		schema,
	)
	if err != nil {
		log.Printf("[%s] Table Creation failed: %v", testName, err)
		t.Fatalf("[%s] Coulnd not create glue table for testing: %v", testName, err)
	}
	t.Cleanup(func() { glueClient.DeleteTable(travelerTableName) })

	// Add to list of domains linked to the Industry Connector
	// Adding would be done manually, but we need to ensure link is deleted if domain is deleted
	configDbClient.WaitForTableCreation()  // wait if table is still being created
	errorDbClient.WaitForTableCreation()   // wait if table is still being created
	privacyDbClient.WaitForTableCreation() // wait if table is still being created
	domainList := commonModel.IndustryConnectorDomainList{
		ItemID:     common.CONFIG_DB_LINKED_CONNECTORS_PK,
		ItemType:   common.CONFIG_DB_LINKED_CONNECTORS_SK,
		DomainList: []string{testDomain},
	}
	_, err = configDbClient.Save(domainList)
	if err != nil {
		t.Fatalf("[%s] Error linking domain to Industry Connector: %v", testName, err)
	}
	// Create ACCP event stream destination
	destinationStreamName := "create-delete-async-test-" + testPostFix
	kinesisConfig, err := kinesis.InitAndCreate(destinationStreamName, envCfg.Region, "", "", core.LogLevelDebug)
	if err != nil {
		t.Fatalf("[%s] Error creating kinesis config %v", testName, err)
	}
	t.Cleanup(func() {
		err = kinesisConfig.Delete(destinationStreamName)
		if err != nil {
			t.Errorf("[%s] Error deleting stream %v", testName, err)
		}
	})
	_, err = kinesisConfig.WaitForStreamCreation(300)
	if err != nil {
		t.Fatalf("[%s] Error waiting for kinesis config to be created %v", testName, err)
	}
	destinationStream, err := kinesisConfig.Describe()
	if err != nil {
		t.Fatalf("[%s] Error describing kinesis config %v", testName, err)
	}

	lcsConfigTable, err := db.InitWithNewTable("lcs-test-"+testDomain, "pk", "sk", "", "")
	if err != nil {
		t.Fatalf("error creating lcs config table: %v", err)
	}
	t.Cleanup(func() {
		err = lcsConfigTable.DeleteTable(lcsConfigTable.TableName)
		if err != nil {
			t.Errorf("error deleting lcs config table: %v", err)
		}
	})
	err = lcsConfigTable.WaitForTableCreation()
	if err != nil {
		t.Fatalf("error waiting for lcs config table: %v", err)
	}
	mergeQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = mergeQueueClient.CreateRandom("createDeleteDomainAsyncTest")
	if err != nil {
		t.Errorf("error creating merge queue: %v", err)
	}
	t.Cleanup(func() {
		err = mergeQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting merge queue: %v", err)
		}
	})
	cpWriterQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = cpWriterQueueClient.CreateRandom("cpWriterCreateDeleteAsyncTest")
	if err != nil {
		t.Errorf("error creating cp writer queue: %v", err)
	}
	initParams := testutils.ProfileStorageParams{
		DomainName:          &testDomain,
		LcsConfigTable:      &lcsConfigTable,
		LcsKinesisStream:    kinesisConfig,
		MergeQueueClient:    &mergeQueueClient,
		CPWriterQueueClient: &cpWriterQueueClient,
	}
	profiles, err := testutils.InitProfileStorage(initParams)
	if err != nil {
		t.Fatalf("Error creating customer profiles %v", err)
	}
	solutionsConfig := awssolutions.InitMock()

	return setupResources{
		lambdaConfig:         lambdaConfig,
		qUrl:                 qUrl,
		cognitoClient:        cognitoClient,
		destinationStream:    destinationStream,
		profiles:             profiles,
		solutionsConfig:      solutionsConfig,
		configDbClient:       configDbClient,
		errorDbClient:        errorDbClient,
		glueClient:           glueClient,
		privacyDbClient:      privacyDbClient,
		portalConfigDbClient: portalConfigDbClient,
		lcsConfigDbClient:    lcsConfigTable,
		asyncLambdaConfig:    asyncLambdaConfig,
		keyArn:               keyArn,
		bucketName:           bucketName,
		bucketNameMatch:      bucketNameMatch,
		glueSchemaPath:       glueSchemaPath,
		testPostFix:          testPostFix,
		testDomain:           testDomain,
		s3ExportBucketName:   s3ExportBucketName,
		travelerTableName:    travelerTableName,
		tx:                   tx,
	}
}

func TestCreateDeleteDomainAsync(t *testing.T) {
	t.Parallel()

	testName := t.Name()
	testResources := testSetup(t, testName)
	testDomain := testResources.testDomain
	mockCognitoFunctions(testResources, testDomain)
	// Execute create domain usecase
	env := "test-env"
	createDomainUsecase := InitCreateDomain()
	ucName := createDomainUsecase.Name()
	if ucName != "CreateDomain" {
		t.Fatalf("[%s] Usecase name is incorrect, %v", testName, ucName)
	}
	featureSetVersion := feature.LATEST_FEATURE_SET_VERSION

	createDomainBody := uc.CreateDomainBody{
		KmsArn:                testResources.keyArn,
		AccpDlq:               testResources.qUrl,
		Env:                   env,
		MatchBucketName:       testResources.bucketNameMatch,
		AccpSourceBucketName:  testResources.bucketName,
		AccpDestinationStream: testResources.destinationStream.Arn,
		DomainName:            testResources.testDomain,
		HotelBookingBucket:    "hotel-booking-bucket",
		AirBookingBucket:      "air-booking-bucket",
		GuestProfileBucket:    "guest-profile-bucket",
		PaxProfileBucket:      "pax-profile-bucket",
		StayRevenueBucket:     "stay-revenue-bucket",
		ClickstreamBucket:     "clickstream-bucket",
		CSIBucket:             "customer-service-interaction-bucket",
		GlueSchemaPath:        testResources.glueSchemaPath,
		RequestedVersion:      featureSetVersion,
	}
	payload := model.AsyncInvokePayload{
		EventID:       "test-event-id",
		Usecase:       model.USECASE_CREATE_DOMAIN,
		TransactionID: "test-tx-id",
		Body:          createDomainBody,
	}

	services := model.Services{
		AccpConfig:        testResources.profiles,
		ConfigDB:          testResources.configDbClient,
		ErrorDB:           testResources.errorDbClient,
		SolutionsConfig:   testResources.solutionsConfig,
		CognitoConfig:     testResources.cognitoClient,
		GlueConfig:        testResources.glueClient,
		PrivacyResultsDB:  testResources.privacyDbClient,
		RetryLambdaConfig: testResources.lambdaConfig,
		PortalConfigDB:    testResources.portalConfigDbClient,
		AsyncLambdaConfig: testResources.asyncLambdaConfig,
		Env: map[string]string{
			"S3_EXPORT_BUCKET":              testResources.s3ExportBucketName,
			"GLUE_EXPORT_TABLE_NAME":        testResources.travelerTableName,
			"TRAVELER_S3_ROOT_PATH":         "profiles",
			"CUSTOMER_PROFILE_STORAGE_MODE": "false",
			"DYNAMO_STORAGE_MODE":           "true",
		},
	}
	err := createDomainUsecase.Execute(payload, services, testResources.tx)
	if err != nil {
		t.Fatalf("[%s] Error executing create domain usecase %v", testName, err)
	}

	// Validate domain was created
	tables, err := testResources.glueClient.ListTables()
	if err != nil {
		t.Fatalf("[%s] Error retrieving tables after creation", testName)
	}
	if len(tables) != len(common.BUSINESS_OBJECTS)+1 {
		t.Fatalf(
			"[%s] Wrong number of tables in list, %v tables in database, creation. expected %v (all object and traveler table)",
			testName,
			len(tables),
			len(common.BUSINESS_OBJECTS)+1,
		)
	}

	// validate version
	domainConfig := customerprofileslcs.DynamoLcDomainConfig{}
	err = testResources.lcsConfigDbClient.Get(testDomain, "domain_config", &domainConfig)
	if err != nil {
		t.Fatalf("[%s] Error retrieving domain config", testName)
	}
	if domainConfig.DomainVersion != featureSetVersion {
		t.Fatalf("[%s] Incorrect DomainVersion", testName)
	}
	if domainConfig.CompatibleVersion != featureSetVersion {
		t.Fatalf("[%s] Incorrect CompatibleVersion", testName)
	}

	expectedTables := map[string]string{
		"ucp_" + env + "_hotel_booking_" + testDomain:                "s3://hotel-booking-bucket/" + testDomain,
		"ucp_" + env + "_clickstream_" + testDomain:                  "s3://clickstream-bucket/" + testDomain,
		"ucp_" + env + "_guest_profile_" + testDomain:                "s3://guest-profile-bucket/" + testDomain,
		"ucp_" + env + "_pax_profile_" + testDomain:                  "s3://pax-profile-bucket/" + testDomain,
		"ucp_" + env + "_air_booking_" + testDomain:                  "s3://air-booking-bucket/" + testDomain,
		"ucp_" + env + "_hotel_stay_revenue_" + testDomain:           "s3://stay-revenue-bucket/" + testDomain,
		"ucp_" + env + "_customer_service_interaction_" + testDomain: "s3://customer-service-interaction-bucket/" + testDomain,
		testResources.travelerTableName:                              "s3://" + testResources.s3ExportBucketName + "/data",
	}

	for _, table := range tables {
		retriveTableName, ok := expectedTables[table.Name]
		if !ok {
			t.Fatalf("[%s] Table Name %v not expected", testName, table.Name)
		}
		if retriveTableName != table.Location {
			t.Fatalf("[%s] Table %v has wrong location, %v, should be %v", testName, table.Name, table.Location, retriveTableName)
		}
	}

	/////////////////////
	// Test EventSourceMapping Creation
	//////////////////////////
	streamArn, err := testResources.errorDbClient.GetStreamArn()
	if err != nil || streamArn == "" {
		t.Fatalf("[%s] Error getting stream arn %v", testName, err)
	}
	mock := services.RetryLambdaConfig.(*lambda.LambdaMock)
	if len(mock.EventSourceMappings) != 1 {
		t.Fatalf("[%s] Lambda config Mock should have 1 event source mapping but has %v", testName, len(mock.EventSourceMappings))
	} else {
		if mock.EventSourceMappings[0].StreamArn != streamArn {
			t.Fatalf("[%s] Event source mapping arn does not match, %v, %v", testName, mock.EventSourceMappings[0].StreamArn, streamArn)
		}
		if !strings.HasPrefix(mock.EventSourceMappings[0].FunctionName, "ucpRetry") {
			t.Fatalf("[%s] Event source mapping function name %v does not start with ucpRetry", testName, mock.EventSourceMappings[0].FunctionName)
		}
	}

	///////////////////////////////////////
	// Test Add Partitions to Traveler table
	////////////////////////////////
	partitions, err := services.GlueConfig.GetPartitionsFromTable(testResources.travelerTableName)
	log.Printf("[%s] Partitions %+v", testName, partitions)
	if err != nil {
		t.Fatalf("[%s] Error getting partitions from table %v", testName, err)
	}
	hasDomainPartition := false
	for _, p := range partitions {
		if p.Values[0] == testDomain {
			hasDomainPartition = true
			expected := "s3://" + testResources.s3ExportBucketName + "/profiles/domainname=" + testDomain
			if p.Location != expected {
				t.Fatalf("[%s] Partition location is incorrect, %v, should be %v", testName, p.Location, testDomain)
			}
		}
	}
	if !hasDomainPartition {
		t.Fatalf("[%s] Partition should exists on table %v for domain %s", testName, testResources.travelerTableName, testDomain)
	}

	///////////////////////////////////
	//	Test Search for Privacy Data
	///////////////////////////////////
	//	Scaffold search results for test domain and another
	_, err = services.PrivacyResultsDB.Save(uc.PrivacySearchResult{
		DomainName: testDomain,
		ConnectId:  "12345",
	})
	if err != nil {
		t.Fatalf("[%s] Error saving privacy search result %v", testName, err)
	}
	_, err = services.PrivacyResultsDB.Save(uc.PrivacySearchResult{
		DomainName: testDomain,
		ConnectId:  "23456",
	})
	if err != nil {
		t.Fatalf("[%s] Error saving privacy search result %v", testName, err)
	}
	_, err = services.PrivacyResultsDB.Save(uc.PrivacySearchResult{
		DomainName: "another-domain",
		ConnectId:  "XXXXX",
	})
	if err != nil {
		t.Fatalf("[%s] Error saving privacy search result %v", testName, err)
	}
	_, err = services.PrivacyResultsDB.Save(uc.PrivacySearchResult{
		DomainName: "another-domain",
		ConnectId:  "YYYYY",
	})
	if err != nil {
		t.Fatalf("[%s] Error saving privacy search result %v", testName, err)
	}

	// Execute delete domain usecase
	deleteDomainUsecase := InitDeleteDomain()
	ucName = deleteDomainUsecase.Name()
	if ucName != "DeleteDomain" {
		t.Fatalf("[%s] Usecase name is incorrect, %v", testName, ucName)
	}
	deleteDomainBody := uc.DeleteDomainBody{
		Env:        env,
		DomainName: testDomain,
	}
	payload = model.AsyncInvokePayload{
		EventID:       "test-event-delete",
		Usecase:       model.USECASE_DELETE_DOMAIN,
		TransactionID: "test-tx-id-create",
		Body:          deleteDomainBody,
	}
	services = model.Services{
		AccpConfig:       testResources.profiles,
		ConfigDB:         testResources.configDbClient,
		SolutionsConfig:  testResources.solutionsConfig,
		CognitoConfig:    testResources.cognitoClient,
		GlueConfig:       testResources.glueClient,
		PrivacyResultsDB: testResources.privacyDbClient,
		PortalConfigDB:   testResources.portalConfigDbClient,
		Env: map[string]string{
			"GLUE_EXPORT_TABLE_NAME": testResources.travelerTableName,
			"TRAVELER_S3_ROOT_PATH":  "profiles",
		},
	}
	err = deleteDomainUsecase.Execute(payload, services, testResources.tx)
	if err != nil {
		t.Fatalf("[%s] Error executing delete domain usecase %v", testName, err)
	}

	// Validate domain was deleted
	log.Printf("Verifying all tables are gone")
	tablesAfterDeletion, err := testResources.glueClient.ListTables()
	if err != nil {
		t.Fatalf("[%s] Error retrieving tables after deletion", testName)
	}
	if len(tablesAfterDeletion) != 1 {
		t.Fatalf(
			"[%s] Wrong number of tables in list, %v tables in database after deletion. should be %v (traveler table)",
			testName,
			len(tablesAfterDeletion),
			1,
		)
	}
	updatedList := commonModel.IndustryConnectorDomainList{}
	err = services.ConfigDB.Get(common.CONFIG_DB_LINKED_CONNECTORS_PK, common.CONFIG_DB_LINKED_CONNECTORS_SK, &updatedList)
	if err != nil {
		t.Fatalf("[%s] Error retrieving domain list from config db: %v", testName, err)
	}
	if len(updatedList.DomainList) != 0 {
		t.Fatalf("[%s] Error retrieving domain list from config db, expected empty list but got %v", testName, len(updatedList.DomainList))
	}

	///////////////////////////////////////
	// Test Remove Partitions from Traveler table
	////////////////////////////////
	partitions, err = services.GlueConfig.GetPartitionsFromTable(testResources.travelerTableName)
	if err != nil {
		t.Fatalf("[%s] Error getting partitions from table %v", testName, err)
	}
	hasDomainPartition = false
	for _, p := range partitions {
		if p.Values[0] == testDomain {
			hasDomainPartition = true
		}
	}
	if hasDomainPartition {
		t.Fatalf(
			"[%s] Partition should have been removed from table %v after domain %s deletion",
			testName,
			testResources.travelerTableName,
			testDomain,
		)
	}

	/////////////////////////////////////////////////////////////
	//	Test Privacy Results Table has no results from testDomain
	/////////////////////////////////////////////////////////////
	searchResults := []uc.PrivacySearchResult{}
	err = services.PrivacyResultsDB.FindByPk(testDomain, &searchResults)
	if err != nil {
		t.Fatalf("[%s] Error querying privacy results from db %v", testName, err)
	}
	if len(searchResults) != 0 {
		t.Fatalf("[%s] Error retrieving privacy results from db, expected empty list but got %v", testName, len(searchResults))
	}
}

// This test creates domain without specifying the RequestedVersion. In such case, the version attributes should be assigned the latest version value
func TestCreateDomainAsyncWithoutVersion(t *testing.T) {
	t.Parallel()

	testName := t.Name()
	testResources := testSetup(t, testName)
	testDomain := testResources.testDomain
	mockCognitoFunctions(testResources, testDomain)
	// Execute create domain usecase
	env := "test-env"
	createDomainUsecase := InitCreateDomain()
	ucName := createDomainUsecase.Name()
	if ucName != "CreateDomain" {
		t.Fatalf("[%s] Usecase name is incorrect, %v", testName, ucName)
	}

	// RequestedVersion is skipped in the request
	createDomainBody := uc.CreateDomainBody{
		KmsArn:                testResources.keyArn,
		AccpDlq:               testResources.qUrl,
		Env:                   env,
		MatchBucketName:       testResources.bucketNameMatch,
		AccpSourceBucketName:  testResources.bucketName,
		AccpDestinationStream: testResources.destinationStream.Arn,
		DomainName:            testResources.testDomain,
		HotelBookingBucket:    "hotel-booking-bucket",
		AirBookingBucket:      "air-booking-bucket",
		GuestProfileBucket:    "guest-profile-bucket",
		PaxProfileBucket:      "pax-profile-bucket",
		StayRevenueBucket:     "stay-revenue-bucket",
		ClickstreamBucket:     "clickstream-bucket",
		CSIBucket:             "customer-service-interaction-bucket",
		GlueSchemaPath:        testResources.glueSchemaPath,
	}
	payload := model.AsyncInvokePayload{
		EventID:       "test-event-id",
		Usecase:       model.USECASE_CREATE_DOMAIN,
		TransactionID: "test-tx-id",
		Body:          createDomainBody,
	}

	services := model.Services{
		AccpConfig:        testResources.profiles,
		ConfigDB:          testResources.configDbClient,
		ErrorDB:           testResources.errorDbClient,
		SolutionsConfig:   testResources.solutionsConfig,
		CognitoConfig:     testResources.cognitoClient,
		GlueConfig:        testResources.glueClient,
		PrivacyResultsDB:  testResources.privacyDbClient,
		RetryLambdaConfig: testResources.lambdaConfig,
		PortalConfigDB:    testResources.portalConfigDbClient,
		AsyncLambdaConfig: testResources.asyncLambdaConfig,
		Env: map[string]string{
			"S3_EXPORT_BUCKET":              testResources.s3ExportBucketName,
			"GLUE_EXPORT_TABLE_NAME":        testResources.travelerTableName,
			"TRAVELER_S3_ROOT_PATH":         "profiles",
			"CUSTOMER_PROFILE_STORAGE_MODE": "false",
			"DYNAMO_STORAGE_MODE":           "true",
		},
	}
	err := createDomainUsecase.Execute(payload, services, testResources.tx)
	if err != nil {
		t.Fatalf("[%s] Error executing create domain usecase %v", testName, err)
	}

	// validate version
	domainConfig := customerprofileslcs.DynamoLcDomainConfig{}
	err = testResources.lcsConfigDbClient.Get(testDomain, "domain_config", &domainConfig)
	if err != nil {
		t.Fatalf("[%s] Error retrieving domain config", testName)
	}
	if domainConfig.DomainVersion != feature.LATEST_FEATURE_SET_VERSION {
		t.Fatalf("[%s] Incorrect DomainVersion", testName)
	}
	if domainConfig.CompatibleVersion != feature.LATEST_FEATURE_SET_VERSION {
		t.Fatalf("[%s] Incorrect CompatibleVersion", testName)
	}

	// Execute delete domain usecase
	deleteDomainUsecase := InitDeleteDomain()

	deleteDomainBody := uc.DeleteDomainBody{
		Env:        env,
		DomainName: testDomain,
	}
	payload = model.AsyncInvokePayload{
		EventID:       "test-event-delete",
		Usecase:       model.USECASE_DELETE_DOMAIN,
		TransactionID: "test-tx-id-create",
		Body:          deleteDomainBody,
	}
	services = model.Services{
		AccpConfig:       testResources.profiles,
		ConfigDB:         testResources.configDbClient,
		SolutionsConfig:  testResources.solutionsConfig,
		CognitoConfig:    testResources.cognitoClient,
		GlueConfig:       testResources.glueClient,
		PrivacyResultsDB: testResources.privacyDbClient,
		PortalConfigDB:   testResources.portalConfigDbClient,
		Env: map[string]string{
			"GLUE_EXPORT_TABLE_NAME": testResources.travelerTableName,
			"TRAVELER_S3_ROOT_PATH":  "profiles",
		},
	}
	err = deleteDomainUsecase.Execute(payload, services, testResources.tx)
	if err != nil {
		t.Fatalf("[%s] Error executing delete domain usecase %v", testName, err)
	}

	// Validate domain was deleted
	updatedList := commonModel.IndustryConnectorDomainList{}
	err = services.ConfigDB.Get(common.CONFIG_DB_LINKED_CONNECTORS_PK, common.CONFIG_DB_LINKED_CONNECTORS_SK, &updatedList)
	if err != nil {
		t.Fatalf("[%s] Error retrieving domain list from config db: %v", testName, err)
	}
	if len(updatedList.DomainList) != 0 {
		t.Fatalf("[%s] Error retrieving domain list from config db, expected empty list but got %v", testName, len(updatedList.DomainList))
	}

}

func TestDomainCreationFail(t *testing.T) {
	t.Parallel()

	testResources := testSetup(t, t.Name())
	testDomain := testResources.testDomain

	// Execute create domain usecase
	env := "test-env" + testResources.testPostFix
	createDomainUsecase := InitCreateDomain()

	createDomainBody := uc.CreateDomainBody{
		KmsArn:                testResources.keyArn,
		AccpDlq:               testResources.qUrl,
		Env:                   env,
		MatchBucketName:       testResources.bucketNameMatch,
		AccpSourceBucketName:  testResources.bucketName,
		AccpDestinationStream: testResources.destinationStream.Arn,
		DomainName:            testDomain,
		HotelBookingBucket:    "hotel-booking-bucket",
		AirBookingBucket:      "air-booking-bucket",
		GuestProfileBucket:    "guest-profile-bucket",
		PaxProfileBucket:      "pax-profile-bucket",
		StayRevenueBucket:     "stay-revenue-bucket",
		ClickstreamBucket:     "clickstream-bucket",
		CSIBucket:             "customer-service-interaction-bucket",
		// GlueSchemaPath:      commenting to fail domain creation
	}
	payload := model.AsyncInvokePayload{
		EventID:       "test-event-id",
		Usecase:       model.USECASE_CREATE_DOMAIN,
		TransactionID: "test-tx-id",
		Body:          createDomainBody,
	}

	services := model.Services{
		AccpConfig:        testResources.profiles,
		ConfigDB:          testResources.configDbClient,
		ErrorDB:           testResources.errorDbClient,
		SolutionsConfig:   testResources.solutionsConfig,
		CognitoConfig:     testResources.cognitoClient,
		GlueConfig:        testResources.glueClient,
		PrivacyResultsDB:  testResources.privacyDbClient,
		RetryLambdaConfig: testResources.lambdaConfig,
		PortalConfigDB:    testResources.portalConfigDbClient,
		AsyncLambdaConfig: testResources.asyncLambdaConfig,
		Env: map[string]string{
			"S3_EXPORT_BUCKET":              testResources.s3ExportBucketName,
			"GLUE_EXPORT_TABLE_NAME":        testResources.travelerTableName,
			"TRAVELER_S3_ROOT_PATH":         "profiles",
			"CUSTOMER_PROFILE_STORAGE_MODE": "false",
			"DYNAMO_STORAGE_MODE":           "true",
		},
	}

	mockCognitoFunctions(testResources, testDomain)
	err := createDomainUsecase.Execute(payload, services, testResources.tx)
	if err == nil {
		t.Fatalf("[%s] Domain creation did not fail", t.Name())
	}

	// assert
	groupNamePrefix := buildAppAccessGroupPrefix(testDomain)
	testResources.cognitoClient.AssertNumberOfCalls(t, "CreateGroup", 6)
	testResources.cognitoClient.AssertNumberOfCalls(t, "DeleteGroup", 6)
	testResources.cognitoClient.AssertCalled(t, "CreateGroup", buildDataAccessGroupName(testDomain), "*/*")
	testResources.cognitoClient.AssertCalled(t, "DeleteGroup", buildDataAccessGroupName(testDomain))
	for roleName, rolePermission := range appAccessRoleMap {
		testResources.cognitoClient.AssertCalled(
			t,
			"CreateGroup",
			groupNamePrefix+roleName+fmt.Sprintf("%x", common.AppPermission(rolePermission)),
			"",
		)
		testResources.cognitoClient.AssertCalled(
			t,
			"DeleteGroup",
			groupNamePrefix+roleName+fmt.Sprintf("%x", common.AppPermission(rolePermission)),
		)
	}
	testResources.cognitoClient.AssertExpectations(t)
}

func mockCognitoFunctions(testResources setupResources, testDomain string) {
	groupNamePrefix := buildAppAccessGroupPrefix(testDomain)
	testResources.cognitoClient.On("CreateGroup", buildDataAccessGroupName(testDomain), "*/*")
	testResources.cognitoClient.On("DeleteGroup", buildDataAccessGroupName(testDomain))
	for roleName, rolePermission := range appAccessRoleMap {
		testResources.cognitoClient.On("CreateGroup", groupNamePrefix+roleName+fmt.Sprintf("%x", common.AppPermission(rolePermission)), "")
		testResources.cognitoClient.On("DeleteGroup", groupNamePrefix+roleName+fmt.Sprintf("%x", common.AppPermission(rolePermission)))
	}
	testResources.cognitoClient.On("IsGroupExistsException", mock.Anything)
}
