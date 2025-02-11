// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	core "tah/upt/source/tah-core/core"
)

func validateFieldMappings(fieldMappings []FieldMapping) error {
	nProfileIndexes := 0
	nUniqueIndexes := 0
	for _, field := range fieldMappings {
		if core.ContainsString(field.Indexes, INDEX_PROFILE) {
			nProfileIndexes++
		}
		if core.ContainsString(field.Indexes, INDEX_UNIQUE) {
			nUniqueIndexes++
		}
		if strings.HasPrefix(field.Source, "_source._") {
			return fmt.Errorf("source field name cannot start with '_' after prefix '_source.'")
		}
	}
	if nProfileIndexes != 1 {
		return fmt.Errorf("mapping must have exactly 1 PROFILE index")
	}
	if nUniqueIndexes != 1 {
		return fmt.Errorf("mapping must have exactly 1 UNIQUE index")
	}
	return nil
}

func validateMappingName(name string) error {
	if name == "" {
		return fmt.Errorf("mapping name should not be empty")
	}
	if len(name) > MAX_MAPPING_NAME_LENGTH {
		return fmt.Errorf("mapping name %v exceeds limit of %v characters", name, MAX_MAPPING_NAME_LENGTH)
	}
	if !core.IsSnakeCaseLower(name) {
		return errors.New("mapping name must be in snake case")
	}
	if core.ContainsString(RESERVED_FIELDS, name) {
		return fmt.Errorf(
			"%v name is a reserved field and cannot be used for object type name (along with %v)",
			name,
			strings.Join(RESERVED_FIELDS, ","),
		)
	}

	return nil
}

func ValidateDomainName(name string) error {
	if !core.IsSnakeCaseLower(name) {
		return fmt.Errorf("Domain name %v invalid, must be in snake case", name)
	}
	if len(name) > MAX_DOMAIN_NAME_LENGTH || len(name) < MIN_DOMAIN_NAME_LENGTH {
		return fmt.Errorf("Domain name length should be between %v and %v chars", MIN_DOMAIN_NAME_LENGTH, MAX_DOMAIN_NAME_LENGTH)
	}
	return nil
}

func validateDomainTags(tags map[string]string) error {
	// Ensure domain is tagged with correct environment.
	// This is required when selecting domains to display in UPT.
	if tags[DOMAIN_CONFIG_ENV_KEY] == "" {
		return errors.New("env name is required in the tags")
	}
	return nil
}

func uniqueElements[T comparable](slice []T) []T {
	uniqueMap := make(map[T]bool) // A map to keep track of unique elements.
	var result []T                // A slice to store the unique elements.

	for _, element := range slice {
		if _, exists := uniqueMap[element]; !exists {
			uniqueMap[element] = true        // Mark the element as seen.
			result = append(result, element) // Add to result slice.
		}
	}

	return result
}

// rule set validation
func validateConditions(conditions []Condition, mappings map[string][]string) error {
	var err error
	var errorMsgs error
	matchConditionFound := false
	matchConditionTypes := []string{}
	filterConditionTypes := []string{}
	for _, condition := range conditions {
		incomingObjectType := condition.IncomingObjectType
		incomingObjectField := condition.IncomingObjectField

		if incomingObjectType == "" || incomingObjectField == "" {
			errorMsgs = errors.Join(
				errorMsgs,
				errors.New(getErrorMessagePrefix(condition.Index)+"incoming object type or field cannot be empty"),
			)
			continue
		}

		expectedFields, ok := mappings[incomingObjectType]
		if !ok {
			errorMsgs = errors.Join(errorMsgs, fmt.Errorf("%s no mapping not found for object type %s", getErrorMessagePrefix(condition.Index), incomingObjectType))
			continue
		}

		isMapping := slices.Contains(expectedFields, incomingObjectField)
		if condition.ConditionType == CONDITION_TYPE_SKIP {
			errorMsgs = errors.Join(errorMsgs, validateSkipCondition(condition, mappings, isMapping))
		} else if condition.ConditionType == CONDITION_TYPE_MATCH {
			matchConditionFound = true
			matchConditionTypes, err = validateMatchCondition(condition, mappings, isMapping, expectedFields, matchConditionTypes)
			errorMsgs = errors.Join(errorMsgs, err)
		} else if condition.ConditionType == CONDITION_TYPE_FILTER {
			filterConditionTypes, err = validateFilterCondition(condition, mappings, filterConditionTypes)
			errorMsgs = errors.Join(errorMsgs, err)
		} else {
			errorMsgs = errors.Join(errorMsgs, errors.New(getErrorMessagePrefix(condition.Index)+"invalid condition type found"))
		}
	}
	if !matchConditionFound {
		errorMsgs = errors.Join(errorMsgs, errors.New("requires at least one match condition"))

	}
	errorMsgs = errors.Join(errorMsgs, validateMatchFilter(matchConditionTypes, filterConditionTypes))
	return errorMsgs
}

