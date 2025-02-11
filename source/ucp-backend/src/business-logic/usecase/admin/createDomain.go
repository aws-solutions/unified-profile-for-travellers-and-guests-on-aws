// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	limits "tah/upt/source/ucp-common/src/constant/limits"
	"tah/upt/source/ucp-common/src/feature"
	async "tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	commonServices "tah/upt/source/ucp-common/src/services/admin"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
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
	u.tx.Debug("Creating request wrapper")
	return registry.CreateRequest(u, req)
}

func (u *CreateDomain) AccessPermission() constant.AppPermission {
	return constant.CreateDomainPermission
}

func (u *CreateDomain) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())

	// Validate name requirements
	err := customerprofileslcs.ValidateDomainName(rq.Domain.Name)
	if err != nil {
		return err
	}

	// Validate against existing domains
	existingDomains, err := commonServices.SearchDomain(*u.tx, u.reg.Accp, u.reg.Env["LAMBDA_ENV"])
	if err != nil {
		u.tx.Error("Error while searching existing domains %s", err)
		return errors.New("unable to validate request")
	}
	if len(existingDomains) >= limits.UPT_LIMIT_MAX_DOMAINS {
		return fmt.Errorf("limit of %d domains reached", limits.UPT_LIMIT_MAX_DOMAINS)
	}
	for _, domain := range existingDomains {
		if domain.Name == rq.Domain.Name {
			return fmt.Errorf("domain %s already exists", rq.Domain.Name)
		}
	}

	return nil
}

func (u *CreateDomain) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	body := uc.CreateDomainBody{
		KmsArn:                u.reg.Env["KMS_KEY_PROFILE_DOMAIN"],
		AccpDlq:               u.reg.Env["ACCP_DOMAIN_DLQ"],
		AccpDestinationStream: u.reg.Env["ACCP_DESTINATION_STREAM"],
		Env:                   u.reg.Env["LAMBDA_ENV"],
		MatchBucketName:       u.reg.Env["MATCH_BUCKET_NAME"],
		AccpSourceBucketName:  u.reg.Env["CONNECT_PROFILE_SOURCE_BUCKET"],
		DomainName:            req.Domain.Name,
		HotelBookingBucket:    u.reg.Env["S3_HOTEL_BOOKING"],
		AirBookingBucket:      u.reg.Env["S3_AIR_BOOKING"],
		GuestProfileBucket:    u.reg.Env["S3_GUEST_PROFILE"],
		PaxProfileBucket:      u.reg.Env["S3_PAX_PROFILE"],
		StayRevenueBucket:     u.reg.Env["S3_STAY_REVENUE"],
		ClickstreamBucket:     u.reg.Env["S3_CLICKSTREAM"],
		CSIBucket:             u.reg.Env["S3_CSI"],
		GlueSchemaPath:        u.reg.Env["GLUE_SCHEMA_PATH"],
		RequestedVersion:      feature.LATEST_FEATURE_SET_VERSION,
	}

	eventId := uuid.NewString()
	payload := async.AsyncInvokePayload{
		EventID:       eventId,
		Usecase:       async.USECASE_CREATE_DOMAIN,
		TransactionID: u.tx.TransactionID,
		Body:          body,
	}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		u.tx.Error("[%s] Error marshalling async payload: %v", u.Name(), err)
		return model.ResponseWrapper{}, err
	}

	_, err = u.reg.AsyncLambda.InvokeAsync(payloadJson)
	if err != nil {
		u.tx.Error("[%s] Error invoking async lambda: %v", u.Name(), err)
		return model.ResponseWrapper{}, err
	}

	asyncEvent := async.AsyncEvent{
		EventID:     eventId,
		Usecase:     async.USECASE_CREATE_DOMAIN,
		Status:      async.EVENT_STATUS_INVOKED,
		LastUpdated: time.Now(),
	}
	_, err = u.reg.ConfigDB.Save(asyncEvent)
	if err != nil {
		return model.ResponseWrapper{}, err
	}

	res := model.ResponseWrapper{
		AsyncEvent: &asyncEvent,
	}
	return res, nil
}

func (u *CreateDomain) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
