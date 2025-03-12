// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	core "tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"time"
)

// creating this struct to convey the transaction id for logging.
type LCSUtils struct {
	Tx core.Transaction
}

func (u *LCSUtils) SetTx(tx core.Transaction) {
	u.Tx = tx
}

//Profile data map

/*
* This function creates a ProfileDataMap from Profile. ProfileDataMap is used to quickly access values withing a profile
* by field name
***/
func prepareProfile(p profilemodel.Profile) ProfileDataMap {
	pdm := &ProfileDataMap{}
	for _, f := range RESERVED_FIELDS {
		pdm.SetProfileValue(f, p.Get(f))
	}
	for f, v := range p.Attributes {
		pdm.SetProfileValue("Attributes."+f, v)
	}
	return *pdm
}

func parseTimestampFromProfileObjectData(po map[string]string) (time.Time, error) {
	return time.Parse(TIMESTAMP_FORMAT, po[PROFILE_OBJECT_FIELD_NAME_TIMESTAMP])
}

// this function hecks whether a condition applies to timestamp. in the future we may want to use more OP types
func isTimestampCondition(op string) bool {
	return op == RULE_OP_WITHIN_SECONDS
}

// this Function creates a filter field name from an object type and field name
func filterFieldName(objType, fieldName string) string {
	return objType + "_" + fieldName
}

// Aurora

func auroraResToRowArray(data []map[string]interface{}, fields []string) [][]string {
	var rows [][]string
	for _, row := range data {
		var rowArray []string
		for _, field := range fields {
			value := row[field]
			switch v := value.(type) {
			case string:
				rowArray = append(rowArray, v)
			case time.Time:
				rowArray = append(rowArray, v.Format(TIMESTAMP_FORMAT))
			}
		}
		rows = append(rows, rowArray)
	}
	return rows
}

func (u LCSUtils) auroraToProfileObjects(rows []map[string]interface{}, mappings []ObjectMapping) []profilemodel.ProfileObject {
	u.Tx.Debug("[auroraToProfileObjects] creating object list form database result")
	objects := []profilemodel.ProfileObject{}
	objectUniqueKey := map[string]string{}
	for _, mapping := range mappings {
		objectUniqueKey[mapping.Name] = getIndexKey(mapping, INDEX_UNIQUE)
	}
	rows = filterDuplicatesKeepLatest(rows, objectUniqueKey)
	for _, row := range rows {
		objectType, ok := row[OBJECT_TYPE_NAME_FIELD].(string)
		if !ok {
			u.Tx.Warn("[auroraToProfileObjects] invalid object %+v", row)
			continue
		}
		objects = append(objects, u.auroraToProfileObject(objectType, row, objectUniqueKey[objectType]))
	}
	return objects
}

func (u LCSUtils) auroraToProfileObject(objectType string, record map[string]interface{}, objectIDKey string) profilemodel.ProfileObject {
	objectID, ok := record[objectIDKey].(string)
	if !ok {
		u.Tx.Warn("[auroraToProfileObject] invalid object %+v", record)
		return profilemodel.ProfileObject{}
	}
	//getting object count for pagination
	objCount, ok := record["total_records"].(int64)
	if !ok {
		u.Tx.Warn("[auroraToProfileObject] no object type count provided")
	}
	return profilemodel.ProfileObject{
		ID:                  objectID,
		Type:                objectType,
		AttributesInterface: record,
		Attributes:          core.MapStringInterfaceToMapStringString(record),
		TotalCount:          objCount,
	}
}

func (u LCSUtils) objectMapToProfileObject(objectType string, objMap map[string]string, objectIDKey string) profilemodel.ProfileObject {
	rec := map[string]interface{}{}
	for key, val := range objMap {
		rec[key] = val
	}
	return u.auroraToProfileObject(objectType, rec, objectIDKey)
}

// this function takes aurora search response for profile level data in Search index table
// and returns a profile list
func (u LCSUtils) auroraToProfiles(rows []map[string]interface{}) []profilemodel.Profile {
	u.Tx.Debug("[auroraToProfileObjects] creating profile list from database result")
	list := []profilemodel.Profile{}
	for _, row := range rows {
		profile, _, _ := u.auroraToProfile(row)
		list = append(list, profile)
	}
	return list
}

