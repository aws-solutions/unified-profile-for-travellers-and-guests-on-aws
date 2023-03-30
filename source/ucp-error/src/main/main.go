package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	core "tah/core/core"
	db "tah/core/db"
	model "tah/ucp-sync/src/business-logic/model"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
)

//Resources
var LAMBDA_ENV = os.Getenv("LAMBDA_ENV")
var LAMBDA_ACCOUNT_ID = os.Getenv("LAMBDA_ACCOUNT_ID")
var LAMBDA_REGION = os.Getenv("AWS_REGION")

var DYNAMO_TABLE = os.Getenv("DYNAMO_TABLE")
var DYNAMO_PK = os.Getenv("DYNAMO_PK")
var DYNAMO_SK = os.Getenv("DYNAMO_SK")

var errDb = db.Init(DYNAMO_TABLE, DYNAMO_PK, DYNAMO_SK)

func HandleRequest(ctx context.Context, req events.SQSEvent) (model.ResponseWrapper, error) {
	return HandleRequestWithServices(ctx, req, errDb)
}

func HandleRequestWithServices(ctx context.Context, req events.SQSEvent, configDb db.DBConfig) (model.ResponseWrapper, error) {
	tx := core.NewTransaction("ucp-sync", "")
	tx.Log("Received SQS event %+v with context %+v", req, ctx)
	var lastErr error
	for _, rec := range req.Records {
		ucpErr, err := parseSQSRecord(rec)
		if err != nil {
			ucpErr = createUnknownErrorError(rec, err)
		}
		err = saveRecord(configDb, ucpErr)
		if err != nil {
			lastErr = err
			tx.Log("An error occured while saving the error in DynamoDB")
		}
	}
	return model.ResponseWrapper{}, lastErr
}

func main() {
	lambda.Start(HandleRequest)
}

func saveRecord(cdb db.DBConfig, ucpErr model.UcpIngestionError) error {
	_, err := cdb.Save(ucpErr)
	return err
}

func parseSQSRecord(rec events.SQSMessage) (model.UcpIngestionError, error) {
	errType := aws.StringValue(rec.MessageAttributes["UcpErrorType"].StringValue)
	domain := aws.StringValue(rec.MessageAttributes["DomainName"].StringValue)
	message := aws.StringValue(rec.MessageAttributes["Message"].StringValue)
	bizObject := aws.StringValue(rec.MessageAttributes["BusinessObjectTypeName"].StringValue)
	accpRec := aws.StringValue(rec.MessageAttributes["ObjectTypeName"].StringValue)
	if errType == "" {
		if domain != "" {
			errType = model.ACCP_INGESTION_ERROR
		} else {
			return model.UcpIngestionError{}, errors.New("Unknown record type (no 'UcpErrorType' or 'Domain' attribute")
		}

	}
	ucpIngErr := model.UcpIngestionError{
		Type:               errType,
		ID:                 core.UUID(),
		Message:            message,
		Domain:             domain,
		BusinessObjectType: bizObject,
		AccpRecordType:     accpRec,
		Record:             rec.Body,
		TravellerID:        parseTravellerID(rec.Body),
	}
	return ucpIngErr, nil
}

func createUnknownErrorError(rec events.SQSMessage, err error) model.UcpIngestionError {
	return model.UcpIngestionError{
		Type:    model.ERROR_PARSING_ERROR,
		ID:      core.UUID(),
		Message: fmt.Sprintf("Unknown SQS record type with Body '%s' and attributes %v. Error: %v", rec.Body, rec.MessageAttributes, err),
	}
}

func parseTravellerID(body string) string {
	accpRecord := model.AccpRecord{}
	json.Unmarshal([]byte(body), &accpRecord)
	return accpRecord.TravellerID
}
