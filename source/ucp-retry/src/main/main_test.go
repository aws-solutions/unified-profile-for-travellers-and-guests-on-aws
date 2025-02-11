// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"
	"time"

	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	model "tah/upt/source/ucp-common/src/model/admin"
	testutils "tah/upt/source/ucp-common/src/utils/test"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var UCP_REGION = testutils.GetTestRegion()

func TestMain(t *testing.T) {
	// Set up resources
	retryTableName := "ucp-retry-test-" + time.Now().Format("2006-01-02-15-04-05")
	retryDb, err := db.InitWithNewTable(retryTableName, "traveller_id", "object_id", "", "")
	if err != nil {
		t.Fatalf("Error creating retry table: %s", err)
	}
	t.Cleanup(func() { retryDb.DeleteTable(retryTableName) })
	err = retryDb.WaitForTableCreation()
	if err != nil {
		t.Fatalf("Error creating retry table: %s", err)
	}
	errorTableName := "ucp-error-test-" + time.Now().Format("2006-01-02-15-04-05")
	errorDb, err := db.InitWithNewTable(errorTableName, "error_type", "error_id", "", "")
	if err != nil {
		t.Fatalf("Error creating error table: %s", err)
	}
	t.Cleanup(func() { errorDb.DeleteTable(errorTableName) })
	err = errorDb.WaitForTableCreation()
	if err != nil {
		t.Fatalf("Error creating error table: %s", err)
	}
	existingError := model.UcpIngestionError{
		Type: model.ACCP_INGESTION_ERROR,
		ID:   "error-1",
	}
	_, err = errorDb.Save(existingError)
	if err != nil {
		t.Errorf("Error saving error object: %s", err)
	}
	existingRetry := RetryObject{
		Traveller:  "traveller-2",
		Object:     "hotel_booking_record-2",
		RetryCount: 2,
		TTL:        int(time.Now().Add(time.Hour).Unix()),
	}
	maxedRetry := RetryObject{
		Traveller:  "traveller-4",
		Object:     "hotel_booking_record-4",
		RetryCount: 10,
		TTL:        int(time.Now().Add(time.Hour).Unix()),
	}
	_, err = retryDb.Save(existingRetry)
	if err != nil {
		t.Errorf("Error saving retry object: %s", err)
	}
	_, err = retryDb.Save(maxedRetry)
	if err != nil {
		t.Errorf("Error saving retry object: %s", err)
	}
	accpConfig := customerprofiles.InitMock(nil, nil, nil, nil, nil)
	solutionsConfig := awssolutions.InitMock()
	tx := core.NewTransaction("ucp-retry-test", "", core.LogLevelDebug)
	event := DynamoEvent{
		Records: []DynamoEventRecord{
			// Error being processed for the first time
			{
				EventID:   "1",
				EventName: EVENT_NAME_INSERT,
				Change: DynamoEventChange{
					NewImage: map[string]*dynamodb.AttributeValue{
						"error_type": {
							S: aws.String(model.ACCP_INGESTION_ERROR),
						},
						"error_id": {
							S: aws.String("error-1"),
						},
						"travelerId": {
							S: aws.String("traveller-1"),
						},
						"recordId": {
							S: aws.String("record-1"),
						},
						"accpRecordType": {
							S: aws.String("hotel_booking"),
						},
						"category": {
							S: aws.String(model.ACCP_INGESTION_ERROR),
						},
					},
				},
			},
			// Existing error being removed from the table, no action should be taken
			{
				EventID:   "2",
				EventName: EVENT_NAME_REMOVE,
				Change: DynamoEventChange{
					OldImage: map[string]*dynamodb.AttributeValue{
						"travelerId": {
							S: aws.String("traveller-1"),
						},
						"recordId": {
							S: aws.String("record-1"),
						},
						"accpRecordType": {
							S: aws.String("hotel_booking"),
						},
						"category": {
							S: aws.String(model.ACCP_INGESTION_ERROR),
						},
					},
				},
			},
			// Error being retried nth time
			{
				EventID:   "3",
				EventName: EVENT_NAME_MODIFY,
				Change: DynamoEventChange{
					OldImage: map[string]*dynamodb.AttributeValue{
						"travelerId": {
							S: aws.String("traveller-2"),
						},
						"recordId": {
							S: aws.String("record-2"),
						},
						"accpRecordType": {
							S: aws.String("hotel_booking"),
						},
						"category": {
							S: aws.String(model.ACCP_INGESTION_ERROR),
						},
					},
					NewImage: map[string]*dynamodb.AttributeValue{
						"travelerId": {
							S: aws.String("traveller-2"),
						},
						"recordId": {
							S: aws.String("record-2"),
						},
						"accpRecordType": {
							S: aws.String("hotel_booking"),
						},
						"category": {
							S: aws.String(model.ACCP_INGESTION_ERROR),
						},
					},
				},
			},
			// Parsing error that cannot be retried due to missing fields
			{
				EventID:   "4",
				EventName: EVENT_NAME_INSERT,
				Change: DynamoEventChange{
					NewImage: map[string]*dynamodb.AttributeValue{
						"travelerId": {
							S: aws.String("traveller-3"),
						},
						"recordId": {
							S: aws.String("record-3"),
						},
						"accpRecordType": {
							S: aws.String("hotel_booking"),
						},
						"category": {
							S: aws.String(model.ERROR_PARSING_ERROR),
						},
					},
				},
			},
			// Error that has been retried the max number of times
			{
				EventID:   "5",
				EventName: EVENT_NAME_MODIFY,
				Change: DynamoEventChange{
					NewImage: map[string]*dynamodb.AttributeValue{
						"travelerId": {
							S: aws.String("traveller-4"),
						},
						"recordId": {
							S: aws.String("record-4"),
						},
						"accpRecordType": {
							S: aws.String("hotel_booking"),
						},
						"category": {
							S: aws.String(model.ACCP_INGESTION_ERROR),
						},
					},
				},
			},
			// Additional errors for a single traveller
			{
				EventID:   "6",
				EventName: EVENT_NAME_INSERT,
				Change: DynamoEventChange{
					NewImage: map[string]*dynamodb.AttributeValue{
						"travelerId": {
							S: aws.String("traveller-5"),
						},
						"recordId": {
							S: aws.String("record-5"),
						},
						"accpRecordType": {
							S: aws.String("hotel_booking"),
						},
						"category": {
							S: aws.String(model.ACCP_INGESTION_ERROR),
						},
					},
				},
			},
			{
				EventID:   "6",
				EventName: EVENT_NAME_INSERT,
				Change: DynamoEventChange{
					NewImage: map[string]*dynamodb.AttributeValue{
						"travelerId": {
							S: aws.String("traveller-5"),
						},
						"recordId": {
							S: aws.String("record-6"),
						},
						"accpRecordType": {
							S: aws.String("hotel_booking"),
						},
						"category": {
							S: aws.String(model.ACCP_INGESTION_ERROR),
						},
					},
				},
			},
			{
				EventID:   "6",
				EventName: EVENT_NAME_INSERT,
				Change: DynamoEventChange{
					NewImage: map[string]*dynamodb.AttributeValue{
						"travelerId": {
							S: aws.String("traveller-5"),
						},
						"recordId": {
							S: aws.String("record-7"),
						},
						"accpRecordType": {
							S: aws.String("hotel_booking"),
						},
						"category": {
							S: aws.String(model.ACCP_INGESTION_ERROR),
						},
					},
				},
			},
		},
	}

	// Test
	err = HandleRequestWithServices(context.Background(), event, tx, accpConfig, retryDb, errorDb, solutionsConfig)
	if err != nil {
		t.Errorf("Error running lambda: %s", err)
	}
	rec1 := RetryObject{}
	err = retryDb.Get("traveller-1", "hotel_booking_record-1", &rec1)
	if err != nil {
		t.Errorf("Error getting retry object: %s", err)
	}
	if rec1.RetryCount != 1 {
		t.Errorf("Expected retry count to be 1, got %d", rec1.RetryCount)
	}
	rec2 := RetryObject{}
	err = retryDb.Get("traveller-2", "hotel_booking_record-2", &rec2)
	if err != nil {
		t.Errorf("Error getting retry object: %s", err)
	}
	if rec2.RetryCount != 3 {
		t.Errorf("Expected retry count to be 3, got %d", rec2.RetryCount)
	}
	rec3 := RetryObject{}
	err = retryDb.Get("traveller-5", "hotel_booking_record-5", &rec3)
	if err != nil {
		t.Errorf("Error getting retry object: %s", err)
	}
	if rec3.RetryCount != 1 {
		t.Errorf("Expected retry count to be 1, got %d", rec3.RetryCount)
	}
	if len(solutionsConfig.Metrics) == 0 || solutionsConfig.Metrics["service"] != "ucp-retry" {
		t.Errorf("Expected metrics to be recorded")
	}
	errors := []map[string]interface{}{}
	_, err = errorDb.FindAll(&errors, nil)
	if err != nil {
		t.Errorf("Error scanning items in error table %v", err)
	}
	if len(errors) > 0 {
		t.Errorf("Retried errors should not be in error table")
	}
}
