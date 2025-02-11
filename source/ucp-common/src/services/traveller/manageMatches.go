// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"errors"
	"log"
	"sort"
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/admin"
)

var ERROR_MSG_PROFILE_PAIR_DOESNT_EXIST = "profile pair does not exist"

func SearchMatches(matchDb db.DBConfig, domainName string, inputId string) ([]model.MatchPair, error) {
	data := []model.MatchPair{}
	filter := db.DynamoFilterExpression{
		EqualBool: []db.DynamoBoolFilterCondition{{Key: "falsePositive", Value: false}},
	}
	err := matchDb.FindStartingWithAndFilter(domainName+"_"+inputId, constant.MATCH_PREFIX, &data, filter)
	return data, err
}

// this function returns all match pairs in order of confidence score
func FindAllMatches(tx core.Transaction, matchDb db.DBConfig, po model.PaginationOptions) ([]model.MatchPair, error) {
	data := []model.MatchPair{}
	// the index has 2 records for each pair of matches (since we create one source-target and one target-source)
	// index does  not ensure hash key uniqueness To solve this, we retreive tiwce as many profiles and remove the duplicates
	// as post filtering
	tx.Info("[FindAllMatches] Search matches with pagination settings: page %v of size %v", po.Page, po.PageSize)
	dqo := db.QueryOptions{
		Index: db.DynamoIndex{
			Name: constant.MATCH_INDEX_NAME,
			Pk:   constant.MATCH_INDEX_PK_NAME,
			Sk:   constant.MATCH_INDEX_SK_NAME,
		},
		ReverseOrder: true,
		PaginOptions: db.DynamoPaginationOptions{
			Page:     int64(po.Page),
			PageSize: int32(po.PageSize),
		},
	}
	tx.Debug("[FindAllMatches] Dynamo query options: %+v", dqo)
	err := matchDb.FindStartingWithAndFilterWithIndex(constant.MATCH_LATEST_RUN, constant.MATCH_PREFIX, &data, dqo)
	if err != nil {
		return data, err
	}
	tx.Debug("[FindAllMatches] Total items in dynamo before removing duplicates:  %v", len(data))
	data = filterDuplicates(data)
	tx.Info("[FindAllMatches] %v items remaining after filtering duplicates", len(data))
	return data, err
}

func GetFalsePositives(matchDb db.DBConfig, page, pageSize int64) []model.MatchPair {
	dqo := db.QueryOptions{
		Index: db.DynamoIndex{
			Name: constant.MATCH_INDEX_NAME,
			Pk:   constant.MATCH_INDEX_PK_NAME,
			Sk:   constant.MATCH_INDEX_SK_NAME,
		},
		ReverseOrder: true,
		Filter: db.DynamoFilterExpression{
			EqualBool: []db.DynamoBoolFilterCondition{{Key: "falsePositive", Value: true}},
		},
		PaginOptions: db.DynamoPaginationOptions{
			Page:     page,
			PageSize: int32(pageSize),
		},
	}

	fps := []model.MatchPair{}
	err := matchDb.FindStartingWithAndFilterWithIndex(constant.MATCH_LATEST_RUN, "fp_"+constant.MATCH_PREFIX, &fps, dqo)
	if err != nil {
		log.Printf("[GetFalsePositives] error: %v", err)
	}
	return fps
}

func GetMatchPair(tx core.Transaction, matchDb db.DBConfig, domainName string, sourceId string, targetId string) (model.MatchPair, error) {
	tx.Info("[GetMatchPair] searching for match pair %v-%v", sourceId, targetId)
	pk := buildPk(domainName, sourceId)
	sk := buildSk(targetId)
	mp := model.MatchPair{}
	err := matchDb.Get(pk, sk, &mp)
	if err != nil {
		tx.Error("[GetMatchPair] error: %v", err)
	}
	return mp, err
}

func filterDuplicates(data []model.MatchPair) []model.MatchPair {
	filtered := []model.MatchPair{}
	mpMap := map[string]bool{}
	for _, mp := range data {
		if !mpMap[mp.ScoreTargetID] {
			filtered = append(filtered, mp)
			mpMap[mp.ScoreTargetID] = true
		}
	}
	return filtered
}

func MarkFalsePositive(tx core.Transaction, matchDb db.IDBConfig, domain string, sourceId, targetId string) error {
	return MarkFalsePositives(tx, matchDb, domain, []model.MergeRq{{SourceProfileID: sourceId, TargetProfileID: targetId}})
}

func MarkFalsePositives(tx core.Transaction, matchDb db.IDBConfig, domain string, rqs []model.MergeRq) error {
	tx.Info("[MarkFalsePositives] marking %v profile pairs as false positive", len(rqs))
	mps := []model.MatchPair{}
	errs := []error{}
	for _, rq := range rqs {
		sourceId := rq.SourceProfileID
		targetId := rq.TargetProfileID
		tx.Debug("Marking profile pair (%s, %s) as false positive", sourceId, targetId)

		matchPairSource := model.MatchPair{}
		err := matchDb.Get(domain+"_"+sourceId, constant.MATCH_PREFIX+targetId, &matchPairSource)
		if err != nil {
			tx.Error("[MarkFalsePositives] Error getting profile pair (%s, %s) from dynamo", sourceId, targetId)
			errs = append(errs, err)
		}

		matchPairTarget := model.MatchPair{}
		err = matchDb.Get(domain+"_"+targetId, constant.MATCH_PREFIX+sourceId, &matchPairTarget)
		if err != nil {
			tx.Error("[MarkFalsePositives] Error getting profile pair (%s, %s) from dynamo", targetId, sourceId)
			errs = append(errs, err)
		}

		if matchPairTarget.TargetProfileID == "" && matchPairSource.TargetProfileID == "" {
			tx.Error("[MarkFalsePositives] Profile pair (%s, %s) does not exist", sourceId, targetId)
			errs = append(errs, errors.New(ERROR_MSG_PROFILE_PAIR_DOESNT_EXIST))
		}
		if matchPairTarget.TargetProfileID != "" {
			mps = append(mps, UpdateFalsePositiveRecord(matchPairTarget))
		}
		if matchPairSource.TargetProfileID != "" {
			mps = append(mps, UpdateFalsePositiveRecord(matchPairSource))
		}
	}
	errs = append(errs, matchDb.SaveMany(mps))
	return core.ErrorsToError(errs)
}

