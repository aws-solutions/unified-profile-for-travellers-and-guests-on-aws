// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/aurora"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	secrets "tah/upt/source/tah-core/secret"
	"tah/upt/source/tah-core/sqs"
	model "tah/upt/source/ucp-common/src/model/admin"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoEvent struct {
	Records []DynamoEventRecord `json:"records"`
}

type DynamoEventRecord struct {
	EventID   string            `json:"eventID"`
	EventName string            `json:"eventName"`
	Change    DynamoEventChange `json:"dynamodb"`
	// more fields if needed: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_streams_GetRecords.html
}

type DynamoEventChange struct {
	NewImage map[string]*dynamodb.AttributeValue `json:"NewImage"`
	OldImage map[string]*dynamodb.AttributeValue `json:"OldImage"`
}

type RetryData struct {
	RetryObject RetryObject
	ErrorObject model.UcpIngestionError
}

type RetryObject struct {
	Traveller  string `json:"traveller_id"`
	Object     string `json:"object_id"`
	RetryCount int    `json:"retryCount"`
	TTL        int    `json:"ttl"`
}

// Event Names
const (
	EVENT_NAME_INSERT string = "INSERT"
	EVENT_NAME_MODIFY string = "MODIFY"
	EVENT_NAME_REMOVE string = "REMOVE"
)

// Retry Delays
const (
	FIRST_RECORD_DELAY      = time.Second * 30
	SEQUENTIAL_RECORD_DELAY = time.Second * 5
)

// Table Info
var RETRY_TABLE = os.Getenv("RETRY_TABLE")
var RETRY_PK = os.Getenv("RETRY_PK")
var RETRY_SK = os.Getenv("RETRY_SK")
var ERROR_TABLE = os.Getenv("ERROR_TABLE")
var ERROR_PK = os.Getenv("ERROR_PK")
var ERROR_SK = os.Getenv("ERROR_SK")

// Solution Info
var LAMBDA_REGION = os.Getenv("LAMBDA_REGION")
var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
var SEND_ANONYMIZED_DATA = os.Getenv("SEND_ANONYMIZED_DATA")
var METRICS_UUID = os.Getenv("METRICS_UUID")
var METRICS_NAMESPACE = "upt"

// Low cost storage
var AURORA_PROXY_ENDPOINT = os.Getenv("AURORA_PROXY_ENDPOINT")
var AURORA_DB_NAME = os.Getenv("AURORA_DB_NAME")
var AURORA_DB_SECRET_ARN = os.Getenv("AURORA_DB_SECRET_ARN")
var STORAGE_CONFIG_TABLE_NAME = os.Getenv("STORAGE_CONFIG_TABLE_NAME")
var STORAGE_CONFIG_TABLE_PK = os.Getenv("STORAGE_CONFIG_TABLE_PK")
var STORAGE_CONFIG_TABLE_SK = os.Getenv("STORAGE_CONFIG_TABLE_SK")
var CHANGE_PROC_KINESIS_STREAM_NAME = os.Getenv("CHANGE_PROC_KINESIS_STREAM_NAME")
var MERGE_QUEUE_URL = os.Getenv("MERGE_QUEUE_URL")
var CP_WRITER_QUEUE_URL = os.Getenv("CP_WRITER_QUEUE_URL")
var LCS_CACHE = customerprofiles.InitLcsCache()

var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

