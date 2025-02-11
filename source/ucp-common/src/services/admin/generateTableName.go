// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	model "tah/upt/source/ucp-common/src/model/admin"
)

func BuildTableName(env string, bo model.BusinessObject, domainName string) string {
	return "ucp_" + env + "_" + bo.Name + "_" + domainName
}

func BuildGlueDestination(bucket string, domainName string) string {
	return bucket + "/" + domainName
}