func UpdateFalsePositiveRecord(mp model.MatchPair) model.MatchPair {
	//by prefixing false positives with fp we make sure they never show up in the index serach which simplifies pagination implementation
	mp.ScoreTargetID = "fp_" + mp.ScoreTargetID
	mp.FalsePositive = true
	return mp
}

func MarkMergeInProgress(tx core.Transaction, matchDb db.IDBConfig, domain string, rqs []model.MergeRq) error {
	tx.Info("[MarkMergeInProgress] marking %v profile pairs as 'MergeInProgress", len(rqs))
	mps := []model.MatchPair{}
	errs := []error{}
	for _, rq := range rqs {
		sourceId := rq.SourceProfileID
		targetId := rq.TargetProfileID
		tx.Debug("Marking profile pair (%s, %s) as MergeInProgress", sourceId, targetId)

		matchPairSource := model.MatchPair{}
		err := matchDb.Get(domain+"_"+sourceId, constant.MATCH_PREFIX+targetId, &matchPairSource)
		if err != nil {
			tx.Error("[MarkMergeInProgress] Error getting profile pair (%s, %s) from dynamo", sourceId, targetId)
			errs = append(errs, err)
		}

		matchPairTarget := model.MatchPair{}
		err = matchDb.Get(domain+"_"+targetId, constant.MATCH_PREFIX+sourceId, &matchPairTarget)
		if err != nil {
			tx.Error("[MarkMergeInProgress] Error getting profile pair (%s, %s) from dynamo", targetId, sourceId)
			errs = append(errs, err)
		}

		if matchPairSource.TargetProfileID == "" && matchPairTarget.TargetProfileID == "" {
			tx.Error("[MarkMergeInProgress] Profile pair (%s, %s) does not exist", sourceId, targetId)
			errs = append(errs, errors.New(ERROR_MSG_PROFILE_PAIR_DOESNT_EXIST))
		}
		if matchPairSource.TargetProfileID != "" {
			matchPairSource.MergeInProgress = true
			mps = append(mps, matchPairSource)
		}

		if matchPairTarget.TargetProfileID != "" {
			matchPairTarget.MergeInProgress = true
			mps = append(mps, matchPairTarget)
		}
	}
	tx.Info("[MarkMergeInProgress] Saving %v items to dynamo", len(mps))
	errs = append(errs, matchDb.SaveMany(mps))
	return core.ErrorsToError(errs)
}

func DeleteMatchPairs(tx core.Transaction, matchDb db.IDBConfig, domain string, rqs []model.MergeRq) error {
	tx.Info("[DeleteMatchPairs] deleting %v match pairs", len(rqs))
	mps := []model.MatchPairForDelete{}
	for _, rq := range rqs {
		sourceId := rq.SourceProfileID
		targetId := rq.TargetProfileID
		tx.Debug("Adding Profile pair (%s, %s) for deletion", sourceId, targetId)
		mps = append(mps, model.MatchPairForDelete{
			Pk: buildPk(domain, sourceId),
			Sk: buildSk(targetId),
		})
		tx.Debug("Adding Profile pair (%s, %s) for deletion", targetId, sourceId)
		mps = append(mps, model.MatchPairForDelete{
			Pk: buildPk(domain, targetId),
			Sk: buildSk(sourceId),
		})
	}
	tx.Info("[DeleteMatchPairs] Deleting %v items from dynqmo", len(mps))
	return matchDb.DeleteMany(mps)

}

// note: match pair records built with this function will not have diffdata
func BuildMatchPairRecord(sourceId, targetId string, confidenceScore string, domain string, mergeInProgress bool) model.MatchPair {
	return model.MatchPair{
		Pk:              buildPk(domain, sourceId),
		Sk:              buildSk(targetId),
		TargetProfileID: targetId,
		Score:           confidenceScore,
		Domain:          domain,
		//we use these attributes for the global secondary index used to return all matches sorted by score
		RunID:           constant.MATCH_LATEST_RUN,
		ScoreTargetID:   buildScoreTargetId(confidenceScore, sourceId, targetId),
		MergeInProgress: mergeInProgress,
	}
}

func buildPk(domain, sourceId string) string {
	return domain + "_" + sourceId
}

func buildSk(targetId string) string {
	return constant.MATCH_PREFIX + targetId
}

func buildScoreTargetId(confidenceScore, sourceId, targetId string) string {
	return constant.MATCH_PREFIX + confidenceScore + "_" + buildUniquePairKey(sourceId, targetId)
}

// this function builds aunique key per pair by sorting and ocncatenating Ids
// this allows the index to be populated with only one record per pair
func buildUniquePairKey(id1 string, id2 string) string {
	ids := []string{id1, id2}
	sort.Strings(ids)
	return strings.Join(ids, "_")
}
