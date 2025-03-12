// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	aurora "tah/upt/source/tah-core/aurora"
	"tah/upt/source/tah-core/cloudwatch"
	core "tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"time"
)

type IdentityResolutionHandler struct {
	AurSvc *aurora.PostgresDBConfig

	// Optional Aurora client that can be used for queries that don't need to run on the writer instance.
	// Currently, most of our workload is run directly on the writer. This is a pattern that we can continue to adopt and consider
	// initializing with every LCS handler.
	AurReaderSvc *aurora.PostgresDBConfig
	DynSvc       *db.DBConfig
	Tx           core.Transaction
	MetricLogger *cloudwatch.MetricLogger
	utils        LCSUtils
}

func (irh *IdentityResolutionHandler) SetTx(tx core.Transaction) {
	tx.LogPrefix = "customerprofiles-lc-irh"
	irh.Tx = tx
	irh.AurSvc.SetTx(tx)
	irh.DynSvc.SetTx(tx)
	irh.utils.SetTx(tx)
}

func NewIdentityResolutionHandler(aur aurora.PostgresDBConfig, dyn db.DBConfig, tx core.Transaction, cwm cloudwatch.MetricLogger) IdentityResolutionHandler {
	return IdentityResolutionHandler{
		AurSvc:       &aur,
		DynSvc:       &dyn,
		Tx:           tx,
		MetricLogger: &cwm,
		utils: LCSUtils{
			Tx: tx,
		},
	}
}

/**
* This function creates in initial IR table if rule based identity resolution is enabled
**/
func (i IdentityResolutionHandler) CreateIRTable(domain string) error {
	i.Tx.Info("Creating IR Table for %s", domain)
	_, err := i.AurSvc.Query(createIRTableSQL(domain))
	if err != nil {
		i.Tx.Error("Error creating IR Table: %v", err)
		return err
	}
	i.Tx.Info("Creating IR table index for rule index values")
	_, err = i.AurSvc.Query(createIRTableRuleIndex1Sql(domain))
	if err != nil {
		i.Tx.Error("Error creating IR Table rule index 1: %v", err)
		return err
	}
	_, err = i.AurSvc.Query(createIRTableRuleIndex2Sql(domain))
	if err != nil {
		i.Tx.Error("Error creating IR Table rule index 2: %v", err)
		return err
	}
	return nil
}

/**
* This function deletes the IR table
**/
func (i IdentityResolutionHandler) DeleteIRTable(domain string) error {
	i.Tx.Info("Deleting IR Table for %s", domain)
	_, err := i.AurSvc.Query(fmt.Sprintf(`DROP TABLE IF EXISTS %s`, irTableName(domain)))
	if err != nil {
		i.Tx.Error("Error deleting IR Table: %v", err)
		return err
	}
	return nil
}

/**
* This function empties the IR table
**/
func (i IdentityResolutionHandler) EmptyIRTable(domain string) error {
	i.Tx.Debug("Deleting IR Table for %s", domain)
	_, err := i.AurSvc.Query(fmt.Sprintf(`TRUNCATE TABLE %s`, irTableName(domain)))
	if err != nil {
		i.Tx.Error("Error deleting IR Table: %v", err)
		return err
	}
	return nil
}

/**
* This function finds matches in the IR table from an incoming object type, profile-level data, and a rule set
**/
func (i IdentityResolutionHandler) FindMatches(conn aurora.DBConnection, domain string, objectTypeName string, objectData map[string]string, profile profilemodel.Profile, ruleSet RuleSet) ([]RuleBasedMatch, error) {
	conn.Tx.Info("[FindMatches] FindMatches in IR Table %s for object type %s ", domain, objectTypeName)
	matches := []RuleBasedMatch{}
	conn.Tx.Debug("[FindMatches] Create index records for rule set")
	indexRecords := i.createRuleSetIndex(profile, objectTypeName, objectData, ruleSet)
	if len(indexRecords) == 0 {
		i.Tx.Error("[FindMatches] No records to index")
		return []RuleBasedMatch{}, nil
	}
	conn.Tx.Debug("[FindMatches] index records for rule set: %+v", indexRecords)

	sql, args, err := i.createFindMatchSql(domain, indexRecords, profile.ProfileId)
	if err != nil {
		conn.Tx.Error("[FindMatches] Error creating find match sql: %v", err)
		return []RuleBasedMatch{}, err
	}
	res, err := conn.Query(sql, args...)
	if err != nil {
		conn.Tx.Error("Error querying IR Table: %v", err)
		return []RuleBasedMatch{}, err
	}
	matchesByRule := map[string][]string{}
	for _, match := range res {
		ruleIndex := match[`rule`].(string)
		matchId := match[`connect_id`].(string)
		if ruleIndex == "" || matchId == "" {
			conn.Tx.Warn("the IR database table query return an invalid (rule,cid) pair: (%s, %s)", ruleIndex, matchId)
			continue
		}
		matchesByRule[ruleIndex] = append(matchesByRule[ruleIndex], matchId)
	}
	for rule, matchingCids := range matchesByRule {
		matches = append(matches, RuleBasedMatch{
			RuleIndex: rule,
			SourceID:  profile.ProfileId,
			MatchIDs:  matchingCids,
		})
	}

	return matches, nil
}

