// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package generators

import (
	"encoding/json"
	"fmt"
	"tah/upt/schemas/src/tah-common/air"
	"tah/upt/source/tah-core/kinesis"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

// Simplified Passenger Profile Object
type PassengerProfile struct {
	ModelVersion  string    `json:"modelVersion"`
	ID            string    `json:"id"`
	LastUpdatedOn time.Time `json:"lastUpdatedOn"`
	CreatedOn     time.Time `json:"createdOn"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
	CreatedBy     string    `json:"createdBy"`
	IsBooker      bool      `json:"isBooker"` //when set to true within a booking, this attribute indicates that this passenger is the booker

	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`

	Honorific string `json:"honorific"`
}

func (gp PassengerProfile) MarshalJSON() ([]byte, error) {
	type Alias PassengerProfile
	return json.Marshal(&struct {
		Alias
		LastUpdatedOn string `json:"lastUpdatedOn"`
	}{
		Alias:         (Alias)(gp),
		LastUpdatedOn: gp.LastUpdatedOn.Format(INGEST_TIMESTAMP_FORMAT),
	})
}

func GenerateSimplePaxProfileRecords(travellerId, domainName string, numRecords int) ([]kinesis.Record, error) {
	if numRecords > 100 {
		return []kinesis.Record{}, fmt.Errorf("too many records, num records should be lower than or equal to 10000")
	}
	recs := []kinesis.Record{}
	for i := 0; i < numRecords; i++ {
		rec, err := buildSimplePaxProfileRecord(travellerId, domainName)
		if err != nil {
			continue
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func BuildPassenger(pax_id string) air.PassengerProfile {
	now := time.Now()
	createdOn := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))
	return air.PassengerProfile{
		ModelVersion:  "1.0",
		ID:            pax_id,
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		LastUpdatedOn: createdOn.AddDate(0, gofakeit.Number(1, 4), 0),
		CreatedOn:     createdOn,
		LastUpdatedBy: gofakeit.Name(),
		CreatedBy:     gofakeit.Name(),
		Honorific:     gofakeit.RandomString([]string{"Mrs", "Mr", "Miss", "Sir", "Ms"}),
	}
}

func buildSimplePaxProfileRecord(travellerId, domainName string) (kinesis.Record, error) {
	pax := BuildPassenger(travellerId)
	return GetKinesisRecord(domainName, "pax_profile", pax)
}
