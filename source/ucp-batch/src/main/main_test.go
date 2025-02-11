// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"testing"
	"time"

	"strings"

	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/cloudwatch"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/s3"
	"tah/upt/source/tah-core/sqs"

	"github.com/aws/aws-lambda-go/events"

	testutils "tah/upt/source/ucp-common/src/utils/test"
)

var UCP_REGION = testutils.GetTestRegion()

type TestCase struct {
	Name            string
	S3FileName      string
	Domain          string
	AccpObjectType  string
	Data            [][]string
	ExpectedErrors  []sqs.Message
	ExpectedRecords []TestKinesisRecord
}

type TestKinesisRecord struct {
	Data BusinessObjectRecord
}

var TEST_DOMAIN_1 = "test_domain_1"
var TEST_CASES = []TestCase{
	{
		Name:           "Happy Path",
		S3FileName:     "happy_path.csv",
		Domain:         TEST_DOMAIN_1,
		AccpObjectType: "hotel_booking",
		Data: [][]string{
			{"traveller_id", "object_type", "accp_object_id", "profile_id"},
			{"t1", "hotel_booking", "b1", "t1"},
			{"t1", "hotel_booking", "b2", "t1"},
			{"t2", "hotel_booking", "b1", "t2"},
		},
		ExpectedErrors: []sqs.Message{},
		ExpectedRecords: []TestKinesisRecord{
			{
				Data: BusinessObjectRecord{
					Domain:     "test_domain_1",
					ObjectType: "hotel_booking",
					Data:       []interface{}{map[string]interface{}{"accp_object_id": "b1", "object_type": "hotel_booking", "profile_id": "t1", "traveller_id": "t1"}},
				},
			},
			{
				Data: BusinessObjectRecord{
					Domain:     "test_domain_1",
					ObjectType: "hotel_booking",
					Data:       []interface{}{map[string]interface{}{"accp_object_id": "b2", "object_type": "hotel_booking", "profile_id": "t1", "traveller_id": "t1"}},
				},
			},
			{
				Data: BusinessObjectRecord{
					Domain:     "test_domain_1",
					ObjectType: "hotel_booking",
					Data:       []interface{}{map[string]interface{}{"accp_object_id": "b1", "object_type": "hotel_booking", "profile_id": "t2", "traveller_id": "t2"}}},
			},
		},
	},
	{
		Name:            "Large File",
		S3FileName:      "largeFile.csv",
		Domain:          TEST_DOMAIN_1,
		AccpObjectType:  "hotel_booking",
		Data:            buildLargeFile(2010),
		ExpectedErrors:  []sqs.Message{},
		ExpectedRecords: buildExpectedResult(TEST_DOMAIN_1, "hotel_booking", 2010),
	},
	{
		Name:           "Empty File",
		S3FileName:     "emptyFile.csv",
		Domain:         TEST_DOMAIN_1,
		AccpObjectType: "hotel_booking",
		Data:           [][]string{},
		ExpectedErrors: []sqs.Message{
			{
				MessageAttributes: map[string]string{
					"BusinessObjectTypeName": "hotel_booking",
					"Message":                "no data to process in file test_domain_1/hotel_booking/emptyFile.csv",
					"UcpErrorType":           "Batch ingestion error",
				},
			},
		},
	},
	{
		Name:           "empty domain",
		S3FileName:     "emptyDomain.csv",
		Domain:         "",
		AccpObjectType: "hotel_booking",
		Data:           [][]string{},
		ExpectedErrors: []sqs.Message{
			{
				MessageAttributes: map[string]string{
					"BusinessObjectTypeName": "Unknown",
					"Message":                "Could not parse domain and business object from key /hotel_booking/emptyDomain.csv",
					"UcpErrorType":           "Batch ingestion error",
				},
			},
		},
	},
	{
		Name:           "Invalid CVS",
		S3FileName:     "invalidCSV.csv",
		Domain:         TEST_DOMAIN_1,
		AccpObjectType: "hotel_booking",
		Data: [][]string{
			{"traveller_id", "object_type", "accp_object_id", "profile_id"},
			{"t1", "hotel_booking", "b1"},
		},
		ExpectedErrors: []sqs.Message{
			{
				MessageAttributes: map[string]string{
					"BusinessObjectTypeName": "hotel_booking",
					"Message":                "record on line 2: wrong number of fields",
					"UcpErrorType":           "Batch ingestion error",
				},
			},
		},
	},
	{
		Name:           "Invalid ACCP Object",
		S3FileName:     "invalidACCPObjects.csv",
		Domain:         TEST_DOMAIN_1,
		AccpObjectType: "hotel_booking",
		Data: [][]string{
			{"traveller_id", "invalid_object_type", "accp_object_id", "profile_id"},
			{"t1", "hotel_booking", "b1", "t1"},
		},
		ExpectedErrors: []sqs.Message{
			{
				MessageAttributes: map[string]string{
					"BusinessObjectTypeName": "hotel_booking",
					"Message":                "invalid batch 0. file test_domain_1/hotel_booking/invalidACCPObjects.csv format might be wrong. example error: 'could not infer accp object type'",
					"UcpErrorType":           "Batch ingestion error",
				},
			},
		},
	},
	{
		Name:           "Invalid TravelerID",
		S3FileName:     "invalidTravelerIDs.csv",
		Domain:         TEST_DOMAIN_1,
		AccpObjectType: "hotel_booking",
		Data: [][]string{
			{"invalid_traveller_id", "object_type", "accp_object_id", "profile_id"},
			{"t1", "hotel_booking", "b1", "t1"},
		},
		ExpectedErrors: []sqs.Message{
			{
				MessageAttributes: map[string]string{
					"BusinessObjectTypeName": "hotel_booking",
					"Message":                "invalid batch 0. file test_domain_1/hotel_booking/invalidTravelerIDs.csv format might be wrong. example error: 'could not infer accp object traveler id'",
					"UcpErrorType":           "Batch ingestion error",
				},
			},
		},
	},
	{
		Name:           "Invalid Object ID",
		S3FileName:     "invalidObjectId.csv",
		Domain:         TEST_DOMAIN_1,
		AccpObjectType: "hotel_booking",
		Data: [][]string{
			{"traveller_id", "object_type", "invalid_accp_object_id", "profile_id"},
			{"t1", "hotel_booking", "b1", "t1"},
		},
		ExpectedErrors: []sqs.Message{
			{
				MessageAttributes: map[string]string{
					"BusinessObjectTypeName": "hotel_booking",
					"Message":                "invalid batch 0. file test_domain_1/hotel_booking/invalidObjectId.csv format might be wrong. example error: 'could not infer accp object id'",
					"UcpErrorType":           "Batch ingestion error",
				},
			},
		},
	},
	{
		Name:           "Empty Line",
		S3FileName:     "invalidACCPObjects.csv",
		Domain:         TEST_DOMAIN_1,
		AccpObjectType: "hotel_booking",
		Data: [][]string{
			{"traveller_id", "invalid_object_type", "accp_object_id", "profile_id"},
			{"t1", "hotel_booking", "b1", "t1"},
		},
		ExpectedErrors: []sqs.Message{
			{
				MessageAttributes: map[string]string{
					"BusinessObjectTypeName": "hotel_booking",
					"Message":                "invalid batch 0. file test_domain_1/hotel_booking/invalidACCPObjects.csv format might be wrong. example error: 'could not infer accp object type'",
					"UcpErrorType":           "Batch ingestion error",
				},
			},
		},
	},
}

