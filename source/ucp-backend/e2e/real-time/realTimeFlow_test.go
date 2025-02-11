// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/sqs"
	backendmodel "tah/upt/source/ucp-backend/src/business-logic/model/common"
	constants "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/traveler"
	travModel "tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/config"
	utils "tah/upt/source/ucp-common/src/utils/utils"
	upt_sdk "tah/upt/source/ucp-sdk/src"
)

// End to end test for real time flow
//
// The test leverages Newman, the open-source Postman CLI tool that allows us
// to run collections locally, without a Postman account.
// https://learning.postman.com/docs/postman-cli/postman-cli-overview/
//
// Using Newman alongside Go gives us the ability to easily use the Postman
// Collection Runner (running collection requests in a specified order) to
// test the API, while also managing/validating the data with Go.
// https://learning.postman.com/docs/collections/running-collections/intro-to-collection-runs/

// ACCP Events
type ProfileChangeEvent struct {
	Version    string                   `json:"version"`
	ID         string                   `json:"id"`
	DetailType string                   `json:"detail-type"`
	Source     string                   `json:"source"`
	Detail     ProfileChangeEventDetail `json:"detail"`
}
type ProfileChangeEventDetail struct {
	TravellerID string        `json:"travelerID"`
	Domain      string        `json:"domain"`
	Events      []DetailEvent `json:"events"`
}
type DetailEvent struct {
	EventType  string          `json:"eventType"`
	ObjectType string          `json:"objectType"`
	Data       DetailEventData `json:"data"`
}
type DetailEventData struct {
	ModelVersion         string `json:"modelVersion"`
	PersonalEmailAddress string `json:"personalEmailAddress"`
	// ...additional fields available
}

