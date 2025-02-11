// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package upt_sdk

import (
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/utils/config"
)

////////////////////////////////////////////////////////////////////////////////////////////////
// Unified Profile for Travelers and Guests on AWS (UPT) Software Development Kit for Golang
////////////////////////////////////////////////////////////////////////////////////////////////
// This module defines go structs and functions allowing to interact with UPT programmatically through its API.
// It is composed of teh following files
// - upt_sdk.go includes struct definition and the init function for the SDK
// - upt_sdk_auth.go defines the functions for the Cognito authentication process (this include test user creation/deletion)
// - upt_sdk_admin.go defines the functions for the admin API (CreateDomain, DeleteDomain, RunJobs...)
// - upt_sdk_profile defines the functions for the profile API (SearchProfile, RetrieveProfile...)
// - uptSdk_test.go includes the test code for this module
// - upt_sdk_usecases.go defines functions that execute more sophisticated use cases orchestrating multiple api call (For example CreateDomainAndWait for async domain creation)
////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////////
// UPT SDK
// For now this is only sued for testing advanced
//
//////////////////////////////

const DOMAIN_HEADER = "Customer-Profiles-Domain"

type UptHandler struct {
	Client              core.HttpService
	CognitoTokenEnpoint core.HttpService
	CognitoClientID     string
	CognitoClient       *cognito.CognitoConfig
	RefreshToken        string
	AccessToken         string
}

type UPTConfig struct {
	ApiEndpoint     string
	CognitoEndpoint string
}

func Init(cfg config.InfraConfig) (UptHandler, error) {
	return UptHandler{
		Client:              core.HttpInit(cfg.ApiBaseUrl+"/api/ucp", core.LogLevelDebug),
		CognitoTokenEnpoint: core.HttpInit(cfg.CognitoTokenEndpoint, core.LogLevelDebug),
		CognitoClientID:     cfg.CognitoClientId,
		CognitoClient:       cognito.InitWithClientID(cfg.CognitoUserPoolId, cfg.CognitoClientId, "", "", core.LogLevelDebug),
	}, nil
}
