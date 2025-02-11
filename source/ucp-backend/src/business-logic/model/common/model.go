// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"encoding/json"
	customerprofiles "tah/upt/source/storage"
	core "tah/upt/source/tah-core/core"
	adminModel "tah/upt/source/ucp-common/src/model/admin"
	asyncModel "tah/upt/source/ucp-common/src/model/async"
	privacySearchResultModel "tah/upt/source/ucp-common/src/model/async/usecase"
	travellerModel "tah/upt/source/ucp-common/src/model/traveler"

	"time"
)

type SearchRq struct {
	TravellerID    string
	AirBookingD    string
	HotelBookingID string
	Email          string
	LastName       string
	Phone          string
	Address1       string
	Address2       string
	AddressType    string
	City           string
	State          string
	Country        string
	Province       string
	PostalCode     string
	LoyaltyID      string
	Matches        string
}

type LinkIndustryConnectorRq struct {
	DomainName string `json:"domainName"`
	RemoveLink bool   `json:"removeLink"`
}

type LinkIndustryConnectorRes struct {
	GlueRoleArn  string
	BucketPolicy string
}

type RequestWrapper struct {
	ID                      string
	TxID                    string
	SearchRq                SearchRq
	StartJobRq              StartJobRq `json:"startJobRq"`
	Domain                  Domain
	EnvName                 string
	MergeRq                 []adminModel.MergeRq           `json:"mergeRq"`
	UnmergeRq               adminModel.UnmergeRq           `json:"unmergeRq"`
	LinkIndustryConnectorRq LinkIndustryConnectorRq        `json:"linkIndustryConnectorRq"`
	UcpErrorToDelete        UcpIngestionErrorToDelete      `json:"ucpErrorToDelete"`
	Pagination              []adminModel.PaginationOptions `json:"pagination"`
	PortalConfig            PortalConfig                   `json:"portalConfig"`
	SaveRuleSet             SaveRuleSetRq                  `json:"saveRuleSetRq"`
	ListRuleSets            ListRuleSetsRq                 `json:"listRuleSetsRq"`
	CreatePrivacySearchRq   []CreatePrivacySearchRq        `json:"privacySearchRq"`
	ListPrivacySearches     ListPrivacySearchesRq          `json:"listPrivacySearchesRq"`
	GetPrivacySearch        GetPrivacySearchRq             `json:"getPrivacySearchRq"`
	DeletePrivacySearches   DeletePrivacySearchesRq        `json:"deletePrivacySearchesRq"`
	PrivacyPurge            CreatePrivacyPurgeRq           `json:"createPrivacyPurgeRq"`
	GetInteractionHistory   GetInteractionHistoryRq        `json:"getInteractionHistoryRq"`
	GetPurgeStatus          GetPurgeStatusRq               `json:"getPurgeStatusRq"`
	DomainSetting           DomainSetting                  `json:"domainSetting"`
	RebuildCache            RebuildCacheRq                 `json:"rebuildCacheRq"`
}

type StartJobRq struct {
	JobName    string `json:"jobName"`
	DomainName string `json:"domainName"`
}

type CreatePrivacySearchRq struct {
	ConnectId string `json:"connectId"`
}

type ListPrivacySearchesRq struct {
	DomainName string `json:"domainName"`
}

type GetPrivacySearchRq struct {
	ConnectId  string `json:"connectId"`
	DomainName string `json:"domainName"`
}

type DeletePrivacySearchesRq struct {
	ConnectIds []string `json:"connectIds"`
	DomainName string   `json:"domainName"`
}

type CreatePrivacyPurgeRq struct {
	ConnectIds     []string `json:"connectIds"`
	AgentCognitoId string   `json:"agentCognitoId"`
}

type GetPurgeStatusRq struct {
	DomainName string `json:"domainName"`
}

type RebuildCacheRq struct {
	CacheMode customerprofiles.CacheMode `json:"cacheMode"`
}

type GetInteractionHistoryRq struct {
	InteractionId string `json:"interactionId"`
	ObjectType    string `json:"objectType"`
	ConnectId     string `json:"connectId"`
	DomainName    string `json:"domainName"`
}