func TestLogInvalidJson(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	qName := "ucp-batch-" + core.GenerateUniqueId()
	sqsc := sqs.Init(UCP_REGION, "", "")
	_, err := sqsc.Create(qName)
	if err != nil {
		t.Errorf("[%s] Could not create test queue: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		log.Printf("[%s] Deleting queue", qName)
		err = sqsc.DeleteByName(qName)
		if err != nil {
			t.Errorf("[%s] Error deleting queue: %s", t.Name(), err)
		}
	})
	logIngestionError(tx, errors.New("test_error"), math.Inf(1), "objectType", sqsc, awssolutions.InitMock())
	queueContent, err := sqsc.Get(sqs.GetMessageOptions{})
	if err != nil {
		t.Errorf("[%s] Error getting messages: %s", t.Name(), err)
	}
	if len(queueContent.Peek) != 1 {
		t.Errorf("[%s] Expected 1 message, got %d", t.Name(), len(queueContent.Peek))
	}
	msg := queueContent.Peek[0]
	if msg.Body != `+Inf` {
		t.Errorf("[%s] Expected message data '+Inf', got '%s'", t.Name(), msg.Body)
	}
	if msg.MessageAttributes["Message"] != "json: unsupported value: +Inf" {
		t.Errorf("[%s] Expected message data 'json: unsupported value: +Inf', got '%s'", t.Name(), msg.MessageAttributes["Message"])
	}
}

