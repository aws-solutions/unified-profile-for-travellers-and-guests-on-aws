// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/sqs"
	common "tah/upt/source/ucp-common/src/constant/admin"
	uc "tah/upt/source/ucp-cp-writer/src/usecase"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type messageAttributes struct {
	requestType string
	domain      string
	txid        string
}

const sqsErrType string = "Write to Customer Profiles error"

var lambdaRegion = os.Getenv("AWS_REGION")
var dlqUrl = os.Getenv("DLQ_URL")
var cpWriterUrl = os.Getenv("CP_WRITER_QUEUE_URL")
var metricsSolutionId = os.Getenv("METRICS_SOLUTION_ID")
var metricsSolutionVersion = os.Getenv("METRICS_SOLUTION_VERSION")
var metricsUuid = os.Getenv("METRICS_UUID")
var metricsSendAnonymizedData = os.Getenv("SEND_ANONYMIZED_DATA")
var indexTable = os.Getenv("CP_INDEX_TABLE")
var indexTablePk = os.Getenv("INDEX_PK")
var indexTableSk = os.Getenv("INDEX_SK")

var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

var customerProfilesConfig = customerprofiles.Init(lambdaRegion, metricsSolutionId, metricsSolutionVersion)
var dlqConfig = sqs.InitWithQueueUrl(dlqUrl, lambdaRegion, metricsSolutionId, metricsSolutionVersion)
var cpQueueConfig = sqs.InitWithQueueUrl(cpWriterUrl, lambdaRegion, metricsSolutionId, metricsSolutionVersion)
var solutionsConfig = awssolutions.Init(metricsSolutionId, metricsSolutionVersion, metricsUuid, metricsSendAnonymizedData, LOG_LEVEL)
var indexDb = db.Init(indexTable, indexTablePk, indexTableSk, metricsSolutionId, metricsSolutionVersion)

/*
	Supported use cases
	- PutProfile
	- PutProfileObject
	- DeleteProfile
	- (merge profile is a combination of puts and delete)

	Messages are sent to the CP Writer queue with attributes. The attributes provide metadata
	for how the request should be handled. If attributes are not provided, the message will
	be sent to the error ddb table and presented to the customer.

	Otherwise, the message will be handled based on the request type and sent to Amazon Connect Customer Profiles.
	If there's an error with the request, we will send it to the DLQ. The request to ACCP is async, and we will not
	know if there is an error with ingestion, instead ACCP will send it to the DLQ provided at domain creation.
	That DLQ will also write to the error ddb table (and potentially retried based on the Retry Lambda).
*/

func HandleRequest(ctx context.Context, req events.SQSEvent) error {
	tx := core.NewTransaction("cp_writer", "", LOG_LEVEL)
	tx.SetPrefix("cp_writer")
	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")
	sr := uc.ServiceRegistry{
		Tx:              tx,
		CPConfig:        customerProfilesConfig,
		DLQConfig:       &dlqConfig,
		SolutionsConfig: solutionsConfig,
		DbConfig:        indexDb,
		CpQueue:         &cpQueueConfig,
	}
	return HandleRequestWithServices(ctx, req, sr)
}

func HandleRequestWithServices(ctx context.Context, req events.SQSEvent, sr uc.ServiceRegistry) error {
	sr.Tx.Info("Processing %d records", len(req.Records))

	for _, msg := range req.Records {
		attributes, err := validateMessage(msg)
		if err != nil {
			sr.Tx.Error("error validating message: %v", err)
			sendToDLQ("", msg, fmt.Sprintf("error validating request: %v", err), sr.Tx, sr.DLQConfig)
			continue
		}

		// Configure clients
		sr.Tx.TransactionID = attributes.txid
		sr.CPConfig.SetDomain(attributes.domain)
		sr.CPConfig.SetTx(sr.Tx)
		sr.Tx.Debug("Performing request: %s", attributes.requestType)

		// Handle usecase
		switch attributes.requestType {
		case uc.CPWriterRequestType_PutProfile:
			err = uc.PutProfile(msg, sr)
			if err != nil {
				sr.Tx.Error("error performing put_profile request: %v", err)
				sendToDLQ(attributes.domain, msg, fmt.Sprintf("error performing put_profile request: %v", err), sr.Tx, sr.DLQConfig)
			}
		case uc.CPWriterRequestType_PutProfileObject:
			err = uc.PutProfileObject(msg, sr)
			if err != nil {
				sr.Tx.Error("error performing put_profile_object request: %v", err)
				sendToDLQ(attributes.domain, msg, fmt.Sprintf("error performing put_profile_object request: %v", err), sr.Tx, sr.DLQConfig)
			}
		case uc.CPWriterRequestType_DeleteProfile:
			err = uc.DeleteProfile(msg, attributes.domain, sr)
			if err != nil {
				sr.Tx.Error("error performing delete_profile request: %v", err)
				sendToDLQ(attributes.domain, msg, fmt.Sprintf("error performing delete_profile request: %v", err), sr.Tx, sr.DLQConfig)
			}
		default:
			sr.Tx.Error("unsupported request type: %s", attributes.requestType)
			sendToDLQ(attributes.domain, msg, fmt.Sprintf("unsupported request type: %s", attributes.requestType), sr.Tx, sr.DLQConfig)
		}
	}

	// In case of errors, we send a formatted error message to the DLQ instead of returning an error and relying on the service.
	// This way, we are able to control the shape of the message sent to the DLQ (including metadata that is used by the Error Lambda).
	return nil
}

func validateMessage(msg events.SQSMessage) (messageAttributes, error) {
	requestType := msg.MessageAttributes["request_type"].StringValue
	if aws.ToString(requestType) == "" {
		return messageAttributes{}, errors.New("request type not provided")
	}
	domain := msg.MessageAttributes["domain"].StringValue
	if aws.ToString(domain) == "" {
		return messageAttributes{}, errors.New("domain not provided")
	}
	txid := msg.MessageAttributes["txid"].StringValue
	if aws.ToString(txid) == "" {
		return messageAttributes{}, errors.New("txid not provided")
	}

	return messageAttributes{
		requestType: aws.ToString(requestType),
		domain:      aws.ToString(domain),
		txid:        aws.ToString(txid),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

// Send error message directly to the DLQ. The Error Lambda processes error messages and checks for certain attributes.
// We set attributes here that are saved to the error item in DDB and displayed to users.
func sendToDLQ(domainName string, msg events.SQSMessage, errorMessage string, tx core.Transaction, dlqConfig sqs.IConfig) {
	messageJson, err := json.Marshal(msg)
	if err != nil {
		tx.Error("Error marshalling message, unable to send to DLQ: %v", err)
		return
	}
	err = dlqConfig.SendWithStringAttributes(string(messageJson), map[string]string{
		common.SQS_MES_ATTR_DOMAIN:         domainName,
		common.SQS_MES_ATTR_UCP_ERROR_TYPE: sqsErrType,
		common.SQS_MES_ATTR_MESSAGE:        errorMessage,
		common.SQS_MES_ATTR_TX_ID:          tx.TransactionID,
	})
	if err != nil {
		tx.Error("Error sending message to DLQ: %v", err)
	}
}
