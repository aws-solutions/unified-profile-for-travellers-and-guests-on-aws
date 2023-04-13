package admin

import (
	"errors"
	"sync"
	"tah/core/core"
	"tah/core/customerprofiles"
	common "tah/ucp-common/src/constant/admin"
	commonModel "tah/ucp-common/src/model/admin"
	commonServices "tah/ucp-common/src/services/admin"
	accpmappings "tah/ucp/src/business-logic/model/accp-mappings"
	assetsSchema "tah/ucp/src/business-logic/model/assetsSchema"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type CreateDomain struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewCreateDomain() *CreateDomain {
	return &CreateDomain{name: "CreateDomain"}
}

func (u *CreateDomain) Name() string {
	return u.name
}
func (u *CreateDomain) Tx() core.Transaction {
	return *u.tx
}
func (u *CreateDomain) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *CreateDomain) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *CreateDomain) Registry() *registry.Registry {
	return u.reg
}

func (u *CreateDomain) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	u.tx.Log("Creating request wrapper")
	return registry.CreateRequest(u, req)
}

func (u *CreateDomain) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *CreateDomain) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	kmsArn := u.reg.Env["KMS_KEY_PROFILE_DOMAIN"]
	env := u.reg.Env["LAMBDA_ENV"]
	accpSourceBucket := u.reg.Env["CONNECT_PROFILE_SOURCE_BUCKET"]

	if kmsArn == "" || env == "" || accpSourceBucket == "" {
		return model.ResponseWrapper{}, errors.New("Missing Registry Environment (KMS_KEY_PROFILE_DOMAIN,LAMBDA_ENV,CONNECT_PROFILE_SOURCE_BUCKET)")
	}

	err := u.reg.Accp.CreateDomain(req.Domain.Name, true, kmsArn, map[string]string{DOMAIN_TAG_ENV_NAME: env})
	if err != nil {
		return model.ResponseWrapper{}, err
	}

	businessMap := map[string]func() customerprofiles.FieldMappings{
		ACCP_SUB_FOLDER_AIR_BOOKING:        accpmappings.BuildAirBookingMapping,
		ACCP_SUB_FOLDER_EMAIL_HISTORY:      accpmappings.BuildEmailHistoryMapping,
		ACCP_SUB_FOLDER_PHONE_HISTORY:      accpmappings.BuildPhoneHistoryMapping,
		ACCP_SUB_FOLDER_AIR_LOYALTY:        accpmappings.BuildAirLoyaltyMapping,
		ACCP_SUB_FOLDER_CLICKSTREAM:        accpmappings.BuildClickstreamMapping,
		ACCP_SUB_FOLDER_GUEST_PROFILE:      accpmappings.BuildGuestProfileMapping,
		ACCP_SUB_FOLDER_HOTEL_LOYALTY:      accpmappings.BuildHotelLoyaltyMapping,
		ACCP_SUB_FOLDER_HOTEL_BOOKING:      accpmappings.BuildHotelBookingMapping,
		ACCP_SUB_FOLDER_PAX_PROFILE:        accpmappings.BuildPassengerProfileMapping,
		ACCP_SUB_FOLDER_HOTEL_STAY_MAPPING: accpmappings.BuildHotelStayMapping,
	}

	bizObjectBuckets := map[string]string{
		common.BIZ_OBJECT_HOTEL_BOOKING: u.reg.Env["S3_HOTEL_BOOKING"],
		common.BIZ_OBJECT_AIR_BOOKING:   u.reg.Env["S3_AIR_BOOKING"],
		common.BIZ_OBJECT_GUEST_PROFILE: u.reg.Env["S3_GUEST_PROFILE"],
		common.BIZ_OBJECT_PAX_PROFILE:   u.reg.Env["S3_PAX_PROFILE"],
		common.BIZ_OBJECT_CLICKSTREAM:   u.reg.Env["S3_STAY_REVENUE"],
		common.BIZ_OBJECT_STAY_REVENUE:  u.reg.Env["S3_CLICKSTREAM"],
	}

	for keyBusiness := range businessMap {
		err = u.reg.Accp.CreateMapping(keyBusiness,
			"Primary Mapping for the "+keyBusiness+" object", businessMap[keyBusiness]())
		if err != nil {
			u.tx.Log("[CreateUcpDomain] Error creating Mapping: %s. deleting domain", err)
			err2 := u.reg.Accp.DeleteDomain()
			if err2 != nil {
				u.tx.Log("[CreateUcpDomain][warning] Error cleaning up domain after failed mapping creation %v", err2)
			}
			return model.ResponseWrapper{}, err
		}
		integrationName := keyBusiness + "_" + req.Domain.Name
		err = u.reg.Accp.PutIntegration(integrationName, keyBusiness, accpSourceBucket, businessMap[keyBusiness]())
		if err != nil {
			u.tx.Log("Error creating integration %s", err)
			return model.ResponseWrapper{}, err
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(common.BUSINESS_OBJECTS))
	output := make(chan error, 100)
	for _, bizObject := range common.BUSINESS_OBJECTS {
		go func(bizObject commonModel.BusinessObject) {
			schema, err := assetsSchema.LoadSchema(bizObject)
			if err != nil {
				u.tx.Log("[CreateUcpDomain] Error loading schema: %v", err)
				output <- err
			}
			tableName := commonServices.BuildTableName(env, bizObject, req.Domain.Name)

			err2 := u.reg.Glue.CreateTable(tableName, bizObjectBuckets[bizObject.Name], map[string]string{"year": "int", "month": "int", "day": "int"}, schema)
			if err2 != nil {
				u.tx.Log("[CreateUcpDomain] Error creating table: %v", err2)
				output <- err2
			}
			defer wg.Done()
		}(bizObject)
	}
	wg.Wait()
	if len(output) > 0 {
		return model.ResponseWrapper{}, errors.New("Error occured during table creation")
	}

	return model.ResponseWrapper{}, err
}

func (u *CreateDomain) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
