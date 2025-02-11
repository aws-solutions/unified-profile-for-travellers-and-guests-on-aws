// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e_batch

import (
	"errors"
	"log"
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/s3"
	constants "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/config"
	testutils "tah/upt/source/ucp-common/src/utils/test"
	"tah/upt/source/ucp-common/src/utils/utils"
	upt_sdk "tah/upt/source/ucp-sdk/src"
	"testing"
	"time"
)

type BatchIngestionTestConfig struct {
	ObjectName   string
	GlueJob      string
	SourceBucket string
	Files        []string
	Profiles     map[string][]testutils.ExpectedField
}

type ProfileObjectRef struct {
	ObjectType string
	ObjectID   string
}

const TEST_DATA_DIR = "../../test_data/"

func createTestConfig(cfg config.InfraConfig) []BatchIngestionTestConfig {
	return []BatchIngestionTestConfig{
		{
			ObjectName:   "air_booking",
			Files:        []string{},
			SourceBucket: cfg.BucketAirBooking,
			Profiles:     map[string][]testutils.ExpectedField{},
		},
		{
			ObjectName: "clickstream",
			Files: []string{
				TEST_DATA_DIR + "batch/clickstream_customer_1.jsonl",
			},
			SourceBucket: cfg.BucketClickstream,
			Profiles: map[string][]testutils.ExpectedField{
				"6FpjO7XiF3urkvTD8S468k": {
					{ObjectType: "clickstream", ObjectID: "6FpjO7XiF3urkvTD8S468k" + "_" + "2023-07-03T10:54:57.943Z", FieldName: "bookingId", FieldValue: "78214857"},
					{ObjectType: "clickstream", ObjectID: "6FpjO7XiF3urkvTD8S468k" + "_" + "2023-07-03T10:54:57.943Z", FieldName: "eventType", FieldValue: "confirm_booking"},
					{ObjectType: "clickstream", ObjectID: "6FpjO7XiF3urkvTD8S468k" + "_" + "2023-07-03T10:54:57.943Z", FieldName: "eventTimestamp", FieldValue: "2023-07-03T10:54:57.943Z"},
				},
				"3W10sBWGQlekfdH_FtNmHQ": {
					{ObjectType: "clickstream", ObjectID: "3W10sBWGQlekfdH_FtNmHQ" + "_" + "2023-07-09T02:56:11.727Z", FieldName: "eventType", FieldValue: "login"},
					{ObjectType: "clickstream", ObjectID: "3W10sBWGQlekfdH_FtNmHQ" + "_" + "2023-07-09T02:56:11.727Z", FieldName: "customerId", FieldValue: "334398523"},
				},
				"7pmZvY8kSwWnCKs3rPQD9Q": {
					{ObjectType: "clickstream", ObjectID: "7pmZvY8kSwWnCKs3rPQD9Q" + "_" + "2023-07-02T10:52:32.431Z", FieldName: "languageCode", FieldValue: "EN"},
				},
			},
		},
		{
			ObjectName: "guest_profile",
			Files: []string{
				TEST_DATA_DIR + "batch/guest_profile_customer_1.jsonl",
			},
			SourceBucket: cfg.BucketGuestProfile,
			Profiles: map[string][]testutils.ExpectedField{
				"112779536": {
					{ObjectType: "_profile", FieldName: "firstName", FieldValue: "ROBERT"},
					{ObjectType: "_profile", FieldName: "lastName", FieldValue: "DOWNEY"},
					{ObjectType: "hotel_loyalty", ObjectID: "22841965", FieldName: "level", FieldValue: "Blue"},
				},
				"115791287": {
					{ObjectType: "_profile", FieldName: "firstName", FieldValue: "CHRIS"},
					{ObjectType: "_profile", FieldName: "lastName", FieldValue: "EVANS"},
					{ObjectType: "hotel_loyalty", ObjectID: "23140152", FieldName: "level", FieldValue: "Silver"},
				},
			},
		},
		{
			ObjectName: "hotel_booking",
			Files: []string{
				TEST_DATA_DIR + "batch/hotel_booking_customer_1.jsonl",
			},
			SourceBucket: cfg.BucketHotelBooking,
			Profiles: map[string][]testutils.ExpectedField{
				"396422331": {
					{ObjectType: "_profile", FieldName: "lastName", FieldValue: "JACKSON"},
					{ObjectType: "_profile", FieldName: "homePhoneNumber", FieldValue: "6068707944"},
					{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "roomTypeCode", FieldValue: "SNK"},
				},
			},
		},
		{
			ObjectName: "hotel_stay",
			Files: []string{
				TEST_DATA_DIR + "batch/hotel_stay.jsonl",
			},
			SourceBucket: cfg.BucketHotelStay,
			Profiles: map[string][]testutils.ExpectedField{
				"2950645095": {
					{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Myrtie"},
					{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Sanford"},
				},
			},
		},
		{
			ObjectName: "customer_service_interaction",
			Files: []string{
				TEST_DATA_DIR + "batch/csi.jsonl",
			},
			SourceBucket: cfg.BucketCsi,
			Profiles: map[string][]testutils.ExpectedField{
				"d55a8191-bbf9-48e6-a2ad-cac5cd649500": {
					{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Sophie"},
					{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Bernier"},
				},
			},
		},
		{
			ObjectName: "pax_profile",
			Files: []string{
				TEST_DATA_DIR + "batch/pax_profile.jsonl",
			},
			SourceBucket: cfg.BucketPaxProfile,
			Profiles: map[string][]testutils.ExpectedField{
				"6207056237": {
					{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Sylvester"},
					{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Turcotte"},
					{ObjectType: "air_loyalty", ObjectID: "0890498174", FieldName: "programName", FieldValue: "ascend"},
					{ObjectType: "loyalty_transaction", ObjectID: "9724977878", FieldName: "corporateReferenceNumber", FieldValue: "NDDXHHH2HN"},
				},
			},
		},
	}
}

func TestBatch(t *testing.T) {
	envCfg, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}
	domainName := "batch_tst_" + strings.ToLower(core.GenerateUniqueId())

	uptSdk, err := upt_sdk.Init(infraConfig)
	if err != nil {
		t.Fatalf("Error initializing SDK: %v", err)
	}

	err = uptSdk.SetupTestAuth(domainName)
	if err != nil {
		t.Fatalf("Error setting up auth: %v", err)
	}
	t.Cleanup(func() {
		err := uptSdk.CleanupAuth(domainName)
		if err != nil {
			t.Errorf("Error cleaning up auth: %v", err)
		}
	})

	err = uptSdk.CreateDomainAndWait(domainName, 300)
	if err != nil {
		t.Fatalf("Error creating domain: %v", err)
	}
	t.Cleanup(func() {
		err := uptSdk.DeleteDomainAndWait(domainName)
		if err != nil {
			t.Errorf("Error deleting domain: %v", err)
		}
	})

	bizObjectConfigs := createTestConfig(infraConfig)

	// Run ETL jobs
	year := "2024"
	month := "01"
	day := "01"
	hour := "01"

	log.Printf("Uploading CSV files for all objects")
	for _, c := range bizObjectConfigs {
		sourceBucketHandler := s3.Init(c.SourceBucket, "", envCfg.Region, "", "")
		for _, file := range c.Files {
			pathSegs := strings.Split(file, "/")
			fileName := pathSegs[len(pathSegs)-1]
			s3loc := domainName + "/" + year + "/" + month + "/" + day + "/" + hour + "/" + fileName
			log.Printf("Uploading CSV %s file for %s to s3://%s/%s", c.ObjectName, file, c.SourceBucket, s3loc)
			err := sourceBucketHandler.UploadFile(domainName+"/"+year+"/"+month+"/"+day+"/"+hour+"/"+fileName, file)
			if err != nil {
				t.Fatalf("Error uploading file: %v", err)
			}
		}
	}

	err = uptSdk.RunAllJobsAndWait(domainName, 300)
	if err != nil {
		t.Fatalf("Error running all Glue Jobs: %v", err)
	}

	log.Printf("Verifying ingested profiles for all objects")
	for _, c := range bizObjectConfigs {
		log.Printf("Verifying ingested profiles for for %s", c.ObjectName)
		for profileId, expectedFields := range c.Profiles {
			log.Printf("Verifying profile for %s", profileId)
			profileObjectRefs := buildProfileObjectRefs(expectedFields)
			traveler, err := waitForIngestionCompletion(uptSdk, domainName, profileId, profileObjectRefs, 60)
			if err == nil {
				log.Printf("Profile %s ingested successfully", profileId)
				for _, expectedField := range expectedFields {
					testutils.CheckExpected(t, traveler, expectedField)
				}
			} else {
				t.Fatalf("Error waiting for profile %s creation: %v", profileId, err)
			}
		}
	}
	log.Printf("Batch succeeded for all business objects")
}

// in this function, we use a map to deduplicate object by object type
func buildProfileObjectRefs(expectedFields []testutils.ExpectedField) []ProfileObjectRef {
	objestByRef := map[string]ProfileObjectRef{}
	objRefs := []ProfileObjectRef{}
	for _, expected := range expectedFields {
		if expected.ObjectType != "_profile" {
			objestByRef[expected.ObjectType+expected.ObjectID] = ProfileObjectRef{ObjectType: expected.ObjectType, ObjectID: expected.ObjectID}
		}
	}
	for _, ref := range objestByRef {
		objRefs = append(objRefs, ref)
	}
	return objRefs
}

func waitForMerge(uptSdk upt_sdk.UptHandler, domainName string, profileId1 string, profileId2 string, timeout int) error {
	it := 0
	for it/5 < timeout {
		profile1, err := uptSdk.WaitForProfile(domainName, profileId1, timeout)
		if err != nil {
			return err
		}
		profile2, err := uptSdk.WaitForProfile(domainName, profileId1, timeout)
		if err != nil {
			return err
		}
		if profile1.ConnectID == profile2.ConnectID {
			log.Printf("Merge converged. Profile %s merged into %s", profileId1, profileId2)
			return nil
		} else {
			log.Printf("Merge not converged (cid1:%s,cid2:%s)", profile1.ConnectID, profile2.ConnectID)
		}
		log.Printf("Waiting 5 seconds")
		time.Sleep(5 * time.Second)
		it += 1
	}
	return errors.New("could profile ids %s and %s ddid not converge after merge timout expired")
}

func waitForIngestionCompletion(uptSdk upt_sdk.UptHandler, domainName string, profileId string, profileObjects []ProfileObjectRef, timeout int) (model.Traveller, error) {
	log.Printf("Waiting for profile '%v' full ingestion completion (timeout: %v seconds)", profileId, timeout)
	profile, err := uptSdk.WaitForProfile(domainName, profileId, timeout)
	if err != nil {
		return model.Traveller{}, err
	}
	for _, objectRef := range profileObjects {
		log.Printf("Waiting for object type '%v' with id '%v' to be fully ingested on profile %v", objectRef.ObjectType, objectRef.ObjectID, profile.ConnectID)
		_, err = uptSdk.WaitForProfileObject(domainName, objectRef.ObjectID, profile.ConnectID, objectRef.ObjectType, timeout)
		if err != nil {
			return model.Traveller{}, err
		}
	}
	res, err := uptSdk.RetreiveProfile(domainName, profile.ConnectID, constants.AccpRecordsNames())
	if err != nil {
		return model.Traveller{}, err
	}
	profiles := utils.SafeDereference(res.Profiles)
	if len(profiles) == 0 {
		return model.Traveller{}, errors.New("profile not found")
	}
	return profiles[0], nil
}
