// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	model "tah/upt/source/ucp-common/src/model/async"
	"testing"
)

func TestMain(t *testing.T) {
	stubContext := context.Background()
	payload := model.AsyncInvokePayload{}
	rawPayload, _ := json.Marshal(payload)
	HandleRequest(stubContext, string(rawPayload))
}
