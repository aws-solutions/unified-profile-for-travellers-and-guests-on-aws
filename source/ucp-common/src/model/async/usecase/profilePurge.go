// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

type PurgeProfileDataBody struct {
	ConnectIds     []string `json:"connectIds"`
	DomainName     string   `json:"domainName"`
	AgentCognitoId string   `json:"agentCognitoId"`
}

type PurgeOptions struct {
	OperatorID string `json:"operatorId"`
}
