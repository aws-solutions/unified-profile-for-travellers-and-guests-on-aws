// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	customerprofiles "tah/upt/source/storage"
	appflow "tah/upt/source/tah-core/appflow"
	appregistry "tah/upt/source/tah-core/appregistry"
	"tah/upt/source/tah-core/aurora"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/bedrock"
	"tah/upt/source/tah-core/cognito"
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/datasync"
	db "tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/ecs"
	glue "tah/upt/source/tah-core/glue"
	iam "tah/upt/source/tah-core/iam"
	"tah/upt/source/tah-core/kinesis"
	secrets "tah/upt/source/tah-core/secret"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/tah-core/ssm"

	tahCoreLambda "tah/upt/source/tah-core/lambda"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"

	"tah/upt/source/ucp-backend/src/business-logic/usecase/admin"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/privacy"
	registry "tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	traveller "tah/upt/source/ucp-backend/src/business-logic/usecase/traveller"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Solution info
var METRICS_SOLUTION_ID = os.Getenv("METRICS_SOLUTION_ID")
var METRICS_SOLUTION_VERSION = os.Getenv("METRICS_SOLUTION_VERSION")
var METRICS_UUID = os.Getenv("METRICS_UUID")
var SEND_ANONYMIZED_DATA = os.Getenv("SEND_ANONYMIZED_DATA")
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

// Resources
var LAMBDA_ENV = os.Getenv("LAMBDA_ENV")
var LAMBDA_ACCOUNT_ID = os.Getenv("LAMBDA_ACCOUNT_ID")
var LAMBDA_REGION = os.Getenv("AWS_REGION")
var NAMESPACE_NAME = "cloudRackServiceDiscoveryNamespace" + LAMBDA_ENV
var ATHENA_WORKGROUP = os.Getenv("ATHENA_WORKGROUP")
var ATHENA_DB = os.Getenv("ATHENA_DB")
var CONNECTOR_CRAWLER_QUEUE = os.Getenv("CONNECTOR_CRAWLER_QUEUE")
var CONNECTOR_CRAWLER_DLQ = os.Getenv("CONNECTOR_CRAWLER_DLQ")
var GLUE_DB = os.Getenv("GLUE_DB")
var DATALAKE_ADMIN_ROLE_ARN = os.Getenv("DATALAKE_ADMIN_ROLE_ARN")
var COGNITO_USER_POOL_ID = os.Getenv("COGNITO_USER_POOL_ID")

// S3 buckets
var CONNECT_PROFILE_SOURCE_BUCKET = os.Getenv("CONNECT_PROFILE_SOURCE_BUCKET")
var S3_HOTEL_BOOKING = os.Getenv("S3_HOTEL_BOOKING")
var S3_AIR_BOOKING = os.Getenv("S3_AIR_BOOKING")
var S3_GUEST_PROFILE = os.Getenv("S3_GUEST_PROFILE")
var S3_PAX_PROFILE = os.Getenv("S3_PAX_PROFILE")
var S3_STAY_REVENUE = os.Getenv("S3_STAY_REVENUE")
var S3_CLICKSTREAM = os.Getenv("S3_CLICKSTREAM")
var S3_CSI = os.Getenv("S3_CSI")

// Job Names
var HOTEL_BOOKING_JOB_NAME_CUSTOMER = os.Getenv("HOTEL_BOOKING_JOB_NAME_CUSTOMER")
var AIR_BOOKING_JOB_NAME_CUSTOMER = os.Getenv("AIR_BOOKING_JOB_NAME_CUSTOMER")
var GUEST_PROFILE_JOB_NAME_CUSTOMER = os.Getenv("GUEST_PROFILE_JOB_NAME_CUSTOMER")
var PAX_PROFILE_JOB_NAME_CUSTOMER = os.Getenv("PAX_PROFILE_JOB_NAME_CUSTOMER")
var CLICKSTREAM_JOB_NAME_CUSTOMER = os.Getenv("CLICKSTREAM_JOB_NAME_CUSTOMER")
var HOTEL_STAY_JOB_NAME_CUSTOMER = os.Getenv("HOTEL_STAY_JOB_NAME_CUSTOMER")
var CSI_JOB_NAME_CUSTOMER = os.Getenv("CSI_JOB_NAME_CUSTOMER")
var RUN_ALL_JOBS = "run-all-jobs"

