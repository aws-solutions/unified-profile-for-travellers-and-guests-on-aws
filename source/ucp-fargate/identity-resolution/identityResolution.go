// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package ir

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"

	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/aurora"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	secrets "tah/upt/source/tah-core/secret"
	"tah/upt/source/tah-core/sqs"

	"github.com/google/uuid"
)

var AWS_REGION = os.Getenv("AWS_REGION")
var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
var METRICS_NAMESPACE = "upt"

// Low cost storage
var AURORA_PROXY_ENDPOINT = os.Getenv("AURORA_PROXY_ENDPOINT")
var AURORA_DB_NAME = os.Getenv("AURORA_DB_NAME")
var AURORA_DB_SECRET_ARN = os.Getenv("AURORA_DB_SECRET_ARN")
var STORAGE_CONFIG_TABLE_NAME = os.Getenv("STORAGE_CONFIG_TABLE_NAME")
var STORAGE_CONFIG_TABLE_PK = os.Getenv("STORAGE_CONFIG_TABLE_PK")
var STORAGE_CONFIG_TABLE_SK = os.Getenv("STORAGE_CONFIG_TABLE_SK")
var CHANGE_PROC_KINESIS_STREAM_NAME = os.Getenv("CHANGE_PROC_KINESIS_STREAM_NAME")
var MERGE_QUEUE_URL = os.Getenv("MERGE_QUEUE_URL")
var CP_WRITER_QUEUE_URL = os.Getenv("CP_WRITER_QUEUE_URL")
var LCS_CACHE = customerprofileslcs.InitLcsCache()

var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

