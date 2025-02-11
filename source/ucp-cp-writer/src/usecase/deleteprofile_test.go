// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"encoding/json"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/sqs"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func TestDeleteProfile(t *testing.T) {
	tx, cpMock, sqsMock, awsSolutionsMock := initMocks()
	tx.TransactionID = "abcd1234"
	indexTb := "ucp-cp-index-" + core.GenerateUniqueId()
	cpIndexPkName := "domainName"
	cpIndexSkName := "connect_id"
	indexDb, err := db.InitWithNewTable(indexTb, cpIndexPkName, cpIndexSkName, "", "")
	if err != nil {
		t.Errorf("[%v] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = indexDb.DeleteTable(indexTb)
		if err != nil {
			t.Errorf("error deleting table: %v", err)
		}
	})
	indexDb.WaitForTableCreation()
	sr := ServiceRegistry{
		Tx:              tx,
		CPConfig:        cpMock,
		DLQConfig:       sqsMock,
		SolutionsConfig: awsSolutionsMock,
		DbConfig:        indexDb,
	}

	var opts []customerprofiles.DeleteProfileOptions
	cpMock.On("DeleteProfile", "id", opts).Return(nil)

	DeleteProfile(events.SQSMessage{Body: "{}"}, "", sr)
}

func TestDeleteProfileWithRetry(t *testing.T) {
	tx, cpMock, sqsMock, awsSolutionsMock := initMocks()
	cpQueueMock := sqs.InitMock()
	tx.TransactionID = "abcd1234"
	indexTb := "ucp-cp-index-" + core.GenerateUniqueId()
	cpIndexPkName := "domainName"
	cpIndexSkName := "connect_id"
	indexDb, err := db.InitWithNewTable(indexTb, cpIndexPkName, cpIndexSkName, "", "")
	if err != nil {
		t.Errorf("[%v] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = indexDb.DeleteTable(indexTb)
		if err != nil {
			t.Errorf("error deleting table: %v", err)
		}
	})
	indexDb.WaitForTableCreation()
	sr := ServiceRegistry{
		Tx:              tx,
		CPConfig:        cpMock,
		DLQConfig:       sqsMock,
		CpQueue:         cpQueueMock,
		SolutionsConfig: awsSolutionsMock,
		DbConfig:        indexDb,
	}

	deleteProfileRq := DeleteProfileRq{
		ProfileID:  "prof_1",
		NumRetries: 2,
	}
	body, err := json.Marshal(deleteProfileRq)
	if err != nil {
		t.Fatalf("[%s] error while marshalling request: %v", t.Name(), err)
	}
	deleteProfileMsg := events.SQSMessage{
		Body: string(body),
		MessageAttributes: map[string]events.SQSMessageAttribute{
			"request_type": {
				DataType:    "String",
				StringValue: aws.String(CPWriterRequestType_DeleteProfile),
			},
			"domain": {
				DataType:    "String",
				StringValue: aws.String("writer_test"),
			},
			"txid": {
				DataType:    "String",
				StringValue: aws.String("abcd1234"),
			},
		},
	}

	var opts []customerprofiles.DeleteProfileOptions
	cpMock.On("DeleteProfile", "id", opts).Return(nil)

	newDeleteProfileRq := deleteProfileRq
	newDeleteProfileRq.NumRetries = deleteProfileRq.NumRetries - 1
	newDeleteProfileRqByte, err := json.Marshal(newDeleteProfileRq)
	msgAttr := map[string]string{
		"request_type": CPWriterRequestType_DeleteProfile,
		"domain":       "writer_test",
		"txid":         "abcd1234",
	}
	cpQueueMock.On("SendWithStringAttributesAndDelay", string(newDeleteProfileRqByte), msgAttr).Return(nil)

	DeleteProfile(deleteProfileMsg, "writer_test", sr)
}

func initMocks() (core.Transaction, *customerprofiles.Mock, *sqs.MockConfig, awssolutions.MockConfig) {
	tx := core.NewTransaction("cp_writer", "", core.LogLevelDebug)
	cpConfig := customerprofiles.InitMock()
	dlqMock := sqs.InitMock()
	solutionsConfigMock := awssolutions.InitMock()

	return tx, cpConfig, dlqMock, solutionsConfigMock
}