var ACCP_DOMAIN_DLQ = os.Getenv("ACCP_DOMAIN_DLQ")
var ACCP_DESTINATION_STREAM = os.Getenv("ACCP_DESTINATION_STREAM")
var SYNC_LAMBDA_NAME = os.Getenv("SYNC_LAMBDA_NAME")
var ASYNC_LAMBDA_NAME = os.Getenv("ASYNC_LAMBDA_NAME")
var RETRY_LAMBDA_NAME = os.Getenv("RETRY_LAMBDA_NAME")

var KMS_KEY_PROFILE_DOMAIN = os.Getenv("KMS_KEY_PROFILE_DOMAIN")
var UCP_CONNECT_DOMAIN = ""

var CONFIG_TABLE_NAME = os.Getenv("CONFIG_TABLE_NAME")
var CONFIG_TABLE_NAME_TEST = os.Getenv("CONFIG_TABLE_NAME_TEST")
var CONFIG_TABLE_PK = os.Getenv("CONFIG_TABLE_PK")
var CONFIG_TABLE_SK = os.Getenv("CONFIG_TABLE_SK")

var ERROR_TABLE_NAME = os.Getenv("ERROR_TABLE_NAME")
var ERROR_TABLE_PK = os.Getenv("ERROR_TABLE_PK")
var ERROR_TABLE_SK = os.Getenv("ERROR_TABLE_SK")

var DYNAMO_TABLE_MATCH = os.Getenv("DYNAMO_TABLE_MATCH")
var DYNAMO_TABLE_MATCH_PK = os.Getenv("DYNAMO_TABLE_MATCH_PK")
var DYNAMO_TABLE_MATCH_SK = os.Getenv("DYNAMO_TABLE_MATCH_SK")

var PORTAL_CONFIG_TABLE_NAME = os.Getenv("PORTAL_CONFIG_TABLE_NAME")
var PORTAL_CONFIG_TABLE_PK = os.Getenv("PORTAL_CONFIG_TABLE_PK")
var PORTAL_CONFIG_TABLE_SK = os.Getenv("PORTAL_CONFIG_TABLE_SK")

var PRIVACY_RESULTS_TABLE_NAME = os.Getenv("PRIVACY_RESULTS_TABLE_NAME")
var PRIVACY_RESULTS_TABLE_PK = os.Getenv("PRIVACY_RESULTS_TABLE_PK")
var PRIVACY_RESULTS_TABLE_SK = os.Getenv("PRIVACY_RESULTS_TABLE_SK")

var SECRET_ARN = os.Getenv("SECRET_ARN")

// Matches
var MATCH_BUCKET_NAME = os.Getenv("MATCH_BUCKET_NAME")

// Industry Connector DataSync
var INDUSTRY_CONNECTOR_BUCKET_NAME = os.Getenv("INDUSTRY_CONNECTOR_BUCKET_NAME")
var DATASYNC_ROLE_ARN = os.Getenv("DATASYNC_ROLE_ARN")
var DATASYNC_LOG_GROUP_ARN = os.Getenv("DATASYNC_LOG_GROUP_ARN")
var IS_PLACEHOLDER_CONNECTOR_BUCKET = os.Getenv("IS_PLACEHOLDER_CONNECTOR_BUCKET")
var LOG_LEVEL = core.LogLevelFromString(os.Getenv("LOG_LEVEL"))

var SSM_PARAM_NAMESPACE = os.Getenv("SSM_PARAM_NAMESPACE")

// Config Options
var USE_PERMISSION_SYSTEM = os.Getenv("USE_PERMISSION_SYSTEM")

