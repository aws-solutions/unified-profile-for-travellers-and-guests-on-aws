// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/aurora"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	secrets "tah/upt/source/tah-core/secret"
	"tah/upt/source/tah-core/sqs"

	common "tah/upt/source/ucp-common/src/constant/admin"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/google/uuid"
)

var LAMBDA_REGION = os.Getenv("LAMBDA_REGION")
var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
var METRICS_NAMESPACE = "upt"

// Low cost storage
var USE_LOW_COST_STORAGE = os.Getenv("USE_LOW_COST_STORAGE")
var AURORA_PROXY_ENDPOINT = os.Getenv("AURORA_PROXY_ENDPOINT")
var AURORA_DB_NAME = os.Getenv("AURORA_DB_NAME")
var AURORA_DB_SECRET_ARN = os.Getenv("AURORA_DB_SECRET_ARN")
var STORAGE_CONFIG_TABLE_NAME = os.Getenv("STORAGE_CONFIG_TABLE_NAME")
var STORAGE_CONFIG_TABLE_PK = os.Getenv("STORAGE_CONFIG_TABLE_PK")
var STORAGE_CONFIG_TABLE_SK = os.Getenv("STORAGE_CONFIG_TABLE_SK")
var CHANGE_PROC_KINESIS_STREAM_NAME = os.Getenv("CHANGE_PROC_KINESIS_STREAM_NAME")
var CP_WRITER_QUEUE_URL = os.Getenv("CP_WRITER_QUEUE_URL")
var MERGE_QUEUE_URL = os.Getenv("MERGE_QUEUE_URL")

var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

