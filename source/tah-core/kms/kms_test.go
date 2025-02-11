// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kms

import (
	"testing"

	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/utils/config"
)

func TestCreateDelete(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	cfg := Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	keyArn, err := cfg.CreateKey("tah core unit test key")
	if err != nil {
		t.Errorf("Error creating KMS key: %v", err)
	}
	err = cfg.DeleteKey(keyArn)
	if err != nil {
		t.Errorf("Error deleting KMS key: %v", err)
	}
}
