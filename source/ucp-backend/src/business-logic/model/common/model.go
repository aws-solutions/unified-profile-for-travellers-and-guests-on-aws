package model

import (
	core "tah/core/core"
	travellerModel "tah/ucp/src/business-logic/model/traveller"

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

type RequestWrapper struct {
	ID                       string
	TxID                     string
	SearchRq                 SearchRq
	Domain                   Domain
	EnvName                  string
	LinkIndustryConnectorRq  LinkIndustryConnectorRq  `json:"linkIndustryConnectorRq"`
	CreateConnectorCrawlerRq CreateConnectorCrawlerRq `json:"createConnectorCrawlerRq"`
	Pagination               PaginationOptions        `json:"pagination"`
}

type ResponseWrapper struct {
	TxID            string
	Profiles        []travellerModel.Traveller `json:"profiles"`
	IngestionErrors []UcpIngestionError        `json:"ingestionErrors"`
	DataValidation  []ValidationError          `json:"dataValidation"`
	TotalErrors     int64                      `json:"totalErrors"`
	UCPConfig       UCPConfig                  `json:"config"`
	Matches         []Match                    `json:"matches"`
	Error           core.ResError              `json:"error"`
	Pagination      PaginationOptions          `json:"pagination"`
	Connectors      []Connector                `json:"connectors"`
	AwsResources    AwsResources               `json:"awsResources"`
}

type JobSummary struct {
	JobName     string    `json:"jobName"`
	LastRunTime time.Time `json:"lastRunTime"`
	Status      string    `json:"status"`
}

type AwsResources struct {
	GlueRoleArn              string            `json:"glueRoleArn"`
	TahConnectorBucketPolicy string            `json:"tahConnectorBucketPolicy"`
	S3Buckets                map[string]string `json:"S3Buckets"`
	Jobs                     []JobSummary      `json:"jobs"`
}

type Connector struct {
	Name string `json:"name"`
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

type UcpIngestionError struct {
	Type               string `json:"error_type"`
	ID                 string `json:"error_id"`
	Message            string `json:"message"`
	Domain             string `json:"domain"`
	BusinessObjectType string `json:"businessObjectType"`
	AccpRecordType     string `json:"accpRecordType"`
	Record             string `json:"reccord"`
	TravellerID        string `json:"travelerId"`
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
