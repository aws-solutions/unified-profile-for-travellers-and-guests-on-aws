// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofiles

import (
	"time"
)

/********************************************************
* Setup
*********************************************************/

type Domain struct {
	Name                 string
	NObjects             int64
	NProfiles            int64
	MatchingEnabled      bool
	RuleBaseStitchingOn  bool
	IsLowCostEnabled     bool
	DefaultEncryptionKey string
	Created              time.Time
	LastUpdated          time.Time
	Tags                 map[string]string
}

type IngestionError struct {
	Reason  string
	Message string
}

type MatchList struct {
	ConfidenceScore float64
	ProfileIds      []string
}
type ProfilePair struct {
	SourceID string
	TargetID string
}

type ObjectMapping struct {
	Name    string         `json:"name"`
	Version string         `json:"version"`
	Fields  []FieldMapping `json:"fields"`
}

type FieldMapping struct {
	Type        string   `json:"type"`
	Source      string   `json:"source"`
	Target      string   `json:"target"`
	Indexes     []string `json:"indexes"`
	Searcheable bool     `json:"searchable"`
	KeyOnly     bool     `json:"keyOnly"`
}

type FieldMappings []FieldMapping

type Integration struct {
	Name           string
	Source         string
	Target         string
	Status         string
	StatusMessage  string
	LastRun        time.Time
	LastRunStatus  string
	LastRunMessage string
	Trigger        string
}

type PaginationOptions struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	ObjectType string `json:"objectType"`
}

type DeleteProfileOptions struct {
	DeleteHistory bool `json:"deleteProfileHistoryOnDelete"`
}

/************************************************************
* Control Plane
*******************************************************/

// user-defined options
type DomainOptions struct {
	AiIdResolutionOn       bool     `json:"aiIdResolutionOn"`
	RuleBaseIdResolutionOn bool     `json:"ruleBaseIdResolutionOn"` //true is rule based id resolution enabled
	ObjectTypePriority     []string `json:"objectTypePriority"`     //optionally define the relative priority between object type, this will define how profile-level data is updated
	DynamoMode             *bool    `json:"dynamoMode"`
	CustomerProfileMode    *bool    `json:"cusstomerProfileMode"`
}

// supported data types field mappings
const (
	MappingTypeString  string = "STRING"
	MappingTypeText    string = "TEXT"
	MappingTypeInteger string = "INTEGER"
)
