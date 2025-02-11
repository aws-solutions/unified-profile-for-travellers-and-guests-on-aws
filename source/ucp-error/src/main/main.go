// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/awssolutions"
	core "tah/upt/source/tah-core/core"
	db "tah/upt/source/tah-core/db"
	model "tah/upt/source/ucp-common/src/model/admin"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
)

type ResponseWrapper struct{}

// a versino of the ACCP records with only mandatory fields
type AccpRecord struct {
	ModelVersion string `json:"model_version"`
	ObjectType   string `json:"object_type"`
	LastUpdated  string `json:"last_updated"`
	TravellerID  string `json:"traveller_id"`
	RecordID     string `json:"accp_object_id"`
}

// Resources
var LAMBDA_ENV = os.Getenv("LAMBDA_ENV")
var LAMBDA_ACCOUNT_ID = os.Getenv("LAMBDA_ACCOUNT_ID")
var LAMBDA_REGION = os.Getenv("AWS_REGION")

var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
var METRICS_UUID = os.Getenv("METRICS_UUID")
var SEND_ANONYMIZED_DATA = os.Getenv("SEND_ANONYMIZED_DATA")

var DYNAMO_TABLE = os.Getenv("DYNAMO_TABLE")
var DYNAMO_PK = os.Getenv("DYNAMO_PK")
var DYNAMO_SK = os.Getenv("DYNAMO_SK")
var TTL = os.Getenv("TTL")

var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

var ERROR_PK = "ucp_ingestion_error"
var ERROR_SK_PREFIX = "error_"

var TX_ERROR = "ucp-error"

var errDb = db.Init(DYNAMO_TABLE, DYNAMO_PK, DYNAMO_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var solutionsConfig = awssolutions.Init(METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, METRICS_UUID, SEND_ANONYMIZED_DATA, LOG_LEVEL)

func HandleRequest(ctx context.Context, req events.SQSEvent) (ResponseWrapper, error) {
	return HandleRequestWithServices(ctx, req, errDb, solutionsConfig)
}

func HandleRequestWithServices(ctx context.Context, req events.SQSEvent, configDb db.DBConfig, solutionsConfig awssolutions.IConfig) (ResponseWrapper, error) {
	tx := core.NewTransaction(TX_ERROR, "", LOG_LEVEL)
	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")
	tx.Info("Received SQS event %+v with context %+v", req, ctx)
	var lastErr error
	now := time.Now()
	for _, rec := range req.Records {
		txID := aws.StringValue(rec.MessageAttributes["TransactionId"].StringValue)
		if txID != "" {
			tx = core.NewTransaction(TX_ERROR, txID, LOG_LEVEL)
		}
		ucpErr, err := parseSQSRecord(rec, tx, TTL)
		if err != nil {
			ucpErr = createUnknownErrorError(rec, err, tx, TTL)
		}
		tx.Debug("Saving error in DynamoDB")
		err = saveRecord(configDb, ucpErr)
		if err != nil {
			lastErr = err
			tx.Error("An error occured while saving the error in DynamoDB")
		}
	}
	_, err := solutionsConfig.SendMetrics(map[string]interface{}{
		"service":      TX_ERROR,
		"duration":     time.Since(now).Milliseconds(),
		"record_count": len(req.Records),
	})
	if err != nil {
		tx.Warn("Error while sending metrics to AWS Solutions: %s", err)
	}
	return ResponseWrapper{}, lastErr
}

func main() {
	lambda.Start(HandleRequest)
}

func saveRecord(cdb db.DBConfig, ucpErr model.UcpIngestionError) error {
	_, err := cdb.Save(ucpErr)
	return err
}

func parseSQSRecord(rec events.SQSMessage, tx core.Transaction, TTL string) (model.UcpIngestionError, error) {
	now := time.Now()
	errType := aws.StringValue(rec.MessageAttributes["UcpErrorType"].StringValue)
	domain := aws.StringValue(rec.MessageAttributes["DomainName"].StringValue)
	message := aws.StringValue(rec.MessageAttributes["Message"].StringValue)
	bizObject := aws.StringValue(rec.MessageAttributes["BusinessObjectTypeName"].StringValue)
	accpRec := aws.StringValue(rec.MessageAttributes["ObjectTypeName"].StringValue)
	txID := aws.StringValue(rec.MessageAttributes["TransactionId"].StringValue)
	isMergeRetryError, mergeReq := getMergeRetryError(rec)
	if isMergeRetryError {
		errType = "UCP merge error"
		message = "Failed to Merge profiles"
		domain = mergeReq.DomainName
		txID = mergeReq.TransactionID
	}
	tx.Info("Parsing error type: '%s' from SQS. domain: '%s', travellerID: '%s', error message: '%s', transaction ID: %v, bizObject: '%s' accpRec: '%s'",
		errType, domain, parseTravellerID(rec.Body), message, txID, bizObject, accpRec)
	if errType == "" {
		if domain != "" {
			tx.Error("No error type: error comes from ACCP")
			errType = model.ACCP_INGESTION_ERROR
		} else {
			return model.UcpIngestionError{}, errors.New("unknown record type (no 'UcpErrorType' or 'Domain' attribute")
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
		RecordID:           parseRecordID(rec.Body),
		TransactionID:      txID,
		Timestamp:          now,
		TTL:                createTTL(TTL, tx),
	}
	return ucpIngErr, nil
}

func createErrorSk(ts time.Time) string {
	return ERROR_SK_PREFIX + ts.Format("2006-01-02-15-04-05") + core.UUID()
}

func createUnknownErrorError(rec events.SQSMessage, err error, tx core.Transaction, TTL string) model.UcpIngestionError {
	now := time.Now()
	return model.UcpIngestionError{
		Type:      ERROR_PK,
		Category:  model.ERROR_PARSING_ERROR,
		ID:        createErrorSk(now),
		Message:   fmt.Sprintf("Unknown SQS record type with Body '%s' and attributes %v. Error: %v", rec.Body, rec.MessageAttributes, err),
		Timestamp: now,
		TTL:       createTTL(TTL, tx),
	}
}

func parseTravellerID(body string) string {
	accpRecord := AccpRecord{}
	json.Unmarshal([]byte(body), &accpRecord)
	return accpRecord.TravellerID
}

func parseRecordID(body string) string {
	accpRecord := AccpRecord{}
	json.Unmarshal([]byte(body), &accpRecord)
	return accpRecord.RecordID
}

func createTTL(ttl string, tx core.Transaction) int {
	days, err := strconv.Atoi(ttl)
	if err != nil {
		tx.Error("Error parsing TTL '%s', defaulting to 7 days", ttl)
		days = 7
	}
	now := time.Now()
	duration := time.Duration(days * int(time.Hour) * 24)
	return int(now.Add(duration).Unix())
}

func getMergeRetryError(rec events.SQSMessage) (bool, customerprofiles.MergeRequest) {
	var req customerprofiles.MergeRequest
	json.Unmarshal([]byte(rec.Body), &req)
	if req.SourceID != "" && req.TargetID != "" && req.DomainName != "" && req.TransactionID != "" && req.Context != (customerprofiles.ProfileMergeContext{}) {
		return true, req
	}
	return false, customerprofiles.MergeRequest{}
}
