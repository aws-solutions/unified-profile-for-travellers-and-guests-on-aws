// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package generators

import (
	"fmt"
	"tah/upt/schemas/src/tah-common/common"
	"tah/upt/source/tah-core/kinesis"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

// Simplified Clickstream Object
type Clickstream struct {
	ModelVersion   string       `json:"modelVersion"`
	Session        EventSession `json:"session"`
	UserID         string       `json:"userId,omitempty"` //if known, this field identifies the user unambigiously (can come from a CDP or similar application)
	EventTimestamp time.Time    `json:"event_timestamp"`
}
type EventSession struct {
	ID string `json:"id"`
}

func GenerateSimpleClickstreamRecords(travellerId, domainName string, numRecords int) ([]kinesis.Record, error) {
	if numRecords > 10000 {
		return []kinesis.Record{}, fmt.Errorf("too many records, num records should be lower than or equal to 10000")
	}
	recs := []kinesis.Record{}
	for i := 0; i < numRecords; i++ {
		rec, err := buildSimpleClickstreamRecord(travellerId, domainName)
		if err != nil {
			continue
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func buildSimpleClickstreamRecord(travellerId, domainName string) (kinesis.Record, error) {
	clickstream := Clickstream{
		ModelVersion:   "1.0",
		Session:        EventSession{ID: travellerId},
		EventTimestamp: time.Now().Add(-time.Second * time.Duration(gofakeit.IntRange(0, 10000))),
	}
	return GetKinesisRecord(domainName, "clickstream", clickstream)
}

func BuildSpecificClickstreamRecord(travellerId, domainName string, profileOverrides ...func(*common.ClickEvent)) (kinesis.Record, error) {
	click := common.ClickEvent{
		ModelVersion:   "1.0",
		Session:        common.EventSession{ID: travellerId},
		EventTimestamp: time.Now().Add(-time.Second * time.Duration(gofakeit.IntRange(0, 10000))),
	}
	for _, override := range profileOverrides {
		override(&click)
	}
	return GetKinesisRecord(domainName, "clickstream", click)
}
