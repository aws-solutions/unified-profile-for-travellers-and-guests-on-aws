// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"testing"

	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"

	model "tah/upt/source/ucp-common/src/model/dynamo-schema"

	"github.com/aws/aws-lambda-go/events"
)

func TestReceivedEvent(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	indexTb := "ucp-cp-index-" + core.GenerateUniqueId()
	cpIndexPkName := "domainName"
	cpIndexSkName := "connectId"
	testDomainName := "test"
	indexDb, err := db.InitWithNewTable(indexTb, cpIndexPkName, cpIndexSkName, "", "")
	if err != nil {
		t.Errorf("[TestUpdatePortalConfig] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = indexDb.DeleteTable(indexTb)
		if err != nil {
			t.Errorf("error deleting table: %v", err)
		}
	})
	indexDb.WaitForTableCreation()

	sr := ServiceRegistry{
		DbConfig: indexDb,
		Tx:       tx,
	}

	profSerializable := map[string]interface{}{}
	prof := profilemodel.Profile{
		ProfileId: "1",
		Attributes: map[string]string{
			"external_connect_id": "2",
		},
	}
	profByte, err := json.Marshal(prof)
	if err != nil {
		t.Errorf("error marshalling profile: %v", err)
	}
	err = json.Unmarshal(profByte, &profSerializable)
	if err != nil {
		t.Errorf("error unmarshalling profile: %v", err)
	}

	cpEvent := CustomerProfileEvent{
		ObjectTypeName: "_profile",
		EventType:      "CREATED",
		DomainName:     testDomainName,
		Object:         profSerializable,
	}
	cpEventByte, err := json.Marshal(cpEvent)
	if err != nil {
		t.Errorf("error marshalling cp event: %v", err)
	}

	event := events.KinesisEvent{
		Records: []events.KinesisEventRecord{
			{
				Kinesis: events.KinesisRecord{
					Data:           cpEventByte,
					SequenceNumber: "1",
				},
			},
		},
	}

	batch, err := HandleRequestWithServices(context.Background(), event, sr)
	batchFailures := batch["batchItemFailures"]
	batchData := batchFailures.([]map[string]interface{})
	if len(batchData) != 0 {
		t.Fatalf("expected no failure, got %v", len(batchData))
	}

	var dynamoRecords []model.CpIdMap
	err = indexDb.FindStartingWith(testDomainName, "lcs_", &dynamoRecords)
	if err != nil {
		t.Errorf("error finding records: %v", err)
	}
	if len(dynamoRecords) != 1 {
		t.Errorf("expected 1 record, got %v", len(dynamoRecords))
	}
	if dynamoRecords[0].ConnectId != "lcs_2" {
		t.Errorf("expected connect id to be 2, got %v", dynamoRecords[0].ConnectId)
	}
	if dynamoRecords[0].CpId != "1" {
		t.Errorf("expected cp id to be 1, got %v", dynamoRecords[0].CpId)
	}

	err = indexDb.FindStartingWith(testDomainName, "cp_", &dynamoRecords)
	if err != nil {
		t.Errorf("error finding records: %v", err)
	}
	if len(dynamoRecords) != 1 {
		t.Errorf("expected 1 record, got %v", len(dynamoRecords))
	}
	if dynamoRecords[0].ConnectId != "cp_1" {
		t.Errorf("expected connect id to be 2, got %v", dynamoRecords[0].ConnectId)
	}
	if dynamoRecords[0].CpId != "2" {
		t.Errorf("expected cp id to be 1, got %v", dynamoRecords[0].CpId)
	}
}

