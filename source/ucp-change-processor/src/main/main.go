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
	"strings"
	"sync"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/aurora"
	core "tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	eventbridge "tah/upt/source/tah-core/eventbridge"
	firehose "tah/upt/source/tah-core/firehose"
	"tah/upt/source/tah-core/kinesis"
	s3 "tah/upt/source/tah-core/s3"
	secrets "tah/upt/source/tah-core/secret"
	"time"

	sqs "tah/upt/source/tah-core/sqs"

	"context"
	"os"

	adapter "tah/upt/source/ucp-common/src/adapter"
	common "tah/upt/source/ucp-common/src/constant/admin"
	changeEvent "tah/upt/source/ucp-common/src/model/change-event"
	model "tah/upt/source/ucp-common/src/model/traveler"
	travSvc "tah/upt/source/ucp-common/src/services/traveller"

	"github.com/aws/aws-lambda-go/events"
	awsLambda "github.com/aws/aws-lambda-go/lambda"
)

// Environment Variable passed from infrastructure
var LAMBDA_REGION = os.Getenv("LAMBDA_REGION")
var INPUT_STREAM = os.Getenv("INPUT_STREAM")
var EVENTBRIDGE_ENABLED = os.Getenv("EVENTBRIDGE_ENABLED")
var TRAVELLER_EVENT_BUS = os.Getenv("TRAVELLER_EVENT_BUS")
var FIREHOSE_STREAM_NAME = os.Getenv("FIREHOSE_STREAM_NAME")

var S3_OUTPUT = os.Getenv("S3_OUTPUT")

var DLQ_URL = os.Getenv("DEAD_LETTER_QUEUE_URL")
var SECRET_ARN = os.Getenv("SECRET_ARN")

var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
var METRICS_NAMESPACE = "upt"

// Low cost storage
var AURORA_PROXY_READONLY_ENDPOINT = os.Getenv("AURORA_PROXY_READONLY_ENDPOINT")
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

