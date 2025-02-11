// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"

	model "tah/upt/source/ucp-common/src/model/dynamo-schema"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type CustomerProfileEvent struct {
	SchemaVersion          int                    `json:"schemaVersion"`
	EventId                string                 `json:"eventId"`
	EventTimestamp         string                 `json:"eventTimestamp"`
	EventType              string                 `json:"eventType"`
	DomainName             string                 `json:"domainName"`
	ObjectTypeName         string                 `json:"objectTypeName"`
	AssociatedProfileId    string                 `json:"associatedProfileId"`
	ProfileObjectUniqueKey string                 `json:"profileObjectUniqueKey"`
	Object                 map[string]interface{} `json:"object"`
	IsMessageRealTime      bool                   `json:"isMessageRealTime"`
}

type ServiceRegistry struct {
	Tx       core.Transaction
	DbConfig db.DBConfig
}

var LAMBDA_REGION = os.Getenv("AWS_REGION")
var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")

var DYNAMO_TABLE = os.Getenv("DYNAMO_TABLE")
var DYNAMO_PK = os.Getenv("DYNAMO_PK")
var DYNAMO_SK = os.Getenv("DYNAMO_SK")

var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

var indexDb = db.Init(DYNAMO_TABLE, DYNAMO_PK, DYNAMO_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, req events.KinesisEvent) (map[string]interface{}, error) {
	tx := core.NewTransaction("cp_indexer", "", LOG_LEVEL)
	tx.SetPrefix("cp_indexer")
	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")

	kinesisBatchResponse, err := HandleRequestWithServices(ctx, req, ServiceRegistry{Tx: tx, DbConfig: indexDb})
	return kinesisBatchResponse, err
}

func HandleRequestWithServices(ctx context.Context, req events.KinesisEvent, services ServiceRegistry) (map[string]interface{}, error) {
	batchItemFailures := []map[string]interface{}{}
	for _, request := range req.Records {
		curRecordSequenceNumber := ""
		err := handleRec(ctx, request, services)
		if err != nil {
			curRecordSequenceNumber = request.Kinesis.SequenceNumber
			if curRecordSequenceNumber != "" {
				batchItemFailures = append(batchItemFailures, map[string]interface{}{"itemIdentifier": curRecordSequenceNumber})
			}
			services.Tx.Error("Error: %v", err)
		}
	}
	kinesisBatchResponse := map[string]interface{}{
		"batchItemFailures": batchItemFailures,
	}
	return kinesisBatchResponse, nil
}

func handleRec(ctx context.Context, req events.KinesisEventRecord, services ServiceRegistry) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if castError, ok := r.(error); ok {
				err = castError
			} else {
				panic(r)
			}
		}
	}()

	var cpEvent CustomerProfileEvent
	err = json.Unmarshal(req.Kinesis.Data, &cpEvent)
	if err != nil {
		services.Tx.Error("Error processing event: %v", err)
		return err
	}
	if cpEvent.ObjectTypeName == "_profile" && cpEvent.EventType == "CREATED" {
		profileId, ok := cpEvent.Object["ProfileId"].(string)
		if !ok {
			services.Tx.Error("Error: ProfileId not found in the object")
			return errors.New("error: profileId not found in the object")
		}
		attributeMap, ok := cpEvent.Object["Attributes"].(map[string]interface{})
		if !ok {
			services.Tx.Error("Error: Attributes not found in the object")
			return errors.New("error: attributes not found in the object")
		}
		connectId, ok := attributeMap["external_connect_id"].(string)
		if !ok {
			services.Tx.Error("Error: external_connect_id not found in attributes")
			return errors.New("error: external_connect_id not found in attributes")
		}
		cpIdMap := model.CpIdMap{
			DomainName: cpEvent.DomainName,
			ConnectId:  "lcs_" + connectId,
			CpId:       profileId,
		}
		_, err := services.DbConfig.Save(cpIdMap)
		if err != nil {
			services.Tx.Error("Error saving lcs to cp map: %v", err)
			return err
		}
		cpIdMap = model.CpIdMap{
			DomainName: cpEvent.DomainName,
			ConnectId:  "cp_" + profileId,
			CpId:       connectId,
		}
		_, err = services.DbConfig.Save(cpIdMap)
		if err != nil {
			services.Tx.Error("Error saving cp to lcs: %v", err)
			return err
		}
	}
	return err
}
