// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"time"
)

type CreatePrivacySearchBody struct {
	ConnectIds []string `json:"connectIds"`
	DomainName string   `json:"domainName"`
}

type LocationType string

const (
	S3LocationType             LocationType = "S3"
	ProfileStorageLocationType LocationType = "Profile Storage"
	MatchDbLocationType        LocationType = "Entity Resolution Matches"
)

type PrivacySearchResult struct {
	DomainName        string                    `json:"domainName"`
	ConnectId         string                    `json:"connectId"`
	Status            PrivacySearchStatus       `json:"status"`
	ErrorMessage      string                    `json:"errorMessage"`
	SearchDate        time.Time                 `json:"searchDate"`
	LocationResults   map[LocationType][]string `json:"locationResults"`
	TxID              string                    `json:"txId"`
	TotalResultsFound int                       `json:"totalResultsFound"`
}

type PurgeStatusResult struct {
	DomainName     string `json:"domainName"`
	IsPurgeRunning bool   `json:"isPurgeRunning"`
}

type PurgeFromS3Body struct {
	S3Path     string   `json:"s3Path"`
	ConnectIds []string `json:"connectIds"`
	DomainName string   `json:"domainName"`
	TxID       string   `json:"txId"`
}

// Event status options
type PrivacySearchStatus string

const (
	PRIVACY_STATUS_SEARCH_INVOKED PrivacySearchStatus = "search_invoked"
	PRIVACY_STATUS_SEARCH_RUNNING PrivacySearchStatus = "search_running"
	PRIVACY_STATUS_SEARCH_SUCCESS PrivacySearchStatus = "search_success"
	PRIVACY_STATUS_SEARCH_FAILED  PrivacySearchStatus = "search_failed"
	PRIVACY_STATUS_PURGE_INVOKED  PrivacySearchStatus = "purge_invoked"
	PRIVACY_STATUS_PURGE_RUNNING  PrivacySearchStatus = "purge_running"
	PRIVACY_STATUS_PURGE_SUCCESS  PrivacySearchStatus = "purge_success"
	PRIVACY_STATUS_PURGE_FAILED   PrivacySearchStatus = "purge_failed"
)

type PrivacySearchPurgeStatusPriorityMap map[PrivacySearchStatus]int

var PrivacySearchStatusPriority = PrivacySearchPurgeStatusPriorityMap{
	PRIVACY_STATUS_PURGE_FAILED:  3,
	PRIVACY_STATUS_PURGE_RUNNING: 2,
	PRIVACY_STATUS_PURGE_INVOKED: 1,
	PRIVACY_STATUS_PURGE_SUCCESS: 0,
}

type UpdateMode uint8

const (
	MAIN_RECORD UpdateMode = 1 << iota
	LOCATION_RECORD
)

type UpdatePurgeStatusOptions struct {
	UpdateMode   UpdateMode
	LocationType LocationType
	ErrorMessage string
}