/**
* This function takes an incoming object type, object data and profile-level date and indexes it in the IR table based on provided rule set
* This means the data will be stored in a way that is efficient to query using the provided rule set
**/
func (i IdentityResolutionHandler) IndexInIRTable(conn aurora.DBConnection, domain string, objectTypeName string, objectData map[string]string, prof profilemodel.Profile, ruleSet RuleSet) error {
	conn.Tx.Info("[IndexInIRTable] Indexing in IR Table for %s", domain)
	indexRecords := i.createRuleSetIndex(prof, objectTypeName, objectData, ruleSet)
	if len(indexRecords) == 0 {
		conn.Tx.Error("[IndexInIRTable] No records to index")
		return nil
	}
	query, args := i.createIndexSql(domain, indexRecords)
	_, err := conn.Query(query, args...)
	if err != nil {
		conn.Tx.Error("[IndexInIRTable] Error deleting IR Table: %v", err)
		return err
	}
	return nil
}

/**
* This function return the content of the IR table limited by nRows. mostly to be used for testing and debugging
**/
func (i IdentityResolutionHandler) ShowIRTable(domain string, nFilterCol, nRows int) ([]map[string]interface{}, error) {
	i.Tx.Info("Showing IR Table for %s with %d filter columns and %d rows", domain, nFilterCol, nRows)
	filterFields := []string{}
	for i := 0; i < nFilterCol; i++ {
		filterFields = append(filterFields, filterColumnName(i))
	}
	data, err := i.AurSvc.Query(fmt.Sprintf(`SELECT connect_id,rule, index_value_1, index_value_2, timestamp, %s from %s LIMIT ?`, strings.Join(filterFields, `,`), irTableName(domain)), nRows)
	if err != nil {
		i.Tx.Error("Error showing IR Table: %v", err)
		return []map[string]interface{}{}, err
	}
	return data, nil
}

// This function updates the index after two profiles are merged without iterating on all object types. We reindex the profile in four steps:
// Step 1: update connect Ids for all index records
// Step 2: delete index value for all profile only rules
// Step 3: recreate index value for profile only rules for the new profile
// Step 4: check if the new profile triggers a profile-only rule (the new profile should never trigger an object-only rule since the rule would also apply to individual profiles)
// Note: for scalability reason we don't allow hybrid rules that have both profile and object level fields. if not we need to reindex every
// object in the profile with the new profile level attribute which is very expensive. this function should never iterate on all objects inside the merge profile
func (i IdentityResolutionHandler) ReindexProfileAfterMerge(domain string, connectId, connectIdToMerge string, objectData map[string]string, profile profilemodel.Profile, ruleSet RuleSet) error {
	conn, err := i.AurSvc.AcquireConnection(i.Tx)
	if err != nil {
		i.Tx.Error("[ReindexProfileAfterMerge] Error while acquiring connection: %v", err)
		return err
	}
	defer conn.Release()
	err = conn.StartTransaction()
	if err != nil {
		i.Tx.Error("[ReindexProfileAfterMerge] Error while starting transaction: %v", err)
		return err
	}
	//Step One: update connect Ids
	i.Tx.Debug("[ReindexProfileAfterMerge] removing duplicate indexes before updating connect ids")
	sql, args := deleteConflictingIndexesSql(domain, connectIdToMerge, connectId)
	_, err = conn.Query(sql, args...)
	if err != nil {
		conn.Tx.Error("[ReindexProfileAfterMerge] Error deleting conflicting indexes: %v", err)
		conn.RollbackTransaction()
		return err
	}
	i.Tx.Debug("[ReindexProfileAfterMerge] updating connect ids")
	_, err = conn.Query(updateConnectIdAfterMergeSql(domain), connectId, connectIdToMerge)
	if err != nil {
		conn.Tx.Error("[ReindexProfileAfterMerge] Error updating connect ID: %v", err)
		conn.RollbackTransaction()
		return err
	}
	profileOnlyRules, _ := i.organizeRules(ruleSet.Rules)
	//Step Two: update index value for profile only rules
	if len(profileOnlyRules) > 0 {
		i.Tx.Debug("[ReindexProfileAfterMerge] deleting profile-rule related level indexes (%d profile-only rules)", len(profileOnlyRules))
		err = i.deleteIndexesForRules(conn, domain, connectId, profileOnlyRules)
		if err != nil {
			conn.Tx.Error("[ReindexProfileAfterMerge] Error deleting index: %v", err)
			conn.RollbackTransaction()
			return err
		}
		//Step Three: recreating index for profile only rule
		pdm := prepareProfile(profile)
		indexesToApply := []RuleIndexRecord{}
		for _, rule := range profileOnlyRules {
			rec, applies := i.createRuleIndex(connectId, pdm, PROFILE_OBJECT_TYPE_NAME, objectData, rule)
			if applies {
				indexesToApply = append(indexesToApply, rec)
			}
		}
		i.Tx.Debug("[ReindexProfileAfterMerge] re-adding profile level index")
		query, args := i.createIndexSql(domain, indexesToApply)
		_, err = conn.Query(query, args...)
		if err != nil {
			conn.Tx.Error("[ReindexProfileAfterMerge] Error creating index in IR Table: %v", err)
			conn.RollbackTransaction()
			return err
		}
	} else {
		i.Tx.Debug("[ReindexProfileAfterMerge] no profile only rules to update")
	}

	err = conn.CommitTransaction()
	if err != nil {
		i.Tx.Error("[ReindexProfileAfterMerge] Error while committing transaction for profile indexing: %v", err)
		return err
	}
	return nil
}

// this function organize rules in 2 bucket: profile only and object only
func (i IdentityResolutionHandler) organizeRules(rules []Rule) (profileLevelRules []Rule, objectLevelRules []Rule) {
	for _, rule := range rules {
		if len(rule.Conditions) == 0 {
			i.Tx.Warn("Rule %d is an empty rule. It should not be support %v", rule.Index)
			continue
		}
		if rule.IsProfileLevel() {
			profileLevelRules = append(profileLevelRules, rule)
		} else if rule.IsObjectLevel() {
			objectLevelRules = append(objectLevelRules, rule)
		} else {
			i.Tx.Warn("Rule %d is an Hybrid rule. It should not be supported %v", rule.Index)
		}
	}
	return profileLevelRules, objectLevelRules
}

