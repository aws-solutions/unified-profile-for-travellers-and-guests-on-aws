// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"encoding/json"
	"errors"
	"time"

	"fmt"
	"strings"

	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	async "tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	utils "tah/upt/source/ucp-common/src/utils/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

type DeleteError struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewDeleteError() *DeleteError {
	return &DeleteError{name: "DeleteError"}
}

func (u *DeleteError) Name() string {
	return u.name
}
func (u *DeleteError) Tx() core.Transaction {
	return *u.tx
}
func (u *DeleteError) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *DeleteError) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *DeleteError) Registry() *registry.Registry {
	return u.reg
}

func (u *DeleteError) AccessPermission() constant.AppPermission {
	return constant.ClearAllErrorsPermission
}

func (u *DeleteError) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return model.RequestWrapper{
		UcpErrorToDelete: model.UcpIngestionErrorToDelete{
			Type: ERROR_PK,
			ID:   req.PathParameters["id"],
		},
	}, nil
}

func (u *DeleteError) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	if rq.UcpErrorToDelete.ID == "*" {
		return nil
	}
	if rq.UcpErrorToDelete.ID == "" {
		return errors.New("an error ID is required for deletion")
	}
	if len(rq.UcpErrorToDelete.ID) != 61 {
		return fmt.Errorf("invalid error ID format should be 61 char long and not %d", len(rq.UcpErrorToDelete.ID))
	}
	if !strings.HasPrefix(rq.UcpErrorToDelete.ID, "error_") {
		return errors.New("invalid error ID format: should start with error_")
	}
	if utils.IsUUID(rq.UcpErrorToDelete.ID[23:]) {
		return errors.New("invalid error ID format: last 38 char should be a uuid")
	}
	return nil
}

func (u *DeleteError) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	if req.UcpErrorToDelete.ID == "*" {
		u.tx.Info("Deleting all errors")
		return deleteErrorsAsync(u, req)
	}

	u.tx.Info("Deleting single error: %s", req.UcpErrorToDelete.ID)
	return deleteError(u, req)
}

func (u *DeleteError) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

// Delete a single error by ID. This request is completed synchronously and the result is returned.
func deleteError(u *DeleteError, req model.RequestWrapper) (model.ResponseWrapper, error) {
	_, err := u.reg.ErrorDB.Delete(req.UcpErrorToDelete)
	return model.ResponseWrapper{}, err
}

// Delete all errors in the DynamoDB table. This request is completed asynchronously and the async event details are returned.
// The client must poll with the async event details to determine when the request is finished.
func deleteErrorsAsync(u *DeleteError, req model.RequestWrapper) (model.ResponseWrapper, error) {
	body := uc.EmptyTableBody{
		TableName: u.reg.ErrorDB.TableName,
	}
	eventId := uuid.NewString()
	payload := async.AsyncInvokePayload{
		EventID:       eventId,
		Usecase:       async.USECASE_EMPTY_TABLE,
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
		Usecase:     async.USECASE_EMPTY_TABLE,
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
