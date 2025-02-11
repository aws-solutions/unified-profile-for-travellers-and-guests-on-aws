// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package generators

import (
	"fmt"
	"tah/upt/source/tah-core/kinesis"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

// Simplified Hotel Stay Object
type HotelStay struct {
	ModelVersion string            `json:"modelVersion"`
	ID           string            `json:"id"`
	GuestID      string            `json:"guestId"`
	Revenue      []StayRevenueItem `json:"revenue"`
	ExtendedData interface{}       `json:"extendedData"`
}

type StayRevenueItem struct {
	Type string    `json:"type"`
	Date time.Time `json:"date"`
}

func GenerateSimpleHotelStayRecords(travellerId, domainName string, numRecords int) ([]kinesis.Record, error) {
	if numRecords > 10000 {
		return []kinesis.Record{}, fmt.Errorf("too many records, num records should be lower than or equal to 10000")
	}
	recs := []kinesis.Record{}
	for i := 0; i < numRecords; i++ {
		rec, err := buildSimpleHotelStayRecord(travellerId, domainName)
		if err != nil {
			continue
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func GenerateSpecificHotelStayRecords(travellerId, accpId, domainName string) ([]kinesis.Record, error) {
	recs := []kinesis.Record{}
	rec, err := buildSpecificHotelStayRecord(travellerId, accpId, domainName)
	if err != nil {
		return []kinesis.Record{}, err
	}
	recs = append(recs, rec)

	return recs, nil
}

func buildSimpleHotelStayRecord(travellerId, domainName string) (kinesis.Record, error) {
	revenueItem := buildRevenueItem()
	stay := HotelStay{
		ModelVersion: "1.0",
		ID:           gofakeit.UUID(),
		GuestID:      travellerId,
		Revenue:      []StayRevenueItem{revenueItem},
	}
	return GetKinesisRecord(domainName, "hotel_stay", stay)
}

func buildSpecificHotelStayRecord(travellerId, accpId, domainName string) (kinesis.Record, error) {
	revenueItem := buildRevenueItem()
	stay := HotelStay{
		ModelVersion: "1.0",
		ID:           accpId,
		GuestID:      travellerId,
		Revenue:      []StayRevenueItem{revenueItem},
	}
	return GetKinesisRecord(domainName, "hotel_stay", stay)
}

func buildRevenueItem() StayRevenueItem {
	now := time.Now()
	return StayRevenueItem{
		Type: gofakeit.RandomString([]string{"snack bar", "just-walk-out-purchase", "restaurant", "sodas"}),
		Date: gofakeit.DateRange(now.AddDate(-1, 0, 0), now.AddDate(0, 0, -1)),
	}
}
