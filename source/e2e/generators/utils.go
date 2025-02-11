// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package generators

import (
	"encoding/json"
	"tah/upt/source/tah-core/kinesis"

	"github.com/brianvoe/gofakeit/v6"
)

func GetKinesisRecord[T any](domainName, objType string, record T) (kinesis.Record, error) {
	bytes, err := json.Marshal(record)
	if err != nil {
		return kinesis.Record{}, err
	}
	var data map[string]interface{}
	json.Unmarshal(bytes, &data)

	obj := BusinessObjectRecord{
		Domain:       domainName,
		ObjectType:   objType,
		ModelVersion: "1.0",
		Data:         data,
	}

	jsonData, err := json.Marshal(obj)
	if err != nil {
		return kinesis.Record{}, err
	}

	return kinesis.Record{
		Pk:   gofakeit.UUID(),
		Data: string(jsonData),
	}, nil
}
