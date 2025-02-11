// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"encoding/json"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	commonServices "tah/upt/source/ucp-common/src/services/admin"
)

type EmptyDynamoTableUsecase struct {
	name string
}

func InitEmptyDynamoTable() *EmptyDynamoTableUsecase {
	return &EmptyDynamoTableUsecase{name: "EmptyDynamoTable"}
}

func (u *EmptyDynamoTableUsecase) Name() string {
	return u.name
}

func (u *EmptyDynamoTableUsecase) Execute(payload async.AsyncInvokePayload, services async.Services, tx core.Transaction) error {
	// Convert payload from string to json data
	payloadJson, err := json.Marshal(payload.Body)
	if err != nil {
		tx.Error("[%s] Error marshalling payload %v", u.name, err)
	}

	// Convert json to EmptyTableBody
	data := uc.EmptyTableBody{}
	err = json.Unmarshal(payloadJson, &data)
	if err != nil {
		tx.Error("[%s] Error unmarshalling payload %v", u.name, err)
	}

	// Replace table
	dbConfig := db.Init(data.TableName, "", "", "", "") // only table name is used, keys will be based on existing table
	err = dbConfig.ReplaceTable()                       // this will take several seconds
	if err != nil {
		tx.Error("[%s] Error replacing DynamoDB Table %s", u.name, err)
		return err
	}
	if dbConfig.TableName == services.ErrorDB.TableName {
		tx.Info("[%s] Table being recreated is the Error DB table. Recreating the event source Mapping", u.name)
		//defined in common.go
		_, err = commonServices.CreateRetryLambdaEventSourceMapping(services.RetryLambdaConfig, services.ErrorDB, tx)
		if err != nil {
			tx.Error("[%s] Error creating ACCP event stream %s", u.name, err)
		}
	} else {
		tx.Info("[%s] The table recreated is not the Error DB table. no need to recreate the eventsource stream", u.name)
	}

	return err
}
