// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/sqs"
	common "tah/upt/source/ucp-common/src/constant/admin"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-lambda-go/events"
)

func setUpMocks() (*customerprofiles.MockV2, *sqs.MockConfig) {
	cpMock := customerprofiles.InitMockV2()
	sqsMock := sqs.InitMock()
	return cpMock, sqsMock
}

func buildRequestString(profileId1, profileId2, testDomainName, txId string, mergeContext customerprofiles.ProfileMergeContext) (string, error) {
	requestBody := customerprofiles.MergeRequest{
		SourceID:      profileId1,
		TargetID:      profileId2,
		Context:       mergeContext,
		DomainName:    testDomainName,
		TransactionID: txId,
	}
	request, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	return string(request), nil
}

func TestMain(t *testing.T) {
	// Set up mocks
	cpMock, sqsMock := setUpMocks()

	testDomainName := "test_domain_name"
	txId := "test_transaction_id"
	profileId1 := uuid.NewString()
	profileId2 := uuid.NewString()
	mergeContext := customerprofiles.ProfileMergeContext{
		Timestamp:              time.Now().UTC(),
		MergeType:              "manual",
		ConfidenceUpdateFactor: 1,
		OperatorID:             "test_operator_id",
	}
	request, err := buildRequestString(profileId1, profileId2, testDomainName, txId, mergeContext)
	if err != nil {
		t.Fatal("[TestMain] Error building request")
	}

	req := events.SQSEvent{Records: []events.SQSMessage{
		{
			MessageId: "test_id1",
			Body:      request,
		},
	}}

	cpMock.On("FindCurrentParentOfProfile", testDomainName, profileId1).Return(profileId1, nil)
	cpMock.On("FindCurrentParentOfProfile", testDomainName, profileId2).Return(profileId2, nil)
	cpMock.On("MergeProfiles", profileId2, profileId1, mergeContext).Return("SUCCESS", nil)
	tx := core.NewTransaction(TX_MERGER, txId, core.LogLevelDebug)
	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")
	cpMock.On("SetTx", tx).Return()

	// Call Handler
	requests := []DeduplicatedMergeRequest{}
	for _, rec := range req.Records {
		rq := validateMergeRequest(tx, rec, sqsMock)
		if rq.TargetID != "" {
			tx.Info("Valid merge request. Adding to requests to process")
			requests = append(requests, rq)
		}
	}
	batchItemFailures, err := mergeProfiles(tx, cpMock, requests)

	if err != nil {
		t.Fatal("[TestMain] Failed to process request")
	}
	if len(batchItemFailures) != 0 {
		t.Fatal("[TestMain] Failed to process request")
	}
}

func TestUnmarshalFailure(t *testing.T) {
	// Set up mocks
	cpMock, sqsMock := setUpMocks()

	body := "INVALID_BODY"

	req := events.SQSEvent{Records: []events.SQSMessage{
		{
			MessageId: "test_id1",
			Body:      body,
		},
	}}

	sqsMock.On("SendWithStringAttributes", body, "mock.Anything").Return(nil)

	// Call Handler
	sqsBatchResponse, err := HandleRequestWithServices(core.NewTransaction(t.Name(), "", core.LogLevelDebug), context.Background(), req, cpMock, sqsMock)

	if err != nil {
		t.Fatal("[TestUnmarshalFailure] Failed to process request")
	}
	if len(sqsBatchResponse["batchItemFailures"]) != 0 {
		t.Fatal("[TestUnmarshalFailure] Failed to process request")
	}
}

