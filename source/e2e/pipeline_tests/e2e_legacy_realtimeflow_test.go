// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2epipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	common "tah/upt/schemas/src/tah-common/common"
	"tah/upt/schemas/src/tah-common/lodging"
	util "tah/upt/source/e2e"
	"tah/upt/source/e2e/generators"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	customerprofiles "tah/upt/source/tah-core/customerprofiles"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	kinesis "tah/upt/source/tah-core/kinesis"
	"tah/upt/source/ucp-common/src/utils/config"
	"tah/upt/source/ucp-common/src/utils/utils"

	upt_sdk "tah/upt/source/ucp-sdk/src"

	corelodge "tah/upt/schemas/src/tah-common/core"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awskinesis "github.com/aws/aws-sdk-go/service/kinesis"

	"testing"
	"time"
)

var id = `0660005985`
var objectId = `2519308635`
var secondObjectId = `2519308636`
var emailId = `maryjanecole@example.com`
var secondEmailId = `lorenzcollier@example.com`
var oldPointUnit = `reward points`

var guestId = `0771115986`
var partialObjectType = "hotel_loyalty"

// note: transform will parse these as integers, so test could fail if using floats for the strings
var oldPointsToNextLevel = `9673.0`
var newPointsToNextLevel = `10.0`
var oldPoints = `884081.0`
var programName = `ascend`
var oldLevel = `platinium`
var newLevel = `gold`
var loyaltyJoined = `2023-03-08T03:43:20.871590Z`

var guestProfileRecord = buildGuestProfileRecordWithLoyaltyTwoEmails(id, secondEmailId, emailId, objectId, programName, oldPoints, oldPointUnit, oldPointsToNextLevel, oldLevel, loyaltyJoined)

var profileToMergeId1 = "guest_1"
var profileToMergeId2 = "guest_2"
var profileToMergeId3 = "guest_3"
var profileToMergeEmail1 = "guest1@example.com"
var profileToMergeEmail2 = "guest2@example.com"
var middlename = "testMiddlename"
var guestProfileRecordForMergeMain = buildGuestProfileRecord(profileToMergeId1, "", profileToMergeEmail1, time.Now())
var guestProfileRecordFormMergeDupe = buildGuestProfileRecord(profileToMergeId2, middlename, profileToMergeEmail2, time.Now().Add(time.Hour))

type BusinessObjectRecord struct {
	Domain             string                 `json:"domain"`
	ObjectType         string                 `json:"objectType"`
	ModelVersion       string                 `json:"modelVersion"`
	Data               map[string]interface{} `json:"data"`
	Mode               string                 `json:"mode"`
	MergeModeProfileID string                 `json:"mergeModeProfileId"`
	PartialModeOptions PartialModeOptions     `json:"partialModeOptions"`
}

type PartialModeOptions struct {
	Fields []string `json:"fields"`
}

//These tests are legacy tests, designed to test the real time ingestor through kinesis.
//Specific older features tested here are not covered under our more modern ingestion e2e tests,
//such as partial and merge modes in the kinesis wrapper. These tests were ported over from the real time ingestor and remain
//the same excepting one change: hardcoded data was removed and the tests were made to use the generators.
//These will run as part of the e2e package.

// overriding httpclient to prevent  Error: send request failed caused by: Post "https://kinesis.eu-central-1.amazonaws.com/": EOF
// https://github.com/aws/aws-sdk-go/issues/206
func InitKinesis(streamName, region, solutionId, solutionVersion string) *kinesis.Config {
	mySession := session.Must(session.NewSession())
	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	cfg := aws.NewConfig().WithRegion(region).WithMaxRetries(0).WithHTTPClient(httpClient)
	svc := awskinesis.New(mySession, cfg)
	return &kinesis.Config{
		Client:    svc,
		Region:    region,
		Stream:    streamName,
		BatchSize: 10,
	}
}

