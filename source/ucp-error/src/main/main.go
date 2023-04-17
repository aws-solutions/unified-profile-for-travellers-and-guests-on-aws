package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	core "tah/core/core"
	customerprofiles "tah/core/customerprofiles"
	db "tah/core/db"
	model "tah/ucp-sync/src/business-logic/model"
	"time"

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

var ERROR_PK = "ucp_ingestion_error"
var ERROR_SK_PREFIX = "error_"

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
		retriedSuccess := false
		if err != nil {
			ucpErr = createUnknownErrorError(rec, err)
		}
		if ucpErr.Category == model.ACCP_INGESTION_ERROR {
			tx.Log("Retrying ACCP Error")
			err = retryAccpError(tx, ucpErr)
			if err != nil {
				tx.Log("Retry Failed")
				ucpErr.Message = ucpErr.Message + " (Retried)"
			} else {
				retriedSuccess = true
			}
		}
		if !retriedSuccess {
			tx.Log("Saving error in DynamoDB")
			err = saveRecord(configDb, ucpErr)
			if err != nil {
				lastErr = err
				tx.Log("An error occured while saving the error in DynamoDB")
			}
		}
	}
	return model.ResponseWrapper{}, lastErr
}

func retryAccpError(tx core.Transaction, ucpErr model.UcpIngestionError) error {
	customCfg := customerprofiles.InitWithDomain(ucpErr.Domain, LAMBDA_REGION)
	tx.Log("Putting profile object to ACCP in domain %s", ucpErr.Domain)
	return customCfg.PutProfileObject(ucpErr.Record, ucpErr.AccpRecordType)
}

func main() {
	lambda.Start(HandleRequest)
}

func saveRecord(cdb db.DBConfig, ucpErr model.UcpIngestionError) error {
	_, err := cdb.Save(ucpErr)
	return err
}

func parseSQSRecord(rec events.SQSMessage) (model.UcpIngestionError, error) {
	now := time.Now()
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
		Type:               ERROR_PK,
		ID:                 createErrorSk(now),
		Category:           errType,
		Message:            message,
		Domain:             domain,
		BusinessObjectType: bizObject,
		AccpRecordType:     accpRec,
		Record:             rec.Body,
		TravellerID:        parseTravellerID(rec.Body),
		Timestamp:          now,
	}
	return ucpIngErr, nil
}

func createErrorSk(ts time.Time) string {
	return ERROR_SK_PREFIX + ts.Format("2006-01-02-15-04-05") + core.UUID()
}

func createUnknownErrorError(rec events.SQSMessage, err error) model.UcpIngestionError {
	now := time.Now()
	return model.UcpIngestionError{
		Type:      ERROR_PK,
		Category:  model.ERROR_PARSING_ERROR,
		ID:        createErrorSk(now),
		Message:   fmt.Sprintf("Unknown SQS record type with Body '%s' and attributes %v. Error: %v", rec.Body, rec.MessageAttributes, err),
		Timestamp: now,
	}
}

func parseTravellerID(body string) string {
	accpRecord := model.AccpRecord{}
	json.Unmarshal([]byte(body), &accpRecord)
	return accpRecord.TravellerID
}
