// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"tah/upt/source/ucp-common/src/model/traveler"
	"testing"
	"time"
)

func TestUtils(t *testing.T) {
	ParseFloat64("10.53")
	ParseInt64("19")
	ParseInt("21")
	ParseTime("2019-01-01 00:00:00", "2006-01-02 15:04:05")
	TryParseTime("2019-01-01 00:00:00", "2006-01-02 15:04:05")
	TryParseFloat("12.1")
	TryParseInt("12")
	ParseIntArray([]string{"12", "13", "14"})
	boole := ContainsString([]string{"a", "b"}, "a")
	if !boole {
		t.Errorf("Should be true")
	}
	boole = ContainsString([]string{"a", "b"}, "c")
	if boole {
		t.Errorf("Should be false")
	}
}

func TestSnakeCase(t *testing.T) {
	test := "already_in_snake_case"
	res := ToSnakeCase(test)
	if res != test {
		t.Errorf("Should be %v", test)
	}

	test = "NotInSnakeCase"
	if IsSnakeCaseLower(test) {
		t.Errorf("IsSnakeCaseLower should be false")
	}
	res = ToSnakeCase(test)
	if res != "not_in_snake_case" {
		t.Errorf("Should be %v", "not_in_snake_case")
	}
	if !IsSnakeCaseLower(res) {
		t.Errorf("IsSnakeCaseLower should be true")
	}

	test = "test-kebab-case"
	res = ToSnakeCase(test)
	if res != "test_kebab_case" {
		t.Errorf("Should be %v", "not_in_snake_case")
	}
}

func TestJsonTags(t *testing.T) {
	type User struct {
		Name     string `json:"name"`
		Age      int    `json:"age,omitempty"`
		Password string `json:"-"`
	}
	tags := GetJSONTags(User{})
	expected := []string{"name", "age"}
	if strings.Join(tags, ",") != strings.Join(expected, ",") {
		t.Errorf("Should be %v,  not %v", expected, tags)
	}
}

func TestGetJsonTag(t *testing.T) {
	type User struct {
		Name     string `json:"name"`
		Age      int    `json:"age,omitempty"`
		Password string `json:"-"`
	}
	tag := GetJSONTag(User{}, "Name")
	if tag != "name" {
		t.Errorf("Should be %v, not %v", "name", tag)
	}
	tag = GetJSONTag(User{}, "Age")
	if tag != "age" {
		t.Errorf("Should be %v, not %v", "age", tag)
	}
	tag = GetJSONTag(User{}, "Password")
	if tag != "" {
		t.Errorf("Should be %v, not %v", "", tag)
	}
}

func TestDiffAndCombineAccpMaps(t *testing.T) {
	newMap := map[string]interface{}{
		"loyaltyId":    "123",
		"profileId":    89,
		"status":       "gold",
		"point_type":   "debit",
		"point_number": 10,
		"booli":        true,
	}

	oriMap := map[string]interface{}{
		"loyaltyId":    "890",
		"profileId":    100,
		"point_type":   "credit",
		"status":       "silver",
		"days":         "4",
		"point_number": "",
	}
	objectTypeName := "testObj"
	fields := []string{"testObj.loyaltyId", "testObj.profileId", "testObj.booli", "status", "foo.status", "foo.*"}

	combinedMap := DiffAndCombineAccpMaps(newMap, oriMap, objectTypeName, fields)

	if loyaltyId, ok := combinedMap["loyaltyId"]; !ok || loyaltyId != "123" {
		t.Errorf("Expected loyaltyId 123, but got %v", loyaltyId)
	}
	if profileId, ok := combinedMap["profileId"]; !ok || profileId != 89 {
		t.Errorf("Expected profileId 89, but got %v", profileId)
	}
	if status, ok := combinedMap["status"]; !ok || status != "silver" {
		t.Errorf("Expected status silver, but got %v", status)
	}
	if pointType, ok := combinedMap["point_type"]; !ok || pointType != "credit" {
		t.Errorf("Expected point type credit, but got %v", pointType)
	}
	if days, ok := combinedMap["days"]; !ok || days != "4" {
		t.Errorf("Expected days 4, but got %v. Field should remain even if not in partial record", days)
	}
	if pointNumber, ok := combinedMap["point_number"]; !ok || pointNumber != "" {
		t.Errorf("Expected point number empty, but got %v", pointNumber)
	}
	if booli, ok := combinedMap["booli"]; !ok || booli != true {
		t.Errorf("Expected booli true, but got %v", booli)
	}

	wildFields := []string{"testObj.*", "foo.status"}
	newCombinedMap := DiffAndCombineAccpMaps(newMap, oriMap, objectTypeName, wildFields)
	if pointNumber, ok := newCombinedMap["point_number"]; !ok || pointNumber != 10 {
		t.Errorf("Expected point number 10, but got %v", pointNumber)
	}
	if status, ok := newCombinedMap["status"]; !ok || status != "gold" {
		t.Errorf("Expected status gold, but got %v", status)
	}
}

