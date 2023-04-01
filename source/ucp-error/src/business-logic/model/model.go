package model

import "time"

var ERROR_PARSING_ERROR = "error_parsing_error"
var ACCP_INGESTION_ERROR = "accp_ingestion_error"

type ResponseWrapper struct {
}

type ACCPIngestionError struct {
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
	Timestamp          time.Time `json:"timestamp"`
}

type AccpRecord struct {
	ModelVersion string `json:"model_version"`
	ObjectType   string `json:"object_type"`
	LastUpdated  string `json:"last_updated"`
	TravellerID  string `json:"traveller_id"`
}
