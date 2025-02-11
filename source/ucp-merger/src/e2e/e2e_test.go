// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package merger

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	testutils "tah/upt/source/e2e"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/sqs"
	model "tah/upt/source/ucp-common/src/model/admin"
	"tah/upt/source/ucp-common/src/utils/config"
	upt_sdk "tah/upt/source/ucp-sdk/src"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsKinesis "github.com/aws/aws-sdk-go/service/kinesis"
)

type BusinessObjectRecord struct {
	Domain             string                 `json:"domain"`
	ObjectType         string                 `json:"objectType"`
	ModelVersion       string                 `json:"modelVersion"`
	Data               map[string]interface{} `json:"data"`
	Mode               string                 `json:"mode"`
	MergeModeProfileID string                 `json:"mergeModeProfileId"`
	PartialModeOptions PartialModeOptions     `json:"partialModeOptions"`
}

type PartialModeOptions struct {
	Fields []string `json:"fields"`
}

// overriding httpclient to prevent  Error: send request failed caused by: Post "https://kinesis.eu-central-1.amazonaws.com/": EOF
// https://github.com/aws/aws-sdk-go/issues/206
func InitKinesis(streamName, region, solutionId, solutionVersion string) *kinesis.Config {
	mySession := session.Must(session.NewSession())
	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	cfg := aws.NewConfig().WithRegion(region).WithMaxRetries(0).WithHTTPClient(httpClient)
	svc := awsKinesis.New(mySession, cfg)
	return &kinesis.Config{
		Client:    svc,
		Region:    region,
		Stream:    streamName,
		BatchSize: 10,
	}
}

