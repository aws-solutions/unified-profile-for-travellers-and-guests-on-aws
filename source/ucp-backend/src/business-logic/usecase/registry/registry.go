// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/appflow"
	"tah/upt/source/tah-core/appregistry"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/bedrock"
	"tah/upt/source/tah-core/cognito"
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/datasync"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/ecs"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/iam"
	"tah/upt/source/tah-core/lambda"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	filter "tah/upt/source/ucp-backend/src/business-logic/usecase/task/filter"
	"tah/upt/source/ucp-common/src/constant/admin"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	travModel "tah/upt/source/ucp-common/src/model/traveler"
	utils "tah/upt/source/ucp-common/src/utils/utils"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

const shouldUsePermissionSystemKey = "USE_PERMISSION_SYSTEM"
const packageName = "ucp-backend"

type Registry struct {
	Tx                   *core.Transaction
	Reg                  map[string]Usecase
	Region               string
	Env                  map[string]string
	AppRegistry          *appregistry.Config
	Iam                  *iam.Config
	Glue                 *glue.Config
	Accp                 customerprofiles.ICustomerProfileLowCostConfig
	Cognito              cognito.ICognitoConfig
	ConfigDB             *db.DBConfig
	ErrorDB              *db.DBConfig
	MatchDB              *db.DBConfig
	PrivacyDB            *db.DBConfig
	PortalConfigDB       *db.DBConfig
	AppFlow              appflow.IConfig
	SyncLambda           lambda.IConfig
	AsyncLambda          lambda.IConfig
	RetryLambda          lambda.IConfig
	DataSync             datasync.IConfig
	SolutionUtils        awssolutions.IConfig
	Bedrock              *bedrock.Config
	DataAccessPermission string
	AppAccessPermission  admin.AppPermission
	Ecs                  EcsApi
	Ssm                  SsmApi
}

type Usecase interface {
	Name() string               //Returns name of the use case
	Tx() core.Transaction       //Return the transaction object associated to the use case
	SetTx(tx *core.Transaction) //Associate a transaction to be used for traceability
	SetRegistry(
		r *Registry,
	) //Associate the use case registry to the use case object.
	Registry() *Registry //Associate the use case registry to the use case object.
	ValidateRequest(
		req model.RequestWrapper,
	) error //Validate the incoming request based on use case functional requirements
	CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error)    //Create the request object from API gateway
	CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) //Create the API Gateway response object
	Run(req model.RequestWrapper) (model.ResponseWrapper, error)                      //Execute the use case business logic
	AccessPermission() admin.AppPermission                                            //Fetch access permission associated with the usecase
}

type SsmApi interface {
	GetParametersByPath(context.Context, string) (map[string]string, error)
}

type EcsApi interface {
	ListTasks(context.Context, ecs.TaskConfig, *string, *string) ([]string, error)
}

type ServiceHandlers struct {
	AppRegistry    *appregistry.Config
	Iam            *iam.Config
	Glue           *glue.Config
	Accp           customerprofiles.ICustomerProfileLowCostConfig
	Cognito        cognito.ICognitoConfig
	ConfigDB       *db.DBConfig
	ErrorDB        *db.DBConfig
	MatchDB        *db.DBConfig
	PrivacyDB      *db.DBConfig
	PortalConfigDB *db.DBConfig
	AppFlow        *appflow.Config
	SyncLambda     *lambda.Config
	AsyncLambda    lambda.IConfig
	RetryLambda    lambda.IConfig
	DataSync       datasync.IConfig
	SolutionUtils  awssolutions.IConfig
	Bedrock        *bedrock.Config
	SsmCfg         SsmApi
	EcsCfg         EcsApi
}

