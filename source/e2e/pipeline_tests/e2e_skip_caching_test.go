// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2epipeline

import (
	"tah/upt/schemas/src/tah-common/lodging"
	util "tah/upt/source/e2e"
	"tah/upt/source/e2e/generators"
	lcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/customerprofiles"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

func TestCacheSkipConditions(t *testing.T) {
	t.Parallel()
	kinesisRate := 1000
	firstFirstName := "John"
	firstLastName := "Doe"
	secondFirstName := "John"
	secondLastName := "Smith"
	thirdFirstName := "Josh"
	thirdLastName := "Cena"
	// Constants
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}

	domainName, uptHandler, dbConfig, cpConfig, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	_, err = uptHandler.SaveCacheRuleSet(domainName, []lcs.Rule{
		{
			Index:       0,
			Name:        "sameFirstName",
			Description: "Merge profiles with matching first name",
			Conditions: []lcs.Condition{
				{
					Index:               0,
					ConditionType:       lcs.CONDITION_TYPE_SKIP,
					IncomingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					IncomingObjectField: "FirstName",
					Op:                  lcs.RULE_OP_EQUALS_VALUE,
					SkipConditionValue:  lcs.StringVal(firstFirstName),
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

	initialTravellerId := "testTravellerOne"
	secondTravellerId := "testTravellerTwo"
	thirdTravellerId := "testTravellerThree"
	guestOne, err := generators.GenerateSpecificGuestProfileRecord(initialTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.LastName = firstLastName
		gp.FirstName = firstFirstName
	})
	if err != nil {
		t.Fatalf("Error generating guest record: %v", err)
	}
	guestTwo, err := generators.GenerateSpecificGuestProfileRecord(secondTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.LastName = secondLastName
		gp.FirstName = secondFirstName
	})
	if err != nil {
		t.Fatalf("Error generating guest record: %v", err)
	}
	guestThree, err := generators.GenerateSpecificGuestProfileRecord(thirdTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.LastName = thirdLastName
		gp.FirstName = thirdFirstName
	})
	if err != nil {
		t.Fatalf("Error generating guest record: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestOne)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestTwo)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestThree)
	err = util.WaitForExpectedCondition(func() bool {
		connectIdOne, err := uptHandler.GetProfileId(domainName, initialTravellerId)
		if err != nil {
			t.Logf("Error getting profile: %v", err)
			return false
		}
		cpId1, err := cpConfig.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, connectIdOne)
		if err != nil {
			t.Logf("Error getting accp profile: %v", err)
			return false
		}
		connectIdTwo, err := uptHandler.GetProfileId(domainName, secondTravellerId)
		if err != nil {
			t.Logf("Error getting profile: %v", err)
			return false
		}
		cpId2, err := cpConfig.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, connectIdTwo)
		if err != nil {
			t.Logf("Error getting accp profile: %v", err)
			return false
		}
		connectIdThree, err := uptHandler.GetProfileId(domainName, thirdTravellerId)
		if err != nil {
			t.Logf("Error getting profile: %v", err)
			return false
		}
		cpId3, err := cpConfig.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, connectIdThree)
		if err != nil {
			t.Logf("Error getting accp profile: %v", err)
			return false
		}
		if cpId1 != "" {
			t.Logf("Expected first connect id to be skipped, was not %v", cpId1)
			return false
		}
		if cpId2 != "" {
			t.Logf("Expected second connect id to be skipped, was not %v", cpId2)
			return false
		}
		if cpId3 == "" {
			t.Logf("Expected third connect id to be not be skipped, was not %v", cpId3)
			return false
		}
		profileRecords := []lcs.DynamoProfileRecord{}
		err = dbConfig.FindStartingWith(connectIdOne, lcs.PROFILE_OBJECT_TYPE_PREFIX, &profileRecords)
		if err != nil {
			t.Logf("Error getting dynamo profile: %v", err)
			return false
		}
		profileRecordsTwo := []lcs.DynamoProfileRecord{}
		err = dbConfig.FindStartingWith(connectIdTwo, lcs.PROFILE_OBJECT_TYPE_PREFIX, &profileRecordsTwo)
		if err != nil {
			t.Logf("Error getting dynamo profile: %v", err)
			return false
		}
		profileRecordsThree := []lcs.DynamoProfileRecord{}
		err = dbConfig.FindStartingWith(connectIdThree, lcs.PROFILE_OBJECT_TYPE_PREFIX, &profileRecordsThree)
		if err != nil {
			t.Logf("Error getting dynamo profile: %v", err)
			return false
		}
		if len(profileRecords) != 0 {
			t.Logf("Expected no profile records, got %d", len(profileRecords))
			return false
		}
		if len(profileRecordsTwo) != 0 {
			t.Logf("Expected no profile records, got %d", len(profileRecords))
			return false
		}
		if len(profileRecordsThree) != 2 {
			t.Logf("Expected 2 profile records, got %d", len(profileRecords))
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}
}
