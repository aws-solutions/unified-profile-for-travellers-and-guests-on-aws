// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/s3"
	constants "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/admin"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestMain(t *testing.T) {
	// DynamoDB mock
	tableName := "connector-transfer-test-" + core.GenerateUniqueId()
	dbClient, err := db.InitWithNewTable(tableName, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("Error creating test table: %v", err)
	}
	t.Cleanup(func() { dbClient.DeleteTable(tableName) })
	domainList := model.IndustryConnectorDomainList{
		ItemID:     constants.CONFIG_DB_LINKED_CONNECTORS_PK,
		ItemType:   constants.CONFIG_DB_LINKED_CONNECTORS_SK,
		DomainList: []string{"domain1", "domain2"},
	}
	dbClient.WaitForTableCreation()
	_, err = dbClient.Save(domainList)
	if err != nil {
		t.Errorf("Error saving domain list: %v", err)
	}

	// S3 mock
	s3Client, err := s3.InitWithRandBucket("connector-transfer-test", "", "us-east-1", "", "")
	if err != nil {
		t.Errorf("Error creating test bucket: %v", err)
	}
	t.Cleanup(func() { s3Client.EmptyAndDelete() })
	clickstreamFile1Key := "clickstream/test-key-1.jsonl"
	clickstreamFile2Key := "clickstream/test-key-2.jsonl"
	airBookingFile1Key := "bookings/test-key-1.jsonl"
	err = s3Client.UploadFile(clickstreamFile1Key, "../../test_data/clickstream/data1.jsonl")
	if err != nil {
		t.Errorf("Error uploading test data: %v", err)
	}
	err = s3Client.UploadFile(clickstreamFile2Key, "../../test_data/clickstream/data2.jsonl")
	if err != nil {
		t.Errorf("Error uploading test data: %v", err)
	}
	err = s3Client.UploadFile(airBookingFile1Key, "../../test_data/air_booking/data1.jsonl")
	if err != nil {
		t.Errorf("Error uploading test data: %v", err)
	}

	// S3 PutObject mock event
	s3Event := events.S3Event{
		Records: []events.S3EventRecord{
			{
				EventName: "ObjectCreated:Put",
				S3: events.S3Entity{
					Bucket: events.S3Bucket{
						Name: s3Client.Bucket,
					},
					Object: events.S3Object{
						Key: clickstreamFile1Key,
					},
				},
			},
			{
				EventName: "ObjectCreated:Put",
				S3: events.S3Entity{
					Bucket: events.S3Bucket{
						Name: s3Client.Bucket,
					},
					Object: events.S3Object{
						Key: clickstreamFile2Key,
					},
				},
			},
			{
				EventName: "ObjectCreated:Put",
				S3: events.S3Entity{
					Bucket: events.S3Bucket{
						Name: s3Client.Bucket,
					},
					Object: events.S3Object{
						Key: airBookingFile1Key,
					},
				},
			},
		},
	}

	// Kinesis mock
	kinesisClientMock := kinesis.InitMock(nil, nil, nil, nil)

	// Run Lambda function
	err = HandleRequestWithServices(context.TODO(), s3Event, dbClient, s3Client, kinesisClientMock)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	// Validate Kinesis records match expected output
	//	Expect:
	//		4 records from clickstream/data1.jsonl
	//			2 clickstream records * 2 domains
	//		4 records from clickstream/data2.jsonl
	//			2 clickstream records * 2 domains
	//		4 records from air_booking/data1.jsonl
	//			2 individual records * 2 domains
	//	4 + 4 + 4 = 12
	recs := *kinesisClientMock.Records
	if len(recs) != 12 {
		t.Errorf("Expected 12 records, got %d", len(recs))
	}
	if recs[0].Pk != "clickstream/test-key-1.jsonl" {
		t.Errorf("Expected clickstream/test-key-1.jsonl, got %s", recs[0].Pk)
	}

	content, err := os.ReadFile("../../test_data/clickstream/data1.jsonl")
	if err != nil {
		t.Errorf("Error reading test data: %v", err)
	}

	contentLineOne := strings.Split(string(content), "\n")[0]

	clickstream1ExpectedData := `{"domain":"domain1","objectType":"clickstream","modelVersion":"1.0","data":` + contentLineOne + `}`
	var expectedClickstreamMap map[string]interface{}
	err = json.Unmarshal([]byte(clickstream1ExpectedData), &expectedClickstreamMap)
	if err != nil {
		t.Errorf("Error unmarshalling expected clickstream data: %v", err)
	}

	var dataMap map[string]interface{}
	err = json.Unmarshal([]byte(recs[0].Data), &dataMap)
	if err != nil {
		t.Errorf("Error unmarshalling clickstream data: %v", err)
	}

	if !reflect.DeepEqual(dataMap, expectedClickstreamMap) {
		t.Errorf("Expected %s, got %s", clickstream1ExpectedData, recs[0].Data)
	}
}

func TestCombineAllErrors(t *testing.T) {
	err := combineAllErrors(errors.New("test_errors"), []kinesis.IngestionError{{ErrorMessage: "kinsis_error"}})
	if err == nil {
		t.Errorf("combineAllErrors shoudl resturn a non nil error")
	}
}
