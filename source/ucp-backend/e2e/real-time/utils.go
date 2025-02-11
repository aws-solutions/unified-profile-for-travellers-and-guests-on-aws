// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"testing"
	"time"

	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/db"
	eb "tah/upt/source/tah-core/eventbridge"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/tah-core/s3"
	"tah/upt/source/tah-core/sqs"
	constants "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/config"
	testutils "tah/upt/source/ucp-common/src/utils/test"
)

type TestConfig struct {
	Files    []string
	Profiles map[string][]testutils.ExpectedField
}

// Postman collection folders
const (
	AUTH                              string = "Auth"
	CREATE_DOMAIN                     string = "CreateDomain"
	MANAGE_DOMAIN                     string = "ManageDomain"
	SEARCH_PROFILES                   string = "SearchProfiles"
	MERGE_PROFILES                    string = "MergeProfiles"
	DELETE_DOMAIN                     string = "DeleteDomain"
	POLL_ASYNC_EVENT                  string = "PollAsyncEvent"
	ERROR_MANAGEMENT                  string = "ErrorManagement"
	CREATE_PRIVACY_SEARCH             string = "CreatePrivacySearch"
	LIST_PRIVACY_SEARCHES             string = "ListPrivacySearches"
	PRIVACY_PURGE                     string = "PrivacyPurge"
	LIST_PRIVACY_SEARCHES_AFTER_PURGE string = "ListPrivacySearchesAfterPurge"
)

type Config struct {
	APIGatewayURL  string
	TokenEndpoint  string
	ClientID       string
	RefreshToken   string
	AccessToken    string
	DomainName     string
	EnvName        string
	AdminUserName  string
	AdminGroupName string
}

type Services struct {
	Cognito        *cognito.CognitoConfig
	ProfileStorage customerprofiles.ICustomerProfileLowCostConfig
	IngestorStream kinesis.IConfig
	EventBridge    *eb.Config
	SQS            *sqs.Config
	OutputBucket   *s3.S3Config
	BackupBucket   *s3.S3Config
	RetryLambda    lambda.IConfig
	ErrorTable     *db.DBConfig
	Glue           glue.IConfig
}

func InitServices() (Services, error) {
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		return Services{}, err
	}

	dynamoClient := db.Init(infraConfig.LcsConfigTableName, infraConfig.LcsConfigTablePk, infraConfig.LcsConfigTableSk, "", "")
	kinesisClient := kinesis.Init(infraConfig.ChangeProcessorStreamName, envConfig.Region, "", "", core.LogLevelDebug)
	mergeQueueClient := sqs.InitWithQueueUrl(infraConfig.MergeQueueUrl, envConfig.Region, "", "")
	cpWriterQueueClient := sqs.InitWithQueueUrl(infraConfig.CPWriterQueueUrl, envConfig.Region, "", "")

	initParams := testutils.ProfileStorageParams{
		LcsConfigTable:      &dynamoClient,
		LcsKinesisStream:    kinesisClient,
		MergeQueueClient:    &mergeQueueClient,
		CPWriterQueueClient: &cpWriterQueueClient,
	}
	profileStorage, err := testutils.InitProfileStorage(initParams)
	ebCfg := eb.InitWithBus(envConfig.Region, infraConfig.EventBridgeBusName, "", "")
	sqsCfg := sqs.Init(envConfig.Region, "", "")
	outputCfg := s3.Init(infraConfig.BucketProfileOutput, "", envConfig.Region, "", "")
	backupCfg := s3.Init(infraConfig.BucketIngestorBackup, "", envConfig.Region, "", "")
	errorCfg := db.Init("ucp-error-table-"+envConfig.EnvName, "error_type", "error_id", "", "")
	if err != nil {
		return Services{}, err
	}
	return Services{
		Cognito:        cognito.InitWithClientID(infraConfig.CognitoUserPoolId, infraConfig.CognitoClientId, "", "", core.LogLevelDebug),
		ProfileStorage: profileStorage,
		IngestorStream: kinesis.Init(infraConfig.RealTimeIngestorStreamName, envConfig.Region, "", "", core.LogLevelDebug),
		EventBridge:    &ebCfg,
		SQS:            &sqsCfg,
		OutputBucket:   &outputCfg,
		BackupBucket:   &backupCfg,
		ErrorTable:     &errorCfg,
		RetryLambda:    lambda.Init("ucpRetry"+envConfig.EnvName, "", ""),
		Glue:           glue.Init(envConfig.Region, infraConfig.GlueDbName, "", ""),
	}, nil
}

