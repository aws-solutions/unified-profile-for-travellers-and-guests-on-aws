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
	core "tah/core/core"
	customerprofiles "tah/core/customerprofiles"
	kinesis "tah/core/kinesis"

	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	awsLamdba "github.com/aws/aws-lambda-go/lambda"
)

type businessObjectRecord struct {
	ObjectType   string                 `json:"objectType"`
	ModelVersion string                 `json:"modelVersion"`
	Data         map[string]interface{} `json:"data"`
}

type Event struct {
	TxID           string
	Connector      string
	Usecase        string
	IngestionError kinesis.IngestionError
	OriginalEvent  string
}

// Environment Variable passed from infrastructure
var LAMBDA_REGION = os.Getenv("TAH_REGION")

//var ANONYMOUS_USAGE = os.Getenv("ANONYMOUS_USAGE")
//var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
//var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
//var METRICS_UUID = os.Getenv("METRICS_UUID")

var INPUT_STREAM = os.Getenv("INPUT_STREAM")
var DLQ = os.Getenv("dlqname")

// initlialize resource handlers
var kinesisCfg = kinesis.Init(INPUT_STREAM, LAMBDA_REGION)

var DOMAIN_NAME string = "test-domain"
var profile_id string = "id-123"

//This is from the connectors solution, probably will want something like this in UCP but not used for now
//var solutionUtils awssolutions.IConfig = awssolutions.Init(METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, METRICS_UUID, ANONYMOUS_USAGE)

type Controller struct {
	Tx core.Transaction
}

func HandleRequest(ctx context.Context, req events.KinesisEvent) error {
	return HandleRequestWithServices(ctx, req, &kinesisCfg)
}

func HandleRequestWithServices(ctx context.Context, req events.KinesisEvent, kinesisCfg kinesis.IConfig) error {
	customCfg := customerprofiles.InitWithDomain(DOMAIN_NAME, LAMBDA_REGION)
	tx := core.NewTransaction("ucp-real-time", "")
	tx.Log("Received %v Kinesis Data Stream Events", len(req.Records))
	outputStream, err := kinesisCfg.Describe()
	if err != nil {
		tx.Log("Error retreiving output stream config %v. Exiting", err)
		return nil
	}
	tx.Log("Input stream configuration: %+v", outputStream)

	events := []Event{}
	for _, rec := range req.Records {
		event, err := parseEvent(tx, rec)
		if err != nil {
			tx.Log("Error parsing event")
		}
		events = append(events, event)

	}
	processEvents(tx, events, customCfg)
	//We do not want to reprocess events on failure
	return nil
}

func main() {
	awsLamdba.Start(HandleRequest)
}

func parseEvent(tx core.Transaction, kinesisEvt events.KinesisEventRecord) (Event, error) {
	tx.Log("Parsing kinesis event record")
	var pnrNotif businessObjectRecord
	err := json.Unmarshal(kinesisEvt.Kinesis.Data, &pnrNotif)
	evt := Event{
		OriginalEvent: string(kinesisEvt.Kinesis.Data),
	}
	if err != nil {
		return evt, err
	}

	return evt, errors.New("could not match use case")
}

func processEvents(tx core.Transaction, events []Event, customCfg customerprofiles.CustomerProfileConfig) error {
	tx.Log("Processing events")
	var pnrNotif businessObjectRecord
	for _, event := range events {
		byteArray, _ := json.Marshal(event.OriginalEvent)
		err1 := json.Unmarshal(byteArray, &pnrNotif)
		if err1 != nil {
			return err1
		}
		pnrNotif.Data["profile_id"] = profile_id
		serialProfile, err2 := json.Marshal(pnrNotif.Data)
		if err2 != nil {
			return err2
		}

		err := customCfg.PutProfileObject(string(serialProfile), pnrNotif.ObjectType)
		if err != nil {
			return err
		}
	}
	return nil
}

func sendMetrics(events []Event) {
	//for connector, evts := range eventsByConnector {
	// solutionUtils.SendMetrics(map[string]interface{}{
	// 	"use_case":    "real_time_ingestion",
	// 	"connector":   connector,
	// 	"num_records": len(evts),
	// })
	//}
}
