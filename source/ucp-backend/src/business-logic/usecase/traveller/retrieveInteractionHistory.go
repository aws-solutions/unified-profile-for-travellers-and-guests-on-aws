// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"errors"
	"fmt"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	utils "tah/upt/source/ucp-common/src/utils/utils"

	"github.com/aws/aws-lambda-go/events"
)

type RetrieveInteractionHistory struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewRetrieveInteractionHistory() *RetrieveInteractionHistory {
	return &RetrieveInteractionHistory{name: "RetrieveInteractionHistory"}
}

func (u *RetrieveInteractionHistory) Name() string {
	return u.name
}

func (u *RetrieveInteractionHistory) Tx() core.Transaction {
	return *u.tx
}

func (u *RetrieveInteractionHistory) SetTx(tx *core.Transaction) {
	u.tx = tx
}

func (u *RetrieveInteractionHistory) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}

func (u *RetrieveInteractionHistory) Registry() *registry.Registry {
	return u.reg
}

func (u *RetrieveInteractionHistory) AccessPermission() constant.AppPermission {
	return constant.PublicAccessPermission
}

func (u *RetrieveInteractionHistory) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return model.RequestWrapper{
		GetInteractionHistory: model.GetInteractionHistoryRq{
			InteractionId: req.PathParameters["id"],
			ObjectType:    req.QueryStringParameters["objectType"],
			ConnectId:     req.QueryStringParameters["connectId"],
			DomainName:    u.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR],
		},
	}, nil
}

func (u *RetrieveInteractionHistory) ValidateRequest(rq model.RequestWrapper) error {
	isValidObjectType := map[string]bool{}
	for _, obj := range constant.ACCP_RECORDS {
		isValidObjectType[obj.Name] = true
	}
	invalidChars := []string{"\\", "/", "\b", "\f", "\n", "\r", "\t", "{", "}", "<", ">", "^", "="}
	u.tx.Debug("[%v] Validating request", u.Name())
	if rq.GetInteractionHistory.InteractionId == "" {
		return errors.New("interaction id needed")
	}
	if len(rq.GetInteractionHistory.InteractionId) > 255 {
		return errors.New("interaction Id should be less than 255 characters")
	}
	if utils.ValidateString(rq.GetInteractionHistory.InteractionId, invalidChars) != nil {
		return errors.New("interaction Id contains invalid characters")
	}
	if !isValidObjectType[rq.GetInteractionHistory.ObjectType] {
		return fmt.Errorf("invalid object type %s, should be within %v", rq.GetInteractionHistory.ObjectType, constant.ACCP_RECORDS)
	}
	if rq.GetInteractionHistory.DomainName == "" {
		return errors.New("no domain name in the context if the request")
	}
	if rq.GetInteractionHistory.ConnectId == "" {
		return errors.New("connectId needed")
	}
	if len(rq.GetInteractionHistory.ConnectId) > 255 {
		return errors.New("connectId should be less than 255 characters")
	}
	if utils.ValidateString(rq.GetInteractionHistory.ConnectId, invalidChars) != nil {
		return errors.New("connectId contains invalid characters")
	}
	return nil
}

func (u *RetrieveInteractionHistory) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	interactionHistory, err := u.reg.Accp.BuildInteractionHistory(req.GetInteractionHistory.DomainName, req.GetInteractionHistory.InteractionId, req.GetInteractionHistory.ObjectType, req.GetInteractionHistory.ConnectId)
	if err != nil {
		u.tx.Error("[RetrieveInteraction RUN] Error  %s", err)
		return model.ResponseWrapper{}, err
	}

	return model.ResponseWrapper{InteractionHistory: &interactionHistory}, err
}

func (u *RetrieveInteractionHistory) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
