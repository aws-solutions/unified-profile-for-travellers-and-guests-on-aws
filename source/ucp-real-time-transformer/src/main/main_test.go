// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/awssolutions"
	core "tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	sqs "tah/upt/source/tah-core/sqs"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

var UCP_REGION = getRegion()

func TestSqsErrors(t *testing.T) {
	t.Parallel()

	sqsConfig := sqs.Init(UCP_REGION, "", "")
	solutionsConfig := awssolutions.InitMock()
	qName := "Test-Queue-2" + core.GenerateUniqueId()
	_, err := sqsConfig.Create(qName)
	if err != nil {
		t.Errorf("[%v] Could not create test queue", t.Name())
	}
	t.Cleanup(func() {
		log.Printf("[%v] Deleting SQS queue", t.Name())
		err = sqsConfig.DeleteByName(qName)
		if err != nil {
			t.Errorf("[%v] error deleting queue by name: %v", t.Name(), err)
		}
	})

	// Case 1: illformatted record
	// Case 2: non existent domain
	req := events.KinesisEvent{Records: []events.KinesisEventRecord{
		{
			Kinesis: events.KinesisRecord{
				Data: []byte("illformated record"),
			},
		},
		{
			Kinesis: events.KinesisRecord{
				Data: []byte("{\"objectType\": \"clickstream\", \"modelVersion\": \"1.0\",\"domain\": \"nonexistent_domain\", \"data\": [{}]}"),
			},
		},
	}}
	HandleRequestWithServices(context.Background(), req, sqsConfig, solutionsConfig)

	// Case 3: empty data
	lcsMock := customerprofiles.InitMockV2()
	var emptyData interface{}
	bytes := []byte("{}")
	err = json.Unmarshal(bytes, &emptyData)
	if err != nil {
		t.Fatalf("[%v] error unmarshalling: %v", t.Name(), err)
	}
	emptyBusinessRecord := BusinessObjectRecord{KinesisRecord: events.KinesisEventRecord{Kinesis: events.KinesisRecord{Data: []byte("empty")}}, ObjectType: "clickstream"}
	err = processDataRecord(core.NewTransaction(t.Name(), "", core.LogLevelDebug), emptyBusinessRecord, emptyData, sqsConfig, lcsMock)

	content, err1 := sqsConfig.Get(sqs.GetMessageOptions{})
	log.Printf("content: %+v", content)
	if err1 != nil {
		t.Fatalf("[%v] error getting msg from queue: %v", t.Name(), err1)
	}
	if len(content.Peek) == 0 {
		t.Fatalf("[%v] no record returned", t.Name())
	}
	if content.NMessages != 3 {
		t.Fatalf("[%v] Should have 3 records", t.Name())
	}

}

func TestParseAccpRecord(t *testing.T) {
	t.Parallel()
	count := map[string]int{}
	rec, expectedCount := buildRecord("air_booking", "air_booking", "test_domain", "")
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	for _, accpRec := range rec.Data {
		_, rectype, _, _, _, err := parseAccpRecord(tx, accpRec)
		if err != nil {
			t.Errorf("[parseAccpRecord] error parsing record: %v", err)
		}
		count[rectype]++
	}
	log.Printf("count: %v", count)
	log.Printf("expected: %v", expectedCount)

	for key, c := range count {
		if c != expectedCount[key] {
			t.Errorf("[parseAccpRecord] should have  %v accp records of type %v got %v", expectedCount[key], key, c)
		}
	}
}