func TestRealTime(t *testing.T) {
	_, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}
	services, err := InitServices()
	if err != nil {
		t.Fatalf("Error initializing services: %v", err)
	}
	domainName := "e2e_" + strings.ToLower(core.GenerateUniqueId())

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

	// Create SQS queue to validate EventBridge events
	queueUrl, err := services.SQS.Create(domainName)
	if err != nil {
		t.Errorf("[TestRealTime] Error creating SQS queue: %s", err)
	}
	eventPattern := fmt.Sprintf("{\"source\": [\"ucp-%s\"]}", domainName) // rule fires on events for this domain only
	err = services.EventBridge.CreateSQSTarget(domainName, infraConfig.EventBridgeBusName, queueUrl, eventPattern)
	if err != nil {
		t.Errorf("[TestRealTime] Error creating SQS target: %s", err)
	}
	t.Cleanup(func() {
		err = services.SQS.Delete()
		if err != nil {
			t.Errorf("[TestRealTime] Error deleting SQS queue: %s", err)
		}
		err = services.EventBridge.DeleteRuleByName(domainName)
		if err != nil {
			t.Errorf("[TestRealTime] Error deleting event bridge rule")
		}
	})

	// ADD PROFILES TO DOMAIN
	// Define two customer profiles to create
	airBookingEmail1 := "air_traveller_1@example.com"
	tid1 := "0982031183"
	airBookingEmail2 := "air_traveller_2@example.com"
	tid2 := "5410295782"
	airBooking1 := generateKinesisRec("air_booking.jsonl", constants.BIZ_OBJECT_AIR_BOOKING, domainName, 0)
	airBooking2 := generateKinesisRec("air_booking.jsonl", constants.BIZ_OBJECT_AIR_BOOKING, domainName, 1)
	records := []kinesis.Record{airBooking1, airBooking2}
	// Send profiles with objects to new domain, poll in Go for profiles
	err, ingestionErrs := services.IngestorStream.PutRecords(records)
	if err != nil {
		t.Fatalf("[TestRealTime] Error sending profiles to domain: %s", err)
	}
	if len(ingestionErrs) > 0 {
		t.Fatalf("[TestRealTime] Errors sending profiles to domain: %s", ingestionErrs)
	}

	t1, err := uptSdk.WaitForProfileWithCondition(domainName, tid1, 60, func(trav model.Traveller) bool {
		return len(trav.AirBookingRecords) == 6 && len(trav.EmailHistoryRecords) == 1 && len(trav.PhoneHistoryRecords) == 1 && len(trav.AirLoyaltyRecords) == 1
	})
	if err != nil {
		t.Fatalf("[TestRealTime] Error searching profiles: %s", err)
	}
	log.Printf("[TestRealTime] successfully create profile %v with profile id: %s", t1.ConnectID, tid1)
	t2, err := uptSdk.WaitForProfileWithCondition(domainName, tid2, 60, func(trav model.Traveller) bool {
		return len(trav.AirBookingRecords) == 2 && len(trav.EmailHistoryRecords) == 1 && len(trav.PhoneHistoryRecords) == 2 && len(trav.AirLoyaltyRecords) == 2
	})
	if err != nil {
		t.Fatalf("[TestRealTime] Error searching profiles: %s", err)
	}
	log.Printf("[TestRealTime] successfully create profile %v with profile id %s", t2.ConnectID, tid2)

	if len(t1.AirBookingRecords) != 6 {
		t.Fatalf("[TestRealTime] Profile 1 should have %d AirBookingRecords and not %d", 6, len(t1.AirBookingRecords))
	}
	if len(t1.EmailHistoryRecords) != 1 {
		t.Fatalf("[TestRealTime] Profile 1 should have %d EmailHistoryRecords and not %d", 2, len(t1.EmailHistoryRecords))
	}
	if len(t1.PhoneHistoryRecords) != 1 {
		t.Fatalf("[TestRealTime] Profile 1 should have %d PhoneHistoryRecords and not %d", 1, len(t1.PhoneHistoryRecords))
	}
	if len(t1.AirLoyaltyRecords) != 1 {
		t.Fatalf("[TestRealTime] Profile 1 should have %d AirLoyaltyRecords and not %d", 1, len(t1.AirLoyaltyRecords))
	}
	if len(t2.AirBookingRecords) != 2 {
		t.Fatalf("[TestRealTime] Profile 2 should have %d AirBookingRecords and not %d", 2, len(t2.AirBookingRecords))
	}
	if len(t2.EmailHistoryRecords) != 1 {
		t.Fatalf("[TestRealTime] Profile 2 should have %d EmailHistoryRecords and not %d", 2, len(t2.EmailHistoryRecords))
	}
	if len(t2.PhoneHistoryRecords) != 2 {
		t.Fatalf("[TestRealTime] Profile 2 should have %d PhoneHistoryRecords and not %d", 2, len(t2.PhoneHistoryRecords))
	}
	if len(t2.AirLoyaltyRecords) != 2 {
		t.Fatalf("[TestRealTime] Profile 2 should have %d AirLoyaltyRecords and not %d", 2, len(t2.AirLoyaltyRecords))
	}

	res, err := uptSdk.SearchProfile(domainName, "email", []string{airBookingEmail1})
	if err != nil {
		t.Fatalf("[TestRealTime] Error searching for profile %s by email: %v", airBookingEmail1, err)
	}
	profiles := utils.SafeDereference(res.Profiles)
	if len(profiles) != 1 {
		t.Fatalf("[TestRealTime] profile search by email should one profile and not: %v", profiles)
	}
	if profiles[0].ConnectID != t1.ConnectID {
		t.Fatalf("[TestRealTime] profile search by email should return connectID %s and not %s", t1.ConnectID, profiles[0].ConnectID)
	}

	res, err = uptSdk.SearchProfile(domainName, "travellerId", []string{t1.TravellerID})
	if err != nil {
		t.Fatalf("[TestRealTime] Error searching for profile %s by traveler id: %v", t1.TravellerID, err)
	}
	profiles = utils.SafeDereference(res.Profiles)
	if len(profiles) != 1 {
		t.Fatalf("[TestRealTime] profile search should one profile and not: %v", profiles)
	}
	if profiles[0].ConnectID != t1.ConnectID {
		t.Fatalf("[TestRealTime] profile search by traveler ID should return connectID %s and not %s", t1.ConnectID, profiles[0].ConnectID)
	}
	res, _ = uptSdk.SearchProfileAdvanced(domainName, map[string]string{"travellerId": "definitelydoesnotexist", "email": airBookingEmail1})
	profiles = utils.SafeDereference(res.Profiles)
	if len(profiles) != 0 {
		t.Fatalf("[TestRealTime] advanced profile search with invalid traveler id should return no result and not : %v", profiles)
	}

	t1AfterMerge, err := uptSdk.MergeProfilesAndWait(domainName, t1.ConnectID, t2.ConnectID, 120)
	if err != nil {
		t.Fatalf("[TestRealTime] error waiting for profiles to be merged: %v", err)
	}
	log.Printf("[TestRealTime] Remaining profile content: %+v", t1AfterMerge)
	if sumRecords(t1AfterMerge) != 16 {
		t.Fatalf("[TestRealTime] Merged profile has not been merged correctly. profile should have 16 objects and has %v", sumRecords(t1AfterMerge))
	}
	hasMergedEmail := false
	for _, v := range t1AfterMerge.EmailHistoryRecords {
		if v.Address == airBookingEmail2 {
			hasMergedEmail = true
			break
		}
	}
	if !hasMergedEmail {
		t.Fatalf("[TestRealTime] Merged profile has not been merged correctly. Profile is missing email %v in history", airBookingEmail2)
	}
	if len(t1AfterMerge.AirBookingRecords) != 8 {
		t.Fatalf("[TestRealTime] Merged profile has not been merged correctly: profile should have %v AirBookingRecords and has %v", 8, len(t1AfterMerge.AirBookingRecords))
	}
	if len(t1AfterMerge.EmailHistoryRecords) != 2 {
		t.Fatalf("[TestRealTime] Merged profile has not been merged correctly: profile should have %v EmailHistoryRecords and has %v", 4, len(t1AfterMerge.EmailHistoryRecords))
	}
	if len(t1AfterMerge.PhoneHistoryRecords) != 3 {
		t.Fatalf("[TestRealTime] Merged profile has not been merged correctly: profile should have %v AirBookingRecords and has %v", 3, len(t1AfterMerge.PhoneHistoryRecords))
	}
	if len(t1AfterMerge.AirLoyaltyRecords) != 3 {
		t.Fatalf("[TestRealTime] Merged profile has not been merged correctly: profile should have %v AirBookingRecords and has %v", 3, len(t1AfterMerge.AirLoyaltyRecords))
	}

	// VALIDATE EVENT EXPORT
	// Validating EventBridge export
	isProfile1Exported := false // // Validate profile1 was updated (CREATED event might be too far back in queue and not fetched)
	log.Printf("[TestRealTime][event-export-test] Getting events from SQS")
	max := 6
	for it := 0; it < max; it++ {
		log.Printf("[TestRealTime][event-export-test] Looking for UPDATED event. try %d of %d", it, max)
		ebEvents, err := getEventsFromSqs(services.SQS, 60)
		if err != nil {
			log.Printf("[TestRealTime][event-export-test]  Non blocking Error getting events from SQS: %s", err)
		} else {
			log.Printf("[TestRealTime][event-export-test]  Successfully go Events: %+v", ebEvents)
		}
		log.Printf("[TestRealTime] Looking for UPDATED event with email %v", airBookingEmail1)
		for _, e := range ebEvents {
			for _, event := range e.Detail.Events {
				if event.EventType == "UPDATED" && event.ObjectType == "_profile" && event.Data.PersonalEmailAddress == airBookingEmail1 {
					log.Printf("[TestRealTime] Found UPDATED event with email %v", airBookingEmail1)
					isProfile1Exported = true
					break
				}
			}
			log.Printf("[TestRealTime][event-export-test]  Did not find UPDATED event with email %v. looking for next SQS batch", airBookingEmail1)
		}
		if !isProfile1Exported {
			log.Printf("[TestRealTime][event-export-test]  Waiting 5 seconds and retrying")
			time.Sleep(5 * time.Second)
		} else {
			log.Printf("Found UPDATED event. Test successful!")
			break
		}
	}
	if !isProfile1Exported {
		t.Fatalf("[TestRealTime] Profile 1 was not properly exported to EventBridge (could not find UPDATE event with email %v)", airBookingEmail1)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	//Validating S3 Backup
	go func() {
		if infraConfig.RealTimeBackupEnabled == "true" {
			log.Printf("[TestRealTime] Validating S3 backup")
			err = services.BackupBucket.WaitForObject("data/domainname="+domainName, 300)
			if err != nil {
				t.Errorf("[TestRealTime] Error waiting for S3 export: %s", err)
			}
			for _, expected := range []string{"data/domainname=" + domainName} {
				res, err := services.BackupBucket.Search(expected, 100)

				if err != nil || len(res) == 0 {
					t.Errorf("S3 backup failed: object with prefix %v should be uploaded in S3: got error: %v or no data", expected, err)
				} else {
					nItemExpected := 1
					if len(res) != nItemExpected {
						t.Errorf("S3 backup failed: expected %v  and got %v", nItemExpected, len(res))
					}
				}
			}
		} else {
			log.Printf("[TestRealTime] Realtime Backup Disabled by CFN Parameter.")
			_ = services.BackupBucket.WaitForObject("data/domainName="+domainName, 240)
			for _, expected := range []string{"data/domainName=" + domainName} {
				res, err := services.BackupBucket.Search(expected, 100)

				if err != nil || len(res) > 0 {
					t.Errorf("S3 backup failed: object with prefix %v should not be uploaded in S3: got error: %v or data", expected, err)
				}
			}
		}
		wg.Done()
	}()

	// Validating S3 export
	objectsToCleanup := []string{}
	go func() {
		key := "profiles/domainname=" + domainName
		log.Printf("[TestRealTime] Validating S3 export, searching for file in %s", key)
		err = services.OutputBucket.WaitForObject("profiles/domainname="+domainName, 300)
		if err != nil {
			t.Errorf("[TestRealTime] Error waiting for S3 export: %s", err)
		}
		hasProfile1 := false
		hasProfile2 := false
		for _, expected := range []string{"profiles/domainname=" + domainName} {
			res, err := services.OutputBucket.Search(expected, 100)
			if err != nil || len(res) == 0 {
				t.Errorf("object with prefix %v should be uploaded in S3: got error: %v or no data", expected, err)
			} else {

				for _, r := range res {
					log.Printf("Checking domain %s content", expected)
					data, err := services.OutputBucket.GetParquetObj(r)
					if err != nil {
						t.Errorf("[TestMain] error getting object from S3: %v", err)
					}
					log.Printf("data: %v", data)
					var recs []travModel.Traveller
					if err = json.Unmarshal([]byte(data), &recs); err != nil {
						log.Printf("[TestMain] error unmarshalling json: %v", err)
					}
					if len(recs) == 0 {
						t.Errorf("[TestMain] S3 file should have at least one traveller")
						continue
					}
					for _, s3Traveller := range recs {
						log.Printf("Processing S3 line: %v", s3Traveller)
						if s3Traveller.PersonalEmailAddress != airBookingEmail1 {
							hasProfile1 = true
						}
						if s3Traveller.PersonalEmailAddress != airBookingEmail2 {
							hasProfile2 = true
						}
					}
					objectsToCleanup = append(objectsToCleanup, r)
				}

			}
		}
		if !hasProfile1 {
			t.Errorf("[TestRealTime] Profile 1 was not properly exported to S3")
		}
		if !hasProfile2 {
			t.Errorf("[TestRealTime] Profile 2 was not properly exported to S3")
		}
		wg.Done()
	}()
	wg.Wait()

	t.Cleanup(func() {
		log.Printf("Cleaning up S3 objects")
		for _, obj := range objectsToCleanup {
			err = services.OutputBucket.Delete("", obj)
			if err != nil {
				t.Errorf("[TestMain] error deleting object %v from S3: %v", obj, err)
			}
		}
	})

	t2AfterUnmerge, err := uptSdk.UnmergeProfilesAndWait(domainName, t1.ConnectID, t2.ConnectID, 120)
	if err != nil {
		t.Fatalf("Error unmerging profiles %s and %s", t1.ConnectID, t2.ConnectID)
	}
	if t2AfterUnmerge.ConnectID != t2.ConnectID {
		t.Fatalf("Error invalid unmerged profile %+v returned. should be %+v", t2AfterUnmerge.ConnectID, t2.ConnectID)
	}

	//Privacy Search
	err = uptSdk.CreatePrivacySearchAndWait(domainName, []string{t1.ConnectID, t2.ConnectID}, 120)
	if err != nil {
		t.Fatalf("Error creating privacy search %v", err)
	}

	res, err = uptSdk.ListPrivacySearches(domainName)
	if err != nil {
		t.Fatalf("Error listing privacy searches %v", err)
	}
	if len(utils.SafeDereference(res.PrivacySearchResults)) != 2 {
		t.Fatalf("Error listing privacy searches should return 2 results and have %+v", utils.SafeDereference(res.PrivacySearchResults))
	}
	err = validatePrivacySearchResults(uptSdk, res, t1.ConnectID, domainName, 2)
	if err != nil {
		t.Fatalf("invalid privacy search results: %v", err)
	}
	err = validatePrivacySearchResults(uptSdk, res, t2.ConnectID, domainName, 2)
	if err != nil {
		t.Fatalf("invalid privacy search results: %v", err)
	}
	err = uptSdk.PrivacyPurgeProfilesAndWait(domainName, []string{t1.ConnectID, t2.ConnectID}, 120)
	if err != nil {
		t.Fatalf("Error purging profiles %v", err)
	}
	err = uptSdk.CreatePrivacySearchAndWait(domainName, []string{t1.ConnectID, t2.ConnectID}, 120)
	if err != nil {
		t.Fatalf("Error creating privacy search after purge %v", err)
	}
	res, err = uptSdk.ListPrivacySearches(domainName)
	if err != nil {
		t.Fatalf("Error listing privacy searches after purge  %v", err)
	}
	if len(utils.SafeDereference(res.PrivacySearchResults)) != 2 {
		t.Fatalf("Error listing privacy searches should return 2 results and have %+v", utils.SafeDereference(res.PrivacySearchResults))
	}
	err = validatePrivacySearchResults(uptSdk, res, t1.ConnectID, domainName, 0)
	if err != nil {
		t.Fatalf("invalid privacy search results: %v", err)
	}
	err = validatePrivacySearchResults(uptSdk, res, t2.ConnectID, domainName, 0)
	if err != nil {
		t.Fatalf("invalid privacy search results: %v", err)
	}
}