type ResponseWrapper struct {
	TxID                 string
	Profiles             *[]travellerModel.Traveller                     `json:"profiles,omitempty"`
	IngestionErrors      *[]UcpIngestionError                            `json:"ingestionErrors,omitempty"`
	DataValidation       *[]ValidationError                              `json:"dataValidation,omitempty"`
	TotalErrors          *int64                                          `json:"totalErrors,omitempty"`
	TotalMatchPairs      *int64                                          `json:"totalMatchPairs,omitempty"`
	UCPConfig            *UCPConfig                                      `json:"config,omitempty"`
	Matches              *[]travellerModel.Match                         `json:"matches,omitempty"`
	MatchPairs           *[]adminModel.MatchPair                         `json:"matchPairs,omitempty"`
	Error                core.ResError                                   `json:"error"`
	Pagination           *[]adminModel.PaginationOptions                 `json:"pagination,omitempty"`
	PaginationMetadata   *PaginationMetadata                             `json:"paginationMetadata,omitempty"`
	Connectors           *[]Connector                                    `json:"connectors,omitempty"`
	AwsResources         AwsResources                                    `json:"awsResources,omitempty"`
	MergeResponse        *string                                         `json:"mergeResponse,omitempty"`
	PortalConfig         *PortalConfig                                   `json:"portalConfig,omitempty"`
	DomainSetting        *DomainSetting                                  `json:"domainSetting,omitempty"`
	RuleSets             *[]customerprofiles.RuleSet                     `json:"ruleSets,omitempty"`
	ProfileMappings      *[]string                                       `json:"profileMappings,omitempty"`
	AccpRecords          []adminModel.AccpRecord                         `json:"accpRecords,omitempty"`
	AsyncEvent           *asyncModel.AsyncEvent                          `json:"asyncEvent,omitempty"`
	PrivacySearchResults *[]privacySearchResultModel.PrivacySearchResult `json:"privacySearchResults,omitempty"`
	PrivacySearchResult  *privacySearchResultModel.PrivacySearchResult   `json:"privacySearchResult,omitempty"`
	PrivacyPurgeStatus   *privacySearchResultModel.PurgeStatusResult     `json:"privacyPurgeStatus,omitempty"`
	InteractionHistory   *[]customerprofiles.ProfileMergeContext         `json:"interactionHistory,omitempty"`
	ProfileSummary       *string                                         `json:"profileSummary,omitempty"`
}

// this is needed to for the UPT SDK
func (p ResponseWrapper) Decode(dec json.Decoder) (error, core.JSONObject) {
	return dec.Decode(&p), p
}

type PortalConfig struct {
	HyperlinkMappings []HyperlinkMapping `json:"hyperlinkMappings"`
}

type HyperlinkMapping struct {
	AccpObject string `json:"accpObject"`
	FieldName  string `json:"fieldName"`
	Template   string `json:"hyperlinkTemplate"`
}

type DynamoHyperlinkMapping struct {
	Pk         string `json:"config_item"`
	Sk         string `json:"config_item_category"`
	AccpObject string `json:"accpObject"`
	FieldName  string `json:"fieldName"`
	Template   string `json:"hyperlinkTemplate"`
}
type DynamoHyperlinkMappingForDelete struct {
	Pk string `json:"config_item"`
	Sk string `json:"config_item_category"`
}

type DynamoDomainConfig struct {
	Pk       string `json:"config_item"`
	Sk       string `json:"config_item_category"`
	Value    string `json:"value"`
	IsActive bool   `json:"isActive"`
}

type SaveRuleSetRq struct {
	Rules []customerprofiles.Rule `json:"rules"`
}

type ListRuleSetsRq struct {
	IncludeHistorical bool `json:"includeHistorical"`
}

type AwsResources struct {
	GlueRoleArn              string                  `json:"glueRoleArn,omitempty"`
	TahConnectorBucketPolicy string                  `json:"tahConnectorBucketPolicy,omitempty"`
	S3Buckets                map[string]string       `json:"S3Buckets,omitempty"`
	Jobs                     []adminModel.JobSummary `json:"jobs,omitempty"`
}

