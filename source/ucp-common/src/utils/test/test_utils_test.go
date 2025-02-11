// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"os"
	"testing"
)

var region = os.Getenv("UCP_REGION")
var aws_region = os.Getenv("AWS_REGION")

func TestGetTestRegion(t *testing.T) {
	tr := GetTestRegion()
	if tr != region && tr != aws_region {
		t.Errorf("Invalid test regions %v", tr)
	}
}
