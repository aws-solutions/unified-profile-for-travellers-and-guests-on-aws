// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"log"
	model "tah/upt/source/ucp-common/src/model/traveler"
	"testing"
	"time"
)

func TestParsePermission(t *testing.T) {
	// Valid admin user
	admin := "*/*"
	isAdmin, filters := parsePermission(admin)
	if !isAdmin || len(filters) != 0 {
		t.Errorf("[TestParsePermission] String %v should return a valid admin user", admin)
	}

	// Invalid admin user
	invalidAdmin := "*/*\nclickstream/*?from=RDU"
	isAdmin, filters = parsePermission(invalidAdmin)
	if isAdmin || len(filters) != 0 {
		t.Errorf("[TestParsePermission] String %v should return an invalid admin user", invalidAdmin)
	}

	// Clickstream all data
	clickstreamAll := "clickstream/*"
	expected := model.Filter{
		ObjectName:    "clickstream",
		Fields:        []string{"*"},
		ExcludeFields: true,
		Condition:     "",
	}
	isAdmin, filters = parsePermission(clickstreamAll)
	if isAdmin || len(filters) != 1 || !areFiltersEqual(filters["clickstream"], expected) {
		t.Errorf("[TestParsePermission] Permission was not parsed as expected: %v", clickstreamAll)
	}

	// Clickstream include some fields with filter
	clickstreamSomeFieldsFromRDU := "clickstream/to,from?from=RDU"
	expected = model.Filter{
		ObjectName:    "clickstream",
		Fields:        []string{"to", "from"},
		ExcludeFields: false,
		Condition:     "from=RDU",
	}
	isAdmin, filters = parsePermission(clickstreamSomeFieldsFromRDU)
	if isAdmin || len(filters) != 1 || !areFiltersEqual(filters["clickstream"], expected) {
		t.Errorf("[TestParsePermission] Permission was not parsed as expected: %v", clickstreamSomeFieldsFromRDU)
	}

	// Multiple objects with include and exclude
	obj1 := "clickstream/to,from?from=RDU or to=RDU"
	obj2 := "traveller/*,BirthDate"
	multiple := obj1 + "\n" + obj2
	expected1 := model.Filter{
		ObjectName:    "clickstream",
		Fields:        []string{"to", "from"},
		ExcludeFields: false,
		Condition:     "from=RDU or to=RDU",
	}
	expected2 := model.Filter{
		ObjectName:    "traveller",
		Fields:        []string{"*", "BirthDate"},
		ExcludeFields: true,
		Condition:     "",
	}
	isAdmin, filters = parsePermission(multiple)
	if isAdmin || len(filters) != 2 || !areFiltersEqual(filters["clickstream"], expected1) || !areFiltersEqual(filters["traveller"], expected2) {
		t.Errorf("[TestParsePermission] Permission was not parsed as expected: %v", multiple)
	}
}

func areFiltersEqual(actual, expected model.Filter) bool {
	if actual.ObjectName == expected.ObjectName &&
		isEqual(actual.Fields, expected.Fields) &&
		actual.ExcludeFields == expected.ExcludeFields &&
		actual.Condition == expected.Condition {
		return true
	}

	log.Printf("Actual: %v", actual)
	log.Printf("Expected: %v", expected)
	return false
}

var traveller1 model.Traveller = model.Traveller{
	FirstName: "John",
	LastName:  "Doe",
	PMSID:     "123",
	BirthDate: time.Now(),
	AirBookingRecords: []model.AirBooking{
		{
			SegmentID:    "id-123",
			From:         "RDU",
			To:           "BOS",
			FlightNumber: "1234",
			Status:       "CONFIRMED",
		},
		{
			SegmentID:    "id-456",
			From:         "BOS",
			To:           "RDU",
			FlightNumber: "6789",
			Status:       "BOOKED",
		},
		{
			SegmentID:    "id-789",
			From:         "JFK",
			To:           "MIA",
			FlightNumber: "4321",
			Status:       "TBD",
		},
	},
}

func TestEval(t *testing.T) {
	// Test cases with only binary operators
	context1 := map[string]interface{}{"a": 1, "b": 2}
	testCases1 := []struct {
		expr   string
		result bool
	}{
		{"a = 1", true},
		{"b != 1", true},
		{"a in [1,2,3]", true},
		{"b in [1,2,3]", true},
		{"a in [4,5,6]", false},
		{"b in [4,5,6]", false},
	}
	for _, tc := range testCases1 {
		if shouldIncludeObject(tc.expr, context1) != tc.result {
			t.Errorf("EvalBoolExpr(%v, %v) = %v; want %v", tc.expr, context1, !tc.result, tc.result)
		}
	}

	// Test cases with complex expressions
	d := time.Date(2022, time.June, 30, 0, 0, 0, 0, time.UTC)
	context2 := map[string]interface{}{"a": "RDU", "b": 5, "c": d, "d": 100}
	testCases2 := []struct {
		expr   string
		result bool
	}{
		{"a = RDU", true},
		{"a = RDU or b = 5", true},
		{"a = RDU and b in [3,4,5,6]", true},
		{"a = RDU and b in [1,2,3,4]", false},
		{"(a = RDU or a = BOS) and b > 4", true},
		{"(a = BOS or a = BOS) and b > 4", false},
		{"(a = RDU) or (c > 2010/01/01)", true},
		{"(a = RDU) and (b in [1,2,3])", false},
		{"(a != BOS) or (b in [1,2,3])", true},
		{"(a != RDU) or (b in [1,2,3] and d > 100.01)", false},
		{"(a != RDU) or (b in [1,2,3] and d = 100.0)", false},
		{"(a != RDU) or (b in [4,5,6] and d = 100.0)", false},
		{"(a != BOS) and ((b in [1,2,3]) or (c > 2010/01/01))", true},
		{"(a != BOS) and ((b in [1,2,3]) or (c < 2010/01/01))", false},
	}
	for _, tc := range testCases2 {
		if shouldIncludeObject(tc.expr, context2) != tc.result {
			t.Errorf("EvalBoolExpr: %v, %v = %v; want %v", tc.expr, context2, !tc.result, tc.result)
		}
	}
}

