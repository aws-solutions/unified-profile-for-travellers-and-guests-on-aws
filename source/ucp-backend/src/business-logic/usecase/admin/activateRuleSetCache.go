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

type ActivateCacheRuleSet struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewActivateCacheRuleSet() *ActivateCacheRuleSet {
	return &ActivateCacheRuleSet{name: "ActivateCacheRuleSet"}
}

func (u *ActivateCacheRuleSet) Name() string {
	return u.name
}
func (u *ActivateCacheRuleSet) Tx() core.Transaction {
	return *u.tx
}
func (u *ActivateCacheRuleSet) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *ActivateCacheRuleSet) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *ActivateCacheRuleSet) Registry() *registry.Registry {
	return u.reg
}

func (u *ActivateCacheRuleSet) AccessPermission() constant.AppPermission {
	return constant.ActivateRuleSetPermission
}

func (u *ActivateCacheRuleSet) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *ActivateCacheRuleSet) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	domainName := u.reg.Env["ACCP_DOMAIN_NAME"]
	if domainName == "" {
		return errors.New("no domain found, cannot activate rule set")
	}
	return nil
}

func (u *ActivateCacheRuleSet) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	err := u.reg.Accp.ActivateCacheRuleSet()
	emptyResponse := model.ResponseWrapper{}
	if err != nil {
		return emptyResponse, err
	}
	return emptyResponse, nil
}

func (u *ActivateCacheRuleSet) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
