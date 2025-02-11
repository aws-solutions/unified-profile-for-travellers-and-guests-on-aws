// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudwatch

import (
	"encoding/json"
	"errors"
	"log"
	"tah/upt/source/tah-core/core"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	awsCloudwatchLogs "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type MetricLogger struct {
	Namespace string
	Metrics   []Metric
}

// this go struct represent the metrics object. once it needs to serlialize in a format th cloudwatch understands for custom metrics
type Metric struct {
	Dimensions map[string]string
	Metadata   EmfMetadata
	Values     map[string]float64
}

func (metric Metric) Serlialize() (string, error) {
	obj := map[string]interface{}{}
	for key, value := range metric.Dimensions {
		//dimentionName  : dimentionValue
		obj[key] = value
	}
	for key, value := range metric.Values {
		//metricName  : metricValue
		obj[key] = value
	}
	obj["_aws"] = metric.Metadata
	json, err := json.Marshal(obj)
	if err != nil {
		log.Printf("Could not serlialize metadata")
	}
	return string(json), err
}

type EmfMetadata struct {
	Timestamp         int64             `json:"Timestamp"`
	CloudwatchMetrics []MetricDirective `json:"CloudWatchMetrics"`
}

type MetricDirective struct {
	Namespace  string             `json:"Namespace"`
	Dimensions [][]string         `json:"Dimensions"`
	Metrics    []MetricDefinition `json:"Metrics"`
}

type MetricDefinition struct {
	Name string      `json:"Name"`
	Unit METRIC_UNIT `json:"Unit,omitempty"`
}

type METRIC_UNIT string

const (
	Seconds            METRIC_UNIT = "Seconds"
	Microseconds       METRIC_UNIT = "Microseconds"
	Milliseconds       METRIC_UNIT = "Milliseconds"
	Bytes              METRIC_UNIT = "Bytes"
	Kilobytes          METRIC_UNIT = "Kilobytes"
	Megabytes          METRIC_UNIT = "Megabytes"
	Gigabytes          METRIC_UNIT = "Gigabytes"
	Terabytes          METRIC_UNIT = "Terabytes"
	Bits               METRIC_UNIT = "Bits"
	Kilobits           METRIC_UNIT = "Kilobits"
	Megabits           METRIC_UNIT = "Megabits"
	Gigabits           METRIC_UNIT = "Gigabits"
	Terabits           METRIC_UNIT = "Terabits"
	Percent            METRIC_UNIT = "Percent"
	Count              METRIC_UNIT = "Count"
	BytesPerSecond     METRIC_UNIT = "Bytes/Second"
	KilobytesPerSecond METRIC_UNIT = "Kilobytes/Second"
	MegabytesPerSecond METRIC_UNIT = "Megabytes/Second"
	GigabytesPerSecond METRIC_UNIT = "Gigabytes/Second"
	TerabytesPerSecond METRIC_UNIT = "Terabytes/Second"
	BitsPerSecond      METRIC_UNIT = "Bits/Second"
	KilobitsPerSecond  METRIC_UNIT = "Kilobits/Second"
	MegabitsPerSecond  METRIC_UNIT = "Megabits/Second"
	GigabitsPerSecond  METRIC_UNIT = "Gigabits/Second"
	TerabitsPerSecond  METRIC_UNIT = "Terabits/Second"
	CountPerSecond     METRIC_UNIT = "Count/Second"
	None               METRIC_UNIT = "None"
)

type MetricData struct {
	NameSpace  string
	Data       []float64
	ID         string
	Timestamps []time.Time
}

type CloudwatchConfig struct {
	Client          *awsCloudwatch.CloudWatch
	LogClient       *awsCloudwatchLogs.CloudWatchLogs
	SolutionId      string
	SolutionVersion string
	Region          string
}