// This test ensures that if we send 2 records, one of them being bad, we still write the second one
// and produce an error for the first. That way, a single error does not cause a batch failure.
func TestReceivedEventWithBadRecord(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	indexTb := "ucp-cp-index-" + core.GenerateUniqueId()
	cpIndexPkName := "domainName"
	cpIndexSkName := "connectId"
	testDomainName := "test"
	indexDb, err := db.InitWithNewTable(indexTb, cpIndexPkName, cpIndexSkName, "", "")
	if err != nil {
		t.Errorf("[TestUpdatePortalConfig] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = indexDb.DeleteTable(indexTb)
		if err != nil {
			t.Errorf("error deleting table: %v", err)
		}
	})
	indexDb.WaitForTableCreation()

	sr := ServiceRegistry{
		DbConfig: indexDb,
		Tx:       tx,
	}

	profSerializable := map[string]interface{}{}
	prof := profilemodel.Profile{
		ProfileId: "1",
		Attributes: map[string]string{
			"external_connect_id": "2",
		},
	}
	profByte, err := json.Marshal(prof)
	if err != nil {
		t.Errorf("error marshalling profile: %v", err)
	}
	err = json.Unmarshal(profByte, &profSerializable)
	if err != nil {
		t.Errorf("error unmarshalling profile: %v", err)
	}

	cpEvent := CustomerProfileEvent{
		ObjectTypeName: "_profile",
		EventType:      "CREATED",
		DomainName:     testDomainName,
		Object:         profSerializable,
	}
	cpEventByte, err := json.Marshal(cpEvent)
	if err != nil {
		t.Errorf("error marshalling cp event: %v", err)
	}

	profSerializableBad := map[string]interface{}{}
	profBad := profilemodel.Profile{
		ProfileId: "1",
		Attributes: map[string]string{
			"bad_data": "2",
		},
	}
	profByteBad, err := json.Marshal(profBad)
	if err != nil {
		t.Errorf("error marshalling profile: %v", err)
	}
	err = json.Unmarshal(profByteBad, &profSerializableBad)
	if err != nil {
		t.Errorf("error unmarshalling profile: %v", err)
	}

	cpEventBad := CustomerProfileEvent{
		ObjectTypeName: "_profile",
		EventType:      "CREATED",
		DomainName:     testDomainName,
		Object:         profSerializableBad,
	}
	cpEventByteBad, err := json.Marshal(cpEventBad)
	if err != nil {
		t.Errorf("error marshalling cp event: %v", err)
	}

	event := events.KinesisEvent{
		Records: []events.KinesisEventRecord{
			{
				Kinesis: events.KinesisRecord{
					Data:           cpEventByte,
					SequenceNumber: "1",
				},
			},
			{
				Kinesis: events.KinesisRecord{
					Data:           cpEventByteBad,
					SequenceNumber: "2",
				},
			},
		},
	}

	batch, err := HandleRequestWithServices(context.Background(), event, sr)
	batchFailures := batch["batchItemFailures"]
	batchData := batchFailures.([]map[string]interface{})
	if len(batchData) != 1 {
		t.Fatalf("expected 1 failure, got %v", len(batchData))
	}
	item, ok := batchData[0]["itemIdentifier"].(string)
	if !ok {
		t.Fatalf("unable to parse response")
	}
	if item != "2" {
		t.Fatalf("expected failure for sequence number 2, got %v", batchData[0]["itemIdentifier"])
	}

	if err != nil {
		t.Fatalf("error running request, %v: ", err)
	}

	var dynamoRecords []model.CpIdMap
	err = indexDb.FindStartingWith(testDomainName, "lcs_", &dynamoRecords)
	if err != nil {
		t.Errorf("error finding records: %v", err)
	}
	if len(dynamoRecords) != 1 {
		t.Errorf("expected 1 record, got %v", len(dynamoRecords))
	}
	if dynamoRecords[0].ConnectId != "lcs_2" {
		t.Errorf("expected connect id to be 2, got %v", dynamoRecords[0].ConnectId)
	}
	if dynamoRecords[0].CpId != "1" {
		t.Errorf("expected cp id to be 1, got %v", dynamoRecords[0].CpId)
	}

	err = indexDb.FindStartingWith(testDomainName, "cp_", &dynamoRecords)
	if err != nil {
		t.Errorf("error finding records: %v", err)
	}
	if len(dynamoRecords) != 1 {
		t.Errorf("expected 1 record, got %v", len(dynamoRecords))
	}
	if dynamoRecords[0].ConnectId != "cp_1" {
		t.Errorf("expected connect id to be 2, got %v", dynamoRecords[0].ConnectId)
	}
	if dynamoRecords[0].CpId != "2" {
		t.Errorf("expected cp id to be 1, got %v", dynamoRecords[0].CpId)
	}
}

func TestGracefulExitWithoutDynamoTable(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	testDomainName := "test"
	sr := ServiceRegistry{
		Tx: tx,
	}

	profSerializable := map[string]interface{}{}
	prof := profilemodel.Profile{
		ProfileId: "1",
		Attributes: map[string]string{
			"external_connect_id": "2",
		},
	}
	profByte, err := json.Marshal(prof)
	if err != nil {
		t.Errorf("error marshalling profile: %v", err)
	}
	err = json.Unmarshal(profByte, &profSerializable)
	if err != nil {
		t.Errorf("error unmarshalling profile: %v", err)
	}

	cpEvent := CustomerProfileEvent{
		ObjectTypeName: "_profile",
		EventType:      "CREATED",
		DomainName:     testDomainName,
		Object:         profSerializable,
	}
	cpEventByte, err := json.Marshal(cpEvent)
	if err != nil {
		t.Errorf("error marshalling cp event: %v", err)
	}

	event := events.KinesisEvent{
		Records: []events.KinesisEventRecord{
			{
				Kinesis: events.KinesisRecord{
					Data:           cpEventByte,
					SequenceNumber: "1",
				},
			},
		},
	}

	batch, err := HandleRequestWithServices(context.Background(), event, sr)
	if err != nil {
		t.Fatalf("should not error, got %v", err)
	}
	batchFailures := batch["batchItemFailures"]
	batchData := batchFailures.([]map[string]interface{})
	if len(batchData) != 1 {
		t.Fatalf("expected 1 failure, got %v", len(batchData))
	}
	item, ok := batchData[0]["itemIdentifier"].(string)
	if !ok {
		t.Fatalf("unable to parse response")
	}
	if item != "1" {
		t.Fatalf("expected failure for sequence number 2, got %v", batchData[0]["itemIdentifier"])
	}

}
