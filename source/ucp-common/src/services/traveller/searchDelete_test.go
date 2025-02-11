// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	"testing"
)

func TestSearchDeleteMatches(t *testing.T) {
	domain := "fakedomain"
	tableName := "common-table-test_delete_many-" + core.GenerateUniqueId()
	dbConfig, err := db.InitWithNewTable(tableName, "domain_sourceProfileId", "match_targetProfileId", "", "")
	if err != nil {
		t.Fatalf("Error initializing db: %s", err)
	}
	t.Cleanup(func() { dbConfig.DeleteTable(tableName) })
	dbConfig.WaitForTableCreation()

	matchPair := commonModel.MatchPair{
		Pk:              domain + "_" + "source_id",
		Sk:              "match_" + "match_id",
		TargetProfileID: "match_id",
		Score:           "0.98",
	}

	_, err = dbConfig.Save(matchPair)
	if err != nil {
		t.Errorf("Error saving matchPair: %s", err)
	}

	newPairList, err := SearchMatches(dbConfig, domain, "source_id")
	if err != nil {
		t.Errorf("Error searching matches: %s", err)
	}
	if len(newPairList) != 1 {
		t.Errorf("SearchMatches returned wrong number of matches")
	}
	newPair := newPairList[0]
	if newPair.TargetProfileID != matchPair.TargetProfileID {
		t.Errorf("SearchMatches returned wrong match")
	}
	err = DeleteMatches(dbConfig, domain, "source_id")
	if err != nil {
		t.Errorf("Error deleting matches: %s", err)
	}
}

func TestSearchDeleteMatch(t *testing.T) {
	domain := "fakedomain"
	tableName := "common-table-test_delete-" + core.GenerateUniqueId()
	dbConfig, err := db.InitWithNewTable(tableName, "domain_sourceProfileId", "match_targetProfileId", "", "")
	if err != nil {
		t.Errorf("Error initializing db: %s", err)
	}
	t.Cleanup(func() { dbConfig.DeleteTable(tableName) })
	dbConfig.WaitForTableCreation()

	matchPair := commonModel.MatchPair{
		Pk:              domain + "_" + "source_id",
		Sk:              "match_" + "match_id",
		TargetProfileID: "match_id",
		Score:           "0.56",
	}

	_, err = dbConfig.Save(matchPair)
	if err != nil {
		t.Errorf("Error saving matchPair: %s", err)
	}
	newPairList, err1 := SearchMatches(dbConfig, domain, "source_id")
	if err1 != nil {
		t.Errorf("Error searching matches: %s", err1)
	}
	if len(newPairList) != 1 {
		t.Errorf("There should be 1 match after save")
	}
	err = DeleteMatch(dbConfig, domain, "source_id", "match_id")
	if err != nil {
		t.Errorf("Error deleting matches: %s", err)
	}
	newPairList, err = SearchMatches(dbConfig, domain, "source_id")
	if err != nil {
		t.Errorf("Error searching matches: %s", err)
	}
	if len(newPairList) != 0 {
		t.Errorf("There should be no match after delete")
	}
}