func NewRegistry(region string, logLevel core.LogLevel, handlers ServiceHandlers) Registry {
	//we initialize a transaction for the pre-handler logic (before the lambda function actually processes it's first request)
	tx := core.NewTransaction(packageName, "", logLevel)
	return Registry{
		Tx:             &tx,
		Reg:            map[string]Usecase{},
		Region:         region,
		AppRegistry:    handlers.AppRegistry,
		Iam:            handlers.Iam,
		Glue:           handlers.Glue,
		Accp:           handlers.Accp,
		Cognito:        handlers.Cognito,
		ConfigDB:       handlers.ConfigDB,
		ErrorDB:        handlers.ErrorDB,
		MatchDB:        handlers.MatchDB,
		PrivacyDB:      handlers.PrivacyDB,
		PortalConfigDB: handlers.PortalConfigDB,
		AppFlow:        handlers.AppFlow,
		SyncLambda:     handlers.SyncLambda,
		AsyncLambda:    handlers.AsyncLambda,
		RetryLambda:    handlers.RetryLambda,
		DataSync:       handlers.DataSync,
		SolutionUtils:  handlers.SolutionUtils,
		Bedrock:        handlers.Bedrock,
		Env:            map[string]string{},
		Ssm:            handlers.SsmCfg,
		Ecs:            handlers.EcsCfg,
	}
}

func (r *Registry) Register(httpMethod string, path string, uc Usecase) {
	r.Tx.Debug("[Register] Registering path %s, to endpoint %s -%s", uc.Name(), httpMethod, path)
	uc.SetRegistry(r)
	uc.SetTx(r.Tx)
	r.Reg[httpMethod+path] = uc
}

func (r *Registry) AddEnv(key string, value string) {
	r.Tx.Debug("[Register] Adding environment variable %s=>%s to registry", key, value)
	if len(r.Env) == 0 {
		r.Env = map[string]string{}
	}
	r.Env[key] = value
}

func (r *Registry) SetDataAccessPermission(permission string) {
	r.DataAccessPermission = permission
}

func (r *Registry) SetAppAccessPermission(permission admin.AppPermission) {
	r.AppAccessPermission = permission
}

// Setting the transaction to the registry and all use case (for log traceability)
func (r *Registry) SetTx(tx *core.Transaction) {
	r.Tx.Debug("[Register] Setting transaction %+v in use case registry", tx)
	r.Tx = tx
	for i := range r.Reg {
		//we create a new transaction with the same transaction ID and the use case name as log Prefix
		ucTx := core.NewTransaction(r.Reg[i].Name(), tx.TransactionID, tx.LogLevel)
		ucTx.Debug("[Register] Setting transaction %+v in use case %v", ucTx, r.Reg[i].Name())
		r.Reg[i].SetTx(&ucTx)
	}
}
func (r *Registry) SetRegion(region string) {
	r.Region = region
}

func (r *Registry) ValidateUserAccess(uc Usecase) bool {
	return uc.AccessPermission() == admin.PublicAccessPermission || uc.AccessPermission()&r.AppAccessPermission == uc.AccessPermission()
}

