// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package eventbridge

import (
	"log"
	"sync"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
)

/**********
* EventBridge
************/
type TestEventbridgeBody struct {
	Value string
}

func TestSendMessage(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	cfg, err := InitWithRandomBus(envCfg.Region, "test-bus-", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("[TestSendMessage] error initializing eventbridge with random bus: %v", err)
	}
	err = cfg.SendMessage("test_source", "test_description", TestEventbridgeBody{Value: "yoouhou"})
	if err != nil {
		t.Errorf("[TestSendMessage] error sending message: %v", err)
	}
	err = cfg.DeleteBus(cfg.EventBus)
	if err != nil {
		t.Errorf("[TestSendMessage] error deleting bus: %v", err)
	}

}

func TestMockThreadSafety(t *testing.T) {
	mock := MockConfig{}
	nThreads := 100
	var wg sync.WaitGroup
	wg.Add(nThreads)
	log.Printf("starting %d threads", nThreads)
	for i := 0; i < nThreads; i++ {
		log.Printf("starting thread %d", i)
		go func(index int) {
			mock.SendMessage("test", "test decsription", map[string]interface{}{"value": index})
			wg.Done()
		}(i)
	}
	wg.Wait()
	if len(mock.Events) != nThreads {
		log.Printf("invalid mock content: %+v", mock.Events)
		t.Errorf("[TestMock] expected %d events, got %d", nThreads, len(mock.Events))
	}
}
