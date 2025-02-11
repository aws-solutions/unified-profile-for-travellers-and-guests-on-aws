// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"strings"
	"sync"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/aurora"
	core "tah/upt/source/tah-core/core"
	db "tah/upt/source/tah-core/db"
	glue "tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/kinesis"
	secrets "tah/upt/source/tah-core/secret"
	"tah/upt/source/tah-core/sqs"
	common "tah/upt/source/ucp-common/src/constant/admin"
	commonModel "tah/upt/source/ucp-common/src/model/admin"

	model "tah/upt/source/ucp-sync/src/business-logic/model"
	maintainGluePartitions "tah/upt/source/ucp-sync/src/business-logic/usecase/maintainGluePartitions"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Resources
var LAMBDA_ENV = os.Getenv("LAMBDA_ENV")
var LAMBDA_ACCOUNT_ID = os.Getenv("LAMBDA_ACCOUNT_ID")
var LAMBDA_REGION = os.Getenv("AWS_REGION")
var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
var METRICS_UUID = os.Getenv("METRICS_UUID")
var METRICS_NAMESPACE = "upt/sync"

var ATHENA_WORKGROUP = os.Getenv("ATHENA_WORKGROUP")
var ATHENA_TABLE = os.Getenv("ATHENA_TABLE")
var ATHENA_DB = os.Getenv("ATHENA_DB")

var HOTEL_BOOKING_JOB_NAME = os.Getenv("HOTEL_BOOKING_JOB_NAME_CUSTOMER")
var AIR_BOOKING_JOB_NAME = os.Getenv("AIR_BOOKING_JOB_NAME_CUSTOMER")
var GUEST_PROFILE_JOB_NAME = os.Getenv("GUEST_PROFILE_JOB_NAME_CUSTOMER")
var PAX_PROFILE_JOB_NAME = os.Getenv("PAX_PROFILE_JOB_NAME_CUSTOMER")
var CLICKSTREAM_JOB_NAME = os.Getenv("CLICKSTREAM_JOB_NAME_CUSTOMER")
var HOTEL_STAY_JOB_NAME = os.Getenv("HOTEL_STAY_JOB_NAME_CUSTOMER")
var CSI_JOB_NAME = os.Getenv("CSI_JOB_NAME_CUSTOMER")

var S3_HOTEL_BOOKING = os.Getenv("S3_HOTEL_BOOKING")
var S3_AIR_BOOKING = os.Getenv("S3_AIR_BOOKING")
var S3_GUEST_PROFILE = os.Getenv("S3_GUEST_PROFILE")
var S3_PAX_PROFILE = os.Getenv("S3_PAX_PROFILE")
var S3_CLICKSTREAM = os.Getenv("S3_CLICKSTREAM")
var S3_STAY_REVENUE = os.Getenv("S3_STAY_REVENUE")
var S3_CSI = os.Getenv("S3_CSI")
var CONNECT_PROFILE_SOURCE_BUCKET = os.Getenv("CONNECT_PROFILE_SOURCE_BUCKET")

var DYNAMO_TABLE = os.Getenv("DYNAMO_TABLE")
var DYNAMO_PK = os.Getenv("DYNAMO_PK")
var DYNAMO_SK = os.Getenv("DYNAMO_SK")

var ORIGIN_DATE = os.Getenv("PARTITION_START_DATE")
var SKIP_JOB_RUN = os.Getenv("SKIP_JOB_RUN")

var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

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
var LCS_CACHE = customerprofiles.InitLcsCache()