func TestMain(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)

	s3Config, err := s3.InitWithRandBucket("ucp-batch-test", "", UCP_REGION, "", "")
	if err != nil {
		t.Errorf("[%s] Could not create test bucket", t.Name())
	}
	t.Cleanup(func() {
		err = s3Config.EmptyAndDelete()
		if err != nil {
			t.Errorf("[%s] Error emptying bucket: %s", t.Name(), err)
		}
	})

	for i, test := range TEST_CASES {
		log.Printf("[%s] Starting test case", test.Name)
		log.Printf("[%s] creating SQS queue", test.Name)
		sqsc := sqs.Init(UCP_REGION, "", "")
		qName := "ucp-batch-" + core.GenerateUniqueId() + "-test-" + strconv.Itoa(i)
		_, err = sqsc.Create(qName)
		if err != nil {
			t.Fatalf("[%s] Could not create test queue: %v", test.Name, err)
		}
		t.Cleanup(func() {
			log.Printf("[%s] Deleting queue", qName)
			err = sqsc.DeleteByName(qName)
			if err != nil {
				t.Errorf("[%s] Error deleting queue %s: %s", test.Name, qName, err)
			}
		})

		log.Printf("[%s] creating S3 events", test.Name)
		s3Event, err := sendTestDataToS3AndReturnEvents(test, s3Config)
		if err != nil {
			t.Errorf("[%s] Error sending data to s3: %s", test.Name, err)
		}

		stream := kinesis.Stream{}
		status := ""
		records := []kinesis.Record{}
		ingestionError := []kinesis.IngestionError{}
		kinesisClient := kinesis.InitMock(&stream, &status, &records, &ingestionError)

		log.Printf("[%s] Running HandleRequestWithServices", test.Name)
		err = HandleRequestWithServices(Request{
			Tx:              tx,
			S3Event:         s3Event,
			SolutionsConfig: solutionsConfig,
			Region:          UCP_REGION,
			Sqsc:            sqsc,
			SolId:           "",
			SolVersion:      "",
			MetricLogger:    cloudwatch.NewMetricLogger("test/namespace"),
			KinesisClient:   kinesisClient,
		},
		)
		if err != nil {
			t.Errorf("[%s] Error running test case: %s", test.Name, err)
		}

		log.Printf("[%s] Checking kinesis content", test.Name)
		differs := false
		if len(*kinesisClient.Records) != len(test.ExpectedRecords) {
			t.Fatalf("[%s] Expected %d records in kinesis, got %d: %+v", test.Name, len(test.ExpectedRecords), len(*kinesisClient.Records), *kinesisClient.Records)
			differs = true
		}
		expRecordsByTidOid := map[string]BusinessObjectRecord{}
		for _, expRec := range test.ExpectedRecords {
			if len(expRec.Data.Data) != 1 {
				t.Fatalf("[%s] Expected 1 data item in record, got %d", test.Name, len(expRec.Data.Data))
			}
			dataMap, ok := expRec.Data.Data[0].(map[string]interface{})
			if !ok {
				t.Fatalf("[%s] Could not convert expected data to map[string]interface{} : %+v", test.Name, expRec.Data.Data[0])
			}
			tid, ok := dataMap["traveller_id"].(string)
			if !ok {
				t.Fatalf("[%s] Could not find traveller_id in record: %v", test.Name, dataMap)
			}
			oid, ok := dataMap["accp_object_id"].(string)
			if !ok {
				t.Fatalf("[%s] Could not find accp_object_id in record %v", test.Name, dataMap)
			}
			expRecordsByTidOid[tid+oid] = expRec.Data
		}
		actualRecordsByTidOid := map[string]BusinessObjectRecord{}
		for _, actRec := range *kinesisClient.Records {
			actualData := BusinessObjectRecord{}
			err := json.Unmarshal([]byte(actRec.Data), &actualData)
			if err != nil {
				t.Fatalf("[%s] Could not unmarshal json record for record: %+v", test.Name, actRec)
			}
			dataMap, ok := actualData.Data[0].(map[string]interface{})
			if !ok {
				t.Fatalf("[%s] Could not convert actual data to map[string]interface{}: %+v", test.Name, actualData.Data[0])
			}
			tid, ok := dataMap["traveller_id"].(string)
			if !ok {
				t.Fatalf("[%s] Could not find traveller_id in record: %v", test.Name, dataMap)
			}
			oid, ok := dataMap["accp_object_id"].(string)
			if !ok {
				t.Fatalf("[%s] Could not find accp_object_id in record %v", test.Name, dataMap)
			}
			actualRecordsByTidOid[tid+oid] = actualData
		}
		for id, expRec := range expRecordsByTidOid {
			actRec, ok := actualRecordsByTidOid[id]
			if !ok {
				t.Fatalf("[%s] Expected record with id %s not found in actual records", test.Name, id)
			}
			matches, err := compareExpectedData(expRec, actRec)
			if err != nil {
				t.Fatalf("[%s] Error comparing expected and actual data: %s", test.Name, err)
			}
			if !matches {
				t.Fatalf("[%s] Expected record %+v does not match actual record %+v", test.Name, expRec, actRec)
			}
		}

		log.Printf("[%s] Checking SQS content", test.Name)

		maxTries := 6
		for i := 0; i < maxTries; i++ {
			log.Printf("[%s] Checking error queue", test.Name)
			queueContent, err := sqsc.Get(sqs.GetMessageOptions{})
			if err != nil {
				t.Errorf("[%s] Error getting messages: %s", test.Name, err)
			}
			log.Printf("[%s] Queue content: %+v", test.Name, queueContent)
			if len(queueContent.Peek) != len(test.ExpectedErrors) {
				log.Printf("[%s] queue length (%d) != expected (%d)", test.Name, len(queueContent.Peek), len(test.ExpectedErrors))
				log.Printf("[%s] Waiting 5 sec for messages to be available", test.Name)
				//wait for messages to be processed
				time.Sleep(5 * time.Second)
			} else {
				log.Printf("[%s] expected queue length found (%d)", test.Name, len(test.ExpectedErrors))
				break
			}
		}

		queueContent, err := sqsc.Get(sqs.GetMessageOptions{})
		log.Printf("[%s] Queue content: %+v", test.Name, queueContent)
		if err != nil {
			t.Errorf("[%s] Error getting messages: %s", test.Name, err)
		}
		if i >= maxTries {
			log.Printf("[%s] Max attempt %d reached. doing one last check", test.Name, maxTries)
			if len(queueContent.Peek) != len(test.ExpectedErrors) {
				t.Fatalf("[%s] invalid queue length: Expected %d messages, got %d", test.Name, len(test.ExpectedErrors), len(queueContent.Peek))
			}
		}
		expected := map[string]sqs.Message{}
		for _, sqsErr := range test.ExpectedErrors {
			msg := sqsErr.MessageAttributes["Message"]
			expected[msg] = sqsErr
		}
		inQueue := map[string]sqs.Message{}
		for _, sqsErr := range queueContent.Peek {
			msg := sqsErr.MessageAttributes["Message"]
			inQueue[msg] = sqsErr
		}

		differs = false
		for _, exp := range expected {
			if actual, ok := inQueue[exp.MessageAttributes["Message"]]; !ok {
				t.Errorf("[%s] Expected message '%s' not found in queue", test.Name, exp.MessageAttributes["Message"])
				differs = true
			} else {
				differs = compareSQSMessages(t, exp, actual)
			}
		}
		for _, act := range inQueue {
			if _, ok := expected[act.MessageAttributes["Message"]]; !ok {
				t.Errorf("[%s] message '%s' in queue found and not expected", test.Name, act.MessageAttributes["Message"])
				differs = true
			}
		}

		if !differs {
			log.Printf("[%s] SQS content check Successfull", test.Name)
		} else {
			t.Fatalf("[%s] SQS content check Failed", test.Name)
		}
	}
}

