package main

import (
	"context"
	"log"
	db "tah/core/db"
	model "tah/ucp-sync/src/business-logic/model"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	//model "cloudrack-lambda-core/config/model"
)

func TestHandler(t *testing.T) {
	log.Printf("INIT FUNCTION TEST")
	log.Printf("*******************")
	log.Printf("Create test table")
	cfg, err := db.InitWithNewTable("ucp-errors-tests", "error_type", "error_id")
	if err != nil {
		t.Errorf("[TestHandler] error init with new table: %v", err)
	}
	cfg.WaitForTableCreation()
	req := events.SQSEvent{
		Records: []events.SQSMessage{
			events.SQSMessage{
				Body: "{}",
				MessageAttributes: map[string]events.SQSMessageAttribute{
					"UcpErrorType":           events.SQSMessageAttribute{StringValue: aws.String("etl_transformer")},
					"BusinessObjectTypeName": events.SQSMessageAttribute{StringValue: aws.String("e2e_biz_object_type")},
					"Message":                events.SQSMessageAttribute{StringValue: aws.String("A random error message")},
				},
			},
			events.SQSMessage{
				Body: "some weird text",
			},
			events.SQSMessage{
				Body: "{\"traveller_id\": \"123456789\"}",
				MessageAttributes: map[string]events.SQSMessageAttribute{
					"Message":        events.SQSMessageAttribute{StringValue: aws.String("Accp error message")},
					"DomainName":     events.SQSMessageAttribute{StringValue: aws.String("error_e2e_domain")},
					"ObjectTypeName": events.SQSMessageAttribute{StringValue: aws.String("e2e_accp_rec_type")},
				},
			},
		},
	}

	_, err = HandleRequestWithServices(context.Background(), req, cfg)
	if err != nil {
		t.Errorf("[TestHandler] Error executing lambda handler %v", err)
	}

	recs := []model.UcpIngestionError{}
	err1 := cfg.FindAll(&recs)
	if err1 != nil {
		t.Errorf("[TestHandler] Error fetching records fromr dynamo %v", err1)
	}
	log.Printf("Errors in dynamo: %+v", recs)
	if len(recs) != len(req.Records) {
		t.Errorf("[TestHandler] Dynamo table shoudl have %v records and has %v", len(req.Records), len(recs))
	}

	log.Printf("5. Deleteing table")
	err = cfg.DeleteTable(cfg.TableName)
	if err != nil {
		t.Errorf("[TestHandler] Error deleting test table %v", err)
	}
}
