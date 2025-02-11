// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"os"
	"time"

	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/aurora"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/ecs"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/kinesis"
	tahCoreLambda "tah/upt/source/tah-core/lambda"
	secrets "tah/upt/source/tah-core/secret"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/tah-core/ssm"
	"tah/upt/source/ucp-async/src/business-logic/usecase"
	model "tah/upt/source/ucp-common/src/model/async"

	"github.com/aws/aws-lambda-go/lambda"
)

var TIMESTAMP_FORMAT = "2006-01-02 15:04:05.0"

// Solution Info
var LAMBDA_ENV = os.Getenv("LAMBDA_ENV")
var LAMBDA_REGION = os.Getenv("LAMBDA_REGION")
var ASYNC_LAMBDA_NAME = os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
var SEND_ANONYMIZED_DATA = os.Getenv("SEND_ANONYMIZED_DATA")
var METRICS_UUID = os.Getenv("METRICS_UUID")
var RETRY_LAMBDA_NAME = os.Getenv("RETRY_LAMBDA_NAME")
var S3_EXPORT_BUCKET = os.Getenv("S3_EXPORT_BUCKET")
var GLUE_EXPORT_TABLE_NAME = os.Getenv("GLUE_EXPORT_TABLE_NAME")
var TRAVELER_S3_ROOT_PATH = os.Getenv("TRAVELER_S3_ROOT_PATH")
var ATHENA_WORKGROUP_NAME = os.Getenv("ATHENA_WORKGROUP_NAME")
var ATHENA_OUTPUT_BUCKET = os.Getenv("ATHENA_OUTPUT_BUCKET")
var GET_S3_PATHS_PREPARED_STATEMENT_NAME = os.Getenv("GET_S3_PATHS_PREPARED_STATEMENT_NAME")
var GET_S3PATH_MAPPING_PREPARED_STATEMENT_NAME = os.Getenv("GET_S3PATH_MAPPING_PREPARED_STATEMENT_NAME")
var CP_EXPORT_STREAM = os.Getenv("CP_EXPORT_STREAM")

// Table Info
var CONFIG_TABLE = os.Getenv("CONFIG_TABLE_NAME")
var CONFIG_PK = os.Getenv("CONFIG_TABLE_PK")
var CONFIG_SK = os.Getenv("CONFIG_TABLE_SK")

var MATCH_TABLE_NAME = os.Getenv("MATCH_TABLE_NAME")
var MATCH_TABLE_PK = os.Getenv("MATCH_TABLE_PK")
var MATCH_TABLE_SK = os.Getenv("MATCH_TABLE_SK")

var ERROR_TABLE_NAME = os.Getenv("ERROR_TABLE_NAME")
var ERROR_TABLE_PK = os.Getenv("ERROR_TABLE_PK")
var ERROR_TABLE_SK = os.Getenv("ERROR_TABLE_SK")

var PRIVACY_RESULTS_TABLE_NAME = os.Getenv("PRIVACY_RESULTS_TABLE_NAME")
var PRIVACY_RESULTS_TABLE_PK = os.Getenv("PRIVACY_RESULTS_TABLE_PK")
var PRIVACY_RESULTS_TABLE_SK = os.Getenv("PRIVACY_RESULTS_TABLE_SK")

var COGNITO_USER_POOL_ID = os.Getenv("COGNITO_USER_POOL_ID")
var GLUE_DB = os.Getenv("GLUE_DB")

var PORTAL_CONFIG_TABLE_NAME = os.Getenv("PORTAL_CONFIG_TABLE_NAME")
var PORTAL_CONFIG_TABLE_PK = os.Getenv("PORTAL_CONFIG_TABLE_PK")
var PORTAL_CONFIG_TABLE_SK = os.Getenv("PORTAL_CONFIG_TABLE_SK")
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
var LCS_CACHE = customerprofiles.InitLcsCache()

// Config Options
var USE_PERMISSION_SYSTEM = os.Getenv("USE_PERMISSION_SYSTEM")
var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