func validateCacheConditions(conditions []Condition, mappings map[string][]string) error {
	var errorMsgs error
	for _, condition := range conditions {
		incomingObjectType := condition.IncomingObjectType
		incomingObjectField := condition.IncomingObjectField

		if incomingObjectType == "" || incomingObjectField == "" {
			errorMsgs = errors.Join(errors.New(getErrorMessagePrefix(condition.Index) + "incoming object type or field cannot be empty"))
			continue
		}

		expectedFields, ok := mappings[incomingObjectType]
		if !ok {
			errorMsgs = errors.Join(errorMsgs, fmt.Errorf("%s no mapping not found for object type %s", getErrorMessagePrefix(condition.Index), incomingObjectType))
			continue
		}

		if condition.ConditionType != CONDITION_TYPE_SKIP {
			errorMsgs = errors.Join(errors.New(getErrorMessagePrefix(condition.Index) + "not skip condition"))
			continue
		}

		isMapping := slices.Contains(expectedFields, incomingObjectField)
		errorMsgs = errors.Join(errorMsgs, validateSkipCondition(condition, mappings, isMapping))
	}
	return errorMsgs
}

func validateSkipCondition(condition Condition, mappings map[string][]string, isIncomingMapping bool) error {
	if !isIncomingMapping {
		return errors.New(getErrorMessagePrefix(condition.Index) + "invalid incoming field")
	}
	if condition.Op == "" {
		return errors.New(getErrorMessagePrefix(condition.Index) + "operation field cannot be empty")
	}
	if condition.Op == RULE_OP_EQUALS_VALUE ||
		condition.Op == RULE_OP_NOT_EQUALS_VALUE ||
		condition.Op == RULE_OP_MATCHES_REGEXP {
		skipValue := condition.SkipConditionValue
		if skipValue.ValueType != CONDITION_VALUE_TYPE_STRING {
			return errors.New(getErrorMessagePrefix(condition.Index) + "skip condition value should not be empty")
		}
	} else if condition.Op == RULE_OP_EQUALS ||
		condition.Op == RULE_OP_NOT_EQUALS {
		incomingObjectType2 := condition.IncomingObjectType2
		incomingObjectField2 := condition.IncomingObjectField2

		err := validateTypeFieldMapping(incomingObjectType2, incomingObjectField2, mappings)
		if err != nil {
			return errors.New(getErrorMessagePrefix(condition.Index) + err.Error())
		}
	} else {
		return errors.New(getErrorMessagePrefix(condition.Index) + "invalid operation")
	}
	return nil
}

func validateTypeFieldMapping(objectType string, objectField string, mappings map[string][]string) error {
	if objectType == "" || objectField == "" {
		return errors.New("object type or field cannot be empty")
	}
	expectedFields, ok := mappings[objectType]
	if !ok {
		return fmt.Errorf("no mapping not found for object type %s", objectType)
	}
	if !slices.Contains(expectedFields, objectField) {
		return errors.New("invalid incoming field")
	}
	return nil
}

func validateMatchCondition(
	condition Condition,
	mappings map[string][]string,
	isIncomingMapping bool,
	expectedIncomingFields []string,
	matchConditionTypes []string,
) ([]string, error) {
	existingObjectType := condition.ExistingObjectType
	existingObjectField := condition.ExistingObjectField
	if !isIncomingMapping {
		err := validateTemplating(condition.IncomingObjectField, expectedIncomingFields)
		if err != nil {
			return []string{}, errors.New(getErrorMessagePrefix(condition.Index) + err.Error())
		}
	}
	if existingObjectType == "" && existingObjectField == "" {
		return []string{}, nil
	}
	if existingObjectType == "" {
		return []string{}, errors.New(
			getErrorMessagePrefix(condition.Index) + "existing object type cannot be empty if existing object field is present",
		)
	}
	if existingObjectField == "" {
		return []string{}, errors.New(
			getErrorMessagePrefix(condition.Index) + "existing object field cannot be empty if existing object type is present",
		)
	}

	expectedFields, ok := mappings[existingObjectType]
	if !ok {
		return []string{}, errors.New(getErrorMessagePrefix(condition.Index) + "existing object type not found")
	}

	isMapping := slices.Contains(expectedFields, existingObjectField)
	if !isMapping {
		err := validateTemplating(existingObjectField, expectedFields)
		if err != nil {
			return []string{}, errors.New(getErrorMessagePrefix(condition.Index) + err.Error())
		}
	}
	matchConditionTypes = append(matchConditionTypes, condition.IncomingObjectType+existingObjectType)
	return matchConditionTypes, nil
}