func Init(region, solutionId, solutionVersion string) CloudwatchConfig {
	mySession := session.Must(session.NewSession())
	client := core.CreateClient(solutionId, solutionVersion)
	// Set max retries to 0. This disables any default retry functionality provided
	// by the AWS SDK. https://pkg.go.dev/github.com/aws/aws-sdk-go/aws/request#Retryer
	// This avoid issues with patterns like DeleteIfExists, where a resource
	// that does not exist fails, then is created, finally the failed DeleteIfExists
	// call is retried and unintentionally deletes the new resource
	cfg := aws.NewConfig().WithRegion(region).WithHTTPClient(client).WithMaxRetries(0)
	svc := awsCloudwatch.New(mySession, cfg)
	svc2 := awsCloudwatchLogs.New(mySession, cfg)
	return CloudwatchConfig{
		Client:          svc,
		LogClient:       svc2,
		SolutionId:      solutionId,
		SolutionVersion: solutionVersion,
		Region:          region,
	}
}

/********************
* standard SDK functions
*********************/

func (c CloudwatchConfig) GetMetricData(start, end time.Time, namespace, name string, stat string, dimentions map[string]string) ([]MetricData, error) {
	dims := []*awsCloudwatch.Dimension{}
	for k, v := range dimentions {
		dims = append(dims, &awsCloudwatch.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		})

	}
	in := &awsCloudwatch.GetMetricDataInput{
		StartTime: aws.Time(start),
		EndTime:   aws.Time(end),
		MetricDataQueries: []*awsCloudwatch.MetricDataQuery{
			{
				Id: aws.String("customeMetric" + core.GenerateUniqueId()),
				MetricStat: &awsCloudwatch.MetricStat{
					Period: aws.Int64(60),
					Metric: &awsCloudwatch.Metric{
						Namespace:  aws.String(namespace),
						MetricName: aws.String(name),
						Dimensions: dims,
					},
					Stat: aws.String(stat),
				},
			},
		},
	}
	log.Printf("Metric data request: %+v", in)
	out, err := c.Client.GetMetricData(in)
	if err != nil {
		log.Printf("Error getting metric data: %s", err)
		return []MetricData{}, err
	}
	log.Printf("Metric data response: %+v", out)

	data := []MetricData{}
	for _, res := range out.MetricDataResults {
		data = append(data, MetricData{
			Data:       aws.Float64ValueSlice(res.Values),
			Timestamps: aws.TimeValueSlice(res.Timestamps),
			ID:         *res.Id,
		})
	}
	return data, nil
}

// create log group
func (c CloudwatchConfig) CreateLogGroup(logGroupName string) error {
	log.Printf("[cloudwatch] creating log group %s", logGroupName)
	in := &awsCloudwatchLogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	}
	_, err := c.LogClient.CreateLogGroup(in)
	if err != nil {
		log.Printf("[cloudwatch] error creating log group %s: %s", logGroupName, err)
	}
	return err
}

// create log stream
func (c CloudwatchConfig) CreateLogStream(logGroupName, logStreamName string) error {
	log.Printf("[cloudwatch] creating log stream %s/%s", logGroupName, logStreamName)
	in := &awsCloudwatchLogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	}
	_, err := c.LogClient.CreateLogStream(in)
	if err != nil {
		log.Printf("[cloudwatch] error creating log stream %s/%s: %s", logGroupName, logStreamName, err)
	}
	return err
}

// delete log group
func (c CloudwatchConfig) DeleteLogGroup(logGroupName string) error {
	log.Printf("Deleting log group %v", logGroupName)
	in := &awsCloudwatchLogs.DeleteLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	}
	_, err := c.LogClient.DeleteLogGroup(in)
	if err != nil {
		log.Printf("Error deleting log group %v: %v", logGroupName, err)
	}
	return err
}

// delete log stream
func (c CloudwatchConfig) DeleteLogStream(logGroupName, logStreamName string) error {
	log.Printf("Deleting log stream %v/%v", logGroupName, logStreamName)
	in := &awsCloudwatchLogs.DeleteLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	}
	_, err := c.LogClient.DeleteLogStream(in)
	if err != nil {
		log.Printf("Error deleting log stream %v/%v: %v", logGroupName, logStreamName, err)
	}
	return err
}

// mostly used for testing for now. do not use in prodction
func (c CloudwatchConfig) SendLog(logroup, logStream, logLine string) error {
	in := &awsCloudwatchLogs.PutLogEventsInput{
		LogGroupName:  aws.String(logroup),
		LogStreamName: aws.String(logStream),
		LogEvents: []*awsCloudwatchLogs.InputLogEvent{
			{
				Message:   aws.String(logLine),
				Timestamp: aws.Int64(time.Now().UnixMilli()),
			},
		},
	}
	log.Printf("[cloudwatch] PutLogEvents: %+v", in)
	_, err := c.LogClient.PutLogEvents(in)
	if err != nil {
		log.Printf("[cloudwatch] error sending log %s: %s", logLine, err)
	}
	return err
}

