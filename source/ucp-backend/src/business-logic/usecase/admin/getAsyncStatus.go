// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"errors"
	"strings"
	"time"

	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/model/async"
	"tah/upt/source/ucp-common/src/utils/utils"

	"github.com/aws/aws-lambda-go/events"
)

type GetAsyncStatus struct {
	name    string
	tx      *core.Transaction
	reg     *registry.Registry
	eventId string
	usecase string
}

func NewGetAsyncStatus() *GetAsyncStatus {
	return &GetAsyncStatus{name: "GetAsyncStatus"}
}

func (u *GetAsyncStatus) Name() string {
	return u.name
}

func (u *GetAsyncStatus) Tx() core.Transaction {
	return *u.tx
}

func (u *GetAsyncStatus) SetTx(tx *core.Transaction) {
	u.tx = tx
}

func (u *GetAsyncStatus) Registry() *registry.Registry {
	return u.reg
}

func (u *GetAsyncStatus) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}

func (u *GetAsyncStatus) AccessPermission() constant.AppPermission {
	return constant.PublicAccessPermission
}

func (u *GetAsyncStatus) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	u.tx.Debug("Creating request wrapper")
	u.eventId = req.QueryStringParameters["id"]
	u.usecase = req.QueryStringParameters["usecase"]
	return registry.CreateRequest(u, req)
}

func (u *GetAsyncStatus) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	if u.eventId == "" || u.usecase == "" {
		return errors.New("invalid request: missing event id or use case")
	}
	if !utils.ContainsString(async.ASYNC_USECASE_LIST, u.usecase) {
		return errors.New("invalid usecase. supported use cases are: " + strings.Join(async.ASYNC_USECASE_LIST, ","))
	}

	return nil
}

func (u *GetAsyncStatus) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	u.tx.Info("[%s] Running usecase %s for event %s", u.name, u.usecase, u.eventId)
	asyncEvent := async.AsyncEvent{}
	err := u.reg.ConfigDB.Get(u.usecase, u.eventId, &asyncEvent)

	if err != nil {
		return model.ResponseWrapper{}, err
	}

	// Check if the event is taking longer than expected, and if so, time out the request
	if asyncEvent.Status != async.EVENT_STATUS_SUCCESS && asyncEvent.Status != async.EVENT_STATUS_FAILED {
		now := time.Now()
		timeout := 5 * time.Minute
		if now.Sub(asyncEvent.LastUpdated) > timeout {
			u.tx.Error("Event timed out")
			asyncEvent.Status = async.EVENT_STATUS_FAILED
			asyncEvent.LastUpdated = now
			_, err = u.reg.ConfigDB.Save(asyncEvent)
			if err != nil {
				u.tx.Error("Error timing out async event status")
			}
		}
	}

	res := model.ResponseWrapper{
		AsyncEvent: &asyncEvent,
	}
	return res, nil
}

func (u *GetAsyncStatus) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
