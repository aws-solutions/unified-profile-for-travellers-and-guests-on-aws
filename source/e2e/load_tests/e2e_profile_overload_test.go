// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package loadTests

import (
	"log"
	util "tah/upt/source/e2e"
	"tah/upt/source/e2e/generators"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"

	"github.com/google/uuid"
)

/*
	DO NOT RUN THIS TEST WITHIN THE PIPELINE
*/

func TestProfileOverload(t *testing.T) {
	t.Parallel()
	// Constants
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}
	domainName, uptHandler, dynamoCache, cpCache, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	// Create Profile
	log.Printf("[%v] Creating profile", t.Name())
	travellerId := uuid.NewString()

	totalRecords := 0

	// Set numRecords to generate {numRecords} of each object type
	// Set kinesisRate to insert into kinesis in batch of {kinesisRate}
	numRecords := 5
	kinesisRate := 1000

	var records []kinesis.Record

	// air booking objects
	log.Printf("[%v] generating %v air_booking records", t.Name(), numRecords)
	records, err = generators.GenerateSimpleAirBookingRecords(travellerId, domainName, numRecords)
	if err != nil {
		t.Fatalf("[%v] error generating records: %v", t.Name(), err)
	}
	log.Printf("[%v] sending %v air_booking records to kinesis", t.Name(), numRecords)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, records)
	totalRecords += len(records)

	// clickstream objects
	log.Printf("[%v] generating %v clickstream records", t.Name(), numRecords)
	records, err = generators.GenerateSimpleClickstreamRecords(travellerId, domainName, numRecords)
	if err != nil {
		t.Fatalf("[%v] error generating records: %v", t.Name(), err)
	}
	log.Printf("[%v] sending %v clickstream records to kinesis", t.Name(), numRecords)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, records)
	totalRecords += len(records)

	// hotel booking objects
	log.Printf("[%v] generating %v hotel booking records", t.Name(), numRecords)
	records, err = generators.GenerateSimpleHotelBookingRecords(travellerId, domainName, numRecords)
	if err != nil {
		t.Fatalf("[%v] error generating records: %v", t.Name(), err)
	}
	log.Printf("[%v] sending %v hotel_booking records to kinesis", t.Name(), numRecords)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, records)
	totalRecords += len(records)

	// customer service interactions objects
	log.Printf("[%v] generating %v csi records", t.Name(), numRecords)
	records, err = generators.GenerateSimpleCsiRecords(domainName, numRecords)
	if err != nil {
		t.Fatalf("[%v] error generating records: %v", t.Name(), err)
	}
	log.Printf("[%v] sending %v customer_service_interaction records to kinesis", t.Name(), numRecords)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, records)
	totalRecords += len(records)

	// hotel stay objects
	log.Printf("[%v] generating %v hotel stay records", t.Name(), numRecords)
	records, err = generators.GenerateSimpleHotelStayRecords(travellerId, domainName, numRecords)
	if err != nil {
		t.Fatalf("[%v] error generating records: %v", t.Name(), err)
	}
	log.Printf("[%v] sending %v hotel_stay records to kinesis", t.Name(), numRecords)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, records)
	totalRecords += len(records)

	// Not inserting guest profile, pax profile, loyalty, phone, etc. as they're negligible compared to how much these objects will be ingested

	err = util.WaitForExpectedCondition(func() bool {
		res, err := uptHandler.GetDomainConfig(domainName)
		if err != nil {
			t.Fatalf("[%v] error retrieving domain config: %v", t.Name(), err)
		}
		dom := res.UCPConfig.Domains[0]
		if dom.NProfiles != 1 {
			log.Printf("[%v] invalid number of profiles found expected 1, found %v", t.Name(), dom.NProfiles)
			return false
		}
		if int(dom.NObjects) != totalRecords {
			log.Printf("[%v] invalid number of objects found expected %v, found %v", t.Name(), totalRecords, dom.NObjects)
			return false
		}
		return true
	}, 60, 10*time.Second)
	if err != nil {
		t.Fatalf("[%v] error waiting for profile creation: %v", t.Name(), err)
	}
	log.Printf("[%v] successfully found all profiles and objects in Aurora", t.Name())

	// Check Dynamo Cache
	// DynamoDB Cache should have 1 profile record + {totalRecords} object records
	err = util.WaitForExpectedCondition(func() bool {
		dynamoCount := util.GetDynamoCacheCounts(t, dynamoCache)
		if dynamoCount != (totalRecords + 1) {
			log.Printf("[%v] dynamo should have %v records, found %v", t.Name(), totalRecords+1, dynamoCount)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("[%v] dynamo does not have the correct number of records: %v", t.Name(), err)
	}
	log.Printf("[%v] successfully found all profiles and objects in DynamoDB", t.Name())

	// Check CP Cache
	// CP Cache should have 1 profile record and {totalRecords} object records
	err = util.WaitForExpectedCondition(func() bool {
		nProfiles, _ := util.GetCpCacheCounts(t, cpCache)
		if int(nProfiles) != 1 {
			log.Printf("[%v] CP should have 1 profile, found %v", t.Name(), nProfiles)
			return false
		}
		// Avoiding CP Object check since CP hit's a limit for inserting too many objects into the same profile
		// if int(nObjects) != totalRecords {
		// 	log.Printf("[%v] CP should have %v objects, found %v", t.Name(), totalRecords, nObjects)
		// 	return false
		// }
		return true
	}, 10, 5*time.Second)
	if err != nil {
		t.Fatalf("[%v] dynamo does not have the correct number of records: %v", t.Name(), err)
	}
	log.Printf("[%v] successfully found all profiles and objects in CP", t.Name())
}