func TestKinesis(t *testing.T) {
	// Set up clients
	envCfg, infraCfg, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("error loading initial config %v", err)
	}
	t.Logf("[E2EDataStream] Running E2E test for data stream")
	outcfg := kinesis.Init(infraCfg.RealTimeOutputStreamName, envCfg.Region, "", "", core.LogLevelDebug)
	outcfg.InitConsumer(kinesis.ITERATOR_TYPE_LATEST)
	if err != nil {
		t.Fatalf("Error initializing profile storage %v", err)
	}

	// Set up test environment
	domainName, uptSdk, _, _, incfg, err := util.InitializeEndToEndTestEnvironment(t, infraCfg, envCfg)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	t.Logf("[E2EDataStream] 1- sending clickstream records to kinesis stream %v", infraCfg.RealTimeIngestorStreamName)
	//Full profiles
	clickstreamRecord := buildRecord("clickstream", "clickstream", domainName, "", 0)
	clickstreamSerialized, _ := json.Marshal(clickstreamRecord)
	hotelBookingRecord := buildRecord("hotel_booking", "hotel_booking", domainName, "", 0)
	hotelBookingSerialized, _ := json.Marshal(hotelBookingRecord)
	guestProfileRecord := buildRecordFromString("guest_profile", guestProfileRecord, domainName, "", "")
	guestProfileSerialized, _ := json.Marshal(guestProfileRecord)
	guestProfileMsgForMergeMain := buildRecordFromString("guest_profile", guestProfileRecordForMergeMain, domainName, "", "")
	guestProfileMsgForMergeMainSerialized, _ := json.Marshal(guestProfileMsgForMergeMain)
	guestProfileMsgForMergeDupe := buildRecordFromString("guest_profile", guestProfileRecordFormMergeDupe, domainName, "", "")
	guestProfileMsgForMergeDupeSerialized, _ := json.Marshal(guestProfileMsgForMergeDupe)
	guestProfileMergeRq := buildRecordFromString("guest_profile", guestProfileRecordFormMergeDupe, domainName, "merge", profileToMergeId1)
	guestProfileMergeRqSerialized, _ := json.Marshal(guestProfileMergeRq)

	/////////////////////////////
	//Full Record Tests
	/////////////////////////////

	records := []kinesis.Record{
		{
			Pk:   clickstreamRecord.ObjectType + "-" + clickstreamRecord.ModelVersion + "-1",
			Data: string(clickstreamSerialized),
		},
		{
			Pk:   hotelBookingRecord.ObjectType + "-" + hotelBookingRecord.ModelVersion + "-2",
			Data: string(hotelBookingSerialized),
		},
		{
			Pk:   guestProfileRecord.ObjectType + "-" + guestProfileRecord.ModelVersion + "-3",
			Data: string(guestProfileSerialized),
		},
	}
	t.Logf("[E2EDataStream] 2- sending records %+v to kinesis stream %v", records, infraCfg.RealTimeIngestorStreamName)
	err, errs := incfg.PutRecords(records)
	if err != nil {
		t.Fatalf("[E2EDataStream] error sending data to stream: %v. %+v", err, errs)
	}

	t.Logf("[E2EDataStream] Waiting for profile creation")
	t.Logf("%+v", clickstreamRecord)
	profile_id := ((clickstreamRecord.Data["session"].(map[string]interface{}))["id"]).(string)
	_, err = uptSdk.WaitForProfile(domainName, profile_id, 60)
	if err != nil {
		t.Fatalf("Could not find profile with ID %s for clickstream rec. Error: %v", profile_id, err)
	}
	hotel_profile_id := ((hotelBookingRecord.Data["holder"].(map[string]interface{}))["id"]).(string)
	_, err = uptSdk.WaitForProfile(domainName, hotel_profile_id, 60)
	if err != nil {
		t.Fatalf("Could not find profile with ID %s for hotel booking record. Error: %v", hotel_profile_id, err)
	}
	guest_profile_id := (guestProfileRecord.Data["id"]).(string)
	_, err = uptSdk.WaitForProfile(domainName, guest_profile_id, 60)
	if err != nil {
		t.Fatalf("Could not find profile with ID %s for guest booking record. Error: %v", guest_profile_id, err)
	}
	t.Logf("[E2EDataStream] All full profiles have been created")

	/////////////////////////////
	//Partial Record Tests
	/////////////////////////////
	guestProfileString := buildGuestProfileRecordWithLoyalty(guestId, middlename, profileToMergeEmail1, objectId, oldPoints, oldPointsToNextLevel, oldLevel)
	partialGuestProfileObject := buildRecordFromString("guest_profile", guestProfileString, domainName, "", "")
	partialGuestProfileObjectSerialized, _ := json.Marshal(partialGuestProfileObject)
	err, errs = incfg.PutRecords([]kinesis.Record{
		{
			Pk:   partialGuestProfileObject.ObjectType + "-" + partialGuestProfileObject.ModelVersion + "-1",
			Data: string(partialGuestProfileObjectSerialized),
		},
	})
	if err != nil {
		t.Fatalf("[E2EDataStream] error sending partial record to stream: %v. %+v", err, errs)
	}
	trav, err := uptSdk.WaitForProfile(domainName, guestId, 60)
	if err != nil {
		t.Fatalf("Could not find profile with ID %s for partial record. Error: %v", guestId, err)
	}
	profObject, err := uptSdk.WaitForProfileObject(domainName, objectId, trav.ConnectID, partialObjectType, 60)
	if err != nil {
		t.Fatalf("Could not get profile object %v", err)
	}
	oldPointsInt, _ := strconv.ParseFloat(oldPoints, 64)
	oldPointsToNextLevelInt, _ := strconv.ParseFloat(oldPointsToNextLevel, 64)
	err = runProfileObjectLoyaltyTests(profObject, guestId, objectId, oldLevel, oldPointsInt, oldPointsToNextLevelInt)
	if err != nil {
		t.Fatalf("Error running profile object loyalty tests: %v", err)
	}

	newGuestProfileString := buildGuestProfileRecordWithLoyalty(guestId, middlename, profileToMergeEmail1, objectId, oldPoints, newPointsToNextLevel, newLevel)
	newPartialGuestProfileObject := buildPartialRecordFromString("guest_profile", newGuestProfileString, domainName, "partial", []string{"hotel_loyalty.level"})
	newPartialGuestProfileObjectSerialized, _ := json.Marshal(newPartialGuestProfileObject)
	t.Logf("Sending partial record to stream: %+v", string(newPartialGuestProfileObjectSerialized))
	err, errs = incfg.PutRecords([]kinesis.Record{
		{
			Pk:   newPartialGuestProfileObject.ObjectType + "-" + newPartialGuestProfileObject.ModelVersion + "-1",
			Data: string(newPartialGuestProfileObjectSerialized),
		},
	})
	if err != nil || len(errs) > 0 {
		t.Fatalf("[E2EDataStream] error sending partial record to stream: %v. %+v", err, errs)
	}
	_, err = uptSdk.WaitForProfileObject(domainName, objectId, trav.ConnectID, partialObjectType, 60)
	if err != nil {
		t.Fatalf("Could not get profile object %v", err)
	}
	err = repeatRunTestsUntilUpdateOrTimeout(t, uptSdk, guestId, objectId, "hotel_loyalty", 60, newLevel, oldPointsInt, oldPointsToNextLevelInt, domainName)
	if err != nil {
		t.Fatalf("Error running profile object loyalty tests, first partial test: %v", err)
	}

	newPointsToNextLevelInt, _ := strconv.ParseFloat(newPointsToNextLevel, 64)
	newGuestProfileString = buildGuestProfileRecordWithLoyalty(guestId, middlename, profileToMergeEmail1, objectId, oldPoints, newPointsToNextLevel, newLevel)
	newPartialGuestProfileObject = buildPartialRecordFromString("guest_profile", newGuestProfileString, domainName, "partial", []string{"hotel_loyalty.*"})
	newPartialGuestProfileObjectSerialized, _ = json.Marshal(newPartialGuestProfileObject)
	err, errs = incfg.PutRecords([]kinesis.Record{
		{
			Pk:   newPartialGuestProfileObject.ObjectType + "-" + newPartialGuestProfileObject.ModelVersion + "-1",
			Data: string(newPartialGuestProfileObjectSerialized),
		},
	})
	if err != nil {
		t.Fatalf("[E2EDataStream] error sending partial record to stream: %v. %+v", err, errs)
	}
	_, err = uptSdk.WaitForProfileObject(domainName, objectId, trav.ConnectID, partialObjectType, 60)
	if err != nil {
		t.Fatalf("Could not get profile object %v", err)
	}
	err = repeatRunTestsUntilUpdateOrTimeout(t, uptSdk, guestId, objectId, "hotel_loyalty", 60, newLevel, oldPointsInt, newPointsToNextLevelInt, domainName)
	if err != nil {
		t.Fatalf("Error running profile object loyalty tests, second partial test: %v", err)
	}

	guestProfileString = buildGuestProfileRecordWithLoyalty(guestId, middlename, profileToMergeEmail1, secondObjectId, oldPoints, oldPointsToNextLevel, oldLevel)
	partialGuestProfileObject = buildPartialRecordFromString("guest_profile", guestProfileString, domainName, "partial", []string{"hotel_loyalty.*"})
	partialGuestProfileObjectSerialized, _ = json.Marshal(partialGuestProfileObject)
	err, errs = incfg.PutRecords([]kinesis.Record{
		{
			Pk:   partialGuestProfileObject.ObjectType + "-" + partialGuestProfileObject.ModelVersion + "-1",
			Data: string(partialGuestProfileObjectSerialized),
		},
	})
	if err != nil {
		t.Fatalf("[E2EDataStream] error sending partial record to stream: %v. %+v", err, errs)
	}
	//in partial mode, if partial mode validationfails, because no object exists, we create a new object,
	profObject, err = uptSdk.WaitForProfileObject(domainName, secondObjectId, trav.ConnectID, partialObjectType, 60)
	log.Printf("Got profile object %+v", profObject)
	if err != nil {
		t.Fatalf("Profile object should have been created.")
	}

	/////////////////////////////
	//Merge Record Tests
	/////////////////////////////
	records = []kinesis.Record{
		{
			Pk:   guestProfileRecord.ObjectType + "-" + guestProfileRecord.ModelVersion + "-4",
			Data: string(guestProfileMsgForMergeMainSerialized),
		},
		{
			Pk:   guestProfileRecord.ObjectType + "-" + guestProfileRecord.ModelVersion + "-5",
			Data: string(guestProfileMsgForMergeDupeSerialized),
		},
	}
	t.Logf("[E2EDataStream] 2- sending records %+v to kinesis stream %v (merge test)", records, infraCfg.RealTimeIngestorStreamName)
	err, errs = incfg.PutRecords(records)
	if err != nil {
		t.Fatalf("[E2EDataStream] error sending data to stream: %v. %+v", err, errs)
	}

	t.Logf("[E2EDataStream] Checking that both profiles exist")
	_, err = uptSdk.WaitForProfile(domainName, profileToMergeId1, 90)
	if err != nil {
		t.Fatalf("Could not find profile with ID %s for guest booking record. Error: %v", profileToMergeId1, err)
	}
	_, err = uptSdk.WaitForProfile(domainName, profileToMergeId2, 60)
	if err != nil {
		t.Fatalf("Could not find profile with ID %s for guest booking record. Error: %v", profileToMergeId2, err)
	}
	resp, _ := uptSdk.SearchProfile(domainName, "travellerId", []string{profileToMergeId2})
	profiles := utils.SafeDereference(resp.Profiles)
	if len(profiles) > 0 {
		t.Logf("Profile to merge: %+v ", profiles[0])
	}
	cid1 := profiles[0].ConnectID
	resp, _ = uptSdk.SearchProfile(domainName, "travellerId", []string{profileToMergeId1})
	profiles = utils.SafeDereference(resp.Profiles)
	if len(profiles) > 0 {
		t.Logf("Profile being merged into: %+v ", profiles[0])
	}
	cid2 := profiles[0].ConnectID

	t.Logf("[E2EDataStream] Sending merge request")
	err, errs = incfg.PutRecords([]kinesis.Record{
		{
			Pk:   guestProfileMergeRq.ObjectType + "-" + guestProfileMergeRq.ModelVersion + "-1",
			Data: string(guestProfileMergeRqSerialized),
		},
	})
	if err != nil {
		t.Fatalf("[E2EDataStream] error sending merge request to stream: %v. %+v", err, errs)
	}
	t.Logf("[E2EDataStream] Waiting for merge to complete")
	err = waitForProfileMerge(t, uptSdk, cid1, cid2, 60, domainName)
	if err != nil {
		t.Fatalf("Error waiting for merge to complete. Error: %v", err)
	}
	t.Logf("[E2EDataStream] Waiting for profile 1 to be searchable again")
	trav1, err := uptSdk.WaitForProfile(domainName, profileToMergeId1, 90)
	if err != nil {
		t.Fatalf("Could not find profile with ID %s after merge. Error: %v", profileToMergeId1, err)
	}
	t.Logf("[E2EDataStream] Waiting for profile 2 to be searchable again")
	trav2, err := uptSdk.WaitForProfile(domainName, profileToMergeId2, 60)
	if err != nil {
		t.Fatalf("Could not find profile with ID %s after merge. Error: %v", profileToMergeId2, err)
	}
	if trav1.ConnectID != trav2.ConnectID {
		t.Fatalf("Search on both merged profiles traveller_id should return the same connect id")
	}
	resp, err = uptSdk.RetreiveProfile(domainName, trav1.ConnectID, []string{"email_history", "phone_history"})
	if err != nil {
		t.Fatalf("Error getting full surviving profile %v", err)
	}
	profile := (utils.SafeDereference(resp.Profiles))[0]

	t.Logf("full surviving profile: %+v", profile)
	hasObjectFromMergedPRofile := false
	hasObjectFromSurvivingProfile := false
	for _, obj := range profile.EmailHistoryRecords {
		if obj.Address == profileToMergeEmail1 {
			hasObjectFromSurvivingProfile = true
		}
		if obj.Address == profileToMergeEmail2 {
			hasObjectFromMergedPRofile = true
		}
	}
	if !hasObjectFromMergedPRofile || !hasObjectFromSurvivingProfile {
		t.Fatalf("Surviving profile should contain both %v and %v", profileToMergeEmail1, profileToMergeEmail2)
	}

}

