// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"tah/core/appflow"
	"tah/core/appregistry"
	core "tah/core/core"
	"tah/core/customerprofiles"
	"tah/core/db"
	"tah/core/glue"
	"tah/core/iam"
	"tah/core/lambda"

	"time"

	model "tah/ucp/src/business-logic/model/common"
	utils "tah/ucp/src/business-logic/utils"

	"github.com/aws/aws-lambda-go/events"
)

type Registry struct {
	Tx          *core.Transaction
	Reg         map[string]Usecase
	Region      string
	Env         map[string]string
	AppRegistry *appregistry.Config
	Iam         *iam.Config
	Glue        *glue.Config
	Accp        *customerprofiles.CustomerProfileConfig
	ErrorDB     *db.DBConfig
	AppFlow     *appflow.Config
	SyncLambda  *lambda.Config
}

type Usecase interface {
	Name() string                                                                     //Returns name of the use case
	Tx() core.Transaction                                                             //Return the traneaction object assocuated to the use case
	SetTx(tx *core.Transaction)                                                       //Associate a transaction to be used for traceability
	SetRegistry(r *Registry)                                                          //Associate the uuse case registry to the use case object.
	Registry() *Registry                                                              //Associate the uuse case registry to the use case object.
	ValidateRequest(req model.RequestWrapper) error                                   //Validate the incoming request based on use case functional requirements
	CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error)    //Create the request object from API gateway
	CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) //Create the API Gateway response object
	Run(req model.RequestWrapper) (model.ResponseWrapper, error)                      //Excecute the use case business logic
}

type ServiceHandlers struct {
	AppRegistry *appregistry.Config
	Iam         *iam.Config
	Glue        *glue.Config
	Accp        *customerprofiles.CustomerProfileConfig
	ErrorDB     *db.DBConfig
	AppFlow     *appflow.Config
	SyncLambda  *lambda.Config
}

func NewRegistry(region string, handlers ServiceHandlers) Registry {
	//we initialize a transaction for the pre-handler locgic (before the lambda function actually processes it's first request)
	tx := core.NewTransaction("ind_connector", "")
	return Registry{
		Tx:          &tx,
		Reg:         map[string]Usecase{},
		Region:      region,
		AppRegistry: handlers.AppRegistry,
		Iam:         handlers.Iam,
		Glue:        handlers.Glue,
		Accp:        handlers.Accp,
		ErrorDB:     handlers.ErrorDB,
		AppFlow:     handlers.AppFlow,
		SyncLambda:  handlers.SyncLambda,
		Env:         map[string]string{},
	}
}

func (r *Registry) Register(httpMethod string, path string, uc Usecase) {
	r.Log("[Register] Registering path %s, to endpoint %s -%s", uc.Name(), httpMethod, path)
	uc.SetRegistry(r)
	uc.SetTx(r.Tx)
	r.Reg[httpMethod+path] = uc
}

func (r *Registry) AddEnv(key string, value string) {
	r.Log("[Register] Adding evitonment variable %s=>%s to registry", key, value)
	r.Env[key] = value
}

func (r *Registry) Log(format string, v ...interface{}) {
	r.Tx.Log(format, v...)
}

//Setting the transaction to the registry and all use case (for log traceability)
func (r *Registry) SetTx(tx *core.Transaction) {
	r.Log("[Register] Setting transaction %+v in use case registery", tx)
	r.Tx = tx
	for i, _ := range r.Reg {
		//we create a new transaction with the same transactiono ID and the use case name as log Prefix
		ucTx := core.NewTransaction(r.Reg[i].Name(), tx.TransactionID)
		r.Log("[Register] Setting transaction %+v in use case %v", ucTx, r.Reg[i].Name())
		r.Reg[i].SetTx(&ucTx)
	}
}
func (r *Registry) SetRegion(region string) {
	r.Region = region
}