///////////////////////////////////////////////////
// TEST UTILS
//////////////////////////////////////////////////

func Setup(t *testing.T, services Services, envCfg config.EnvConfig, infraCfg config.InfraConfig) (Config, error) {
	now := time.Now()
	ts := now.Format("20060102150405")
	domainName := "e2e_" + strings.ToLower(core.GenerateUniqueId())
	rootDir, err := config.AbsolutePathToProjectRootDir()
	if err != nil {
		t.Errorf("Error getting absolute path to project root dir: %v", err)
		return Config{}, err
	}
	postmanDir := rootDir + "source/ucp-backend/e2e/"

	// Create Cognito groups + users
	adminGroupName := "ucp-" + domainName + "-admin"
	adminGroupDescription := "*/*"
	err = services.Cognito.CreateGroup(adminGroupName, adminGroupDescription)
	if err != nil {
		t.Errorf("[TestRealTime] Error creating admin group: %s", err)
	}
	adminUserName := "e2e-admin-" + ts + "@example.com"
	adminPassword := core.RandomPassword(40)
	err = services.Cognito.CreateUser(adminUserName, adminGroupName, true)
	if err != nil {
		t.Errorf("[TestRealTime] Error creating admin user: %s", err)
	}
	err = services.Cognito.SetUserPassword(adminUserName, adminPassword, true)
	if err != nil {
		t.Errorf("[TestRealTime] Error setting admin user password: %s", err)
	}
	adminRefreshToken, adminAccessToken, err := services.Cognito.AuthenticateWithUsernamePassword(adminUserName, adminPassword)
	if err != nil {
		t.Errorf("[TestRealTime] Error getting refresh access token: %s", err)
	}
	if adminRefreshToken == "" {
		t.Errorf("[TestRealTime] Refresh token should not be empty")
	}
	// Create SQS queue to validate EventBridge events
	queueUrl, err := services.SQS.Create(domainName)
	if err != nil {
		t.Errorf("[TestRealTime] Error creating SQS queue: %s", err)
	}
	eventPattern := fmt.Sprintf("{\"source\": [\"ucp-%s\"]}", domainName) // rule fires on events for this domain only
	err = services.EventBridge.CreateSQSTarget(domainName, infraCfg.EventBridgeBusName, queueUrl, eventPattern)
	if err != nil {
		t.Errorf("[TestRealTime] Error creating SQS target: %s", err)
	}
	t.Cleanup(func() {
		err = services.EventBridge.DeleteRuleByName(domainName)
		if err != nil {
			t.Errorf("[%s] error deleting rule: %v", t.Name(), err)
		}
	})

	adminConfig := Config{
		APIGatewayURL:  infraCfg.ApiBaseUrl,
		TokenEndpoint:  infraCfg.CognitoTokenEndpoint,
		ClientID:       infraCfg.CognitoClientId,
		RefreshToken:   adminRefreshToken,
		AccessToken:    adminAccessToken,
		AdminUserName:  adminUserName,
		AdminGroupName: adminGroupName,
		DomainName:     domainName,
		EnvName:        envCfg.EnvName,
	}

	appAccessGroup := fmt.Sprintf("app-global-realtime-e2e-%s/%x", t.Name(), constants.AdminPermission)
	err = services.Cognito.CreateGroup(appAccessGroup, "")
	if err != nil {
		t.Fatalf("Error creating app access group %v", err)
	}
	err = services.Cognito.AddUserToGroup(adminUserName, appAccessGroup)
	if err != nil {
		t.Fatalf("Error adding user to app access group %v", err)
	}

	// CREATE DOMAIN
	// Run newman script for create domain
	folders := []string{AUTH, CREATE_DOMAIN, POLL_ASYNC_EVENT}
	createDomainCmd := buildCommand(postmanDir, folders, adminConfig, []string{})
	err = createDomainCmd.Run()
	if err != nil {
		t.Errorf("[TestRealTime] Error running create domain command: %s", err)
	}

	travelerTable := fmt.Sprintf("ucp_traveler_%s", envCfg.EnvName)
	partitions, err := services.Glue.GetPartitionsFromTable(travelerTable)
	log.Printf("Partitions %+v", partitions)
	if err != nil {
		t.Errorf("Error getting partitions from table %v", err)
	}
	hasDomainPartition := false
	for _, p := range partitions {
		if p.Values[0] == domainName {
			hasDomainPartition = true
			expected := "s3://" + infraCfg.BucketProfileOutput + "/profiles/domainname=" + domainName
			if p.Location != expected {
				t.Errorf("Partition location is incorrect, %v, should be %v", p.Location, domainName)
			}
		}
	}
	if !hasDomainPartition {
		t.Errorf("Partition should exists on table %v for domain %s", travelerTable, domainName)
	}

	return adminConfig, err
}