/*
Test Cases:
1. Happy Path -> Successful Merge
2. Invalid Domain Name -> Error created with no retries
3. Nonexistent Domain -> Error created with retries
*/
func TestEndToEndMergeQueue(t *testing.T) {
	envCfg, infraCfg, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %s", err)
	}
	// SQS Merge Queue Client
	mergeQueueConfig := sqs.InitWithQueueUrl(infraCfg.MergeQueueUrl, envCfg.Region, "", "")

	// Error DB Client
	errorDbPk := "error_type"
	errorDbPkValue := "ucp_ingestion_error"
	errorDbClient := db.Init("ucp-error-table-"+envCfg.EnvName, errorDbPk, "error_id", "", "")

	// Kinesis Setup
	inCfg := InitKinesis(infraCfg.RealTimeIngestorStreamName, envCfg.Region, "", "")
	outCfg := kinesis.Init(infraCfg.RealTimeOutputStreamName, envCfg.Region, "", "", core.LogLevelDebug)
	outCfg.InitConsumer(kinesis.ITERATOR_TYPE_LATEST) // QQ: what does this do?

	domainName := "e2e_merger" + strings.ToLower(core.GenerateUniqueId())

	uptSdk, err := upt_sdk.Init(infraCfg)
	if err != nil {
		t.Fatalf("Error initializing SDK: %v", err)
	}
	err = uptSdk.SetupTestAuth(domainName)
	if err != nil {
		t.Fatalf("Error setting up auth: %v", err)
	}
	t.Cleanup(func() {
		err := uptSdk.CleanupAuth(domainName)
		if err != nil {
			t.Errorf("Error cleaning up auth: %v", err)
		}
	})
	err = uptSdk.CreateDomainAndWait(domainName, 300)
	if err != nil {
		t.Fatalf("Error creating domain: %v", err)
	}
	t.Cleanup(func() {
		err := uptSdk.DeleteDomainAndWait(domainName)
		if err != nil {
			t.Errorf("Error deleting domain: %v", err)
		}
	})
	// Create 2 Profiles
	log.Printf("[E2EMerger] 1- sending clickstream records to kinesis stream %v", infraCfg.RealTimeIngestorStreamName)
	record1 := buildRecord("clickstream", "clickstream", domainName, "", 0)
	record1Serialized, _ := json.Marshal(record1)
	record2 := buildRecord("hotel_booking", "hotel_booking", domainName, "", 0)
	record2Serialized, _ := json.Marshal(record2)

	records := []kinesis.Record{
		{
			Pk:   record1.ObjectType + "-" + record1.ModelVersion + "-1",
			Data: string(record1Serialized),
		},
		{
			Pk:   record2.ObjectType + "-" + record2.ModelVersion + "-2",
			Data: string(record2Serialized),
		},
	}
	log.Printf("[E2EMerger] 2- sending records %+v to kinesis stream %v", records, infraCfg.RealTimeIngestorStreamName)
	err, errs := inCfg.PutRecords(records)
	if err != nil {
		t.Fatalf("[E2EMerger] error sending data to stream: %v. %+v", err, errs)
	}
	//	Wait for ingestion merge rules (e.g. Alt Prof ID) to finish
	time.Sleep(30 * time.Second)

	log.Printf("[E2EMerger] Waiting for profile creation")
	log.Printf("%+v", record1)
	targetProfileId := ((record1.Data["session"].(map[string]interface{}))["id"]).(string)
	traveler1, err := uptSdk.WaitForProfile(domainName, targetProfileId, 90)
	if err != nil {
		t.Fatalf("Could not find profile with ID %s for clickstream rec. Error: %v", targetProfileId, err)
	}
	consumedProfileId := ((record2.Data["holder"].(map[string]interface{}))["id"]).(string)
	traveler2, err := uptSdk.WaitForProfile(domainName, consumedProfileId, 60)
	if err != nil {
		t.Fatalf("Could not find profile with ID %s for hotel booking record. Error: %v", consumedProfileId, err)
	}

	// Success Case: Valid Merge Request
	txId := "merge-e2e-txId-success" + core.GenerateUniqueId()
	mergeContext := customerprofiles.ProfileMergeContext{
		Timestamp:              time.Now(),
		MergeType:              customerprofiles.MergeTypeManual,
		ConfidenceUpdateFactor: 1,
		OperatorID:             "test_operator",
	}
	mergeReq := customerprofiles.MergeRequest{
		SourceID:      traveler2.ConnectID,
		TargetID:      traveler1.ConnectID,
		Context:       mergeContext,
		DomainName:    domainName,
		TransactionID: txId,
	}
	mergeReqSerialized, _ := json.Marshal(mergeReq)
	mergeQueueConfig.Send(string(mergeReqSerialized))

	log.Printf("[E2EMerger] Waiting for merge request to be processed. Waiting 15 seconds")

	// Check aurora profile results
	err = testutils.WaitForExpectedCondition(func() bool {
		connectId1, err := uptSdk.GetProfileId(domainName, targetProfileId)
		if err != nil {
			t.Logf("[%s] Error getting profile ID %v", t.Name(), err)
			return false
		}
		connectId2, err := uptSdk.GetProfileId(domainName, consumedProfileId)
		if err != nil {
			t.Logf("[%s] Error getting profile ID %v", t.Name(), err)
			return false
		}
		if connectId1 != connectId2 {
			t.Logf("[%s] ConnectIds do not match, merge was not successful", t.Name())
			return false
		}
		if connectId1 != traveler1.ConnectID && connectId2 != traveler1.ConnectID {
			t.Logf("[%s] Both connectIds should equal: %v", t.Name(), traveler1.ConnectID)
			return false
		}
		return true
	}, 10, 15*time.Second)
	if err != nil {
		t.Fatalf("[%s] error waiting for expected condition: %v", t.Name(), err)
	}

	t.Cleanup(func() {
		data := []model.UcpIngestionError{}
		filter := db.DynamoFilterExpression{
			BeginsWith: []db.DynamoFilterCondition{{Key: "message", Value: "Error retreiving profile with connect ID " + traveler2.ConnectID + " after 2 retries."}},
		}
		errorDbClient.FindStartingWithAndFilter("ucp_ingestion_error", "error_", &data, filter)
		if len(data) > 0 {
			err = errorDbClient.DeleteByKey(errorDbPkValue, data[0].ID)
			if err != nil {
				t.Errorf("Error deleting error from error table: %v", err)
			}
		} else {
			t.Log("Could not find error created from getting deleted profile")
		}
	})
	log.Println("Successfully Merged Profiles")

	// Fail Case: Invalid DomainName
	log.Printf("[E2EMerger] Testing merge with invalid domain name")
	invalidDomainName := "INVALID-DOMAIN"
	txId = "merge-e2e-txId-invalid-domain" + core.GenerateUniqueId()
	mergeContext = customerprofiles.ProfileMergeContext{
		Timestamp:              time.Now(),
		MergeType:              customerprofiles.MergeTypeManual,
		ConfidenceUpdateFactor: 1,
		OperatorID:             "test_operator",
	}
	invalidDomainMergeReq := customerprofiles.MergeRequest{
		SourceID:      traveler2.ConnectID,
		TargetID:      traveler1.ConnectID,
		Context:       mergeContext,
		DomainName:    invalidDomainName,
		TransactionID: txId,
	}
	invalidDomainMergeReqSerialized, _ := json.Marshal(invalidDomainMergeReq)
	mergeQueueConfig.Send(string(invalidDomainMergeReqSerialized))

	// Request should be sent directly to error table
	log.Printf("[E2EMerger] Waiting for error to be created")
	// Check Error Table
	errorId, err := waitForError(&errorDbClient, txId, 90)
	if err != nil {
		t.Fatalf("Error waiting for error to be created: %v", err)
	}
	t.Cleanup(func() {
		err = errorDbClient.DeleteByKey(errorDbPkValue, errorId)
		if err != nil {
			t.Errorf("Error deleting error from error table: %v", err)
		}
	})
	log.Println("Successfully errored on invalid domain name")

	// Fail Case: Invalid Merge
	log.Printf("[E2EMerger] Testing merge with valid context but nonexistent domain")
	nonexistentDomainName := "nonexistent_domain"
	txId = "merge-e2e-txId-nonexistent-domain" + core.GenerateUniqueId()
	mergeContext = customerprofiles.ProfileMergeContext{
		Timestamp:              time.Now().Round(0),
		MergeType:              customerprofiles.MergeTypeManual,
		ConfidenceUpdateFactor: 1,
		OperatorID:             "test_operator",
	}
	nonexistentDomainMergeReq := customerprofiles.MergeRequest{
		SourceID:      traveler2.ConnectID,
		TargetID:      traveler1.ConnectID,
		Context:       mergeContext,
		DomainName:    nonexistentDomainName,
		TransactionID: txId,
	}
	nonexistentDomainMergeReqSerialized, _ := json.Marshal(nonexistentDomainMergeReq)
	mergeQueueConfig.Send(string(nonexistentDomainMergeReqSerialized))

	// Request should be sent to error table
	log.Printf("[E2EMerger] Waiting for error to be created")
	// Check Error Table
	// Timeout is set to 4 minutes because sqs visibility timeout is set to 60 seconds and retries 3 times
	errorId2, err := waitForError(&errorDbClient, txId, 240)
	if err != nil {
		t.Fatalf("Error waiting for error to be created: %v", err)
	}
	log.Println("Successfully errored on nonexistent domain name")
	t.Cleanup(func() {
		err = errorDbClient.DeleteByKey(errorDbPkValue, errorId2)
		if err != nil {
			t.Errorf("Error deleting error from error table: %v", err)
		}
	})
}

