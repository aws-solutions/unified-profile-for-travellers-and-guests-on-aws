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
	core "tah/core/core"
	customerprofiles "tah/core/customerprofiles"
	kinesis "tah/core/kinesis"
	sqs "tah/core/sqs"

	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	awsLamdba "github.com/aws/aws-lambda-go/lambda"
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

var sqsConfig = sqs.InitWithQueueUrl(DLQ_URL, LAMBDA_REGION)

//This is from the connectors solution, probably will want something like this in UCP but not used for now
//var solutionUtils awssolutions.IConfig = awssolutions.Init(METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, METRICS_UUID, ANONYMOUS_USAGE)

type Controller struct {
	Tx core.Transaction
}

type ProcessingError struct {
	Error      error
	Data       string
	AccpRecord string
}

type BusinessObjectRecord struct {
	Domain       string        `json:"domain"`
	ObjectType   string        `json:"objectType"`
	ModelVersion string        `json:"modelVersion"`
	Data         []interface{} `json:"data"`
}

var SQS_MES_ATTR_UCP_ERROR_TYPE = "UcpErrorType"
var SQS_MES_ATTR_BUSINESS_OBJECT_TYPE_NAME = "BusinessObjectTypeName"
var SQS_MES_ATTR_MESSAGE = "Message"

func HandleRequest(ctx context.Context, req events.KinesisEvent) error {
	return HandleRequestWithServices(ctx, req, sqsConfig)
}

func HandleRequestWithServices(ctx context.Context, req events.KinesisEvent, sqsc sqs.Config) error {
	tx := core.NewTransaction("ucp-real-time", "")
	tx.Log("Received %v Kinesis Data Stream Events", len(req.Records))
	for _, rec := range req.Records {
		processKinesisRecord(tx, rec, sqsc)
	}
	return nil
}

func main() {
	awsLamdba.Start(HandleRequest)
}

func processKinesisRecord(tx core.Transaction, rec events.KinesisEventRecord, sqsc sqs.Config) error {
	//Using panic/defer/recover as a try/catch
	defer func() {
		if r := recover(); r != nil {
			tx.Log("Error while processing record. Logging to SQS queue")
			logProcessingError(tx, r.(ProcessingError), sqsc)
		}
	}()
	tx.Log("Processing Kineiss record")
	businessRecord := BusinessObjectRecord{}
	tx.Log("Parsing Kinesis Data")
	err := json.Unmarshal(rec.Kinesis.Data, &businessRecord)
	if err != nil {
		tx.Log("Unexpected format for record inputted %v", err)
		panic(ProcessingError{Error: err, Data: string(rec.Kinesis.Data)})
	}
	//TODO: To remove
	tx.Log("Parsed: %v", businessRecord)

	tx.Log("Extracting ACCP Reccords")
	//TODO: parallelize this
	for _, accpRec := range businessRecord.Data {
		accpRecByte, err2 := json.Marshal(accpRec)
		if err2 != nil {
			tx.Log("Error processing data field, perhaps data not flattened: %v", err2)
			panic(ProcessingError{Error: err, Data: string(rec.Kinesis.Data), AccpRecord: businessRecord.ObjectType})
		}
		customCfg := customerprofiles.InitWithDomain(businessRecord.Domain, LAMBDA_REGION)
		tx.Log("Putting profile object to ACCP in domain %s", businessRecord.Domain)
		err = customCfg.PutProfileObject(string(accpRecByte), businessRecord.ObjectType)
		if err != nil {
			tx.Log("error putting profile object %v", err)
			panic(ProcessingError{Error: err, Data: string(rec.Kinesis.Data), AccpRecord: businessRecord.ObjectType})
		}
	}

	return nil
}

func logProcessingError(tx core.Transaction, errObj ProcessingError, sqsc sqs.Config) {
	accpRec := errObj.AccpRecord
	if accpRec == "" {
		accpRec = "Unknown"
	}
	err := sqsc.SendWithStringAttributes(errObj.Data, map[string]string{
		SQS_MES_ATTR_UCP_ERROR_TYPE:            "ACP real time Ingestion error",
		SQS_MES_ATTR_BUSINESS_OBJECT_TYPE_NAME: accpRec,
		SQS_MES_ATTR_MESSAGE:                   errObj.Error.Error(),
	})
	if err != nil {
		tx.Log("Could not log message to SQS queue %v. Error: %v", sqsc.QueueUrl, err)
	}
}