func validateFilterCondition(condition Condition, mappings map[string][]string, filterConditionTypes []string) ([]string, error) {
	incomingObjectField := condition.IncomingObjectField
	existingObjectType := condition.ExistingObjectType
	existingObjectField := condition.ExistingObjectField
	filterConditionValue := condition.FilterConditionValue

	if condition.Op != RULE_OP_WITHIN_SECONDS {
		return []string{}, errors.New(
			getErrorMessagePrefix(condition.Index) + "filter condition operation should be of type within seconds",
		)
	}
	if incomingObjectField != "timestamp" {
		return []string{}, errors.New(getErrorMessagePrefix(condition.Index) + "invalid incoming object field")
	}
	if existingObjectField != "timestamp" {
		return []string{}, errors.New(getErrorMessagePrefix(condition.Index) + "invalid existing object field")
	}
	_, ok := mappings[existingObjectType]
	if !ok {
		return []string{}, errors.New(getErrorMessagePrefix(condition.Index) + "existing object type not found")
	}
	if filterConditionValue.ValueType != CONDITION_VALUE_TYPE_INT || filterConditionValue.IntValue < 0 {
		return []string{}, errors.New(
			getErrorMessagePrefix(condition.Index) + "filter condition value must be of type int and greater than or equal to 0",
		)
	}
	filterConditionTypes = append(filterConditionTypes, condition.IncomingObjectType+existingObjectType)
	return filterConditionTypes, nil
}

func validateTemplating(templateValue string, objectFieldNames []string) error {
	// Validate bracket syntax
	brackets := 0
	for _, c := range templateValue {
		if c == '{' {
			brackets++
		}
		if c == '}' {
			brackets--
		}
	}
	if brackets > 0 {
		return errors.New("invalid template syntax. validate that you have the right number of brackets. Ex: {{fieldName}}")
	}

	// Validate if template holds mapped fields
	// Create a regular expression to match placeholders in the template
	re := regexp.MustCompile(`{{\w+}}`)

	// Find all matches in the template string
	matches := re.FindAllString(templateValue, -1)
	if len(matches) == 0 {
		return errors.New("invalid template formatting")
	}
	for _, match := range matches {
		match = strings.Replace(match, "{{", "", -1)
		match = strings.Replace(match, "}}", "", -1)
		if !slices.Contains(objectFieldNames, match) {
			return fmt.Errorf("invalid placeholder, should be within %v", objectFieldNames)
		}
	}
	return nil
}

func validateMatchFilter(matchConditionTypes []string, filterConditionTypes []string) error {
	for _, filterCondition := range filterConditionTypes {
		if !slices.Contains(matchConditionTypes, filterCondition) {
			return errors.New("one or more filter conditions do not match up with a match condition")
		}
	}
	return nil
}

func getErrorMessagePrefix(conditionIndex int) string {
	return fmt.Sprintf("condition %d | ", conditionIndex)
}

func validatePagination(pagination []PaginationOptions) error {
	for _, option := range pagination {
		if option.PageSize*option.Page > MAX_INTERACTIONS_PER_PROFILE {
			return fmt.Errorf("invalid Pagination settings: PageSize*Page should be lower than MAX_INTERACTIONS_PER_PROFILE=%d", MAX_INTERACTIONS_PER_PROFILE)
		}
		if option.PageSize > MAX_INTERACTIONS_PER_PROFILE {
			return fmt.Errorf("invalid Pagination settings: PageSize should be lower than MAX_INTERACTIONS_PER_PROFILE=%d", MAX_INTERACTIONS_PER_PROFILE)
		}
		if option.Page < 0 {
			return fmt.Errorf("invalid Pagination settings: Page must not be negative")
		}
	}
	return nil
}
