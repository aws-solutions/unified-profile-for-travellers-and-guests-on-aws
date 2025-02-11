// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

type BusinessObjectRecord struct {
	Domain       string      `json:"domain"`
	ObjectType   string      `json:"objectType"`
	ModelVersion string      `json:"modelVersion"`
	Data         interface{} `json:"data"`
}