func Cleanup(t *testing.T, adminConfig Config, services Services, envCfg config.EnvConfig, infraCfg config.InfraConfig) error {
	// DELETE DOMAIN
	// Run newman script for delete domain
	folders := []string{AUTH, DELETE_DOMAIN, POLL_ASYNC_EVENT}
	rootDir, err := config.AbsolutePathToProjectRootDir()
	if err != nil {
		t.Errorf("Error getting absolute path to project root dir: %v", err)
		return err
	}
	postmanDir := rootDir + "source/ucp-backend/e2e/"
	deleteDomainCmd := buildCommand(postmanDir, folders, adminConfig, []string{})
	err = deleteDomainCmd.Run()
	if err != nil {
		t.Errorf("[TestRealTime] Error running delete domain command: %s", err)
	}

	//validate that partition has been removed
	travelerTable := fmt.Sprintf("ucp_traveler_%s", envCfg.EnvName)
	partitions, err := services.Glue.GetPartitionsFromTable(travelerTable)
	log.Printf("Partitions %+v", partitions)
	if err != nil {
		t.Errorf("Error getting partitions from table %v", err)
	}
	hasDomainPartition := false
	for _, p := range partitions {
		if p.Values[0] == adminConfig.DomainName {
			hasDomainPartition = true
		}
	}
	if hasDomainPartition {
		t.Errorf("Partition should have been deleted on table %v for domain %s", travelerTable, adminConfig.DomainName)
	}

	err = services.Cognito.DeleteGroup(fmt.Sprintf("app-global-realtime-e2e-%s/%x", t.Name(), constants.AdminPermission))
	if err != nil {
		t.Errorf("[TestRealTime] Error deleting group: %s", err)
	}
	// Clean up created resources
	err = services.Cognito.DeleteUser(adminConfig.AdminUserName)
	if err != nil {
		t.Errorf("[TestRealTime] Error deleting admin user: %s", err)
	}

	err = services.SQS.Delete()
	if err != nil {
		t.Errorf("[TestRealTime] Error deleting SQS queue: %s", err)
	}

	err = services.EventBridge.DeleteRuleByName(adminConfig.DomainName)
	if err != nil {
		t.Errorf("[TestRealTime] Error deleting event bridge rule")
	}
	return err
}