var sqsConfig = sqs.InitWithQueueUrl(DLQ_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var eventbridgeClient = eventbridge.InitWithBus(LAMBDA_REGION, TRAVELLER_EVENT_BUS, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var s3Client = s3.Init(S3_OUTPUT, "", LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var firehoseClient = firehose.Init(FIREHOSE_STREAM_NAME, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

// Init low-cost clients
var smClient = secrets.InitWithRegion(AURORA_DB_SECRET_ARN, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var user, pwd = getUserCredentials(smClient)
var auroraClient = aurora.InitPostgresql(AURORA_PROXY_READONLY_ENDPOINT, AURORA_DB_NAME, user, pwd, 50, LOG_LEVEL)

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

var OBJECT_TYPE_NAME_PROFILE = "_profile"

var EVENT_TYPE_CREATED = "CREATED"
var EVENT_TYPE_DELETED = "DELETED"
var EVENT_TYPE_UPDATED = "UPDATED"
var EVENT_TYPE_MERGED = "MERGED"
var EVENT_TYPE_UNMERGED = "UNMERGED"

var UNEXPECTED_FORMAT_ERROR = "Unexpected format for record input"

var EVENT_TYPE_MAP = map[string]string{
	EVENT_TYPE_CREATED: adapter.OPERATION_CREATE,
	EVENT_TYPE_DELETED: adapter.OPERATION_DELETE,
	EVENT_TYPE_UPDATED: adapter.OPERATION_UPDATE,
}

/**
{
 "SchemaVersion": 0,
 "EventId": "b88d5339-69a3-4ef0-ac0e-208be1823a55", # record id
 "EventTimestamp": "2023-08-24T02:43:20.000Z", # timestamp of when profile standard fields are merged
 "EventType": "MERGED",
 "DomainName": "IrvinDomain",
 "ObjectTypeName": "_profile",
 "AssociatedProfileId": null,
 "ProfileObjectUniqueKey": null,
 "Object": {
   "FirstName": "FirstName",
   "LastName": "LastName",
   "AdditionalInformation": "Main",
   "ProfileId": "111"
   ...
 },
 "MergeId": "045d69ae-42c2-11ee-be56-0242ac120002",
 "DuplicateProfiles": [
   {
     "FirstName": "FirstName",
     "LastName": "LastName",
     "AdditionalInformation": "Duplicate1",
     "ProfileId": "222"
     ...
   },
   {
     "FirstName": "FirstName",
     "LastName": "LastName",
     "AdditionalInformation": "Duplicate2",
     "ProfileId": "333"
     ...
   },
   ...
 ],
 "IsMessageRealTime": true
}
**/

type ProcessingError struct {
	Error        error
	Data         string
	AccpRecord   string
	AWSRequestID string
	MessageID    string
}

type TravelerEvent struct {
	TravelerID string      `json:"connectId"`
	Events     []AccpEvent `json:"events"`
	Domain     string      `json:"domain"`
	Time       time.Time   `json:"time"`
}

type AccpEvent struct {
	EventType         string          `json:"eventType"`
	ObjectType        string          `json:"objectType"`
	PartialTraveler   model.Traveller `json:"data"`
	MergeEventDetails MergeEvent      `json:"mergeEventDetails"`
}

type Runner struct {
	Tx       core.Transaction
	State    State
	Sqs      sqs.Config
	Accp     customerprofiles.ICustomerProfileLowCostConfig
	EvBridge eventbridge.IConfig
	S3       s3.S3Config
	Firehose firehose.Config
	Env      map[string]string
}

type State struct {
	RawData         string
	PartialTraveler []model.Traveller
	MergeEvent      MergeEvent
	Domain          string
	ObjectType      string
	EventType       string
}

type MergeEvent struct {
	SourceConnectID  string   `json:"sourceConnectId"`
	TargetConnectIDs []string `json:"targetConnectIds"`
	SourceProfileID  string   `json:"sourceTravelerId"`
	TargetProfileIDs []string `json:"targetTravelerIds"`
}

type DomainEvents struct {
	Events []AccpEvent
	Domain string
}

func HandleRequest(ctx context.Context, req events.KinesisEvent) error {
	r := buildRunnerFromGlobals()
	return r.HandleRequestWithServices(ctx, req)
}

func buildRunnerFromGlobals() Runner {
	accpClient := initStorage()
	return Runner{
		Sqs:      sqsConfig,
		Accp:     accpClient,
		EvBridge: &eventbridgeClient,
		S3:       s3Client,
		Firehose: *firehoseClient,
		Env: map[string]string{
			"EVENTBRIDGE_ENABLED":  EVENTBRIDGE_ENABLED,
			"FIREHOSE_STREAM_NAME": FIREHOSE_STREAM_NAME,
		},
	}
}

func (r *Runner) HandleRequestWithServices(ctx context.Context, req events.KinesisEvent) error {
	r.Tx = core.NewTransaction("ucp-change-processor", "", LOG_LEVEL)
	r.Tx.Info("Received %v Kinesis Data Stream Events", len(req.Records))
	r.S3.SetTx(r.Tx)
	r.Accp.SetTx(r.Tx)
	eventsByTravellersId := map[string]DomainEvents{}
	for _, rec := range req.Records {
		r.Tx.Debug("Parsing change record")
		err := r.parseKinesisRecord(rec)
		if err != nil {
			//we don't return the error to avoid auto retry for now give that the error is logged
			r.Tx.Error("Error parsing Kinesis record")
			return nil
		}
		eventsByTravellersId = r.indexEventsByTravelerID(eventsByTravellersId)
	}

	if r.Env["EVENTBRIDGE_ENABLED"] == "true" {
		r.Tx.Debug("EventBridge enabled")
		r.sendToDownstreamTargets(eventsByTravellersId)
	} else {
		r.Tx.Debug("EventBridge is disabled")
	}
	//Cleanup S3 on merge/delete requests
	r.cleanupS3(eventsByTravellersId)
	return nil
}

func main() {
	awsLambda.Start(HandleRequest)
}

func (r *Runner) cleanupS3(eventsByTravellersId map[string]DomainEvents) {
	r.Tx.Debug("Marking profiles that have been merged or deleted")
	travellers := []model.Traveller{}
	for sourcePid, domainEvents := range eventsByTravellersId {
		for _, accpEvent := range domainEvents.Events {
			if accpEvent.EventType == EVENT_TYPE_MERGED {
				r.Tx.Debug("Cleaning up profiles: %v", accpEvent.MergeEventDetails.TargetConnectIDs)
				for _, idToDelete := range accpEvent.MergeEventDetails.TargetConnectIDs {
					travellers = append(travellers, model.Traveller{
						Domain:      domainEvents.Domain,
						MergedIn:    sourcePid,
						ConnectID:   idToDelete,
						LastUpdated: time.Now(),
					})
				}
			}
			if accpEvent.EventType == EVENT_TYPE_DELETED {
				r.Tx.Debug("Cleaning up profile: %s", accpEvent.PartialTraveler.ConnectID)
				travellers = append(travellers, model.Traveller{
					Domain:      domainEvents.Domain,
					ConnectID:   accpEvent.PartialTraveler.ConnectID,
					LastUpdated: time.Now(),
				})
			}
		}
	}
	if len(travellers) > 0 {
		r.MarkAsMerged(travellers)
	} else {
		r.Tx.Debug("No profiles to mark as merged")
	}
}

func (r *Runner) sendToDownstreamTargets(eventsByTravellersId map[string]DomainEvents) {
	r.Tx.Info("Sending data to eventbridge and S3")
	wg := sync.WaitGroup{}
	nSentEb := 0
	mu := sync.Mutex{}
	firehoseRecords := []firehose.Record{}

	for connectId, domEvents := range eventsByTravellersId {
		wg.Add(1)
		go func(cid string, domainEvents DomainEvents) {
			defer wg.Done()

			// Send event to EventBridge
			tevt, err := r.sendToEventBridge(cid, domainEvents)
			r.LogErrorIfNotNil(err, fmt.Sprintf("[cid:%s] Error sending traveler event to EventBridge: %v", cid, err), tevt)
			if err == nil {
				nSentEb++
			}

			// Send event to S3
			for _, event := range domainEvents.Events {
				if event.EventType == EVENT_TYPE_DELETED {
					return // no need to send an updated profile to S3, profile is deleted and will be removed during `cleanupS3()`
				}
			}
			tJson, err := r.buildTravelerForS3(cid, domainEvents)
			r.LogErrorIfNotNil(err, fmt.Sprintf("[cid:%s] Error sending traveler event to S3: %v", cid, err), "")
			if err == nil {
				mu.Lock()
				firehoseRecords = append(firehoseRecords, firehose.Record{Data: string(tJson)})
				mu.Unlock()
			}
		}(connectId, domEvents)
		// we don't log the error here as the profile might have been deleted automatically by the auto-merge
	}
	r.Tx.Debug("Waiting for %v request to complete", len(eventsByTravellersId))
	wg.Wait()
	r.Tx.Info("Sending %v records to Firehose", len(firehoseRecords))
	ingErrs, err := r.Firehose.PutRecords(r.Env["FIREHOSE_STREAM_NAME"], firehoseRecords)
	if err != nil {
		r.Tx.Error("Error sending batch of travellers to S3 via firehose %v : %v", err)
		r.logProcessingError(ProcessingError{Error: err, Data: ""})
	}
	for _, ingErr := range ingErrs {
		r.Tx.Error("[cid:%s] Error sending batch of travellers to S3 via firehose %v : %v", ingErr)
		errJson, _ := json.Marshal(ingErr)
		r.logProcessingError(ProcessingError{Error: errors.New("error ingesting record to Firehose"), Data: string(errJson)})
	}
	r.Tx.Info("Successfully sent %v events to EventBridge and %v profiles to S3", nSentEb, len(firehoseRecords)-len(ingErrs))

}

func (r *Runner) retrieveTravellerWithRetries(connectId string) (model.Traveller, error) {
	r.Tx.Info("Retrieving traveler profile with id %v", connectId)
	traveller, err := travSvc.RetreiveTraveller(r.Tx, r.Accp, connectId, []customerprofiles.PaginationOptions{})
	if err == nil {
		return traveller, nil
	}

	r.Tx.Info("Error retrieving traveller with id %v:  %v. retrying once after 5 sec", connectId, err)
	time.Sleep(time.Second * 5)
	traveller, err = travSvc.RetreiveTraveller(r.Tx, r.Accp, connectId, []customerprofiles.PaginationOptions{})
	if err == nil {
		return traveller, nil
	}

	r.Tx.Info("Error retrieving traveller with id %v:  %v. retrying twice after 10 sec", connectId, err)
	time.Sleep(time.Second * 10)
	traveller, err = travSvc.RetreiveTraveller(r.Tx, r.Accp, connectId, []customerprofiles.PaginationOptions{})
	if err == nil {
		return traveller, nil
	}

	r.Tx.Info("Error retrieving traveller with id %v:  %v. giving up. traveller may have been deleted", connectId, err)
	r.logProcessingError(
		ProcessingError{
			Error: errors.New("Error retrieving profile with connect ID " + connectId + " after 2 retries."),
			Data:  r.State.RawData,
		},
	)
	return model.Traveller{}, err
}

// This function marks the object as delete in S3 by moving it to a different folder
func (r *Runner) MarkAsMerged(travellers []model.Traveller) error {
	r.Tx.Info("Marking %v profiles as merged", len(travellers))
	recs := []firehose.Record{}
	for _, traveller := range travellers {
		tJson, err := json.Marshal(traveller)
		if err != nil {
			r.Tx.Error("Could not marshal merged traveler %v : %v", traveller.ConnectID, err)
			r.logProcessingError(ProcessingError{Error: err, Data: string(tJson)})
		}
		recs = append(recs, firehose.Record{Data: string(tJson)})
	}
	r.Tx.Debug("Sending %v merged records to Firehose", len(recs))
	ingErrs, err := r.Firehose.PutRecords(r.Env["FIREHOSE_STREAM_NAME"], recs)
	if err != nil {
		r.Tx.Error("Error sending batch of merged travellers to S3 via firehose %v : %v", err)
		r.logProcessingError(ProcessingError{Error: err, Data: ""})
	}
	for _, ingErr := range ingErrs {
		r.Tx.Error("Error sending batch of travellers to S3 via firehose %+v", ingErr)
		errJson, _ := json.Marshal(ingErr)
		r.logProcessingError(ProcessingError{Error: errors.New("error ingesting record to Firehose"), Data: string(errJson)})
	}
	return nil
}

func (r *Runner) indexEventsByTravelerID(eventsByTravellersId map[string]DomainEvents) map[string]DomainEvents {
	state := r.State
	// unmerge operation involves recreating the "unmerged" profile and state.PartialTraveler can comprise of two profiles
	// Iterating through each traveller to index each by its ID
	for _, traveller := range state.PartialTraveler {
		r.Tx.Debug("Indexing event %v by traveller ID %v ", state.EventType, traveller.ConnectID)
		if _, ok := eventsByTravellersId[traveller.ConnectID]; !ok {
			eventsByTravellersId[traveller.ConnectID] = DomainEvents{Events: []AccpEvent{}, Domain: state.Domain}
		}
		//Grouping events by traveller ID to support traveller bound event generation
		evts := eventsByTravellersId[traveller.ConnectID]
		evts.Events = append(evts.Events, AccpEvent{
			EventType:         state.EventType,
			ObjectType:        state.ObjectType,
			MergeEventDetails: state.MergeEvent,
			PartialTraveler:   traveller,
		})
		eventsByTravellersId[traveller.ConnectID] = evts
	}
	return eventsByTravellersId
}

func (r *Runner) parseKinesisRecord(rec events.KinesisEventRecord) error {
	//Using panic/defer/recover as a try/catch
	defer func() {
		if recovery := recover(); recovery != nil {
			r.Tx.Error("Error while processing record. Logging to SQS queue")
			r.logProcessingError(recovery.(ProcessingError))
		}
	}()
	r.Tx.Debug("Processing Kinesis record with partition key %v", rec.Kinesis.PartitionKey)
	profile := profilemodel.Profile{}
	unmergedProfile := profilemodel.Profile{}
	changeRec := changeEvent.ProfileChangeRecord{}
	mergeTargetProfileIds := []string{}
	mergeTargetConnectIds := []string{}
	//Unprettify json just in case
	rec.Kinesis.Data = []byte(strings.Replace(string(rec.Kinesis.Data), "\n", "", -1))
	rec.Kinesis.Data = []byte(strings.Replace(string(rec.Kinesis.Data), "\t", "", -1))
	err := json.Unmarshal(rec.Kinesis.Data, &changeRec)
	if err != nil {
		r.Tx.Error(UNEXPECTED_FORMAT_ERROR+" %v", err)
		r.logProcessingError(ProcessingError{Error: err, Data: string(rec.Kinesis.Data)})
		return err
	}
	domain := changeRec.DomainName
	r.Tx.Debug("Processing ACCP Record of type: %s for connectId %s", changeRec.ObjectTypeName, changeRec.AssociatedProfileId)
	if changeRec.ObjectTypeName == OBJECT_TYPE_NAME_PROFILE {
		r.Tx.Debug("Parsing Profile")
		changeRecProfile := changeEvent.ProfileChangeRecordWithProfile{}
		err = json.Unmarshal(rec.Kinesis.Data, &changeRecProfile)
		if err != nil {
			r.Tx.Error(UNEXPECTED_FORMAT_ERROR+" %v", err)
			r.logProcessingError(ProcessingError{Error: err, Data: string(rec.Kinesis.Data)})
			return err
		}
		profile = changeRecProfile.Object
		if changeRecProfile.EventType == EVENT_TYPE_MERGED {
			for _, target := range changeRecProfile.DuplicateProfiles {
				r.Tx.Debug("Merged target profile has connect ID: %s and profile id %v", target.ProfileId, target.Attributes["profile_id"])
				mergeTargetProfileIds = append(mergeTargetProfileIds, target.Attributes["profile_id"])
				mergeTargetConnectIds = append(mergeTargetConnectIds, target.ProfileId)
			}
		} else if changeRecProfile.EventType == EVENT_TYPE_UNMERGED {
			unmergedProfile = changeRecProfile.DuplicateProfiles[0]
		}

	} else {
		r.Tx.Debug("Parsing %s", changeRec.ObjectTypeName)
		changeRecProfileObject := changeEvent.ProfileChangeRecordWithProfileObject{}
		err = json.Unmarshal(rec.Kinesis.Data, &changeRecProfileObject)
		if err != nil {
			r.Tx.Error(UNEXPECTED_FORMAT_ERROR+" %v", err)
			r.logProcessingError(ProcessingError{Error: err, Data: string(rec.Kinesis.Data)})
			return err
		}
		profile.ProfileId = changeRecProfileObject.AssociatedProfileId
		objectData := core.MapStringInterfaceToMapStringString(changeRecProfileObject.Object)
		profile.ProfileObjects = append(profile.ProfileObjects, profilemodel.ProfileObject{Type: changeRec.ObjectTypeName, Attributes: objectData})
	}

	partialTravellers := []model.Traveller{adapter.ProfileToTraveller(r.Tx, profile)}
	if unmergedProfile.ProfileId != "" {
		// if event type is unmerge, append the unmerged profile
		partialTravellers = append(partialTravellers, adapter.ProfileToTraveller(r.Tx, unmergedProfile))
	}

	r.State = State{ObjectType: changeRec.ObjectTypeName,
		RawData:   string(rec.Kinesis.Data),
		EventType: changeRec.EventType,
		Domain:    domain,
		MergeEvent: MergeEvent{
			SourceConnectID:  profile.ProfileId,
			TargetConnectIDs: mergeTargetConnectIds,
			SourceProfileID:  profile.Attributes["profile_id"],
			TargetProfileIDs: mergeTargetProfileIds,
		},
		PartialTraveler: partialTravellers,
	}
	return nil

}

func (r *Runner) LogErrorIfNotNil(err error, logLine string, sourceData interface{}) {
	if err != nil {
		r.Tx.Error(logLine)
		data, err2 := json.Marshal(sourceData)
		str := string(data)
		if err2 != nil {
			str = "Could not marshal data. error" + err2.Error()
		}
		r.logProcessingError(ProcessingError{Error: err, Data: str})
	}
}

func (r *Runner) logProcessingError(errObj ProcessingError) {
	accpRec := errObj.AccpRecord
	if accpRec == "" {
		accpRec = "Unknown"
	}
	err := r.Sqs.SendWithStringAttributes(errObj.Data, map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE:            "ACCP Change Processing Error",
		common.SQS_MES_ATTR_BUSINESS_OBJECT_TYPE_NAME: accpRec,
		common.SQS_MES_ATTR_MESSAGE:                   errObj.Error.Error(),
	})
	if err != nil {
		r.Tx.Error("Could not log message to SQS queue %v. Error: %v", r.Sqs.QueueUrl, err)
	}
}

func (r *Runner) sendToEventBridge(id string, events DomainEvents) (TravelerEvent, error) {
	r.Tx.Info("Sending %d events to EventBridge for Profile ID %s", len(events.Events), id)
	tevt := TravelerEvent{
		TravelerID: id,
		Events:     events.Events,
		Domain:     events.Domain,
		Time:       time.Now(),
	}
	err := r.EvBridge.SendMessage("ucp-"+events.Domain, "traveller change event", tevt)
	return tevt, err
}

func (r *Runner) buildTravelerForS3(id string, events DomainEvents) ([]byte, error) {
	r.Tx.Info("Sending %d events to S3 for Profile ID %s", len(events.Events), id)
	r.Accp.SetDomain(events.Domain)
	traveler, err := r.retrieveTravellerWithRetries(id)
	if err != nil {
		return nil, err
	}

	traveler.LastUpdated = time.Now()
	tJson, err := json.Marshal(traveler)
	if err != nil {
		r.logProcessingError(ProcessingError{Error: err, Data: string(tJson)})
		return nil, err
	}

	return tJson, nil
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