func TestFilter(t *testing.T) {
	adminPermission := "*/*"
	zeroPermission := ""
	noMatchingTravellerPermission := "traveller/*?pmsId = xyz\nair_booking/*?from = RDU"
	allFromRDU := "traveller/*\nair_booking/*?from = RDU"
	complex := "traveller/firstName,lastName\nair_booking/to,from?from in [RDU,JFK]"

	adminTraveller := Filter(traveller1, adminPermission)
	zeroTraveller := Filter(traveller1, zeroPermission)
	noMatchingTraveller := Filter(traveller1, noMatchingTravellerPermission)
	allFromRDUTraveller := Filter(traveller1, allFromRDU)
	complexTraveller := Filter(traveller1, complex)

	defaultTime := time.Time{}

	if adminTraveller.FirstName != traveller1.FirstName ||
		adminTraveller.LastName != traveller1.LastName ||
		adminTraveller.BirthDate != traveller1.BirthDate ||
		len(adminTraveller.AirBookingRecords) != len(traveller1.AirBookingRecords) ||
		adminTraveller.AirBookingRecords[0] != traveller1.AirBookingRecords[0] ||
		adminTraveller.AirBookingRecords[1] != traveller1.AirBookingRecords[1] ||
		adminTraveller.AirBookingRecords[2] != traveller1.AirBookingRecords[2] {
		t.Errorf("Objects do not match as expected")
	}

	if zeroTraveller != nil {
		t.Errorf("Objects do not match as expected")
	}

	if noMatchingTraveller != nil {
		t.Errorf("Objects do not match as expected")
	}

	if allFromRDUTraveller.FirstName != traveller1.FirstName ||
		allFromRDUTraveller.LastName != traveller1.LastName ||
		allFromRDUTraveller.BirthDate != traveller1.BirthDate ||
		len(allFromRDUTraveller.AirBookingRecords) != 1 ||
		allFromRDUTraveller.AirBookingRecords[0].From != traveller1.AirBookingRecords[0].From ||
		allFromRDUTraveller.AirBookingRecords[0].To != traveller1.AirBookingRecords[0].To ||
		allFromRDUTraveller.AirBookingRecords[0].BookingID != traveller1.AirBookingRecords[0].BookingID ||
		allFromRDUTraveller.AirBookingRecords[0].Status != traveller1.AirBookingRecords[0].Status {
		t.Errorf("Objects do not match as expected")
		log.Println(adminTraveller)
		log.Println(traveller1)
	}

	if complexTraveller.FirstName != traveller1.FirstName ||
		complexTraveller.LastName != traveller1.LastName ||
		complexTraveller.BirthDate != defaultTime ||
		len(complexTraveller.AirBookingRecords) != 2 ||
		complexTraveller.AirBookingRecords[0].To != traveller1.AirBookingRecords[0].To ||
		complexTraveller.AirBookingRecords[0].From != traveller1.AirBookingRecords[0].From ||
		complexTraveller.AirBookingRecords[0].SegmentID != "" ||
		complexTraveller.AirBookingRecords[1].To != traveller1.AirBookingRecords[2].To ||
		complexTraveller.AirBookingRecords[1].From != traveller1.AirBookingRecords[2].From ||
		complexTraveller.AirBookingRecords[1].SegmentID != "" {
		t.Errorf("Objects do not match as expected")
	}

	// log.Printf("Admin DOB: %v", adminTraveller.BirthDate)
	// log.Printf("Zero DOB: %v", zeroTraveller.BirthDate)
	// log.Printf("RDU DOB: %v", allFromRDUTraveller.BirthDate)
	// log.Printf("Complex DOB: %v", complexTraveller.BirthDate)

	// log.Printf("Admin Air Booking:\n%v", adminTraveller.AirBookingRecords)
	// log.Printf("Zero Air Booking:\n%v", zeroTraveller.AirBookingRecords)
	// log.Printf("RDU Air Booking:\n%v", allFromRDUTraveller.AirBookingRecords)
	// log.Printf("Complex Air Booking:\n%v", complexTraveller.AirBookingRecords)
}

// Test equality between two slices of the same type
func isEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