// There is an edge case where two profiles have a rule index that would cause a primary key collision on merge.
// PRIMARY KEY (connect_id, rule, index_value_1, index_value_2)
//
// Example:
//
// rule: customer_service_interaction related_loyalty_id = air_loyalty  id
// we ingest two customer service interactions with a related_loyalty_id but no matching air_loyalty ids (no merge occurs)
// we now have two indexes
// id1, rule_0, loyalty_id_123, empty_value
// id2, rule_0, loyalty_id_123, empty_value
//
// When we see a loyalty profile come in for loyalty_id_123 and merge these profiles, we might try to update id1 to id2, which violates our primary key.
// Instead, we first delete any indexes with this edge case on the source profile, so only one copy will exist after merge.
func deleteConflictingIndexesSql(domain, connectIdToMerge, connectId string) (query string, args []interface{}) {
	tableName := irTableName(domain)
	query = fmt.Sprintf(`
			DELETE FROM %s as idr
			WHERE connect_id = ?
			AND EXISTS (
				SELECT 1
				FROM %s AS sub
				WHERE
					idr.rule = sub.rule
					AND idr.index_value_1 = sub.index_value_1
					AND idr.index_value_2 = sub.index_value_2
					AND sub.connect_id IN (?, ?)
				GROUP BY
					sub.rule,
					sub.index_value_1,
					sub.index_value_2
				HAVING
					COUNT(DISTINCT sub.connect_id) = 2
			);
		`, tableName, tableName)
	args = append(args, connectIdToMerge, connectIdToMerge, connectId)
	return query, args
}

func updateConnectIdAfterMergeSql(domain string) string {
	return fmt.Sprintf(`
			UPDATE %s
			SET connect_id = ?
			WHERE connect_id = ?;
		`, irTableName(domain))
}

// Delete rule based match indexes for a given  UPT ID and rules indexes
func (i IdentityResolutionHandler) deleteIndexesForRules(conn aurora.DBConnection, domain string, uptId string, rules []Rule) error {
	query, args := i.deleteIndexesForRuleSql(domain, uptId, rules)
	_, err := conn.Query(query, args...)
	return err
}

// Delete rule based match indexes for a given list of UPT IDs
func (i IdentityResolutionHandler) deleteIndexes(conn aurora.DBConnection, domain string, uptIds []string) error {
	query, args := i.deleteIndexesSql(domain, uptIds)
	_, err := conn.Query(query, args...)
	return err
}

func (i IdentityResolutionHandler) deleteIndexesSql(domain string, uptIds []string) (query string, args []interface{}) {
	var paramPlaceholders []string
	for _, uptId := range uptIds {
		paramPlaceholders = append(paramPlaceholders, "?")
		args = append(args, uptId)
	}
	query = fmt.Sprintf(`
		DELETE FROM %s
		WHERE connect_id IN (%s);
	`, irTableName(domain), strings.Join(paramPlaceholders, ", "))
	return query, args
}

func (i IdentityResolutionHandler) deleteIndexesForRuleSql(domain string, uptId string, rules []Rule) (query string, args []interface{}) {
	var paramPlaceholders []string
	args = append(args, uptId)
	for _, rule := range rules {
		paramPlaceholders = append(paramPlaceholders, "?")
		args = append(args, strconv.Itoa(rule.Index))
	}
	query = fmt.Sprintf(`
		DELETE FROM %s
		WHERE connect_id=? AND rule IN (%s);
	`, irTableName(domain), strings.Join(paramPlaceholders, ", "))
	return query, args
}

/**
* This function create an array of RuleIndexRecord for a given interaction and Ruleset
* a RuleIndexRecord describes the fields used for indexing an object for matching purpose.
**/
func (i IdentityResolutionHandler) createRuleSetIndex(p profilemodel.Profile, profileObjectTypeName string, profileObject map[string]string, ruleSet RuleSet) []RuleIndexRecord {
	pdm := prepareProfile(p)
	var records []RuleIndexRecord

	for _, rule := range ruleSet.Rules {
		if rec, applies := i.createRuleIndex(p.ProfileId, pdm, profileObjectTypeName, profileObject, rule); applies {
			records = append(records, rec)
		}
	}
	return records
}

