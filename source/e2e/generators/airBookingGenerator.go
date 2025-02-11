// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package generators

import (
	"fmt"
	"tah/upt/schemas/generators/real-time/src/example"
	"tah/upt/schemas/src/tah-common/air"
	"tah/upt/source/tah-core/kinesis"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func GenerateAirBooking(nPax int, fieldOverrides ...FieldOverrides) air.Booking {
	paxProfiles := make([]air.PassengerProfile, nPax)
	for i := 0; i < nPax; i++ {
		paxProfiles[i] = example.CreatePassengerProfileExample(1)
	}
	booking := example.CreateAirBookingExample(paxProfiles)

	applyFieldOverrides(&booking, fieldOverrides)

	return booking
}

func GenerateRandomSimpleAirBooking(domainName string, fieldOverrides ...FieldOverrides) air.Booking {
	booking := buildSimpleAirBookingRecord(gofakeit.UUID())
	applyFieldOverrides(&booking, fieldOverrides)
	return booking
}

func GenerateSimpleAirBookingRecords(travellerId, domainName string, numRecords int) ([]kinesis.Record, error) {
	if numRecords > 10000 {
		return []kinesis.Record{}, fmt.Errorf("too many records, num records should be lower than or equal to 10000")
	}
	recs := []kinesis.Record{}
	for i := 0; i < numRecords; i++ {
		booking := buildSimpleAirBookingRecord(travellerId)
		rec, err := GetKinesisRecord(domainName, "air_booking", booking)
		if err != nil {
			continue
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func buildSimpleAirBookingRecord(travellerId string) air.Booking {
	defaultPax := BuildPassenger(travellerId)
	return air.Booking{
		ModelVersion:  "1.0",
		ID:            gofakeit.Regex("[0-9A-Z]{6}"),
		PassengerInfo: air.PassengerInfo{Passengers: []air.PassengerProfile{defaultPax}},
		Itinerary:     buildItinerary(),
		LastUpdatedOn: time.Now(),
	}
}

func buildItinerary() air.Itinerary {
	from := example.AirportCode()
	to := example.AirportCode()
	return air.Itinerary{
		Segments: []air.FlightSegment{buildAirSegment(from, to)},
	}
}

func buildAirSegment(from, to string) air.FlightSegment {
	return air.FlightSegment{
		From: from,
		To:   to,
	}
}
