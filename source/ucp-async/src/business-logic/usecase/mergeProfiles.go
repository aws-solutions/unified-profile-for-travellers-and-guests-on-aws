// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"encoding/json"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-common/src/model/admin"
	"tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	commonServices "tah/upt/source/ucp-common/src/services/traveller"
)

type MergeProfiles struct {
	name string
}

func InitMergeProfiles() *MergeProfiles {
	return &MergeProfiles{name: "MergeProfiles"}
}

func (u *MergeProfiles) Name() string {
	return u.name
}

func (u *MergeProfiles) Execute(payload async.AsyncInvokePayload, services async.Services, tx core.Transaction) error {
	// Convert payload from string to json data
	payloadJson, err := json.Marshal(payload.Body)
	if err != nil {
		tx.Error("[%s] Error marshalling payload %v", u.name, err)
	}
	rq := uc.MergeProfilesBody{}
	err = json.Unmarshal(payloadJson, &rq)
	if err != nil {
		tx.Error("[%s] Error unmarshalling payload %v", u.name, err)
	}
	tx.Debug("Setting domain to ACCP config to %s", rq.Domain)
	services.AccpConfig.SetDomain(rq.Domain)
	tx.Info("Executing Merge profiles service")
	err = commonServices.MergeProfiles(tx, services.AccpConfig, rq.Domain, rq.Rq)
	if err != nil {
		tx.Error("Error merging profiles: %+v", err)
		return err
	}
	successfullyMerges, failedMerges, err := commonServices.WaitForProfileMergeCompletion(tx, services.AccpConfig, rq.Domain, rq.Rq, 60)
	if err != nil {
		tx.Error("Error merging profiles: %+v. %v profiles have not been successfully merged", err, len(failedMerges))
		return err
	}
	tx.Info("Merge Successfull: Deleting match pair from DynamoDB")
	toDelete := []model.MergeRq{}
	for _, merged := range successfullyMerges {
		tx.Debug("Identifying all matches for target profile %v", merged.TargetProfileID)
		matches, err := commonServices.SearchMatches(services.MatchDB, rq.Domain, merged.TargetProfileID)
		if err != nil {
			tx.Error("Could not identify matches for target profile %v. error: %v", merged.TargetProfileID, err)
		}
		for _, match := range matches {
			toDelete = append(toDelete, model.MergeRq{
				SourceProfileID: merged.TargetProfileID,
				TargetProfileID: match.TargetProfileID,
			})
		}
	}
	tx.Info("Deleting the following profiles: %v", toDelete)
	err = commonServices.DeleteMatchPairs(tx, &services.MatchDB, rq.Domain, toDelete)
	if err != nil {
		tx.Error("Error deleting match pairs: %+v", err)
	}
	return err
}