// Init low-cost clients
var smClient = secrets.InitWithRegion(AURORA_DB_SECRET_ARN, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var user, pwd = getUserCredentials(smClient)
var auroraClient = aurora.InitPostgresql(AURORA_PROXY_ENDPOINT, AURORA_DB_NAME, user, pwd, 50, LOG_LEVEL)
var configTableClient = db.Init(
	STORAGE_CONFIG_TABLE_NAME,
	STORAGE_CONFIG_TABLE_PK,
	STORAGE_CONFIG_TABLE_SK,
	METRICS_SOLUTION_ID,
	METRICS_SOLUTION_VERSION,
)
var kinesisClient = kinesis.Init(CHANGE_PROC_KINESIS_STREAM_NAME, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, LOG_LEVEL)
var cpWriterQueue = sqs.InitWithQueueUrl(CP_WRITER_QUEUE_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var mergeQueue = sqs.InitWithQueueUrl(MERGE_QUEUE_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var LCS_CACHE = customerprofiles.InitLcsCache()

// Init SQS client
var DEAD_LETTER_QUEUE_URL = os.Getenv("DEAD_LETTER_QUEUE_URL")
var dlqConfig = sqs.InitWithQueueUrl(DEAD_LETTER_QUEUE_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

var TX_MERGER = "ucp-merger"

type ProcessingError struct {
	Error      error
	AccpRecord string
	Data       string
}

type ItemIdentifier struct {
	ItemIdentifier string `json:"itemIdentifier"`
}

func getUserCredentials(sm secrets.Config) (string, string) {
	user := sm.Get("username")
	pwd := sm.Get("password")
	return user, pwd
}

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) (map[string][]ItemIdentifier, error) {
	customerProfile := initStorage()
	tx := core.NewTransaction(TX_MERGER, "", LOG_LEVEL)
	return HandleRequestWithServices(tx, ctx, sqsEvent, customerProfile, &dlqConfig)
}

func HandleRequestWithServices(tx core.Transaction, ctx context.Context, sqsEvent events.SQSEvent, cp customerprofiles.ICustomerProfileLowCostConfig, dlqConfig sqs.IConfig) (map[string][]ItemIdentifier, error) {
	lc, ok := lambdacontext.FromContext(ctx)
	if ok {
		tx.Info("Lambda Request ID: %s", lc.AwsRequestID)
	} else {
		tx.Warn("Unable to read Lambda context")
	}

	batchItemFailures := []ItemIdentifier{}
	requests := []DeduplicatedMergeRequest{}
	tx.Info("Received %d Merge Requests", len(sqsEvent.Records))
	tx.Debug("1. Processing and validating all requests")
	for _, rec := range sqsEvent.Records {
		rq := validateMergeRequest(tx, rec, dlqConfig)
		if rq.TargetID != "" {
			tx.Debug("Valid merge request. Adding to requests to process")
			requests = append(requests, rq)
		}
	}
	if len(requests) == 0 {
		tx.Info("No valid merge requests found. Exiting without retry")
		return map[string][]ItemIdentifier{}, nil
	}
	tx.Debug("2. Organizing by domain")
	requestsByDomain := organizeByDomain(requests)
	tx.Debug("Found request for %d different domains", len(requestsByDomain))
	// Process each domain in parallel

	for domain, requests := range requestsByDomain {
		tx.Debug("processing %d requests for domain '%s'", len(requests), domain)
		domainTx := core.NewTransaction(TX_MERGER, tx.TransactionID, LOG_LEVEL)
		domainTx.Debug("Deduplicating requests")
		dedupedrequests := deduplicate(tx, requests)
		domainTx.Debug("Initializing LCS handler for domain %s", domain)
		customCfg := initStorage()
		customCfg.SetTx(tx)
		customCfg.SetDomain(domain)
		domainTx.Debug("Processing %d requests for domain %s", len(dedupedrequests), domain)
		failures, err := mergeProfiles(domainTx, customCfg, dedupedrequests)
		if err != nil {
			log.Printf("Error processing one or multiple merge requests: %+v", err)
			batchItemFailures = append(batchItemFailures, failures...)
		}
	}
	sqsBatchResponse := map[string][]ItemIdentifier{
		"batchItemFailures": batchItemFailures,
	}
	if len(batchItemFailures) > 0 {
		log.Printf("lambda complete with %d failures", len(batchItemFailures))
	}
	return sqsBatchResponse, nil
}

func validateMergeRequest(tx core.Transaction, rec events.SQSMessage, dlqConfig sqs.IConfig) DeduplicatedMergeRequest {
	tx.Debug("Validating merge request: %+v", rec)
	var mergeRequest customerprofiles.MergeRequest
	err := json.Unmarshal([]byte(rec.Body), &mergeRequest)
	if err != nil {
		tx := core.NewTransaction(TX_MERGER, "", LOG_LEVEL)
		tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")
		tx.Error("Error unmarshalling SQS event")
		processingError := ProcessingError{
			Error: err,
			Data:  rec.Body,
		}
		logProcessingError(tx, processingError, dlqConfig)
		return DeduplicatedMergeRequest{}
	}
	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")
	tx.Debug("Received Merge Request: %v", mergeRequest)
	// Validate Domain Name is Valid
	err = customerprofiles.ValidateDomainName(mergeRequest.DomainName)
	if err != nil {
		tx.Error("Error validating domain name in SQS event: %v", err)
		data, err2 := json.Marshal(mergeRequest)
		str := string(data)
		if err2 != nil {
			str = getMarshalErrorString(err)
		}
		processingError := ProcessingError{
			Error: err,
			Data:  str,
		}
		logProcessingError(tx, processingError, dlqConfig)
		return DeduplicatedMergeRequest{}
	}

	// Validate Source ID and Target ID are Valid
	if !isValidUUID(mergeRequest.SourceID) {
		tx.Error("Error source UUID")
		data, err := json.Marshal(mergeRequest)
		str := string(data)
		if err != nil {
			str = getMarshalErrorString(err)
		}
		processingError := ProcessingError{
			Error: errors.New("invalid source id"),
			Data:  str,
		}
		logProcessingError(tx, processingError, dlqConfig)
		return DeduplicatedMergeRequest{}
	}
	if !isValidUUID(mergeRequest.TargetID) {
		tx.Error("Error target UUID")
		data, err := json.Marshal(mergeRequest)
		str := string(data)
		if err != nil {
			str = getMarshalErrorString(err)
		}
		processingError := ProcessingError{
			Error: errors.New("invalid target id"),
			Data:  str,
		}
		logProcessingError(tx, processingError, dlqConfig)
		return DeduplicatedMergeRequest{}
	}

	// Validate Context
	err = validateMergeContext(mergeRequest.Context)
	if err != nil {
		tx.Error("Error validating context: %v", err)
		data, err2 := json.Marshal(mergeRequest)
		str := string(data)
		if err2 != nil {
			str = getMarshalErrorString(err2)
		}
		processingError := ProcessingError{
			Error: err,
			Data:  str,
		}
		logProcessingError(tx, processingError, dlqConfig)
		return DeduplicatedMergeRequest{}
	}

	return DeduplicatedMergeRequest{
		TargetID:   mergeRequest.TargetID,
		SourceID:   mergeRequest.SourceID,
		DomainName: mergeRequest.DomainName,
		Context:    mergeRequest.Context,
		MsgID:      rec.MessageId,
	}
}

type DeduplicatedMergeRequest struct {
	TargetID   string                               `json:"targetId"`
	SourceID   string                               `json:"sourceId"`
	DomainName string                               `json:"domainName"`
	Context    customerprofiles.ProfileMergeContext `json:"context"`
	MsgID      string                               `json:"msgId"`
}

func organizeByDomain(request []DeduplicatedMergeRequest) map[string][]DeduplicatedMergeRequest {
	byDomain := make(map[string][]DeduplicatedMergeRequest)
	for _, r := range request {
		byDomain[r.DomainName] = append(byDomain[r.DomainName], r)
	}
	return byDomain
}

func deduplicate(tx core.Transaction, requests []DeduplicatedMergeRequest) []DeduplicatedMergeRequest {
	tx.Debug("Deduplicating %d Merge Requests", len(requests))
	deduped := []DeduplicatedMergeRequest{}
	dedupedMap := make(map[string]DeduplicatedMergeRequest)
	// Deduplicate by source and target ID
	for _, r := range requests {
		key := r.SourceID + "-" + r.TargetID
		dedupedMap[key] = r
	}
	// Convert map to slice
	for _, v := range dedupedMap {
		deduped = append(deduped, v)
	}
	tx.Debug("%d Merge Requests remains after deduping (found %d dupes)", len(deduped), len(requests)-len(deduped))
	// Deduplicate by domain
	return deduped
}

func mergeProfiles(tx core.Transaction, cp customerprofiles.ICustomerProfileLowCostConfig, deduplicatedMargeRequests []DeduplicatedMergeRequest) ([]ItemIdentifier, error) {
	errors := []error{}
	tx.Debug("Processing deduplicated %d Merge Requests", len(deduplicatedMargeRequests))
	batchItemFailures := []ItemIdentifier{}
	batchItemFailuresMyMsgId := map[string]ItemIdentifier{}

	for _, mergeRequest := range deduplicatedMargeRequests {
		tx.Debug("Starting merge for profile %s and %s", mergeRequest.TargetID, mergeRequest.SourceID)
		// TODO: When implementing merge optimization to merge the smaller profile into larger profile, update this look up to use 1 query instead of two.
		sourceConnectId, err := cp.FindCurrentParentOfProfile(mergeRequest.DomainName, mergeRequest.SourceID)
		if err != nil {
			tx.Error("Error finding current parent of profile %s", mergeRequest.SourceID)
			errors = append(errors, err)
			batchItemFailuresMyMsgId[mergeRequest.MsgID] = ItemIdentifier{ItemIdentifier: mergeRequest.MsgID}
			continue
		}
		targetConnectId, err := cp.FindCurrentParentOfProfile(mergeRequest.DomainName, mergeRequest.TargetID)
		if err != nil {
			tx.Error("Error finding current parent of profile %s", mergeRequest.TargetID)
			errors = append(errors, err)
			batchItemFailuresMyMsgId[mergeRequest.MsgID] = ItemIdentifier{ItemIdentifier: mergeRequest.MsgID}
			continue
		}
		if sourceConnectId == targetConnectId {
			tx.Info("Profiles are already merged, skipping request")
			continue
		}
		_, err = cp.MergeProfiles(targetConnectId, sourceConnectId, mergeRequest.Context)
		if err != nil {
			tx.Error("Error merging profiles %s and %s", mergeRequest.TargetID, mergeRequest.SourceID)
			errors = append(errors, err)
			batchItemFailuresMyMsgId[mergeRequest.MsgID] = ItemIdentifier{ItemIdentifier: mergeRequest.MsgID}
		}
	}

	for _, failsTyRetry := range batchItemFailuresMyMsgId {
		batchItemFailures = append(batchItemFailures, failsTyRetry)
	}

	if len(errors) > 0 {
		tx.Error("%d errors while merging profiles: %+v", errors)
	}
	return batchItemFailures, core.ErrorsToError(errors)
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func validateMergeContext(ctx customerprofiles.ProfileMergeContext) error {
	mergeType := ctx.MergeType
	if mergeType == "" {
		return errors.New("merge type is required")
	}
	if mergeType == customerprofiles.MergeTypeAI {
		if ctx.ConfidenceUpdateFactor < 0 || ctx.ConfidenceUpdateFactor > 1 {
			return errors.New("confidence update factor is invalid")
		}
	} else if mergeType == customerprofiles.MergeTypeRule {
		ruleId, err := strconv.ParseInt(ctx.RuleID, 10, 64)
		if err != nil || ruleId < 0 {
			return errors.New("rule id is invalid")
		}
		if ctx.RuleSetVersion == "" {
			return errors.New("rule set version is required")
		}
	} else if mergeType == customerprofiles.MergeTypeManual {
		if ctx.OperatorID == "" {
			return errors.New("operator id is required")
		}
	} else {
		return errors.New("invalid merge type")
	}
	return nil
}

func logProcessingError(tx core.Transaction, errObj ProcessingError, dlqConfig sqs.IConfig) {
	err := dlqConfig.SendWithStringAttributes(errObj.Data, map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: "UCP merge error",
		common.SQS_MES_ATTR_MESSAGE:        errObj.Error.Error(),
		common.SQS_MES_ATTR_TX_ID:          tx.TransactionID,
	})
	if err != nil {
		tx.Error("Could not log message to SQS queue. Error: %v", err)
	}
}

func getMarshalErrorString(err error) string {
	return "Could not marshal data. error: " + err.Error()
}

func initStorage() customerprofiles.ICustomerProfileLowCostConfig {
	options := customerprofiles.CustomerProfileInitOptions{
		MergeQueueClient: &mergeQueue,
	}
	return customerprofiles.InitLowCost(
		LAMBDA_REGION,
		auroraClient,
		&configTableClient,
		kinesisClient,
		&cpWriterQueue,
		METRICS_NAMESPACE,
		METRICS_SOLUTION_ID,
		METRICS_SOLUTION_VERSION,
		LOG_LEVEL,
		&options,
		LCS_CACHE,
	)
}

func main() {
	lambda.Start(HandleRequest)
}