// Init low-cost clients (optimally, this is conditional for low-cost storage)
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

var S3_EXCISE_QUEUE_URL = os.Getenv("S3_EXCISE_QUEUE_URL")
var GDPR_PURGE_LOG_GROUP_NAME = os.Getenv("GDPR_PURGE_LOG_GROUP_NAME")

// Services
var configDb = db.Init(CONFIG_TABLE, CONFIG_PK, CONFIG_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var matchDb = db.Init(MATCH_TABLE_NAME, MATCH_TABLE_PK, MATCH_TABLE_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var errorDb = db.Init(ERROR_TABLE_NAME, ERROR_TABLE_PK, ERROR_TABLE_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

var portalConfigDB = db.Init(
	PORTAL_CONFIG_TABLE_NAME,
	PORTAL_CONFIG_TABLE_PK,
	PORTAL_CONFIG_TABLE_SK,
	METRICS_SOLUTION_ID,
	METRICS_SOLUTION_VERSION,
)

var privacyResultsDb = db.Init(
	PRIVACY_RESULTS_TABLE_NAME,
	PRIVACY_RESULTS_TABLE_PK,
	PRIVACY_RESULTS_TABLE_SK,
	METRICS_SOLUTION_ID,
	METRICS_SOLUTION_VERSION,
)
var solutionsConfig = awssolutions.Init(METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, METRICS_UUID, SEND_ANONYMIZED_DATA, LOG_LEVEL)
var cognitoConfig = cognito.Init(COGNITO_USER_POOL_ID, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, LOG_LEVEL)
var glueConfig = glue.Init(LAMBDA_REGION, GLUE_DB, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var retryLambdaConfig = tahCoreLambda.Init(RETRY_LAMBDA_NAME, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var asyncLambdaConfig = tahCoreLambda.Init(ASYNC_LAMBDA_NAME, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var s3ExciseSqsConfig = sqs.InitWithQueueUrl(S3_EXCISE_QUEUE_URL, LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

var ssmCfg, _ = ssm.InitSsm(context.Background(), LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var ecsCfg, _ = ecs.InitEcs(context.Background(), LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

// storage modes
var CUSTOMER_PROFILE_STORAGE_MODE = os.Getenv("CUSTOMER_PROFILE_STORAGE_MODE")
var DYNAMO_STORAGE_MODE = os.Getenv("DYNAMO_STORAGE_MODE")

var SSM_PARAM_NAMESPACE = os.Getenv("SSM_PARAM_NAMESPACE")

func HandleRequest(ctx context.Context, raw string) {
	payload, err := model.Parse(raw) // parse base64 encoded payload
	if err != nil {
		tx := core.NewTransaction("async", "", LOG_LEVEL)
		tx.Error("Error while parsing payload: %v", err)
	}
	tx := core.NewTransaction("async", payload.TransactionID, LOG_LEVEL)
	tx.Debug("Async Lambda invoked with payload %+v", payload)

	accpConfig := initStorage()
	accpConfig.SetTx(tx)
	configDb.SetTx(tx)
	matchDb.SetTx(tx)
	errorDb.SetTx(tx)
	portalConfigDB.SetTx(tx)
	privacyResultsDb.SetTx(tx)

	services := model.Services{
		AccpConfig:        accpConfig,
		ConfigDB:          configDb,
		MatchDB:           matchDb,
		ErrorDB:           errorDb,
		PortalConfigDB:    portalConfigDB,
		PrivacyResultsDB:  privacyResultsDb,
		SolutionsConfig:   solutionsConfig,
		CognitoConfig:     cognitoConfig,
		GlueConfig:        glueConfig,
		RetryLambdaConfig: retryLambdaConfig,
		S3ExciseSqsConfig: s3ExciseSqsConfig,
		SsmConfig:         ssmCfg,
		EcsConfig:         ecsCfg,
		AsyncLambdaConfig: asyncLambdaConfig,
		Env: map[string]string{
			"S3_EXPORT_BUCKET":                           S3_EXPORT_BUCKET,
			"GLUE_EXPORT_TABLE_NAME":                     GLUE_EXPORT_TABLE_NAME,
			"TRAVELER_S3_ROOT_PATH":                      TRAVELER_S3_ROOT_PATH,
			"GLUE_DB":                                    glueConfig.DbName,
			"ATHENA_WORKGROUP_NAME":                      ATHENA_WORKGROUP_NAME,
			"REGION":                                     LAMBDA_REGION,
			"ATHENA_OUTPUT_BUCKET":                       "s3://" + ATHENA_OUTPUT_BUCKET,
			"METRICS_SOLUTION_ID":                        METRICS_SOLUTION_ID,
			"METRICS_SOLUTION_VERSION":                   METRICS_SOLUTION_VERSION,
			"GET_S3_PATHS_PREPARED_STATEMENT_NAME":       GET_S3_PATHS_PREPARED_STATEMENT_NAME,
			"GET_S3PATH_MAPPING_PREPARED_STATEMENT_NAME": GET_S3PATH_MAPPING_PREPARED_STATEMENT_NAME,
			"GDPR_PURGE_LOG_GROUP_NAME":                  GDPR_PURGE_LOG_GROUP_NAME,
			"CUSTOMER_PROFILE_STORAGE_MODE":              CUSTOMER_PROFILE_STORAGE_MODE,
			"DYNAMO_STORAGE_MODE":                        DYNAMO_STORAGE_MODE,
			"SSM_PARAM_NAMESPACE":                        SSM_PARAM_NAMESPACE,
			"CP_EXPORT_STREAM":                           CP_EXPORT_STREAM,
			"USE_PERMISSION_SYSTEM":                      USE_PERMISSION_SYSTEM,
		},
	}

	HandleRequestWithServices(ctx, payload, services, tx)
}

func HandleRequestWithServices(ctx context.Context, payload model.AsyncInvokePayload, services model.Services, tx core.Transaction) {
	now := time.Now()
	usecases := usecase.GetUsecases()
	uc := usecases[payload.Usecase]

	// Check for valid usecase
	if uc == nil {
		tx.Error("[%s] Usecase not found", payload.Usecase)
		setStatus(payload, services, tx, now, model.EVENT_STATUS_FAILED)
		return
	}

	// Update event status
	setStatus(payload, services, tx, now, model.EVENT_STATUS_RUNNING)

	// Execute usecase
	err := uc.Execute(payload, services, tx)
	if err != nil {
		tx.Error("[%s] Error while executing usecase: %s", payload.Usecase, err)
		setStatus(payload, services, tx, now, model.EVENT_STATUS_FAILED)
		return
	}

	tx.Info("[%s] Usecase executed successfully", payload.Usecase)
	setStatus(payload, services, tx, now, model.EVENT_STATUS_SUCCESS)
}

func main() {
	lambda.Start(HandleRequest)
}

func setStatus(payload model.AsyncInvokePayload, services model.Services, tx core.Transaction, now time.Time, status string) {
	// Update status
	event := model.AsyncEvent{
		EventID:     payload.EventID,
		Usecase:     payload.Usecase,
		Status:      status,
		LastUpdated: time.Now(),
	}
	_, err := services.ConfigDB.Save(event)
	if err != nil {
		tx.Error("Error while updating status: %s", err)
	}

	// Send metrics when usecase is successful or failed
	if status == model.EVENT_STATUS_SUCCESS || status == model.EVENT_STATUS_FAILED {
		_, err = services.SolutionsConfig.SendMetrics(map[string]interface{}{
			"service":  "ucp-async",
			"duration": time.Since(now).Milliseconds(),
			"status":   status,
		})
		if err != nil {
			tx.Warn("Error while sending metrics to AWS Solutions: %s", err)
		}
	}
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
	log.Printf("Retrieving DB credentials from Secrets Manager: %v", sm.SecretArn)
	user := sm.Get("username")
	pwd := sm.Get("password")
	return user, pwd
}