// Init clients
var configDb = db.Init(DYNAMO_TABLE, DYNAMO_PK, DYNAMO_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var glueCfg = glue.InitWithRetries(LAMBDA_REGION, ATHENA_DB, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

// Init low-cost clients
var smClient = secrets.InitWithRegion(AURORA_DB_SECRET_ARN, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var user, pwd = getUserCredentials(smClient)
var auroraClient = aurora.InitPostgresql(AURORA_PROXY_ENDPOINT, AURORA_DB_NAME, user, pwd, 50, LOG_LEVEL)

var configTableClient = db.Init(
	STORAGE_CONFIG_TABLE_NAME,
	STORAGE_CONFIG_TABLE_PK,
	STORAGE_CONFIG_TABLE_SK,
	METRICS_SOLUTION_ID,
	METRICS_SOLUTION_VERSION,
)
var kinesisClient = kinesis.Init(CHANGE_PROC_KINESIS_STREAM_NAME, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, LOG_LEVEL)
var mergeQueue = sqs.InitWithQueueUrl(MERGE_QUEUE_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var cpWriterQueue = sqs.InitWithQueueUrl(CP_WRITER_QUEUE_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

var buckets = map[string]string{
	common.BIZ_OBJECT_HOTEL_BOOKING: S3_HOTEL_BOOKING,
	common.BIZ_OBJECT_AIR_BOOKING:   S3_AIR_BOOKING,
	common.BIZ_OBJECT_GUEST_PROFILE: S3_GUEST_PROFILE,
	common.BIZ_OBJECT_PAX_PROFILE:   S3_PAX_PROFILE,
	common.BIZ_OBJECT_CLICKSTREAM:   S3_CLICKSTREAM,
	common.BIZ_OBJECT_STAY_REVENUE:  S3_STAY_REVENUE,
	common.BIZ_OBJECT_CSI:           S3_CSI,
}
var jobs = map[string]string{
	common.BIZ_OBJECT_HOTEL_BOOKING: HOTEL_BOOKING_JOB_NAME,
	common.BIZ_OBJECT_AIR_BOOKING:   AIR_BOOKING_JOB_NAME,
	common.BIZ_OBJECT_GUEST_PROFILE: GUEST_PROFILE_JOB_NAME,
	common.BIZ_OBJECT_PAX_PROFILE:   PAX_PROFILE_JOB_NAME,
	common.BIZ_OBJECT_CLICKSTREAM:   CLICKSTREAM_JOB_NAME,
	common.BIZ_OBJECT_STAY_REVENUE:  HOTEL_STAY_JOB_NAME,
	common.BIZ_OBJECT_CSI:           CSI_JOB_NAME,
}

func buildConfig(req events.CloudWatchEvent) model.AWSConfig {
	tx := core.NewTransaction("ucp-sync", "", LOG_LEVEL)
	accpCfg := initStorage()
	configDb.SetTx(tx)
	accpCfg.SetTx(tx)
	uuid := METRICS_UUID
	if uuid == "" {
		//we prefer replacing the UUID with a random one to no UUID as it would fail the Glue Job
		tx.Warn("Warning: No UUID found, generating a new one")
		uuid = core.UUID()
	}
	cfg := model.AWSConfig{
		Tx:                   tx,
		Glue:                 glueCfg,
		Accp:                 accpCfg,
		Dynamo:               configDb,
		BizObjectBucketNames: buckets,
		JobNames:             jobs,
		Env: map[string]string{
			"LAMBDA_ENV":                    LAMBDA_ENV,
			"ORIGIN_DATE":                   ORIGIN_DATE,
			"CONNECT_PROFILE_SOURCE_BUCKET": CONNECT_PROFILE_SOURCE_BUCKET,
			"METRICS_UUID":                  uuid,
			"SKIP_JOB_RUN":                  SKIP_JOB_RUN,
			"METHOD_OF_INVOCATION":          req.DetailType,
		},
	}
	if req.DetailType == common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER {
		cfg.Domains = parseDomain(req.Resources)
		cfg.RequestedJob = parseJob(req.Resources)
	}
	return cfg
}

func HandleRequest(ctx context.Context, req events.CloudWatchEvent) (model.ResponseWrapper, error) {
	config := buildConfig(req)
	return HandleRequestWithParams(ctx, req, config)
}

func parseDomain(names []string) []commonModel.Domain {
	domains := []commonModel.Domain{}
	for _, resourceName := range names {
		if strings.HasPrefix(resourceName, "domain:") {
			segs := strings.Split(resourceName, ":")
			domains = append(domains, commonModel.Domain{Name: segs[1]})
		}
	}
	return domains
}

func parseJob(names []string) string {
	for _, resourceName := range names {
		if strings.HasPrefix(resourceName, "job:") {
			segs := strings.Split(resourceName, ":")
			return segs[1]
		}
	}
	return ""
}

func HandleRequestWithParams(ctx context.Context, req events.CloudWatchEvent, cfg model.AWSConfig) (model.ResponseWrapper, error) {
	cfg.Tx.Info("Received Request %+v with context %+v", req, ctx)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		cfg.Tx.Info("Starting use case %+v", "maintainGluePartitions")
		allErrs := []error{}
		perr, jerr := maintainGluePartitions.Run(cfg)
		allErrs = append(allErrs, perr...)
		allErrs = append(allErrs, jerr...)
		if len(allErrs) > 0 {
			cfg.Tx.Error("Multiple errors while running use case %+v: %v", "maintainGluePartitions", allErrs)
		}
		wg.Done()
	}()
	wg.Wait()
	return model.ResponseWrapper{}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

func initStorage() customerprofiles.ICustomerProfileLowCostConfig {
	var customCfg customerprofiles.ICustomerProfileLowCostConfig
	options := customerprofiles.CustomerProfileInitOptions{
		MergeQueueClient: &mergeQueue,
	}
	customCfg = customerprofiles.InitLowCost(
		LAMBDA_REGION,
		auroraClient,
		&configTableClient,
		kinesisClient,
		&cpWriterQueue,
		METRICS_NAMESPACE,
		METRICS_SOLUTION_ID,
		METRICS_SOLUTION_VERSION,
		LOG_LEVEL,
		&options,
		LCS_CACHE,
	)

	return customCfg
}

func getUserCredentials(sm secrets.Config) (string, string) {
	user := sm.Get("username")
	pwd := sm.Get("password")
	return user, pwd
}
