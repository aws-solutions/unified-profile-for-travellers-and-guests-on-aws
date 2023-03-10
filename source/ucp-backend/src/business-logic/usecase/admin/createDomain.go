package admin

import (
	"errors"
	"tah/core/core"
	"tah/core/customerprofiles"
	accpmappings "tah/ucp/src/business-logic/model/accp-mappings"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

// Key Names for Business Objects
const HOTEL_BOOKING string = "hotel_booking"
const HOTEL_STAY_REVENUE string = "hotel_stay_revenue"
const CLICKSTREAM string = "clickstream"
const AIR_BOOKING string = "air_booking"
const GUEST_PROFILE string = "guest_profile"
const PASSENGER_PROFILE string = "passenger_profile"

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

	businessMap := map[string]func() []customerprofiles.FieldMapping{
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
		err = u.reg.Accp.PutIntegration(keyBusiness, accpSourceBucket, businessMap[keyBusiness]())
		if err != nil {
			u.tx.Log("Error creating integration %s", err)
			return model.ResponseWrapper{}, err
		}
	}
	return model.ResponseWrapper{}, err
}

func (u *CreateDomain) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