func TestDomainNameValidation(t *testing.T) {
	// Set up mock CP
	cpMock, sqsMock := setUpMocks()

	testDomainName := "INVALID-DOMAIN-NAME"
	txId := "test_transaction_id"
	profileId1 := uuid.NewString()
	profileId2 := uuid.NewString()
	mergeContext := customerprofiles.ProfileMergeContext{
		Timestamp:              time.Now().UTC(),
		MergeType:              "manual",
		ConfidenceUpdateFactor: 1,
		OperatorID:             "test_operator_id",
	}
	request, err := buildRequestString(profileId1, profileId2, testDomainName, txId, mergeContext)
	if err != nil {
		t.Fatal("[TestDomainNameValidation] Error building request")
	}

	tx := core.NewTransaction(TX_MERGER, txId, core.LogLevelDebug)
	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")
	cpMock.On("SetTx", tx).Return()

	messageId := "test_id1"
	req := events.SQSEvent{Records: []events.SQSMessage{
		{
			MessageId: messageId,
			Body:      request,
		},
	}}

	messageAttributes := map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: "UCP merge error",
		common.SQS_MES_ATTR_TX_ID:          tx.TransactionID,
		common.SQS_MES_ATTR_MESSAGE:        errors.New("Domain name " + testDomainName + " invalid, must be in snake case").Error(),
	}
	sqsMock.On("SendWithStringAttributes", request, messageAttributes).Return(nil)

	// Call Handler map[string]interface{}, error
	sqsBatchResponse, err := HandleRequestWithServices(tx, context.Background(), req, cpMock, sqsMock)

	if err != nil {
		t.Fatal("[TestDomainNameValidation] Failed to process request")
	}
	if len(sqsBatchResponse["batchItemFailures"]) != 0 {
		t.Fatal("[TestDomainNameValidation] Failed to process request")
	}
}

func TestUUIDValidation(t *testing.T) {
	// Set up mocks
	cpMock, sqsMock := setUpMocks()

	testDomainName := "test_domain_name"
	txId := "test_transaction_id"
	profileId1 := uuid.NewString()
	profileId2 := uuid.NewString()
	nonValidProfileId := "non-valid-uuid"
	mergeContext := customerprofiles.ProfileMergeContext{
		Timestamp:              time.Now().UTC(),
		MergeType:              "manual",
		ConfidenceUpdateFactor: 1,
		OperatorID:             "test_operator_id",
	}
	request, err := buildRequestString(nonValidProfileId, profileId2, testDomainName, txId, mergeContext)
	if err != nil {
		t.Fatal("[TestUUIDValidation] Error building request")
	}

	cpMock.On("SetDomain", testDomainName).Return(nil)
	tx := core.NewTransaction(TX_MERGER, txId, core.LogLevelDebug)
	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")
	cpMock.On("SetTx", tx).Return()
	messageId := "test_id1"
	req := events.SQSEvent{Records: []events.SQSMessage{
		{
			MessageId: messageId,
			Body:      request,
		},
	}}

	messageAttributes := map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: "UCP merge error",
		common.SQS_MES_ATTR_MESSAGE:        errors.New("invalid source id").Error(),
		common.SQS_MES_ATTR_TX_ID:          txId,
	}
	sqsMock.On("SendWithStringAttributes", request, mock.Anything).Return(nil)

	// Call Handler
	sqsBatchResponse, err := HandleRequestWithServices(tx, context.Background(), req, cpMock, sqsMock)

	if err != nil {
		t.Fatal("[TestUUIDValidation] Failed to process request")
	}
	if len(sqsBatchResponse["batchItemFailures"]) != 0 {
		t.Fatal("[TestUUIDValidation] Failed to process request")
	}

	request2, err := buildRequestString(profileId1, nonValidProfileId, testDomainName, txId, mergeContext)
	if err != nil {
		t.Fatal("[TestUUIDValidation] Error building request")
	}

	req2 := events.SQSEvent{Records: []events.SQSMessage{
		{
			MessageId: messageId,
			Body:      request2,
		},
	}}

	messageAttributes = map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: "UCP merge error",
		common.SQS_MES_ATTR_MESSAGE:        errors.New("invalid target id").Error(),
		common.SQS_MES_ATTR_TX_ID:          txId,
	}
	sqsMock.On("SendWithStringAttributes", request2, messageAttributes).Return(nil)

	// Call Handler map[string]interface{}, error
	sqsBatchResponse2, err := HandleRequestWithServices(tx, context.Background(), req2, cpMock, sqsMock)

	if err != nil {
		t.Fatal("[TestUUIDValidation] Failed to process request")
	}
	if len(sqsBatchResponse2["batchItemFailures"]) != 0 {
		t.Fatal("[TestUUIDValidation] Failed to process request")
	}

}

