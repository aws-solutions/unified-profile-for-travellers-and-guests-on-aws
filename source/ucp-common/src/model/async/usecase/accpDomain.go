// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	admin "tah/upt/source/ucp-common/src/model/admin"
)

type CreateDomainBody struct {
	KmsArn                string `json:"kmsArn"`
	AccpDlq               string `json:"accpDlq"`
	AccpDestinationStream string `json:"accpDestinationStream"`
	MatchBucketName       string `json:"matchBucket"`
	AccpSourceBucketName  string `json:"accpSourceBucket"`
	DomainName            string `json:"domainName"`
	Env                   string `json:"env"`
	HotelBookingBucket    string `json:"hotelBookingBucket"`
	AirBookingBucket      string `json:"airBookingBucket"`
	GuestProfileBucket    string `json:"guessProfileBucket"`
	PaxProfileBucket      string `json:"paxProfileBucket"`
	StayRevenueBucket     string `json:"stayRevenueBucket"`
	ClickstreamBucket     string `json:"clickstreamBucket"`
	CSIBucket             string `json:"csiBucket"`
	GlueSchemaPath        string `json:"glueSchemaPath"`
	RequestedVersion      int    `json:"requestedVersion"`
}

type DeleteDomainBody struct {
	Env        string `json:"env"`
	DomainName string `json:"domainName"`
}

type EmptyTableBody struct {
	TableName string `json:"tableName"`
}

type MergeProfilesBody struct {
	Rq     []admin.MergeRq
	Domain string
}

type UnmergeProfilesBody struct {
	Rq     admin.UnmergeRq
	Domain string
}
