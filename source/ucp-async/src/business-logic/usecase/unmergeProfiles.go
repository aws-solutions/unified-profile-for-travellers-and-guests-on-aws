// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"encoding/json"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	commonServices "tah/upt/source/ucp-common/src/services/traveller"
)

type UnmergeProfiles struct {
	name string
}

func InitUnmergeProfiles() *UnmergeProfiles {
	return &UnmergeProfiles{name: "UnmergeProfiles"}
}

func (u *UnmergeProfiles) Name() string {
	return u.name
}

func (u *UnmergeProfiles) Execute(payload async.AsyncInvokePayload, services async.Services, tx core.Transaction) error {

	// Convert payload from string to json data
	rq := uc.UnmergeProfilesBody{}
	payloadJson, err := json.Marshal(payload.Body)
	if err != nil {
		tx.Error("[%s] Error marshalling payload %v", u.name, err)
	}
	err = json.Unmarshal(payloadJson, &rq)
	if err != nil {
		tx.Error("[%s] Error unmarshalling payload %v", u.name, err)
	}

	tx.Debug("[%s] Setting domain to ACCP config to %s", u.name, rq)
	services.AccpConfig.SetDomain(rq.Domain)

	tx.Info("[%s] Executing Unmerge profiles service", u.name)
	err = commonServices.UnmergeProfiles(tx, services.AccpConfig, rq.Domain, rq.Rq)
	if err != nil {
		tx.Error("[%s] Error unmerging profiles: %+v", u.name, err)
		return err
	}
	return err
}
