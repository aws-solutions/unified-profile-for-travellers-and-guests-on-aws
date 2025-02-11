// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package privacy

import (
	"encoding/json"
	"fmt"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

type CreatePrivacySearchData struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

// Static interface check
var _ registry.Usecase = &CreatePrivacySearchData{}

func NewCreatePrivacySearch() *CreatePrivacySearchData {
	return &CreatePrivacySearchData{
		name: "SearchData",
	}
}

func (u *CreatePrivacySearchData) Name() string {
	return u.name
}

func (u *CreatePrivacySearchData) Tx() core.Transaction {
	return *u.tx
}

func (u *CreatePrivacySearchData) SetTx(tx *core.Transaction) {
	u.tx = tx
}

func (u *CreatePrivacySearchData) SetRegistry(r *registry.Registry) {
	u.reg = r
}

func (u *CreatePrivacySearchData) Registry() *registry.Registry {
	return u.reg
}

func (u *CreatePrivacySearchData) AccessPermission() admin.AppPermission {
	return admin.CreatePrivacySearchPermission
}

func (u *CreatePrivacySearchData) ValidateRequest(req model.RequestWrapper) error {
	u.tx.Debug("[%s] Validating Request", u.name)

	if len(req.CreatePrivacySearchRq) == 0 {
		return fmt.Errorf("[%s] no connect ids found in request", u.name)
	}

	for _, searchRq := range req.CreatePrivacySearchRq {
		_, err := uuid.Parse(searchRq.ConnectId)
		if err != nil {
			return fmt.Errorf("[%s] invalid connect ID %s found", searchRq.ConnectId, u.name)
		}
	}
	return nil
}
func (u *CreatePrivacySearchData) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *CreatePrivacySearchData) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

func (u *CreatePrivacySearchData) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	connectIds := []string{}
	for _, rq := range req.CreatePrivacySearchRq {
		connectIds = append(connectIds, rq.ConnectId)
	}
	u.tx.Info("[%s] Running Privacy Search for: %v", u.name, connectIds)

	body := uc.CreatePrivacySearchBody{
		ConnectIds: connectIds,
		DomainName: u.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR],
	}

	eventId := uuid.NewString()
	payload := async.AsyncInvokePayload{
		EventID:       eventId,
		Usecase:       async.USECASE_CREATE_PRIVACY_SEARCH,
		TransactionID: u.tx.TransactionID,
		Body:          body,
	}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		u.tx.Error("[%s] Error marshalling async payload: %v", u.name, err)
		return model.ResponseWrapper{}, err
	}

	_, err = u.reg.AsyncLambda.InvokeAsync(payloadJson)
	if err != nil {
		u.tx.Error("[%s] Error invoking async lambda: %v", u.name, err)
		return model.ResponseWrapper{}, err
	}

	asyncEvent := async.AsyncEvent{
		EventID:     eventId,
		Usecase:     async.USECASE_CREATE_PRIVACY_SEARCH,
		Status:      async.EVENT_STATUS_INVOKED,
		LastUpdated: time.Now(),
	}
	_, err = u.reg.ConfigDB.Save(asyncEvent)
	if err != nil {
		u.tx.Error("[%s] Error saving async event: %v", u.name, err)
		return model.ResponseWrapper{}, err
	}

	// Save search result with status
	searchResults := []uc.PrivacySearchResult{}
	for _, connectId := range connectIds {
		searchResults = append(searchResults, uc.PrivacySearchResult{
			DomainName: body.DomainName,
			ConnectId:  connectId,
			Status:     uc.PRIVACY_STATUS_SEARCH_INVOKED,
			SearchDate: time.Now(),
			TxID:       u.tx.TransactionID,
		})
	}
	err = u.reg.PrivacyDB.SaveMany(searchResults)
	if err != nil {
		u.tx.Error("[%s] Error saving search results: %v", u.name, err)
		return model.ResponseWrapper{}, err
	}

	res := model.ResponseWrapper{
		AsyncEvent: &asyncEvent,
	}

	return res, nil
}
