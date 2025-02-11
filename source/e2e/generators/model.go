// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package generators

type BusinessObjectRecord struct {
	Domain       string                 `json:"domain"`
	ObjectType   string                 `json:"objectType"`
	ModelVersion string                 `json:"modelVersion"`
	Data         map[string]interface{} `json:"data"`
}

const INGEST_TIMESTAMP_FORMAT = "2006-01-02T15:04:05.000000Z"
