// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"fmt"
	"os"
	"strings"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/athena"
	"tah/upt/source/tah-core/awssolutions"
	cw "tah/upt/source/tah-core/cloudwatch"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/s3"
	"tah/upt/source/tah-core/sqs"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	model "tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	"tah/upt/source/ucp-common/src/utils/config"
	testutils "tah/upt/source/ucp-common/src/utils/test"
	"testing"
)

func TestSearchOutputBucket(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}

	rootDir, err := config.AbsolutePathToProjectRootDir()
	if err != nil {
		t.Fatalf("unable to determine root dir")
	}
	glueSchemaPath := rootDir + "source/tah-common/"
	testName := testutils.GetTestName(TestSearchOutputBucket)
	testPostFix := strings.ToLower(core.GenerateUniqueId())

	//	ARRANGE resources
	//	Create S3 Source Bucket
	sourceBucket, err := s3.InitWithRandBucket("test-source-bucket", "", envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] Error creating s3 source bucket: %v", testName, err)
	}
	t.Cleanup(func() { sourceBucket.EmptyAndDelete() })

	//	Create S3 Athena Output Bucket
	outputBucket, err := s3.InitWithRandBucket("test-athena-output-bucket", "", envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] Error creating s3 output bucket: %v", testName, err)
	}
	t.Cleanup(func() { outputBucket.EmptyAndDelete() })

	//	Create Glue Tables
	glueDbName := "test_glue_database-" + testPostFix
	glueTableName := "test_glue_table-" + testPostFix
	glueCfg := glue.Init(envCfg.Region, glueDbName, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	schemaFilePath := glueSchemaPath + "/" + "tah-common-glue-schemas/traveller.glue.json"
	schemaBytes, err := os.ReadFile(schemaFilePath)
	if err != nil {
		t.Fatalf("[%s] error reading glue schema file: %v", testName, err)
	}
	glueSchema, err := glue.ParseSchema(string(schemaBytes))
	if err != nil {
		t.Fatalf("[%s] error parsing glue schema: %v", testName, err)
	}
	err = glueCfg.CreateDatabase(glueDbName)
	if err != nil {
		t.Fatalf("[%s] error creating glue database: %v", testName, err)
	}
	t.Cleanup(func() { glueCfg.DeleteDatabase(glueDbName) })

	athenaCfg := athena.Init(glueDbName, glueTableName, "test_athena_workgroup_"+testPostFix, envCfg.Region, outputBucket.Bucket, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION, core.LogLevelDebug)
	err = athenaCfg.CreateWorkGroup(athenaCfg.Workgroup)
	if err != nil {
		t.Fatalf("[%s] error creating athena workgroup: %v", testName, err)
	}
	t.Cleanup(func() {
		athenaCfg.DeleteWorkGroup(athenaCfg.Workgroup)
	})
	parameterizedQuery := fmt.Sprintf("SELECT \"$path\" FROM \"%s\".\"%s\" where connectid = ? and domainname = ? group by \"$path\"", athenaCfg.DBName, athenaCfg.TableName)
	const GET_S3_PATHS_PREPARED_STATEMENT_NAME = "get_s3_paths"
	err = athenaCfg.CreatePreparedStatement(parameterizedQuery, GET_S3_PATHS_PREPARED_STATEMENT_NAME)
	if err != nil {
		t.Fatalf("[%s] error creating prepared statement: %v", testName, err)
	}
	t.Cleanup(func() {
		athenaCfg.DeletePreparedStatement(GET_S3_PATHS_PREPARED_STATEMENT_NAME)
	})
	const GET_S3PATH_MAPPING_PREPARED_STATEMENT_NAME = "get_s3_mappings"
	queryStatement := fmt.Sprintf(`SELECT "$path", connectid FROM UNNEST(CAST(JSON_PARSE(?) as ARRAY<VARCHAR>)) AS t(cId) JOIN "%s"."%s" on cId = connectid WHERE domainname = ? GROUP BY "$path", connectid`, athenaCfg.DBName, athenaCfg.TableName)
	err = athenaCfg.CreatePreparedStatement(queryStatement, GET_S3PATH_MAPPING_PREPARED_STATEMENT_NAME)
	if err != nil {
		t.Fatalf("[%s] error creating prepared statement: %v", testName, err)
	}
	t.Cleanup(func() { athenaCfg.DeletePreparedStatement(GET_S3PATH_MAPPING_PREPARED_STATEMENT_NAME) })

	partitions := []glue.PartitionKey{}
	partitions = append(partitions, glue.PartitionKey{
		Name: "domainname",
		Type: "string",
	})
	err = glueCfg.CreateParquetTable(glueTableName, sourceBucket.Bucket, partitions, glueSchema)
	if err != nil {
		t.Fatalf("[%s] error creating parquet table: %v", testName, err)
	}
	t.Cleanup(func() { glueCfg.DeleteTable(glueTableName) })

	domainName := "dev"
	glueCfg.AddParquetPartitionsToTable(glueTableName, []glue.Partition{{
		Values:   []string{domainName},
		Location: "s3://" + sourceBucket.Bucket + "/domainname=" + domainName,
	}})

	//	Create Privacy Search Results Dynamo Table
	privacySearchResultsTableName := "privacy-search-results-table-create-privacy-search-async-test-" + testPostFix
	privacyDbClient, err := db.InitWithNewTable(privacySearchResultsTableName, "domainName", "connectId", "", "")
	if err != nil {
		t.Fatalf("Could not create privacy db client %v", err)
	}
	t.Cleanup(func() { privacyDbClient.DeleteTable(privacySearchResultsTableName) })
	err = privacyDbClient.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] error waiting for table creation: %v", t.Name(), err)
	}

	//	Create Match Dynamo Table
	matchDbTableName := "match-table-create-privacy-search-async-test-" + testPostFix
	matchDbClient, err := db.InitWithNewTable(matchDbTableName, "domain_sourceProfileId", "match_targetProfileId", "", "")
	if err != nil {
		t.Fatalf("Could not create match db client %v", err)
	}
	t.Cleanup(func() { matchDbClient.DeleteTable(matchDbTableName) })
	err = matchDbClient.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[%s] error waiting for table creation: %v", t.Name(), err)
	}

	//	Create SQS Queue
	s3ExciseSqsConfig := sqs.Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	s3ExciseSqsConfig.CreateRandom(testPostFix)
	t.Cleanup(func() { s3ExciseSqsConfig.Delete() })

	//	Create Customer Profiles Mock
	connectIds := []string{"420bdbf6f581401695e60fa650958e97"}
	profileId := core.GenerateUniqueId()
	profiles := customerprofiles.InitMock(&customerprofiles.Domain{
		Name: domainName,
	}, nil, &profilemodel.Profile{}, &[]profilemodel.Profile{
		{
			ProfileId: connectIds[0],
			Attributes: map[string]string{
				"profile_id": profileId,
			},
		},
	}, nil)
	solutionsConfig := awssolutions.InitMock()

	cwConfig := cw.Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	gdprLogGroupName := "gdpr-log-group-" + testPostFix
	if err = cwConfig.CreateLogGroup(gdprLogGroupName); err != nil {
		t.Fatalf("Error creating log group: %v", err)
	}
	t.Cleanup(func() { cwConfig.DeleteLogGroup(gdprLogGroupName) })

	//	Arrange UseCase
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)

	createPrivacySearchBody := uc.CreatePrivacySearchBody{
		ConnectIds: connectIds,
		DomainName: domainName,
	}
	payload := model.AsyncInvokePayload{
		EventID:       "test-event-id",
		Usecase:       model.USECASE_CREATE_PRIVACY_SEARCH,
		TransactionID: "test-transaction-id",
		Body:          createPrivacySearchBody,
	}
	services := model.Services{
		AccpConfig:        profiles,
		SolutionsConfig:   solutionsConfig,
		PrivacyResultsDB:  privacyDbClient,
		MatchDB:           matchDbClient,
		GlueConfig:        glueCfg,
		S3ExciseSqsConfig: s3ExciseSqsConfig,
		Env: map[string]string{
			"GLUE_DB":                                    glueCfg.DbName,
			"GLUE_EXPORT_TABLE_NAME":                     glueTableName,
			"ATHENA_WORKGROUP_NAME":                      athenaCfg.Workgroup,
			"REGION":                                     envCfg.Region,
			"ATHENA_OUTPUT_BUCKET":                       "s3://" + outputBucket.Bucket,
			"METRICS_SOLUTION_ID":                        core.TEST_SOLUTION_ID,
			"METRICS_SOLUTION_VERSION":                   core.TEST_SOLUTION_VERSION,
			"GET_S3_PATHS_PREPARED_STATEMENT_NAME":       GET_S3_PATHS_PREPARED_STATEMENT_NAME + testPostFix,
			"GET_S3PATH_MAPPING_PREPARED_STATEMENT_NAME": GET_S3PATH_MAPPING_PREPARED_STATEMENT_NAME,
			"GDPR_PURGE_LOG_GROUP_NAME":                  gdprLogGroupName,
		},
	}
	createPrivacySearchUsecase := InitCreatePrivacySearch()
	ucName := createPrivacySearchUsecase.Name()
	if ucName != "CreatePrivacySearch" {
		t.Fatalf("[%s] UseCase Name is not correct: %v", testName, ucName)
	}

	//	Arrange - Upload Test File to Source Bucket
	filePath := "../../../../test_data/change_proc_output/test_change_proc_output.parquet"
	err = sourceBucket.UploadFile("domainname=dev/change_proc.parquet", filePath)
	if err != nil {
		t.Fatalf("[%s] Error uploading file: %v", testName, err)
	}
	privacyDbClient.WaitForTableCreation()
	matchDbClient.WaitForTableCreation()

	matchingConnectId := core.GenerateUniqueId()
	_, err = matchDbClient.Save(commonModel.MatchPair{
		Pk: fmt.Sprintf("%s_%s", domainName, connectIds[0]),
		Sk: fmt.Sprintf("match_%s", matchingConnectId),
	})
	if err != nil {
		t.Fatalf("[%s] Error saving match pair: %v", testName, err)
	}

	//	Act - Test Execute UseCase with incorrect prepared statement name
	err = createPrivacySearchUsecase.Execute(payload, services, tx)
	if err == nil {
		t.Fatalf("[%s] Expected error executing usecase", testName)
	}

	//	Set prepared statement to proper name
	services.Env["GET_S3_PATHS_PREPARED_STATEMENT_NAME"] = GET_S3_PATHS_PREPARED_STATEMENT_NAME
	err = createPrivacySearchUsecase.Execute(payload, services, tx)
	if err != nil {
		t.Fatalf("[%s] Error executing usecase: %v", testName, err)
	}

	//	Assert
	searchResult := uc.PrivacySearchResult{}
	err = privacyDbClient.Get(domainName, connectIds[0], &searchResult)
	if err != nil {
		t.Fatalf("[%s] Error getting privacy search result: %v", testName, err)
	}
	if searchResult.ConnectId != connectIds[0] {
		t.Fatalf("[%s] Privacy Search Result ConnectId is not correct. Expected: %s, Actual: %s", testName, connectIds[0], searchResult.ConnectId)
	}
	if len(searchResult.LocationResults) == 0 {
		t.Fatalf("[%s] Privacy Search Result LocationResults is empty", testName)
	}

	expectedSourcePath := fmt.Sprintf("s3://%s/domainname=dev/change_proc.parquet", sourceBucket.Bucket)
	if len(searchResult.LocationResults[uc.S3LocationType]) == 0 {
		t.Fatalf("[%s] Privacy Search Result S3SearchResults is empty", testName)
	}
	if searchResult.LocationResults[uc.S3LocationType][0] != expectedSourcePath {
		t.Fatalf("[%s] Privacy Search Result S3SearchResults is not correct. Expected: %s, Actual: %s", testName, expectedSourcePath, searchResult.LocationResults[uc.S3LocationType][0])
	}
	if len(searchResult.LocationResults[uc.ProfileStorageLocationType]) == 0 {
		t.Fatalf("[%s] Privacy Search Result ProfileSearchResults is empty", testName)
	}
	if searchResult.LocationResults[uc.ProfileStorageLocationType][0] != "profileId: "+connectIds[0] {
		t.Fatalf("[%s] Privacy Search Result ProfileSearchResults is not correct. Expected: %s, Actual: %s", testName, "profileId: "+connectIds[0], searchResult.LocationResults[uc.ProfileStorageLocationType][0])
	}
	if len(searchResult.LocationResults[uc.MatchDbLocationType]) == 0 {
		t.Fatalf("[%s] Privacy Search Result MatchSearchResults is empty", testName)
	}
	expectedMatchDbResult := fmt.Sprintf("%s://%s_%s/match_%s", matchDbClient.TableName, domainName, connectIds[0], matchingConnectId)
	if expectedMatchDbResult != searchResult.LocationResults[uc.MatchDbLocationType][0] {
		t.Fatalf("[%s] Privacy Search Result MatchSearchResults is not correct. Expected: %s, Actual: %s", testName, expectedMatchDbResult, searchResult.LocationResults[uc.MatchDbLocationType][0])
	}

	//	Act - Purge Profile Data
	purgeProfileDataBody := uc.PurgeProfileDataBody{
		ConnectIds:     connectIds,
		DomainName:     domainName,
		AgentCognitoId: core.GenerateUniqueId(),
	}
	payload = model.AsyncInvokePayload{
		EventID:       "test-event-" + testPostFix,
		Usecase:       model.USECASE_PURGE_PROFILE_DATA,
		TransactionID: "test-transaction-" + testPostFix,
		Body:          purgeProfileDataBody,
	}

	purgeProfileDataUsecase := InitPurgeProfileData()
	ucName = purgeProfileDataUsecase.Name()
	if ucName != "PurgeProfileData" {
		t.Fatalf("[%s] Unexpected usecase name: %s", testName, ucName)
	}

	//	Act - Test Execute UseCase
	err = purgeProfileDataUsecase.Execute(payload, services, tx)
	if err != nil {
		t.Fatalf("[%s] Error while executing usecase: %v", testName, err)
	}

	//	Mock the results of the S3ExciseQueueProcessor
	//	This usecase should not return an error from putting messages onto the S3ExciseQueue
	//	The S3ExciseQueueProcessor lambda is tested separately
	sourceBucket.Delete("", "domainname=dev/change_proc.parquet")

	//	Act - Test Execute UseCase again
	err = createPrivacySearchUsecase.Execute(payload, services, tx)
	if err != nil {
		t.Fatalf("[%s] Expected error executing usecase: %v", testName, err)
	}

	//	Assert
	searchResult = uc.PrivacySearchResult{}
	err = privacyDbClient.Get(domainName, connectIds[0], &searchResult)
	if err != nil {
		t.Fatalf("[%s] Error getting privacy search result: %v", testName, err)
	}
	if searchResult.ConnectId != connectIds[0] {
		t.Fatalf("[%s] Privacy Search Result ConnectId is not correct. Expected: %s, Actual: %s", testName, connectIds[0], searchResult.ConnectId)
	}

	resultsLength := len(searchResult.LocationResults[uc.S3LocationType])
	t.Logf("[%s] s3 location type results: %d", testName, resultsLength)
	if resultsLength != 0 {
		t.Fatalf("[%s] Privacy Search Result S3SearchResults is not empty", testName)
	}
	resultsLength = len(searchResult.LocationResults[uc.ProfileStorageLocationType])
	if resultsLength != 0 {
		t.Fatalf("[%s] Privacy Search Result ProfileSearchResults is not empty; expected 0, got: %d", testName, resultsLength)
	}
	resultsLength = len(searchResult.LocationResults[uc.MatchDbLocationType])
	if resultsLength != 0 {
		t.Fatalf("[%s] Privacy Search Result MatchSearchResults is not empty; expected 0, get: %d", testName, resultsLength)
	}
}