// Init clients
var retryDb = db.Init(RETRY_TABLE, RETRY_PK, RETRY_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var errorDb = db.Init(ERROR_TABLE, ERROR_PK, ERROR_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var solutionsConfig = awssolutions.Init(METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, METRICS_UUID, SEND_ANONYMIZED_DATA, LOG_LEVEL)

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
var mergeQueue = sqs.InitWithQueueUrl(MERGE_QUEUE_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var cpWriterQueue = sqs.InitWithQueueUrl(CP_WRITER_QUEUE_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

var MAX_RETRY_COUNT = 3

var TX_RETRY = "ucp-retry"

func HandleRequest(ctx context.Context, event DynamoEvent) error {
	tx := core.NewTransaction(TX_RETRY, "", LOG_LEVEL)
	accpConfig := initStorage()
	retryDb.SetTx(tx)
	errorDb.SetTx(tx)
	accpConfig.SetTx(tx)
	return HandleRequestWithServices(ctx, event, tx, accpConfig, retryDb, errorDb, solutionsConfig)
}

func HandleRequestWithServices(
	ctx context.Context,
	event DynamoEvent,
	tx core.Transaction,
	accpConfig customerprofiles.ICustomerProfileLowCostConfig,
	retryDb, errorDb db.DBConfig,
	solutionsConfig awssolutions.IConfig,
) error {
	now := time.Now()
	tx.Debug("Processing %d records", len(event.Records))

	// Retry errors sequentially per traveller
	recordsByTraveller := parseRecords(event.Records, tx)
	err := retryByTraveller(recordsByTraveller, tx, accpConfig, retryDb, errorDb)
	if err != nil {
		tx.Error("Error while retrying: %s", err)
	}

	// Send metrics
	duration := time.Since(now).Milliseconds()
	recordCount := len(event.Records)
	_, err = solutionsConfig.SendMetrics(map[string]interface{}{
		"service":      TX_RETRY,
		"duration":     duration,
		"record_count": recordCount,
	})
	if err != nil {
		tx.Warn("Error while sending metrics to AWS Solutions: %s", err)
	}

	return nil
}

func parseRecords(records []DynamoEventRecord, tx core.Transaction) map[string]map[string]RetryData {
	recsByTraveller := map[string]map[string]RetryData{}
	for _, record := range records {
		// Ignore REMOVE event, we don't want to retry errors when they're being deleted
		if record.EventName == EVENT_NAME_REMOVE {
			continue
		}
		// Ignore errors that cannot be retried
		errRecord := model.UcpIngestionError{}
		err := dynamodbattribute.UnmarshalMap(record.Change.NewImage, &errRecord)
		if err != nil {
			tx.Error("Error unmarshalling record: %s", err)
			continue
		}
		tx.Error(
			"Parsing error type: '%s' from SQS. domain: '%s', travellerID: '%s', error message: '%s',  transaction ID: %v, bizObject: '%s' accpRec: '%s', recordId: '%s'",
			errRecord.Type,
			errRecord.Domain,
			errRecord.TravellerID,
			errRecord.Message,
			errRecord.TransactionID,
			errRecord.BusinessObjectType,
			errRecord.AccpRecordType,
			errRecord.RecordID,
		)
		if errRecord.TransactionID != "" {
			tx = core.NewTransaction(TX_RETRY, errRecord.TransactionID, LOG_LEVEL)
		}
		if errRecord.Category != model.ACCP_INGESTION_ERROR {
			tx.Error("Error category %s is not ACCP_INGESTION_ERROR. Skipping", errRecord.Category)
			continue
		}
		if errRecord.RecordID == "" {
			tx.Error("Could not find record ID. Skipping in record: %+v. Skipping", errRecord)
			continue
		}

		// Build retry data
		traveller := errRecord.TravellerID
		object := errRecord.AccpRecordType + "_" + errRecord.RecordID
		retryObject := RetryObject{
			Traveller: traveller,
			Object:    object,
		}
		record := RetryData{
			RetryObject: retryObject,
			ErrorObject: errRecord,
		}

		// Create map of objects for traveller
		_, ok := recsByTraveller[traveller]
		if !ok {
			recsByTraveller[traveller] = map[string]RetryData{}
		}
		recsByTraveller[traveller][object] = record
	}
	return recsByTraveller
}

func retryByTraveller(
	recordsByTraveller map[string]map[string]RetryData,
	tx core.Transaction,
	accpConfig customerprofiles.ICustomerProfileLowCostConfig,
	retryDb, errorDb db.DBConfig,
) error {
	// Concurrently retry records for each unique traveller
	var wg sync.WaitGroup
	wg.Add(len(recordsByTraveller))

	// Retry traveller objects sequentially
	for _, travellerRecords := range recordsByTraveller {
		go func(travellerRecords map[string]RetryData) {
			isFirstRecord := true
			hasMultipleRecords := len(travellerRecords) > 1
			for _, record := range travellerRecords {
				retry(record, isFirstRecord, hasMultipleRecords, tx, accpConfig, retryDb, errorDb)
				isFirstRecord = false
			}
			wg.Done()
		}(travellerRecords)
	}

	wg.Wait()
	return nil
}

func retry(
	record RetryData,
	isFirstRecord bool,
	hasMultipleRecords bool,
	tx core.Transaction,
	accpConfig customerprofiles.ICustomerProfileLowCostConfig,
	retryDb, errorDb db.DBConfig,
) {
	// Check if the object has already been retried
	err := retryDb.Get(record.RetryObject.Traveller, record.RetryObject.Object, &record.RetryObject)
	if err != nil {
		tx.Error("[%s] Error getting retry object: %s", record.RetryObject.Object, err)
		return
	}
	if record.RetryObject.Traveller == "" {
		tx.Error("[%s] Could not find retry object for traveller %s", record.RetryObject.Object, record.RetryObject.Traveller)
		return
	}
	if record.RetryObject.RetryCount >= MAX_RETRY_COUNT {
		tx.Error("[%s] Max retry count %d reached. stop retrying", record.RetryObject.Object, record.RetryObject.RetryCount)
		return
	}

	// Error is being retried, delete the existing error
	tx.Info("[%s] Error is being retried, delete the existing error", record.RetryObject)
	err = errorDb.DeleteByKey(record.ErrorObject.Type, record.ErrorObject.ID)
	if err != nil {
		tx.Error("[%s] Error deleting old error object: %s", record.ErrorObject, err)
	}

	// Update retry count
	tx.Debug("[%s] Update retry count: %d => %d", record.RetryObject, record.RetryObject.RetryCount, record.RetryObject.RetryCount+1)
	record.RetryObject.RetryCount++
	record.RetryObject.TTL = int(time.Now().Add(24 * time.Hour).Unix())
	_, err = retryDb.Save(record.RetryObject)
	if err != nil {
		tx.Error("[%s] Error updating retry: %s", record.RetryObject, err)
	}

	// Retry based on retry count
	accpConfig.SetDomain(record.ErrorObject.Domain)
	delay := setRetryDelay(record.RetryObject.RetryCount)
	tx.Debug("[%s] Retrying in %s", record.RetryObject, delay)
	time.Sleep(delay)
	err = accpConfig.PutProfileObject(record.ErrorObject.Record, record.ErrorObject.AccpRecordType)
	if err != nil {
		tx.Error("[%s] Error retrying %s", record.RetryObject, err)
	}
	tx.Info("[%s] Retrying complete", record.RetryObject)
	if isFirstRecord && hasMultipleRecords {
		// add buffer before addional profile objects for same traveller
		// in case Customer Profiles needs to create new profile
		time.Sleep(FIRST_RECORD_DELAY)
	}
}

// Retry delay based on
// https://aws.amazon.com/builders-library/timeouts-retries-and-backoff-with-jitter/
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
func setRetryDelay(retryCount int) time.Duration {
	base := 5000.0  // 5 seconds
	max := 300000.0 // 5 minutes
	delay := base * math.Pow(2, float64(retryCount))
	temp := math.Min(max, delay)
	equalJitter := temp/2 + float64(rand.Intn(int(temp)/2))
	return time.Duration(equalJitter) * time.Millisecond
}

func main() {
	lambda.Start(HandleRequest)
}

func initStorage() customerprofiles.ICustomerProfileLowCostConfig {
	var customCfg customerprofiles.ICustomerProfileLowCostConfig
	options := customerprofiles.CustomerProfileInitOptions{
		MergeQueueClient: &mergeQueue,
	}
	customCfg = customerprofiles.InitLowCost(
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
	user := sm.Get("username")
	pwd := sm.Get("password")
	return user, pwd
}