// Use cases
var FN_RETREIVE_UCP_PROFILE = "retreive_ucp_profile"
var FN_DELETE_UCP_PROFILE = "delete_ucp_profile"
var FN_SEARCH_UCP_PROFILES = "search_ucp_profiles"
var FN_RETREIVE_UCP_CONFIG = "retreive_ucp_config"
var FN_LIST_UCP_DOMAINS = "list_ucp_domains"
var FN_CREATE_UCP_DOMAIN = "create_ucp_domain"
var FN_DELETE_UCP_DOMAIN = "delete_ucp_domain"
var FN_MERGE_UCP_PROFILE = "merge_ucp_profile"
var FN_LIST_CONNECTORS = "list_connectors"
var FN_GET_DATA_VALIDATION_STATUS = "get_data_validation_status"
var FN_LINK_INDUSTRY_CONNECTOR = "link_industry_connector"
var FN_CREATE_CONNECTOR_CRAWLER = "create_connector_crawler"
var FN_LIST_UCP_INGESTION_ERROR = "list_ucp_ingestion_errors"
var CUSTOMER_PROFILE_DOMAIN_HEADER = "customer-profiles-domain"

var appregistryClient = appregistry.Init(LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var iamClient = iam.Init(METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var glueClient = glue.Init(LAMBDA_REGION, GLUE_DB, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var cognitoClient = cognito.Init(COGNITO_USER_POOL_ID, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, LOG_LEVEL)
var configDB = db.Init(CONFIG_TABLE_NAME, CONFIG_TABLE_PK, CONFIG_TABLE_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var errorDB = db.Init(ERROR_TABLE_NAME, ERROR_TABLE_PK, ERROR_TABLE_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

var portalConfigDB = db.Init(
	PORTAL_CONFIG_TABLE_NAME,
	PORTAL_CONFIG_TABLE_PK,
	PORTAL_CONFIG_TABLE_SK,
	METRICS_SOLUTION_ID,
	METRICS_SOLUTION_VERSION,
)
var matchDB = db.Init(DYNAMO_TABLE_MATCH, DYNAMO_TABLE_MATCH_PK, DYNAMO_TABLE_MATCH_SK, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

var privacyDB = db.Init(
	PRIVACY_RESULTS_TABLE_NAME,
	PRIVACY_RESULTS_TABLE_PK,
	PRIVACY_RESULTS_TABLE_SK,
	METRICS_SOLUTION_ID,
	METRICS_SOLUTION_VERSION,
)
var appFlowSvc = appflow.Init(METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var lambdaSvc = tahCoreLambda.Init(SYNC_LAMBDA_NAME, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var retryLambdaSvc = tahCoreLambda.Init(RETRY_LAMBDA_NAME, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var asyncLambda = tahCoreLambda.Init(ASYNC_LAMBDA_NAME, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var datasyncClient = datasync.InitRegion(LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, LOG_LEVEL)
var solutionUtils = awssolutions.Init(METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION, METRICS_UUID, SEND_ANONYMIZED_DATA, LOG_LEVEL)
var bedrockClient = bedrock.Init(LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

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

var ecsCfg, _ = ecs.InitEcs(context.Background(), LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
var ssmCfg, _ = ssm.InitSsm(context.Background(), LAMBDA_REGION, METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)

const PROFILE_PATH string = "/ucp/profile"
const PROFILE_PATH_ID string = PROFILE_PATH + "/{id}"
const PROFILE_SUMMARY_PATH_ID string = PROFILE_PATH + "/summary/{id}"
const MERGE_PATH string = "/ucp/merge"
const UNMERGE_PATH string = "/ucp/unmerge"
const API_PATH string = "/ucp/admin"
const API_PATH_ADMIN_ID string = API_PATH + "/{id}"
const CONNECTOR_PATH string = "/ucp/connector"
const ERROR_PATH string = "/ucp/error"
const ERROR_PATH_ID string = ERROR_PATH + "/{id}"
const JOBS_PATH string = "/ucp/jobs"
const FLOWS_PATH string = "/ucp/flows"
const PORTAL_CONFIG_PATH string = "/ucp/portalConfig"
const ASYNC_PATH string = "/ucp/async"
const RULE_SET_PATH string = "/ucp/ruleSet"
const RULE_SET_PATH_ACTIVATE string = RULE_SET_PATH + "/activate"
const RULE_SET_PATH_CACHE string = "/ucp/ruleSetCache"
const RULE_SET_PATH_ACTIVATE_CACHE string = RULE_SET_PATH_CACHE + "/activate"
const REBUILD_CACHE_PATH string = "/ucp/cache"
const PRIVACY_PATH string = "/ucp/privacy"
const PRIVACY_PATH_ID string = PRIVACY_PATH + "/{id}"
const PRIVACY_PATH_PURGE = PRIVACY_PATH + "/purge"
const INTERACTION_HISTORY_PATH = "/ucp/interactionHistory"
const INTERACTION_HISTORY_PATH_ID = INTERACTION_HISTORY_PATH + "/{id}"
const PROMPT_CONFIG_PATH = "/ucp/promptConfig"

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := ValidateApiGatewayRequest(req)
	if err != nil {
		return buildResponseError(err), nil
	}
	tx := core.NewTransaction("ucp-backend", "", LOG_LEVEL)
	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")
	tx.AddLogObfuscationPattern("authorization:", "eyJra\\S*", " ")
	tx.AddLogObfuscationPattern("authorization:", "\\[eyJra\\S*", " ")

	tx.Info("Received Request %s %s with context %+v", req.HTTPMethod, req.Path, ctx)

	configDB.SetTx(tx)
	errorDB.SetTx(tx)
	portalConfigDB.SetTx(tx)
	privacyDB.SetTx(tx)
	accpDomainName := req.Headers[CUSTOMER_PROFILE_DOMAIN_HEADER]
	profiles := initStorage(accpDomainName)
	profiles.SetTx(tx)

	var reg = registry.NewRegistry(LAMBDA_REGION, LOG_LEVEL, registry.ServiceHandlers{
		AppRegistry:    &appregistryClient,
		Iam:            &iamClient,
		Glue:           &glueClient,
		Accp:           profiles,
		ConfigDB:       &configDB,
		ErrorDB:        &errorDB,
		MatchDB:        &matchDB,
		PrivacyDB:      &privacyDB,
		Cognito:        cognitoClient,
		PortalConfigDB: &portalConfigDB,
		AppFlow:        &appFlowSvc,
		SyncLambda:     lambdaSvc,
		AsyncLambda:    asyncLambda,
		RetryLambda:    retryLambdaSvc,
		DataSync:       datasyncClient,
		SolutionUtils:  solutionUtils,
		Bedrock:        &bedrockClient,
		SsmCfg:         ssmCfg,
		EcsCfg:         ecsCfg,
	})
	reg.SetRegion(LAMBDA_REGION)
	reg.SetTx(&tx)
	//Setting environment variables to the registry (this allows to pass CloudFormation created resources)
	reg.AddEnv("LAMBDA_ENV", LAMBDA_ENV)
	reg.AddEnv("KMS_KEY_PROFILE_DOMAIN", KMS_KEY_PROFILE_DOMAIN)
	reg.AddEnv("CONNECT_PROFILE_SOURCE_BUCKET", CONNECT_PROFILE_SOURCE_BUCKET)
	reg.AddEnv("CONNECTOR_CRAWLER_QUEUE", CONNECTOR_CRAWLER_QUEUE)
	reg.AddEnv("CONNECTOR_CRAWLER_DLQ", CONNECTOR_CRAWLER_DLQ)
	reg.AddEnv("AWS_ACCOUNT_ID", LAMBDA_ACCOUNT_ID)
	reg.AddEnv("DATALAKE_ADMIN_ROLE_ARN", DATALAKE_ADMIN_ROLE_ARN)

	reg.AddEnv("S3_HOTEL_BOOKING", S3_HOTEL_BOOKING)
	reg.AddEnv("S3_AIR_BOOKING", S3_AIR_BOOKING)
	reg.AddEnv("S3_GUEST_PROFILE", S3_GUEST_PROFILE)
	reg.AddEnv("S3_PAX_PROFILE", S3_PAX_PROFILE)
	reg.AddEnv("S3_STAY_REVENUE", S3_STAY_REVENUE)
	reg.AddEnv("S3_CLICKSTREAM", S3_CLICKSTREAM)
	reg.AddEnv("S3_CSI", S3_CSI)

	reg.AddEnv("HOTEL_BOOKING_JOB_NAME_CUSTOMER", HOTEL_BOOKING_JOB_NAME_CUSTOMER)
	reg.AddEnv("AIR_BOOKING_JOB_NAME_CUSTOMER", AIR_BOOKING_JOB_NAME_CUSTOMER)
	reg.AddEnv("GUEST_PROFILE_JOB_NAME_CUSTOMER", GUEST_PROFILE_JOB_NAME_CUSTOMER)
	reg.AddEnv("PAX_PROFILE_JOB_NAME_CUSTOMER", PAX_PROFILE_JOB_NAME_CUSTOMER)
	reg.AddEnv("CLICKSTREAM_JOB_NAME_CUSTOMER", CLICKSTREAM_JOB_NAME_CUSTOMER)
	reg.AddEnv("HOTEL_STAY_JOB_NAME_CUSTOMER", HOTEL_STAY_JOB_NAME_CUSTOMER)
	reg.AddEnv("CSI_JOB_NAME_CUSTOMER", CSI_JOB_NAME_CUSTOMER)
	reg.AddEnv("RUN_ALL_JOBS", RUN_ALL_JOBS)

	reg.AddEnv("MATCH_BUCKET_NAME", MATCH_BUCKET_NAME)

	reg.AddEnv(constants.ACCP_DOMAIN_NAME_ENV_VAR, accpDomainName)
	reg.AddEnv("ACCP_DOMAIN_DLQ", ACCP_DOMAIN_DLQ)
	reg.AddEnv("ACCP_DESTINATION_STREAM", ACCP_DESTINATION_STREAM)

	reg.AddEnv("CONFIG_TABLE_NAME", CONFIG_TABLE_NAME)
	reg.AddEnv("CONFIG_TABLE_NAME_TEST", CONFIG_TABLE_NAME_TEST)
	reg.AddEnv("CONFIG_TABLE_PK", CONFIG_TABLE_PK)
	reg.AddEnv("CONFIG_TABLE_SK", CONFIG_TABLE_SK)

	reg.AddEnv("PRIVACY_RESULTS_TABLE_NAME", PRIVACY_RESULTS_TABLE_NAME)
	reg.AddEnv("PRIVACY_RESULTS_TABLE_PK", PRIVACY_RESULTS_TABLE_PK)
	reg.AddEnv("PRIVACY_RESULTS_TABLE_SK", PRIVACY_RESULTS_TABLE_SK)

	reg.AddEnv("INDUSTRY_CONNECTOR_BUCKET_NAME", INDUSTRY_CONNECTOR_BUCKET_NAME)
	reg.AddEnv("DATASYNC_ROLE_ARN", DATASYNC_ROLE_ARN)
	reg.AddEnv("DATASYNC_LOG_GROUP_ARN", DATASYNC_LOG_GROUP_ARN)
	reg.AddEnv("IS_PLACEHOLDER_CONNECTOR_BUCKET", IS_PLACEHOLDER_CONNECTOR_BUCKET)

	reg.AddEnv("SSM_PARAM_NAMESPACE", SSM_PARAM_NAMESPACE)

	reg.AddEnv("USE_PERMISSION_SYSTEM", USE_PERMISSION_SYSTEM)

	reg.Register("GET", PROFILE_PATH_ID, traveller.NewRetreiveProfile())
	reg.Register("DELETE", PROFILE_PATH_ID, traveller.NewDeleteProfile())
	reg.Register("GET", PROFILE_PATH, traveller.NewSearchProfile())
	reg.Register("GET", PROFILE_SUMMARY_PATH_ID, traveller.NewGetProfileSummary())
	reg.Register("POST", MERGE_PATH, traveller.NewMergeProfile())
	reg.Register("POST", UNMERGE_PATH, traveller.NewUnmergeProfile())
	reg.Register("GET", API_PATH_ADMIN_ID, admin.NewRetreiveConfig())
	reg.Register("DELETE", API_PATH_ADMIN_ID, admin.NewDeleteDomain())
	reg.Register("PUT", API_PATH_ADMIN_ID, admin.NewUpdateDomain())
	reg.Register("POST", API_PATH, admin.NewCreateDomain())
	reg.Register("GET", API_PATH, admin.NewListUcpDomains())
	reg.Register("GET", CONNECTOR_PATH, admin.NewListConnectors())
	reg.Register("POST", CONNECTOR_PATH+"/link", admin.NewLinkIndustryConnector())
	reg.Register("GET", ERROR_PATH, admin.NewListErrors())
	reg.Register("DELETE", ERROR_PATH_ID, admin.NewDeleteError())
	reg.Register("GET", JOBS_PATH, admin.NewGetJobsStatus())
	reg.Register("POST", JOBS_PATH, admin.NewStartJobs())
	reg.Register("POST", FLOWS_PATH, admin.NewStartFlows())
	reg.Register("POST", PORTAL_CONFIG_PATH, admin.NewUpdatePortalConfig())
	reg.Register("GET", PORTAL_CONFIG_PATH, admin.NewGetPortalConfig())
	reg.Register("GET", ASYNC_PATH, admin.NewGetAsyncStatus())
	reg.Register("GET", RULE_SET_PATH, admin.NewListIdResRuleSets())
	reg.Register("POST", RULE_SET_PATH, admin.NewSaveIdResRuleSet())
	reg.Register("POST", RULE_SET_PATH_ACTIVATE, admin.NewActivateIdResRuleSet())
	reg.Register("GET", RULE_SET_PATH_CACHE, admin.NewListCacheRuleSets())
	reg.Register("POST", RULE_SET_PATH_CACHE, admin.NewSaveCacheRuleSet())
	reg.Register("POST", RULE_SET_PATH_ACTIVATE_CACHE, admin.NewActivateCacheRuleSet())
	reg.Register("POST", REBUILD_CACHE_PATH, admin.NewRebuildCache())
	reg.Register("GET", PRIVACY_PATH, privacy.NewListPrivacySearches())
	reg.Register("GET", PRIVACY_PATH_ID, privacy.NewGetPrivacySearch())
	reg.Register("POST", PRIVACY_PATH, privacy.NewCreatePrivacySearch())
	reg.Register("DELETE", PRIVACY_PATH, privacy.NewDeletePrivacySearch())
	reg.Register("POST", PRIVACY_PATH_PURGE, privacy.NewCreatePrivacyDataPurge())
	reg.Register("GET", PRIVACY_PATH_PURGE, privacy.NewGetPurgeStatus())
	reg.Register("GET", INTERACTION_HISTORY_PATH_ID, traveller.NewRetrieveInteractionHistory())
	reg.Register("GET", PROMPT_CONFIG_PATH, admin.NewGetPromptConfig())
	reg.Register("POST", PROMPT_CONFIG_PATH, admin.NewUpdatePromptConfig())
	return reg.Run(req)
}

// this function validates the Api gateway request first before we do anything
// this prevents issues with unsecure headers that could impact logs and other behaviors downstream
func ValidateApiGatewayRequest(req events.APIGatewayProxyRequest) error {
	tidHeader := req.Headers[core.TRANSACTION_ID_HEADER]
	domainHeader := req.Headers[CUSTOMER_PROFILE_DOMAIN_HEADER]
	//restricting length of transaction ID to 36 char (UUID)
	if tidHeader != "" && len(tidHeader) > 36 {
		return errors.New("transaction ID header is too long. Should be limited to 36 characters")
	}
	if domainHeader != "" {
		err := customerprofiles.ValidateDomainName(domainHeader)
		if err != nil {
			return fmt.Errorf("domain name provided, but invalid: %v", err)
		}
	}
	return nil
}

func buildResponseError(err error) events.APIGatewayProxyResponse {
	log.Printf("Error: %v", err)
	wrapper := model.ResponseWrapper{
		Error: core.BuildResError(err),
	}
	body, err := json.Marshal(wrapper)
	if err != nil {
		body = []byte("unable to marshal error response")
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 400,
		Body:       string(body),
	}
}

func main() {
	lambda.Start(HandleRequest)
}

func initStorage(domainName string) customerprofiles.ICustomerProfileLowCostConfig {
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
	customCfg.SetDomain(domainName)

	return customCfg
}

func getUserCredentials(sm secrets.Config) (string, string) {
	log.Printf("Retrieving DB credentials from Secrets Manager: %v", sm.SecretArn)
	user := sm.Get("username")
	pwd := sm.Get("password")
	return user, pwd
}
