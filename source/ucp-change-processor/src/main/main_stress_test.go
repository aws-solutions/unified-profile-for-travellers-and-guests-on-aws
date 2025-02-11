// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	customerprofiles "tah/upt/source/storage"
	core "tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	eventbridge "tah/upt/source/tah-core/eventbridge"
	firehose "tah/upt/source/tah-core/firehose"
	iam "tah/upt/source/tah-core/iam"
	s3 "tah/upt/source/tah-core/s3"
	sqs "tah/upt/source/tah-core/sqs"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

//////////////////////////////////////////////////////////////////////////////////////////////////
// These test are not run in the CICD pipline. They are meant to be run manually to test
// highly parralel S3 ingestion using firehose. usage: sh ./test_scalability.sh
///////////////////////////////////////////////////////////////////////////////////////////

func TestScalability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress testing in short mode")
	}
	domain := "change_proc_stress_tests_domain_" + strings.ToLower(core.GenerateUniqueId())

	svc := sqs.Init(UCP_REGION, "", "")
	ebQueue := sqs.Init(UCP_REGION, "", "")
	qName := "Test-Queue-2" + core.GenerateUniqueId()
	_, err := svc.Create(qName)
	if err != nil {
		t.Errorf("[testMain] Could not create test queue")
	}
	s3c, err := s3.InitWithRandBucket("test-change-proc-main", "", UCP_REGION, "", "")
	if err != nil {
		t.Errorf("[testMain] Could not create test bucket %v", err)
	}

	name := s3c.Bucket
	bucketArn := "arn:aws:s3:::" + name
	resources := []string{bucketArn, bucketArn + "/*"}
	actions := []string{"s3:PutObject"}
	principal := map[string][]string{"Service": {"firehose.amazonaws.com"}}
	err = s3c.AddPolicy(name, resources, actions, principal)
	if err != nil {
		t.Errorf("[TestCreateDeleteBucket] error adding bucket policy %+v", err)
	}

	iamCfg := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	roleName := "firehose-unit-test-role"
	roleArn, _, err := iamCfg.CreateRoleWithActionsResource(roleName, "firehose.amazonaws.com", []string{"s3:PutObject",
		"s3:PutObjectAcl",
		"s3:ListBucket"}, resources)
	if err != nil {
		t.Errorf("[TestIam] error %v", err)
	}
	log.Printf("Waiting 10 seconds role to propagate")
	time.Sleep(10 * time.Second)
	streamName := "test_stream_change_proc"
	log.Printf("0-Creating stream %v", streamName)
	firehoseCfg, err := firehose.InitAndCreateWithPartitioning(
		streamName,
		bucketArn,
		roleArn,
		[]firehose.PartitionKeyVal{{Key: "domainname", Val: ".domain"}},
		UCP_REGION,
		"solutionId",
		"solutionVersion",
	)
	if err != nil {
		t.Errorf("[TestMain] error creating firehose stream %v", err)
	}
	firehoseCfg.WaitForStreamCreation(streamName, 120)

	evBridge, err := eventbridge.InitWithRandomBus(UCP_REGION, "change_proc_stress_tests", "solutionId", "solVersion")
	if err != nil {
		t.Errorf("[testMain] Could not create test bus %v", err)
	}
	// Create SQS queue to validate EventBridge events
	ebQueueName := "eb_queue_" + domain
	ebQueueUrl, err := ebQueue.Create(ebQueueName)
	if err != nil {
		t.Errorf("[TestRealTime] Error creating SQS queue: %s", err)
	}
	eventPattern := fmt.Sprintf("{\"source\": [\"ucp-%s\"]}", domain) // rule fires on events for this domain only
	err = evBridge.CreateSQSTarget(domain, evBridge.EventBus, ebQueueUrl, eventPattern)
	if err != nil {
		t.Errorf("[TestRealTime] Error creating SQS target: %s", err)
	}
	t.Cleanup(func() {
		err = evBridge.DeleteRuleByName(domain)
		if err != nil {
			t.Errorf("[%s] error deleting rule: %v", t.Name(), err)
		}
	})

	nBatch := 10
	batchSize := 100
	profileIds := []string{}
	createEvents := []events.KinesisEvent{}
	mergeEvents := []events.KinesisEvent{}
	log.Printf("Creating Kineisis requetsts for %+v batches of %+v events in paralel", nBatch, batchSize)
	for i := 0; i < nBatch; i++ {
		log.Printf("Batch %v", i)
		createBatch := []events.KinesisEventRecord{}
		mergeBatch := []events.KinesisEventRecord{}
		log.Printf("[batch-%v] Adding %v events", i, batchSize)
		for j := 0; j < batchSize; j++ {
			profileID := core.UUID()
			profileIds = append(profileIds, profileID)
			createBatch = append(
				createBatch,
				events.KinesisEventRecord{
					Kinesis: events.KinesisRecord{Data: []byte(buildEvent(domain, "CREATED", "_profile", profileID, ""))},
				},
			)
			if i > nBatch/2-1 {
				mergeBatch = append(
					mergeBatch,
					events.KinesisEventRecord{
						Kinesis: events.KinesisRecord{
							Data: []byte(buildEvent(domain, "MERGED", "_profile", profileIds[(i-nBatch/2)*batchSize+j], profileID)),
						},
					},
				)
			}
		}
		log.Printf("Adding %v item in create batch", len(createBatch))
		createEvents = append(createEvents, events.KinesisEvent{Records: createBatch})
		log.Printf("Adding %v items in merge batch", len(mergeBatch))
		mergeEvents = append(mergeEvents, events.KinesisEvent{Records: mergeBatch})
	}
	evts := []events.KinesisEvent{}
	evts = append(evts, createEvents...)
	evts = append(evts, mergeEvents...)
	profilesById := map[string]profilemodel.Profile{}
	for _, id := range profileIds {
		profilesById[id] = profilemodel.Profile{
			Domain:    domain,
			ProfileId: id,
		}
	}
	r := Runner{
		Sqs: svc,
		Accp: &customerprofiles.MockCustomerProfileLowCostConfig{
			ProfilesByID: profilesById,
		},
		EvBridge: &evBridge,
		S3:       s3c,
		Firehose: *firehoseCfg,
		Env: map[string]string{
			"EVENTBRIDGE_ENABLED":  "true",
			"FIREHOSE_STREAM_NAME": streamName,
		},
	}

	log.Printf("Sending %v CREATE events", len(evts))
	wg := sync.WaitGroup{}
	for i, evt := range evts {
		wg.Add(1)
		log.Printf("Sending event %v with %v records", i, len(evt.Records))
		go func(event events.KinesisEvent) {
			r.HandleRequestWithServices(context.Background(), event)
			wg.Done()
		}(evt)
	}
	log.Printf("Waiting for all requests to be sent")
	wg.Wait()

	err = firehose.WaitForData(s3c, 300)
	if err != nil {
		t.Errorf("[TestKinesis] error waiting for data: %v", err)
	}
	res, err := s3c.Search("", 10000)
	log.Printf("S3 bucket constains %v items", len(res))
	if err != nil {
		t.Errorf("Could not list results in S3")
	}
	nItems := 0
	for _, file := range res {
		log.Printf("File: %v", file)
		content, err := s3c.GetTextObj(file)
		if err != nil {
			t.Errorf("Could not get text object %s from S3: %v", file, err)
		}
		log.Printf("File %v contains %v lines", file, len(strings.Split(content, "\n")))
		nItems += len(strings.Split(content, "\n"))
		log.Printf("Total items in S3 up to now: %v", nItems)
	}

	//we expect the S3 bucket contain 1 object per profile created
	//+ 1 object per source profile for each merge event
	//+ 1 object per target profile for each merge event
	nCreate := nBatch * batchSize
	nProfileSource := nBatch * batchSize / 2
	nProfiletarget := nBatch * batchSize / 2
	nMerge := nProfileSource + nProfiletarget
	expectedObjects := nCreate + nMerge
	if nItems != expectedObjects {
		t.Errorf("Bucket Should have %v items, not %v", expectedObjects, nItems)
	}

	res, err = s3c.Search("error", nBatch*batchSize)
	log.Printf("S3 bucket constains %v items", len(res))
	if err != nil {
		t.Errorf("Could not list results in S3: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("Bucket Should have %v items in the error folder, not %v: %v", 0, len(res), res)
	}

	i := 0
	queuAttr := map[string]string{}
	expectedEvents := nBatch*batchSize + nBatch*batchSize/2
	for i < 5 && queuAttr["ApproximateNumberOfMessages"] != strconv.Itoa(expectedEvents) {
		queuAttr, err = ebQueue.GetAttributes()
		log.Printf("#Items in Queue: %+v", queuAttr["ApproximateNumberOfMessages"])
		if err != nil {
			t.Errorf("Could not  get EB queue attributes")
		}
		log.Printf("Waiting for ApproximateNumberOfMessages stabilization")
		time.Sleep(5 * time.Second)
		i++
	}

	log.Printf("Queueu attr: %+v", queuAttr)
	if queuAttr["ApproximateNumberOfMessages"] != strconv.Itoa(expectedEvents) {
		messages, _ := ebQueue.Get(sqs.GetMessageOptions{})
		t.Errorf("Eb queue Should have %v items not %v", strconv.Itoa(expectedEvents), queuAttr["ApproximateNumberOfMessages"])
		log.Printf("Queue content: %+v", messages)
	}

	err = svc.DeleteByName(qName)
	if err != nil {
		t.Errorf("[TestMain]error deleting queue by name: %v", err)
	}
	err = svc.DeleteByName(ebQueueName)
	if err != nil {
		t.Errorf("[TestMain]error deleting queue by name: %v", err)
	}
	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestMain]error deleting s3 by name: %v", err)
	}
	log.Printf("Deleting rules by name %v", domain)
	err = evBridge.DeleteRuleByName(domain)
	if err != nil {
		t.Errorf("[TestMain] error deleting event bridge rule by name: %v", err)
	}
	log.Printf("Deleting Bus %v", evBridge.EventBus)
	err = evBridge.DeleteBus(evBridge.EventBus)
	if err != nil {
		log.Printf("ERROR Deleting Bus %v. (keeping error silent)", evBridge.EventBus)
		//t.Errorf("[TestMain]error deleting event bus: %v", err)
	}
	log.Printf("4-Deleting stream %v", streamName)
	err = firehoseCfg.Delete(streamName)
	if err != nil {
		t.Errorf("[TestKinesis] error deleting stream: %v", err)
	}
	err = iamCfg.DeleteRole(roleName)
	if err != nil {
		t.Errorf("[TestIam] error deleting role: %v", err)
	}
}

func buildEvent(domain, evType, objectType, connectID, dupeID string) string {
	return fmt.Sprintf(`
{
	"SchemaVersion": 0,
	"EventId": "b88d5339-69a3-4ef0-ac0e-208be1823a55",
	"EventTimestamp": "2023-08-24T02:43:20.000Z",
	"EventType": "%v",
	"DomainName": "%v",
	"ObjectTypeName": "%v",
	"Object": {
	  "FirstName": "FirstName",
	  "LastName": "LastName",
	  "AdditionalInformation": "Main",
	  "ProfileId": "%v"
	},
	"MergeId": "045d69ae-42c2-11ee-be56-0242ac120002",
	"DuplicateProfiles": [
	  {
		"FirstName": "FirstName",
		"LastName": "LastName",
		"AdditionalInformation": "Duplicate1",
		"ProfileId": "%v"
	  }
	],
	"IsMessageRealTime": true
   }`, evType, domain, objectType, connectID, dupeID)
}
