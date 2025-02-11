// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"errors"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
)

/**
 * Creates the event source mapping for the retry lambda if it does not already exist
 * this function is necessary because when we empty the error dynamoDB, we do it by recreating the table
 * recreating a DYnamoDB table with stream enabled generate a new stream ARN which means that any lambda consumer of this
 * stream must get a new EventSourceMapping
**/
func CreateRetryLambdaEventSourceMapping(lambdaCfg lambda.IConfig, dbConfig db.DBConfig, tx core.Transaction) (bool, error) {
	retryLambdaName := lambdaCfg.Get("FunctionName")
	errorTableName := dbConfig.TableName
	tx.Info("[CreateUcpDomain][eventSourceMapping] Creating Retry Lambda Event Mapping (if not exists) for lambda %v and table %v", retryLambdaName, errorTableName)
	streamArn, err := dbConfig.GetStreamArn()
	if err != nil {
		tx.Error("[CreateUcpDomain][eventSourceMapping] error getting stream arn: %v", err)
		return false, errors.New("retry Lambda Event Mapping not created because stream arn could not be fetched")
	}

	if streamArn == "" {
		tx.Error("[CreateUcpDomain][eventSourceMapping] Retry Lambda Event Mapping not created because Error table streamArn is empty")
		return false, errors.New("retry Lambda Event Mapping not created because Error table streamArn is empty")
	}

	mappings, err := lambdaCfg.SearchDynamoEventSourceMappings(streamArn, retryLambdaName)
	if err != nil {
		tx.Error("[CreateUcpDomain][eventSourceMapping] Error looking for exising eventSource mapping for retry lambda %v", err)
		return false, err
	}
	if len(mappings) > 0 {
		tx.Debug("[CreateUcpDomain][eventSourceMapping] Eventsource mapping already exists for lambda %v and stream %v", retryLambdaName, streamArn)
		return false, nil
	}
	tx.Debug("[CreateUcpDomain][eventSourceMapping] Creating Retry Lambda Event Mapping for lambda %v and table %v", retryLambdaName, errorTableName)
	err = lambdaCfg.CreateDynamoEventSourceMapping(streamArn, retryLambdaName)
	if err != nil {
		tx.Error("[CreateUcpDomain][eventSourceMapping] Error creating Retry Lambda Event Mapping for lambda %v and table %v", retryLambdaName, errorTableName)
		return false, err
	}
	tx.Info("[CreateUcpDomain][eventSourceMapping] Successfully created Retry Lambda Event Mapping for lambda %v and table %v", retryLambdaName, errorTableName)
	return true, nil
}

// this function checks wether an event source mapping exist on the retry lambda function
func HasEventSourceMapping(lambdaCfg lambda.IConfig, dbConfig db.DBConfig, tx core.Transaction) (bool, error) {
	retryLambdaName := lambdaCfg.Get("FunctionName")
	errorTableName := dbConfig.TableName
	tx.Debug("[CreateUcpDomain][eventSourceMapping] Creating Retry Lambda Event Mapping (if not exists) for lambda %v and table %v", retryLambdaName, errorTableName)
	streamArn, err := dbConfig.GetStreamArn()
	if err != nil {
		tx.Error("[CreateUcpDomain][eventSourceMapping] error getting stream arn: %v", err)
		return false, errors.New("retry Lambda Event Mapping not created because stream arn could not be fetched")
	}
	if streamArn == "" {
		tx.Error("[CreateUcpDomain][eventSourceMapping] Error table streamArn is empty")
		return false, errors.New("error table streamArn is empty")
	}

	mappings, err := lambdaCfg.SearchDynamoEventSourceMappings(streamArn, retryLambdaName)
	if err != nil {
		tx.Error("[CreateUcpDomain][eventSourceMapping] Error looking for exising eventSource mapping for retry lambda %v", err)
		return false, err
	}
	if len(mappings) > 0 {
		tx.Debug("[CreateUcpDomain][eventSourceMapping] Eventsource mapping exists for lambda %v and stream %v", retryLambdaName, streamArn)
		return true, nil
	}
	return false, nil
}
