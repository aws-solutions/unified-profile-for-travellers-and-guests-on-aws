// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"fmt"
	"strings"
	core "tah/upt/source/tah-core/core"
	sqs "tah/upt/source/tah-core/sqs"
	"testing"
)

// this file contains function used to facilitate LCS testing
func randomDomain(prefix string) string {
	return fmt.Sprintf("%s_%s", strings.ToLower(prefix), strings.ToLower(core.GenerateUniqueId()))
}

func randomTable(prefix string) string {
	return fmt.Sprintf("%s_%s", strings.ToLower(prefix), strings.ToLower(core.GenerateUniqueId()))
}

// Build an SQS queue used for testing. The queue will be deleted after the test execution is completed.
func buildTestMergeQueue(t *testing.T, region string) sqs.IConfig {
	t.Helper()
	sqsCfg := sqs.Init(region, "", "")
	_, err := sqsCfg.CreateRandom(strings.ReplaceAll((truncateString(t.Name(), 60) + core.GenerateUniqueId() + "merger"), "/", ""))
	if err != nil {
		t.Fatalf("[%s] error creating merge queue: %v", t.Name(), err)
	}

	t.Cleanup(func() {
		err = sqsCfg.Delete()
		if err != nil {
			t.Errorf("[%s] Delete failed: %s", t.Name(), err)
		}
	})

	return &sqsCfg
}

func buildTestCPWriterQueue(t *testing.T, region string) sqs.IConfig {
	t.Helper()
	sqsCfg := sqs.Init(region, "", "")
	_, err := sqsCfg.CreateRandom(strings.ReplaceAll((truncateString(t.Name(), 60) + core.GenerateUniqueId() + "cpwriter"), "/", ""))
	if err != nil {
		t.Fatalf("[%s] error creating CP writer queue: %v", t.Name(), err)
	}

	t.Cleanup(func() {
		err = sqsCfg.Delete()
		if err != nil {
			t.Errorf("[%s] Delete failed: %s", t.Name(), err)
		}
	})

	return &sqsCfg
}

func truncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}

	return str[:length]
}
