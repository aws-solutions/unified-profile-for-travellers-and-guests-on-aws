package appflow

import (
	"tah/core/s3"

	"os"
	"testing"
)

var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")

func TestAppflow(t *testing.T) {
	// Setup
	s3c := s3.InitRegion(TAH_CORE_REGION)
	bucketName, err0 := s3c.CreateRandomBucket("tah-core-glue-test-bucket")
	if err0 != nil {
		t.Errorf("error creating bucket: %+v", err0)
	}
	s3c.Bucket = bucketName
	flowName := "appflow-test-flow"

	resource := []string{"arn:aws:s3:::" + bucketName, "arn:aws:s3:::" + bucketName + "/*"}
	actions := []string{"s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy", "s3:getbucketacl", "s3:putobjectacl"}
	principal := map[string][]string{"Service": {"appflow.amazonaws.com"}}
	err := s3c.AddPolicy(bucketName, resource, actions, principal)
	if err != nil {
		t.Errorf("error adding bucket policy %+v", err)
	}

	// Test AppFlow
	client := Init()
	err = client.createTestFlow(flowName, bucketName)
	if err != nil {
		t.Errorf("[TestAppflow] Error creating flow: %v", err)
	}
	var flow = Flow{}
	flow, err = client.GetFlow(flowName)
	if err != nil {
		t.Errorf("[TestAppflow] Error getting flow: %v", err)
	}
	if flow.Name != flowName {
		t.Errorf("[TestAppflow] Error with data received from GetFlow: %v", err)
	}
	var flows = []Flow{}
	flows, err = client.GetFlows([]string{flowName})
	if err != nil {
		t.Errorf("[TestAppflow] Error getting flows: %v", err)
	}
	if flows[0].Name != flowName || len(flows) != 1 {
		t.Errorf("[TestAppflow] Error with data received from GetFlows: %v", err)
	}
	var status = FlowStatusOutput{}
	status, err = client.StartFlow(flowName)
	if err != nil {
		t.Errorf("[TestAppflow] Error starting flow: %v", err)
	}
	if status.FlowStatus != FLOW_STATUS_ACTIVE {
		t.Errorf("[TestAppflow] Flow did not start successfully")
	}
	status, err = client.StopFlow(flowName)
	if err != nil {
		t.Errorf("[TestAppflow] Error stopping flow: %v", err)
	}
	if status.FlowStatus != FLOW_STATUS_SUSPENDED {
		t.Errorf("[TestAppflow] Flow did not stop successfully")
	}
	err = client.DeleteFlow(flowName, true)
	if err != nil {
		t.Errorf("[TestAppflow] Error deleting flow: %v", err)
	}

	// Tear Down
	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestAppflow] error deleting bucket: %v", err)
	}
}
