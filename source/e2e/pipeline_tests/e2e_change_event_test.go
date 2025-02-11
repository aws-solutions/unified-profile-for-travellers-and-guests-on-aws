// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package e2epipeline

import (
	"fmt"
	"log"
	util "tah/upt/source/e2e"
	"tah/upt/source/e2e/generators"
	eb "tah/upt/source/tah-core/eventbridge"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestChangeEvent(t *testing.T) {
	t.Parallel()
	// Constants
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}
	domainName, uptHandler, _, _, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	// EventBridge & Queue
	ebCfg := eb.InitWithBus(envConfig.Region, infraConfig.EventBridgeBusName, "", "")
	sqsConfig := sqs.Init(envConfig.Region, "", "")
	queueUrl, err := sqsConfig.Create(domainName)
	if err != nil {
		t.Fatalf("[%v] Error creating SQS queue: %s", t.Name(), err)
	}
	t.Cleanup(func() {
		err = sqsConfig.DeleteByName(domainName)
		if err != nil {
			t.Errorf("[%v] error deleting sqs queue", t.Name())
		}
	})
	eventPattern := fmt.Sprintf("{\"source\": [\"ucp-%s\"]}", domainName) // rule fires on events for this domain only
	err = ebCfg.CreateSQSTarget(domainName, infraConfig.EventBridgeBusName, queueUrl, eventPattern)
	if err != nil {
		t.Fatalf("[%v] Error creating SQS target: %s", t.Name(), err)
	}
	t.Cleanup(func() {
		err = ebCfg.DeleteRuleByName(domainName)
		if err != nil {
			t.Errorf("[%v] error deleting event bridge rule", t.Name())
		}
	})

	// Create Profile
	log.Printf("[%v] Creating profile with 10 air_booking objects", t.Name())
	travellerId := uuid.NewString()

	totalRecords := 0

	NUM_OBJECTS := 5

	var records []kinesis.Record

	// air booking objects
	log.Printf("[%v] generating %v air_booking records", t.Name(), NUM_OBJECTS)
	records, err = generators.GenerateSimpleAirBookingRecords(travellerId, domainName, NUM_OBJECTS)
	if err != nil {
		t.Fatalf("[%v] error generating records: %v", t.Name(), err)
	}
	log.Printf("[%v] sending %v air_booking records to kinesis", t.Name(), NUM_OBJECTS)
	util.SendKinesisBatch(t, realTimeStream, NUM_OBJECTS, records)
	totalRecords += len(records)

	// Check Aurora
	err = util.WaitForExpectedCondition(func() bool {
		res, err := uptHandler.GetDomainConfig(domainName)
		if err != nil {
			t.Fatalf("[%v] error retrieving domain config: %v", t.Name(), err)
		}
		dom := res.UCPConfig.Domains[0]
		if dom.NProfiles != 1 {
			t.Logf("[%v] invalid number of profiles found expected 1, found %v", t.Name(), dom.NProfiles)
			return false
		}
		if int(dom.NObjects) != totalRecords {
			t.Logf("[%v] invalid number of objects found expected %v, found %v", t.Name(), totalRecords, dom.NObjects)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("[%v] error waiting for profile creation: %v", t.Name(), err)
	}
	log.Printf("[%v] successfully found all profiles and objects in Aurora", t.Name())

	// Check Event Bus
	// Validating EventBridge export
	log.Printf("[%v] Getting events from SQS", t.Name())
	EXPECTED_PROFILE_COUNT := NUM_OBJECTS
	EXPECTED_OBJECT_COUNT := NUM_OBJECTS
	EXPECTED_UPDATED_COUNT := (NUM_OBJECTS - 1) * 2
	EXPECTED_CREATED_COUNT := 2
	var profileCount, objectCount, updatedCount, createdCount int
	ebEvents, err := util.GetEventsFromSqs(&sqsConfig, 120)
	if err != nil {
		log.Printf("[%v] Non blocking Error getting events from SQS: %s", t.Name(), err)
	} else {
		log.Printf("[%v] Successfully got Events: %+v", t.Name(), ebEvents)
	}
	for _, eventBatch := range ebEvents {
		log.Printf("[%v] found %v events in batch", t.Name(), len(eventBatch.Detail.Events))
		for _, event := range eventBatch.Detail.Events {
			log.Printf("[%v] found event %v", t.Name(), event)
			if event.EventType == "UPDATED" {
				updatedCount++
			}
			if event.EventType == "CREATED" {
				createdCount++
			}
			if event.ObjectType == "_profile" {
				if event.Data.TravellerID != travellerId {
					t.Errorf("Profile event should have traveller id at profile level: %v", event)
				}
				profileCount++
			}
			if event.ObjectType == "air_booking" {
				if event.Data.AirBookingRecords[0].TravellerID != travellerId {
					t.Errorf("Profile event should have traveller id at air booking object level: %v", event)
				}
				objectCount++
			}
		}
	}
	if profileCount != EXPECTED_PROFILE_COUNT {
		t.Fatalf("[%v] expected %v profile events, found %v", t.Name(), EXPECTED_PROFILE_COUNT, profileCount)
	}
	if objectCount != EXPECTED_OBJECT_COUNT {
		t.Fatalf("[%v] expected %v object events, found %v", t.Name(), EXPECTED_OBJECT_COUNT, objectCount)
	}
	if updatedCount != EXPECTED_UPDATED_COUNT {
		t.Fatalf("[%v] expected %v UPDATED events, found %v", t.Name(), EXPECTED_UPDATED_COUNT, updatedCount)
	}
	if createdCount != EXPECTED_CREATED_COUNT {
		t.Fatalf("[%v] expected %v CREATED events, found %v", t.Name(), EXPECTED_CREATED_COUNT, createdCount)
	}
}