func TestContextValidation(t *testing.T) {
	// Set up mocks
	cpMock, sqsMock := setUpMocks()
	testDomainName := "test_domain_name"
	txId := "test_transaction_id"
	profileId1 := uuid.NewString()
	profileId2 := uuid.NewString()
	noMergeTypeContext := customerprofiles.ProfileMergeContext{
		Timestamp:              time.Now().UTC(),
		ConfidenceUpdateFactor: 1,
		OperatorID:             "test_operator_id",
	}
	invalidConfidenceFactorContext := customerprofiles.ProfileMergeContext{
		Timestamp:              time.Now().UTC(),
		MergeType:              "ai",
		ConfidenceUpdateFactor: -1,
	}
	invalidRuleIdContext := customerprofiles.ProfileMergeContext{
		Timestamp:      time.Now().UTC(),
		MergeType:      "rule",
		RuleID:         "-123",
		RuleSetVersion: "test_operator_id",
	}
	invalidRuleSetVersionContext := customerprofiles.ProfileMergeContext{
		Timestamp:      time.Now().UTC(),
		MergeType:      "rule",
		RuleID:         "1",
		RuleSetVersion: "",
	}
	invalidOperatorIdContext := customerprofiles.ProfileMergeContext{
		Timestamp:  time.Now().UTC(),
		MergeType:  "manual",
		OperatorID: "",
	}

	cpMock.On("SetDomain", testDomainName).Return(nil)
	tx := core.NewTransaction(TX_MERGER, txId, core.LogLevelDebug)
	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")
	cpMock.On("SetTx", tx).Return()

	noMergeTypeRequest, err := buildRequestString(profileId1, profileId2, testDomainName, txId, noMergeTypeContext)
	if err != nil {
		t.Fatal("[TestContextValidation] Error building request")
	}
	invalidConfidenceFactorRequest, err := buildRequestString(profileId1, profileId2, testDomainName, txId, invalidConfidenceFactorContext)
	if err != nil {
		t.Fatal("[TestContextValidation] Error building request")
	}
	invalidRuleIdRequest, err := buildRequestString(profileId1, profileId2, testDomainName, txId, invalidRuleIdContext)
	if err != nil {
		t.Fatal("[TestContextValidation] Error building request")
	}
	invalidRuleSetVersionRequest, err := buildRequestString(profileId1, profileId2, testDomainName, txId, invalidRuleSetVersionContext)
	if err != nil {
		t.Fatal("[TestContextValidation] Error building request")
	}
	invalidOperatorIdRequest, err := buildRequestString(profileId1, profileId2, testDomainName, txId, invalidOperatorIdContext)
	if err != nil {
		t.Fatal("[TestContextValidation] Error building request")
	}

	req := events.SQSEvent{Records: []events.SQSMessage{
		{
			MessageId: "test_id1",
			Body:      noMergeTypeRequest,
		},
		{
			MessageId: "test_id2",
			Body:      invalidConfidenceFactorRequest,
		},
		{
			MessageId: "test_id3",
			Body:      invalidRuleIdRequest,
		},
		{
			MessageId: "test_id4",
			Body:      invalidRuleSetVersionRequest,
		},
		{
			MessageId: "test_id5",
			Body:      invalidOperatorIdRequest,
		},
	}}

	noMergeTypeAttributes := map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: "UCP merge error",
		common.SQS_MES_ATTR_MESSAGE:        errors.New("merge type is required").Error(),
		common.SQS_MES_ATTR_TX_ID:          txId,
	}
	invalidConfidenceFactorAttributes := map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: "UCP merge error",
		common.SQS_MES_ATTR_MESSAGE:        errors.New("confidence update factor is invalid").Error(),
		common.SQS_MES_ATTR_TX_ID:          txId,
	}
	invalidRuleIdAttributes := map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: "UCP merge error",
		common.SQS_MES_ATTR_MESSAGE:        errors.New("rule id is invalid").Error(),
		common.SQS_MES_ATTR_TX_ID:          txId,
	}
	invalidRuleSetVersionAttributes := map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: "UCP merge error",
		common.SQS_MES_ATTR_MESSAGE:        errors.New("rule set version is required").Error(),
		common.SQS_MES_ATTR_TX_ID:          txId,
	}
	invalidOperatorIdAttributes := map[string]string{
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: "UCP merge error",
		common.SQS_MES_ATTR_MESSAGE:        errors.New("operator id is required").Error(),
		common.SQS_MES_ATTR_TX_ID:          txId,
	}
	sqsMock.On("SendWithStringAttributes", noMergeTypeRequest, noMergeTypeAttributes).Return(nil)
	sqsMock.On("SendWithStringAttributes", invalidConfidenceFactorRequest, invalidConfidenceFactorAttributes).Return(nil)
	sqsMock.On("SendWithStringAttributes", invalidRuleIdRequest, invalidRuleIdAttributes).Return(nil)
	sqsMock.On("SendWithStringAttributes", invalidRuleSetVersionRequest, invalidRuleSetVersionAttributes).Return(nil)
	sqsMock.On("SendWithStringAttributes", invalidOperatorIdRequest, invalidOperatorIdAttributes).Return(nil)

	// Call Handler
	sqsBatchResponse2, err := HandleRequestWithServices(tx, context.Background(), req, cpMock, sqsMock)

	if err != nil {
		t.Fatal("[TestContextValidation] Failed to process request")
	}
	if len(sqsBatchResponse2["batchItemFailures"]) != 0 {
		t.Fatal("[TestContextValidation] Failed to process request")
	}
}

