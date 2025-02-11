// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	_, _, err := LoadConfigs()
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			// ignore failure when infra hasn't been deployed and infra-config-env.json doesn't exist yet
			return
		}
		t.Errorf("Error loading env config %v", err)
	}
}