func TestTransformerWithRulesBasedMergingAndCache(t *testing.T) {
	// Set up clients
	envCfg, infraCfg, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}
	incfg := InitKinesis(infraCfg.RealTimeIngestorStreamName, envCfg.Region, "", "")
	domainName := "rule_based_cache" + strings.ToLower(core.GenerateUniqueId())

	uptSdk, err := upt_sdk.Init(infraCfg)
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
	_, err = uptSdk.SaveIdResRuleSet(domainName, []customerprofileslcs.Rule{
		{
			Index:       0,
			Name:        "sameMiddleName",
			Description: "Merge profiles with matching middle name",
			Conditions: []customerprofileslcs.Condition{
				{
					Index:               0,
					ConditionType:       customerprofileslcs.CONDITION_TYPE_MATCH,
					IncomingObjectType:  customerprofileslcs.PROFILE_OBJECT_TYPE_NAME,
					IncomingObjectField: "MiddleName",
					ExistingObjectType:  customerprofileslcs.PROFILE_OBJECT_TYPE_NAME,
					ExistingObjectField: "MiddleName",
					IndexNormalization: customerprofileslcs.IndexNormalizationSettings{
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

	_, err = uptSdk.ActivateIdResRuleSetWithRetries(domainName, 5)
	if err != nil {
		t.Fatalf("Error activating rule set: %v", err)
	}

	_, err = uptSdk.SaveCacheRuleSet(domainName, []customerprofileslcs.Rule{
		{
			Index:       0,
			Name:        "skipMiddleName",
			Description: "Skipping profiles with a particular middle name",
			Conditions: []customerprofileslcs.Condition{
				{
					Index:               0,
					ConditionType:       customerprofileslcs.CONDITION_TYPE_SKIP,
					IncomingObjectType:  customerprofileslcs.PROFILE_OBJECT_TYPE_NAME,
					IncomingObjectField: "MiddleName",
					Op:                  customerprofileslcs.RULE_OP_EQUALS_VALUE,
					SkipConditionValue:  customerprofileslcs.StringVal(middlename),
					IndexNormalization: customerprofileslcs.IndexNormalizationSettings{
						Lowercase: true,
						Trim:      true,
					},
				},
			},
		},
		{
			Index:       1,
			Name:        "skipEmptyMiddleName",
			Description: "Skipping profiles with an empty middle name",
			Conditions: []customerprofileslcs.Condition{
				{
					Index:               0,
					ConditionType:       customerprofileslcs.CONDITION_TYPE_SKIP,
					IncomingObjectType:  customerprofileslcs.PROFILE_OBJECT_TYPE_NAME,
					IncomingObjectField: "MiddleName",
					Op:                  customerprofileslcs.RULE_OP_EQUALS_VALUE,
					SkipConditionValue:  customerprofileslcs.StringVal(""),
					IndexNormalization: customerprofileslcs.IndexNormalizationSettings{
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

	_, err = uptSdk.ActivateCacheRuleSet(domainName)
	if err != nil {
		t.Fatalf("Error activating rule set: %v", err)
	}

	baseTime := time.Now()
	guestProfileToMerge := buildGuestProfileRecord(profileToMergeId1, middlename, profileToMergeEmail1, baseTime)
	guestProfileToMergeRecord := buildRecordFromString("guest_profile", guestProfileToMerge, domainName, "", "")
	guestProfileToMergeSerialized, _ := json.Marshal(guestProfileToMergeRecord)
	finalGuestProfile := buildGuestProfileRecord(profileToMergeId2, middlename, profileToMergeEmail2, baseTime.Add(time.Hour))
	finalGuestProfileRecord := buildRecordFromString("guest_profile", finalGuestProfile, domainName, "", "")
	finalGuestProfileSerialized, _ := json.Marshal(finalGuestProfileRecord)
	guestProfileCached := buildGuestProfileRecord(profileToMergeId3, "cache middle name", profileToMergeEmail1, baseTime)
	guestProfileCachedRecord := buildRecordFromString("guest_profile", guestProfileCached, domainName, "", "")
	guestProfileCachedSerialized, _ := json.Marshal(guestProfileCachedRecord)

	records := []kinesis.Record{
		{
			Pk:   "guest_profile-1-1",
			Data: string(guestProfileToMergeSerialized),
		},
		{
			Pk:   "guest_profile-1-3",
			Data: string(guestProfileCachedSerialized),
		},
	}
	t.Logf("[E2EDataStream] 2- sending records %+v to kinesis stream %v (rules based merging test)", records, infraCfg.RealTimeIngestorStreamName)
	err, errs := incfg.PutRecords(records)
	if err != nil {
		t.Fatalf("[E2EDataStream] error sending data to stream: %v. %+v", err, errs)
	}
	_, err = uptSdk.WaitForProfile(domainName, profileToMergeId1, 300)
	if err != nil {
		t.Fatalf("Error waiting for profile %v", err)
	}
	trav3, err := uptSdk.WaitForProfile(domainName, profileToMergeId3, 300)
	if err != nil {
		t.Fatalf("Error waiting for profile %v", err)
	}
	records = []kinesis.Record{
		{
			Pk:   "guest_profile-1-2",
			Data: string(finalGuestProfileSerialized),
		},
	}
	t.Logf("[E2EDataStream] 2- sending records %+v to kinesis stream %v (rules based merging test)", records, infraCfg.RealTimeIngestorStreamName)
	err, errs = incfg.PutRecords(records)
	if err != nil {
		t.Fatalf("[E2EDataStream] error sending data to stream: %v. %+v", err, errs)
	}
	err = util.WaitForExpectedCondition(func() bool {
		//first record
		res, err := uptSdk.SearchProfile(domainName, "travellerId", []string{profileToMergeId1})
		if err != nil {
			t.Fatalf("Error searching profile %v", err)
		}
		profiles := utils.SafeDereference(res.Profiles)
		if len(profiles) != 1 {
			t.Logf("Expected 1 profile, got %v", len(profiles))
			return false
		}
		//should be second email, not first, because it got merged
		if profiles[0].PersonalEmailAddress != profileToMergeEmail2 {
			t.Logf("Expected profile to have email %v, got %v", profileToMergeEmail2, profiles[0].PersonalEmailAddress)
			return false
		}
		return true
	}, 24, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for profile %v", err)
	}
	trav2, err := uptSdk.WaitForProfile(domainName, profileToMergeId2, 300)
	if err != nil {
		t.Fatalf("Error waiting for profile %v", err)
	}
	if trav2.PersonalEmailAddress != profileToMergeEmail2 {
		t.Fatalf("Expected profile to have email %v, got %v", profileToMergeEmail2, trav2.PersonalEmailAddress)
	}

	if infraCfg.CustomerProfileStorageMode == "true" {
		cpCfg := customerprofiles.InitWithDomain(domainName, envCfg.Region, "", "")
		domExists, err := cpCfg.DomainExists(domainName)
		if err != nil {
			t.Fatalf("Error getting domain: %v", err)
		}
		if !domExists {
			t.Fatalf("Domain %v does not exist", domainName)
		}
		err = util.WaitForExpectedCondition(func() bool {
			//first record
			cpId, err := cpCfg.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, trav2.ConnectID)
			if err != nil {
				t.Fatalf("Error getting profile id: %v", err)
			}
			if cpId != "" {
				t.Fatalf("[TestSkipCPIntegration] profile should not exist in cache")
				return false
			}
			cpId, err = cpCfg.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, trav3.ConnectID)
			if err != nil {
				t.Logf("[TestSkipCPIntegration] error, failed to get profile id for %v", trav3.ConnectID)
				return false
			}
			if cpId == "" {
				t.Logf("[TestSkipCPIntegration] profile exist in cache")
				return false
			}
			return true
		}, 24, 5*time.Second)
		if err != nil {
			t.Fatalf("Error waiting for expected condition %v", err)
		}
	}
}

func waitForProfileMerge(t *testing.T, uptSdk *upt_sdk.UptHandler, connectId1, connectId2 string, timeout int, domainName string) error {
	t.Logf("Waiting %v seconds for profiles %s and %s to be merged", timeout, connectId1, connectId2)
	it := 0
	for it*5 < timeout {
		res1, err := uptSdk.RetreiveProfile(domainName, connectId1, []string{})
		if err != nil {
			log.Printf("error retrieving profile %s %v", connectId1, err)
			return errors.New("error retrieving profile")
		}
		res2, err := uptSdk.RetreiveProfile(domainName, connectId2, []string{})
		if err != nil {
			log.Printf("error retrieving profile %s %v", connectId2, err)
			return errors.New("error retrieving profile")
		}
		if res1.Profiles != nil && res2.Profiles != nil {
			p1 := *res1.Profiles
			p2 := *res2.Profiles
			if p1[0].ConnectID == p2[0].ConnectID {
				log.Printf("Profile %s and %s have been merged. returning", connectId1, connectId2)
				return nil
			}
		}
		t.Logf("profiles have not been merged: Waiting 5 seconds")
		time.Sleep(5 * time.Second)
		it += 1
	}
	return fmt.Errorf("profile has not been deleted in %v seconds", timeout)
}

func buildRecord(objectType string, folder string, domain string, mode string, line int) BusinessObjectRecord {
	newMode := ""
	if mode != "" {
		newMode = "_" + mode
	}
	content, err := os.ReadFile("../../test_data/" + folder + "/data1" + newMode + ".jsonl")
	if err != nil {
		log.Printf("Error reading JSON file: %s", err)
		return BusinessObjectRecord{}
	}
	content = []byte(strings.Split(string(content), "\n")[line])

	var v map[string]interface{}
	json.Unmarshal(content, &v)

	return BusinessObjectRecord{
		Domain:       domain,
		ObjectType:   objectType,
		ModelVersion: "1",
		Data:         v,
		Mode:         mode,
	}
}

func buildRecordFromString(objectType string, objectString string, domain string, mode string, mergeProfileId string) BusinessObjectRecord {
	content := []byte(objectString)
	var v map[string]interface{}
	json.Unmarshal(content, &v)

	return BusinessObjectRecord{
		Domain:             domain,
		ObjectType:         objectType,
		ModelVersion:       "1",
		Data:               v,
		Mode:               mode,
		MergeModeProfileID: mergeProfileId,
	}
}

func buildPartialRecordFromString(objectType string, objectString string, domain string, mode string, partialModeOptions []string) BusinessObjectRecord {
	content := []byte(objectString)
	var v map[string]interface{}
	json.Unmarshal(content, &v)

	return BusinessObjectRecord{
		Domain:       domain,
		ObjectType:   objectType,
		ModelVersion: "1",
		Data:         v,
		Mode:         mode,
		PartialModeOptions: PartialModeOptions{
			Fields: partialModeOptions,
		},
	}
}

func runProfileObjectLoyaltyTests(profObject profilemodel.ProfileObject, id string, objectId string, level string, points float64, pointsNextLevel float64) error {
	attr := profObject.AttributesInterface

	if traveller_id, ok := attr["TravellerID"]; !ok || traveller_id != id {
		return errors.New("traveller_id is not correct")
	}

	if objID, ok := attr["AccpObjectID"]; !ok || objID != objectId {
		return errors.New("accp_object_id is not correct")
	}

	if lvl, ok := attr["Level"]; !ok || lvl != level {
		return errors.New("level is not correct")
	}

	pointsNextLevelObj, ok := attr["PointsToNextLevel"]
	if !ok {
		return errors.New("points_to_next_level not found")
	}
	pointsNextLevelObjFloat, err := strconv.ParseFloat(pointsNextLevelObj.(string), 64)
	if err != nil {
		return errors.New("points_to_next_level can not be converted to int")
	}
	if pointsNextLevelObjFloat != pointsNextLevel {
		return errors.New("points_to_next_level is not correct")
	}
	pointsObj, ok := attr["Points"]
	if !ok {
		return errors.New("points not found")
	}
	pointsObjInt, err := strconv.ParseFloat(pointsObj.(string), 64)
	if err != nil {
		return errors.New("points can not be converted to int")
	}
	if pointsObjInt != points {
		return errors.New("points is not correct")
	}
	return nil
}

func repeatRunTestsUntilUpdateOrTimeout(t *testing.T, uptSdk *upt_sdk.UptHandler, profileId string, objectId string, objectType string, timeout int, level string, points float64, pointsNextLevel float64, domainName string) error {
	it := 0
	trav, err := uptSdk.WaitForProfile(domainName, profileId, 1)
	if err != nil {
		return err
	}
	for it*5 < timeout {
		profObject, err := uptSdk.GetSpecificProfileObject(domainName, objectId, trav.ConnectID, objectType)
		if err != nil {
			return err
		}
		err = runProfileObjectLoyaltyTests(profObject, profileId, objectId, level, points, pointsNextLevel)
		if err != nil {
			t.Logf("Error running profile object loyalty tests:  %v", err)
			t.Logf("Current profile object: %+v", profObject.Attributes)
			t.Logf("Waiting 5 seconds")
			time.Sleep(5 * time.Second)
			it += 1
			continue
		} else {
			return nil
		}
	}
	return errors.New("could not find profile object that works: timeout expired")
}

func buildGuestProfileRecord(id string, middleName string, email string, lastUpdatedOn time.Time) string {
	gp := generators.BuildGuest(id)
	gp.LastUpdatedOn = lastUpdatedOn
	gp.MiddleName = middleName
	gp.ID = id
	gp.Emails = []common.Email{
		{
			Type:    "",
			Address: email,
			Primary: true,
		},
	}
	businessByte, err := json.Marshal(gp)
	if err != nil {
		log.Fatalf("Error marshalling guest profile record: %v", err)
	}
	return string(businessByte)
}

func buildGuestProfileRecordWithLoyalty(id string, middleName string, email string, loyaltyId string, points string, pointsToNextLevel string, level string) string {
	gp := generators.BuildGuest(id)
	gp.MiddleName = middleName
	gp.ID = id
	gp.Emails = []common.Email{
		{
			Type:    "",
			Address: email,
			Primary: true,
		},
	}
	pointFloat, err := strconv.ParseFloat(points, 64)
	if err != nil {
		log.Fatalf("Error converting string to float: %v", err)
	}
	pointToNextLevelFloat, err := strconv.ParseFloat(pointsToNextLevel, 64)
	if err != nil {
		log.Fatalf("Error converting string to float: %v", err)
	}
	gp.LoyaltyPrograms = []lodging.LoyaltyProgram{
		{
			ID:                loyaltyId,
			Points:            corelodge.Float(pointFloat),
			PointsToNextLevel: corelodge.Float(pointToNextLevelFloat),
			Level:             level,
			PointUnit:         "miles",
		},
	}
	businessByte, err := json.Marshal(gp)
	if err != nil {
		log.Fatalf("Error marshalling guest profile record: %v", err)
	}
	return string(businessByte)
}

func buildGuestProfileRecordWithLoyaltyTwoEmails(id string, email string, email2 string, loyaltyId string, programName string, points string, unit string, pointsToNextLevel string, level string, loyaltyJoined string) string {
	gp := generators.BuildGuest(id)
	gp.ID = id
	gp.Emails = []common.Email{
		{
			Type:    "",
			Address: email,
			Primary: true,
		},
		{
			Type:    "",
			Address: email2,
			Primary: true,
		},
	}
	pointFloat, err := strconv.ParseFloat(points, 64)
	if err != nil {
		log.Fatalf("Error converting string to float: %v", err)
	}
	pointToNextLevelFloat, err := strconv.ParseFloat(pointsToNextLevel, 64)
	if err != nil {
		log.Fatalf("Error converting string to float: %v", err)
	}
	const layout = "2006-01-02T15:04:05.999999Z"
	joined, err := time.Parse(layout, loyaltyJoined)
	if err != nil {
		log.Fatalf("Error converting string to time: %v", err)
	}
	gp.LoyaltyPrograms = []lodging.LoyaltyProgram{
		{
			ID:                loyaltyId,
			ProgramName:       programName,
			Points:            corelodge.Float(pointFloat),
			PointsToNextLevel: corelodge.Float(pointToNextLevelFloat),
			Level:             level,
			PointUnit:         unit,
			Joined:            joined,
		},
	}
	businessByte, err := json.Marshal(gp)
	if err != nil {
		log.Fatalf("Error marshalling guest profile record: %v", err)
	}
	return string(businessByte)
}
