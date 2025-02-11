// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package sqs

import (
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
)

/**********
* Test SQS queue
******************/

func TestSQSCrud(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	svc := Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	_, err = svc.Create("Test-Queue" + core.GenerateUniqueId())
	if err != nil {
		t.Errorf("[TestSQS]error creating queue: %v", err)
	}
	err = svc.SetPolicy("Service", "profile.amazonaws.com", []string{"SQS:*"})
	if err != nil {
		t.Errorf("[TestSQS]error setting queue policy queue: %v", err)
	}
	_, err2 := svc.GetAttributes()
	if err2 != nil {
		t.Errorf("[TestSQS] error getting queue atttribute: %v", err2)
	}
	err = svc.Delete()
	if err != nil {
		t.Errorf("[TestSQS]error deleting queue: %v", err)
	}
}

func TestSQSMessages(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	svc := Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	qName := "Test-Queue-2" + core.GenerateUniqueId()
	_, err = svc.Create(qName)
	if err != nil {
		t.Errorf("[TestSQS]error creating queue: %v", err)
	}
	err = svc.Send("test msg")
	if err != nil {
		t.Errorf("[TestSQS]error sending msg to  queue: %v", err)
	}
	content, err1 := svc.Get(GetMessageOptions{})
	if err1 != nil {
		t.Errorf("[TestSQS]error getting msg from queue: %v", err1)
	}
	if len(content.Peek) == 0 {
		t.Errorf("[TestSQS] no record returned")
	}
	if content.Peek[0].Body != "test msg" {
		t.Errorf("[TestSQS] message in queue should be %v", "test msg")
	}
	err = svc.SendWithStringAttributes("test msg 2", map[string]string{"key": "testVal"})
	if err != nil {
		t.Errorf("[TestSQS]error sending msg to  queue: %v", err)
	}
	content, err1 = svc.Get(GetMessageOptions{})
	if err1 != nil {
		t.Errorf("[TestSQS] error getting msg from queue: %v", err1)
	}
	if len(content.Peek) == 0 {
		t.Errorf("[TestSQS] no record returned")
	}
	if content.Peek[0].Body != "test msg 2" {
		t.Errorf("[TestSQS] message in queue should be %v", "test msg 2")
	}
	if content.Peek[0].MessageAttributes["key"] != "testVal" {
		t.Errorf("[TestSQS] message in queue should have message attribute %v", map[string]string{"key": "testVal"})
	}
	err = svc.DeleteByName(qName)
	if err != nil {
		t.Errorf("[TestSQS]error deleting queue by name: %v", err)
	}
}
