// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package generators

import (
	"fmt"
	"tah/upt/schemas/generators/real-time/src/example"
	"tah/upt/schemas/src/tah-common/lodging"
	"tah/upt/source/tah-core/kinesis"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

const ID_REGREX = "[0-9A-Z]{10}"

func GenerateHotelBooking(basicGuest, basicBooking bool, fieldOverrides ...FieldOverrides) lodging.Booking {
	guestProfile := example.CreateGuestProfileExample(0, basicGuest)
	booking := example.CreateHotelBookingExample([]lodging.GuestProfile{guestProfile}, basicBooking)
	applyFieldOverrides(&booking, fieldOverrides)

	return booking
}

func GenerateSimpleHotelBookingRecords(travellerId, domainName string, numRecords int) ([]kinesis.Record, error) {
	if numRecords > 10000 {
		return []kinesis.Record{}, fmt.Errorf("too many records, num records should be lower than or equal to 10000")
	}
	recs := []kinesis.Record{}
	for i := 0; i < numRecords; i++ {
		rec, err := buildSimpleHotelBookingRecord(travellerId, domainName)
		if err != nil {
			continue
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func buildSimpleHotelBookingRecord(travellerId, domainName string) (kinesis.Record, error) {
	hotelCode := example.HotelCode()
	segment := buildBookingSegment(travellerId, hotelCode)
	hotelBooking := lodging.Booking{
		ModelVersion:  "1.0",
		ID:            gofakeit.Regex(ID_REGREX),
		Segments:      []lodging.BookingSegment{segment},
		HotelCode:     hotelCode,
		LastUpdatedOn: time.Now(),
	}
	return GetKinesisRecord(domainName, "hotel_booking", hotelBooking)
}

func buildBookingSegment(travellerId, hotelCode string) lodging.BookingSegment {
	return lodging.BookingSegment{
		ID:        gofakeit.Regex(ID_REGREX),
		HotelCode: hotelCode,
		Holder:    BuildGuest(travellerId),
		Products:  []lodging.Product{buildProduct()},
	}
}

func buildProduct() lodging.Product {
	return lodging.Product{
		ID: gofakeit.Regex(ID_REGREX),
	}
}
