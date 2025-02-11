// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/cloudwatch"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/s3"

	"tah/upt/source/tah-core/sqs"
	common "tah/upt/source/ucp-common/src/constant/admin"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Solution Info
var LAMBDA_REGION = os.Getenv("LAMBDA_REGION")
var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
var SEND_ANONYMIZED_DATA = os.Getenv("SEND_ANONYMIZED_DATA")
var METRICS_UUID = os.Getenv("METRICS_UUID")
var METRICS_NAMESPACE = "upt"

// Low cost storage

var BATCH_PROCESSOR_DLQ = os.Getenv("BATCH_PROCESSOR_DLQ")
var KINESIS_INGESTOR_STREAM_NAME = os.Getenv("KINESIS_INGESTOR_STREAM_NAME")
var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

var solutionsConfig = awssolutions.Init(METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, METRICS_UUID, SEND_ANONYMIZED_DATA, LOG_LEVEL)

var kinesisClient = kinesis.Init(KINESIS_INGESTOR_STREAM_NAME, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, LOG_LEVEL)
var sqsConfig = sqs.InitWithQueueUrl(BATCH_PROCESSOR_DLQ, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var metricLogger = cloudwatch.NewMetricLogger(METRICS_NAMESPACE)

var TX_BATCH = "ucp-batch"
var METRIC_N_CSV_ROWS_RECEIVED = "nCSVRowsReceived"
var METRIC_N_CSV_ROWS_FAILED = "nCSVRowsFailed"

var MAX_CONCURRENCY_LIMIT = 20

type BatchProcessingError struct {
	Error      error
	Data       string
	AccpRecord string
	TxID       string
}

type Request struct {
	Ctx             context.Context
	S3Event         events.S3Event
	Tx              core.Transaction
	SolutionsConfig awssolutions.IConfig
	Region          string
	Sqsc            sqs.Config
	SolId           string
	SolVersion      string
	MetricLogger    *cloudwatch.MetricLogger
	Domain          string
	KinesisClient   kinesis.IConfig
}

func HandleRequest(ctx context.Context, s3Event events.S3Event) error {
	tx := core.NewTransaction(TX_BATCH, "", LOG_LEVEL)
	tx.Debug("S3Event: %+v", s3Event)
	return HandleRequestWithServices(
		Request{
			Tx:              tx,
			S3Event:         s3Event,
			SolutionsConfig: solutionsConfig,
			Region:          LAMBDA_REGION,
			Sqsc:            sqsConfig,
			SolId:           METRICS_SOLUTION_ID,
			SolVersion:      METRICS_SOLUTION_VERSION,
			MetricLogger:    metricLogger,
			KinesisClient:   kinesisClient,
		},
	)
}

func HandleRequestWithServices(rq Request) error {
	s3Event := rq.S3Event
	tx := rq.Tx
	solConfg := rq.SolutionsConfig

	tx.Info("Processing %d records", len(s3Event.Records))
	for i, record := range s3Event.Records {
		csvData, objectType, domain := parseCsvData(tx, i, record, rq)
		rq.Domain = domain
		//we create the handler dynamically for every row to support a unique transaction ID per row
		if len(csvData) == 0 {
			continue
		}
		tx.Debug("[%d] 3-Ingesting Csv rows", i)
		successCount := 0
		//skipping header row
		headerFreeData := csvData[1:]
		headerRow := csvData[0]
		rq.MetricLogger.LogMetric(
			map[string]string{"domain": domain},
			METRIC_N_CSV_ROWS_RECEIVED,
			cloudwatch.Count,
			float64(len(headerFreeData)),
		)
		batSize := 500
		tx.Debug("[%d] 1-Processing %d rows in batch of %d", i, len(headerFreeData), batSize)
		batches := splitByBatch(headerFreeData, batSize)
		tx.Debug("[%d] 2- created %d batches", i, len(batches))
		for i, batch := range batches {
			errorCount := ingestInteractionBatch(tx, i, batch, objectType, headerRow, rq, record)
			rq.MetricLogger.LogMetric(map[string]string{"domain": domain}, METRIC_N_CSV_ROWS_FAILED, cloudwatch.Count, float64(errorCount))
		}
		//AWS solution custom metrics
		logMetric(tx, solConfg, successCount)
	}
	return nil
}

func splitByBatch(csvData [][]string, batchSize int) [][][]string {
	// Split the CSV data into batches of size `batchSize`
	batches := make([][][]string, 0)
	for i := 0; i < len(csvData); i += batchSize {
		end := i + batchSize
		if end > len(csvData) {
			end = len(csvData)
		}
		batches = append(batches, csvData[i:end])
	}
	return batches
}

func logMetric(tx core.Transaction, solConfg awssolutions.IConfig, successCount int) {
	if successCount > 0 {
		_, err := solConfg.SendMetrics(map[string]interface{}{
			"service":      TX_BATCH,
			"status":       "success",
			"record_count": successCount,
		})
		if err != nil {
			tx.Warn("Error while sending metrics to AWS Solutions: %s", err)
		}
	}
}
func logFailureMetric(tx core.Transaction, solConfg awssolutions.IConfig, err error) {
	if err != nil {
		_, err := solConfg.SendMetrics(map[string]interface{}{
			"service": TX_BATCH,
			"status":  "failed",
			"message": err.Error(),
		})
		if err != nil {
			tx.Warn("Error while sending metrics to AWS Solutions: %s", err)
		}
	}
}

func ingestInteractionBatch(tx core.Transaction, batchId int, csvData [][]string, objectType string, headerRow []string, rq Request, s3Rec events.S3EventRecord) int {
	tx.Info("Ingesting batch %d", batchId)
	sqsc := rq.Sqsc
	solConfg := rq.SolutionsConfig
	domain := rq.Domain
	kinesisBatch := []kinesis.Record{}
	invalidRows := [][]string{}
	errors := []error{}
	fileName := s3Rec.S3.Object.Key
	for i, row := range csvData {
		tx.Debug("Creating accp records for batch %d and row %d", batchId, i)
		rec, err := createAccpRecordFromCsv(tx, headerRow, row)
		if err != nil {
			tx.Error("Error while parsing row %d of csv batch %d: %v", i, batchId, err)
			invalidRows = append(invalidRows, row)
			errors = append(errors, err)
			continue
		}
		//this field '_model_version' is added during partitioning by the Glue ETL. We remove it to avoid LCS error.
		delete(rec, "_model_version")
		tx.Debug("Validating accp record for batch %d and row %d", batchId, i)
		err = validateRecord(tx, rec)
		if err != nil {
			tx.Error("Error while validating batch %d row %d: %v", batchId, i, err)
			invalidRows = append(invalidRows, row)
			errors = append(errors, err)
			continue
		}
		tx.Debug("Creating Kinesis record for row %d (batch %d) ", i, batchId)
		kinesisRec, err := createKinesisRecord(tx, rec, domain, objectType)
		if err != nil {
			tx.Error("[%d] Error while creating kinesis record for batch %d row %d: %v", batchId, i, err)
			invalidRows = append(invalidRows, row)
			errors = append(errors, err)
			continue
		}
		kinesisBatch = append(kinesisBatch, kinesisRec)
	}

	if len(invalidRows) == len(csvData) {
		tx.Error("Sending Invalid batch error %d. example error: %v", batchId, errors[0])
		logIngestionError(tx, fmt.Errorf("invalid batch %d. file %s format might be wrong. example error: '%v'", batchId, fileName, errors[0]), invalidRows[0], objectType, sqsc, solConfg)

	} else if len(invalidRows) > 0 {
		for i, err := range errors {
			tx.Error("Sending Partial batch errors %d: %v", batchId, invalidRows)
			logIngestionError(tx, fmt.Errorf("invalid row %d in batch %d of file %s. error: %v", i, batchId, fileName, err), invalidRows[i], objectType, sqsc, solConfg)
		}
	}

	tx.Info("Ingesting %d records into Kinesis (batch %d) ", len(kinesisBatch), batchId)
	err, errs := rq.KinesisClient.PutRecords(kinesisBatch)
	if err != nil {
		for _, ingErr := range errs {
			tx.Error("[%d] error putting batch %d into kinesis: %v", batchId, err)
			logIngestionError(tx, fmt.Errorf("error putting batch %d into Kinesis: %v", batchId, ingErr), kinesisBatch, objectType, sqsc, solConfg)
		}
	}
	tx.Info("Ingested %d records into Kinesis (batch %d) with %d ingestion errors", len(kinesisBatch), batchId, len(errs)+len(invalidRows))
	return len(errs) + len(invalidRows)
}

type BusinessObjectRecord struct {
	TravellerID  string        `json:"travellerId"`
	TxID         string        `json:"transactionId"`
	Domain       string        `json:"domain"`
	ObjectType   string        `json:"objectType"`
	ModelVersion string        `json:"modelVersion"`
	Data         []interface{} `json:"data"`
}

func createKinesisRecord(tx core.Transaction, rec map[string]string, domain, objectTypeName string) (kinesis.Record, error) {
	wrappedRecords := BusinessObjectRecord{
		TxID:       tx.TransactionID,
		ObjectType: objectTypeName,
		Domain:     domain,
		Data:       []interface{}{rec},
	}
	data, err := json.Marshal(wrappedRecords)
	if err != nil {
		return kinesis.Record{}, err
	}

	return kinesis.Record{
		Pk:   core.UUID(),
		Data: string(data),
	}, nil
}

func parseCsvData(tx core.Transaction, i int, record events.S3EventRecord, rq Request) ([][]string, string, string) {
	tx.Debug("[%d] 0-processing record %d", i)
	tx.Debug("[%d] 1-Extracting S3 key from record", i)
	region := rq.Region
	sqsc := rq.Sqsc
	solId := rq.SolId
	solVersion := rq.SolVersion
	s3Rec := record.S3
	sourceKey := s3Rec.Object.Key
	sourceBucket := s3Rec.Bucket.Name
	solConfg := rq.SolutionsConfig
	tx.Debug("[%d] 0-ingesting data in file: s3://%s/%s", i, s3Rec.Bucket.Name, s3Rec.Object.Key)

	if !strings.HasSuffix(sourceKey, ".csv") {
		tx.Info("Skipping non-CSV file: %s", sourceKey)
		return [][]string{}, "", ""
	}

	domain, objectType, err := extractDomainObjectTypeFromKey(tx, sourceKey)
	if err != nil {
		tx.Error("[%d] Error while extracting domain from S3 key '%s': ", i, sourceKey, err)
		logIngestionError(tx, err, record, objectType, sqsc, solConfg)
		return [][]string{}, objectType, ""
	}
	tx.Debug("[%d] Detected object type %s and domain %s", i, objectType, domain)

	s3Client := s3.Init(sourceBucket, "", region, solId, solVersion)

	tx.Debug("[%d] 2-Downloading file from S3 and parsing CSV", i)
	sourceKeyDecoded, err := url.QueryUnescape(sourceKey)
	if err != nil {
		tx.Error("[%d] Error while url-decoding S3 key: %v", i, err)
	}
	csvData, err := s3Client.ParseCsvFromS3(sourceKeyDecoded)
	if err != nil {
		tx.Error("[%d] Error parsing CSV from S3: %v", i, err)
		logIngestionError(tx, err, record, objectType, sqsc, solConfg)
		return [][]string{}, objectType, ""
	}
	if len(csvData) == 0 {
		tx.Info("[%d] No data to process. returning", i)
		logIngestionError(tx, fmt.Errorf("no data to process in file %s", sourceKey), record, objectType, sqsc, solConfg)
		return [][]string{}, objectType, ""
	}
	if len(csvData[0]) == 0 {
		tx.Error("[%d] Invalid data format (empty line). returning", i)
		logIngestionError(tx, errors.New("invalid data format (empty line)"), record, objectType, sqsc, solConfg)
		return [][]string{}, objectType, ""
	}
	tx.Debug("[%d] file: s3://%s/%s as %d rows (incl header)", i, s3Rec.Bucket.Name, s3Rec.Object.Key, len(csvData))
	return csvData, objectType, domain
}

func createAccpRecordFromCsv(tx core.Transaction, header, csvRow []string) (map[string]string, error) {
	rec := map[string]string{}
	if len(header) != len(csvRow) {
		tx.Error("Error while parsing csv row. Header length is not equal to csv row length")
		return rec, errors.New("error while parsing csv row. Header length is not equal to csv row length")
	}
	for i, fieldName := range header {
		rec[fieldName] = csvRow[i]
	}
	return rec, nil
}

func validateRecord(tx core.Transaction, record map[string]string) error {
	_, ok := record[common.ACCP_OBJECT_TYPE_KEY]
	if !ok {
		tx.Error("Error casting Accp object could not infer object type")
		return errors.New("could not infer accp object type")
	}
	_, ok = record[common.ACCP_OBJECT_ID_KEY]
	if !ok {
		tx.Error("Error casting Accp object could not infer object id")
		return errors.New("could not infer accp object id")
	}
	_, ok = record[common.ACCP_OBJECT_TRAVELLER_ID_KEY]
	if !ok {
		tx.Error("Error casting Accp object could not infer object traveler id")
		return errors.New("could not infer accp object traveler id")
	}
	tx.Debug("traveler_id: %s", record[common.ACCP_OBJECT_TRAVELLER_ID_KEY])
	tx.Debug("accpID: %s", record[common.ACCP_OBJECT_ID_KEY])
	tx.Debug("objectType: %s", record[common.ACCP_OBJECT_TYPE_KEY])
	return nil
}

func logIngestionError(
	tx core.Transaction,
	err error,
	data interface{},
	objectType string,
	sqsc sqs.Config,
	solConfg awssolutions.IConfig,
) {
	errToLog := err
	recStr, marshErr := json.Marshal(data)
	if marshErr != nil {
		tx.Error("Error while marshalling accp record for error logging: ", err)
		recStr = []byte(fmt.Sprintf("%+v", data))
		errToLog = marshErr
	}
	logProcessingError(tx, BatchProcessingError{Error: errToLog, Data: string(recStr), AccpRecord: objectType}, sqsc, solConfg)
}

func logProcessingError(tx core.Transaction, errObj BatchProcessingError, sqsc sqs.Config, solConfg awssolutions.IConfig) {
	accpRec := errObj.AccpRecord
	if accpRec == "" {
		accpRec = "Unknown"
	}
	err := sqsc.SendWithStringAttributes(errObj.Data, map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE:            "Batch ingestion error",
		common.SQS_MES_ATTR_BUSINESS_OBJECT_TYPE_NAME: accpRec,
		common.SQS_MES_ATTR_MESSAGE:                   errObj.Error.Error(),
		common.SQS_MES_ATTR_TX_ID:                     tx.TransactionID,
	})
	logFailureMetric(tx, solConfg, err)
	if err != nil {
		tx.Warn("Could not log message to SQS queue %v. Error: %v", sqsc.QueueUrl, err)
	}
}

func extractDomainObjectTypeFromKey(tx core.Transaction, key string) (string, string, error) {
	tx.Debug("[extractDomainObjectTypeFromKey] processing key %s", key)
	parts := strings.Split(key, "/")
	var domain, objectType string
	if len(parts) > 1 {
		domain = parts[0]
		objectType = parts[1]
	} else {
		return "", "", errors.New("Could not parse domain and business object from key " + key)
	}
	if domain == "" || objectType == "" {
		return "", "", errors.New("Could not parse domain and business object from key " + key)
	}
	tx.Debug("[extractDomainObjectTypeFromKey] found domain %s, objectType %s", domain, objectType)
	return domain, objectType, nil
}

func main() {
	lambda.Start(HandleRequest)
}
