// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

/*********************************************************************************************************************
 *  Copyright 2023 Amazon.com, Inc. or its affiliates. All Rights Reserved.                                           *
 *                                                                                                                    *
 *  Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance    *
 *  with the License. A copy of the License is located at                                                             *
 *                                                                                                                    *
 *      http://www.apache.org/licenses/LICENSE-2.0                                                                    *
 *                                                                                                                    *
 *  or in the 'license' file accompanying this file. This file is distributed on an 'AS IS' BASIS, WITHOUT WARRANTIES *
 *  OR CONDITIONS OF ANY KIND, express or implied. See the License for the specific language governing permissions    *
 *  and limitations under the License.                                                                                *
 *********************************************************************************************************************/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/aurora"
	"tah/upt/source/tah-core/awssolutions"
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	secrets "tah/upt/source/tah-core/secret"
	common "tah/upt/source/ucp-common/src/constant/admin"
	utils "tah/upt/source/ucp-common/src/utils/utils"
	"time"

	kinesis "tah/upt/source/tah-core/kinesis"
	sqs "tah/upt/source/tah-core/sqs"

	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	awsLambda "github.com/aws/aws-lambda-go/lambda"
)

type Event struct {
	TxID           string
	Connector      string
	Usecase        string
	IngestionError kinesis.IngestionError
	OriginalEvent  string
}

// Environment Variable passed from infrastructure
var LAMBDA_REGION = os.Getenv("LAMBDA_REGION")
var INPUT_STREAM = os.Getenv("INPUT_STREAM")
var DLQ_URL = os.Getenv("DEAD_LETTER_QUEUE_URL")

// Solution Info
var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
var METRICS_UUID = os.Getenv("METRICS_UUID")
var SEND_ANONYMIZED_DATA = os.Getenv("SEND_ANONYMIZED_DATA")
var METRICS_NAMESPACE = "upt"

// Low cost storage
var AURORA_PROXY_ENDPOINT = os.Getenv("AURORA_PROXY_ENDPOINT")
var AURORA_PROXY_READONLY_ENDPOINT = os.Getenv("AURORA_PROXY_READONLY_ENDPOINT")
var AURORA_DB_NAME = os.Getenv("AURORA_DB_NAME")
var AURORA_DB_SECRET_ARN = os.Getenv("AURORA_DB_SECRET_ARN")
var STORAGE_CONFIG_TABLE_NAME = os.Getenv("STORAGE_CONFIG_TABLE_NAME")
var STORAGE_CONFIG_TABLE_PK = os.Getenv("STORAGE_CONFIG_TABLE_PK")
var STORAGE_CONFIG_TABLE_SK = os.Getenv("STORAGE_CONFIG_TABLE_SK")
var CHANGE_PROC_KINESIS_STREAM_NAME = os.Getenv("CHANGE_PROC_KINESIS_STREAM_NAME")
var MERGE_QUEUE_URL = os.Getenv("MERGE_QUEUE_URL")
var CP_WRITER_QUEUE_URL = os.Getenv("CP_WRITER_QUEUE_URL")
var LCS_CACHE = customerprofileslcs.InitLcsCache()

var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