/**
* This function creates a RuleIndexRecord associated with a given Rule. this records allows indexing of an incoming
* interaction (and associated updated profile) into the IR table
* the process of creation of RuleIndexRecord index records follows the same steps than matching:
* 1/ we first process the SKIP condition. if a skip condition is met, we will not create the record and the function return false allowing the indexing process to stop
* 2/ the we process the MATCH conditions and create the values to be indexed. Noe tha this step normalized the fields and also compute template results is a template is used in the fields
* 3/ Finally this function processes the FILTER conditions and identify the extra fields needed in teh Filter section of teh IR table to support filtering
**/
func (i IdentityResolutionHandler) createRuleIndex(connectId string, pdm ProfileDataMap, profileObjectTypeName string, profileObject map[string]string, rule Rule) (RuleIndexRecord, bool) {
	//Run Match conditions
	i.Tx.Debug("[cid-%s][rule-%d] 1. Evaluating object %v against Skip conditions", connectId, rule.Index, profileObjectTypeName)
	for _, condition := range rule.Conditions {
		if condition.ConditionType == CONDITION_TYPE_SKIP && i.skipConditionApplies(condition, pdm, profileObjectTypeName, profileObject) {
			i.Tx.Error("[rule-%d][skip-cond-%d] object %v meets skip condition: stopping now!", rule.Index, condition.Index, profileObjectTypeName)
			return RuleIndexRecord{}, false
		}
	}
	i.Tx.Debug("[cid-%s][rule-%d] no Skip condition applies to object %s", connectId, rule.Index, profileObjectTypeName)

	//we need to create 2 indexes because we don't know which will arrive first (incoming or existing)
	i.Tx.Debug("[cid-%s][rule-%d] 2. create match condition indexes for object %v", connectId, rule.Index, profileObjectTypeName)
	matchConditionIndex1Segments := []string{}
	matchConditionIndex2Segments := []string{}
	//we use this map to ensure a field is not used
	usedFieldNames := map[string]bool{}
	usedFieldNames2 := map[string]bool{}

	for _, condition := range rule.Conditions {
		if condition.ConditionType == CONDITION_TYPE_MATCH {
			//matches is false if the field used for the match is empty in the incoming record
			seg, objectType, fieldName, matches := i.createIndexSegment(true, rule, condition, pdm, profileObjectTypeName, profileObject)
			seg2, objectType2, fieldName2, matches2 := i.createIndexSegment(false, rule, condition, pdm, profileObjectTypeName, profileObject)
			if matches && !usedFieldNames[objectType+fieldName] {
				usedFieldNames[objectType+fieldName] = true
				matchConditionIndex1Segments = append(matchConditionIndex1Segments, seg)
			}
			if matches2 && !usedFieldNames2[objectType2+fieldName2] {
				usedFieldNames2[objectType2+fieldName2] = true
				matchConditionIndex2Segments = append(matchConditionIndex2Segments, seg2)
			}
			if !matches && !matches2 {
				i.Tx.Debug("[cid-%s][rule-%d][match-cond-%s] Match condition does not apply to record %s because both fields %s.%s are empty %s.%s", connectId, rule.Index, condition.Index, profileObjectTypeName, condition.IncomingObjectType, condition.IncomingObjectField, condition.ExistingObjectType, condition.ExistingObjectField)
				return RuleIndexRecord{}, false
			}
		}
	}
	matchConditionIndexValue := strings.Join(matchConditionIndex1Segments, "_")
	matchConditionIndexValue2 := strings.Join(matchConditionIndex2Segments, "_")

	//we extract all the field involved in the filter conditions. we MUST keep the order of the condition in the rules
	// so that for a given rule, the field list is always in the same order (since our IR table use virtual column names in the filter section)
	i.Tx.Debug("[cid-%s][rule-%d] 3. extract filter fields from filter conditions", connectId, rule.Index)
	//we use this map to deduplicate fields
	filterFieldExists := map[string]bool{}
	filterFields := []FilterField{}
	filterConditions := []Condition{}
	for _, condition := range rule.Conditions {
		if condition.ConditionType == CONDITION_TYPE_FILTER {
			for _, field := range i.createFilterFields(rule, condition, pdm, profileObjectTypeName, profileObject) {
				if !filterFieldExists[field.Name] {
					i.Tx.Debug("[cid-%s][rule-%d][filter-cond-%d] Adding field %v to filter fields", connectId, rule.Index, condition.Index, field.Name)
					filterFieldExists[field.Name] = true
					filterFields = append(filterFields, field)
				} else {
					i.Tx.Debug("[cid-%s][rule-%d][filter-cond-%d] Skipping field %v for filter fields (already exists)", connectId, rule.Index, condition.Index, field.Name)
				}
			}
			filterConditions = append(filterConditions, condition)
		}
	}

	ts, err := parseTimestampFromProfileObjectData(profileObject)
	if err != nil {
		i.Tx.Warn("[cid-%s][rule-%d] Warning: could not parsing timestamp from profile object: %v. Defaulting to now", connectId, rule.Index, err)
		ts = time.Now()
	}

	rec := RuleIndexRecord{
		ConnectID:        connectId,
		RuleIndex:        rule.Index,
		RuleName:         rule.Name,
		IndexValue1:      matchConditionIndexValue,
		IndexValue2:      matchConditionIndexValue2,
		FilterFields:     filterFields,
		FilterConditions: filterConditions,
		Timestamp:        ts,
	}
	return rec, true
}

/**
* This function checks whether a skip condition applies to a rule and a profile object.
 */
