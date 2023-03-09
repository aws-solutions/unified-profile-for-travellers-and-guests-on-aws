package model

import (
	core "tah/core/core"
	"tah/ucp/src/business-logic/common"
	"time"
)

type SearchRq struct {
	Email     string
	LastName  string
	Phone     string
	LoyaltyID string
}

type LinkIndustryConnectorRq struct {
	AgwUrl        string
	TokenEndpoint string
	ClientId      string
	ClientSecret  string
	BucketArn     string
}

type LinkIndustryConnectorRes struct {
	GlueRoleArn  string
	BucketPolicy string
}

type CreateConnectorCrawlerRq struct {
	GlueRoleArn string
	BucketPath  string
	ConnectorId string
}

type UCPRequest struct {
	ID         string
	Cx         *common.Context
	SearchRq   SearchRq
	Domain     Domain
	EnvName    string
	Pagination PaginationOptions `json:"pagination"`
}

type UCPConfig struct {
	Domains []Domain `json:"domains"`
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

type ResWrapper struct {
	Profiles        []Traveller       `json:"profiles"`
	IngestionErrors []IngestionErrors `json:"ingestionErrors"`
	DataValidation  []ValidationError `json:"dataValidation"`
	TotalErrors     int64             `json:"totalErrors"`
	UCPConfig       UCPConfig         `json:"config"`
	Matches         []Match           `json:"matches"`
	Error           core.ResError     `json:"error"`
	Pagination      PaginationOptions `json:"pagination"`
}

var PAGINATION_OPTION_PAGE = "page"
var PAGINATION_OPTION_PAGE_SIZE = "pageSize"

type PaginationOptions struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type Match struct {
	ConfidenceScore float64 `json:"confidence"`
	ID              string  `json:"id"`
	FirstName       string  `json:"firstName"`
	LastName        string  `json:"lastName"`
	BirthDate       string  `json:"birthDate"`
	PhoneNumber     string  `json:"phone"`
	EmailAddress    string  `json:"email"`
}

type IngestionErrors struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type ObjectMapping struct {
	Name      string         `json:"name"`
	Timestamp string         `json:"timestamp"`
	Fields    []FieldMapping `json:"fields"`
}

type FieldMapping struct {
	Type        string   `json:"type"`
	Source      string   `json:"source"`
	Target      string   `json:"target"`
	Indexes     []string `json:"indexes"`
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

var ERR_TYPE_MISSING_MAPPING_FIELD = "missing_column_mapping"
var ERR_TYPE_MISSING_INDEX_FIELD = "missing_index_field"
var ERR_TYPE_MISSING_MANDATORY_FIELD = "missing_column_mandatory_field"
var ERR_TYPE_MISSING_MANDATORY_FIELD_VALUE = "missing_column_mandatory_field_value"
var ERR_TYPE_MISSING_INDEX_FIELD_VALUE = "missing_index_field_value"
var ERR_TYPE_NO_HEADER = "missing_header"

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