func compareExpectedData(expected BusinessObjectRecord, actual BusinessObjectRecord) (bool, error) {
	log.Printf("Comparing expected and actual data: %+v, %+v", expected, actual)

	if expected.Domain != actual.Domain {
		log.Printf("Expected and actual do not match on domain")
		return false, nil
	}
	if expected.ObjectType != actual.ObjectType {
		log.Printf("Expected and actual do not match on object type")
		return false, nil
	}
	if len(expected.Data) != len(actual.Data) {
		log.Printf("Expected and actual do not match on data length")
		return false, nil
	}

	if len(expected.Data) > 0 {
		expectedDataMap, ok := expected.Data[0].(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("Could not cast expected data to map[string]interface{}: %+v", expected.Data[0])
		}
		actualDataMap, ok := actual.Data[0].(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("Could not cast actual data to map[string]interface{} %+v", actual.Data[0])
		}
		for k, v := range expectedDataMap {
			if actualDataMap[k] != v {
				log.Printf("Expected and actual do not match on key %s", k)
				return false, nil
			}
		}
	}

	log.Printf("Expected and actual match on al keys")
	return true, nil
}

func buildLargeFile(fileSize int) [][]string {
	data := [][]string{
		{"traveller_id", "object_type", "accp_object_id", "profile_id"},
	}
	for i := 0; i < fileSize; i++ {
		data = append(data, []string{
			fmt.Sprintf("t%d", i),
			"hotel_booking",
			fmt.Sprintf("b%d", i),
			fmt.Sprintf("t%d", i),
		})
	}
	return data
}
func buildExpectedResult(domain string, objectType string, fileSize int) []TestKinesisRecord {
	data := []TestKinesisRecord{}
	for i := 0; i < fileSize; i++ {
		data = append(data, TestKinesisRecord{
			Data: BusinessObjectRecord{
				Domain:     domain,
				ObjectType: objectType,
				Data: []interface{}{map[string]interface{}{
					"accp_object_id": fmt.Sprintf("b%d", i),
					"object_type":    "hotel_booking",
					"profile_id":     fmt.Sprintf("t%d", i),
					"traveller_id":   fmt.Sprintf("t%d", i)},
				}},
		})
	}

	return data
}

