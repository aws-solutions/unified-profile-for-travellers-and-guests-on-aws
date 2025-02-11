// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	cpmodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/sqs"
	common "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/dynamo-schema"
	"tah/upt/source/ucp-cp-writer/src/usecase"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestPutProfile(t *testing.T) {
	t.Parallel()
	tx, cpMock, sqsMock, awsSolutionsMock := initMocks()
	tx.TransactionID = "abcd1234"
	sr := usecase.ServiceRegistry{
		Tx:              tx,
		CPConfig:        cpMock,
		DLQConfig:       sqsMock,
		SolutionsConfig: awsSolutionsMock,
	}
	p := cpmodel.Profile{
		ProfileId: "prof_1",
	}
	putProfileRq := usecase.PutProfileRq{
		Profile: p,
	}
	body, err := json.Marshal(putProfileRq)
	if err != nil {
		t.Fatalf("[%s] error while marshalling request: %v", t.Name(), err)
	}
	putProfileMsg := events.SQSMessage{
		Body: string(body),
		MessageAttributes: map[string]events.SQSMessageAttribute{
			"request_type": {
				DataType:    "String",
				StringValue: aws.String(usecase.CPWriterRequestType_PutProfile),
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

	cpMock.On("SetDomain", "writer_test").Return(nil)
	cpMock.On("SetTx", tx).Return(nil)
	cpMock.On("PutProfileAsObject", p).Return(nil)

	err = HandleRequestWithServices(context.Background(), events.SQSEvent{Records: []events.SQSMessage{putProfileMsg}}, sr)
	assert.NoError(t, err)

	cpMock.AssertExpectations(t)
	sqsMock.AssertExpectations(t)
	// awsSolutionsMock is not a testify mock struct, nothing to assert
}

func TestPutProfileObject(t *testing.T) {
	t.Parallel()
	tx, cpMock, sqsMock, awsSolutionsMock := initMocks()
	tx.TransactionID = "abcd1234"
	sr := usecase.ServiceRegistry{
		Tx:              tx,
		CPConfig:        cpMock,
		DLQConfig:       sqsMock,
		SolutionsConfig: awsSolutionsMock,
	}
	po := cpmodel.ProfileObject{
		ID: "obj_1",
	}
	putProfileObjRq := usecase.PutProfileObjectRq{
		ProfileID:     "prof_1",
		ProfileObject: po,
	}
	body, err := json.Marshal(putProfileObjRq)
	if err != nil {
		t.Fatalf("[%s] error while marshalling request: %v", t.Name(), err)
	}
	putProfileObjMsg := events.SQSMessage{
		Body: string(body),
		MessageAttributes: map[string]events.SQSMessageAttribute{
			"request_type": {
				DataType:    "String",
				StringValue: aws.String(usecase.CPWriterRequestType_PutProfileObject),
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

	cpMock.On("SetDomain", "writer_test").Return(nil)
	cpMock.On("SetTx", tx).Return(nil)
	cpMock.On("PutProfileObjectFromLcs", po, "prof_1").Return(nil)

	err = HandleRequestWithServices(context.Background(), events.SQSEvent{Records: []events.SQSMessage{putProfileObjMsg}}, sr)
	assert.NoError(t, err)

	cpMock.AssertExpectations(t)
	sqsMock.AssertExpectations(t)
	// awsSolutionsMock is not a testify mock struct, nothing to assert
}

func TestDeleteProfile(t *testing.T) {
	t.Parallel()
	tx, cpMock, sqsMock, awsSolutionsMock := initMocks()
	tx.TransactionID = "abcd1234"

	indexTb := "ucp-cp-index-" + core.GenerateUniqueId()
	cpIndexPkName := "domainName"
	cpIndexSkName := "connectId"
	indexDb, err := db.InitWithNewTable(indexTb, cpIndexPkName, cpIndexSkName, "", "")
	if err != nil {
		t.Fatalf("[%v] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = indexDb.DeleteTable(indexTb)
		if err != nil {
			t.Errorf("error deleting table: %v", err)
		}
	})
	indexDb.WaitForTableCreation()
	testObj := model.CpIdMap{
		DomainName: "writer_test",
		ConnectId:  "lcs_prof_1",
		CpId:       "connect_id_1",
	}
	indexDb.Save(testObj)

	sr := usecase.ServiceRegistry{
		Tx:              tx,
		CPConfig:        cpMock,
		DLQConfig:       sqsMock,
		DbConfig:        indexDb,
		SolutionsConfig: awsSolutionsMock,
	}
	deleteProfileRq := usecase.DeleteProfileRq{
		ProfileID: "prof_1",
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
				StringValue: aws.String(usecase.CPWriterRequestType_DeleteProfile),
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
	var expectedDeleteOptions []customerprofiles.DeleteProfileOptions

	cpMock.On("SetDomain", "writer_test").Return(nil)
	cpMock.On("SetTx", tx).Return(nil)

	cpMock.On("DeleteProfile", "connect_id_1", expectedDeleteOptions).Return(nil)

	err = HandleRequestWithServices(context.Background(), events.SQSEvent{Records: []events.SQSMessage{deleteProfileMsg}}, sr)
	assert.NoError(t, err)

	cpMock.AssertExpectations(t)
	sqsMock.AssertExpectations(t)
	// awsSolutionsMock is not a testify mock struct, nothing to assert
}

// ensure retry logic works correctly
func TestDeleteProfileWithRetries(t *testing.T) {
	t.Parallel()
	tx, cpMock, sqsMock, awsSolutionsMock := initMocks()
	cpQueueMock := sqs.InitMock()
	tx.TransactionID = "abcd1234"

	indexTb := "ucp-cp-index-" + core.GenerateUniqueId()
	cpIndexPkName := "domainName"
	cpIndexSkName := "connectId"
	indexDb, err := db.InitWithNewTable(indexTb, cpIndexPkName, cpIndexSkName, "", "")
	if err != nil {
		t.Fatalf("[%v] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = indexDb.DeleteTable(indexTb)
		if err != nil {
			t.Errorf("error deleting table: %v", err)
		}
	})
	indexDb.WaitForTableCreation()

	sr := usecase.ServiceRegistry{
		Tx:              tx,
		CPConfig:        cpMock,
		DLQConfig:       sqsMock,
		DbConfig:        indexDb,
		CpQueue:         cpQueueMock,
		SolutionsConfig: awsSolutionsMock,
	}
	deleteProfileRq := usecase.DeleteProfileRq{
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
				StringValue: aws.String(usecase.CPWriterRequestType_DeleteProfile),
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

	newDeleteProfileRq := deleteProfileRq
	newDeleteProfileRq.NumRetries = deleteProfileRq.NumRetries - 1
	newDeleteProfileRqByte, err := json.Marshal(newDeleteProfileRq)
	msgAttr := map[string]string{
		"request_type": usecase.CPWriterRequestType_DeleteProfile,
		"domain":       "writer_test",
		"txid":         "abcd1234",
	}

	cpQueueMock.On("SendWithStringAttributesAndDelay", string(newDeleteProfileRqByte), msgAttr).Return(nil)

	cpMock.On("SetDomain", "writer_test").Return(nil)
	cpMock.On("SetTx", tx).Return(nil)

	err = HandleRequestWithServices(context.Background(), events.SQSEvent{Records: []events.SQSMessage{deleteProfileMsg}}, sr)
	assert.NoError(t, err)

	cpMock.AssertExpectations(t)
	sqsMock.AssertExpectations(t)
	cpQueueMock.AssertExpectations(t)
	// awsSolutionsMock is not a testify mock struct, nothing to assert
}

// Messages are expected to have certain attributes for processing.
// If the attributes are not provided, send to the DLQ for processing and entry into the error ddb table.
func TestInvalidMessage(t *testing.T) {
	t.Parallel()
	tx, cpMock, sqsMock, awsSolutionsMock := initMocks()
	tx.TransactionID = "abcd1234"
	sr := usecase.ServiceRegistry{
		Tx:              tx,
		CPConfig:        cpMock,
		DLQConfig:       sqsMock,
		SolutionsConfig: awsSolutionsMock,
	}
	invalidMsg := events.SQSMessage{
		Body: "invalid",
		MessageAttributes: map[string]events.SQSMessageAttribute{
			"invalid": {
				DataType:    "String",
				StringValue: aws.String("invalid"),
			},
		},
	}

	msgJson, err := json.Marshal(invalidMsg)
	if err != nil {
		t.Fatalf("[%s] error while marshalling request: %v", t.Name(), err)
	}
	attributes := map[string]string{
		common.SQS_MES_ATTR_DOMAIN:         "",
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: sqsErrType,
		common.SQS_MES_ATTR_MESSAGE:        "error validating request: request type not provided",
		common.SQS_MES_ATTR_TX_ID:          "abcd1234",
	}
	sqsMock.On("SendWithStringAttributes", string(msgJson), attributes).Return(nil)

	err = HandleRequestWithServices(context.Background(), events.SQSEvent{Records: []events.SQSMessage{invalidMsg}}, sr)
	assert.NoError(t, err) // no error, bad message was sent to DLQ and ends up in error ddb table

	cpMock.AssertExpectations(t)
	sqsMock.AssertExpectations(t)
	// awsSolutionsMock is not a testify mock struct, nothing to assert
}

// Same setup as PutProfile, except
// 1. service returns an error message
// 2. expect error is sent to DLQ
func TestPutProfileError(t *testing.T) {
	t.Parallel()
	tx, cpMock, sqsMock, awsSolutionsMock := initMocks()
	tx.TransactionID = "abcd1234"
	sr := usecase.ServiceRegistry{
		Tx:              tx,
		CPConfig:        cpMock,
		DLQConfig:       sqsMock,
		SolutionsConfig: awsSolutionsMock,
	}
	p := cpmodel.Profile{
		ProfileId: "prof_1",
	}
	putProfileRq := usecase.PutProfileRq{
		Profile: p,
	}
	body, err := json.Marshal(putProfileRq)
	if err != nil {
		t.Fatalf("[%s] error while marshalling request: %v", t.Name(), err)
	}
	putProfileMsg := events.SQSMessage{
		Body: string(body),
		MessageAttributes: map[string]events.SQSMessageAttribute{
			"request_type": {
				DataType:    "String",
				StringValue: aws.String(usecase.CPWriterRequestType_PutProfile),
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

	cpMock.On("SetDomain", "writer_test").Return(nil)
	cpMock.On("SetTx", tx).Return(nil)
	cpMock.On("PutProfileAsObject", p).Return(fmt.Errorf("test error"))

	msgJson, err := json.Marshal(putProfileMsg)
	if err != nil {
		t.Fatalf("[%s] error while marshalling request: %v", t.Name(), err)
	}
	attributes := map[string]string{
		common.SQS_MES_ATTR_DOMAIN:         "writer_test",
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: sqsErrType,
		common.SQS_MES_ATTR_MESSAGE:        "error performing put_profile request: test error",
		common.SQS_MES_ATTR_TX_ID:          "abcd1234",
	}
	sqsMock.On("SendWithStringAttributes", string(msgJson), attributes).Return(nil)

	err = HandleRequestWithServices(context.Background(), events.SQSEvent{Records: []events.SQSMessage{putProfileMsg}}, sr)
	assert.NoError(t, err)

	cpMock.AssertExpectations(t)
	sqsMock.AssertExpectations(t)
	// awsSolutionsMock is not a testify mock struct, nothing to assert
}

func initMocks() (core.Transaction, *customerprofiles.Mock, *sqs.MockConfig, awssolutions.MockConfig) {
	tx := core.NewTransaction("cp_writer", "", core.LogLevelDebug)
	cpConfig := customerprofiles.InitMock()
	dlqMock := sqs.InitMock()
	solutionsConfigMock := awssolutions.InitMock()

	return tx, cpConfig, dlqMock, solutionsConfigMock
}
