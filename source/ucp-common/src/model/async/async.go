// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package async

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/tah-core/sqs"
)

type SsmApi interface {
	GetParametersByPath(context.Context, string) (map[string]string, error)
}

type Services struct {
	AccpConfig        customerprofileslcs.ICustomerProfileLowCostConfig
	ConfigDB          db.DBConfig
	MatchDB           db.DBConfig
	ErrorDB           db.DBConfig
	PrivacyResultsDB  db.DBConfig
	PortalConfigDB    db.DBConfig
	SolutionsConfig   awssolutions.IConfig
	CognitoConfig     cognito.ICognitoConfig
	GlueConfig        glue.IConfig
	RetryLambdaConfig lambda.IConfig
	S3ExciseSqsConfig sqs.Config
	SsmConfig         SsmApi
	EcsConfig         customerprofileslcs.EcsApi
	AsyncLambdaConfig lambda.IConfig
	Env               map[string]string
}

type AsyncInvokePayload struct {
	EventID       string      `json:"eventId"`
	Usecase       string      `json:"usecase"`
	TransactionID string      `json:"transactionId"`
	Body          interface{} `json:"body"`
}

type AsyncEvent struct {
	EventID     string    `json:"item_type"`
	Usecase     string    `json:"item_id"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"lastUpdated"`
	Progress    int       `json:"progress"`
}

// Event status options
const (
	EVENT_STATUS_INVOKED string = "invoked"
	EVENT_STATUS_RUNNING string = "running"
	EVENT_STATUS_SUCCESS string = "success"
	EVENT_STATUS_FAILED  string = "failed"
)

// Usecases
const (
	USECASE_CREATE_DOMAIN         string = "createDomain"
	USECASE_DELETE_DOMAIN         string = "deleteDomain"
	USECASE_EMPTY_TABLE           string = "emptyTable"
	USECASE_MERGE_PROFILES        string = "mergeProfiles"
	USECASE_UNMERGE_PROFILES      string = "unmergeProfiles"
	USECASE_CREATE_PRIVACY_SEARCH string = "createPrivacySearch"
	USECASE_PURGE_PROFILE_DATA    string = "purgeProfileData"
	USECASE_START_BATCH_ID_RES    string = "startBatchIdRes"
	USECASE_REBUILD_CACHE         string = "rebuildCache"
)

var ASYNC_USECASE_LIST = []string{
	USECASE_CREATE_DOMAIN,
	USECASE_DELETE_DOMAIN,
	USECASE_EMPTY_TABLE,
	USECASE_MERGE_PROFILES,
	USECASE_UNMERGE_PROFILES,
	USECASE_CREATE_PRIVACY_SEARCH,
	USECASE_PURGE_PROFILE_DATA,
	USECASE_REBUILD_CACHE,
}

func Parse(raw string) (AsyncInvokePayload, error) {
	// Decode base64 string
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return AsyncInvokePayload{}, err
	}

	// Unmarshal JSON
	var payload AsyncInvokePayload
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return AsyncInvokePayload{}, err
	}

	return payload, nil
}
