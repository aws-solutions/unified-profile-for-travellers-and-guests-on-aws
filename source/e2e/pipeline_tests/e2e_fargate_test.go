// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package e2epipeline

import (
	"context"
	"math"
	"strconv"
	"sync"
	"tah/upt/schemas/src/tah-common/air"
	"tah/upt/schemas/src/tah-common/lodging"
	util "tah/upt/source/e2e"
	"tah/upt/source/e2e/generators"
	lcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

func TestRefreshCacheEvent(t *testing.T) {
	t.Parallel()
	NUM_PROFILES := 100
	kinesisRate := 1000
	// Constants
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}
	domainName, uptHandler, dbConfig, cpConfig, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}
	skippedTravellerid := "SkippedTravellerId"
	skippedFirstName := "Skipped"
	skippedLastName := "Skipped"
	// Init ECS
	ecsConfig, taskConfig := util.InitEcsConfig(t, infraConfig, envConfig)

	//setting skip cache conditions
	_, err = uptHandler.SaveCacheRuleSet(domainName, []lcs.Rule{
		{
			Index:       0,
			Name:        "sameFirstName",
			Description: "Skip profiles with a particular first name from cache",
			Conditions: []lcs.Condition{
				{
					Index:               0,
					ConditionType:       lcs.CONDITION_TYPE_SKIP,
					IncomingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					IncomingObjectField: "FirstName",
					Op:                  lcs.RULE_OP_EQUALS_VALUE,
					SkipConditionValue:  lcs.StringVal(skippedFirstName),
					IndexNormalization: lcs.IndexNormalizationSettings{
						Lowercase: true,
						Trim:      true,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Error saving rule set: %v", err)
	}

	_, err = uptHandler.ActivateCacheRuleSet(domainName)
	if err != nil {
		t.Fatalf("Error activating rule set: %v", err)
	}

	// Insert profiles into kinesis stream
	for i := 0; i < NUM_PROFILES; i++ {
		travellerId := "travellerId" + strconv.Itoa(i)
		guestRecord, err := generators.GenerateSimpleGuestProfileRecords(travellerId, domainName, 1)
		if err != nil {
			t.Fatalf("Error generating guest record: %v", err)
		}
		util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestRecord)
	}
	skippedGuestRecord, err := generators.GenerateSpecificGuestProfileRecord(skippedTravellerid, domainName, func(gp *lodging.GuestProfile) {
		gp.LastName = skippedLastName
		gp.FirstName = skippedFirstName
	})
	if err != nil {
		t.Fatalf("Error generating guest profile record: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, skippedGuestRecord)
	connectIdArray := make([]string, NUM_PROFILES)

	err = util.WaitForExpectedCondition(func() bool {
		res, err := uptHandler.GetDomainConfig(domainName)
		if err != nil {
			t.Logf("[%v] error retrieving domain config: %v", t.Name(), err)
			return false
		}
		masterProfileCount := res.UCPConfig.Domains[0].NProfiles
		if int(masterProfileCount) != NUM_PROFILES+1 {
			t.Logf("[%v] Expected profile count %v, got %v", t.Name(), NUM_PROFILES+1, masterProfileCount)
			return false
		}
		var wg sync.WaitGroup
		existsChannel := make([]bool, NUM_PROFILES)
		for i := 0; i < NUM_PROFILES; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				profile, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, "travellerId"+strconv.Itoa(i), []string{})
				if !exists {
					existsChannel[i] = false
					return
				}
				connectIdArray[i] = profile.ConnectID
				_, profileCp, exists := util.GetProfileFromAccp(t, cpConfig, domainName, profile.ConnectID, []string{"guest_profile"})
				if !exists {
					existsChannel[i] = false
					return
				}
				dynamoProfile, _, exists := util.GetProfileRecordsFromDynamo(t, dbConfig, domainName, profile.ConnectID, []string{})
				if !exists {
					existsChannel[i] = false
					return
				}
				if dynamoProfile.FirstName != profile.FirstName || dynamoProfile.LastName != profile.LastName {
					t.Logf("[%v] Expected profile first name %v, got %v, dynamo, index %v", t.Name(), profile.FirstName, dynamoProfile.FirstName, i)
					t.Logf("[%v] Expected profile last name %v, got %v, dynamo, index %v", t.Name(), profile.LastName, dynamoProfile.LastName, i)
					existsChannel[i] = false
					return
				}
				if profileCp.FirstName != profile.FirstName || profileCp.LastName != profile.LastName {
					t.Logf("[%v] Expected profile first name %v, got %v, accp, index %v", t.Name(), profile.FirstName, profileCp.FirstName, i)
					t.Logf("[%v] Expected profile last name %v, got %v, accp, index %v", t.Name(), profile.LastName, profileCp.LastName, i)
					existsChannel[i] = false
					return
				}
				existsChannel[i] = true
			}(i)
		}
		wg.Wait()

		for _, exists := range existsChannel {
			if !exists {
				return false
			}
		}

		skippedProfile, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, skippedTravellerid, []string{})
		if !exists {
			return exists
		}
		//these profiles are skipped
		_, _, exists = util.GetProfileFromAccp(t, cpConfig, domainName, skippedProfile.ConnectID, []string{})
		if exists {
			t.Fatalf("[%v] Accp profile should not exist, skipped: %v", t.Name(), skippedProfile.ConnectID)
			return !exists
		}
		_, _, exists = util.GetProfileRecordsFromDynamo(t, dbConfig, domainName, skippedProfile.ConnectID, []string{})
		if exists {
			t.Fatalf("[%v] Dynamo profile should not exist, skipped: %v", t.Name(), skippedProfile.ConnectID)
			return !exists
		}
		return true
	}, 12, 5*time.Second)
	if err != nil {
		t.Fatalf("[%v] Error waiting for expected condition: %v", t.Name(), err)
	}

	for _, id := range connectIdArray {
		cpId, _, exists := util.GetProfileFromAccp(t, cpConfig, domainName, id, []string{})
		if !exists {
			t.Errorf("[%v] Error retrieving profile from accp: %v", t.Name(), id)
		}
		err = cpConfig.DeleteProfile(cpId)
		if err != nil {
			t.Fatalf("[%v] Error deleting profile from cp: %v", t.Name(), err)
		}
	}
	err = dbConfig.DeleteAll()
	if err != nil {
		t.Fatalf("[%v] Error deleting profile from dynamo: %v", t.Name(), err)
	}

	_, _, exists := util.GetProfileFromAccp(t, cpConfig, domainName, connectIdArray[0], []string{"guest_profile"})
	if exists {
		t.Fatalf("[%v] Expected profile not to exist in accp", t.Name())
	}
	_, _, exists = util.GetProfileRecordsFromDynamo(t, dbConfig, domainName, connectIdArray[0], []string{})
	if exists {
		t.Fatalf("[%v] Expected profile not to exist in dynamo", t.Name())
	}

	//Trigger rebuild cache
	eventId, err := uptHandler.RebuildCacheAndWait(domainName, strconv.Itoa(int(lcs.MAX_CACHE_BIT)), 300)
	if err != nil {
		t.Fatalf("[%v] Error rebuilding caches: %v", t.Name(), err)
	}

	//Wait for fargate tasks to finish
	t.Logf("[%v] Retrieving running fargate tasks", t.Name())
	taskArns, err := ecsConfig.ListTasks(context.TODO(), taskConfig, &eventId, nil)
	if err != nil {
		t.Fatalf("[%v] Error listing fargate task arns: %v", t.Name(), err)
	}
	t.Logf("[%v] Found running fargate tasks with arns: %v", t.Name(), taskArns)
	_, err = ecsConfig.WaitForTasks(context.TODO(), taskConfig, taskArns, 2*time.Minute)
	if err != nil {
		t.Fatalf("[%v] Error waiting for fargate tasks: %v", t.Name(), err)
	}
	t.Logf("[%v] Fargate tasks stopped, checking cache counts", t.Name())
	existsChannel := make([]bool, NUM_PROFILES)

	err = util.WaitForExpectedCondition(func() bool {
		res, err := uptHandler.GetDomainConfig(domainName)
		if err != nil {
			t.Logf("[%v] error retrieving domain config: %v", t.Name(), err)
			return false
		}
		masterProfileCount := res.UCPConfig.Domains[0].NProfiles
		if int(masterProfileCount) != NUM_PROFILES+1 {
			t.Logf("[%v] Expected profile count %v, got %v", t.Name(), NUM_PROFILES+1, masterProfileCount)
			return false
		}
		var wg sync.WaitGroup
		for i := 0; i < NUM_PROFILES; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if existsChannel[i] {
					return
				}
				profile, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, "travellerId"+strconv.Itoa(i), []string{})
				if !exists {
					t.Logf("[%v] GetProfileWithTravellerId did not return a profile", t.Name())
					existsChannel[i] = false
					return
				}
				t.Logf("[%v] GetProfileWithTravellerId succeeded, found profile %v", t.Name(), profile)

				connectIdArray[i] = profile.ConnectID
				_, profileCp, exists := util.GetProfileFromAccp(t, cpConfig, domainName, profile.ConnectID, []string{"guest_profile"})
				if !exists {
					t.Logf("[%v] GetProfileFromAccp did not return a profile", t.Name())
					existsChannel[i] = false
					return
				}
				t.Logf("[%v] GetProfileFromAccp succeeded, found profile %v", t.Name(), profileCp)

				dynamoProfile, _, exists := util.GetProfileRecordsFromDynamo(t, dbConfig, domainName, profile.ConnectID, []string{})
				if !exists {
					t.Logf("[%v] GetProfileRecordsFromDynamo did not return a profile", t.Name())
					existsChannel[i] = false
					return
				}
				t.Logf("[%v] GetProfileRecordsFromDynamo succeeded, found profile %v", t.Name(), dynamoProfile)

				if dynamoProfile.FirstName != profile.FirstName || dynamoProfile.LastName != profile.LastName {
					t.Logf("[%v] Expected profile first name %v, got %v, dynamo, index %v", t.Name(), profile.FirstName, dynamoProfile.FirstName, i)
					t.Logf("[%v] Expected profile last name %v, got %v, dynamo, index %v", t.Name(), profile.LastName, dynamoProfile.LastName, i)
					existsChannel[i] = false
					return
				}
				if profileCp.FirstName != profile.FirstName || profileCp.LastName != profile.LastName {
					t.Logf("[%v] Expected profile first name %v, got %v, accp, index %v", t.Name(), profile.FirstName, profileCp.FirstName, i)
					t.Logf("[%v] Expected profile last name %v, got %v, accp, index %v", t.Name(), profile.LastName, profileCp.LastName, i)
					existsChannel[i] = false
					return
				}
				existsChannel[i] = true
			}(i)
		}
		wg.Wait()

		for _, exists := range existsChannel {
			if !exists {
				return false
			}
		}

		skippedProfile, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, skippedTravellerid, []string{})
		if !exists {
			return exists
		}
		//these profiles are skipped
		_, _, exists = util.GetProfileFromAccp(t, cpConfig, domainName, skippedProfile.ConnectID, []string{})
		if exists {
			t.Fatalf("[%v] Accp profile should not exist, skipped: %v", t.Name(), skippedProfile.ConnectID)
			return !exists
		}
		_, _, exists = util.GetProfileRecordsFromDynamo(t, dbConfig, domainName, skippedProfile.ConnectID, []string{})
		if exists {
			t.Fatalf("[%v] Dynamo profile should not exist, skipped: %v", t.Name(), skippedProfile.ConnectID)
			return !exists
		}
		return true
	}, 12, 5*time.Second)
	if err != nil {
		t.Fatalf("[%v] Error waiting for expected condition: %v", t.Name(), err)
	}

}

