package main

import (
	"context"
	"os"
	appregistry "tah/core/appregistry"
	core "tah/core/core"
	customerprofiles "tah/core/customerprofiles"
	db "tah/core/db"
	glue "tah/core/glue"
	iam "tah/core/iam"
	"tah/ucp/src/business-logic/usecase/admin"
	registry "tah/ucp/src/business-logic/usecase/registry"
	traveller "tah/ucp/src/business-logic/usecase/traveller"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

//Resources
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

//S3 buckets
var CONNECT_PROFILE_SOURCE_BUCKET = os.Getenv("CONNECT_PROFILE_SOURCE_BUCKET")
var S3_HOTEL_BOOKING = os.Getenv("S3_HOTEL_BOOKING")
var S3_AIR_BOOKING = os.Getenv("S3_AIR_BOOKING")
var S3_GUEST_PROFILE = os.Getenv("S3_GUEST_PROFILE")
var S3_PAX_PROFILE = os.Getenv("S3_PAX_PROFILE")
var S3_STAY_REVENUE = os.Getenv("S3_STAY_REVENUE")
var S3_CLICKSTREAM = os.Getenv("S3_CLICKSTREAM")

//Job Names
var HOTEL_BOOKING_JOB_NAME = os.Getenv("HOTEL_BOOKING_JOB_NAME")
var AIR_BOOKING_JOB_NAME = os.Getenv("AIR_BOOKING_JOB_NAME")
var GUEST_PROFILE_JOB_NAME = os.Getenv("GUEST_PROFILE_JOB_NAME")
var PAX_PROFILE_JOB_NAME = os.Getenv("PAX_PROFILE_JOB_NAME")
var CLICKSTREAM_JOB_NAME = os.Getenv("CLICKSTREAM_JOB_NAME")
var HOTEL_STAY_JOB_NAME = os.Getenv("HOTEL_STAY_JOB_NAME")

var HOTEL_BOOKING_JOB_NAME_CUSTOMER = os.Getenv("HOTEL_BOOKING_JOB_NAME_CUSTOMER")
var AIR_BOOKING_JOB_NAME_CUSTOMER = os.Getenv("AIR_BOOKING_JOB_NAME_CUSTOMER")
var GUEST_PROFILE_JOB_NAME_CUSTOMER = os.Getenv("GUEST_PROFILE_JOB_NAME_CUSTOMER")
var PAX_PROFILE_JOB_NAME_CUSTOMER = os.Getenv("PAX_PROFILE_JOB_NAME_CUSTOMER")
var CLICKSTREAM_JOB_NAME_CUSTOMER = os.Getenv("CLICKSTREAM_JOB_NAME_CUSTOMER")
var HOTEL_STAY_JOB_NAME_CUSTOMER = os.Getenv("HOTEL_STAY_JOB_NAME_CUSTOMER")

var KMS_KEY_PROFILE_DOMAIN = os.Getenv("KMS_KEY_PROFILE_DOMAIN")
var UCP_CONNECT_DOMAIN = ""

var ERROR_TABLE_NAME = os.Getenv("ERROR_TABLE_NAME")
var ERROR_TABLE_PK = os.Getenv("ERROR_TABLE_PK")
var ERROR_TABLE_SK = os.Getenv("ERROR_TABLE_SK")

//Use cases
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

var appregistryClient = appregistry.Init(LAMBDA_REGION)
var iamClient = iam.Init()
var glueClient = glue.Init(LAMBDA_REGION, GLUE_DB)
var errorDB = db.Init(ERROR_TABLE_NAME, ERROR_TABLE_PK, ERROR_TABLE_SK)

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tx := core.NewTransaction("ucp", req.Headers[core.TRANSACTION_ID_HEADER])
	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")
	tx.Log("Received Request %+v with context %+v", req, ctx)
	var profiles = customerprofiles.InitWithDomain(req.Headers[CUSTOMER_PROFILE_DOMAIN_HEADER], LAMBDA_REGION)

	var reg = registry.NewRegistry(LAMBDA_REGION, &appregistryClient, &iamClient, &glueClient, &profiles, &errorDB)
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

	reg.AddEnv("HOTEL_BOOKING_JOB_NAME", HOTEL_BOOKING_JOB_NAME)
	reg.AddEnv("AIR_BOOKING_JOB_NAME", AIR_BOOKING_JOB_NAME)
	reg.AddEnv("GUEST_PROFILE_JOB_NAME", GUEST_PROFILE_JOB_NAME)
	reg.AddEnv("PAX_PROFILE_JOB_NAME", PAX_PROFILE_JOB_NAME)
	reg.AddEnv("CLICKSTREAM_JOB_NAME", CLICKSTREAM_JOB_NAME)
	reg.AddEnv("HOTEL_STAY_JOB_NAME", HOTEL_STAY_JOB_NAME)
	reg.AddEnv("HOTEL_BOOKING_JOB_NAME_CUSTOMER", HOTEL_BOOKING_JOB_NAME_CUSTOMER)
	reg.AddEnv("AIR_BOOKING_JOB_NAME_CUSTOMER", AIR_BOOKING_JOB_NAME_CUSTOMER)
	reg.AddEnv("GUEST_PROFILE_JOB_NAME_CUSTOMER", GUEST_PROFILE_JOB_NAME_CUSTOMER)
	reg.AddEnv("PAX_PROFILE_JOB_NAME_CUSTOMER", PAX_PROFILE_JOB_NAME_CUSTOMER)
	reg.AddEnv("CLICKSTREAM_JOB_NAME_CUSTOMER", CLICKSTREAM_JOB_NAME_CUSTOMER)
	reg.AddEnv("HOTEL_STAY_JOB_NAME_CUSTOMER", HOTEL_STAY_JOB_NAME_CUSTOMER)

	reg.Register("GET", "/ucp/profile/{id}", traveller.NewRetreiveProfile())
	reg.Register("DELETE", "/ucp/profile/{id}", traveller.NewDeleteProfile())
	reg.Register("GET", "/ucp/profile", traveller.NewSearchProfile())
	reg.Register("POST", "/ucp/merge", traveller.NewMergeProfile())
	reg.Register("GET", "/ucp/admin/{id}", admin.NewRetreiveConfig())
	reg.Register("DELETE", "/ucp/admin/{id}", admin.NewDeleteDomain())
	reg.Register("POST", "/ucp/admin", admin.NewCreateDomain())
	reg.Register("GET", "/ucp/admin", admin.NewListUcpDomains())
	reg.Register("GET", "/ucp/connector", admin.NewListConnectors())
	reg.Register("POST", "/ucp/connector/link", admin.NewLinkIndustryConnector())
	reg.Register("POST", "/ucp/connector/crawler", admin.NewCreateConnectorCrawler())
	reg.Register("GET", "/ucp/error", admin.NewListErrors())
	reg.Register("GET", "/ucp/data", admin.NewGetDataValidationStatus())

	return reg.Run(req)
}