// Build a newman command that is run with specified folders and config env vars
//
// Special setup is required since the command cannot be run as a simple string
// (i.e. newman run --folder=folder1 --env-var key=value)
//
// Instead, each argument must be a separate string in a slice, and the slice
// is combined by os/exec to create the argv (argument vector) for the command.
//
// WARNING: the order of the folders is not the one provided in the command. it is the order of the folder
// in the collection!!! this can lead to strange behaviors
//
// See more:
// https://cs.opensource.google/go/go/+/refs/tags/go1.20.6:src/os/exec/exec.go;l=482
// https://pkg.go.dev/os/exec#Command
// https://pkg.go.dev/os#StartProcess
func buildCommand(directory string, folders []string, config Config, additionalArgs []string) *exec.Cmd {
	folderArgs := []string{}
	for _, f := range folders {
		folderArgs = append(folderArgs, "--folder")
		folderArgs = append(folderArgs, f)
	}
	envVarArgs := []string{
		"--verbose",
		"--env-var",
		"apiGatewayUrl=" + config.APIGatewayURL,
		"--env-var",
		"tokenEndpoint=" + config.TokenEndpoint,
		"--env-var",
		"clientId=" + config.ClientID,
		"--env-var",
		"refreshToken=" + config.RefreshToken,
		"--env-var",
		"domainName=" + config.DomainName,
		"--env-var",
		"environment=" + config.EnvName,
	}
	// allArgs := []string{"run", directory + "/ucp.postman_collection.json"}
	allArgs := []string{"run", "ucp.postman_collection.json"}
	allArgs = append(allArgs, folderArgs...)
	allArgs = append(allArgs, envVarArgs...)
	allArgs = append(allArgs, additionalArgs...)
	log.Printf("Running newman with args: %+v", allArgs)
	cmd := exec.Command("newman", allArgs...)
	cmd.Dir = directory
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func IngestFromFiles(t *testing.T, files []string, adminConfig Config, services Services) error {
	data, err := ReadJsonlFiles(files)
	if err != nil {
		log.Printf("[%s] [%s] Error ingesting files %v", t.Name(), "IngestFromFiles", err)
		return err
	}
	if len(data) == 0 {
		log.Printf("[%s] [%s] No data to ingest", t.Name(), "IngestFromFiles")
		return errors.New("no data to ingest (check that test file ends with an empty line)")
	}
	data = Transform(data, adminConfig)
	recs, err := BuildKinesisRecords(data)
	if err != nil {
		log.Printf("[%s] [%s] Error building kinesis records %v", t.Name(), "IngestFromFiles", err)
		return err
	}
	return IngestToKinesis(recs, services)
}

func Transform(data []map[string]interface{}, adminConfig Config) []map[string]interface{} {
	log.Printf("Transforming %v records with domain %v and timestamp", len(data), adminConfig.DomainName)
	transformed := []map[string]interface{}{}
	for _, rec := range data {
		rec["domain"] = adminConfig.DomainName
		rec["timestamp"] = time.Now().Format("2006-01-02T15:04:05.999Z")
		transformed = append(transformed, rec)
	}
	return transformed
}

func IngestToKinesis(records []kinesis.Record, services Services) error {
	// Send profiles with objects to new domain, poll in Go for profiles
	log.Printf("Sending records %+v", len(records))
	err, ingestionErrs := services.IngestorStream.PutRecords(records)
	if err != nil {
		log.Printf("[IngestToKinesis] Error sending profiles to domain: %s", err)
		return err
	}
	if len(ingestionErrs) > 0 {
		log.Printf("[IngestToKinesis] Errors sending profiles to domain: %s", ingestionErrs)
		return err
	}
	return nil
}

func BuildKinesisRecords(jsonRecs []map[string]interface{}) ([]kinesis.Record, error) {
	records := []kinesis.Record{}
	for _, jsonRec := range jsonRecs {
		jsonData, err := json.Marshal(jsonRec)
		if err != nil {
			log.Printf("[BuildKinesisRccords] Errors marshalling json records: %s", err)
			return records, err
		}
		records = append(records, kinesis.Record{
			Pk:   "pk-" + core.GenerateUniqueId(),
			Data: string(jsonData),
		})
	}
	return records, nil
}

func ReadJsonlFiles(paths []string) ([]map[string]interface{}, error) {
	log.Printf("Reading jsonl files: %v", paths)
	objects := []map[string]interface{}{}
	for _, path := range paths {
		data, err := ReadJsonlFile(path)
		if err != nil {
			return objects, err
		}
		objects = append(objects, data...)
	}
	return objects, nil
}

func ReadJsonlFile(path string) ([]map[string]interface{}, error) {
	log.Printf("Reading jsonl file: %v", path)
	objects := []map[string]interface{}{}
	// Load jsonl file
	file, err := os.Open(path)
	if err != nil {
		log.Printf("[ReadJsonlFile] Error reading from  file %v: %v", path, err)
		return objects, nil
	}
	defer file.Close()

	// Create map of json objects
	reader := bufio.NewReader(file)
	lineNumber := 0
	for {
		log.Printf("[ReadJsonlFile] Reading line %v", lineNumber)
		// Read line
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break // EOF or unexpected error
		}
		lineNumber++
		var data map[string]interface{}
		err = json.Unmarshal([]byte(line), &data)
		if err != nil {
			log.Printf("[ReadJsonlFile] Errors unmarshalling json : %s", err)
			return objects, err
		}
		objects = append(objects, data)
	}
	log.Printf("[ReadJsonlFile] Found %v objects in file %v", len(objects), path)
	return objects, nil
}