func TestIngestProfileObject(t *testing.T) {
	t.Parallel()
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	accpRecString := `{"traveller_id": "abcde"}`
	accpObjectType := "air_booking"
	domain := "test_domain"
	accpDomain := customerprofiles.Domain{Name: domain}
	domains := []customerprofiles.Domain{accpDomain}
	profile := profilemodel.Profile{
		ProfileId: "test_connect_id",
	}
	profiles := []profilemodel.Profile{}
	mappings := []customerprofiles.ObjectMapping{}
	accp := customerprofiles.InitMock(&accpDomain, &domains, &profile, &profiles, &mappings)
	accp.Objects = []profilemodel.ProfileObject{{
		ID: "booking_1",
		AttributesInterface: map[string]interface{}{
			"traveller_id": "abcde",
			"attribute_1":  "value_1",
			"attribute_2":  "value_2",
		},
	}}
	sqsc := sqs.Init(UCP_REGION, "", "")
	qName := "test-queue-ingest-profile-obj" + core.GenerateUniqueId()
	_, err := sqsc.Create(qName)
	if err != nil {
		t.Errorf("[IngestProfileObject] Could not create test queue")
	}
	err = ingestProfileObject(tx, accpRecString, accpObjectType, domain, accp, sqsc, "abcde")
	if err != nil {
		t.Errorf("[IngestProfileObject]error  ingesting to profile %v", err)
	}

	diff, err := handlePartialCase(
		tx,
		accpObjectType,
		"booking_1",
		"abcde",
		map[string]interface{}{
			"attribute_1": "value_1_changed",
			"attribute_2": "value_2_changed",
		},
		accp,
		[]string{"air_booking.attribute_1", "hotel_booking.attribute_2"},
	)
	if err != nil {
		t.Errorf("[IngestProfileObject]error  handling partial case%v", err)
	}
	expected := `{"attribute_1":"value_1_changed","attribute_2":"value_2","traveller_id":"abcde"}`
	if diff != expected {
		t.Errorf("[IngestProfileObject]error partial diff should be %s and not %v", expected, diff)
	}
	diff, err = handlePartialCase(
		tx,
		accpObjectType,
		"booking_1",
		"abcde",
		map[string]interface{}{
			"attribute_1": "value_1_changed",
			"attribute_2": "value_2_changed",
		},
		accp,
		[]string{"air_booking.*", "hotel_booking.attribute_2"},
	)
	if err != nil {
		t.Errorf("[IngestProfileObject]error  handling partial case%v", err)
	}
	expected = `{"attribute_1":"value_1_changed","attribute_2":"value_2_changed","traveller_id":"abcde"}`
	if diff != expected {
		t.Errorf("[IngestProfileObject]error partial diff should be %s and not %v", expected, diff)
	}
	err = sqsc.DeleteByName(qName)
	if err != nil {
		t.Errorf("[IngestProfileObject]error deleting queue by name: %v", err)
	}

}

func TestProcessDataRecord(t *testing.T) {
	t.Parallel()
	sqsc := sqs.Init(UCP_REGION, "", "")
	qName := "test-queue-ingest-profile-obj" + core.GenerateUniqueId()
	_, err := sqsc.Create(qName)
	if err != nil {
		t.Errorf("[IngestProfileObject] Could not create test queue")
	}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	domain := "test_domain"
	accpDomain := customerprofiles.Domain{Name: domain}
	domains := []customerprofiles.Domain{accpDomain}
	profile := profilemodel.Profile{
		ProfileId: "test_connect_id",
	}
	profiles := []profilemodel.Profile{}
	mappings := []customerprofiles.ObjectMapping{}
	accp := customerprofiles.InitMock(&accpDomain, &domains, &profile, &profiles, &mappings)

	hotelBookingRec := map[string]interface{}{
		"traveller_id":   "abcde",
		"object_type":    "hotel_boking",
		"accp_object_id": "abcde",
		"profile_id":     "abcde",
	}
	guestProfileRec := map[string]interface{}{
		"traveller_id":   "abcde",
		"object_type":    "guest_profile",
		"accp_object_id": "abcde",
		"profile_id":     "abcde",
	}
	travelerRecordsMergeWrongObject := BusinessObjectRecord{
		TravellerID:        "abcde",
		TxID:               tx.TransactionID,
		Domain:             domain,
		Mode:               "merge",
		MergeModeProfileID: "abcde",
		ObjectType:         "hotel_boking",
		ModelVersion:       "1.0",
		Data:               []interface{}{hotelBookingRec},
		KinesisRecord:      events.KinesisEventRecord{Kinesis: events.KinesisRecord{Data: []byte("test_data")}},
	}
	travelerRecordsMergeRightObject := BusinessObjectRecord{
		TravellerID:        "abcde",
		TxID:               tx.TransactionID,
		Domain:             domain,
		Mode:               "merge",
		MergeModeProfileID: "abcde",
		ObjectType:         "guest_profile",
		ModelVersion:       "1.0",
		Data:               []interface{}{guestProfileRec},
		KinesisRecord:      events.KinesisEventRecord{Kinesis: events.KinesisRecord{Data: []byte("test_data")}},
	}
	travelerRecordsPartial := BusinessObjectRecord{
		TravellerID:        "abcde",
		TxID:               tx.TransactionID,
		Domain:             domain,
		Mode:               "partial",
		MergeModeProfileID: "abcde",
		ObjectType:         "guest_profile",
		ModelVersion:       "1.0",
		Data:               []interface{}{guestProfileRec},
		KinesisRecord:      events.KinesisEventRecord{Kinesis: events.KinesisRecord{Data: []byte("test_data")}},
	}

	processDataRecord(tx, travelerRecordsMergeWrongObject, hotelBookingRec, sqsc, accp)
	processDataRecord(tx, travelerRecordsMergeRightObject, guestProfileRec, sqsc, accp)
	processDataRecord(tx, travelerRecordsPartial, guestProfileRec, sqsc, accp)

	err = sqsc.DeleteByName(qName)
	if err != nil {
		t.Errorf("[IngestProfileObject]error deleting queue by name: %v", err)
	}
}

