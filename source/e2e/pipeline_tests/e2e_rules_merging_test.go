// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2epipeline

import (
	commonEmail "tah/upt/schemas/src/tah-common/common"
	"tah/upt/schemas/src/tah-common/lodging"
	util "tah/upt/source/e2e"
	"tah/upt/source/e2e/generators"
	lcs "tah/upt/source/storage"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func TestRulesBasedMerge(t *testing.T) {
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

	_, err = uptHandler.SaveIdResRuleSet(domainName, []lcs.Rule{
		{
			Index:       0,
			Name:        "sameFirstName",
			Description: "Merge profiles with matching first name",
			Conditions: []lcs.Condition{
				{
					Index:               0,
					ConditionType:       lcs.CONDITION_TYPE_MATCH,
					IncomingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					IncomingObjectField: "FirstName",
					Op:                  lcs.RULE_OP_EQUALS,
					ExistingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					ExistingObjectField: "FirstName",
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
	// guestTwo is should be newer than guest one; this is based on lastUpdated, so we override it to ensure it's newer than guestOne
	guestTwo, err := generators.GenerateSpecificGuestProfileRecord(secondTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.LastUpdatedOn = gp.LastUpdatedOn.AddDate(2, gofakeit.Number(4, 7), 0)
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
	err = util.WaitForExpectedCondition(func() bool {
		profile, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, initialTravellerId, []string{})
		if !exists {
			return exists
		}
		if profile.LastName != firstLastName {
			t.Logf("Expected last name to be %s, got %s", firstFirstName, profile.LastName)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for expected condition: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestTwo)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestThree)
	err = util.WaitForExpectedCondition(func() bool {
		profile, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, initialTravellerId, []string{})
		if !exists {
			t.Logf("[%s] expected profile to exist, got %v", t.Name(), exists)
			return exists
		}
		if profile.LastName != secondLastName {
			t.Logf("Expected last name to be %s, got %s", secondLastName, profile.LastName)
			return false
		}
		connectIdTwo, err := uptHandler.GetProfileId(domainName, secondTravellerId)
		if err != nil {
			t.Logf("Error getting profile: %v", err)
			return false
		}
		if profile.ConnectID != connectIdTwo {
			t.Logf("Expected connect ids to be the same, got %s and %s", profile.ConnectID, connectIdTwo)
			return false
		}
		connectIdThree, err := uptHandler.GetProfileId(domainName, thirdTravellerId)
		if err != nil {
			t.Logf("Error getting profile: %v", err)
			return false
		}
		if profile.ConnectID == connectIdThree {
			t.Logf("Expected connect ids to be the different, got %s and %s", profile.ConnectID, connectIdThree)
			return false
		}
		_, profileCp, exists := util.GetProfileFromAccp(t, cpConfig, domainName, profile.ConnectID, []string{})
		if !exists {
			return exists
		}
		if profileCp.LastName != secondLastName {
			t.Logf("Expected last name to be %s, got %s", secondLastName, profileCp.LastName)
			return false
		}
		dynamoProfile, _, exists := util.GetProfileRecordsFromDynamo(t, dbConfig, domainName, profile.ConnectID, []string{})
		if !exists {
			t.Logf("[%s] expected dynamoProfile to exist, got %v", t.Name(), exists)
			return exists
		}
		if dynamoProfile.LastName != secondLastName {
			t.Logf("Expected last name to be %s, got %s", secondLastName, dynamoProfile.LastName)
			return false
		}
		profileRecordsTwo := []lcs.DynamoProfileRecord{}
		err = dbConfig.FindStartingWith(connectIdTwo, lcs.PROFILE_OBJECT_TYPE_PREFIX, &profileRecordsTwo)
		if err != nil {
			t.Logf("Error getting profile: %v", err)
			return false
		}
		profileRecordsThree := []lcs.DynamoProfileRecord{}
		err = dbConfig.FindStartingWith(connectIdThree, lcs.PROFILE_OBJECT_TYPE_PREFIX, &profileRecordsThree)
		if err != nil {
			t.Logf("Error getting profile: %v", err)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}
}

// ISSUE: If a profile matched a particular rule, then was deleted
// any profile that would have been merged with that deleted profile
// would no longer be added
// TEST: Add profile, follows a rule, then delete it. Verify profile that matches with
// the deleted profile gets added
func TestRulesBasedMergeWithDelete(t *testing.T) {
	t.Parallel()
	kinesisRate := 1000
	firstFirstName := "John"
	firstLastName := "Doe"
	secondFirstName := "John"
	secondLastName := "Smith"
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}

	domainName, uptHandler, dbConfig, cpConfig, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	_, err = uptHandler.SaveIdResRuleSet(domainName, []lcs.Rule{
		{
			Index:       0,
			Name:        "sameFirstName",
			Description: "Merge profiles with matching first name",
			Conditions: []lcs.Condition{
				{
					Index:               0,
					ConditionType:       lcs.CONDITION_TYPE_MATCH,
					IncomingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					IncomingObjectField: "FirstName",
					Op:                  lcs.RULE_OP_EQUALS,
					ExistingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					ExistingObjectField: "FirstName",
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

	initialTravellerId := "testTravellerOne"
	secondTravellerId := "testTravellerTwo"
	guestOne, err := generators.GenerateSpecificGuestProfileRecord(initialTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.LastName = firstLastName
		gp.FirstName = firstFirstName
	})
	if err != nil {
		t.Fatalf("Error generating guest record: %v", err)
	}
	// guestTwo is should be newer than guest one; this is based on lastUpdated, so we override it to ensure it's newer than guestOne
	guestTwo, err := generators.GenerateSpecificGuestProfileRecord(secondTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.LastUpdatedOn = gp.LastUpdatedOn.AddDate(2, gofakeit.Number(4, 7), 0)
		gp.LastName = secondLastName
		gp.FirstName = secondFirstName
	})
	if err != nil {
		t.Fatalf("Error generating guest record: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestOne)
	connectId := ""
	err = util.WaitForExpectedCondition(func() bool {
		profile, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, initialTravellerId, []string{})
		if !exists {
			return exists
		}
		if profile.LastName != firstLastName {
			t.Logf("Expected last name to be %s, got %s", firstFirstName, profile.LastName)
			return false
		}
		connectId = profile.ConnectID
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for expected condition: %v", err)
	}
	uptHandler.DeleteProfile(domainName, connectId)
	err = util.WaitForExpectedCondition(func() bool {
		_, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, initialTravellerId, []string{})
		return !exists
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for expected condition: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestTwo)
	err = util.WaitForExpectedCondition(func() bool {
		profile, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, secondTravellerId, []string{})
		if !exists {
			t.Logf("[%s] expected profile to exist, got %v", t.Name(), exists)
			return exists
		}
		if profile.LastName != secondLastName {
			t.Logf("Expected last name to be %s, got %s", secondLastName, profile.LastName)
			return false
		}
		_, profileCp, exists := util.GetProfileFromAccp(t, cpConfig, domainName, profile.ConnectID, []string{})
		if !exists {
			return exists
		}
		if profileCp.LastName != secondLastName {
			t.Logf("Expected last name to be %s, got %s", secondLastName, profileCp.LastName)
			return false
		}
		dynamoProfile, _, exists := util.GetProfileRecordsFromDynamo(t, dbConfig, domainName, profile.ConnectID, []string{})
		if !exists {
			t.Logf("[%s] expected dynamoProfile to exist, got %v", t.Name(), exists)
			return exists
		}
		if dynamoProfile.LastName != secondLastName {
			t.Logf("Expected last name to be %s, got %s", secondLastName, dynamoProfile.LastName)
			return false
		}
		profileRecordsTwo := []lcs.DynamoProfileRecord{}
		err = dbConfig.FindStartingWith(profile.ConnectID, lcs.PROFILE_OBJECT_TYPE_PREFIX, &profileRecordsTwo)
		if err != nil {
			t.Logf("Error getting profile: %v", err)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}
}

var rules_MatchAfterMerge = []lcs.Rule{
	{
		Index: 0,
		Name:  "MatchAfterMerge",
		Conditions: []lcs.Condition{
			{
				Index:               0,
				ConditionType:       lcs.CONDITION_TYPE_MATCH,
				IncomingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
				IncomingObjectField: "PersonalEmailAddress",
				Op:                  lcs.RULE_OP_EQUALS,
				ExistingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
				ExistingObjectField: "PersonalEmailAddress",
			},
		},
	},
	{
		Index: 1,
		Conditions: []lcs.Condition{
			{
				Index:               0,
				ConditionType:       lcs.CONDITION_TYPE_MATCH,
				IncomingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
				IncomingObjectField: "FirstName",
				Op:                  lcs.RULE_OP_EQUALS,
				ExistingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
				ExistingObjectField: "FirstName",
				IndexNormalization: lcs.IndexNormalizationSettings{
					Lowercase: true,
					Trim:      true,
				},
			},
			{
				Index:               1,
				ConditionType:       lcs.CONDITION_TYPE_MATCH,
				IncomingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
				IncomingObjectField: "LastName",
				Op:                  lcs.RULE_OP_EQUALS,
				ExistingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
				ExistingObjectField: "LastName",
				IndexNormalization: lcs.IndexNormalizationSettings{
					Lowercase: true,
					Trim:      true,
				},
			},
		},
	},
}

// Issue: Recursive merges were not working, where a merge reveals a prior
// merge and that creates a new match on prior data, which sends another match to
// the merger lambda. This test sees if it was fixed
func TestRecursiveMerges(t *testing.T) {
	t.Parallel()
	kinesisRate := 1000
	firstName := "John"
	lastName := "Doe"
	email := "email@example.com"
	jobTitle := "Engineer"
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}

	domainName, uptHandler, _, cpConfig, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	_, err = uptHandler.SaveIdResRuleSet(domainName, rules_MatchAfterMerge)
	if err != nil {
		t.Fatalf("Error saving rule set: %v", err)
	}

	_, err = uptHandler.ActivateIdResRuleSetWithRetries(domainName, 10)
	if err != nil {
		t.Fatalf("Error activating rule set: %v", err)
	}

	initialTravellerId := "testTravellerOne"
	guestOne, err := generators.GenerateSpecificGuestProfileRecord(initialTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.LastName = lastName
		gp.FirstName = firstName
		gp.JobTitle = jobTitle
	})
	if err != nil {
		t.Fatalf("Error generating guest record: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestOne)

	secondTravellerId := "testTravellerTwo"
	guestTwo, err := generators.GenerateSpecificGuestProfileRecord(secondTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.LastUpdatedOn = gp.LastUpdatedOn.AddDate(2, gofakeit.Number(4, 7), 0)
		gp.LastName = lastName
		gp.FirstName = ""
		gp.Emails = []commonEmail.Email{{Address: email}}
	})
	if err != nil {
		t.Fatalf("Error generating guest record: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestTwo)

	err = util.WaitForExpectedCondition(func() bool {
		profileOne, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, initialTravellerId, []string{})
		if !exists {
			return exists
		}
		profileTwo, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, secondTravellerId, []string{})
		if profileOne.ConnectID == profileTwo.ConnectID {
			t.Logf("Expected connect ids to be different, got %s and %s", profileOne.ConnectID, profileTwo.ConnectID)
			return false
		}
		t.Logf("Profile one: %v", profileTwo)
		return exists
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for expected condition: %v", err)
	}

	//profiles 2 and 3 match on email. Once merged, profile one should be a match, this is the
	//recursive property we are testing.
	thirdTravellerId := "testTravellerThree"
	guestThree, err := generators.GenerateSpecificGuestProfileRecord(thirdTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.LastUpdatedOn = gp.LastUpdatedOn.AddDate(4, gofakeit.Number(4, 7), 0)
		gp.FirstName = firstName
		gp.LastName = ""
		gp.Emails = []commonEmail.Email{{Address: email}}
	})
	if err != nil {
		t.Fatalf("Error generating guest record: %v", err)
	}

	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestThree)
	err = util.WaitForExpectedCondition(func() bool {
		profileOne, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, initialTravellerId, []string{})
		if !exists {
			return exists
		}
		profileTwo, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, secondTravellerId, []string{})
		if !exists {
			return exists
		}
		profileThree, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, thirdTravellerId, []string{})
		if profileThree.LastName != lastName {
			t.Logf("Expected first name to be %s, got %s", firstName, profileThree.FirstName)
			return false
		}
		//job title from profile one should propagate through
		if profileThree.JobTitle != jobTitle {
			t.Logf("Expected job title to be %s, got %s", jobTitle, profileThree.JobTitle)
			return false
		}
		if profileThree.ConnectID != profileTwo.ConnectID {
			t.Logf("Expected connect ids to be different, got %s and %s", profileThree.ConnectID, profileTwo.ConnectID)
			return false
		}
		if profileThree.ConnectID != profileOne.ConnectID {
			t.Logf("Expected connect ids to be different, got %s and %s", profileThree.ConnectID, profileThree.ConnectID)
			return false
		}
		return exists
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for expected condition: %v", err)
	}

	err = util.WaitForExpectedCondition(func() bool {
		cpProfileOne, exists := util.GetProfileFromAccpWithTravellerId(t, cpConfig, domainName, initialTravellerId)
		if !exists {
			return exists
		}
		cpProfileTwo, exists := util.GetProfileFromAccpWithTravellerId(t, cpConfig, domainName, secondTravellerId)
		if !exists {
			return exists
		}
		cpProfileThree, exists := util.GetProfileFromAccpWithTravellerId(t, cpConfig, domainName, thirdTravellerId)
		if !exists {
			return exists
		}
		if cpProfileThree.LastName != lastName {
			t.Logf("Expected first name to be %s, got %s", firstName, cpProfileThree.FirstName)
			return false
		}
		if cpProfileOne.ProfileId != cpProfileTwo.ProfileId {
			t.Logf("Expected profile ids to be different, got %s and %s", cpProfileOne.ProfileId, cpProfileTwo.ProfileId)
			return false
		}
		if cpProfileOne.ProfileId != cpProfileThree.ProfileId {
			t.Logf("Expected profile ids to be different, got %s and %s", cpProfileOne.ProfileId, cpProfileThree.ProfileId)
			return false
		}
		return exists
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for expected condition: %v", err)
	}
}

/*
*
Scenario: Rule change should merge existing profile which now meets the rule
- ingested three profiles; A & B have the same first name, B & C have the same last name
- rule set merges profiles with matching first name
- result: A is merged into B, C remains independent
- rule is now changed so that it merges profiles with matching last name
- final result: C is merged into B (A & B should still remain merged)
*/
func TestRuleChangeMergesExistingProfile(t *testing.T) {
	t.Parallel()
	kinesisRate := 1000

	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("[%s] Error loading configs: %v", t.Name(), err)
	}

	domainName, uptHandler, _, _, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("[%s] Failed to initialize test environment: %v", t.Name(), err)
	}

	// Create initial rule: merge profiles with matching first name
	initialRule := []lcs.Rule{
		{
			Index:       0,
			Name:        "sameFirstName",
			Description: "Merge profiles with matching first name",
			Conditions: []lcs.Condition{
				{
					Index:               0,
					ConditionType:       lcs.CONDITION_TYPE_MATCH,
					IncomingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					IncomingObjectField: "FirstName",
					Op:                  lcs.RULE_OP_EQUALS,
					ExistingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					ExistingObjectField: "FirstName",
					IndexNormalization: lcs.IndexNormalizationSettings{
						Lowercase: true,
						Trim:      true,
					},
				},
			},
		},
	}

	// Save rule set
	_, err = uptHandler.SaveIdResRuleSet(domainName, initialRule)
	if err != nil {
		t.Fatalf("[%s] Error saving initial rule set: %v", t.Name(), err)
	}

	// Activate rule set
	_, err = uptHandler.ActivateIdResRuleSetWithRetries(domainName, 10)
	if err != nil {
		t.Fatalf("[%s] Error activating initial rule set: %v", t.Name(), err)
	}

	sameFirstName := "John"
	otherFirstName := "Jane"
	sameLastName := "Smith"
	otherLastName := "Doe"
	// Create three profiles
	profileA, err := generators.GenerateSpecificGuestProfileRecord("travelerA", domainName, func(gp *lodging.GuestProfile) {
		gp.FirstName = sameFirstName // "John"
		gp.LastName = otherLastName  // "Doe"
	})
	if err != nil {
		t.Fatalf("[%s] Error generating profile A: %v", t.Name(), err)
	}

	profileB, err := generators.GenerateSpecificGuestProfileRecord("travelerB", domainName, func(gp *lodging.GuestProfile) {
		gp.FirstName = sameFirstName // "John"
		gp.LastName = sameLastName   // "Smith"
		gp.LastUpdatedOn = gp.LastUpdatedOn.AddDate(2, gofakeit.Number(4, 7), 0)
	})
	if err != nil {
		t.Fatalf("[%s] Error generating profile B: %v", t.Name(), err)
	}

	profileC, err := generators.GenerateSpecificGuestProfileRecord("travelerC", domainName, func(gp *lodging.GuestProfile) {
		gp.FirstName = otherFirstName // "Jane"
		gp.LastName = sameLastName    // "Smith"
	})
	if err != nil {
		t.Fatalf("[%s] Error generating profile C: %v", t.Name(), err)
	}

	// Send profiles to Kinesis
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, profileA)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, profileB)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, profileC)

	// Wait and verify that A and B are merged, but C is separate
	err = util.WaitForExpectedCondition(func() bool {
		profileA, existsA := util.GetProfileWithTravellerId(t, uptHandler, domainName, "travelerA", []string{})
		profileB, existsB := util.GetProfileWithTravellerId(t, uptHandler, domainName, "travelerB", []string{})
		profileC, existsC := util.GetProfileWithTravellerId(t, uptHandler, domainName, "travelerC", []string{})

		if !existsA || !existsB || !existsC {
			return false
		}

		if profileA.ConnectID != profileB.ConnectID {
			t.Logf("[%s] Expected profiles A and B to be merged, but they have different ConnectIDs", t.Name())
			return false
		}

		if profileA.ConnectID == profileC.ConnectID {
			t.Logf("[%s] Expected profile C to be separate, but it has the same ConnectID as A and B", t.Name())
			return false
		}

		if profileA.FirstName != sameFirstName {
			t.Logf("[%s] Expected first name to be %s, got %s", t.Name(), sameFirstName, profileA.FirstName)
			return false
		}

		if profileB.FirstName != sameFirstName {
			t.Logf("[%s] Expected first name to be %s, got %s", t.Name(), sameFirstName, profileB.FirstName)
			return false
		}

		if profileC.FirstName != otherFirstName {
			t.Logf("[%s] Expected first name to be %s, got %s", t.Name(), otherFirstName, profileC.FirstName)
			return false
		}

		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("[%s] Error waiting for initial merge condition: %v", t.Name(), err)
	}

	// Change the rule to merge profiles with matching last name
	updatedRule := []lcs.Rule{
		{
			Index:       0,
			Name:        "sameLastName",
			Description: "Merge profiles with matching last name",
			Conditions: []lcs.Condition{
				{
					Index:               0,
					ConditionType:       lcs.CONDITION_TYPE_MATCH,
					IncomingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					IncomingObjectField: "LastName",
					Op:                  lcs.RULE_OP_EQUALS,
					ExistingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					ExistingObjectField: "LastName",
					IndexNormalization: lcs.IndexNormalizationSettings{
						Lowercase: true,
						Trim:      true,
					},
				},
			},
		},
	}

	_, err = uptHandler.SaveIdResRuleSet(domainName, updatedRule)
	if err != nil {
		t.Fatalf("[%s] Error saving updated rule set: %v", t.Name(), err)
	}

	_, err = uptHandler.ActivateIdResRuleSetWithRetries(domainName, 10)
	if err != nil {
		t.Fatalf("[%s] Error activating updated rule set: %v", t.Name(), err)
	}

	// Wait for ID resolution job to complete and verify all profiles are merged
	err = util.WaitForExpectedCondition(func() bool {
		profileA, existsA := util.GetProfileWithTravellerId(t, uptHandler, domainName, "travelerA", []string{})
		profileB, existsB := util.GetProfileWithTravellerId(t, uptHandler, domainName, "travelerB", []string{})
		profileC, existsC := util.GetProfileWithTravellerId(t, uptHandler, domainName, "travelerC", []string{})

		if !existsA || !existsB || !existsC {
			return false
		}

		if profileA.ConnectID != profileC.ConnectID {
			t.Logf("[%s] Expected profiles A and C to be merged, but they have different ConnectIDs", t.Name())
			return false
		}

		if profileA.ConnectID != profileB.ConnectID {
			t.Logf("[%s] Expected profiles A and B to be merged, but they have different ConnectIDs", t.Name())
			return false
		}

		if profileA.LastName != sameLastName {
			t.Logf("[%s] Expected last name to be %s, got %s", t.Name(), sameLastName, profileA.LastName)
			return false
		}

		if profileB.LastName != sameLastName {
			t.Logf("[%s] Expected last name to be %s, got %s", t.Name(), sameLastName, profileB.LastName)
			return false
		}

		if profileC.LastName != sameLastName {
			t.Logf("[%s] Expected last name to be %s, got %s", t.Name(), sameLastName, profileC.LastName)
			return false
		}

		return true
	}, 10, 10*time.Second)
	if err != nil {
		t.Fatalf("[%s] Error waiting for final merge condition: %v", t.Name(), err)
	}

	t.Logf("[%s] Test completed successfully: Profiles merged correctly after rule change", t.Name())
}
