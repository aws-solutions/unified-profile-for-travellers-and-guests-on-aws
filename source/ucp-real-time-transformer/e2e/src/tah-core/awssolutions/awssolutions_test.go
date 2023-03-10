package awssolutions

import "testing"

/**********
* IAM
************/
func TestMetricService(t *testing.T) {
	cfg := Init("TEST_TAH_SOLUTION_ID", "1.0", "aaaaaaaa-bbbbbbbb-cccccccc-dddddddd", "Yes")
	sent, err := cfg.SendMetrics(map[string]interface{}{"metric1": 6.4, "m2": "test"})
	if err != nil {
		t.Errorf("[TestMetricService] error sending metrics: %v", err)
	}
	if !sent {
		t.Errorf("[TestMetricService] metric not sent while enabled %v", err)
	}
	cfg = Init("TEST_TAH_SOLUTION_ID", "1.0", "aaaaaaaa-bbbbbbbb-cccccccc-dddddddd", "yes")
	sent, err = cfg.SendMetrics(map[string]interface{}{"metric1": 6.3, "m2": "test"})
	if err != nil {
		t.Errorf("[TestMetricService] error sending metrics: %v", err)
	}
	if !sent {
		t.Errorf("[TestMetricService] metric not sent while enabled %v", err)
	}
	cfg = Init("TEST_TAH_SOLUTION_ID", "1.0", "aaaaaaaa-bbbbbbbb-cccccccc-dddddddd", "No")
	sent, err = cfg.SendMetrics(map[string]interface{}{"metric1": 6, "m2": "test"})
	if err != nil {
		t.Errorf("[TestMetricService] error sending metrics: %v", err)
	}
	if sent {
		t.Errorf("[TestMetricService] metric  sent while disabled %v", err)
	}
}
