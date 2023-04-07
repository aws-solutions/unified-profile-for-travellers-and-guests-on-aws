package awssolutions

import (
	"encoding/json"
	"log"
	"strings"
	core "tah/core/core"
	"time"
)

var METRICS_ENDPOINT = "https://metrics.awssolutionsbuilder.com/generic"
var TIMESTAMP_FORMAT = "2006-02-01 15:04:05.0"
var ANONYMOUS_USAGE_VALUE_YES = "yes"

type IConfig interface {
	SendMetrics(data map[string]interface{}) (bool, error)
}

type Config struct {
	Svc            core.HttpService
	Solution       string
	Version        string
	MetricsEnabled bool
	DeploymentUUID string
}

type MetricRequest struct {
	Solution  string
	Version   string
	TimeStamp string
	UUID      string
	Data      map[string]interface{}
}

type MetricResponse struct {
}

func (p MetricResponse) Decode(dec json.Decoder) (error, core.JSONObject) {
	return dec.Decode(&p), p
}

func Init(solutionID string, version string, deploymentUUID string, anonymousUsage string) Config {
	return Config{
		Svc:            core.HttpInit(METRICS_ENDPOINT),
		Solution:       solutionID,
		Version:        version,
		DeploymentUUID: deploymentUUID,
		MetricsEnabled: (strings.ToLower(anonymousUsage) == ANONYMOUS_USAGE_VALUE_YES),
	}
}

func (c Config) SendMetrics(data map[string]interface{}) (bool, error) {
	if !c.MetricsEnabled {
		log.Printf("Anonymous usage data is disabled. Metrics not sent")
		return false, nil
	}
	rq := MetricRequest{
		Solution:  c.Solution,
		Version:   c.Version,
		UUID:      c.DeploymentUUID,
		TimeStamp: time.Now().Format(TIMESTAMP_FORMAT),
		Data:      data,
	}
	options := core.RestOptions{}
	rs := MetricResponse{}
	res, err := c.Svc.HttpPost(rq, rs, options)
	if err != nil {
		log.Printf("Error sending metric %+v", rs)
		return false, err
	}
	rs = res.(MetricResponse)
	log.Printf("Metrics Service response: %+v", rs)
	return true, nil
}
