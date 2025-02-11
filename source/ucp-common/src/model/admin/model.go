// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"encoding/json"
	customerprofiles "tah/upt/source/storage"
	"time"
)

type BusinessObject struct {
	Name string
}

type AccpRecord struct {
	Name         string   `json:"name,omitempty"`
	ObjectFields []string `json:"objectFields,omitempty"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
// This is necessary because AccpRecord's "Struct" is of type AccpObject, which is
// an interface, and cannot be directly unmarshaled into a struct.
// This custom implementation ignores Struct and preserves Name, which is all that was
// necessary for testing.
func (rec *AccpRecord) UnmarshalJSON(data []byte) error {
	var mapData map[string]interface{}
	if err := json.Unmarshal(data, &mapData); err != nil {
		return err
	}
	rec.Name = mapData["name"].(string)

	return nil
}

type PaginationOptions struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	ObjectType string `json:"objectType"`
}

type Domain struct {
	Name            string          `json:"customerProfileDomain"`
	NObjects        int64           `json:"numberOfObjects"`
	NProfiles       int64           `json:"numberOfProfiles"`
	MatchingEnabled bool            `json:"matchingEnabled"`
	Created         time.Time       `json:"created"`
	LastUpdated     time.Time       `json:"lastUpdated"`
	Mappings        []ObjectMapping `json:"mappings"`
	Integrations    []Integration   `json:"integrations"`
}

type ObjectMapping struct {
	Version   string         `json:"version"`
	Name      string         `json:"name"`
	Timestamp string         `json:"timestamp"`
	Fields    []FieldMapping `json:"fields"`
}

type FieldMapping struct {
	Type        string   `json:"type"`
	Source      string   `json:"source"`
	Target      string   `json:"target"`
	Indexes     []string `json:"indexes"`
	KeyOnly     bool     `json:"keyOnly"`
	Searcheable bool     `json:"searchable"`
	IsTimestamp bool     `json:"isTimestamp"`
}

type Integration struct {
	Source         string    `json:"source"`
	Target         string    `json:"target"`
	Status         string    `json:"status"`
	StatusMessage  string    `json:"statuMessage"`
	LastRun        time.Time `json:"lastRun"`
	LastRunStatus  string    `json:"lastRunStatus"`
	LastRunMessage string    `json:"lastRunMessage"`
	Trigger        string    `json:"trigger"`
}

type IndustryConnectorDomainList struct {
	ItemID     string   `json:"item_id"`
	ItemType   string   `json:"item_type"`
	DomainList []string `json:"domain_list"`
}

type MatchPair struct {
	Pk              string     `json:"domain_sourceProfileId"`
	Sk              string     `json:"match_targetProfileId"`
	TargetProfileID string     `json:"targetProfileID"`
	Domain          string     `json:"domain"`
	Score           string     `json:"score"`
	ProfileDataDiff []DataDiff `json:"profileDataDiff"`
	RunID           string     `json:"runId"`
	ScoreTargetID   string     `json:"scoreTargetId"`
	MergeInProgress bool       `json:"mergeInProgress"`
	FalsePositive   bool       `json:"falsePositive"`
	//indexes
	FirstNameS string `json:"firstNameS"`
	LastNameS  string `json:"lastNameS"`
	FirstNameT string `json:"firstNameT"`
	LastNameT  string `json:"lastNameT"`
}

type MatchPairForDelete struct {
	Pk string `json:"domain_sourceProfileId"`
	Sk string `json:"match_targetProfileId"`
}

type DataDiff struct {
	FieldName   string `json:"fieldName"`
	SourceValue string `json:"sourceValue"`
	TargetValue string `json:"targetValue"`
}

type UcpIngestionError struct {
	Type               string    `json:"error_type"`
	ID                 string    `json:"error_id"`
	Category           string    `json:"category"`
	Message            string    `json:"message"`
	Domain             string    `json:"domain"`
	BusinessObjectType string    `json:"businessObjectType"`
	AccpRecordType     string    `json:"accpRecordType"`
	Record             string    `json:"reccord"`
	TravellerID        string    `json:"travelerId"`
	RecordID           string    `json:"recordId"`
	Timestamp          time.Time `json:"timestamp"`
	TransactionID      string    `json:"transactionId"`
	TTL                int       `json:"ttl"`
}

type MergeRq struct {
	SourceProfileID string                               `json:"source"`
	TargetProfileID string                               `json:"target"`
	FalsePositive   bool                                 `json:"falsePositive"`
	MergeContext    customerprofiles.ProfileMergeContext `json:"mergeContext"`
}

type UnmergeRq struct {
	ToUnmergeConnectID   string                               `json:"toUnmergeConnectID"`
	MergedIntoConnectID  string                               `json:"mergedIntoConnectID"`
	InteractionToUnmerge string                               `json:"interactionToUnmerge"`
	InteractionType      string                               `json:"interactionType"`
	UnmergeContext       customerprofiles.ProfileMergeContext `json:"unmergeContext"`
}

type JobSummary struct {
	JobName     string    `json:"jobName"`
	LastRunTime time.Time `json:"lastRunTime"`
	Status      string    `json:"status"`
}

// Realtime ingestion error categories
// note that these are not the only error categories that may show us as some error categories are dynamically added
const (
	ERROR_PARSING_ERROR  string = "error_parsing_error"
	ACCP_INGESTION_ERROR string = "accp_ingestion_error"
)