func (i IdentityResolutionHandler) skipConditionApplies(condition Condition, pdm ProfileDataMap, profileObjectTypeName string, profileObject map[string]string) bool {
	if condition.ConditionType != CONDITION_TYPE_SKIP {
		i.Tx.Debug("[skip-cond-%d] Condition other than SKIP provided to the skipConditionApplies function. it does not apply", condition.Index)
		return false
	}
	if condition.IncomingObjectType != PROFILE_OBJECT_TYPE_NAME && condition.IncomingObjectType != profileObjectTypeName {
		i.Tx.Debug("[skip-cond-%d] Skip condition does not apply to record %s", condition.Index, profileObjectTypeName)
		return false
	}
	fieldValue := i.computeFieldValue(condition.IncomingObjectType, condition.IncomingObjectField, pdm, profileObjectTypeName, profileObject)
	fieldValue2 := i.computeFieldValue(condition.IncomingObjectType2, condition.IncomingObjectField2, pdm, profileObjectTypeName, profileObject)
	value := condition.SkipConditionValue
	if condition.Op == RULE_OP_EQUALS {
		i.Tx.Debug("[skip-cond-%d] checking skip condition for %v with operator RULE_OP_EQUALS", condition.Index, condition.IncomingObjectType)
		return fieldValue == fieldValue2
	}
	if condition.Op == RULE_OP_EQUALS_VALUE {
		i.Tx.Debug("[skip-cond-%d] checking skip condition for %v with operator RULE_OP_EQUALS_VALUE", condition.Index, condition.IncomingObjectType)
		return fieldValue == value.StringValue
	}
	if condition.Op == RULE_OP_NOT_EQUALS {
		i.Tx.Debug("[skip-cond-%d] checking skip condition for %v with operator RULE_OP_NOT_EQUALS", condition.Index, condition.IncomingObjectType)
		return fieldValue != fieldValue2
	}
	if condition.Op == RULE_OP_NOT_EQUALS_VALUE {
		i.Tx.Debug("[skip-cond-%d] checking skip condition for %v with operator RULE_OP_NOT_EQUALS_VALUE", condition.Index, condition.IncomingObjectType)
		return fieldValue != value.StringValue
	}
	if condition.Op == RULE_OP_MATCHES_REGEXP {
		i.Tx.Debug("[skip-cond-%d] checking skip condition for %v with operator RULE_OP_MATCHES_REGEXP", condition.Index, condition.IncomingObjectType)
		//this should be validated upon submission
		match, err := regexp.MatchString(value.StringValue, fieldValue)
		if err != nil {
			i.Tx.Info("[skip-cond-%d] Skip condition %v failed to apply because of invalid Regexp %s. error: %v", condition.Index, value.StringValue, err)
			return false
		}
		return match
	}
	i.Tx.Info("[skip-cond-%d] skip condition %+v does not apply to profile after interaction %v", condition.Index, condition, profileObjectTypeName)
	return false
}

/**
* This function creates the index segment corresponding to a match condition. it returns the following values:
* 1/ the index segment created from the field name, execution of an optional template  and application of the normalization process specified in teh condition
* 2/ the object type that the condition applies to
* 3/ the field name name applicable to the condition
* 4/ true if the field value is not empty (and therefore the condition applies to the object)
* Note tha this function has 2 modes:
* - if incoming = true we create the segment value form the incoming record
* - if incoming = false we cerate the segment from the Existing records if provided. if not provided we use the incoming record (this allows a condition to only specify the incoming record (for example match on FirstName))
 */
func (i IdentityResolutionHandler) createIndexSegment(incoming bool, rule Rule, condition Condition, pdm ProfileDataMap, profileObjectTypeName string, profileObject map[string]string) (string, string, string, bool) {
	if condition.ConditionType != CONDITION_TYPE_MATCH {
		i.Tx.Debug("[rule-%d][match-cond-%d] Condition other than MATCH provided to the createIndexSegment function. it does not apply", rule.Index, condition.Index)
		return "", "", "", false
	}

	i.Tx.Debug("[rule-%d][match-cond-%d] creating index segment for object type %v and field %v with incoming=%v", rule.Index, condition.Index, condition.IncomingObjectType, condition.IncomingObjectField, incoming)
	field := ""
	objectType := ""
	//if the condition does not provide an existinObjectField, it means we are matching incoming and existing on teh same field
	//if we are not in search mode, we also should use the incomingObject
	if !incoming {
		i.Tx.Debug("[rule-%d][match-cond-%d] using existing field (%s.%s)", rule.Index, condition.Index, condition.ExistingObjectType, condition.ExistingObjectField)
		field = condition.ExistingObjectField
		objectType = condition.ExistingObjectType
		//if the condition does not provide an existinObjectField, it means we are using incoming to specify both incoming and existing (ex match on FirstName)
		if field == "" {
			i.Tx.Debug("[rule-%d][match-cond-%d] Condition expected field is not provided. using the incoming field name")
			field = condition.IncomingObjectField
			objectType = condition.IncomingObjectType
		}
	} else {
		i.Tx.Debug("[rule-%d][match-cond-%d] using incoming field (%s.%s) ", rule.Index, condition.Index, condition.IncomingObjectType, condition.IncomingObjectField)
		field = condition.IncomingObjectField
		objectType = condition.IncomingObjectType
	}
	value := i.computeFieldValue(objectType, field, pdm, profileObjectTypeName, profileObject)
	if strings.Trim(value, " ") != "" {
		i.Tx.Debug("[rule-%d][match-cond-%d] index segment found for object type %v and field %v", rule.Index, condition.Index, objectType, field)
		return normalizeForIRIndex(condition.IndexNormalization, value), objectType, field, true
	}
	return "", "", "", false
}

/**
* this function normalize the value for index creation based on match condition normalization settings
* example: phone numbers should be normalized to remove all non-numeric characters and leading zeros
**/
func normalizeForIRIndex(settings IndexNormalizationSettings, value string) string {
	normalizedValue := value
	if settings.Trim {
		normalizedValue = strings.Trim(value, " ")
	}
	if settings.Lowercase {
		normalizedValue = strings.ToLower(normalizedValue)
	}
	if settings.RemoveNonAlphaNumeric {
		normalizedValue = regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAllString(normalizedValue, "")
	}
	if settings.RemoveNonNumeric {
		normalizedValue = regexp.MustCompile("[^0-9]+").ReplaceAllString(normalizedValue, "")
	}
	if settings.RemoveSpaces {
		normalizedValue = strings.ReplaceAll(normalizedValue, " ", "")
	}
	if settings.RemoveLeadingZeros {
		normalizedValue = strings.TrimLeft(normalizedValue, "0")
	}
	return normalizedValue
}

