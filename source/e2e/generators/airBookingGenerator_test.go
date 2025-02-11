// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package generators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAirBooking(t *testing.T) {
	fieldOverrides1 := []FieldOverrides{
		{
			Fields: []FieldOverride{
				{Name: "MainEmail", Value: "replaced@example.com"},
				{Name: "FirstName", Value: "First"},
				{Name: "LastName", Value: "Last"},
			},
			Weight: 100000,
		},
		{
			Fields: []FieldOverride{
				{Name: "MainEmail", Value: "replaced2@example.com"},
				{Name: "FirstName", Value: "First2"},
				{Name: "LastName", Value: "Last2"},
			},
			Weight: 0,
		},
	}
	booking1 := GenerateAirBooking(1, fieldOverrides1...)
	assert.True(t, booking1.MainEmail == "replaced@example.com", "expected %s, got %s", "replaced@example.com", booking1.MainEmail)
	assert.True(
		t,
		booking1.PassengerInfo.Passengers[0].FirstName == "First",
		"expected %s, got %s",
		"First",
		booking1.PassengerInfo.Passengers[0].FirstName,
	)
	assert.True(
		t,
		booking1.PassengerInfo.Passengers[0].LastName == "Last",
		"expected %s, got %s",
		"Last",
		booking1.PassengerInfo.Passengers[0].LastName,
	)

	fieldOverrides2 := []FieldOverrides{
		{
			Fields: []FieldOverride{
				{Name: "MainEmail", Value: "replaced@example.com"},
				{Name: "FirstName", Value: "First"},
				{Name: "LastName", Value: "Last"},
			},
			Weight: 0,
		},
		{
			Fields: []FieldOverride{
				{Name: "MainEmail", Value: "replaced2@example.com"},
				{Name: "FirstName", Value: "First2"},
				{Name: "LastName", Value: "Last2"},
			},
			Weight: 100000,
		},
	}
	booking2 := GenerateAirBooking(1, fieldOverrides2...)

	assert.True(t, booking2.MainEmail == "replaced2@example.com", "expected %s, got %s", "replaced2@example.com", booking2.MainEmail)
	assert.True(
		t,
		booking2.PassengerInfo.Passengers[0].FirstName == "First2",
		"expected %s, got %s",
		"First2",
		booking2.PassengerInfo.Passengers[0].FirstName,
	)
	assert.True(
		t,
		booking2.PassengerInfo.Passengers[0].LastName == "Last2",
		"expected %s, got %s",
		"Last2",
		booking2.PassengerInfo.Passengers[0].LastName,
	)
}