func main() {
	lambda.Start(HandleRequest)
}

/*rq := model.UCPRequest{
		EnvName: LAMBDA_ENV,
		Cx:      txContext,
		Pagination: model.PaginationOptions{
			Page:     txContext.ParseQueryParamInt(req.QueryStringParameters[model.PAGINATION_OPTION_PAGE]),
			PageSize: txContext.ParseQueryParamInt(req.QueryStringParameters[model.PAGINATION_OPTION_PAGE_SIZE]),
		},
	}

	if subFunction == FN_RETREIVE_UCP_PROFILE {
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			rq.ID = req.PathParameters["id"]
			tx.Log("Retreive request: %+v", rq)
			err = ValidateUCPRetreiveRequest(rq)
			if err != nil {
				tx.Log("Validation error: %v", err)
				return builResponseError(tx, err), nil
			}
			ucpRes, err = usecase.RetreiveUCPProfile(rq, profiles)
			if err == nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(tx, ucpRes), nil
			}
		}
		return builResponseError(tx, err), nil
	} else if subFunction == FN_DELETE_UCP_PROFILE {
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			rq.ID = req.PathParameters["id"]
			tx.Log("Delete request: %+v", rq)
			err = ValidateUCPRetreiveRequest(rq)
			if err != nil {
				tx.Log("Validation error: %v", err)
				return builResponseError(tx, err), nil
			}
			err = usecase.DeleteUCPProfile(rq, profiles)
		}
		return builResponseError(tx, err), nil
	} else if subFunction == FN_SEARCH_UCP_PROFILES {
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			rq.SearchRq.LastName = req.QueryStringParameters["lastName"]
			rq.SearchRq.Phone = req.QueryStringParameters["phone"]
			rq.SearchRq.Email = req.QueryStringParameters["email"]
			rq.SearchRq.LoyaltyID = req.QueryStringParameters["loyaltyId"]
			err = ValidateUCPSearchRequest(rq)
			if err != nil {
				tx.Log("Validation error: %v", err)
				return builResponseError(tx, err), nil
			}
			tx.Log("Search request: %+v", rq)
			ucpRes, err = usecase.SearchUCPProfile(rq, profiles)
			if err == nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(tx, ucpRes), nil
			}
		}
		return builResponseError(tx, err), nil
	} else if subFunction == FN_RETREIVE_UCP_CONFIG {
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			tx.Log("Search request: %+v", rq)
			ucpRes, err = usecase.RetreiveUCPConfig(rq, profiles)
			if err == nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(tx, ucpRes), nil
			}
		}
		return builResponseError(tx, err), nil
	} else if subFunction == FN_LIST_UCP_DOMAINS {
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			tx.Log("Search request: %+v", rq)
			ucpRes, err = usecase.ListUcpDomains(rq, profiles)
			if err == nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(tx, ucpRes), nil
			}
		}
		return builResponseError(tx, err), nil
	} else if subFunction == FN_CREATE_UCP_DOMAIN {

	} else if subFunction == FN_DELETE_UCP_DOMAIN {
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			rq.Domain = model.Domain{
				Name: req.PathParameters["id"],
			}
			tx.Log("Delete request: %+v", rq)
			ucpRes, err = usecase.DeleteUcpDomain(rq, profiles)
			if err == nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(tx, ucpRes), nil
			}
		}
		return builResponseError(tx, err), nil
	} else if subFunction == FN_LIST_CONNECTORS {
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			ucpRes, err := usecase.ListIndustryConnectors(appregistryClient)
			if err != nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builResponseError(tx, err), nil
			}
			res := events.APIGatewayProxyResponse{
				StatusCode: 200,
			}
			jsonRes, err := json.Marshal(ucpRes)
			if err != nil {
				tx.Log("Error while unmarshalling response %v", err)
				return builResponseError(tx, err), nil
			}
			res.Body = string(jsonRes)
			return res, nil
		}
	} else if subFunction == FN_GET_DATA_VALIDATION_STATUS {
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			buckets := map[string]string{
				"S3_HOTEL_BOOKING": S3_HOTEL_BOOKING,
				"S3_AIR_BOOKING":   S3_AIR_BOOKING,
				"S3_GUEST_PROFILE": S3_GUEST_PROFILE,
				"S3_PAX_PROFILE":   S3_PAX_PROFILE,
				"S3_STAY_REVENUE":  S3_STAY_REVENUE,
				"S3_CLICKSTREAM":   S3_CLICKSTREAM,
			}
			ucpRes, err := usecase.GetDataValidationnStatus(rq, buckets, CONNECT_PROFILE_SOURCE_BUCKET)
			if err != nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builResponseError(tx, err), nil
			}
			res := events.APIGatewayProxyResponse{
				StatusCode: 200,
			}
			jsonRes, err := json.Marshal(ucpRes)
			if err != nil {
				tx.Log("Error while unmarshalling response %v", err)
				return builResponseError(tx, err), nil
			}
			res.Body = string(jsonRes)
			return res, nil
		}
	} else if subFunction == FN_LINK_INDUSTRY_CONNECTOR {
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			data, err := decodeLinkConnectorBody(tx, req)
			if err != nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builResponseError(tx, err), nil
			}
			glueRoleArn, bucketPolicy, err := usecase.LinkIndustryConnector(iamClient, data, LAMBDA_ACCOUNT_ID, LAMBDA_REGION, DATALAKE_ADMIN_ROLE_ARN)
			if err != nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builResponseError(tx, err), nil
			}
			res := events.APIGatewayProxyResponse{
				StatusCode: 200,
			}
			resData := model.LinkIndustryConnectorRes{
				GlueRoleArn:  glueRoleArn,
				BucketPolicy: bucketPolicy,
			}
			jsonRes, err := json.Marshal(resData)
			if err != nil {
				tx.Log("Error while unmarshalling response %v", err)
				return builResponseError(tx, err), nil
			}
			res.Body = string(jsonRes)
			return res, nil
		}
	} else if subFunction == FN_CREATE_CONNECTOR_CRAWLER {
		//TODO: move this inside one use case
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			data, err := decodeCreateConnectorCrawlerBody(tx, req)
			if err != nil {
				return builResponseError(tx, err), nil
			}
			split := strings.Split(data.BucketPath, ":")
			bucketName := split[len(split)-1]
			err = usecase.CreateConnectorCrawler(glueClient, data.GlueRoleArn, bucketName, LAMBDA_ENV, CONNECTOR_CRAWLER_QUEUE, CONNECTOR_CRAWLER_DLQ)
			if err != nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builResponseError(tx, err), nil
			}
			var jobNames = []string{}
			crawlerName := "ucp-connector-crawler-" + LAMBDA_ENV
			jobNames, err = usecase.CreateConnectorJobTrigger(glueClient, LAMBDA_ACCOUNT_ID, LAMBDA_REGION, data.ConnectorId, crawlerName)
			if err != nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builResponseError(tx, err), nil
			}
			err = usecase.AddConnectorBucketToJobs(glueClient, bucketName, jobNames)
			if err != nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
			}
			res := events.APIGatewayProxyResponse{
				StatusCode: 200,
			}
			return res, nil
		}
	} else if subFunction == FN_MERGE_UCP_PROFILE {
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			tx.Log("Search request: %+v", rq)
			ucpRes, err = usecase.MergeUCPConfig(rq)
			if err == nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(tx, ucpRes), nil
			}
		}
		return builResponseError(tx, err), nil
	} else if subFunction == FN_LIST_UCP_INGESTION_ERROR {
		tx.Log("Selected Use Case %v", subFunction)
		if err == nil {
			tx.Log("Search request: %+v", rq)
			ucpRes, err = usecase.ListUCPIngestionError(rq, profiles)
			if err == nil {
				tx.Log("Use Case %s failed with error: %v", subFunction, err)
				return builUCPResponse(tx, ucpRes), nil
			}
		}
		return builResponseError(tx, err), nil
	}

	tx.Log("No Use Case Found for %v", method+" "+resource)
	err = errors.New("No Use Case Found for " + method + " " + resource)
	return builResponseError(tx, err), nil
}

func decodeUCPBody(tx core.Transaction, req events.APIGatewayProxyRequest) (model.UCPRequest, error) {
	wrapper := model.UCPRequest{}
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, _ := base64.StdEncoding.DecodeString(req.Body)
		decodedBody = base64Body
	}
	if err := json.Unmarshal(decodedBody, &wrapper); err != nil {
		return model.UCPRequest{}, err
	}
	tx.Log("Decoded Body %+v", wrapper)
	return wrapper, nil
}

func decodeLinkConnectorBody(tx core.Transaction, req events.APIGatewayProxyRequest) (model.LinkIndustryConnectorRq, error) {
	wrapper := model.LinkIndustryConnectorRq{}
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, _ := base64.StdEncoding.DecodeString(req.Body)
		decodedBody = base64Body
	}
	if err := json.Unmarshal(decodedBody, &wrapper); err != nil {
		return model.LinkIndustryConnectorRq{}, err
	}
	tx.Log("Decoded Body %+v", wrapper)
	return wrapper, nil
}

func decodeCreateConnectorCrawlerBody(tx core.Transaction, req events.APIGatewayProxyRequest) (model.CreateConnectorCrawlerRq, error) {
	wrapper := model.CreateConnectorCrawlerRq{}
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, _ := base64.StdEncoding.DecodeString(req.Body)
		decodedBody = base64Body
	}
	if err := json.Unmarshal(decodedBody, &wrapper); err != nil {
		return model.CreateConnectorCrawlerRq{}, err
	}
	tx.Log("Decoded Body %+v", wrapper)
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
	if res == "/ucp/connector" && meth == "GET" {
		return FN_LIST_CONNECTORS
	}
	if res == "/ucp/connector/link" && meth == "POST" {
		return FN_LINK_INDUSTRY_CONNECTOR
	}
	if res == "/ucp/connector/crawler" && meth == "POST" {
		return FN_CREATE_CONNECTOR_CRAWLER
	}
	if res == "/ucp/merge" && meth == "POST" {
		return FN_MERGE_UCP_PROFILE
	}
	if res == "/ucp/error" && meth == "GET" {
		return FN_LIST_UCP_INGESTION_ERROR
	}
	if res == "/ucp/data" && meth == "GET" {
		return FN_GET_DATA_VALIDATION_STATUS
	}
	return ""
}

func builUCPResponse(tx core.Transaction, resWrapper model.ResWrapper) events.APIGatewayProxyResponse {
	res := events.APIGatewayProxyResponse{
		StatusCode: 200,
	}
	jsonRes, err := json.Marshal(resWrapper)
	if err != nil {
		tx.Log("Error while unmarshalling response %v", err)
		return builResponseError(tx, err)
	}
	res.Body = string(jsonRes)
	return res
}

func builResponseError(tx core.Transaction, err error) events.APIGatewayProxyResponse {
	tx.Log("Response Error: %v", err)
	res := events.APIGatewayProxyResponse{
		StatusCode: 400,
	}
	json, _ := json.Marshal(model.ResWrapper{Error: core.BuildResError(err)})
	res.Body = string(json)
	return res
}

func ValidateUCPSearchRequest(wrapper model.UCPRequest) error {
	if wrapper.SearchRq.LastName == "" && wrapper.SearchRq.LoyaltyID == "" && wrapper.SearchRq.Phone == "" && wrapper.SearchRq.Email == "" {
		return errors.New("at least one search creteria within LastName, LoyaltyID, Phone and Email must be provided")
	}
	return nil
}

func ValidateUCPRetreiveRequest(wrapper model.UCPRequest) error {
	if wrapper.ID == "" {
		return errors.New("profile ID is required to retreive profile")
	}
	return nil
}

*/