// Init low-cost clients
var smClient = secrets.InitWithRegion(AURORA_DB_SECRET_ARN, AWS_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var user, pwd = getUserCredentials(smClient)
var auroraClient = aurora.InitPostgresql(AURORA_PROXY_ENDPOINT, AURORA_DB_NAME, user, pwd, 50, LOG_LEVEL)
var configTableClient = db.Init(
	STORAGE_CONFIG_TABLE_NAME,
	STORAGE_CONFIG_TABLE_PK,
	STORAGE_CONFIG_TABLE_SK,
	METRICS_SOLUTION_ID,
	METRICS_SOLUTION_VERSION,
)
var kinesisClient = kinesis.Init(CHANGE_PROC_KINESIS_STREAM_NAME, AWS_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, LOG_LEVEL)
var mergeQueueClient = sqs.InitWithQueueUrl(MERGE_QUEUE_URL, AWS_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var cpWriterClient = sqs.InitWithQueueUrl(CP_WRITER_QUEUE_URL, AWS_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

func getUserCredentials(sm secrets.Config) (string, string) {
	user := sm.Get("username")
	pwd := sm.Get("password")
	return user, pwd
}

func IdentityResolutionMain() error {
	tx := core.NewTransaction("IdentityResolution", "", LOG_LEVEL)
	tx.Info("[FargateTask] Starting IR Fargate Task")
	cp := initStorage()
	err := HandleWithServices(cp, tx)
	if err != nil {
		return err
	}
	return nil
}

// This fargate task is to be run after all existing indices pertaining to deleted rules are removed
func HandleWithServices(cp customerprofileslcs.ICustomerProfileLowCostConfig, tx core.Transaction) error {
	// Task Level Variables
	var DOMAIN_NAME = os.Getenv("DOMAIN_NAME")
	var OBJECT_TYPE_NAME = os.Getenv("OBJECT_TYPE_NAME")
	var RULE_SET_VERSION = os.Getenv("RULE_SET_VERSION")
	var RULE_ID = os.Getenv("RULE_ID")
	var PARTITION_LOWER_BOUND = os.Getenv("PARTITION_LOWER_BOUND")
	var PARTITION_UPPER_BOUND = os.Getenv("PARTITION_UPPER_BOUND")
	// var PARTITION_ID = os.Getenv("PARTITION_ID")
	// var PARTITION_SIZE = os.Getenv("PARTITION_SIZE")
	tx.Debug("[FargateTask] DOMAIN_NAME: %v", DOMAIN_NAME)
	tx.Debug("[FargateTask] OBJECT_TYPE_NAME: %v", OBJECT_TYPE_NAME)
	tx.Debug("[FargateTask] RULE_SET_VERSION: %v", RULE_SET_VERSION)
	tx.Debug("[FargateTask] RULE_ID: %v", RULE_ID)
	tx.Debug("[FargateTask] PARTITION_LOWER_BOUND: %v", PARTITION_LOWER_BOUND)
	tx.Debug("[FargateTask] PARTITION_UPPER_BOUND: %v", PARTITION_UPPER_BOUND)

	// Validate Task Level Variables are set
	if DOMAIN_NAME == "" {
		return errors.New("domain name should be provided")
	}
	if OBJECT_TYPE_NAME == "" {
		return errors.New("object type name should be provided")
	}
	if RULE_SET_VERSION == "" {
		return errors.New("rule set version should be provided")
	}
	if RULE_ID == "" {
		return errors.New("rule id should be provided")
	}
	cp.SetDomain(DOMAIN_NAME)
	ruleSets, err := cp.ListIdResRuleSets(true)
	if err != nil {
		return err
	}
	idx := slices.IndexFunc(ruleSets, func(ruleSet customerprofileslcs.RuleSet) bool { return ruleSet.Name == "active" })
	if idx == -1 {
		return errors.New("no active rule set found")
	}
	currentActive := ruleSets[idx]
	tx.Debug("[FargateTask] current active rule set: %v", currentActive)

	if fmt.Sprint(currentActive.LatestVersion) != RULE_SET_VERSION {
		tx.Error("[FargateTask] Rule Set Version Mismatch")
		return errors.New("rule set version mismatch")
		// Cancel the task, new active rule set was activated
	}
	tx.Debug("[FargateTask] current active version: %v", currentActive.LatestVersion)

	ruleId, err := strconv.Atoi(RULE_ID)
	if err != nil {
		return err
	}
	rule := currentActive.Rules[ruleId]
	tx.Debug("[FargateTask] Rule: %v", rule)
	// Get indexing for rule

	lowerBound, err := uuid.Parse(PARTITION_LOWER_BOUND)
	if err != nil {
		return errors.Join(errors.New("invalid lower bound"), err)
	}
	upperBound, err := uuid.Parse(PARTITION_UPPER_BOUND)
	if err != nil {
		return errors.Join(errors.New("invalid upper bound"), err)
	}
	partition := customerprofileslcs.UuidPartition{LowerBound: lowerBound, UpperBound: upperBound}
	pageSize := 1000
	objects, err := cp.GetInteractionTable(DOMAIN_NAME, OBJECT_TYPE_NAME, partition, uuid.Nil, pageSize)
	if err != nil {
		tx.Error("error getting %v object table from domain %v", OBJECT_TYPE_NAME, DOMAIN_NAME)
		return err
	}
	for len(objects) > 0 {
		tx.Debug("[FargateTask] objects: %v", objects)

		// Temporary Rule Set created with only one rule
		tempRuleSet := customerprofileslcs.RuleSet{
			Name:                    currentActive.Name,
			Rules:                   []customerprofileslcs.Rule{rule},
			LatestVersion:           currentActive.LatestVersion,
			LastActivationTimestamp: currentActive.LastActivationTimestamp,
		}
		// For each object find matches
		// We have 1 Rule and many objects
		var lastConnectId string
		for _, object := range objects {
			if object["connect_id"].(string) == "" {
				tx.Warn("Invalid object found with no connect_id: %v", object)
				continue
			}
			connectID := object["connect_id"].(string)
			lastConnectId = connectID
			// We are removing generated fields to pass CP object parsing
			delete(object, "timestamp")
			delete(object, "overall_confidence_score")
			delete(object, "connect_id")
			delete(object, "profile_id")

			tx.Debug("[FargateTask] Processing Object: %v", object)
			objectStr, err := json.Marshal(object)
			if err != nil {
				tx.Error("Error marshalling object: %v", err)
				continue
			}
			_, err = cp.RunRuleBasedIdentityResolution(string(objectStr), OBJECT_TYPE_NAME, connectID, tempRuleSet)
			if err != nil {
				tx.Error("Error processing object: %v", object)
			}
		}

		if lastConnectId != "" {
			lastConnectUuid, err := uuid.Parse(lastConnectId)
			if err != nil {
				return errors.New("invalid last connect ID")
			}
			objects, err = cp.GetInteractionTable(DOMAIN_NAME, OBJECT_TYPE_NAME, partition, lastConnectUuid, pageSize)
			if err != nil {
				tx.Error("error getting %v object table from domain %v", OBJECT_TYPE_NAME, DOMAIN_NAME)
				return err
			}
		}
	}
	return nil
}

func initStorage() customerprofileslcs.ICustomerProfileLowCostConfig {
	return customerprofileslcs.InitLowCost(
		AWS_REGION,
		auroraClient,
		&configTableClient,
		kinesisClient,
		&cpWriterClient,
		METRICS_NAMESPACE,
		METRICS_SOLUTION_ID,
		METRICS_SOLUTION_VERSION,
		LOG_LEVEL,
		&customerprofileslcs.CustomerProfileInitOptions{MergeQueueClient: &mergeQueueClient},
		LCS_CACHE,
	)
}