func (u LCSUtils) prettyPrint(res []map[string]interface{}) string {
	print := ""
	if len(res) == 0 {
		return "[]"
	}
	for _, row := range res {
		print += "\n-------------------------------------\n"
		for k, v := range row {
			if v != nil {
				print += fmt.Sprintf("%s: %v\n", k, v)
			}
		}
	}
	return print
}

func (u LCSUtils) getLaterDate(t1 time.Time, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}

func (u LCSUtils) auroraToProfile(row map[string]interface{}) (profilemodel.Profile, map[string]string, map[string]time.Time) {
	lastUpdatedObjectTypeMap := map[string]string{}
	lastUpdatedTsMap := map[string]time.Time{}
	connectID, ok := row["connect_id"].(string)
	if !ok {
		u.Tx.Warn("[auroraToProfile] invalid profile (no valid connect_id field)")
	}
	profile := profilemodel.Profile{
		ProfileId:  connectID,
		Attributes: make(map[string]string),
	}
	for field, value := range row {
		// Skip if field ends with _obj_type or _ts
		if strings.HasSuffix(field, "_obj_type") || strings.HasSuffix(field, "_ts") {
			continue
		}
		//	Special Handling for timestamp field
		if field == RESERVED_FIELD_TIMESTAMP {
			if profile.LastUpdated, ok = value.(time.Time); !ok {
				u.Tx.Warn("[auroraToProfile] invalid profile timestamp value: %v", value)
			}
			continue
		}
		//	Skip if field is not a string
		val, ok := value.(string)
		if !ok {
			continue
		}
		profile = SetProfileField(profile, PROFILE_OBJECT_TYPE_NAME+"."+field, val)
		lastUpdatedObjType, ok := row[field+"_obj_type"].(string)
		if !ok {
			u.Tx.Warn("[auroraToProfile] invalid profile last updated object Type value: %v", row[field+"_obj_type"])
		} else {
			profile = SetProfileField(profile, "_profile."+field+"_obj_type", lastUpdatedObjType)
			lastUpdatedObjectTypeMap[field] = lastUpdatedObjType
		}
		lastUpdatedTs, ok := row[field+"_ts"].(time.Time)
		if !ok {
			u.Tx.Warn("[auroraToProfile] invalid profile last updated timestamp value: %v", row[field+"_ts"])
		} else {
			lastUpdatedTsMap[field] = lastUpdatedTs
		}
	}
	return profile, lastUpdatedObjectTypeMap, lastUpdatedTsMap
}

func (u LCSUtils) MappingsToDbColumns(mappings []ObjectMapping) ([]string, map[string][]string) {
	objectMappings := make(map[string][]string)
	profileColumns := u.mappingsToProfileDbColumns(mappings)
	for _, mapping := range mappings {
		objectMappings[mapping.Name] = u.mappingToInteractionDbColumns(mapping)
	}
	return profileColumns, objectMappings
}

// this function returns the object and master table fields from object mapping
func (u LCSUtils) mappingsToProfileDbColumns(mappings []ObjectMapping) []string {
	var masterTableFields []string
	customAttr := map[string]string{}
	masterTableFields = append(masterTableFields, "connect_id")
	for _, reservedField := range RESERVED_FIELDS {
		if reservedField == "Attributes" {
			continue
		}
		//we skip profile_id since it does not exist in the profile search table (only in the master table )
		if reservedField == RESERVED_FIELD_PROFILE_ID {
			continue
		}
		if core.ContainsString([]string{RESERVED_FIELDS_ADDRESS, RESERVED_FIELDS_SHIPPING_ADDRESS, RESERVED_FIELDS_MAILING_ADDRESS, RESERVED_FIELDS_BILLING_ADDRESS}, reservedField) {
			for _, addrField := range ADDRESS_FIELDS {
				masterTableFields = append(masterTableFields, mappingToColumnNameProfile(reservedField+"."+addrField))
			}
		} else {
			masterTableFields = append(masterTableFields, mappingToColumnNameProfile(reservedField))
		}
	}
	for _, mapping := range mappings {
		for _, field := range mapping.Fields {
			if strings.HasPrefix(field.Target, PROFILE_OBJECT_TYPE_NAME+".Attributes.") && customAttr[field.Target] == "" {
				customAttr[field.Target] = field.Target
				masterTableFields = append(masterTableFields, mappingToColumnNameProfile(field.Target))
			}
		}
	}
	return masterTableFields
}

