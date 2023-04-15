package admin

import (
	"errors"
	"math/rand"
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
	"time"

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
	// Set the seed for the random number generator
	rand.Seed(time.Now().UnixNano())
	kmsArn := u.reg.Env["KMS_KEY_PROFILE_DOMAIN"]
	dlq := u.reg.Env["ACCP_DOMAIN_DLQ"]

	env := u.reg.Env["LAMBDA_ENV"]
	accpSourceBucket := u.reg.Env["CONNECT_PROFILE_SOURCE_BUCKET"]

	if kmsArn == "" || env == "" || accpSourceBucket == "" || dlq == "" {
		return model.ResponseWrapper{}, errors.New("Missing Registry Environment (KMS_KEY_PROFILE_DOMAIN,LAMBDA_ENV,CONNECT_PROFILE_SOURCE_BUCKET,ACCP_DOMAIN_DLQ)")
	}

	u.tx.Log("Creating domain")
	err := u.reg.Accp.CreateDomainWithQueue(req.Domain.Name, true, kmsArn, map[string]string{DOMAIN_TAG_ENV_NAME: env, "aws_solution": "SO0244"}, dlq)
	if err != nil {
		return model.ResponseWrapper{}, err
	}

	accpMappings := map[string]func() customerprofiles.FieldMappings{
		common.ACCP_RECORD_AIR_BOOKING:        accpmappings.BuildAirBookingMapping,
		common.ACCP_RECORD_EMAIL_HISTORY:      accpmappings.BuildEmailHistoryMapping,
		common.ACCP_RECORD_PHONE_HISTORY:      accpmappings.BuildPhoneHistoryMapping,
		common.ACCP_RECORD_AIR_LOYALTY:        accpmappings.BuildAirLoyaltyMapping,
		common.ACCP_RECORD_CLICKSTREAM:        accpmappings.BuildClickstreamMapping,
		common.ACCP_RECORD_GUEST_PROFILE:      accpmappings.BuildGuestProfileMapping,
		common.ACCP_RECORD_HOTEL_LOYALTY:      accpmappings.BuildHotelLoyaltyMapping,
		common.ACCP_RECORD_HOTEL_BOOKING:      accpmappings.BuildHotelBookingMapping,
		common.ACCP_RECORD_PAX_PROFILE:        accpmappings.BuildPassengerProfileMapping,
		common.ACCP_RECORD_HOTEL_STAY_MAPPING: accpmappings.BuildHotelStayMapping,
	}

	bizObjectBuckets := map[string]string{
		common.BIZ_OBJECT_HOTEL_BOOKING: u.reg.Env["S3_HOTEL_BOOKING"],
		common.BIZ_OBJECT_AIR_BOOKING:   u.reg.Env["S3_AIR_BOOKING"],
		common.BIZ_OBJECT_GUEST_PROFILE: u.reg.Env["S3_GUEST_PROFILE"],
		common.BIZ_OBJECT_PAX_PROFILE:   u.reg.Env["S3_PAX_PROFILE"],
		common.BIZ_OBJECT_CLICKSTREAM:   u.reg.Env["S3_STAY_REVENUE"],
		common.BIZ_OBJECT_STAY_REVENUE:  u.reg.Env["S3_CLICKSTREAM"],
	}

	i := 0
	var lastErr error
	var wg sync.WaitGroup
	wg.Add(len(common.ACCP_RECORDS))
	wg.Add(len(common.BUSINESS_OBJECTS))
	for _, businessObject := range common.ACCP_RECORDS {
		go func(index int, accpRec commonModel.AccpRecord) {
			accpRecName := accpRec.Name
			u.tx.Log("[CreateUcpDomain] Creating mapping for %s", accpRecName)
			err = u.reg.Accp.CreateMapping(accpRecName,
				"Primary Mapping for the "+accpRecName+" object", accpMappings[accpRecName]())
			if err != nil {
				u.tx.Log("[CreateUcpDomain] Error creating Mapping: %s. deleting domain", err)
				lastErr = err
				err2 := u.reg.Accp.DeleteDomain()
				if err2 != nil {
					u.tx.Log("[CreateUcpDomain][warning] Error cleaning up domain after failed mapping creation %v", err2)
				}
				wg.Done()
				return
			}

			u.tx.Log("[CreateUcpDomain] Creating integration for %s", accpRecName)
			integrationName := accpRecName + "_" + req.Domain.Name
			//we create each flow with a 3 minutes delay compared to he previous to avoid concurtency issues at the ACCP level
			startTime := time.Now().Add(time.Duration(index*3) * time.Minute)
			//we introduce 100 ms delay to avoid
			wait := (1000 * time.Millisecond) * time.Duration(index)
			time.Sleep(wait)
			//each integration creates an AppFlow flow that targets a S3 bucket under the folder <dominaName>/<objectName>
			prefix := req.Domain.Name + "/" + accpRecName
			err = u.reg.Accp.PutIntegration(integrationName, accpRecName, accpSourceBucket, prefix, accpMappings[accpRecName](), startTime)
			if err != nil {
				u.tx.Log("Error creating integration %s, retrying after wait", err)
				time.Sleep((1000 * time.Millisecond))
				//we update the flow name to avoid conflict during retry
				integrationName = accpRecName + "_" + req.Domain.Name + "_1"
				startTime = time.Now().Add(time.Duration(index) * time.Minute)
				err = u.reg.Accp.PutIntegration(integrationName, accpRecName, accpSourceBucket, prefix, accpMappings[accpRecName](), startTime)
				if err != nil {
					u.tx.Log("Error after retry %s", err)
					lastErr = err
				}
			}
			wg.Done()
		}(i, businessObject)
		i++
	}
	j := 0
	for _, businessObject := range common.BUSINESS_OBJECTS {
		go func(index int, bizObject commonModel.BusinessObject) {
			u.tx.Log("[CreateUcpDomain] Creating table for %s", bizObject.Name)
			schema, err := assetsSchema.LoadSchema(bizObject)
			if err != nil {
				u.tx.Log("[CreateUcpDomain] Error loading schema: %v", err)
				lastErr = err
			}
			tableName := commonServices.BuildTableName(env, bizObject, req.Domain.Name)

			err2 := u.reg.Glue.CreateTable(tableName, bizObjectBuckets[bizObject.Name], map[string]string{"year": "int", "month": "int", "day": "int"}, schema)
			if err2 != nil {
				u.tx.Log("[CreateUcpDomain] Error creating table: %v", err2)
				lastErr = err2
			}

			wg.Done()
		}(j, businessObject)
		j++
	}
	wg.Wait()
	return model.ResponseWrapper{}, lastErr
}

func (u *CreateDomain) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