func validatePrivacySearchResults(uptSdk upt_sdk.UptHandler, res backendmodel.ResponseWrapper, connectID string, domainName string, expected int) error {
	privacyRes := utils.SafeDereference(res.PrivacySearchResults)
	privacySearchResult := privacyRes[0]
	for _, result := range privacyRes {
		if result.ConnectId == connectID {
			privacySearchResult = result
			break
		}
	}
	if privacySearchResult.ConnectId != connectID {
		return fmt.Errorf("Expected could not find connect id %s in privacy search results: %+v", connectID, res.PrivacySearchResults)
	}
	if privacySearchResult.TotalResultsFound != expected {
		log.Printf("Invalid privacy search results TotalResultsFound value. retrieving search to investigates (connectid: %s)", privacySearchResult.ConnectId)
		getRes, err := uptSdk.GetPrivacySearch(domainName, privacySearchResult.ConnectId)
		if err != nil {
			log.Printf("Error retrieving privacy search results to investigate %v", err)
		}
		return fmt.Errorf("listing privacy searches should return %d results for connect id %s and have %+v (Full search result if available: %+v)", expected, connectID, privacySearchResult, getRes.PrivacySearchResult)
	}
	return nil
}

func getEventsFromSqs(sqsConfig *sqs.Config, sqsTimeout int) ([]ProfileChangeEvent, error) {
	log.Printf("[TestRealTime] Validating event export")
	allMessages := make(map[string]ProfileChangeEvent)
	i, max, duration := 0, sqsTimeout/5, 5*time.Second
	for i < max {
		messages, err := sqsConfig.Get(sqs.GetMessageOptions{})
		if err != nil {
			log.Printf("Error getting messages from queueu: %+v", err)
			return []ProfileChangeEvent{}, err
		}
		for _, peek := range messages.Peek {
			var event ProfileChangeEvent
			err := json.Unmarshal([]byte(peek.Body), &event)
			if err != nil {
				return []ProfileChangeEvent{}, err
			}
			allMessages[event.ID] = event
		}
		if len(allMessages) == int(messages.NMessages) {
			break
		}
		log.Println("[TestRealTime] Waiting 5 seconds for events to propagate")
		time.Sleep(duration)
		i++
	}

	log.Printf("[TestRealTime] Received %d messages", len(allMessages))
	ebEvents := make([]ProfileChangeEvent, 0, len(allMessages))
	// Create event array
	for _, value := range allMessages {
		ebEvents = append(ebEvents, value)
	}
	if i == max {
		return ebEvents, errors.New("timed out waiting for events to propagate")
	}
	return ebEvents, nil
}