func buildRecord(objectType string, folder string, domain string, mode string, line int) BusinessObjectRecord {
	newMode := ""
	if mode != "" {
		newMode = "_" + mode
	}
	content, err := os.ReadFile("../../../test_data/" + folder + "/data1" + newMode + ".jsonl")
	if err != nil {
		log.Printf("Error reading JSON file: %s", err)
		return BusinessObjectRecord{}
	}
	content = []byte(strings.Split(string(content), "\n")[line])

	var v map[string]interface{}
	json.Unmarshal(content, &v)

	return BusinessObjectRecord{
		Domain:       domain,
		ObjectType:   objectType,
		ModelVersion: "1",
		Data:         v,
		Mode:         mode,
	}
}

func waitForError(errorDbClient *db.DBConfig, txId string, timeout int) (string, error) {
	log.Printf("Waiting %v seconds for error creation", timeout)
	it := 0
	for it*10 < timeout {
		/*
			This query does not perform with large amounts of data in the error DB, since this is only run on the pipeline and
			the pipeline is not exposed to customers, we shouldn't have an issue.
		*/
		data := []model.UcpIngestionError{}
		filter := db.DynamoFilterExpression{
			BeginsWith: []db.DynamoFilterCondition{{Key: "transactionId", Value: txId}},
		}
		errorDbClient.FindStartingWithAndFilter("ucp_ingestion_error", "error_", &data, filter)
		if len(data) > 0 {
			log.Printf("Found error with transactionId: %v", data[0].TransactionID)
			return data[0].ID, nil
		}
		log.Printf("Not found. Waiting 10 seconds")
		time.Sleep(10 * time.Second)
		it += 1
	}
	return "", errors.New("Could not find error: timeout expired")
}
