// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package dynamoSchema

// CP Id indexer object
type CpIdMap struct {
	DomainName string `json:"domainName"`
	ConnectId  string `json:"connectId"`
	CpId       string `json:"cpId"`
}