func compareSQSMessages(t *testing.T, expected sqs.Message, actual sqs.Message) bool {
	exp := expected.MessageAttributes["BusinessObjectTypeName"]
	inQ := actual.MessageAttributes["BusinessObjectTypeName"]
	if exp != inQ {
		t.Errorf("Expected BusinessObjectTypeName=%s and NOT %s", exp, inQ)
		return true
	}
	exp = expected.MessageAttributes["UcpErrorType"]
	inQ = actual.MessageAttributes["UcpErrorType"]
	if exp != inQ {
		t.Errorf("Expected UcpErrorType=%s and NOT %s", exp, inQ)
		return true
	}
	exp = expected.MessageAttributes["Message"]
	inQ = actual.MessageAttributes["Message"]
	if exp != inQ {
		t.Errorf("Expected Message=%s and NOT %s", exp, inQ)
		return true
	}
	return false
}

func sendTestDataToS3AndReturnEvents(test TestCase, s3c s3.S3Config) (events.S3Event, error) {
	csvData := createCsvData(test.Data)
	log.Printf("[%s][sendTestDataToS3AndReturnEvents] csv data created: %+v ", test.Name, csvData)
	err := s3c.Save(test.Domain+"/"+test.AccpObjectType, test.S3FileName, []byte(csvData), "text/csv")
	if err != nil {
		log.Printf("[%s][sendTestDataToS3AndReturnEvents] Error saving file: %s", test.Name, err)
		return events.S3Event{}, err
	}
	return events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: s3c.Bucket},
					Object: events.S3Object{Key: test.Domain + "/" + test.AccpObjectType + "/" + test.S3FileName},
				},
			},
		},
	}, nil
}

func createCsvData(data [][]string) string {
	csv := ""
	if len(data) == 0 {
		return ""
	}
	for _, row := range data {
		csv += strings.Join(row, ",") + "\n"
	}
	return csv
}