func TestPartialBatchFailure(t *testing.T) {
	// Set up mocks
	cpMock, sqsMock := setUpMocks()

	testDomainName := "test_domain_name"
	txId := "test_transaction_id"
	profileId1 := uuid.NewString()
	profileId2 := uuid.NewString()
	profileId3 := uuid.NewString()
	mergeContext := customerprofiles.ProfileMergeContext{
		Timestamp:              time.Now().UTC(),
		MergeType:              "manual",
		ConfidenceUpdateFactor: 1,
		OperatorID:             "test_operator_id",
	}
	request, err := buildRequestString(profileId1, profileId2, testDomainName, txId, mergeContext)
	if err != nil {
		t.Fatal("[TestPartialBatchFailure] Error building request")
	}
	request2, err := buildRequestString(profileId1, profileId3, testDomainName, txId, mergeContext)
	if err != nil {
		t.Fatal("[TestPartialBatchFailure] Error building request")
	}

	req := events.SQSEvent{Records: []events.SQSMessage{
		{
			MessageId: "test_id1",
			Body:      request,
		},
		{
			MessageId: "test_id2",
			Body:      request2,
		},
	}}

	tx := core.NewTransaction(TX_MERGER, txId, core.LogLevelDebug)
	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")

	cpMock.On("FindCurrentParentOfProfile", testDomainName, profileId1).Return(profileId1, nil)
	cpMock.On("FindCurrentParentOfProfile", testDomainName, profileId2).Return(profileId2, nil)
	cpMock.On("FindCurrentParentOfProfile", testDomainName, profileId1).Return(profileId1, nil)
	cpMock.On("FindCurrentParentOfProfile", testDomainName, profileId3).Return(profileId3, nil)
	cpMock.On("MergeProfiles", profileId2, profileId1, mergeContext).Return("SUCCESS", nil)
	cpMock.On("MergeProfiles", profileId3, profileId1, mergeContext).Return("FAILURE", errors.New("something failed"))

	// Call Handler
	requests := []DeduplicatedMergeRequest{}
	for _, rec := range req.Records {
		rq := validateMergeRequest(tx, rec, sqsMock)
		if rq.TargetID != "" {
			tx.Info("Valid merge request. Adding to requests to process")
			requests = append(requests, rq)
		}
	}
	batchItemFailures, err := mergeProfiles(tx, cpMock, requests)
	if err == nil {
		t.Fatal("[TestPartialBatchFailure] mergeProfiles should return an error")
	}
	if len(batchItemFailures) != 1 {
		t.Fatal("[TestPartialBatchFailure] Expected one request to fail")
	}
	if batchItemFailures[0].ItemIdentifier != "test_id2" {
		t.Fatal("[TestPartialBatchFailure] Expected failed request to be request2")
	}
}

func TestGetMarshalErrorString(t *testing.T) {
	testErr := errors.New("test error")
	str := getMarshalErrorString(testErr)
	if str != "Could not marshal data. error: "+testErr.Error() {
		t.Fatal("[TestGetMarshalErrorString] Failed to get expected string")
	}
}
