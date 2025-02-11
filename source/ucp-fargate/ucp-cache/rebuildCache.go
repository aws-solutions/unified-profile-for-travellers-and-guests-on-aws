// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"log"
	"os"
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
var METRICS_NAMESPACE = "upt/rebuild-cache"

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
var cpWriterQueueClient = sqs.InitWithQueueUrl(CP_WRITER_QUEUE_URL, AWS_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

func getUserCredentials(sm secrets.Config) (string, string) {
	user := sm.Get("username")
	pwd := sm.Get("password")
	return user, pwd
}

func CacheMain() {
	log.Print("[CacheBuilder] Starting Rebuild Cache Fargate Task")
	cp := initStorage()
	HandleWithServices(cp)
}

func HandleWithServices(cp customerprofileslcs.ICustomerProfileLowCostConfig) {
	// Task Level Variables
	var DOMAIN_NAME = os.Getenv("DOMAIN_NAME")
	var CACHE_TYPE = os.Getenv("CACHE_TYPE")
	var PARTITION_LOWER_BOUND = os.Getenv("PARTITION_LOWER_BOUND")
	var PARTITION_UPPER_BOUND = os.Getenv("PARTITION_UPPER_BOUND")

	cp.SetDomain(DOMAIN_NAME)

	// Confirm partitions are uuids
	lowerBound, valid := isValidUUID(PARTITION_LOWER_BOUND)
	if !valid {
		log.Fatalf("PARTITION_LOWER_BOUND is not a valid uuid")
	}
	upperBound, valid := isValidUUID(PARTITION_UPPER_BOUND)
	if !valid {
		log.Fatalf("PARTITION_UPPER_BOUND is not a valid uuid")
	}

	// Get Profiles
	partition := customerprofileslcs.UuidPartition{LowerBound: lowerBound, UpperBound: upperBound}
	connectIds, err := cp.GetProfilePartition(partition)
	if err != nil {
		log.Fatalf("Error retrieving partition: %v", err)
	}
	for _, connectId := range connectIds {
		var cacheMode customerprofileslcs.CacheMode
		if CACHE_TYPE == "DYNAMO" {
			cacheMode = customerprofileslcs.DYNAMO_MODE
		}
		if CACHE_TYPE == "CUSTOMER_PROFILES" {
			cacheMode = customerprofileslcs.CUSTOMER_PROFILES_MODE
		}
		err := cp.CacheProfile(connectId, cacheMode)
		if err != nil {
			log.Fatalf("Error updating %v cache with connectId %v: %v", CACHE_TYPE, connectId, err)
		}
	}
}

func initStorage() customerprofileslcs.ICustomerProfileLowCostConfig {
	return customerprofileslcs.InitLowCost(
		AWS_REGION,
		auroraClient,
		&configTableClient,
		kinesisClient,
		&cpWriterQueueClient,
		METRICS_NAMESPACE,
		METRICS_SOLUTION_ID,
		METRICS_SOLUTION_VERSION,
		LOG_LEVEL,
		&customerprofileslcs.CustomerProfileInitOptions{MergeQueueClient: &mergeQueueClient},
		LCS_CACHE,
	)
}

func isValidUUID(u string) (uuid.UUID, bool) {
	uuid, err := uuid.Parse(u)
	return uuid, err == nil
}