func (r *Registry) Run(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r.Tx.Debug("[Run] Processing Request from API Gateway")
	method := req.HTTPMethod
	resource := req.Resource
	r.Tx.Debug("[Run] Checking Data access permission for domain '%s'", r.Env["ACCP_DOMAIN_NAME"])
	dataPermission, appPermission, err := filter.GetUserPermission(r.isPermissionSystemEnabled(), r.Cognito, req, r.Env["ACCP_DOMAIN_NAME"])
	if err != nil {
		return r.BuildErrorResponse("GetUserPermission", "FetchPermissions", 400, errors.New("could not access user permissions")), nil
	}
	r.Tx.Info("[Run] User has data permission '%s' for domain %s", dataPermission, r.Env["ACCP_DOMAIN_NAME"])
	r.SetDataAccessPermission(dataPermission)
	r.SetAppAccessPermission(appPermission)

	r.Tx.Debug("[Run] Looking for registered use case with path %s and Method %s", resource, method)
	if uc, ok := r.Reg[method+resource]; ok {
		r.Tx.Debug("[Run][%s] Validate user access", uc.Name())
		if !r.ValidateUserAccess(uc) {
			r.Tx.Error(
				"[Run][%s] Access denied: User %s is unauthorized to perform this operation",
				uc.Name(),
				cognito.ParseUserFromLambdaRq(req),
			)
			return r.BuildErrorResponse(uc.Name(), "ValidateUserAccess", 403, errors.New("user cannot perform action: "+uc.Name())), nil
		}
		r.Tx.Info("[Run][%s] Found use case for path %s and Method %s", uc.Name(), resource, method)
		r.Tx.Debug("[Run][%s] Creating request wrapper", uc.Name())
		requestWrapper, err1 := uc.CreateRequest(req)
		if err1 != nil {
			r.Tx.Error("[Run][%s] Error creating request: %v", uc.Name(), err1)
			return r.BuildErrorResponse(uc.Name(), "CreateRequest", 400, errors.New("could not parse request: "+err1.Error())), nil
		}
		r.Tx.Debug("[Run][%s] Request wrapper successfully created", uc.Name())
		r.Tx.Debug("[Run][%s] Validating inputs %v", uc.Name())
		err0 := uc.ValidateRequest(requestWrapper)
		if err0 != nil {
			r.Tx.Error("[Run][%s] Request validation error: %v", uc.Name(), err0)
			return r.BuildErrorResponse(uc.Name(), "ValidateRequest", 400, errors.New("validation error: "+err0.Error())), nil
		}
		r.Tx.Info("[Run][%s] Request successfully validated", uc.Name())
		r.Tx.Info("[Run][%s] Executing use case", uc.Name())
		now := time.Now()
		responseWrapper, err2 := uc.Run(requestWrapper)
		ucDuration := time.Since(now)
		if err2 != nil {
			r.Tx.Error("[Run][%s] Use case execution completed in %v with error: %v", uc.Name(), ucDuration, err2)
			return r.BuildErrorResponse(uc.Name(), "Run", 400, errors.New("use case execution failed")), nil
		}
		r.Tx.Info("[Run][%s] Use case execution successfully completed in %v", uc.Name(), ucDuration)
		responseWrapper.TxID = r.Tx.TransactionID
		responseWrapper, err = r.Enrich(responseWrapper)
		if err != nil {
			r.Tx.Error("[Run][%s] Could not enrich response: %v", uc.Name(), err)
			return r.BuildErrorResponse(uc.Name(), "Enrich", 400, errors.New("failed to create response")), nil
		}
		r.FilterTravellerData(&responseWrapper, dataPermission)
		apiGatewayResponse, err3 := uc.CreateResponse(responseWrapper)
		if err3 != nil {
			r.Tx.Error("[Run][%s] Could not create response: %v", uc.Name(), err3)
			return r.BuildErrorResponse(uc.Name(), "CreateResponse", 400, errors.New("failed to create response")), nil
		}
		apiGatewayResponse.Headers[core.TRANSACTION_ID_HEADER] = r.Tx.TransactionID

		_, err = r.SolutionUtils.SendMetrics(map[string]interface{}{
			"service":  packageName,
			"usecase":  uc.Name(),
			"duration": ucDuration.Milliseconds(),
			"status":   "success",
		})
		if err != nil {
			r.Tx.Warn("[Run][%s] Could not send metrics: %v", uc.Name(), err)
		}
		return apiGatewayResponse, nil
	}
	return r.BuildErrorResponse(
		"undefined",
		"SetUsecase",
		404,
		fmt.Errorf("use case not found for path %s and method %s", resource, method),
	), nil
}

func (r Registry) Enrich(res model.ResponseWrapper) (model.ResponseWrapper, error) {
	s3BucketsToReturn := map[string]string{
		"S3_HOTEL_BOOKING":              r.Env["S3_HOTEL_BOOKING"],
		"S3_AIR_BOOKING":                r.Env["S3_AIR_BOOKING"],
		"S3_GUEST_PROFILE":              r.Env["S3_GUEST_PROFILE"],
		"S3_PAX_PROFILE":                r.Env["S3_PAX_PROFILE"],
		"S3_STAY_REVENUE":               r.Env["S3_STAY_REVENUE"],
		"S3_CLICKSTREAM":                r.Env["S3_CLICKSTREAM"],
		"S3_CSI":                        r.Env["S3_CSI"],
		"CONNECT_PROFILE_SOURCE_BUCKET": r.Env["CONNECT_PROFILE_SOURCE_BUCKET"],
		"REGION":                        r.Region,
	}
	if res.AwsResources.S3Buckets == nil {
		res.AwsResources.S3Buckets = map[string]string{}
	}
	for k, v := range s3BucketsToReturn {
		res.AwsResources.S3Buckets[k] = v
	}
	if r.Accp != nil {
		objectMappings, err := r.Accp.GetObjectLevelFields()
		if err != nil {
			return res, err
		}
		if objectMappings == nil {
			return res, nil
		}
		accpMappings := []commonModel.AccpRecord{}
		for objectType, mapping := range objectMappings {
			accpMapping := commonModel.AccpRecord{
				Name:         objectType,
				ObjectFields: mapping,
			}
			accpMappings = append(accpMappings, accpMapping)
		}
		res.AccpRecords = accpMappings
	}
	return res, nil
}

