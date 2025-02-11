// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package changeEvent

import profilemodel "tah/upt/source/tah-core/customerprofilesmodel"

type ProfileChangeRecord struct {
	SchemaVersion       int64  `json:"SchemaVersion"`
	EventId             string `json:"EventId"`
	EventTimestamp      string `json:"EventTimestamp"`
	EventType           string `json:"EventType"`
	DomainName          string `json:"DomainName"`
	ObjectTypeName      string `json:"ObjectTypeName"`
	AssociatedProfileId string `json:"AssociatedProfileId"`
	IsMessageRealTime   bool   `json:"IsMessageRealTime"`
}

type ProfileChangeRecordWithProfile struct {
	SchemaVersion       int64                  `json:"SchemaVersion"`
	EventId             string                 `json:"EventId"`
	EventTimestamp      string                 `json:"EventTimestamp"`
	EventType           string                 `json:"EventType"`
	DomainName          string                 `json:"DomainName"`
	ObjectTypeName      string                 `json:"ObjectTypeName"`
	AssociatedProfileId string                 `json:"AssociatedProfileId"`
	Object              profilemodel.Profile   `json:"Object"`
	MergeId             string                 `json:"MergeId"`           // only included in merge/unmerge events
	DuplicateProfiles   []profilemodel.Profile `json:"DuplicateProfiles"` // only included in merge/unmerge events
	IsMessageRealTime   bool                   `json:"IsMessageRealTime"`
}

type ProfileChangeRecordWithProfileObject struct {
	SchemaVersion       int64                  `json:"SchemaVersion"`
	EventId             string                 `json:"EventId"`
	EventTimestamp      string                 `json:"EventTimestamp"`
	EventType           string                 `json:"EventType"`
	DomainName          string                 `json:"DomainName"`
	ObjectTypeName      string                 `json:"ObjectTypeName"`
	AssociatedProfileId string                 `json:"AssociatedProfileId"`
	Object              map[string]interface{} `json:"Object"`
	IsMessageRealTime   bool                   `json:"IsMessageRealTime"`
}
