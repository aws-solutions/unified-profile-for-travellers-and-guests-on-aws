// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/glue"
	commonModel "tah/upt/source/ucp-common/src/model/admin"

	"time"
)

type ResponseWrapper struct {
}

type PartitionUpdateStatus struct {
	Pk          string    `json:"item_id"`
	Sk          string    `json:"item_type"`
	LastUpdated time.Time `json:"lastUpdated"`
	TableName   string    `json:"tableName"`
	Error       string    `json:"error"`
	Status      string    `json:"status"`
}

type AWSConfig struct {
	Tx                   core.Transaction
	Glue                 glue.IConfig
	Dynamo               db.DBConfig
	Accp                 customerprofiles.ICustomerProfileLowCostConfig
	BizObjectBucketNames map[string]string
	JobNames             map[string]string
	Env                  map[string]string
	Domains              []commonModel.Domain
	RequestedJob         string
}

type JobBookmark struct {
	Pk string `json:"item_id"`
	Sk string `json:"item_type"`
}
