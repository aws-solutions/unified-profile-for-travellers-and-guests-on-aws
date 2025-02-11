// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package awssolutions

import (
	"fmt"
	core "tah/upt/source/tah-core/core"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

/**********
* IAM
************/
func TestMetricService(t *testing.T) {
	t.Parallel()
	cfg := Init("TEST_TAH_SOLUTION_ID", "1.0", "aaaaaaaa-bbbbbbbb-cccccccc-dddddddd", "Yes", core.LogLevelDebug)
	sent, err := cfg.SendMetrics(map[string]interface{}{"metric1": 6.4, "m2": "test"})
	if err != nil {
		t.Fatalf("[TestMetricService] error sending metrics: %v", err)
	}
	if !sent {
		t.Fatalf("[TestMetricService] metric not sent while enabled %v", err)
	}
	cfg = Init("TEST_TAH_SOLUTION_ID", "1.0", "aaaaaaaa-bbbbbbbb-cccccccc-dddddddd", "yes", core.LogLevelDebug)
	sent, err = cfg.SendMetrics(map[string]interface{}{"metric1": 6.3, "m2": "test"})
	if err != nil {
		t.Fatalf("[TestMetricService] error sending metrics: %v", err)
	}
	if !sent {
		t.Fatalf("[TestMetricService] metric not sent while enabled %v", err)
	}
	cfg = Init("TEST_TAH_SOLUTION_ID", "1.0", "aaaaaaaa-bbbbbbbb-cccccccc-dddddddd", "No", core.LogLevelDebug)
	sent, err = cfg.SendMetrics(map[string]interface{}{"metric1": 6, "m2": "test"})
	if err != nil {
		t.Fatalf("[TestMetricService] error sending metrics: %v", err)
	}
	if sent {
		t.Fatalf("[TestMetricService] metric  sent while disabled %v", err)
	}
}

func TestMetricEnabled(t *testing.T) {
	t.Parallel()
	mockHTTPService := core.HttpInitMock()
	defer mockHTTPService.AssertExpectations(t)
	mockHTTPService.On("HttpPost", mock.Anything, mock.Anything, mock.Anything).Return(MetricResponse{}, nil)

	cfg := Init("TEST_TAH_SOLUTION_ID", "1.0", "aaaaaaaa-bbbbbbbb-cccccccc-dddddddd", SEND_ANONYMIZED_DATA_VALUE_YES, core.LogLevelDebug)
	cfg.Svc = mockHTTPService

	sent, err := cfg.SendMetrics(map[string]interface{}{"metric1": 6.4, "m2": "test"})
	assert.NoError(t, err)
	assert.True(t, sent)
	if err != nil {
		t.Fatalf("[%s] error sending metrics: %v", t.Name(), err)
	}
	if !sent {
		t.Fatalf("[%s] metric not sent while enabled %v", t.Name(), err)
	}
	// Metrics are enabled, HttpPost should be invoked
	mockHTTPService.AssertCalled(t, "HttpPost", mock.Anything, mock.Anything, mock.Anything)
	mockHTTPService.AssertNumberOfCalls(t, "HttpPost", 1)
}

func TestMetricDisabled(t *testing.T) {
	t.Parallel()
	mockHTTPService := core.HttpInitMock()
	defer mockHTTPService.AssertExpectations(t)

	cfg := Init("TEST_TAH_SOLUTION_ID", "1.0", "aaaaaaaa-bbbbbbbb-cccccccc-dddddddd", SEND_ANONYMIZED_DATA_VALUE_NO, core.LogLevelDebug)
	cfg.Svc = mockHTTPService

	// Test metric sending when disabled
	sent, err := cfg.SendMetrics(map[string]interface{}{"metric1": 6, "m2": "test"})
	assert.NoError(t, err)
	assert.False(t, sent)
	if err != nil {
		t.Fatalf("[%s] error sending metrics: %v", t.Name(), err)
	}
	if sent {
		t.Fatalf("[%s] metric sent while disabled %v", t.Name(), err)
	}

	// Ensure that HttpPost was not called again (count should remain 1)
	mockHTTPService.AssertNumberOfCalls(t, "HttpPost", 0)
}

// Purpose: disabling metrics in an existing config should not send metrics after the update
func TestMetricDisableAfterEnabling(t *testing.T) {
	t.Parallel()
	mockHTTPService := core.HttpInitMock()
	defer mockHTTPService.AssertExpectations(t)
	mockHTTPService.On("HttpPost", mock.Anything, mock.Anything, mock.Anything).Return(MetricResponse{}, nil)

	cfg := Init("TEST_TAH_SOLUTION_ID", "1.0", "aaaaaaaa-bbbbbbbb-cccccccc-dddddddd", SEND_ANONYMIZED_DATA_VALUE_YES, core.LogLevelDebug)
	cfg.Svc = mockHTTPService

	sent, err := cfg.SendMetrics(map[string]interface{}{"metric1": 6.4, "m2": "test"})
	assert.NoError(t, err)
	assert.True(t, sent)
	if err != nil {
		t.Fatalf("[%s] error sending metrics: %v", t.Name(), err)
	}
	if !sent {
		t.Fatalf("[%s] metric not sent while enabled %v", t.Name(), err)
	}
	// Metrics are enabled, HttpPost should be invoked
	mockHTTPService.AssertCalled(t, "HttpPost", mock.Anything, mock.Anything, mock.Anything)
	mockHTTPService.AssertNumberOfCalls(t, "HttpPost", 1)

	// Deactivate metrics
	err = cfg.setMetricsMode(SEND_ANONYMIZED_DATA_VALUE_NO)
	if err != nil {
		t.Fatalf("[%s] error setting metrics mode: %v", t.Name(), err)
	}

	// Test metric sending when disabled
	sent, err = cfg.SendMetrics(map[string]interface{}{"metric1": 6, "m2": "test"})
	assert.NoError(t, err)
	assert.False(t, sent)
	if err != nil {
		t.Fatalf("[%s] error sending metrics: %v", t.Name(), err)
	}
	if sent {
		t.Fatalf("[%s] metric  sent while disabled %v", t.Name(), err)
	}

	// Ensure that HttpPost was not called again (count should remain 1)
	mockHTTPService.AssertNumberOfCalls(t, "HttpPost", 1)
}

func TestInvalidMetricsMode(t *testing.T) {
	t.Parallel()

	cfg := Init("TEST_TAH_SOLUTION_ID", "1.0", "aaaaaaaa-bbbbbbbb-cccccccc-dddddddd", SEND_ANONYMIZED_DATA_VALUE_YES, core.LogLevelDebug)

	err := cfg.setMetricsMode("blah")
	if err == nil || err.Error() != "[setMetricsMode] invalid metrics mode blah" {
		t.Fatalf("[%s] setMetricsMode should have failed", t.Name())
	}
}

func TestDateFormat(t *testing.T) {
	t.Parallel()
	testDate := time.Date(2023, time.December, 25, 0, 0, 0, 0, time.UTC)
	expectedYear := testDate.Year()
	expectedMonth := testDate.Month()
	expectedDay := testDate.Day()
	actual := testDate.Format(TIMESTAMP_FORMAT)
	if actual[0:4] != fmt.Sprintf("%d", expectedYear) {
		t.Fatalf("[TestDateFormat] wrong year: expected %s, got %s", fmt.Sprintf("%d", expectedYear), actual[0:4])
	}
	if actual[5:7] != fmt.Sprintf("%d", expectedMonth) {
		t.Fatalf("[TestDateFormat] wrong month: expected %s, got %s", fmt.Sprintf("%d", expectedMonth), actual[5:7])
	}
	if actual[8:10] != fmt.Sprintf("%d", expectedDay) {
		t.Fatalf("[TestDateFormat] wrong day: expected %s, got %s", fmt.Sprintf("%d", expectedDay), actual[8:10])
	}
}
