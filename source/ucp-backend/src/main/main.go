package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"os"
	core "tah/core/core"
	customerprofiles "tah/core/customerprofiles"
	model "tah/ucp/src/business-logic/model"
	usecase "tah/ucp/src/business-logic/usecase"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var LAMBDA_ENV = os.Getenv("LAMBDA_ENV")
var S3_MULTIMEDIA = os.Getenv("S3_MULTIMEDIA")
var LAMBDA_REGION = os.Getenv("AWS_REGION")
var NAMESPACE_NAME = "cloudRackServiceDiscoveryNamespace" + LAMBDA_ENV
var ATHENA_WORKGROUP = os.Getenv("ATHENA_WORKGROUP")
var ATHENA_DB = os.Getenv("ATHENA_DB")
var UCP_CONNECT_DOMAIN = ""
var FN_RETREIVE_UCP_PROFILE = "retreive_ucp_profile"
var FN_DELETE_UCP_PROFILE = "delete_ucp_profile"
var FN_SEARCH_UCP_PROFILES = "search_ucp_profiles"
var FN_RETREIVE_UCP_CONFIG = "retreive_ucp_config"
var FN_LIST_UCP_DOMAINS = "list_ucp_domains"
var FN_CREATE_UCP_DOMAIN = "create_ucp_domain"
var FN_DELETE_UCP_DOMAIN = "delete_ucp_domain"
var FN_MERGE_UCP_PROFILE = "merge_ucp_profile"
var FN_LIST_UCP_INGESTION_ERROR = "list_ucp_ingestion_errors"
var CUSTOMER_PROFILE_DOMAIN_HEADER = "customer-profiles-domain"

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received Request %+v with context %+v", req, ctx)
	resource := req.Resource
	method := req.HTTPMethod
	log.Printf("*Resource: %v", resource)
	log.Printf("*Method: %v", method)
	subFunction := identifyUseCase(resource, method)
	var err error
	var ucpRes model.ResWrapper
	var profiles = customerprofiles.InitWithDomain(req.Headers[CUSTOMER_PROFILE_DOMAIN_HEADER], "eu-central-1")

	if subFunction == FN_RETREIVE_UCP_PROFILE {
		log.Printf("Selected Use Case %v", subFunction)
		if err == nil {
			rq := model.UCPRequest{
				ID: req.PathParameters["id"],
			}
			log.Printf("Retreive request: %+v", rq)
			err = ValidateUCPRetreiveRequest(rq)
			if err != nil {
				log.Printf("Validation error: %v", err)
				return builResponseError(err), nil
			}
			ucpRes, err = usecase.RetreiveUCPProfile(rq, profiles)
			if err == nil {
				log.Printf("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(ucpRes), nil
			}
		}
		return builResponseError(err), nil
	} else if subFunction == FN_DELETE_UCP_PROFILE {
		log.Printf("Selected Use Case %v", subFunction)
		if err == nil {
			rq := model.UCPRequest{
				ID: req.PathParameters["id"],
			}
			log.Printf("Delete request: %+v", rq)
			err = ValidateUCPRetreiveRequest(rq)
			if err != nil {
				log.Printf("Validation error: %v", err)
				return builResponseError(err), nil
			}
			err = usecase.DeleteUCPProfile(rq, profiles)
		}
		return builResponseError(err), nil
	} else if subFunction == FN_SEARCH_UCP_PROFILES {
		log.Printf("Selected Use Case %v", subFunction)
		if err == nil {
			rq := model.UCPRequest{SearchRq: model.SearchRq{
				LastName:  req.QueryStringParameters["lastName"],
				Phone:     req.QueryStringParameters["phone"],
				Email:     req.QueryStringParameters["email"],
				LoyaltyID: req.QueryStringParameters["loyaltyId"],
			}}
			err = ValidateUCPSearchRequest(rq)
			if err != nil {
				log.Printf("Validation error: %v", err)
				return builResponseError(err), nil
			}
			log.Printf("Search request: %+v", rq)
			ucpRes, err = usecase.SearchUCPProfile(rq, profiles)
			if err == nil {
				log.Printf("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(ucpRes), nil
			}
		}
		return builResponseError(err), nil
	} else if subFunction == FN_RETREIVE_UCP_CONFIG {
		log.Printf("Selected Use Case %v", subFunction)
		if err == nil {
			rq := model.UCPRequest{}
			log.Printf("Search request: %+v", rq)
			ucpRes, err = usecase.RetreiveUCPConfig(rq, profiles)
			if err == nil {
				log.Printf("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(ucpRes), nil
			}
		}
		return builResponseError(err), nil
	} else if subFunction == FN_LIST_UCP_DOMAINS {
		log.Printf("Selected Use Case %v", subFunction)
		if err == nil {
			rq := model.UCPRequest{}
			log.Printf("Search request: %+v", rq)
			ucpRes, err = usecase.ListUcpDomains(rq, profiles)
			if err == nil {
				log.Printf("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(ucpRes), nil
			}
		}
		return builResponseError(err), nil
	} else if subFunction == FN_CREATE_UCP_DOMAIN {
		log.Printf("Selected Use Case %v", subFunction)
		if err == nil {
			rq := model.UCPRequest{}
			rq, err = decodeUCPBody(req)
			log.Printf("Create Domain request: %+v", rq)
			if err == nil {
				ucpRes, err = usecase.CreateUcpDomain(rq, profiles)
				if err == nil {
					log.Printf("Use Case %s failed with error: %v", subFunction, err)
					return builUCPResponse(ucpRes), nil
				}
			}
		}
		return builResponseError(err), nil
	} else if subFunction == FN_DELETE_UCP_DOMAIN {
		log.Printf("Selected Use Case %v", subFunction)
		if err == nil {
			rq := model.UCPRequest{
				Domain: model.Domain{
					Name: req.PathParameters["id"],
				},
			}
			log.Printf("Delete request: %+v", rq)
			ucpRes, err = usecase.DeleteUcpDomain(rq, profiles)
			if err == nil {
				log.Printf("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(ucpRes), nil
			}
		}
		return builResponseError(err), nil
	} else if subFunction == FN_MERGE_UCP_PROFILE {
		log.Printf("Selected Use Case %v", subFunction)
		if err == nil {
			rq := model.UCPRequest{}
			log.Printf("Search request: %+v", rq)
			ucpRes, err = usecase.MergeUCPConfig(rq)
			if err == nil {
				log.Printf("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(ucpRes), nil
			}
		}
		return builResponseError(err), nil
	} else if subFunction == FN_LIST_UCP_INGESTION_ERROR {
		log.Printf("Selected Use Case %v", subFunction)
		if err == nil {
			rq := model.UCPRequest{}
			log.Printf("Search request: %+v", rq)
			ucpRes, err = usecase.ListUCPIngestionError(rq, profiles)
			if err == nil {
				log.Printf("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(ucpRes), nil
			}
		}
		return builResponseError(err), nil
	}

	log.Printf("No Use Case Found for %v", method+" "+resource)
	err = errors.New("No Use Case Found for " + method + " " + resource)
	return builResponseError(err), nil
}

func decodeUCPBody(req events.APIGatewayProxyRequest) (model.UCPRequest, error) {
	wrapper := model.UCPRequest{}
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, _ := base64.StdEncoding.DecodeString(req.Body)
		decodedBody = base64Body
	}
	if err := json.Unmarshal(decodedBody, &wrapper); err != nil {
		return model.UCPRequest{}, err
	}
	log.Printf("Decoded Body %+v", wrapper)
	return wrapper, nil
}

func identifyUseCase(res string, meth string) string {
	if res == "/ucp/profile/{id}" && meth == "GET" {
		return FN_RETREIVE_UCP_PROFILE
	}
	if res == "/ucp/profile/{id}" && meth == "DELETE" {
		return FN_DELETE_UCP_PROFILE
	}
	if res == "/ucp/profile" && meth == "GET" {
		return FN_SEARCH_UCP_PROFILES
	}
	if res == "/ucp/admin/{id}" && meth == "GET" {
		return FN_RETREIVE_UCP_CONFIG
	}
	if res == "/ucp/admin/{id}" && meth == "DELETE" {
		return FN_DELETE_UCP_DOMAIN
	}
	if res == "/ucp/admin" && meth == "POST" {
		return FN_CREATE_UCP_DOMAIN
	}
	if res == "/ucp/admin" && meth == "GET" {
		return FN_LIST_UCP_DOMAINS
	}
	if res == "/ucp/merge" && meth == "POST" {
		return FN_MERGE_UCP_PROFILE
	}
	if res == "/ucp/error" && meth == "GET" {
		return FN_LIST_UCP_INGESTION_ERROR
	}
	return ""
}

func builUCPResponse(resWrapper model.ResWrapper) events.APIGatewayProxyResponse {
	res := events.APIGatewayProxyResponse{
		StatusCode: 200,
	}
	jsonRes, err := json.Marshal(resWrapper)
	if err != nil {
		log.Printf("Error while unmarshalling response %v", err)
		return builResponseError(err)
	}
	res.Body = string(jsonRes)
	return res
}

func builResponseError(err error) events.APIGatewayProxyResponse {
	log.Printf("[LOYALTY] Response Error: %v", err)
	res := events.APIGatewayProxyResponse{
		StatusCode: 400,
	}
	json, _ := json.Marshal(model.ResWrapper{Error: core.BuildResError(err)})
	res.Body = string(json)
	return res
}

func ValidateUCPSearchRequest(wrapper model.UCPRequest) error {
	if wrapper.SearchRq.LastName == "" && wrapper.SearchRq.LoyaltyID == "" && wrapper.SearchRq.Phone == "" && wrapper.SearchRq.Email == "" {
		return errors.New("At least one search creteria within LastName, LoyaltyID, Phone and Email must be provided")
	}
	return nil
}

func ValidateUCPRetreiveRequest(wrapper model.UCPRequest) error {
	if wrapper.ID == "" {
		return errors.New("Profile ID is required to retreive profile")
	}
	return nil
}
func main() {
	lambda.Start(HandleRequest)
}
