package lodging

import (
	"time"
)

var ORIGINATING_SYSTEN_PMS = "pms"
var ORIGINATING_SYSTEN_CRS = "crs"
var ORIGINATING_SYSTEN_OTA = "ota"
var ORIGINATING_SYSTEN_GDS = "gds"

type DataSearchRq struct {
	HotelCode         string
	ChainCode         string
	LastUpdatedAfter  time.Time
	LastUpdatedBefore time.Time
	Page              int64
	PageSize          int64
}

type DataImportRq struct {
	DataType         string `json:"dataType"`
	NJobs            int    `json:"nJobs"`
	FrequencyMinutes int    `json:"frequencyMinutes"`
}

type ImportStatus struct {
	Status               string        `json:"status"`
	NJobs                int64         `json:"nJobs"`
	NJobsInProgress      int64         `json:"nJobsInProgress"`
	MaxDuration          int64         `json:"maxDuration"`
	PropertyFile         string        `json:"propertyFile"`
	Imported             int64         `json:"imported"`
	Total                int64         `json:"total"`
	Progress             float64       `json:"progress"`
	LastUpdated          time.Time     `json:"lastUpdated"`
	NProperties          int64         `json:"nProperties"`
	NPropertiesCompleted int64         `json:"nPropertiesCompleted"`
	NPropertiesRemaning  int64         `json:"nPropertiesRemaning"`
	Schedule             string        `json:"schedule"`
	SchedulerStatus      string        `json:"schedulerStatus"`
	TotalErrors          int64         `json:"totalErrors"`
	Errors               []ImportError `json:"errors"`
}

type FeedStatus struct {
	Status       string    `json:"status"`
	PropertyCode string    `json:"propertyCode"`
	DataType     string    `json:"dataType"`
	LastUpdated  time.Time `json:"lastUpdated"`
}

type PropertyImportStatus struct {
	Imported    int64     `json:"imported"`
	Total       int64     `json:"total"`
	Progress    float64   `json:"progress"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type ImportError struct {
	Timestamp time.Time `json:"timestamp"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Request   string    `json:"request"`
}
