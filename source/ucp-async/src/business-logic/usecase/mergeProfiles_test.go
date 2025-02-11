// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"testing"

	"log"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	"tah/upt/source/ucp-common/src/model/async"
	ucModel "tah/upt/source/ucp-common/src/model/async/usecase"
	commonServices "tah/upt/source/ucp-common/src/services/traveller"
	"time"
)

func TestMergeProfiles(t *testing.T) {
	pk := "domain_sourceProfileId"
	sk := "match_targetProfileId"
	indexName := "matchesByConfidenceScore"
	indePkName := "runId"
	indexSkName := "scoreTargetId"
	tableName := "merge_profiles_test_table_" + time.Now().Format("20060102150405")
	dbConfig := db.Init(tableName, pk, sk, "", "")
	err := dbConfig.CreateTableWithOptions(tableName, pk, sk, db.TableOptions{
		GSIs: []db.GSI{
			{
				IndexName: indexName,
				PkName:    indePkName,
				PkType:    "S",
				SkName:    indexSkName,
				SkType:    "S",
			},
		},
	})
	if err != nil {
		t.Errorf("Error initializing db: %s", err)
	}
	dbConfig.WaitForTableCreation()

	sourceID := "abcdesdlfheowinfqsdf3"
	targetID := "083h9298dgwiegfowieh"
	mergeRq := []commonModel.MergeRq{{SourceProfileID: sourceID, TargetProfileID: targetID}}
	// Execute usecase
	uc := InitMergeProfiles()
	if uc.Name() != "MergeProfiles" {
		t.Errorf("Expected MergeProfiles, got %v", uc.Name())
	}
	domainName := "my_test_domain_" + time.Now().Format("20060102150405")
	payload := async.AsyncInvokePayload{
		EventID:       "test-event-id",
		Usecase:       async.USECASE_EMPTY_TABLE,
		TransactionID: "test-transaction-id",
		Body: ucModel.MergeProfilesBody{
			Domain: domainName,
			Rq:     mergeRq,
		},
	}

	solutionsConfig := awssolutions.InitMock()
	domain := customerprofiles.Domain{Name: domainName}
	domains := []customerprofiles.Domain{domain}
	profile := profilemodel.Profile{}
	profiles := []profilemodel.Profile{}
	mappings := []customerprofiles.ObjectMapping{}
	var accp = customerprofiles.InitMock(&domain, &domains, &profile, &profiles, &mappings)
	services := async.Services{
		AccpConfig:      accp,
		SolutionsConfig: solutionsConfig,
		MatchDB:         dbConfig,
	}

	mp := commonServices.BuildMatchPairRecord(sourceID, targetID, "0999999", domainName, false)
	mp1 := commonServices.BuildMatchPairRecord(targetID, sourceID, "0999999", domainName, false)
	mp2 := commonServices.BuildMatchPairRecord(targetID, "anotherProfileid", "0999999", domainName, false)
	mp3 := commonServices.BuildMatchPairRecord("anotherProfileid", targetID, "0999999", domainName, false)
	err = dbConfig.SaveMany([]commonModel.MatchPair{mp, mp1, mp2, mp3})
	if err != nil {
		t.Errorf("Error inserting profile pairs for testing: %s", err)
	}
	mpBeforeMerge := commonModel.MatchPair{}
	err = dbConfig.Get(mp.Pk, mp.Sk, &mpBeforeMerge)
	if err != nil {
		t.Errorf("Error getting profile pair for testing: %s", err)
	}
	if mpBeforeMerge.Pk != mp.Pk {
		t.Errorf("match pair before merge pk (%v) should equal match pair after merge (%v)", mpBeforeMerge.Pk, mp.Pk)
	}
	err = uc.Execute(payload, services, core.NewTransaction(t.Name(), "", core.LogLevelDebug))
	if err != nil {
		t.Errorf("Error invoking usecase: %v", err)
	}

	for _, m := range []commonModel.MatchPair{mp, mp1, mp2, mp3} {
		mpAfterMerge := commonModel.MatchPair{}
		err = dbConfig.Get(m.Pk, m.Sk, &mpAfterMerge)
		if err != nil {
			t.Errorf("error fetching record after merge: %s", err)
		}
		if mpAfterMerge.Pk != "" {
			t.Errorf("MatchPair should have beedn deleted but got: %+v", mpAfterMerge)
		}
	}
	log.Printf("TestMergeProfile: deleting test resources")
	err = dbConfig.DeleteTable(tableName)
	if err != nil {
		t.Errorf("Error deleting table: %s", err)
	}
}