func TestUtilsErrorCoverage(t *testing.T) {
	ParseFloat64("invalid")
	ParseInt64("invalid")
	ParseInt("invalid")
	ParseTime("invalid", "2006-01-02 15:04:05")
	TryParseTime("invalid", "2006-01-02 15:04:05")
	TryParseFloat("invalid")
	TryParseInt("invalid")
	ParseIntArray([]string{"invalid"})
}

func TestAccpObjectToFields(t *testing.T) {
	testAccpObject := traveler.EmailHistory{}
	fieldNames := AccpObjectToFields(testAccpObject)
	if len(fieldNames) != 8 {
		t.Errorf("Incorrect number of fields. Expected 8, got %v", len(fieldNames))
	}
	if !slices.Contains(fieldNames, "accpObjectID") {
		t.Error("Expected fieldName accpObjectID")
	}
	if !slices.Contains(fieldNames, "travellerId") {
		t.Error("Expected fieldName travellerId")
	}
	if !slices.Contains(fieldNames, "address") {
		t.Error("Expected fieldName address")
	}
	if !slices.Contains(fieldNames, "type") {
		t.Error("Expected fieldName type")
	}
	if !slices.Contains(fieldNames, "lastUpdated") {
		t.Error("Expected fieldName lastUpdated")
	}
	if !slices.Contains(fieldNames, "lastUpdatedBy") {
		t.Error("Expected fieldName lastUpdatedBy")
	}
	if !slices.Contains(fieldNames, "overallConfidenceScore") {
		t.Error("Expected fieldName overallConfidenceScore")
	}
	if !slices.Contains(fieldNames, "isVerified") {
		t.Error("Expected fieldName isVerified")
	}
}

func TestRetryUtil(t *testing.T) {
	shouldFail := func() (interface{}, error) {
		if true {
			return nil, errors.New("Error")
		}
		return nil, nil
	}
	maxRetries := 3
	interval := 2
	maxInterval := 5
	startTime := time.Now()
	res, err := Retry(shouldFail, maxRetries, interval, maxInterval)
	duration := time.Since(startTime)
	if duration < 3*time.Second {
		t.Fatalf("retries should have taken longer than 3 seconds")
	}
	if err == nil {
		t.Fatalf("retry should have failed")
	}
	if res != nil {
		t.Fatalf("result should be nil")
	}

	type RandomStruct struct {
		Name string
	}
	testStruct := RandomStruct{Name: "test-name"}
	shouldSucceed := func() (interface{}, error) {
		return testStruct, nil
	}
	res, err = Retry(shouldSucceed, maxRetries, interval, maxInterval)
	if err != nil {
		t.Fatalf("retry should not have failed")
	}
	t.Logf("result: %v", res.(RandomStruct))
	t.Logf("result: %v", testStruct)
	if res.(RandomStruct).Name != testStruct.Name {
		t.Fatalf("successful retry should return test struct")
	}
}

func TestSafeDereference(t *testing.T) {
	// Nil Case
	var nilVal *[]string
	emptySlice := SafeDereference(nilVal)
	if emptySlice == nil && len(emptySlice) != 0 {
		t.Fatalf("dereferencing nil slice should return empty slice")
	}

	// Pointer case
	val := []string{"a", "b", "c"}
	res := SafeDereference(&val)
	if res == nil && reflect.DeepEqual(res, val) {
		t.Fatalf("dereferencing pointer should return value")
	}
}
