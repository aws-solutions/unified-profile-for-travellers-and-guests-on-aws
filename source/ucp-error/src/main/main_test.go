// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"strings"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/core"
	db "tah/upt/source/tah-core/db"
	model "tah/upt/source/ucp-common/src/model/admin"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
)

func TestHandler(t *testing.T) {
	log.Printf("INIT FUNCTION TEST")
	log.Printf("*******************")
	log.Printf("Create test table")
	testPostFix := strings.ToLower(core.GenerateUniqueId())
	dbTableName := "ucp-errors-test" + testPostFix
	dbConfig, err := db.InitWithNewTable(dbTableName, "error_type", "error_id", "", "")
	if err != nil {
		t.Fatalf("[TestHandler] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		log.Printf("5. Deleting table")
		err = dbConfig.DeleteTable(dbConfig.TableName)
		if err != nil {
			t.Errorf("[TestHandler] Error deleting test table %v", err)
		}
	})
	dbConfig.WaitForTableCreation()
	solutionsConfig := awssolutions.InitMock()

	dbConfig.WaitForTableCreation()
	req := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				Body: "{}",
				MessageAttributes: map[string]events.SQSMessageAttribute{
					"UcpErrorType":           {StringValue: aws.String("etl_transformer")},
					"BusinessObjectTypeName": {StringValue: aws.String("e2e_biz_object_type")},
					"Message":                {StringValue: aws.String("A random error message")},
				},
			},
			{
				Body: "some weird text",
			},
			{
				Body: "{\"traveller_id\": \"123456789\"}",
				MessageAttributes: map[string]events.SQSMessageAttribute{
					"Message":        {StringValue: aws.String("Accp error message")},
					"DomainName":     {StringValue: aws.String("error_e2e_domain")},
					"ObjectTypeName": {StringValue: aws.String("e2e_accp_rec_type")},
				},
			},
		},
	}

	_, err = HandleRequestWithServices(context.Background(), req, dbConfig, solutionsConfig)
	if err != nil {
		t.Errorf("[TestHandler] Error executing lambda handler %v", err)
	}

	recs := []model.UcpIngestionError{}
	_, err1 := dbConfig.FindAll(&recs, nil)
	if err1 != nil {
		t.Errorf("[TestHandler] Error fetching records fromr dynamo %v", err1)
	}
	log.Printf("Errors in dynamo: %+v", recs)
	if len(recs) != len(req.Records) {
		t.Errorf("[TestHandler] Dynamo table shoudl have %v records and has %v", len(req.Records), len(recs))
	}

}

func TestParseAccpRecID(t *testing.T) {
	expected := "Y1AX8I58K9|Room Charge 2024-01-02|2024-01-02T00:00:00Z"
	actual := parseRecordID(RECORD)
	if actual != expected {
		t.Errorf("[TestParseAccpRecID] AccpRecId should be %v and not %v", expected, actual)
	}
}

var RECORD = `{"accp_object_id":"Y1AX8I58K9|Room Charge 2024-01-02|2024-01-02T00:00:00Z","amount":3863.889,"booking_id":"4N43K5398G","created_by":"Amya Douglas","created_on":"2023-08-26T15:16:54.531350Z","currency_code":"YER","currency_name":"Yemen Rial","currency_symbol":"","date":"2024-01-02T00:00:00Z","description":"Suite with Mini bar,Sea View ( Super Saver rate)","email":"","first_name":"Ola","hotel_code":"TOK19","id":"Y1AX8I58K9","last_name":"Veum","last_updated":"2023-11-17T17:02:10.257061Z","last_updated_by":"Adrien Rodriguez","last_updated_partition":"2023-11-17-17","model_version":"1.0","object_type":"hotel_stay_revenue_items","phone":"10233-438-2134","start_date":"2023-06-29","traveller_id":"9860225761","type":"Room Charge 2024-01-02"}`
