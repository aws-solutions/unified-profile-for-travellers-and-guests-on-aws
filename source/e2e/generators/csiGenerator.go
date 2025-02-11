// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package generators

import (
	"fmt"
	csi "tah/upt/schemas/src/tah-common/common"
	"tah/upt/source/tah-core/kinesis"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func GenerateSimpleCsiRecords(domainName string, numRecords int) ([]kinesis.Record, error) {
	if numRecords > 10000 {
		return []kinesis.Record{}, fmt.Errorf("too many records, num records should be lower than or equal to 10000")
	}
	recs := []kinesis.Record{}
	for i := 0; i < numRecords; i++ {
		rec, err := buildSimpleCsiRecord(domainName)
		if err != nil {
			continue
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func buildSimpleCsiRecord(domainName string) (kinesis.Record, error) {
	conversation := csi.ConversationItem{ID: gofakeit.UUID()}
	csi := csi.CustomerServiceInteraction{
		ModelVersion: "1.0",
		Conversation: []csi.ConversationItem{conversation},
		SessionID:    gofakeit.Regex("[0-9A-Z]{10}"),
	}
	return GetKinesisRecord(domainName, "customer_service_interaction", csi)
}

func GenerateSpecificCsiRecord(travellerId, domainName string, overrides ...func(*csi.CustomerServiceInteraction)) ([]kinesis.Record, error) {
	recs := []kinesis.Record{}
	rec, err := buildSpecificCsiRecord(travellerId, domainName, overrides...)
	if err != nil {
		return []kinesis.Record{}, err
	}
	recs = append(recs, rec)
	return recs, nil
}

func buildSpecificCsiRecord(travellerId, domainName string, overrides ...func(*csi.CustomerServiceInteraction)) (kinesis.Record, error) {
	conversation := csi.ConversationItem{ID: gofakeit.UUID()}
	now := time.Now()
	createdOn := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))
	csi := csi.CustomerServiceInteraction{
		ModelVersion: "1.0",
		Conversation: []csi.ConversationItem{conversation},
		SessionID:    travellerId,
		EndTime:      createdOn,
	}
	// Apply overrides
	for _, override := range overrides {
		override(&csi)
	}
	return GetKinesisRecord(domainName, "customer_service_interaction", csi)
}