func customAttributeFieldsNameFromMapping(objMapping ObjectMapping) []string {
	var customAttrFields []string
	customAttr := map[string]string{}
	for _, field := range objMapping.Fields {
		if strings.HasPrefix(field.Target, "_profile.Attributes.") && customAttr[field.Target] == "" {
			customAttr[field.Target] = field.Target
			customAttrFields = append(customAttrFields, mappingToColumnNameProfile(field.Target))
		}
	}
	return customAttrFields
}

func (u LCSUtils) mappingToInteractionDbColumns(mapping ObjectMapping) []string {

	var objectFields []string
	for _, field := range mapping.Fields {
		objectFields = append(objectFields, mappingToColumnName(field.Source))
	}
	return objectFields
}

func mappingToColumnName(m string) string {
	cl := strings.ReplaceAll(m, "_source.", "")
	return cl
}

func mappingToColumnNameProfile(m string) string {
	cl := strings.ReplaceAll(m, PROFILE_OBJECT_TYPE_NAME+".", "")
	return cl
}

func mappingsToEntityResolution(m []string) []string {
	clList := []string{}
	for _, field := range m {
		cl := strings.ReplaceAll(field, ".", "")
		clList = append(clList, cl)
	}
	return clList
}

// returns the name of teh DB field corresponding to the profile INDEX
func getIndexKey(mappings ObjectMapping, indexType string) string {
	for _, field := range mappings.Fields {
		if core.ContainsString(field.Indexes, indexType) {
			return mappingToColumnName(field.Source)
		}
	}
	return ""
}

// Gets the specified mapping from a list of ObjectMappings
func getObjectMapping(mappings []ObjectMapping, objectType string) (ObjectMapping, error) {
	for _, mapping := range mappings {
		if mapping.Name == objectType {
			return mapping, nil
		}
	}
	return ObjectMapping{}, fmt.Errorf("mapping not found for object type: %v", objectType)
}

// templates
func parseFieldNamesFromTemplate(str string) []string {
	res := []string{}
	//parse group from regexp
	re := regexp.MustCompile(`\{\{\.([A-Za-z0-9]+)\}\}`)
	// Match the regular expression against a string
	matches := re.FindAllStringSubmatch(str, -1)
	log.Printf("%+v", matches)
	// Print the group results
	for _, match := range matches {
		res = append(res, match[1])
	}

	return res
}

// Validate the string matches template requirements: alphanumeric value with leading period and enclosed by double curly braces
// e.g. {{.firstName}}
func isTemplate(str string) bool {
	re := regexp.MustCompile(`\{\{\.([A-Za-z0-9]+)\}\}`)
	return re.MatchString(str)
}

func objectMapToProfileMap(obj map[string]string, mappings ObjectMapping) (map[string]string, error) {
	profileMap := make(map[string]string)
	for key := range obj {
		for _, field := range mappings.Fields {
			fieldName := mappingToColumnName(field.Source)
			if key == fieldName {
				if strings.Contains(field.Target, PROFILE_OBJECT_TYPE_NAME+".") {
					profileFieldName := mappingToColumnNameProfile(field.Target)
					profileMap[profileFieldName] = obj[key]
				}
			}
		}
	}

	return profileMap, nil
}

func objectMapToProfile(obj map[string]string, mappings ObjectMapping) (profile profilemodel.Profile) {
	for key := range obj {
		for _, field := range mappings.Fields {
			fieldName := mappingToColumnName(field.Source)
			if key == fieldName {
				if strings.Contains(field.Target, PROFILE_OBJECT_TYPE_NAME+".") {
					profileFieldName := mappingToColumnNameProfile(field.Target)
					profile.Set(profileFieldName, obj[key])
				}
				continue
			}
		}
	}

	return profile
}

