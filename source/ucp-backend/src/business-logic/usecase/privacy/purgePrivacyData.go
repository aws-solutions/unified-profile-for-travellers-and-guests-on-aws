// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package privacy

import (
	"encoding/json"
	"errors"
	"fmt"
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	"tah/upt/source/ucp-common/src/utils/utils"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

type CreatePrivacyDataPurge struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

// Static interface check
var _ registry.Usecase = (*CreatePrivacyDataPurge)(nil)

func NewCreatePrivacyDataPurge() *CreatePrivacyDataPurge {
	return &CreatePrivacyDataPurge{
		name: "PurgePrivacyData",
	}
}

// CreateRequest implements registry.Usecase.
func (u *CreatePrivacyDataPurge) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	username := cognito.ParseUserFromLambdaRq(req)
	if username == "" {
		u.tx.Error("[%s] Error getting user permission: username is empty", u.name)
		return model.RequestWrapper{}, errors.New("username is empty")
	}
	wrapper, err := registry.CreateRequest(u, req)
	if err != nil {
		u.tx.Error("[%s] Error creating request wrapper: %v", u.name, err)
		return model.RequestWrapper{}, err
	}

	wrapper.PrivacyPurge.AgentCognitoId = username
	return wrapper, nil
}

// CreateResponse implements registry.Usecase.
func (u *CreatePrivacyDataPurge) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

// Name implements registry.Usecase.
func (u *CreatePrivacyDataPurge) Name() string {
	return u.name
}

// Registry implements registry.Usecase.
func (u *CreatePrivacyDataPurge) Registry() *registry.Registry {
	return u.reg
}

// Run implements registry.Usecase.
func (u *CreatePrivacyDataPurge) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	u.tx.Info("[%s] Running Privacy Purge for: %s", u.name, req.PrivacyPurge.ConnectIds)

	// Save privacy search result with status
	records := []db.DynamoRecord{}
	for _, connectId := range req.PrivacyPurge.ConnectIds {
		records = append(records, db.DynamoRecord{
			Pk: u.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR],
			Sk: connectId,
		})
	}

	searchResults := []uc.PrivacySearchResult{}
	err := u.reg.PrivacyDB.GetMany(records, &searchResults)
	if err != nil {
		u.tx.Error("[%s] Error getting search results: %v", u.name, err)
		return model.ResponseWrapper{}, err
	}

	var errors []error
	for _, searchResult := range searchResults {
		for _, locationType := range searchResult.LocationResults {
			for _, location := range locationType {
				err = u.reg.PrivacyDB.UpdateItems(
					u.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR],
					searchResult.ConnectId+"|"+location,
					map[string]interface{}{
						utils.GetJSONTag(uc.PrivacySearchResult{}, "Status"): string(uc.PRIVACY_STATUS_PURGE_INVOKED),
						utils.GetJSONTag(uc.PrivacySearchResult{}, "TxID"):   u.tx.TransactionID,
					},
				)
				if err != nil {
					u.tx.Error("[%s] Error updating privacy search result: %v", u.name, err)
					errors = append(errors, err)
				}
			}
		}
	}
	if len(errors) > 0 {
		return model.ResponseWrapper{}, core.ErrorsToError(errors)
	}

	body := uc.PurgeProfileDataBody{
		ConnectIds:     req.PrivacyPurge.ConnectIds,
		DomainName:     u.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR],
		AgentCognitoId: req.PrivacyPurge.AgentCognitoId,
	}

	eventId := uuid.NewString()
	payload := async.AsyncInvokePayload{
		EventID:       eventId,
		Usecase:       async.USECASE_PURGE_PROFILE_DATA,
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
		Usecase:     async.USECASE_PURGE_PROFILE_DATA,
		Status:      async.EVENT_STATUS_INVOKED,
		LastUpdated: time.Now(),
	}
	_, err = u.reg.ConfigDB.Save(asyncEvent)
	if err != nil {
		u.tx.Error("[%s] Error saving async status: %v", u.name, err)
		return model.ResponseWrapper{}, err
	}

	res := model.ResponseWrapper{
		AsyncEvent: &asyncEvent,
	}

	return res, nil
}

// SetRegistry implements registry.Usecase.
func (u *CreatePrivacyDataPurge) SetRegistry(r *registry.Registry) {
	u.reg = r
}

// SetTx implements registry.Usecase.
func (u *CreatePrivacyDataPurge) SetTx(tx *core.Transaction) {
	u.tx = tx
}

// Tx implements registry.Usecase.
func (u *CreatePrivacyDataPurge) Tx() core.Transaction {
	return *u.tx
}

func (u *CreatePrivacyDataPurge) AccessPermission() admin.AppPermission {
	return admin.PrivacyDataPurgePermission
}

// ValidateRequest implements registry.Usecase.
func (u *CreatePrivacyDataPurge) ValidateRequest(req model.RequestWrapper) error {
	u.tx.Debug("[%s] Validating Request", u.name)
	if len(req.PrivacyPurge.ConnectIds) == 0 {
		return fmt.Errorf("[%s] no connect ids found in request", u.name)
	}
	return nil
}
