package core

import "strconv"

var MODEL_VERSION = "1.2.0"
var DATETIME_FORMAT = "2006-01-02T15:04:05"
var TIME_FORMAT = "15:04"
var DATE_FORMAT = "2006-01-02"
var QUERY_PARAM_PAGE = "page"
var QUERY_PARAM_PAGE_SIZE = "pageSize"
var QUERY_PARAM_DATA_TYPE = "dataType"

const INGEST_TIMESTAMP_FORMAT = "2006-01-02T15:04:05.000000Z"

type BusinessObject interface {
	Version() string
}

// Authirization request object
type AuthRq struct {
	ApiID  string
	ApiKey string
}
type AuthRs struct {
	Token string
}

///////////////
// Customer serializable types
/////////////

///////////////
// Default float64 marshaller remoovethe trailing decimal when a number is integer. this confusing Glue Crawler type
// inference logic so we force at least one decimal
// See https://stackoverflow.com/questions/52446730/stop-json-marshal-from-stripping-trailing-zero-from-floating-point-number
/////////////

type Float float64

func (f Float) MarshalJSON() ([]byte, error) {
	if float64(f) == float64(int(f)) {
		return []byte(strconv.FormatFloat(float64(f), 'f', 1, 32)), nil
	}
	return []byte(strconv.FormatFloat(float64(f), 'f', -1, 32)), nil
}