// Format query string to make it easier to read. Since we frequently use string literals
// to build queries, we want to remove newlines, tabs, etc.
//
// There should never be a functional change to the query after formatting.
func formatQuery(query string) string {
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", "")
	query = strings.TrimSpace(query)
	return query
}

// Build a where clause programatically based on the parameterized search criteria
func buildWhereClause(searchCriteria []BaseCriteria) (whereClause string, err error) {
	for _, criteria := range searchCriteria {
		if c, ok := criteria.(SearchCriterion); ok {
			whereClause += fmt.Sprintf(`"%s" %v`, c.Column, c.Operator)
			if c.Operator == IN || c.Operator == NOTIN {
				if reflect.ValueOf(c.Value).Kind() != reflect.Slice {
					return "", fmt.Errorf("value should be a slice when Operator is IN or NOTIN")
				}
				whereClause += " ("
				for _, val := range c.Value.([]any) {
					if v, ok := val.(string); ok {
						whereClause += fmt.Sprintf("'%s',", v)
					} else if v, ok := val.(int); ok {
						whereClause += fmt.Sprintf("%d", v)
					}
				}
				whereClause = whereClause[:len(whereClause)-1] + ")"
			} else {
				if val, ok := c.Value.(string); ok {
					whereClause += fmt.Sprintf(" '%s'", val)
				} else if val, ok := c.Value.(int); ok {
					whereClause += fmt.Sprintf(" %d", val)
				}
			}
		} else if c, ok := criteria.(SearchOperator); ok {
			whereClause += " " + string(c.Operator) + " "
		} else if c, ok := criteria.(SearchGroup); ok {
			clause, err := buildWhereClause(c.Criteria)
			if err != nil {
				return "", err
			}
			whereClause += fmt.Sprintf("(%s)", clause)
		}
	}

	return whereClause, err
}

func buildParameterizedWhereClause(searchCriteria []BaseCriteria) (whereClause string, parameters []interface{}, err error) {
	for _, criteria := range searchCriteria {
		if c, ok := criteria.(SearchCriterion); ok {
			whereClause += fmt.Sprintf(`"%s" %v`, c.Column, c.Operator)
			if c.Operator == IN || c.Operator == NOTIN {
				if reflect.ValueOf(c.Value).Kind() != reflect.Slice {
					return "", nil, fmt.Errorf("value should be a slice when Operator is IN or NOTIN")
				}
				whereClause += " ("
				for _, val := range c.Value.([]any) {
					whereClause += "?,"
					parameters = append(parameters, val)
				}
				whereClause = whereClause[:len(whereClause)-1] + ")"
			} else {
				whereClause += " ?"
				parameters = append(parameters, c.Value)
			}
		} else if c, ok := criteria.(SearchOperator); ok {
			whereClause += " " + string(c.Operator) + " "
		} else if c, ok := criteria.(SearchGroup); ok {
			clause, params, err := buildParameterizedWhereClause(c.Criteria)
			if err != nil {
				return "", nil, err
			}
			whereClause += fmt.Sprintf("(%s)", clause)
			parameters = append(parameters, params...)
		}
	}

	return whereClause, parameters, err
}

func filterDuplicatesKeepLatest(data []map[string]interface{}, typeKeyMap map[string]string) []map[string]interface{} {
	type itemKey struct {
		objectType string
		keyValue   string
	}

	keyMap := make(map[itemKey]int)
	result := make([]map[string]interface{}, 0, len(data))

	for _, item := range data {
		objectType, ok := item[OBJECT_TYPE_NAME_FIELD].(string)
		if !ok {
			continue
		}
		keyField, exists := typeKeyMap[objectType]
		if !exists {
			continue
		}
		keyValue, ok := item[keyField].(string)
		if !ok {
			continue
		}
		timestamp, ok := item["timestamp"].(time.Time)
		if !ok {
			continue
		}
		key := itemKey{objectType: objectType, keyValue: keyValue}

		if existingIndex, exists := keyMap[key]; exists {
			existingTimestamp, _ := result[existingIndex]["timestamp"].(time.Time)
			if timestamp.After(existingTimestamp) {
				result[existingIndex] = item
			}
		} else {
			keyMap[key] = len(result)
			result = append(result, item)
		}
	}

	return result
}