/**
* This function lists all the field names used in a Filter condition. this is used to identify the DB columns to be filled in the
* FILTER section of teh IR table
**/
func (i IdentityResolutionHandler) createFilterFields(rule Rule, condition Condition, pdm ProfileDataMap, profileObjectTypeName string, profileObject map[string]string) []FilterField {
	if condition.ConditionType != CONDITION_TYPE_FILTER {
		i.Tx.Debug("[rule-%d][filter-cond-%d] Condition other than FILTER provided to the createFilterFields function. it does not apply", rule.Index, condition.Index)
		return []FilterField{}
	}

	i.Tx.Debug("[rule-%d][filter-cond-%d] Extracting filter fields from condition %v", rule.Index, condition.Index, condition)
	filterFields := []FilterField{}
	incoming := i.computeFieldValue(condition.IncomingObjectType, condition.IncomingObjectField, pdm, profileObjectTypeName, profileObject)
	existing := i.computeFieldValue(condition.ExistingObjectType, condition.ExistingObjectField, pdm, profileObjectTypeName, profileObject)
	//We exclude conditions related to timestamps since timestamps are part of the main part of the IR table
	if !isTimestampCondition(condition.Op) {
		filterFields = append(filterFields, FilterField{
			Name:  filterFieldName(condition.IncomingObjectType, condition.IncomingObjectField),
			Value: incoming,
		})
		filterFields = append(filterFields, FilterField{
			Name:  filterFieldName(condition.ExistingObjectType, condition.ExistingObjectField),
			Value: existing,
		})
	}

	return filterFields
}

/**
* This function taks an object type, field name , ProfileDataMap, profileObjectTypeName and the actual object data (profileObject)
* and returns the value of the field for the object.
* The value of the field is determined by either locating the data in the ProfileDataMap (if the condition is not a template)
* or by executing the template logic (if the condition is a template)
*
* example:
* if the condition is a template:
* {{.FirstName}}
**/
func (i IdentityResolutionHandler) computeFieldValue(objectType, fieldName string, pdm ProfileDataMap, profileObjectTypeName string, profileObject map[string]string) string {
	i.Tx.Debug("[computeFieldValue] Computing field value for object %s and field %s", objectType, fieldName)
	if !strings.Contains(fieldName, "{{") {
		i.Tx.Debug("[computeFieldValue] Field name is not a template. ")
		return getValue(objectType, fieldName, pdm, profileObjectTypeName, profileObject)
	}
	var content bytes.Buffer
	t, err := template.New("fieldTemplate").Parse(fieldName)
	if err != nil {
		i.Tx.Warn("Error parsing template name: %v. returning getValue", err)
		return getValue(objectType, fieldName, pdm, profileObjectTypeName, profileObject)
	}
	if objectType == PROFILE_OBJECT_TYPE_NAME {
		//return pdm.GetProfileValue(fieldName)
		err2 := t.Execute(&content, pdm.Data)
		if err2 != nil {
			i.Tx.Warn("[computeFieldValue] Error executing template logic: %v. returning", err2)
			return getValue(objectType, fieldName, pdm, profileObjectTypeName, profileObject)
		}
	} else if objectType == profileObjectTypeName {
		//return profileObject[fieldName]
		err2 := t.Execute(&content, profileObject)
		if err2 != nil {
			i.Tx.Warn("[computeFieldValue] Error executing template logic: %v returning", err2)
			return getValue(objectType, fieldName, pdm, profileObjectTypeName, profileObject)
		}
	}
	return content.String()
}

/**
* This function gets the value of a field based on object data and object type.
 */
func getValue(objectType, fieldName string, pdm ProfileDataMap, profileObjectTypeName string, profileObject map[string]string) string {
	if objectType == PROFILE_OBJECT_TYPE_NAME {
		return pdm.GetProfileValue(fieldName)
	} else if objectType == profileObjectTypeName {
		return profileObject[fieldName]
	}
	return ""
}

// ///////////////////////////////////////////
//
//	SQL FUNCTIONS
//
// ////////////////////////////////////////

func createIRTableSQL(domain string) string {
	fields := []string{
		`connect_id VARCHAR(255) NOT NULL`,
		`rule VARCHAR(255) NOT NULL`,
		`index_value_1 VARCHAR(255) NOT NULL`,
		`index_value_2 VARCHAR(255) NOT NULL`,
		`timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`,
		`PRIMARY KEY (connect_id, rule, index_value_1, index_value_2)`,
	}
	//a condition can only refer to 2 fields, this means we will have a maximum of 2*MAX_CONDITIONS_PER_RULES
	for i := 0; i < 2*MAX_CONDITIONS_PER_RULES; i++ {
		fields = append(fields, fmt.Sprintf(`%s VARCHAR(255)`, filterColumnName(i)))
	}
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS `+irTableName(domain)+` (%s);`, strings.Join(fields, ","))
}

func createIRTableRuleIndex1Sql(domain string) string {
	return fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s_idx_rule_index_value1_connect_id 
		ON %s (rule, index_value_1, connect_id)
	`, domain, irTableName(domain))
}