// Filter out traveller data based on the user's permissions
func (r Registry) FilterTravellerData(res *model.ResponseWrapper, permissionString string) {
	originalProfiles := utils.SafeDereference(res.Profiles)
	filteredProfiles := []travModel.Traveller{}
	for i := range originalProfiles {
		profile := filter.Filter(originalProfiles[i], permissionString)
		if profile != nil {
			filteredProfiles = append(filteredProfiles, *profile)
		}
	}

	res.Profiles = &filteredProfiles
}

func (r Registry) BuildErrorResponse(usecase, errorStep string, statusCode int, err error) events.APIGatewayProxyResponse {
	wrapper := model.ResponseWrapper{
		Error: core.BuildResError(err),
	}
	wrapper.TxID = r.Tx.TransactionID
	body, err2 := json.Marshal(wrapper)
	if err2 != nil {
		r.Tx.Error("[buildErrorResponse] Response Marshalling error: %v", err2)
		body = []byte(`{ "errors" : [{"msg" : "An unexpected error occurred"}]}`)
	}
	rs := events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{core.TRANSACTION_ID_HEADER: r.Tx.TransactionID},
		Body:       string(body),
	}
	r.Tx.Debug("[buildErrorResponse] Error Response successfully created: %v", rs)
	_, err = r.SolutionUtils.SendMetrics(map[string]interface{}{
		"service":    packageName,
		"usecase":    usecase,
		"status":     "failed",
		"error_step": errorStep,
		"error":      err.Error(),
	})
	if err != nil {
		r.Tx.Warn("[buildErrorResponse] Could not send metrics: %v", err)
	}
	return rs
}

// Check if the customer is using the granular permission system. Default is true unless customer explicitly deploys with granular permissions disabled.
func (r Registry) isPermissionSystemEnabled() bool {
	if r.Env[shouldUsePermissionSystemKey] == "" {
		return true // case for customers who deployed infra before the CloudFormation Parameter was added and passed as a Lambda env var
	}

	return r.Env[shouldUsePermissionSystemKey] == "true"
}

/***
* Utility functions
**/

func DecodeBody(uc Usecase, req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	wrapper := model.RequestWrapper{}
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, _ := base64.StdEncoding.DecodeString(req.Body)
		decodedBody = base64Body
	}
	if err := json.Unmarshal(decodedBody, &wrapper); err != nil {
		return model.RequestWrapper{}, errors.New("error parsing body")
	}
	return wrapper, nil
}

func CreateRequest(uc Usecase, req events.APIGatewayProxyRequest) (wrapper model.RequestWrapper, err error) {
	if req.Body != "" {
		wrapper, err = DecodeBody(uc, req)
	}
	wrapper.Pagination = []commonModel.PaginationOptions{
		{
			Page:     utils.ParseInt(req.QueryStringParameters[model.PAGINATION_OPTION_PAGE]),
			PageSize: utils.ParseInt(req.QueryStringParameters[model.PAGINATION_OPTION_PAGE_SIZE]),
		}}
	return wrapper, err
}

func CreateResponse(uc Usecase, res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	uc.Tx().Debug("[%s] Creating Response", uc.Name())
	apiRes := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			core.TRANSACTION_ID_HEADER: uc.Tx().TransactionID,
			"Content-Type":             "application/json"},
	}
	jsonRes, err := json.Marshal(res)
	if err != nil {
		uc.Tx().Error("[%s] Error marshalling response %v", uc.Name(), err)
		return apiRes, err
	}
	apiRes.Body = string(jsonRes)
	return apiRes, nil
}
