// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
)

var PORTAL_CONFIG_PROMPT_PK = "ai_prompt"
var PORTAL_CONFIG_MATCH_THRESHOLD_PK = "match_threshold"

func UpdateDomainSetting(tx core.Transaction, PortalConfigDB db.DBConfig, pKey, domainName, value string, isActive bool) error {
	dynamoRecord := model.DynamoDomainConfig{
		Pk:       pKey,
		Sk:       domainName,
		Value:    value,
		IsActive: isActive,
	}

	_, err := PortalConfigDB.Save(dynamoRecord)
	if err != nil {
		tx.Error("[UpdatePromptConfig] Error saving prompt configuration")
		return err
	}

	return nil
}