// to use for testing
func (c CloudwatchConfig) WaitForMetrics(start, end time.Time, namespace, name, metricStat string, dimentions map[string]string, timeoutSeconds int) ([]MetricData, error) {
	log.Printf("[cloudwatch] Waiting for cloudwatch custom metric %s/%s to be created [%v-%v] with timeout %v sec", namespace, name, start, end, timeoutSeconds)
	max := timeoutSeconds / 5
	for i := 0; i < max; i++ {
		log.Printf("[cloudwatch] getting cloudwatch data (try %v)", i)
		data, err := c.GetMetricData(start, end, namespace, name, metricStat, dimentions)
		if err != nil {
			log.Printf("[cloudwatch] Error retrieving cloudwatch metric %s", err)
			return []MetricData{}, err
		}
		if len(data) > 0 && len(data[0].Data) > 0 {
			log.Printf("[cloudwatch] Successfully retrieved some cloudwatch metric data %v", data)
			return data, nil
		}
		log.Printf("[cloudwatch] no metrics avaialble yet, waiting 5 seconds")
		time.Sleep(5 * time.Second)
	}
	return []MetricData{}, errors.New("[cloudwatch] timeout waiting for custom metric")
}

/********************
* Metric Loggers
*********************/

func NewMetricLogger(namespace string) *MetricLogger {
	log.Printf("Initializing new logger with namespace %v", namespace)
	logger := MetricLogger{
		Namespace: namespace,
	}
	return &logger
}

func (logger *MetricLogger) init(namespace string) {
	logger.Namespace = namespace
}

func (logger *MetricLogger) SetNamespace(namespaceName string) {
	logger.Namespace = namespaceName
}

func (logger *MetricLogger) AddMetric(dims map[string]string, name string, unit METRIC_UNIT, value float64) error {
	log.Printf("Adding metric %v to logger", name)
	if logger.Namespace == "" {
		log.Printf("namespace not set on logger")
		return errors.New("namespace not set on logger")
	}
	dimNames := []string{}
	for dimName := range dims {
		if dimName == "_aws" {
			log.Printf("cannot set _aws as dimension name. reserved key")
			return errors.New("cannot set _aws as dimension name. reserved key")
		}
		dimNames = append(dimNames, dimName)
	}
	logger.Metrics = append(logger.Metrics, Metric{
		Dimensions: dims,
		Metadata: EmfMetadata{
			Timestamp: time.Now().UnixMilli(),
			CloudwatchMetrics: []MetricDirective{
				{
					Namespace:  logger.Namespace,
					Dimensions: [][]string{dimNames},
					Metrics: []MetricDefinition{
						{
							Name: name,
							Unit: unit,
						},
					},
				},
			},
		},
		Values: map[string]float64{name: value},
	})
	return nil
}

// this function logs the metric directly without aggregating in the logger first it is stateless
func (logger *MetricLogger) LogMetric(dims map[string]string, name string, unit METRIC_UNIT, value float64) {
	log.Printf("logging metric directly %v with value %v", name, value)
	//recreate the logger to make sure we don't have persisting data from previous use
	newLogger := NewMetricLogger(logger.Namespace)
	err := newLogger.AddMetric(dims, name, unit, value)
	if err != nil {
		log.Printf("Error adding metric to logger: %s", err)
		return
	}
	newLogger.Log()
}

func (logger *MetricLogger) Log() {
	lines := logger.createLogLines()
	log.Printf("Logging %v lines for customer metric", len(lines))
	for _, logLine := range lines {
		log.Printf(logLine)
	}
}

// stateless logger version

func (logger *MetricLogger) createLogLines() []string {
	logs := []string{}
	for _, metric := range logger.Metrics {
		val, err := metric.Serlialize()
		if err != nil {
			log.Printf("Error serializing metric: %s", err)
			continue
		}
		logs = append(logs, val)
	}
	return logs
}
