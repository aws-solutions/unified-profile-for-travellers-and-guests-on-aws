// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"encoding/json"
	"errors"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	async "tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

type UnmergeProfile struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewUnmergeProfile() *UnmergeProfile {
	return &UnmergeProfile{name: "UnmergeProfile"}
}

func (u *UnmergeProfile) Name() string {
	return u.name
}

func (u *UnmergeProfile) Tx() core.Transaction {
	return *u.tx
}

func (u *UnmergeProfile) SetTx(tx *core.Transaction) {
	u.tx = tx
}

func (u *UnmergeProfile) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}

func (u *UnmergeProfile) Registry() *registry.Registry {
	return u.reg
}

func (u *UnmergeProfile) AccessPermission() constant.AppPermission {
	return constant.UnmergeProfilePermission
}

func (u *UnmergeProfile) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("Validating request for UnmergeProfile")

	if u.reg.DataAccessPermission != constant.DATA_ACCESS_PERMISSION_ALL {
		return errors.New("[ValidateRequest] ony administrator can unmerge profiles")
	}
	if rq.UnmergeRq.MergedIntoConnectID == "" {
		return errors.New("[ValidateRequest] profile to unmerge from is not provided")
	}
	if rq.UnmergeRq.InteractionToUnmerge == "" && rq.UnmergeRq.ToUnmergeConnectID == "" {
		return errors.New("[ValidateRequest] insufficient information to perform unmerge")
	}

	return nil
}

func (u *UnmergeProfile) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *UnmergeProfile) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

func (u *UnmergeProfile) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	domain := u.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR]
	eventId := uuid.NewString()

	payload := u.buildAsyncPayload(domain, eventId, req.UnmergeRq)
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
		Usecase:     async.USECASE_UNMERGE_PROFILES,
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

func (u *UnmergeProfile) buildAsyncPayload(domain string, eventId string, rq commonModel.UnmergeRq) async.AsyncInvokePayload {
	body := uc.UnmergeProfilesBody{
		Rq:     rq,
		Domain: domain,
	}
	payload := async.AsyncInvokePayload{
		EventID:       eventId,
		Usecase:       async.USECASE_UNMERGE_PROFILES,
		TransactionID: u.tx.TransactionID,
		Body:          body,
	}
	return payload
}