func TestHandleMergeCase(t *testing.T) {
	t.Parallel()
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	travelerIdToMergeInto := "tid_1"
	travelerIdToMerge := "tid_2"
	domain := "test_domain"
	accpDomain := customerprofiles.Domain{Name: domain}
	domains := []customerprofiles.Domain{accpDomain}
	profile := profilemodel.Profile{}
	profiles := []profilemodel.Profile{}
	mappings := []customerprofiles.ObjectMapping{}
	accp := customerprofiles.InitMock(&accpDomain, &domains, &profile, &profiles, &mappings)
	err := handleMergeCase(tx, "", travelerIdToMerge, accp)
	if err == nil {
		t.Errorf("[HandleMergeCase] should return and error with an empty profile to merge into")
	}
	connectID := "connectID"
	accp.Profile = &profilemodel.Profile{
		ProfileId: connectID,
	}
	err = handleMergeCase(tx, travelerIdToMergeInto, travelerIdToMerge, accp)
	if err != nil {
		t.Errorf("[HandleMergeCase] returned error: %v", err)
	}
	if len(accp.Merged) != 2 {
		t.Errorf("[HandleMergeCase] ACCP client mock should have 2 profiles")
	}
	if accp.Merged[0] != connectID {
		t.Errorf(
			"[HandleMergeCase] merge profile should be been called with the correct travelerIdToMergeInto=%v and not %v",
			connectID,
			accp.Merged[0],
		)
	}
	if accp.Merged[1] != connectID {
		t.Errorf(
			"[HandleMergeCase] merge profile should be been called with the correct profileIdtoMerge=%v and not %v",
			connectID,
			accp.Merged[1],
		)
	}
}

func buildRecord(objectType string, folder string, domain string, mode string) (BusinessObjectRecord, map[string]int) {
	newMode := ""
	if mode != "" {
		newMode = "_" + mode
	}
	content, err := os.ReadFile("../../../test_data/" + folder + "/data1_expected" + newMode + ".json")
	expectedCount := map[string]int{}
	log.Printf("%s", string(content))

	if err != nil {
		log.Printf("Error reading JSON file: %s", err)
		return BusinessObjectRecord{}, expectedCount
	}

	var v map[string][]interface{}
	var recs []interface{}
	json.Unmarshal(content, &v)
	typeMap := map[string]string{
		"air_booking_recs":  "air_booking",
		"common_email_recs": "email_history",
		"common_phone_recs": "phone_history",
		"air_loyalty_recs":  "air_loyalty",
	}
	for key, accpType := range typeMap {
		expectedCount[accpType] = len(v[key])
		recs = append(recs, v[key]...)
	}

	return BusinessObjectRecord{
		Domain:       domain,
		ObjectType:   objectType,
		ModelVersion: "1",
		Data:         recs,
		Mode:         mode,
	}, expectedCount
}

func getRegion() string {
	//getting region for local testing
	region := os.Getenv("UCP_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}

func TestGroupByTravellerDomain(t *testing.T) {
	t.Parallel()
	businessObjectRecords := []BusinessObjectRecord{
		{
			TravellerID:  "traveller_1",
			Domain:       "test_domain1",
			ObjectType:   "air_booking",
			ModelVersion: "1",
			Data:         []interface{}{},
		},
		{
			TravellerID:  "traveller_1",
			Domain:       "test_domain2",
			ObjectType:   "hotel_booking",
			ModelVersion: "1",
			Data:         []interface{}{},
		},
		{
			TravellerID:  "traveller_2",
			Domain:       "test_domain1",
			ObjectType:   "air_booking",
			ModelVersion: "1",
			Data:         []interface{}{},
		},
		{
			// Edge case without traveller id
			Domain:       "test_domain2",
			ObjectType:   "air_booking",
			ModelVersion: "1",
			Data:         []interface{}{},
		},
		{
			// Edge case without domain
			TravellerID:  "traveller_2",
			ObjectType:   "air_booking",
			ModelVersion: "1",
			Data:         []interface{}{},
		},
	}
	grouped := groupByDomain(businessObjectRecords)
	if len(grouped) != 3 {
		t.Errorf("domain map should have 3 record groups")
	}
	if len(grouped["test_domain2"]) != 2 || len(grouped["test_domain1"]) != 2 || len(grouped[""]) != 1 {
		t.Errorf("records are not grouped by domain properly: %+v", grouped)
	}
}