// Init low-cost clients (optimally, this is conditional for low-cost storage)
var smClient = secrets.InitWithRegion(AURORA_DB_SECRET_ARN, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var user, pwd = getUserCredentials(smClient)
var auroraClient = aurora.InitPostgresql(AURORA_PROXY_ENDPOINT, AURORA_DB_NAME, user, pwd, 50, LOG_LEVEL)
var auroraReadOnlyClient = aurora.InitPostgresql(AURORA_PROXY_READONLY_ENDPOINT, AURORA_DB_NAME, user, pwd, 50, LOG_LEVEL)

var configTableClient = db.Init(
	STORAGE_CONFIG_TABLE_NAME,
	STORAGE_CONFIG_TABLE_PK,
	STORAGE_CONFIG_TABLE_SK,
	METRICS_SOLUTION_ID,
	METRICS_SOLUTION_VERSION,
)
var kinesisClient = kinesis.Init(CHANGE_PROC_KINESIS_STREAM_NAME, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, LOG_LEVEL)
var mergeQueue = sqs.InitWithQueueUrl(MERGE_QUEUE_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var cpWriterQueue = sqs.InitWithQueueUrl(CP_WRITER_QUEUE_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

var sqsConfig = sqs.InitWithQueueUrl(DLQ_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var solutionsConfig = awssolutions.Init(METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, METRICS_UUID, SEND_ANONYMIZED_DATA, LOG_LEVEL)

type ProcessingError struct {
	Error      error
	Data       string
	AccpRecord string
	TxID       string
}

type BusinessObjectRecord struct {
	TravellerID        string                    `json:"travellerId"`
	TxID               string                    `json:"transactionId"`
	Domain             string                    `json:"domain"`
	Mode               string                    `json:"mode"`
	MergeModeProfileID string                    `json:"mergeModeProfileId"`
	PartialModeOptions PartialModeOptions        `json:"partialModeOptions"`
	ObjectType         string                    `json:"objectType"`
	ModelVersion       string                    `json:"modelVersion"`
	Data               []interface{}             `json:"data"`
	KinesisRecord      events.KinesisEventRecord // original record required when sending to error queue
}

type PartialModeOptions struct {
	Fields []string `json:"fields"`
}

func HandleRequest(ctx context.Context, req events.KinesisEvent) error {
	return HandleRequestWithServices(ctx, req, sqsConfig, solutionsConfig)
}

var MAX_CONCURRENCY_LIMIT = 1

func HandleRequestWithServices(ctx context.Context, req events.KinesisEvent, sqsc sqs.Config, solutionsConfig awssolutions.IConfig) error {
	tx := core.NewTransaction("ucp-real-time", "", LOG_LEVEL)
	tx.Info("Received %v Kinesis Data Stream Events", len(req.Records))
	now := time.Now()

	tx.Debug("Parsing kinesis records")
	businessObjectRecords := parseKinesisRecords(tx, req.Records, sqsc)
	tx.Debug("Grouping records by domain")
	groupedBusinessRecords := groupByDomain(businessObjectRecords)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error
	for domain, busObjRecords := range groupedBusinessRecords {
		wg.Add(1)
		go func(dom string, reqs []BusinessObjectRecord) {
			defer wg.Done()
			tx.Info("Ingesting %d records for domain %s", len(reqs), dom)
			err := ingestRecords(tx, dom, reqs, sqsc)
			if err != nil {
				tx.Error("Throttling Error processing grouped records for domain %s: %v", dom, err)
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(domain, busObjRecords)
	}
	wg.Wait()
	if len(errs) > 0 {
		tx.Error("Error processing records: %+v", errs)
		return core.ErrorsToError(errs)
	}
	duration := time.Since(now)
	_, err := solutionsConfig.SendMetrics(map[string]interface{}{
		"service":      "ucp-real-time-transformer",
		"event":        "event-processed",
		"record_count": len(req.Records),
		"duration":     duration.Milliseconds(),
	})
	if err != nil {
		tx.Warn("Error sending metrics %v", err)
	}
	return nil
}

func main() {
	awsLambda.Start(HandleRequest)
}

func parseKinesisRecords(tx core.Transaction, kinesisRecords []events.KinesisEventRecord, sqsc sqs.Config) []BusinessObjectRecord {
	businessObjectRecords := []BusinessObjectRecord{}
	for _, rec := range kinesisRecords {
		tx.Debug("Processing Kinesis record")
		bizObj, err := parseBusinessRecord(tx, rec, sqsc)
		if err == nil {
			businessObjectRecords = append(businessObjectRecords, bizObj)
		}
	}

	return businessObjectRecords
}

func groupByDomain(businessObjectRecords []BusinessObjectRecord) map[string][]BusinessObjectRecord {
	// Group business records by traveller_id
	grouped := map[string][]BusinessObjectRecord{}
	for _, rec := range businessObjectRecords {
		grouped[rec.Domain] = append(grouped[rec.Domain], rec)
	}

	return grouped
}

func ingestRecords(tx core.Transaction, domain string, businessObjectRecords []BusinessObjectRecord, sqsc sqs.Config) error {
	var wg sync.WaitGroup
	//we use this channel to control concurrency
	// see https://encore.dev/blog/advanced-go-concurrency
	tx.Debug("Initializing LCS handler for domain %s", domain)
	customCfg := initStorage()
	err := customCfg.SetDomain(domain)
	customCfg.SetTx(tx)
	sem := make(chan struct{}, MAX_CONCURRENCY_LIMIT)
	tx.Info("Ingesting %v business object records for domain %s", len(businessObjectRecords), domain)
	//in case of throttling error. we want to reprocess the stream
	var throttlingError error
	for _, rec := range businessObjectRecords {
		if err != nil {
			logProcessingError(tx, ProcessingError{Error: err, Data: string(rec.KinesisRecord.Kinesis.Data), AccpRecord: rec.ObjectType}, sqsc)
			continue
		}
		for _, dataRec := range rec.Data {
			sem <- struct{}{}
			wg.Add(1)
			go func(bizRec BusinessObjectRecord, data interface{}) {
				defer wg.Done()
				//processDataRecord only return throttling errors
				err := processDataRecord(tx, bizRec, data, sqsc, customCfg)
				if err != nil {
					throttlingError = err
				}
				<-sem
			}(rec, dataRec)
		}
	}

	wg.Wait()
	return throttlingError
}

func processDataRecord(
	tx core.Transaction,
	rec BusinessObjectRecord,
	dataRec interface{},
	sqsc sqs.Config,
	customCfg customerprofileslcs.ICustomerProfileLowCostConfig,
) error {
	//1. Parsing data record
	accpRecString, accpObjectType, accpObjectId, accpObjectTravellerId, accpMap, err := parseAccpRecord(tx, dataRec)
	tx.Info("Ingesting accp record of type %v in profile with traveler ID %v", accpObjectType, accpObjectTravellerId)
	if err != nil {
		tx.Error("error parsing profile object %v", err)
		logMetric(tx, solutionsConfig, "error", "error parsing profile object: "+err.Error())
		logProcessingError(tx, ProcessingError{Error: err, Data: string(rec.KinesisRecord.Kinesis.Data), AccpRecord: rec.ObjectType}, sqsc)
		return nil
	}

	//If we are in the merge case we will merge the incoming profile (if pax of guest) into the parent profile provided and return
	if rec.Mode == common.INGESTION_MODE_MERGE {
		tx.Debug("Ingestion mode: merge")
		if accpObjectType == common.ACCP_RECORD_GUEST_PROFILE || accpObjectType == common.ACCP_RECORD_PAX_PROFILE {
			err = handleMergeCase(tx, rec.MergeModeProfileID, accpObjectTravellerId, customCfg)
			if err != nil {
				tx.Error("Error handling merge case %v", err)
				logProcessingError(
					tx,
					ProcessingError{Error: err, Data: string(rec.KinesisRecord.Kinesis.Data), AccpRecord: rec.ObjectType},
					sqsc,
				)
			}
		} else {
			tx.Info("Profile object is not a guest or pax profile, skipping")
		}
		return nil
	}

	//Else we ingest the record (Either in partial mode or full mode)
	tx.Debug("Putting profile object of type '%v' to ACCP in domain %s", accpObjectType, rec.Domain)
	if rec.Mode == common.INGESTION_MODE_PARTIAL {
		tx.Debug("Ingestion mode: partial")
		partialModeOptions := rec.PartialModeOptions
		fields := partialModeOptions.Fields
		overrideFields := utils.GetOverrideFields(accpObjectType, fields)
		if len(overrideFields) == 0 {
			tx.Debug("No fields specified, nothing to override for this object")
			return nil
		}
		accpRecString, err = handlePartialCase(tx, accpObjectType, accpObjectId, accpObjectTravellerId, accpMap, customCfg, fields)
		if err != nil {
			tx.Error("error handling partial case %v, object does not exist, continuing ingestion", err)
		}
	}
	return ingestProfileObject(tx, accpRecString, accpObjectType, rec.Domain, customCfg, sqsc, accpObjectTravellerId)
}

func handleMergeCase(
	tx core.Transaction,
	travelerIdToMergeInto string,
	travelerIdToMerge string,
	customCfg customerprofileslcs.ICustomerProfileLowCostConfig,
) error {
	tx.Info("Merging profile tid=%v into profile with tid=%v", travelerIdToMerge, travelerIdToMergeInto)
	if travelerIdToMergeInto == "" {
		return errors.New("Cannot process merge case as profileIdToMergeInto is empty")
	}
	accpProfileIdToMergeInto, err := customCfg.GetProfileId(common.ACCP_PROFILE_ID_KEY, travelerIdToMergeInto)
	if err != nil {
		tx.Error("error getting profile id for profile with travelerIdToMergeInto=%v", travelerIdToMergeInto)
		return err
	}
	accpProfileIdToMerge, err := customCfg.GetProfileId(common.ACCP_PROFILE_ID_KEY, travelerIdToMerge)
	if err != nil {
		tx.Error("error getting profile id for profile with travelerIdToMerge=%v", travelerIdToMerge)
		return err
	}
	_, err = customCfg.MergeProfiles(accpProfileIdToMergeInto, accpProfileIdToMerge, customerprofileslcs.ProfileMergeContext{
		MergeType: customerprofileslcs.MergeTypeManual,
	})
	if err != nil {
		tx.Error(
			"Error merging profiles with profileIdToMergeInto=%v and profileIdToMerge=%v",
			accpProfileIdToMergeInto,
			accpProfileIdToMerge,
		)
	}
	return err
}

func ingestProfileObject(
	tx core.Transaction,
	accpRecString, accpObjectType, domain string,
	customCfg customerprofileslcs.ICustomerProfileLowCostConfig,
	sqsc sqs.Config,
	travelerId string,
) error {
	tx.Info("Putting profile object of type '%v' with traveler Id %v into domain %s", accpObjectType, travelerId, domain)
	err := customCfg.PutProfileObject(accpRecString, accpObjectType)
	tx.Info("Putting profile object of type '%v' with traveler Id %v domain %s", accpObjectType, travelerId, domain)
	if err != nil {
		tx.Warn("error putting profile object %v", err)
		if customCfg.IsThrottlingError(err) {
			//if the errors is a throttling error we fail the function to allow  for retries
			tx.Error("Throttling error: failing function to allow retry")
			return err
		}

		if isRetryableError(err) {
			tx.Error("Retryable error: we will retry")
			return err
		}

		tx.Error("Not a retryable error: sending error message to DLQ")
		logProcessingError(tx, ProcessingError{Error: err, Data: accpRecString, AccpRecord: accpObjectType}, sqsc)
		logMetric(tx, solutionsConfig, "error", "error putting profile object: "+err.Error())
		return nil
	}
	logMetric(tx, solutionsConfig, "success", "successfully put profile object")
	return nil
}

func parseBusinessRecord(tx core.Transaction, rec events.KinesisEventRecord, sqsc sqs.Config) (BusinessObjectRecord, error) {
	businessRecord := BusinessObjectRecord{}
	tx.Debug("Parsing Kinesis Data")
	err := json.Unmarshal(rec.Kinesis.Data, &businessRecord)
	if err != nil {
		tx.Error("Unexpected format for record inputted %v", err)
		logMetric(tx, solutionsConfig, "error", "unexpected format for record: "+err.Error())
		logProcessingError(tx, ProcessingError{Error: err, Data: string(rec.Kinesis.Data)}, sqsc)
		return BusinessObjectRecord{}, err
	}
	//setting transaction ID based on message received
	if businessRecord.TxID != "" {
		tx.TransactionID = businessRecord.TxID
	}
	businessRecord.KinesisRecord = rec
	tx.Debug("Parsed: %v", businessRecord)
	return businessRecord, nil
}

func handlePartialCase(
	tx core.Transaction,
	accpObjectType, accpObjectId, accpObjectTravellerId string,
	accpMap map[string]interface{},
	customCfg customerprofileslcs.ICustomerProfileLowCostConfig,
	fields []string,
) (string, error) {
	accpMapBytes, err := json.Marshal(accpMap)
	if err != nil {
		tx.Error("error marshalling profile object %v", err)
		logMetric(tx, solutionsConfig, "error", "error marshalling profile object: "+err.Error())
		return "", err
	}
	accpProfileId, err := customCfg.GetProfileId(common.ACCP_PROFILE_ID_KEY, accpObjectTravellerId)
	if err != nil {
		tx.Error("error getting profile id %v", err)
		logMetric(tx, solutionsConfig, "error", "error getting profile id: "+err.Error())
		return string(accpMapBytes), err
	}
	if accpProfileId == "" {
		err = errors.New("no profile id found for accpObjectTravellerId =" + accpObjectTravellerId)
		return string(accpMapBytes), err
	}
	profObject, err := customCfg.GetSpecificProfileObject(common.ACCP_OBJECT_ID_KEY, accpObjectId, accpProfileId, accpObjectType)
	if err != nil {
		tx.Error("error getting profile object %v", err)
		logMetric(tx, solutionsConfig, "error", "error getting profile object: "+err.Error())
		return string(accpMapBytes), err
	}
	if profObject.ID == "" {
		return string(accpMapBytes), fmt.Errorf("no profile found for object %s", accpObjectId)
	}
	newMap := utils.DiffAndCombineAccpMaps(accpMap, profObject.AttributesInterface, accpObjectType, fields)
	delete(newMap, "profile_id")
	delete(newMap, "connect_id")
	delete(newMap, "timestamp")
	delete(newMap, "overall_confidence_score")

	newMapBytes, err := json.Marshal(newMap)
	if err != nil {
		tx.Error("error marshalling profile object %v", err)
		logMetric(tx, solutionsConfig, "error", "error marshalling profile object: "+err.Error())
		return "", err
	}
	accpRecString := string(newMapBytes)
	return accpRecString, nil

}

func logMetric(tx core.Transaction, solutionsConfig awssolutions.IConfig, status string, msg string) {
	_, err := solutionsConfig.SendMetrics(map[string]interface{}{
		"service": "ucp-real-time-transformer",
		"event":   "process-record",
		"status":  status,
		"message": msg,
	})
	if err != nil {
		tx.Warn("Error sending metrics %v", err)
	}
}

func parseAccpRecord(tx core.Transaction, accpRec interface{}) (string, string, string, string, map[string]interface{}, error) {
	accpRecByte, err2 := json.Marshal(accpRec)
	if err2 != nil {
		tx.Error("Error marshalling ACCP object: %v", err2)
		return "", "", "", "", map[string]interface{}{}, err2
	}
	accpObjectAsMap, ok := accpRec.(map[string]interface{})
	if !ok {
		tx.Error("Error casting Accp object")
		return "", "", "", "", map[string]interface{}{}, errors.New("could not infer accp map ")
	}
	accpObjectType, ok := accpObjectAsMap[common.ACCP_OBJECT_TYPE_KEY].(string)
	if !ok {
		tx.Error("Error casting Accp object could not infer object type")
		return "", "", "", "", accpObjectAsMap, errors.New("could not infer accp object type ")
	}
	accpObjectId, ok := accpObjectAsMap[common.ACCP_OBJECT_ID_KEY].(string)
	if !ok {
		tx.Error("Error casting Accp object could not infer object id")
		return "", "", "", "", accpObjectAsMap, errors.New("could not infer accp object id ")
	}
	accpObjectTravellerId, ok := accpObjectAsMap[common.ACCP_OBJECT_TRAVELLER_ID_KEY].(string)
	if !ok {
		tx.Error("Error casting Accp object could not infer object traveller id")
		return "", "", "", "", accpObjectAsMap, errors.New("could not infer accp object traveller id")
	}
	return string(accpRecByte), accpObjectType, accpObjectId, accpObjectTravellerId, accpObjectAsMap, nil
}

func logProcessingError(tx core.Transaction, errObj ProcessingError, sqsc sqs.Config) {
	accpRec := errObj.AccpRecord
	if accpRec == "" {
		accpRec = "Unknown"
	}
	err := sqsc.SendWithStringAttributes(errObj.Data, map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE:            "ACP real time Ingestion error",
		common.SQS_MES_ATTR_BUSINESS_OBJECT_TYPE_NAME: accpRec,
		common.SQS_MES_ATTR_MESSAGE:                   errObj.Error.Error(),
		common.SQS_MES_ATTR_TX_ID:                     tx.TransactionID,
	})
	if err != nil {
		tx.Error("Could not log message to SQS queue %v. Error: %v", sqsc.QueueUrl, err)
	}
}

func initStorage() customerprofileslcs.ICustomerProfileLowCostConfig {
	var customCfg customerprofileslcs.ICustomerProfileLowCostConfig

	options := customerprofileslcs.CustomerProfileInitOptions{
		MergeQueueClient: &mergeQueue,
		AuroraReadOnly:   auroraReadOnlyClient,
	}
	customCfg = customerprofileslcs.InitLowCost(
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

	return customCfg
}

func getUserCredentials(sm secrets.Config) (string, string) {
	log.Printf("Retrieving DB credentials from Secrets Manager: %v", sm.SecretArn)
	user := sm.Get("username")
	pwd := sm.Get("password")
	return user, pwd
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "SQLSTATE") || customerprofileslcs.IsRetryableError(err)
}
