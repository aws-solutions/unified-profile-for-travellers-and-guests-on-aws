// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"errors"

	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"

	"github.com/aws/aws-lambda-go/events"
)

type StartFlows struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewStartFlows() *StartFlows {
	return &StartFlows{name: "StartFlows"}
}

func (u *StartFlows) Name() string {
	return u.name
}
func (u *StartFlows) Tx() core.Transaction {
	return *u.tx
}
func (u *StartFlows) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *StartFlows) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *StartFlows) Registry() *registry.Registry {
	return u.reg
}

func (u *StartFlows) AccessPermission() constant.AppPermission {
	return constant.PublicAccessPermission
}

func (u *StartFlows) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *StartFlows) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	if len(rq.Domain.Integrations) == 0 {
		return errors.New("integration list cannot be empty")
	}
	return nil
}

func (u *StartFlows) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	u.tx.Info("Starting flow")
	for _, integration := range req.Domain.Integrations {
		_, err := u.reg.AppFlow.StartFlow(integration.FlowName)
		if err != nil {
			return model.ResponseWrapper{}, err
		}
	}
	return model.ResponseWrapper{}, nil
}

func (u *StartFlows) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
