// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"tah/upt/source/ucp-common/src/model/traveler"
	"time"

	"github.com/google/uuid"
)

func ParseFloat64(val string) float64 {
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("[WARNING] error while parsing float: %s", err)
		return 0
	}
	return parsed
}

func ParseInt64(val string) int64 {
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Printf("[WARNING] error while parsing int: %s", err)
		return 0
	}
	return parsed
}

func ParseInt(val string) int {
	parsed, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("[WARNING] error while parsing int: %s", err)
		return 0
	}
	return parsed
}

func ParseTime(t string, layout string) time.Time {
	parsed, err := time.Parse(layout, t)
	if err != nil {
		log.Printf("[WARNING] error while parsing date: %s", err)
		return time.Time{}
	}
	return parsed
}

// Convert string to time.
//
// Returns type's default value if the value is empty
// or there is an error while parsing.
func TryParseTime(t, layout string) (time.Time, error) {
	if t == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(layout, t)
	if err != nil {
		return time.Time{}, errors.New("error parsing time: \"%s\"")
	}
	return parsed, nil
}

// Convert string to int.
//
// Returns type's default value if the value is empty
// or there is an error while parsing.
func TryParseInt(val string) (int, error) {
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, errors.New("error parsing value \"%s\" to int")
	}
	return parsed, nil
}

func TryParseFloat(val string) (float64, error) {
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, errors.New("error parsing value \"%s\" to int")
	}
	return parsed, nil
}

func ParseIntArray(array []string) []int {
	res := []int{}
	for _, val := range array {
		res = append(res, ParseInt(val))
	}
	return res
}

// this method convert a string to snake case
func ToSnakeCase(input string) string {
	var result strings.Builder
	//we keep track of the upper case char to support uppercase sequences like travelerID that should be come traveller_id and not traveller_i_d
	isPreviousUpper := false
	for i, char := range input {
		if char >= 'A' && char <= 'Z' {
			if !isPreviousUpper && i > 0 {
				result.WriteByte('_')
			}
			isPreviousUpper = true
		} else {
			isPreviousUpper = false
		}
		result.WriteByte(byte(char))
	}
	str := result.String()
	str = strings.Replace(str, "-", "_", -1)
	return strings.ToLower(str)
}

func IsSnakeCaseLower(input string) bool {
	for _, char := range input {
		if !((char >= 'a' && char <= 'z') || char == '_' || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}

func ContainsString(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}

func DiffAndCombineAccpMaps(newMap, oriMap map[string]interface{}, objectTypeName string, fieldNames []string) map[string]interface{} {
	combinedMap := make(map[string]interface{})
	overrideFields := GetOverrideFields(objectTypeName, fieldNames)

	for key, val := range oriMap {
		combinedMap[key] = val
	}

	for _, fieldName := range overrideFields {
		val, ok := newMap[fieldName]
		if ok {
			combinedMap[fieldName] = val
		}
	}

	containsWild := containsElement(overrideFields, "*")
	if containsWild {
		for key, val := range newMap {
			combinedMap[key] = val
		}
	}

	return combinedMap
}

func GetOverrideFields(objectTypeName string, fieldNames []string) []string {
	overrideFields := []string{}
	for _, fieldName := range fieldNames {
		overrideArray := strings.Split(fieldName, ".")
		if len(overrideArray) == 2 {
			if overrideArray[0] == objectTypeName {
				overrideFields = append(overrideFields, overrideArray[1])
			}
		}
	}
	return overrideFields
}

func containsElement(arr []string, element string) bool {
	for _, item := range arr {
		if item == element {
			return true
		}
	}
	return false
}

// Reflection Utils
func GetJSONTags(myStruct interface{}) []string {
	structType := reflect.TypeOf(myStruct)
	jsonTags := []string{}
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get("json")
		if tag != "" && tag != "-" {
			jsonTags = append(jsonTags, strings.Split(tag, ",")[0])
		}
	}
	return jsonTags
}

func GetJSONTag(someStruct interface{}, fieldName string) string {
	structType := reflect.TypeOf(someStruct)
	if structField, ok := structType.FieldByName(fieldName); ok {
		tag := structField.Tag.Get("json")
		if tag != "" && tag != "-" {
			return strings.Split(tag, ",")[0]
		}
	}
	return ""
}

func AccpObjectToFields(data traveler.AccpObject) []string {
	if data == nil {
		return []string{}
	}
	reflectedValue := reflect.ValueOf(data)
	reflectedType := reflectedValue.Type()

	var fieldNames []string

	for i := 0; i < reflectedValue.NumField(); i++ {
		field := reflectedType.Field(i)
		fieldNames = append(fieldNames, field.Tag.Get("json"))
	}
	return fieldNames
}

// Retries {action} function on fail {maxRetries} times until {maxRetries} attempts are reached or {action} returns successfully
// Utilizes Jitter and Backoff in calculating delay, to increase retry efficiency
func Retry(action func() (interface{}, error), maxRetries, baseInterval, maxInterval int) (res interface{}, err error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		log.Printf("Running attempt %v", attempt+1)
		res, err = action()
		// Return if try succeeds
		if err == nil {
			break
		}
		// Don't sleep after last attempt
		if attempt == maxRetries-1 {
			continue
		}
		// Sleep with backoff and jitter: https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
		temp := math.Min(float64(maxInterval), float64(baseInterval)*math.Pow(2, float64(attempt)))
		sleepTime := temp/2 + rand.Float64()*(temp/2.0)
		log.Printf("Attempt %v failed, sleeping %.3v seconds and retrying", attempt+1, sleepTime)
		time.Sleep(time.Duration(sleepTime*1000) * time.Millisecond)
	}
	if err != nil {
		log.Printf("Last attempt failed")
	}
	return res, err
}

func IsUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}

// validate string as DB field
func ValidateString(str string, INVALID_CHARS []string) error {
	if len(str) > 255 {
		return errors.New("string length exceeds 255 characters")
	}
	for _, char := range INVALID_CHARS {
		if strings.Contains(str, string(char)) {
			return fmt.Errorf("string contains invalid characters %s", char)
		}
	}
	return nil
}

// Dereferences a pointer and defaults to zero value if pointer is nil
func SafeDereference[T interface{}](ptr *T) T {
	if ptr == nil {
		var t T
		return t
	}
	return *ptr
}
