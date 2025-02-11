// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package eventbridge

import (
	"log"
	"sync"
)

type MockConfig struct {
	Events []interface{}
	mutex  sync.Mutex //to ensure thread safety of the mock
}

func (c *MockConfig) SendMessage(source string, description string, data interface{}) error {
	log.Printf("[mock] send messages source: %v, descripriono: %v, data: %v", source, description, data)
	c.mutex.Lock()
	c.Events = append(c.Events, data)
	c.mutex.Unlock()
	return nil
}