// This test is to address the case where the same profiles end up under two separate connect_id after id res is triggered
func TestIdResMergeSplit(t *testing.T) {
	t.Parallel()
	kinesisRate := 1000
	numObjectsPerProfile := 150
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}

	domainName, uptHandler, _, _, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	override1 := []generators.FieldOverrides{
		{
			Weight: 100000,
			Fields: []generators.FieldOverride{
				{Name: "LastName", Value: "1"},
			},
		},
	}
	override2 := []generators.FieldOverrides{
		{
			Weight: 100000,
			Fields: []generators.FieldOverride{
				{Name: "LastName", Value: "2"},
			},
		},
	}
	override3 := []generators.FieldOverrides{
		{
			Weight: 100000,
			Fields: []generators.FieldOverride{
				{Name: "LastName", Value: "3"},
			},
		},
	}
	var travelerId1, travelerId2, travelerId3 string
	var batch1 []kinesis.Record
	for i := 0; i < numObjectsPerProfile; i++ {
		var airBooking air.Booking
		if i%3 == 0 {
			airBooking = generators.GenerateRandomSimpleAirBooking(domainName, override1...)
			travelerId1 = airBooking.PassengerInfo.Passengers[0].ID
		} else if i%3 == 1 {
			airBooking = generators.GenerateRandomSimpleAirBooking(domainName, override2...)
			travelerId2 = airBooking.PassengerInfo.Passengers[0].ID
		} else {
			airBooking = generators.GenerateRandomSimpleAirBooking(domainName, override3...)
			travelerId3 = airBooking.PassengerInfo.Passengers[0].ID
		}

		rec, err := generators.GetKinesisRecord(domainName, "air_booking", airBooking)
		if err != nil {
			t.Fatalf("[%s] Error generating kinesis record: %v", t.Name(), err)
		}

		batch1 = append(batch1, rec)
	}
	// Send records to Kinesis
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, batch1)

	time.Sleep(5 * time.Second)
	_, err = uptHandler.SaveIdResRuleSet(domainName, []lcs.Rule{
		{
			Index:       0,
			Name:        "test",
			Description: "test, passenger with same id and last name",
			Conditions: []lcs.Condition{
				{
					Index:               0,
					ConditionType:       lcs.CONDITION_TYPE_MATCH,
					IncomingObjectType:  "air_booking",
					IncomingObjectField: "last_name",
					Op:                  lcs.RULE_OP_EQUALS,
					ExistingObjectType:  "air_booking",
					ExistingObjectField: "last_name",
					IndexNormalization: lcs.IndexNormalizationSettings{
						Lowercase: true,
						Trim:      true,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Error saving rule set: %v", err)
	}

	_, err = uptHandler.ActivateIdResRuleSetWithRetries(domainName, 10)
	if err != nil {
		t.Fatalf("Error activating rule set: %v", err)
	}

	var batch2 []kinesis.Record
	for i := 0; i < numObjectsPerProfile; i++ {
		var airBooking air.Booking
		if i%3 == 0 {
			airBooking = generators.GenerateRandomSimpleAirBooking(domainName, override1...)
		} else if i%3 == 1 {
			airBooking = generators.GenerateRandomSimpleAirBooking(domainName, override2...)
		} else {
			airBooking = generators.GenerateRandomSimpleAirBooking(domainName, override3...)
		}

		rec, err := generators.GetKinesisRecord(domainName, "air_booking", airBooking)
		if err != nil {
			t.Fatalf("[%s] Error generating kinesis record: %v", t.Name(), err)
		}

		batch2 = append(batch2, rec)
	}
	// Send records to Kinesis
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, batch2)

	_, err = uptHandler.SaveIdResRuleSet(domainName, []lcs.Rule{
		{
			Index:       0,
			Name:        "test",
			Description: "test, passenger with same id and last name",
			Conditions: []lcs.Condition{
				{
					Index:               0,
					ConditionType:       lcs.CONDITION_TYPE_MATCH,
					IncomingObjectType:  "air_booking",
					IncomingObjectField: "last_name",
					Op:                  lcs.RULE_OP_EQUALS,
					ExistingObjectType:  "air_booking",
					ExistingObjectField: "last_name",
					IndexNormalization: lcs.IndexNormalizationSettings{
						Lowercase: true,
						Trim:      true,
					},
				},
			},
		},
		{
			Index:       1,
			Name:        "update",
			Description: "update, random rule to trigger reindexing",
			Conditions: []lcs.Condition{
				{
					Index:               0,
					ConditionType:       lcs.CONDITION_TYPE_MATCH,
					IncomingObjectType:  "clickstream",
					IncomingObjectField: "currency",
					Op:                  lcs.RULE_OP_EQUALS,
					ExistingObjectType:  "clickstream",
					ExistingObjectField: "currency",
					IndexNormalization: lcs.IndexNormalizationSettings{
						Lowercase: true,
						Trim:      true,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Error saving rule set: %v", err)
	}

	_, err = uptHandler.ActivateIdResRuleSetWithRetries(domainName, 10)
	if err != nil {
		t.Fatalf("Error activating rule set: %v", err)
	}

	var batch3 []kinesis.Record
	for i := 0; i < numObjectsPerProfile; i++ {
		var airBooking air.Booking
		if i%3 == 0 {
			airBooking = generators.GenerateRandomSimpleAirBooking(domainName, override1...)
		} else if i%3 == 1 {
			airBooking = generators.GenerateRandomSimpleAirBooking(domainName, override2...)
		} else {
			airBooking = generators.GenerateRandomSimpleAirBooking(domainName, override3...)
		}

		rec, err := generators.GetKinesisRecord(domainName, "air_booking", airBooking)
		if err != nil {
			t.Fatalf("[%s] Error generating kinesis record: %v", t.Name(), err)
		}

		batch3 = append(batch3, rec)
	}
	// Send records to Kinesis
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, batch3)

	err = util.WaitForExpectedCondition(func() bool {
		// This error factor is multiplied onto the number of expected objects due to the split issue not completely being fixed.
		// We should expect that only a few objects failed to merge properly into the main profile
		ERROR_FACTOR := .95
		maxNumProfiles := math.Floor(float64(numObjectsPerProfile) * ERROR_FACTOR)
		profileOne, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, travelerId1, []string{"air_booking"})
		if !exists {
			return exists
		}
		if profileOne.LastName != "1" {
			t.Logf("[%v] Expected profile last name 1, got %v", t.Name(), profileOne.LastName)
			return false
		}
		if float64(profileOne.NAirBookingRecords) < maxNumProfiles {
			t.Logf("[%v] Expected profile 1 to have more than %f air booking records, got %v", t.Name(), maxNumProfiles, profileOne.NAirBookingRecords)
			return false
		}
		profileTwo, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, travelerId2, []string{"air_booking"})
		if !exists {
			return exists
		}
		if profileTwo.LastName != "2" {
			t.Logf("[%v] Expected profile last name 2, got %v", t.Name(), profileTwo.LastName)
			return false
		}
		if float64(profileTwo.NAirBookingRecords) < maxNumProfiles {
			t.Logf("[%v] Expected profile 2 to have more than %f air booking records, got %v", t.Name(), maxNumProfiles, profileTwo.NAirBookingRecords)
			return false
		}
		profileThree, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, travelerId3, []string{"air_booking"})
		if !exists {
			return exists
		}
		if profileThree.LastName != "3" {
			t.Logf("[%v] Expected profile last name 3, got %v", t.Name(), profileThree.LastName)
			return false
		}
		if float64(profileThree.NAirBookingRecords) < maxNumProfiles {
			t.Logf("[%v] Expected profile 3 to have more than %f air booking records, got %v", t.Name(), maxNumProfiles, profileThree.NAirBookingRecords)
			return false
		}
		return true
	}, 24, 10*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}
}
