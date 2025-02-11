// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"

	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/s3"
	constants "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/admin"
	services "tah/upt/source/ucp-common/src/services/admin"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type BusinessObjectRecord struct {
	Domain       string      `json:"domain"`
	ObjectType   string      `json:"objectType"`
	ModelVersion string      `json:"modelVersion"`
	Data         interface{} `json:"data"`
}

var LAMBDA_REGION = os.Getenv("AWS_REGION")
var INDUSTRY_CONNECTOR_BUCKET_NAME = os.Getenv("INDUSTRY_CONNECTOR_BUCKET_NAME")
var KINESIS_STREAM_NAME = os.Getenv("KINESIS_STREAM_NAME")
var DYNAMO_TABLE = os.Getenv("DYNAMO_TABLE")
var DYNAMO_PK = os.Getenv("DYNAMO_PK")
var DYNAMO_SK = os.Getenv("DYNAMO_SK")
var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

var dbClient = db.Init(DYNAMO_TABLE, DYNAMO_PK, DYNAMO_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var s3Client = s3.Init(INDUSTRY_CONNECTOR_BUCKET_NAME, "", LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var kinesisClient = kinesis.Init(KINESIS_STREAM_NAME, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, LOG_LEVEL)

var bucketToObjectMap = map[string]string{
	"bookings":    constants.BIZ_OBJECT_AIR_BOOKING,
	"clickstream": constants.BIZ_OBJECT_CLICKSTREAM,
	"profiles":    constants.BIZ_OBJECT_GUEST_PROFILE,
	"stays":       constants.BIZ_OBJECT_STAY_REVENUE,
}

func HandleRequest(ctx context.Context, s3Event events.S3Event) error {
	return HandleRequestWithServices(ctx, s3Event, dbClient, s3Client, kinesisClient)
}

func HandleRequestWithServices(ctx context.Context, s3Event events.S3Event, dbClient db.DBConfig, s3Client s3.S3Config, kinesisClient kinesis.IConfig) error {
	// Get linked domains
	domainList, err := services.GetLinkedDomains(dbClient)
	log.Printf("[IndustryConnectorTransfer] Account has %d ACCP domains", len(domainList.DomainList))
	if err != nil {
		log.Printf("[IndustryConnectorTransfer] Error retrieving domain list from DB: %v", err)
	}
	if len(domainList.DomainList) == 0 {
		log.Printf("[IndustryConnectorTransfer] Linked domain list is empty, no records will be sent")
		return nil
	}

	kinesisRecs := []kinesis.Record{}
	for _, record := range s3Event.Records {
		s3 := record.S3
		sourceKey := s3.Object.Key
		objectPrefix := strings.Split(s3.Object.Key, "/")[0]
		objectType := bucketToObjectMap[objectPrefix]
		log.Printf("[IndustryConnectorTransfer] Getting data from  %s", sourceKey)
		// Get the S3 object from the source bucket
		jsonString, err := s3Client.GetTextObj(sourceKey)
		if err != nil {
			log.Printf("[IndustryConnectorTransfer] Error retrieving object from bucket %v: %v", record.S3.Bucket.Name+"/"+sourceKey, err)
			return err
		}

		// Support JSONL files
		jsonData := strings.Split(jsonString, "\n")
		log.Printf("[IndustryConnectorTransfer] File %s has %d lines", objectPrefix, len(jsonData))
		emptyLine := 0
		for _, v := range jsonData {
			if v == "" {
				emptyLine++
				continue // skip empty line at end of file
			}
			kinesisRecs, err = addKinesisRecords(kinesisRecs, objectType, v, domainList, sourceKey)
			if err != nil {
				log.Printf("[IndustryConnectorTransfer] Error creating kinesis records %v", err)
				return err
			}
		}
		log.Printf("[IndustryConnectorTransfer] records=%d, emptyLines=%d", len(kinesisRecs), emptyLine)
	}
	err, ingestionErrs := kinesisClient.PutRecords(kinesisRecs)
	return combineAllErrors(err, ingestionErrs)
}

func combineAllErrors(err error, ingestionErrs []kinesis.IngestionError) error {
	if len(ingestionErrs) > 0 {
		log.Printf("[IndustryConnectorTransfer] Errors adding records to Kinesis: %v", ingestionErrs)
		errs := []error{}
		if err != nil {
			log.Printf("[IndustryConnectorTransfer] Error adding records to Kinesis: %v", err)
			errs = append(errs, err)
		}
		for _, v := range ingestionErrs {
			errs = append(errs, errors.New(v.ErrorMessage))
		}
		return core.ErrorsToError(errs)
	}
	return nil
}

func addKinesisRecords(kinesisRecs []kinesis.Record, objectType string, data string, domainList model.IndustryConnectorDomainList, sourceKey string) ([]kinesis.Record, error) {
	// Parse jsonData
	var parsedData map[string]interface{}
	err := json.Unmarshal([]byte(data), &parsedData)
	if err != nil {
		log.Printf("[IndustryConnectorTransfer] Error parsing object: %v", err)
		return kinesisRecs, err
	}
	// Check for model version
	if parsedData["modelVersion"] == nil {
		log.Printf("[IndustryConnectorTransfer] Object's model version is required but not found")
		return kinesisRecs, errors.New("object's model version is required but not found")
	}

	// Create Kinesis record for each linked domain
	for _, domain := range domainList.DomainList {
		// Create record to send to Kinesis
		bizObj := BusinessObjectRecord{
			Domain:       domain,
			ObjectType:   objectType,
			ModelVersion: parsedData["modelVersion"].(string),
			Data:         parsedData,
		}
		data, err := json.Marshal(bizObj)
		if err != nil {
			log.Printf("[IndustryConnectorTransfer] Error marshalling record: %v", err)
			return kinesisRecs, err
		}
		kinesisRec := kinesis.Record{
			Pk:   sourceKey,
			Data: string(data),
		}
		kinesisRecs = append(kinesisRecs, kinesisRec)
	}
	return kinesisRecs, nil
}

func main() {
	lambda.Start(HandleRequest)
}
