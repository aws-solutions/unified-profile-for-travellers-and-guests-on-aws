// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"log"
	"sort"
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/admin"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

func TestPaginationLogic(t *testing.T) {
	// Set up resources
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("unable to load configs: %v", err)
	}
	log.Printf("Running test, region: %s", envCfg.Region)
	//Write the test
	pk := "domain_sourceProfileId"
	sk := "match_targetProfileId"
	indexName := "matchesByConfidenceScore"
	indexPkName := "runId"
	indexSkName := "scoreTargetId"
	tableName := "pagination_test_table_" + time.Now().Format("20060102150405")
	dbConfig := db.Init(tableName, pk, sk, "", "")
	domain := "matchpair_test_domain"

	err = dbConfig.CreateTableWithOptions(tableName, pk, sk, db.TableOptions{
		GSIs: []db.GSI{
			{
				IndexName: indexName,
				PkName:    indexPkName,
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

	mps := []model.MatchPair{
		buildTestMatchPair(domain, "s1", "t1", "0.9999", false),
		buildTestMatchPair(domain, "s2", "t2", "0.9998", false),
		buildTestMatchPair(domain, "s3", "t3", "0.9997", false),
		buildTestMatchPair(domain, "s4", "t4", "0.9996", false),
		UpdateFalsePositiveRecord(buildTestMatchPair(domain, "s5", "t5", "0.9995", false)),
		UpdateFalsePositiveRecord(buildTestMatchPair(domain, "s6", "t6", "0.9994", false)),
		UpdateFalsePositiveRecord(buildTestMatchPair(domain, "s7", "t7", "0.9993", false)),
		UpdateFalsePositiveRecord(buildTestMatchPair(domain, "s8", "t8", "0.9992", false)),
		UpdateFalsePositiveRecord(buildTestMatchPair(domain, "s9", "t9", "0.9991", false)),
		buildTestMatchPair(domain, "s10", "t10", "0.9990", false),
		buildTestMatchPair(domain, "t1", "s1", "0.9999", false),
		buildTestMatchPair(domain, "t2", "s2", "0.9998", false),
		buildTestMatchPair(domain, "t3", "s3", "0.9997", false),
		buildTestMatchPair(domain, "t4", "s4", "0.9996", false),
		UpdateFalsePositiveRecord(buildTestMatchPair(domain, "t5", "s5", "0.9995", false)),
		UpdateFalsePositiveRecord(buildTestMatchPair(domain, "t6", "s6", "0.9994", false)),
		UpdateFalsePositiveRecord(buildTestMatchPair(domain, "t7", "s7", "0.9993", false)),
		UpdateFalsePositiveRecord(buildTestMatchPair(domain, "t8", "s8", "0.9992", false)),
		buildTestMatchPair(domain, "t9", "s9", "0.9991", false),
		buildTestMatchPair(domain, "t10", "s10", "0.9990", false),
	}
	err = dbConfig.SaveMany(mps)
	if err != nil {
		t.Errorf("Error saving matchPair: %s", err)
	}
	time.Sleep(time.Second * 10)

	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	// Match pairs for page 0, size 10
	matchPairs, err := FindAllMatches(tx, dbConfig, model.PaginationOptions{
		Page:     0,
		PageSize: 10,
	})
	log.Printf("Full list (0-10): %+v", matchPairs)
	if err != nil {
		t.Errorf("FindAllMatches failed with error: %v", err)
	}
	if len(matchPairs) != 6 {
		t.Errorf("Invalid paginated results. Should return 6 rows. not %v: %+v", len(matchPairs), matchPairs)
	}

	// Match pairs for page 0, size 8
	matchPairs, err = FindAllMatches(tx, dbConfig, model.PaginationOptions{
		Page:     0,
		PageSize: 8,
	})
	if err != nil {
		t.Errorf("FindAllMatches failed with error: %v", err)
	}
	matchPairCount_0_8 := len(matchPairs)
	if matchPairCount_0_8 != 4 {
		t.Errorf("Invalid paginated results. Should return 4 rows. not %v: %+v", len(matchPairs), matchPairs)
	}

	// Match pairs for page 0, size 4
	matchPairs, err = FindAllMatches(tx, dbConfig, model.PaginationOptions{
		Page:     0,
		PageSize: 4,
	})
	log.Printf("Full list (0-4): %+v", matchPairs)
	if err != nil {
		t.Errorf("FindAllMatches failed with error: %v", err)
	}
	matchPairCount_0_4 := len(matchPairs)
	if matchPairCount_0_4 != 2 {
		t.Errorf("Invalid paginated results. Should return 2 rows. not %v: %+v", len(matchPairs), matchPairs)
	}

	// Match pairs for page 1, size 4
	matchPairs, err = FindAllMatches(tx, dbConfig, model.PaginationOptions{
		Page:     1,
		PageSize: 4,
	})
	log.Printf("Full list (1-4): %+v", matchPairs)
	if err != nil {
		t.Errorf("FindAllMatches failed with error: %v", err)
	}
	matchPairCount_1_4 := len(matchPairs)
	if matchPairCount_1_4 != 2 {
		t.Errorf("Invalid paginated results. Should return 2 rows. not %v: %+v", len(matchPairs), matchPairs)
	}

	if matchPairCount_0_8 != matchPairCount_0_4+matchPairCount_1_4 {
		t.Errorf("Invalid paginated results. Total count does not add up with paginated count")
	}

	err = dbConfig.DeleteTable(tableName)
	if err != nil {
		t.Errorf("Error deleting table: %s", err)
	}

}

func buildTestMatchPair(domain, source, target, score string, falsePositive bool) model.MatchPair {
	ids := []string{source, target}
	sort.Strings(ids)
	return model.MatchPair{
		Pk:              domain + "_" + source,
		Sk:              constant.MATCH_PREFIX + target,
		TargetProfileID: target,
		Score:           "0.9999",
		FalsePositive:   falsePositive,
		//we use these attributes for the global secondary index used to return all matches sorted by score
		RunID:         constant.MATCH_LATEST_RUN,
		ScoreTargetID: constant.MATCH_PREFIX + score + "_" + strings.Join(ids, "_"),
	}
}

func TestBuildMatchPairRecord(t *testing.T) {
	sourceId := "123456789"
	targetId := "987654321"
	score := "0.9999999"
	domain := "test_domain"
	log.Printf("Test building match pair for standard records")
	mp := BuildMatchPairRecord(sourceId, targetId, score, domain, false)
	if mp.Pk != domain+"_"+sourceId {
		t.Errorf("Expected Pk %s, got %s", domain+"_"+sourceId, mp.Pk)
	}
	if mp.Sk != constant.MATCH_PREFIX+targetId {
		t.Errorf("Expected Sk %s, got %s", constant.MATCH_PREFIX+targetId, mp.Sk)
	}
	if mp.Score != score {
		t.Errorf("Expected Score %s, got %s", score, mp.Score)
	}
	if mp.TargetProfileID != targetId {
		t.Errorf("Expected TargetProfileID %s, got %s", targetId, mp.TargetProfileID)
	}
	if mp.RunID != constant.MATCH_LATEST_RUN {
		t.Errorf("Expected RunID %s, got %s", constant.MATCH_LATEST_RUN, mp.RunID)
	}
	if mp.ScoreTargetID != constant.MATCH_PREFIX+score+"_"+sourceId+"_"+targetId {
		t.Errorf("Expected ScoreTargetID %s, got %s", constant.MATCH_PREFIX+score+"_"+sourceId+"_"+targetId, mp.ScoreTargetID)
	}
	if mp.MergeInProgress {
		t.Errorf("MergeInProgress should be set to false by default")
	}
	if mp.FalsePositive {
		t.Errorf("FalsePositive should be set to false by default")
	}

	mp = BuildMatchPairRecord(sourceId, targetId, score, domain, true)
	if !mp.MergeInProgress {
		t.Errorf("Merge in porgress should be set to true")
	}
}

func TestFilterDuplicates(t *testing.T) {
	mps := []model.MatchPair{{
		Pk:            "domain_s1",
		Sk:            "match_t1",
		RunID:         constant.MATCH_LATEST_RUN,
		ScoreTargetID: buildScoreTargetId("0.99999", "s1", "t1"),
	}, {
		Pk:            "domain_t1",
		Sk:            "match_s1",
		RunID:         constant.MATCH_LATEST_RUN,
		ScoreTargetID: buildScoreTargetId("0.99999", "t1", "s1"),
	}, {
		Pk:            "domain_s1",
		Sk:            "match_t2",
		RunID:         constant.MATCH_LATEST_RUN,
		ScoreTargetID: buildScoreTargetId("0.99999", "s1", "t2"),
	}}

	filtered := filterDuplicates(mps)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered records, got %d: %+v", len(filtered), filtered)
		return
	}

}

func TestFindAllMatches(t *testing.T) {
	// Set up resources
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("unable to load configs: %v", err)
	}
	log.Printf("Running test, region: %s", envCfg.Region)
	//Write the test
	pk := "domain_sourceProfileId"
	sk := "match_targetProfileId"
	indexName := "matchesByConfidenceScore"
	indexPkName := "runId"
	indexSkName := "scoreTargetId"
	tableName := "all_matches_test_table_" + time.Now().Format("20060102150405")
	dbConfig := db.Init(tableName, pk, sk, "", "")
	domain := "matchpair_test_domain"

	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	err = dbConfig.CreateTableWithOptions(tableName, pk, sk, db.TableOptions{
		GSIs: []db.GSI{
			{
				IndexName: indexName,
				PkName:    indexPkName,
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

	matchPair1 := model.MatchPair{
		Pk:              domain + "_" + "s1",
		Sk:              constant.MATCH_PREFIX + "t1",
		TargetProfileID: "t1",
		Score:           "0.9999",
		//we use these attributes for the global secondary index used to return all matches sorted by score
		RunID:         constant.MATCH_LATEST_RUN,
		ScoreTargetID: constant.MATCH_PREFIX + "0.9999_s1_t1",
	}
	matchPair1Reversed := model.MatchPair{
		Pk:              domain + "_" + "t1",
		Sk:              constant.MATCH_PREFIX + "s1",
		TargetProfileID: "s3",
		Score:           "0.9999",
		//we use these attributes for the global secondary index used to return all matches sorted by score
		RunID:         constant.MATCH_LATEST_RUN,
		ScoreTargetID: constant.MATCH_PREFIX + "0.9999_s1_t1",
	}
	matchPair2 := model.MatchPair{
		Pk:              domain + "_" + "s2",
		Sk:              constant.MATCH_PREFIX + "t2",
		TargetProfileID: "t2",
		Score:           "0.9998",
		//we use these attributes for the global secondary index used to return all matches sorted by score
		RunID:         constant.MATCH_LATEST_RUN,
		ScoreTargetID: constant.MATCH_PREFIX + "0.9998_s2_t2",
	}
	matchPair2Reverse := model.MatchPair{
		Pk:              domain + "_" + "t2",
		Sk:              constant.MATCH_PREFIX + "s2",
		TargetProfileID: "s2",
		Score:           "0.9998",
		//we use these attributes for the global secondary index used to return all matches sorted by score
		RunID:         constant.MATCH_LATEST_RUN,
		ScoreTargetID: constant.MATCH_PREFIX + "0.9998_s2_t2",
	}
	err = dbConfig.SaveMany([]model.MatchPair{matchPair2, matchPair1Reversed, matchPair1, matchPair2Reverse})
	if err != nil {
		t.Errorf("Error saving matchPair: %s", err)
	}

	mp, err := GetMatchPair(tx, dbConfig, domain, "s1", "t1")
	if err != nil {
		t.Errorf("Error getting matchPair: %s", err)
	}
	if mp.Score != "0.9999" || mp.TargetProfileID != "t1" || mp.ScoreTargetID != constant.MATCH_PREFIX+"0.9999_s1_t1" {
		t.Errorf("GetMatchPair failed, expected %+v and %+v", matchPair1, mp)
	}

	time.Sleep(time.Second * 10)
	matchPairs, err := FindAllMatches(tx, dbConfig, model.PaginationOptions{})
	if err != nil {
		t.Errorf("FindAllMatches failed with error: %v", err)
	}
	if len(matchPairs) != 2 {
		t.Errorf("FindAllMatches failed, expected 2 matches, got %d", len(matchPairs))
		return
	}
	if matchPairs[0].Score != "0.9999" || matchPairs[1].Score != "0.9998" {
		t.Errorf("FindAllMatches failed, expected 0.9999 > 0.9998, got %s and %s", matchPairs[0].Score, matchPairs[1].Score)
	}

	err = MarkFalsePositive(tx, &dbConfig, domain, "s2", "t2")
	if err != nil {
		t.Errorf("Error marking false positive: %s", err)
	}
	time.Sleep(time.Second * 10)
	matchPairs, err = FindAllMatches(tx, dbConfig, model.PaginationOptions{})
	if err != nil {
		t.Errorf("FindAllMatches failed with error: %v", err)
	}
	if len(matchPairs) != 1 {
		t.Errorf("FindAllMatches failed, expected 1 match after marking on as false positive, got %d", len(matchPairs))
		return
	}
	if (matchPairs[0].Pk != domain+"_s1" && matchPairs[0].Pk != domain+"_t1") || (matchPairs[0].Sk != constant.MATCH_PREFIX+"t1" && matchPairs[0].Sk != constant.MATCH_PREFIX+"s1") {
		t.Errorf("FindAllMatches failed, expected left pk=%s|%s, sk=%s|%s, got pk=%s and sk=%s", domain+"_s1", domain+"_t1", constant.MATCH_PREFIX+"t1", constant.MATCH_PREFIX+"s1", matchPairs[0].Pk, matchPairs[0].Sk)
	}
	if matchPairs[0].MergeInProgress {
		t.Errorf("Merge in porgress should be set to false by default")
	}

	err = MarkMergeInProgress(tx, &dbConfig, domain, []model.MergeRq{{SourceProfileID: "s1", TargetProfileID: "t1"}})
	if err != nil {
		t.Errorf("MarkMergeInProgress failed with error: %v", err)
	}
	time.Sleep(time.Second * 10)
	matchPairs, err = FindAllMatches(tx, dbConfig, model.PaginationOptions{})
	if err != nil {
		t.Errorf("FindAllMatches failed with error: %v", err)
	}
	if len(matchPairs) != 1 {
		t.Errorf("FindAllMatches failed, expected 1 match after marking on as false positive, got %d: %+v", len(matchPairs), matchPairs)
		return
	}
	if !matchPairs[0].MergeInProgress {
		t.Errorf("Merge in porgress should be set to true")
	}

	fps := GetFalsePositives(dbConfig, 0, 10)
	if len(fps) != 2 {
		t.Errorf("GetFalsePositives failed, expected 2, got %d: %+v", len(fps), fps)
	}

	err = DeleteMatchPairs(tx, &dbConfig, domain, []model.MergeRq{{SourceProfileID: "s1", TargetProfileID: "t1"}})
	if err != nil {
		t.Errorf("DeleteMatchPairs failed with error: %v", err)
	}
	time.Sleep(time.Second * 10)
	matchPairs, err = FindAllMatches(tx, dbConfig, model.PaginationOptions{})
	if err != nil {
		t.Errorf("FindAllMatches failed with error: %v", err)
	}
	if len(matchPairs) != 0 {
		t.Errorf("DeleteMatchPairs failed, expected 0 match after deletion, got %d: %+v", len(matchPairs), matchPairs)
	}

	err = dbConfig.DeleteTable(tableName)
	if err != nil {
		t.Errorf("Error deleting table: %s", err)
	}

}
