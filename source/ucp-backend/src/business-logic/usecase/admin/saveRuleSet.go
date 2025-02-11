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

type SaveRuleSet struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewSaveIdResRuleSet() *SaveRuleSet {
	return &SaveRuleSet{name: "SaveRuleSet"}
}

func (u *SaveRuleSet) Name() string {
	return u.name
}
func (u *SaveRuleSet) Tx() core.Transaction {
	return *u.tx
}
func (u *SaveRuleSet) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *SaveRuleSet) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *SaveRuleSet) Registry() *registry.Registry {
	return u.reg
}

func (u *SaveRuleSet) AccessPermission() constant.AppPermission {
	return constant.SaveRuleSetPermission
}

func (u *SaveRuleSet) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *SaveRuleSet) ValidateRequest(req model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())

	rules := req.SaveRuleSet.Rules
	if rules == nil {
		return errors.New("rule set request has no rules field")
	}
	return nil
}

func (u *SaveRuleSet) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	err := u.reg.Accp.SaveIdResRuleSet(req.SaveRuleSet.Rules)
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	return model.ResponseWrapper{}, nil
}

func (u *SaveRuleSet) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