func (r Registry) Run(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r.Log("[Run] Request from API Gateway: %+v", req)
	method := req.HTTPMethod
	resource := req.Resource
	r.Log("[Run] Looking for registered use case with path %s and Method %s", resource, method)
	if uc, ok := r.Reg[method+resource]; ok {
		r.Log("[Run][%s] Found use case for path %s and Method %s", uc.Name(), resource, method)
		r.Log("[Run][%s] Creating request wrapper", uc.Name())
		requestWrapper, err1 := uc.CreateRequest(req)
		if err1 != nil {
			r.Log("[Run][%s] Error creating request: %v", uc.Name(), err1)
			return r.BuildErrorResponse(400, errors.New("Could not parse request: "+err1.Error())), nil
		}
		r.Log("[Run][%s] Request wrapper successfully created", uc.Name())
		r.Log("[Run][%s] Validating inputs %v", uc.Name())
		err0 := uc.ValidateRequest(requestWrapper)
		if err0 != nil {
			r.Log("[Run][%s] Request validation error: %v", uc.Name(), err0)
			return r.BuildErrorResponse(400, errors.New("Validation Error: "+err0.Error())), nil
		}
		r.Log("[Run][%s] Request sucecssfully validated", uc.Name())
		r.Log("[Run][%s] Executing use case", uc.Name())
		now := time.Now()
		responseWrapper, err2 := uc.Run(requestWrapper)
		ucDuration := time.Since(now)
		if err2 != nil {
			r.Log("[Run][%s] Use case execusion completed in %v with error: %v", uc.Name(), ucDuration, err2)
			return r.BuildErrorResponse(400, errors.New("Use case execusion failed: "+err2.Error())), nil
		}
		r.Log("[Run][%s] Use case execusion successfully completed in %v", uc.Name(), ucDuration)
		responseWrapper.TxID = r.Tx.TransactionID
		responseWrapper = r.Enrich(responseWrapper)
		apiGatewayResponse, err3 := uc.CreateResponse(responseWrapper)
		if err3 != nil {
			r.Log("[Run][%s] Could not create response: %v", uc.Name(), err3)
			return r.BuildErrorResponse(400, errors.New("Could not create response: "+err3.Error())), nil
		}
		apiGatewayResponse.Headers[core.TRANSACTION_ID_HEADER] = r.Tx.TransactionID

		return apiGatewayResponse, nil
	}
	return r.BuildErrorResponse(400, errors.New(fmt.Sprintf("No use case registered for path %s and Method %s", resource, method))), nil
}

func (r Registry) Enrich(res model.ResponseWrapper) model.ResponseWrapper {
	s3BucketsToReturn := map[string]string{
		"S3_HOTEL_BOOKING":              r.Env["S3_HOTEL_BOOKING"],
		"S3_AIR_BOOKING":                r.Env["S3_AIR_BOOKING"],
		"S3_GUEST_PROFILE":              r.Env["S3_GUEST_PROFILE"],
		"S3_PAX_PROFILE":                r.Env["S3_PAX_PROFILE"],
		"S3_STAY_REVENUE":               r.Env["S3_STAY_REVENUE"],
		"S3_CLICKSTREAM":                r.Env["S3_CLICKSTREAM"],
		"CONNECT_PROFILE_SOURCE_BUCKET": r.Env["CONNECT_PROFILE_SOURCE_BUCKET"],
	}
	if res.AwsResources.S3Buckets == nil {
		res.AwsResources.S3Buckets = map[string]string{}
	}
	for k, v := range s3BucketsToReturn {

		res.AwsResources.S3Buckets[k] = v
	}
	return res
}

func (r Registry) BuildErrorResponse(statusCode int, err error) events.APIGatewayProxyResponse {
	wrapper := model.ResponseWrapper{
		Error: core.BuildResError(err),
	}
	body, err2 := json.Marshal(wrapper)
	if err2 != nil {
		r.Log("[buildErrorResponse] Response Marshalling error: %v", err2)
		body = []byte(`{ "errors" : [{"msg" : "An unexpected error occured"}]}`)
	}
	rs := events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{core.TRANSACTION_ID_HEADER: r.Tx.TransactionID},
		Body:       string(body),
	}
	r.Log("[buildErrorResponse] Error Response successfully created: %v", rs)
	return rs
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
		return model.RequestWrapper{}, err
	}
	return wrapper, nil
}

func CreateRequest(uc Usecase, req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	wrapper := model.RequestWrapper{}
	var err error
	if req.Body != "" {
		wrapper, err = DecodeBody(uc, req)
	}
	wrapper.Pagination = []model.PaginationOptions{
		model.PaginationOptions{
			Page:     utils.ParseInt(req.QueryStringParameters[model.PAGINATION_OPTION_PAGE]),
			PageSize: utils.ParseInt(req.QueryStringParameters[model.PAGINATION_OPTION_PAGE_SIZE]),
		}}
	return wrapper, err
}

func CreateResponse(uc Usecase, res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	uc.Tx().Log("[%s] Creating Response", uc.Name())
	apiRes := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			core.TRANSACTION_ID_HEADER: uc.Tx().TransactionID,
			"Content-Type":             "application/json"},
	}
	jsonRes, err := json.Marshal(res)
	if err != nil {
		uc.Tx().Log("[%s] Error while unmarshalling response %v", uc.Name(), err)
		return apiRes, err
	}
	apiRes.Body = string(jsonRes)
	return apiRes, nil
}
