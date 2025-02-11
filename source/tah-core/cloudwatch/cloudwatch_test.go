// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudwatch

import (
	"log"
	"math/rand"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

func TestMetricsLogger(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	cw := Init(envCfg.Region, "sid", "sVersion")
	metricsNS := "tah-core-test"
	metricName := "tah-core-test-metrics" + core.GenerateUniqueId()
	logGroup := "tah-core-test-" + core.GenerateUniqueId()
	logStream := "tah-core-test-stream" + core.GenerateUniqueId()

	err = cw.CreateLogGroup(logGroup)
	if err != nil {
		t.Errorf("Error creating log group: %v", err)
	}
	err = cw.CreateLogStream(logGroup, logStream)
	if err != nil {
		t.Errorf("Error creating log stream: %v", err)
	}

	start := time.Now().Add(-5 * time.Minute)
	logger := NewMetricLogger(metricsNS)
	logger.AddMetric(map[string]string{"LambdaName": "ucpChangeProcessor", "Region": "us-east-1"}, metricName, "Milliseconds", float64(100*rand.Intn(10)))

	logLines := logger.createLogLines()
	if len(logLines) == 0 {
		t.Errorf("Error creating log values")
	} else {
		err = cw.SendLog(logGroup, logStream, logLines[0])
		if err != nil {
			t.Errorf("Error sending log: %v", err)
		}
		//add 1 minute to search window
		end := time.Now().Add(time.Minute * 5)
		log.Printf("searching between %v and %v", start.UnixMilli(), end.UnixMilli())
		data, err := cw.WaitForMetrics(start, end, metricsNS, metricName, "Average", map[string]string{
			"LambdaName": "ucpChangeProcessor",
			"Region":     "us-east-1",
		}, 300)
		if err != nil {
			t.Errorf("Error waiting for metrics: %v", err)
		} else {
			log.Printf("Metrics: %+v", data)

		}
	}

}

func TestMetricsLoggerWithNamespace(t *testing.T) {
	logger := NewMetricLogger("test/namespace")
	err := logger.AddMetric(map[string]string{"LambdaName": "ucpChangeProcessor", "Region": "us-east-1"}, "test/metric", "Milliseconds", float64(100*rand.Intn(10)))
	if err != nil {
		t.Errorf("Error adding metric: %v", err)
	}
	logger.Log()

}
