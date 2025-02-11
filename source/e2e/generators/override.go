// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package generators

import (
	"math/rand"
	"reflect"
	"slices"
)

type FieldOverrides struct {
	Weight int
	Fields []FieldOverride
}

type FieldOverride struct {
	Name  string
	Value interface{}
}

// Apply field override based on weight
// e.g. if a weight is 1, 1 out of 100000 times the generated object will be overridden with the provided values
func applyFieldOverrides[T any](v T, fieldOverrides []FieldOverrides) T {
	if len(fieldOverrides) == 0 {
		return v
	}

	var cumulative int
	randNum := rand.Intn(100000)
	for _, override := range fieldOverrides {
		cumulative += override.Weight
		if cumulative > randNum {
			replaceFields(reflect.ValueOf(v), override.Fields)
			break
		}
	}

	return v
}

func replaceFields(v reflect.Value, fieldOverrides []FieldOverride) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	//	Flatten `fieldOverrides` Names to its own slice
	overrideNames := []string{}
	for _, override := range fieldOverrides {
		overrideNames = append(overrideNames, override.Name)
	}

	for i := 0; i < v.NumField(); i++ {
		//	if field on v matches an override, replace it
		fieldName := v.Type().Field(i).Name
		if slices.Contains(overrideNames, fieldName) {
			for _, override := range fieldOverrides {
				if override.Name == fieldName {
					v.Field(i).Set(reflect.ValueOf(override.Value))
				}
			}
		} else if v.Field(i).Kind() == reflect.Struct {
			// recursively replace fields for nested structs
			replaceFields(v.Field(i), fieldOverrides)
		} else if v.Field(i).Kind() == reflect.Slice {
			// iterate over reach object in slices
			for j := 0; j < v.Field(i).Len(); j++ {
				replaceFields(v.Field(i).Index(j), fieldOverrides)
			}
		}
	}
}
