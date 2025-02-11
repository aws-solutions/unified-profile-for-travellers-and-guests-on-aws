// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"encoding/json"
	"errors"
	"time"

	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	async "tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

type DeleteDomain struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewDeleteDomain() *DeleteDomain {
	return &DeleteDomain{name: "DeleteDomain"}
}

func (u *DeleteDomain) Name() string {
	return u.name
}
func (u *DeleteDomain) Tx() core.Transaction {
	return *u.tx
}
func (u *DeleteDomain) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *DeleteDomain) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *DeleteDomain) Registry() *registry.Registry {
	return u.reg
}

func (u *DeleteDomain) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	rq := model.RequestWrapper{
		Domain: model.Domain{Name: req.PathParameters["id"]},
	}
	return rq, nil
}

func (u *DeleteDomain) AccessPermission() constant.AppPermission {
	return constant.DeleteDomainPermission
}

func (u *DeleteDomain) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	if rq.Domain.Name == "" {
		return errors.New("domain name cannot be empty")
	}
	return customerprofiles.ValidateDomainName(rq.Domain.Name)
}

func (u *DeleteDomain) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	body := uc.DeleteDomainBody{
		Env:        u.reg.Env["LAMBDA_ENV"],
		DomainName: req.Domain.Name,
	}
	eventId := uuid.NewString()
	payload := async.AsyncInvokePayload{
		EventID:       eventId,
		Usecase:       async.USECASE_DELETE_DOMAIN,
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
		Usecase:     async.USECASE_DELETE_DOMAIN,
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

func (u *DeleteDomain) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
