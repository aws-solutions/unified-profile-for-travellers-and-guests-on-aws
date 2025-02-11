// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package e2epipeline

import (
	util "tah/upt/source/e2e"
	"tah/upt/source/e2e/generators"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"

	common "tah/upt/schemas/src/tah-common/common"
	"tah/upt/schemas/src/tah-common/lodging"
)

func TestMergeCache(t *testing.T) {
	t.Parallel()
	kinesisRate := 1000
	firstFirstName := "John"
	firstLastName := "Doe"
	secondFirstName := "Josh"
	secondLastName := "Smith"
	// Constants
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}

	domainName, uptHandler, dbConfig, cpConfig, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	initialTravellerId := "testTravellerOne"
	hotelRecordOne, err := generators.GenerateSimpleHotelStayRecords(initialTravellerId, domainName, 2)
	if err != nil {
		t.Fatalf("Error generating hotel record: %v", err)
	}
	guestOne, err := generators.GenerateSpecificGuestProfileRecord(initialTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.LastUpdatedOn = time.Now().Add(-24 * time.Hour)
		gp.LastName = firstLastName
		gp.FirstName = firstFirstName
		gp.ExtendedData = map[string]interface{}{
			"preferredLanguage": "English",
			"loyaltyTier":       "Gold",
			"specialRequests":   []string{"Late Checkout", "High Floor"},
			"lastStay": map[string]interface{}{
				"date":   "2024-01-15",
				"room":   "1204",
				"rating": 4.5,
			},
		}
	})
	if err != nil {
		t.Fatalf("Error generating guest profile record: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, hotelRecordOne)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestOne)

	secondTravellerId := "testTravellerTwo"
	hotelRecordTwo, err := generators.GenerateSimpleHotelStayRecords(secondTravellerId, domainName, 2)
	if err != nil {
		t.Fatalf("Error generating hotel stay record: %v", err)
	}
	guestTwo, err := generators.GenerateSpecificGuestProfileRecord(secondTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.LastUpdatedOn = time.Now().Add(-6 * time.Hour)
		gp.LastName = secondLastName
		gp.FirstName = secondFirstName
	})
	if err != nil {
		t.Fatalf("Error generating guest record: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, hotelRecordTwo)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestTwo)

	var trueConnectIdOne string
	var trueConnectIdTwo string
	err = util.WaitForExpectedCondition(func() bool {
		profileOne, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, initialTravellerId, []string{"hotel_stay"})
		if !exists {
			return exists
		}
		trueConnectIdOne = profileOne.ConnectID
		if len(profileOne.HotelStayRecords) != 2 {
			t.Logf("Expected 2 hotel stay records, got %d", len(profileOne.HotelStayRecords))
			return false
		}
		profileTwo, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, secondTravellerId, []string{"hotel_stay"})
		if !exists {
			return exists
		}
		trueConnectIdTwo = profileTwo.ConnectID
		if profileOne.ConnectID == profileTwo.ConnectID {
			t.Logf("Expected connect ids to be different, got %s and %s", profileOne.ConnectID, profileTwo.ConnectID)
			return false
		}
		if len(profileTwo.HotelStayRecords) != 2 {
			t.Logf("Expected 2 hotel stay records, got %d", len(profileTwo.HotelStayRecords))
			return false
		}
		// Checking Dynamo GSI
		gsiOne := []customerprofileslcs.DynamoTravelerIndex{}
		gsiTwo := []customerprofileslcs.DynamoTravelerIndex{}
		err = dbConfig.FindByGsi(initialTravellerId, customerprofileslcs.DDB_GSI_NAME, customerprofileslcs.DDB_GSI_PK, &gsiOne, db.FindByGsiOptions{Limit: 1})
		if err != nil {
			t.Logf("GSI did not return a value for traveler id %v", initialTravellerId)
			return false
		}
		if len(gsiOne) != 1 {
			t.Logf("GSI search should return 1 value, found %v", len(gsiOne))
			return false
		}
		err = dbConfig.FindByGsi(secondTravellerId, customerprofileslcs.DDB_GSI_NAME, customerprofileslcs.DDB_GSI_PK, &gsiTwo, db.FindByGsiOptions{Limit: 1})
		if err != nil {
			t.Logf("GSI did not return a value for traveler id %v", secondTravellerId)
			return false
		}
		if len(gsiTwo) != 1 {
			t.Logf("GSI search should return 1 value, found %v", len(gsiTwo))
			return false
		}
		if gsiOne[0].ConnectId != trueConnectIdOne {
			t.Logf("Expected connect id to be %s, found %s", trueConnectIdOne, gsiOne[0].ConnectId)
			return false
		}
		if gsiTwo[0].ConnectId != trueConnectIdTwo {
			t.Logf("Expected connect id to be %s, found %s", trueConnectIdTwo, gsiTwo[0].ConnectId)
			return false
		}
		if gsiOne[0].ConnectId == gsiTwo[0].ConnectId {
			t.Logf("Expected connect ids to be different, found %s", gsiOne[0].ConnectId)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}

	_, err = uptHandler.MergeProfiles(domainName, trueConnectIdOne, trueConnectIdTwo)
	if err != nil {
		t.Fatalf("Error merging profiles: %v", err)
	}
	err = util.WaitForExpectedCondition(func() bool {
		profile, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, initialTravellerId, []string{"hotel_stay_revenue_items"})
		if !exists {
			return exists
		}
		trueConnectIdOne = profile.ConnectID
		connectIdTwo, err := uptHandler.GetProfileId(domainName, secondTravellerId)
		if err != nil {
			t.Logf("Error getting profile: %v", err)
			return false
		}
		if profile.ConnectID != connectIdTwo {
			t.Logf("Expected connect ids to be the same, got %s and %s", profile.ConnectID, connectIdTwo)
			return false
		}
		if len(profile.HotelStayRecords) != 4 {
			t.Logf("Expected 4 hotel stay records, got %d", len(profile.HotelStayRecords))
			return false
		}
		if profile.FirstName != secondFirstName {
			t.Logf("Expected %s, got %s", secondFirstName, profile.FirstName)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}

	err = util.WaitForExpectedCondition(func() bool {
		_, profileCp, exists := util.GetProfileFromAccp(t, cpConfig, domainName, trueConnectIdOne, []string{"hotel_stay_revenue_items", "guest_profile"})
		if !exists {
			return exists
		}
		if len(profileCp.ProfileObjects) != 6 {
			t.Logf("Expected 4 hotel stay records, got %d", len(profileCp.ProfileObjects))
			return false
		}
		//need to make sure the "other profile" was deleted from ACCP
		_, profileCp, exists = util.GetProfileFromAccp(t, cpConfig, domainName, trueConnectIdTwo, []string{"hotel_stay_revenue_items"})
		if exists {
			t.Logf("Expected profile to be deleted, but a profile with connect id %v exists in ACCP", trueConnectIdTwo)
			return !exists
		}
		dynamoProfile, dynamoObjects, exists := util.GetProfileRecordsFromDynamo(t, dbConfig, domainName, trueConnectIdOne, []string{"hotel_stay_revenue_items"})
		if !exists {
			return exists
		}
		if dynamoProfile.FirstName != secondFirstName {
			t.Logf("Expected %s, got %s", secondFirstName, dynamoProfile.FirstName)
			return false
		}
		dynamoHotelObjects, ok := dynamoObjects["hotel_stay_revenue_items"]
		if !ok {
			t.Logf("Expected hotel stay records, got %v", dynamoObjects)
			return false
		}
		if len(dynamoHotelObjects) != 4 {
			t.Logf("Expected 4 hotel stay records, 2 guest records and one profile, got %d", len(dynamoObjects["hotel_stay_revenue_items"]))
			return false
		}

		// Checking Dynamo GSI
		gsiOne := []customerprofileslcs.DynamoTravelerIndex{}
		gsiTwo := []customerprofileslcs.DynamoTravelerIndex{}
		err = dbConfig.FindByGsi(initialTravellerId, customerprofileslcs.DDB_GSI_NAME, customerprofileslcs.DDB_GSI_PK, &gsiOne, db.FindByGsiOptions{Limit: 1})
		if err != nil {
			t.Logf("GSI did not return a value for traveler id %v", initialTravellerId)
			return false
		}
		if len(gsiOne) != 1 {
			t.Logf("GSI search should return 1 value, found %v", len(gsiOne))
			return false
		}
		err = dbConfig.FindByGsi(secondTravellerId, customerprofileslcs.DDB_GSI_NAME, customerprofileslcs.DDB_GSI_PK, &gsiTwo, db.FindByGsiOptions{Limit: 1})
		if err != nil {
			t.Logf("GSI did not return a value for traveler id %v", secondTravellerId)
			return false
		}
		if len(gsiTwo) != 1 {
			t.Logf("GSI search should return 1 value, found %v", len(gsiTwo))
			return false
		}
		if gsiOne[0].ConnectId != gsiTwo[0].ConnectId {
			t.Logf("Expected connect ids to be equal, found %v and %v", gsiTwo[0].ConnectId, gsiOne[0].ConnectId)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}
}

func TestSameObjectIdMerge(t *testing.T) {
	t.Parallel()
	firstFirstName := "John"
	firstLastName := "Doe"
	email := "testOne@example.com"
	secondFirstName := "Josh"
	secondLastName := "Smith"
	kinesisRate := 1000
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}

	domainName, uptHandler, dbConfig, cpConfig, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	initialTravellerId := "testTravellerOne"
	secondTravellerId := "testTravellerTwo"
	emailOne := generators.BuildEmail(email)
	recordOne, err := generators.GenerateSpecificGuestProfileRecord(initialTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.Emails = []common.Email{emailOne}
		gp.FirstName = firstFirstName
		gp.LastName = firstLastName
	})
	if err != nil {
		t.Fatalf("Error generating csi record: %v", err)
	}
	recordTwo, err := generators.GenerateSpecificGuestProfileRecord(secondTravellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.Emails = []common.Email{emailOne}
		gp.FirstName = secondFirstName
		gp.LastName = secondLastName
	})
	if err != nil {
		t.Fatalf("Error generating csi record: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, recordOne)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, recordTwo)

	var trueConnectIdOne string
	var trueConnectIdTwo string
	err = util.WaitForExpectedCondition(func() bool {
		profileOne, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, initialTravellerId, []string{"email_history"})
		if !exists {
			t.Logf("Expected profileOne to exist, got %v", exists)
			return exists
		}
		trueConnectIdOne = profileOne.ConnectID
		profileTwo, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, secondTravellerId, []string{"email_history"})
		if !exists {
			t.Logf("Expected profileTwo to exist, got %v", exists)
			return exists
		}
		trueConnectIdTwo = profileTwo.ConnectID
		cpId1, cpProfileOne, exists := util.GetProfileFromAccp(t, cpConfig, domainName, trueConnectIdOne, []string{"email_history"})
		if !exists {
			t.Logf("Expected cpProfileOne to exist, got %v", exists)
			return exists
		}
		cpId2, cpProfileTwo, exists := util.GetProfileFromAccp(t, cpConfig, domainName, trueConnectIdTwo, []string{"email_history"})
		if !exists {
			t.Logf("Expected cpProfileTwo to exist, got %v", exists)
			return exists
		}
		if profileOne.ConnectID == "" {
			t.Logf("Expected connect ids to be present, got %s and %s", profileOne.ConnectID, cpId1)
			return false
		}
		if profileTwo.ConnectID == profileOne.ConnectID {
			t.Logf("Expected connect ids to be different, got %s and %s", profileOne.ConnectID, profileTwo.ConnectID)
			return false
		}
		if cpId2 == cpId1 {
			t.Logf("Expected accp ids to be different, got %s and %s", cpId1, cpId2)
			return false
		}
		if len(profileOne.EmailHistoryRecords) != 1 {
			t.Logf("Expected one csi record, got %d", len(profileOne.EmailHistoryRecords))
			return false
		}
		if len(profileTwo.EmailHistoryRecords) != 1 {
			t.Logf("Expected one csi record, got %d", len(profileTwo.EmailHistoryRecords))
			return false
		}
		if len(cpProfileOne.ProfileObjects) != 1 {
			t.Logf("Expected one accp profile object, got %d", len(cpProfileOne.ProfileObjects))
			return false
		}
		if len(cpProfileTwo.ProfileObjects) != 1 {
			t.Logf("Expected one accp profile object, got %d", len(cpProfileTwo.ProfileObjects))
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}

	_, err = uptHandler.MergeProfiles(domainName, trueConnectIdOne, trueConnectIdTwo)
	if err != nil {
		t.Fatalf("Error merging profiles: %v", err)
	}

	err = util.WaitForExpectedCondition(func() bool {
		profileOne, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, initialTravellerId, []string{"email_history"})
		if !exists {
			return exists
		}
		connectIdTwo, err := uptHandler.GetProfileId(domainName, secondTravellerId)
		if err != nil {
			t.Logf("Error getting profile: %v", err)
			return false
		}
		if profileOne.ConnectID != connectIdTwo {
			t.Logf("Expected connect ids to be the same, got %s and %s", profileOne.ConnectID, connectIdTwo)
			return false
		}
		//only one customer service record, because they got merged with the same id
		if len(profileOne.EmailHistoryRecords) != 1 {
			t.Logf("Expected one profile interaction, got %d", len(profileOne.ClickstreamRecords))
			return false
		}
		_, profileCp, exists := util.GetProfileFromAccp(t, cpConfig, domainName, profileOne.ConnectID, []string{"email_history"})
		if !exists {
			return exists
		}
		if len(profileCp.ProfileObjects) != 1 {
			t.Logf("Expected one profile interaction, got %d", len(profileCp.ProfileObjects))
			return false
		}
		_, dynamoObjects, exists := util.GetProfileRecordsFromDynamo(t, dbConfig, domainName, trueConnectIdOne, []string{"email_history"})
		if !exists {
			return exists
		}
		dynamoCSIObjects, ok := dynamoObjects["email_history"]
		if !ok {
			t.Logf("Expected csi records, got %v", dynamoObjects)
			return false
		}
		if len(dynamoCSIObjects) != 1 {
			t.Logf("Expected one csi record, got %d", len(dynamoCSIObjects))
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}
}

func TestInteractionHistoryAfterMerge(t *testing.T) {
	t.Parallel()
	//setup
	kinesisRate := 1000
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}
	domainName, uptHandler, _, _, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	objectTypeName := "hotel_booking"

	// create same hotel booking for two profiles
	hotelBookingOne := generators.GenerateHotelBooking(true, true)
	travelerIdOne := hotelBookingOne.Holder.ID
	recordOne, err := generators.GetKinesisRecord(domainName, "hotel_booking", hotelBookingOne)
	if err != nil {
		t.Fatalf("Error generating hotel booking record: %v", err)
	}

	travelerIdTwo := "1273426340"
	hotelBookingSecond := hotelBookingOne
	hotelBookingSecond.Holder.FirstName = "Jane"
	hotelBookingSecond.Holder.LastName = "Doe"
	hotelBookingSecond.Holder.ID = travelerIdTwo
	for i := range hotelBookingSecond.Segments {
		hotelBookingSecond.Segments[i].Holder.FirstName = "Jane"
		hotelBookingSecond.Segments[i].Holder.LastName = "Doe"
		hotelBookingSecond.Segments[i].Holder.ID = travelerIdTwo
	}
	recordTwo, err := generators.GetKinesisRecord(domainName, "hotel_booking", hotelBookingSecond)
	if err != nil {
		t.Fatalf("Error generating hotel booking record: %v", err)
	}

	// send records to kinesis
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, []kinesis.Record{recordOne})
	time.Sleep(10 * time.Second) // sleeping so that both booking records have different creation times
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, []kinesis.Record{recordTwo})

	var lcsIdOne, lcsIdTwo string
	var interactionOne, interactionTwo traveler.HotelBooking

	// validate profile got created and both have the same hotel booking object
	err = util.WaitForExpectedCondition(func() bool {
		profileOne, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, travelerIdOne, []string{objectTypeName})
		if !exists {
			t.Logf("Expected profileOne to exist, got %v", exists)
			return exists
		}
		lcsIdOne = profileOne.ConnectID

		profileTwo, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, travelerIdTwo, []string{objectTypeName})
		if !exists {
			t.Logf("Expected profileTwo to exist, got %v", exists)
			return exists
		}
		lcsIdTwo = profileTwo.ConnectID

		if profileOne.ConnectID == "" {
			t.Logf("connectId is missing for profileOne")
		}

		if profileTwo.ConnectID == "" {
			t.Logf("connectId is missing for profileOne")
		}

		if len(profileOne.HotelBookingRecords) == 0 {
			t.Logf("Expected at least one hotel booking record, got 0")
			return false
		}
		interactionOne = profileOne.HotelBookingRecords[0]
		if len(profileTwo.HotelBookingRecords) == 0 {
			t.Logf("Expected at least one hotel booking record, got 0")
			return false
		}
		interactionTwo = profileTwo.HotelBookingRecords[0]

		if interactionOne.AccpObjectID != interactionTwo.AccpObjectID {
			t.Logf("Expected accp object ids to be the same, got %s and %s", interactionOne.AccpObjectID, interactionTwo.AccpObjectID)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}

	// profile one's interaction history
	interactionHistoryOne, err := uptHandler.RetrieveInteractionHistory(domainName, objectTypeName, interactionOne.AccpObjectID, lcsIdOne)
	if err != nil {
		t.Fatalf("Error retrieving interaction history: %v", err)
	}
	if len(interactionHistoryOne) != 1 {
		t.Fatalf("Expected 1 interaction history, got %d", len(interactionHistoryOne))
	}

	// profile two's interaction history
	interactionHistoryTwo, err := uptHandler.RetrieveInteractionHistory(domainName, objectTypeName, interactionTwo.AccpObjectID, lcsIdTwo)
	if err != nil {
		t.Fatalf("Error retrieving interaction history: %v", err)
	}
	if len(interactionHistoryTwo) != 1 {
		t.Fatalf("Expected 1 interaction history, got %d", len(interactionHistoryTwo))
	}

	// profile one's interaction should be older than profile two's
	if interactionHistoryOne[0].Timestamp.After(interactionHistoryTwo[0].Timestamp) {
		t.Fatalf("Expected profile one's interaction to be older than profile two's")
	}

	_, err = uptHandler.MergeProfiles(domainName, lcsIdOne, lcsIdTwo)
	if err != nil {
		t.Fatalf("Error merging profiles: %v", err)
	}

	err = util.WaitForExpectedCondition(func() bool {
		profileOne, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, travelerIdOne, []string{objectTypeName})
		if !exists {
			t.Logf("Expected profileOne to exist, got %v", exists)
			return exists
		}

		profileTwo, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, travelerIdTwo, []string{objectTypeName})
		if !exists {
			t.Logf("Expected profileTwo to exist, got %v", exists)
			return exists
		}

		if profileOne.ConnectID != profileTwo.ConnectID {
			t.Logf("Expected connect ids to be the same, got %s and %s", profileOne.ConnectID, profileTwo.ConnectID)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}

	interactionHistoryAfterMerge, err := uptHandler.RetrieveInteractionHistory(domainName, objectTypeName, interactionOne.AccpObjectID, lcsIdOne)
	if err != nil {
		t.Fatalf("Error retrieving interaction history: %v", err)
	}

	if len(interactionHistoryAfterMerge) != 2 {
		t.Fatalf("Expected 2 history records, got %d", len(interactionHistoryAfterMerge))
	}

	if interactionHistoryAfterMerge[0].MergeType != "INTERACTION_CREATED" {
		t.Fatalf("Expected merge type to be INTERACTION_CREATED, got %s", interactionHistoryAfterMerge[0].MergeType)
	}
	// INTERACTION_CREATED event timestamp should be the same as profile two's timestamp i.e. the one created later
	if interactionHistoryAfterMerge[0].Timestamp != interactionHistoryTwo[0].Timestamp {
		t.Fatalf("Expected timestamps to be the same, got %s and %s", interactionHistoryAfterMerge[0].Timestamp, interactionHistoryAfterMerge[1].Timestamp)
	}
	if interactionHistoryAfterMerge[1].MergeType != customerprofileslcs.MergeTypeManual {
		t.Fatalf("Expected merge type to be MANUAL, got %s", interactionHistoryAfterMerge[1].MergeType)
	}

}