func createIRTableRuleIndex2Sql(domain string) string {
	return fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s_idx_rule_index_value2_connect_id
		ON %s (rule, index_value_2, connect_id)
	`, domain, irTableName(domain))
}

func (i IdentityResolutionHandler) createFindMatchSql(domain string, indexRecords []RuleIndexRecord, uptId string) (string, []interface{}, error) {
	i.Tx.Debug("[createFindMatchSql] Creating SQL query for matching")
	whereClauseParts := []string{}
	args := []interface{}{}
	for _, rec := range indexRecords {
		colMap := filterFieldToColumns(rec.FilterFields)
		filterConditions, err := i.createFilterConditionClauses(rec, colMap)
		if err != nil {
			i.Tx.Error("[rule-%d] Error: %v", rec.RuleIndex, err)
			return "", nil, err
		}
		//index condition is guaranteed to have a value because if both IndexValue2 and IndexValue1 are empty, the records doe snot match
		//so we would not do the search
		indexCondition := ""
		if rec.IndexValue2 != "" {
			args = append(args, rec.IndexValue2)
			indexCondition = "index_value_1 = ?"
		} else {
			args = append(args, rec.IndexValue1)
			indexCondition = "index_value_2 = ?"
		}
		if len(filterConditions) == 0 {
			whereClauseParts = append(whereClauseParts, fmt.Sprintf(`rule = '%d' AND (%s)`, rec.RuleIndex, indexCondition))
		} else {
			whereClauseParts = append(whereClauseParts, fmt.Sprintf(`rule = '%d' AND (%s) AND (%s)`, rec.RuleIndex, indexCondition, strings.Join(filterConditions, " AND ")))
		}
	}
	whereClause := "(" + strings.Join(whereClauseParts, `) OR (`) + ")"
	sql := fmt.Sprintf(`
		SELECT connect_id, rule
		FROM %s
		WHERE (%s)
		AND NOT connect_id = '%s';
		`, irTableName(domain), whereClause, uptId)

	return formatQuery(sql), args, nil
}

// Max rule index record is the number of rules
func (i IdentityResolutionHandler) createIndexSql(domain string, recs []RuleIndexRecord) (string, []interface{}) {
	var combinedValues []string
	var combinedArgs []interface{}
	maxNumFields := maxNumberOfFilterFields(recs)
	for _, rec := range recs {
		values, args := createRuleIndexValuesSql(rec, maxNumFields)
		combinedValues = append(combinedValues, strings.Join(values, ", "))
		combinedArgs = append(combinedArgs, args...)
	}
	//we create the list of filter fields only if filter fields are needed
	filterFieldArr := buildFilterFieldsArray(maxNumFields)
	filterFields := ""
	if len(filterFieldArr) > 0 {
		filterFields = " " + strings.Join(filterFieldArr, ",")
	}

	sql := fmt.Sprintf(`
		INSERT INTO %s (connect_id, rule, index_value_1, index_value_2, timestamp%s)
		VALUES (%s)
		ON CONFLICT (connect_id, rule, index_value_1, index_value_2)
		DO NOTHING;
	`, irTableName(domain), filterFields, strings.Join(combinedValues, "), ("))
	return formatQuery(sql), combinedArgs
}

// returns the name of the ith filter column
func filterColumnName(i int) string {
	return fmt.Sprintf(`%s%d`, FILTER_FIELD_PREFIX, i)
}

func (i IdentityResolutionHandler) createFilterConditionClauses(rec RuleIndexRecord, columnMap map[string]string) ([]string, error) {
	clauses := []string{}
	for _, cond := range rec.FilterConditions {
		clause, err := i.createFilterConditionClause(rec.Timestamp, cond, columnMap)
		if err != nil {
			return clauses, err
		}
		clauses = append(clauses, clause)
	}
	return clauses, nil
}

func filterFieldToColumns(filterFields []FilterField) map[string]string {
	colMap := map[string]string{}
	for i, field := range filterFields {
		colMap[field.Name] = filterColumnName(i)
	}
	return colMap
}

func (i IdentityResolutionHandler) createFilterConditionClause(incomingRecordTs time.Time, condition Condition, columnMap map[string]string) (string, error) {
	incomingField := filterFieldName(condition.IncomingObjectType, condition.IncomingObjectField)
	existingField := filterFieldName(condition.ExistingObjectType, condition.ExistingObjectField)
	operators := map[string]string{
		RULE_OP_EQUALS:     "=",
		RULE_OP_NOT_EQUALS: "!=",
	}
	if condition.Op == RULE_OP_EQUALS || condition.Op == RULE_OP_NOT_EQUALS {
		return fmt.Sprintf(`("%s" %s '%s')`, columnMap[incomingField], operators, columnMap[existingField]), nil
	}
	if condition.Op == RULE_OP_WITHIN_SECONDS {
		start := incomingRecordTs.Add(-time.Duration(condition.FilterConditionValue.IntValue) * time.Second)
		end := incomingRecordTs.Add(time.Duration(condition.FilterConditionValue.IntValue) * time.Second)
		return fmt.Sprintf(`("timestamp" >= timestamp '%s' AND "timestamp" <= timestamp '%s')`, start.Format(TIMESTAMP_FORMAT), end.Format(TIMESTAMP_FORMAT)), nil
	}
	return "", fmt.Errorf("Rule Operator %s is not supported", condition.Op)
}

func (i IdentityResolutionHandler) selectProfilesToSql(domain string, rule Rule, mappings []ObjectMapping) (string, error) {
	profileColumns := i.utils.mappingsToProfileDbColumns(mappings)
	for i, col := range profileColumns {
		profileColumns[i] = fmt.Sprintf(`m."%s"`, col)
	}
	skipConditions := skipConditionsToSqlClauses(PROFILE_OBJECT_TYPE_NAME, rule)
	return fmt.Sprintf(`SELECT %s FROM %s m WHERE %s`,
		strings.Join(profileColumns, ", "),
		searchIndexTableName(domain),
		strings.Join(skipConditions, " AND ")), nil
}

func (i IdentityResolutionHandler) selectInteractionProfilesToSql(domain string, objectTypeName string, rule Rule, mappings []ObjectMapping) (string, error) {
	profileColumns, objectColumnsTypes := i.utils.MappingsToDbColumns(mappings)
	objectColumns, ok := objectColumnsTypes[objectTypeName]
	if !ok {
		return "", errors.New("objectTypeName not found in mappings")
	}

	for i, col := range objectColumns {
		objectColumns[i] = fmt.Sprintf(`i."%s"`, col)
	}
	for i, col := range profileColumns {
		profileColumns[i] = fmt.Sprintf(`m."%s"`, col)
	}
	skipConditions := skipConditionsToSqlClauses(objectTypeName, rule)
	interactionTableName := createInteractionTableName(domain, objectTypeName)
	return fmt.Sprintf(`SELECT %s, %s FROM %s i LEFT JOIN %s m ON i.connect_id = m.connect_id WHERE %s`,
		strings.Join(objectColumns, ", "),
		strings.Join(profileColumns, ", "),
		interactionTableName,
		searchIndexTableName(domain),
		strings.Join(skipConditions, " AND ")), nil
}

func buildFilterFieldsArray(n int) []string {
	fields := []string{}
	for i := 0; i < n; i++ {
		fields = append(fields, filterColumnName(i))
	}
	return fields
}

func maxNumberOfFilterFields(recs []RuleIndexRecord) int {
	max := 0
	for _, rec := range recs {
		if len(rec.FilterFields) > max {
			max = len(rec.FilterFields)
		}
	}
	return max
}

func createRuleIndexValuesSql(rec RuleIndexRecord, maxNumFields int) ([]string, []interface{}) {
	values := []string{
		fmt.Sprintf(`'%s'`, rec.ConnectID),
		fmt.Sprintf(`'%d'`, rec.RuleIndex),
		"?", // IndexValue1
		"?", // IndexValue2
		fmt.Sprintf(`'%s'`, rec.Timestamp.Format(TIMESTAMP_FORMAT)),
	}
	args := []interface{}{rec.IndexValue1, rec.IndexValue2}

	//padding hte value list with empty value to match the max filter field count
	for i := len(rec.FilterFields); i < maxNumFields; i++ {
		values = append(values, "")
	}

	return values, args
}

func irTableName(domain string) string {
	return fmt.Sprintf(`%s_rule_index`, domain)
}

func IntValue(val int) Value {
	return Value{
		ValueType: CONDITION_VALUE_TYPE_INT,
		IntValue:  val,
	}
}

func StringVal(val string) Value {
	return Value{
		ValueType:   CONDITION_VALUE_TYPE_STRING,
		StringValue: val,
	}
}

// this function iterates on all rules in the current ruleset and return rule_id-object_type pairs
// this is used to trigger separate fargate tasks
func createRuleObjectPairs(ruleSet RuleSet) []RuleObjectPair {
	pairMap := map[string]RuleObjectPair{}
	pairs := []RuleObjectPair{}
	for _, rule := range ruleSet.Rules {
		ruleIndex := fmt.Sprintf("%d", rule.Index)
		for _, cond := range rule.Conditions {
			if cond.ConditionType == CONDITION_TYPE_MATCH {
				pairMap[ruleIndex+cond.IncomingObjectType] = RuleObjectPair{RuleID: ruleIndex, ObjectType: cond.IncomingObjectType}
				pairMap[ruleIndex+cond.ExistingObjectType] = RuleObjectPair{RuleID: ruleIndex, ObjectType: cond.ExistingObjectType}
			}
		}
	}
	for _, pair := range pairMap {
		pairs = append(pairs, pair)
	}
	return pairs
}

func skipConditionsToSqlClauses(objectTypeName string, rule Rule) []string {
	fieldMap := map[string]string{}
	fieldNames := []string{}
	clauses := []string{}
	prefix := "i"
	if objectTypeName == PROFILE_OBJECT_TYPE_NAME {
		prefix = "m"
	}
	for _, cond := range rule.Conditions {

		if cond.ConditionType == CONDITION_TYPE_SKIP && cond.IncomingObjectType == objectTypeName {
			clauses = append(clauses, addWereClauseFromSkipCondition(cond))
		}
		if cond.ConditionType == CONDITION_TYPE_MATCH {
			if cond.IncomingObjectType == objectTypeName {
				for _, field := range parseFieldName(cond.IncomingObjectField) {
					if _, ok := fieldMap[field]; !ok {
						fieldNames = append(fieldNames, field)
						fieldMap[field] = field
					}
				}
			}
			if cond.ExistingObjectType == objectTypeName {
				for _, field := range parseFieldName(cond.ExistingObjectField) {
					if _, ok := fieldMap[field]; !ok {
						fieldNames = append(fieldNames, field)
						fieldMap[field] = field
					}
				}
			}
		}
	}
	for _, field := range fieldNames {
		clauses = append(clauses, fmt.Sprintf(`%s."%s" != ''`, prefix, field))
	}
	return clauses
}

func parseFieldName(fieldValue string) []string {
	if isTemplate(fieldValue) {
		return parseFieldNamesFromTemplate(fieldValue)
	}
	return []string{fieldValue}
}

func addWereClauseFromSkipCondition(cond Condition) string {
	//Note the equals and not equals are reversed since it is a SKIP condition (meaning we only process the object it it is FALSE)
	operator := map[string]string{
		RULE_OP_EQUALS:     "!=",
		RULE_OP_NOT_EQUALS: "=",
	}
	f1 := fieldToDbColumn(cond.IncomingObjectType, cond.IncomingObjectField)
	f2 := fieldToDbColumn(cond.IncomingObjectType2, cond.IncomingObjectField2)
	return fmt.Sprintf(`%s %s %s`, f1, operator[cond.Op], f2)

}

func fieldToDbColumn(objectType string, fieldName string) string {
	if objectType == PROFILE_OBJECT_TYPE_NAME {
		return "m." + fieldName
	}
	return fmt.Sprintf(`i.%s`, fieldName)
}