//////////////////////////////////////////////////////
// PROFILES UTILS
////////////////////////////////////////////

func WaitForProfile(reqHelper testutils.EndToEndRequest, profileID string, timeout int) (model.Traveller, error) {
	log.Printf("Waiting %v seconds for profile ID %v creation", timeout, profileID)
	it := 0
	for it*5 < timeout {
		profiles, _ := testutils.SearchProfiles(reqHelper, "travellerId", []string{profileID})
		if len(profiles) > 0 {
			log.Printf("Found profile ID: %v. (Connetid: %v)", profiles[0].TravellerID, profiles[0].ConnectID)
			fullProfile, err := testutils.GetProfile(
				reqHelper,
				profiles[0].ConnectID,
				constants.AccpRecordsNames(),
				[]customerprofiles.PaginationOptions{},
			)
			if err != nil {
				return model.Traveller{}, err
			}
			return fullProfile, nil
		}
		log.Printf("Not found. Waiting 5 seconds")
		time.Sleep(5 * time.Second)
		it += 1
	}
	return model.Traveller{}, errors.New("could not find profile ID: timeout expired")
}

func buildProfileObjectRefs(expectedFields []testutils.ExpectedField) []testutils.ProfileObjectRef {
	objestByRef := map[string]testutils.ProfileObjectRef{}
	objRefs := []testutils.ProfileObjectRef{}
	for _, expected := range expectedFields {
		if expected.ObjectType != "_profile" {
			objestByRef[expected.ObjectType+expected.ObjectID] = testutils.ProfileObjectRef{
				ObjectType: expected.ObjectType,
				ObjectID:   expected.ObjectID,
			}
		}
	}
	for _, ref := range objestByRef {
		objRefs = append(objRefs, ref)
	}
	return objRefs
}

func waitForIngestionCompletion(
	reqHelper testutils.EndToEndRequest,
	profileId string,
	profileObjects []testutils.ProfileObjectRef,
	timeout int,
) (fullProfile model.Traveller, err error) {
	log.Printf("[waitForIngestionCompletion] Waiting for profile %v full ingestion completion (timeout: %v seconds)", profileId, timeout)
	profile, err := testutils.WaitForProfile(reqHelper, profileId, timeout)
	if err != nil {
		return model.Traveller{}, err
	}
	for _, objectRef := range profileObjects {
		log.Printf(
			"[waitForIngestionCompletion] Waiting for object type %v with id %v to be fully ingested on profile %v",
			objectRef.ObjectType,
			objectRef.ObjectID,
			profile.ConnectID,
		)
		getSpecificProObjChan := make(chan error)
		go func(objectRef testutils.ProfileObjectRef) {
			for i := 0; i < timeout; i += 5 {
				profObject, err := testutils.GetSpecificProfileObject(
					reqHelper,
					objectRef.ObjectID,
					profile.ConnectID,
					objectRef.ObjectType,
				)
				if err != nil {
					log.Printf("[waitForIngestionCompletion] GetSpecificProfileObject error: %v", err)
					log.Printf("[waitForIngestionCompletion] Not Found. Refreshing ConnectID and waiting 5 seconds")
					profile, _ = testutils.WaitForProfile(reqHelper, profileId, 1)
					time.Sleep(5 * time.Second)
				} else {
					log.Printf("[waitForIngestionCompletion] Profile object: %+v", profObject.Attributes)
					getSpecificProObjChan <- nil
					return
				}
			}
			getSpecificProObjChan <- fmt.Errorf("[waitForIngestionCompletion] unable to locate profile object: %s | %s", objectRef.ObjectID, objectRef.ObjectType)
		}(objectRef)

		err := <-getSpecificProObjChan
		if err != nil {
			return model.Traveller{}, err
		}
	}
	fullProfile, err = testutils.GetProfile(
		reqHelper,
		profile.ConnectID,
		constants.AccpRecordsNames(),
		[]customerprofiles.PaginationOptions{},
	)
	if err != nil {
		return model.Traveller{}, err
	}

	return fullProfile, err
}
