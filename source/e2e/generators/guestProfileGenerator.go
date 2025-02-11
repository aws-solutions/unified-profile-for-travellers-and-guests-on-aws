// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package generators

import (
	"fmt"
	"tah/upt/schemas/src/tah-common/lodging"
	"tah/upt/source/tah-core/kinesis"
	"time"

	common "tah/upt/schemas/src/tah-common/common"

	"github.com/brianvoe/gofakeit/v6"
)

func GenerateSimpleGuestProfileRecords(travellerId, domainName string, numRecords int) ([]kinesis.Record, error) {
	if numRecords > 100 {
		return []kinesis.Record{}, fmt.Errorf("too many records, num records should be lower than or equal to 100")
	}
	recs := []kinesis.Record{}
	for i := 0; i < numRecords; i++ {
		rec, err := buildSimpleGuestProfileRecord(travellerId, domainName)
		if err != nil {
			continue
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func BuildGuest(guest_id string) lodging.GuestProfile {
	now := time.Now()
	createdOn := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))
	return lodging.GuestProfile{
		ModelVersion:    "1.0",
		ID:              guest_id,
		FirstName:       gofakeit.FirstName(),
		LastName:        gofakeit.LastName(),
		LastUpdatedOn:   createdOn.AddDate(0, gofakeit.Number(1, 4), 0),
		CreatedOn:       createdOn,
		LastUpdatedBy:   gofakeit.Name(),
		CreatedBy:       gofakeit.Name(),
		Honorific:       gofakeit.RandomString([]string{"Mrs", "Mr", "Miss", "Sir", "Ms"}),
		LoyaltyPrograms: []lodging.LoyaltyProgram{},
	}
}

func BuildEmail(emailAddress string) common.Email {
	return common.Email{
		Type:    "work",
		Address: emailAddress,
		Primary: true,
	}
}

func buildSimpleGuestProfileRecord(travellerId, domainName string) (kinesis.Record, error) {
	guest := BuildGuest(travellerId)
	return GetKinesisRecord(domainName, "guest_profile", guest)
}

func GenerateSpecificGuestProfileRecord(travellerId, domainName string, profileOverrides ...func(*lodging.GuestProfile)) ([]kinesis.Record, error) {
	guest := BuildGuest(travellerId)
	for _, override := range profileOverrides {
		override(&guest)
	}
	record, err := GetKinesisRecord(domainName, "guest_profile", guest)
	return []kinesis.Record{record}, err
}