// Use test_data to create test data that can be sent to the Kinesis
// stream for real-time processing.
func generateKinesisRec(path, objectType, domainName string, line int) kinesis.Record {
	testData, err := ReadJsonlFile(path)
	if err != nil {
		log.Printf("Could not read json file: %v", err)
	}
	if line >= len(testData) {
		log.Printf("Line %d is out of range", line)
		return kinesis.Record{}
	}
	data := testData[line]
	rec := BusinessObjectRecord{
		Domain:       domainName,
		ObjectType:   objectType,
		ModelVersion: data["modelVersion"].(string),
		Data:         data,
	}
	jsonData, err := json.Marshal(rec)
	if err != nil {
		log.Println(err)
		return kinesis.Record{}
	}
	return kinesis.Record{
		Pk:   fmt.Sprintf("%s-%s-%d", domainName, objectType, line),
		Data: string(jsonData),
	}
}

func sumRecords(traveler travModel.Traveller) int {
	var sum int
	sum += len(traveler.AirBookingRecords)
	sum += len(traveler.AirLoyaltyRecords)
	sum += len(traveler.AncillaryServiceRecords)
	sum += len(traveler.ClickstreamRecords)
	sum += len(traveler.CustomerServiceInteractionRecords)
	sum += len(traveler.EmailHistoryRecords)
	sum += len(traveler.HotelBookingRecords)
	sum += len(traveler.HotelLoyaltyRecords)
	sum += len(traveler.HotelStayRecords)
	sum += len(traveler.LoyaltyTxRecords)
	sum += len(traveler.PhoneHistoryRecords)
	return sum
}
