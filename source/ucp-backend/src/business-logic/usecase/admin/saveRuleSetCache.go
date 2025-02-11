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

type SaveCacheRuleSet struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewSaveCacheRuleSet() *SaveCacheRuleSet {
	return &SaveCacheRuleSet{name: "SaveCacheRuleSet"}
}

func (u *SaveCacheRuleSet) Name() string {
	return u.name
}
func (u *SaveCacheRuleSet) Tx() core.Transaction {
	return *u.tx
}
func (u *SaveCacheRuleSet) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *SaveCacheRuleSet) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *SaveCacheRuleSet) Registry() *registry.Registry {
	return u.reg
}

func (u *SaveCacheRuleSet) AccessPermission() constant.AppPermission {
	return constant.SaveRuleSetPermission
}

func (u *SaveCacheRuleSet) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *SaveCacheRuleSet) ValidateRequest(req model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())

	rules := req.SaveRuleSet.Rules
	if rules == nil {
		return errors.New("rule set request has no rules field")
	}
	return nil
}

func (u *SaveCacheRuleSet) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	err := u.reg.Accp.SaveCacheRuleSet(req.SaveRuleSet.Rules)
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	return model.ResponseWrapper{}, nil
}

func (u *SaveCacheRuleSet) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