type Connector struct {
	Name          string   `json:"name"`
	Status        string   `json:"status"`
	LinkedDomains []string `json:"linkedDomains"`
}

// Connector Status options
const (
	ConnectorStatus_DeployedWithBucket    string = "deployed_with_bucket"
	ConnectorStatus_DeployedWithoutBucket string = "deployed_without_bucket"
)

type UCPConfig struct {
	Domains []Domain `json:"domains"`
}

type Domain struct {
	Name               string                     `json:"customerProfileDomain"`
	NObjects           int64                      `json:"numberOfObjects"`
	NProfiles          int64                      `json:"numberOfProfiles"`
	MatchingEnabled    bool                       `json:"matchingEnabled"`
	IsLowCostEnabled   bool                       `json:"isLowCostEnabled"`
	CacheMode          customerprofiles.CacheMode `json:"cacheMode"`
	Created            time.Time                  `json:"created"`
	LastUpdated        time.Time                  `json:"lastUpdated"`
	Mappings           []adminModel.ObjectMapping `json:"mappings"`
	Integrations       []Integration              `json:"integrations"`
	NeedsMappingUpdate bool                       `json:"needsMappingUpdate"` //true when new mapping exists in the solution and the domain needs update
}

var PAGINATION_OPTION_PAGE = "page"
var PAGINATION_OPTION_PAGE_SIZE = "pageSize"

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
	Timestamp          time.Time `json:"timestamp"`
	TransactionID      string    `json:"transactionId"`
}
type UcpIngestionErrorToDelete struct {
	Type string `json:"error_type"`
	ID   string `json:"error_id"`
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
	FlowName       string    `json:"flowName"`
}

var ERR_TYPE_MISSING_MAPPING_FIELD = "missing_column_mapping"
var ERR_TYPE_MISSING_INDEX_FIELD = "missing_index_field"
var ERR_TYPE_MISSING_MANDATORY_FIELD = "missing_column_mandatory_field"
var ERR_TYPE_MISSING_MANDATORY_FIELD_VALUE = "missing_column_mandatory_field_value"
var ERR_TYPE_MISSING_INDEX_FIELD_VALUE = "missing_index_field_value"
var ERR_TYPE_NO_HEADER = "missing_header"
var ERR_TYPE_ACCESSS_ERROR = "no_object_access"

type ValidationError struct {
	ErrType string `json:"errType"`
	Msg     string `json:"msg"`
	Object  string `json:"object"`
	File    string `json:"file"`
	Row     int    `json:"row"`
	Col     int    `json:"col"`
	ColName string `json:"field"`
	Bucket  string `json:"bucket"`
}

type Metadata struct {
	TotalRecords int `json:"total_records"`
}

type PaginationMetadata struct {
	AirBookingRecords                 Metadata `json:"airBookingRecords"`
	AirLoyaltyRecords                 Metadata `json:"airLoyaltyRecords"`
	ClickstreamRecords                Metadata `json:"clickstreamRecords"`
	EmailHistoryRecords               Metadata `json:"emailHistoryRecords"`
	HotelBookingRecords               Metadata `json:"hotelBookingRecords"`
	HotelLoyaltyRecords               Metadata `json:"hotelLoyaltyRecords"`
	HotelStayRecords                  Metadata `json:"hotelStayRecords"`
	PhoneHistoryRecords               Metadata `json:"phoneHistoryRecords"`
	CustomerServiceInteractionRecords Metadata `json:"customerServiceInteractionRecords"`
	LoyaltyTxRecords                  Metadata `json:"lotaltyTxRecords"`
	AncillaryServiceRecords           Metadata `json:"ancillaryServiceRecords"`
}

type DomainConfig struct {
	Value    string `json:"value"`
	IsActive bool   `json:"isActive"`
}

type DomainSetting struct {
	PromptConfig DomainConfig `json:"promptConfig"`
	MatchConfig  DomainConfig `json:"matchConfig"`
}
