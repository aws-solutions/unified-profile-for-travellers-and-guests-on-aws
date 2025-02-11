// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"errors"
	"strconv"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"

	"github.com/aws/aws-lambda-go/events"
)

type ListRuleSets struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewListIdResRuleSets() *ListRuleSets {
	return &ListRuleSets{name: "ListRuleSets"}
}

func (u *ListRuleSets) Name() string {
	return u.name
}
func (u *ListRuleSets) Tx() core.Transaction {
	return *u.tx
}
func (u *ListRuleSets) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *ListRuleSets) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *ListRuleSets) Registry() *registry.Registry {
	return u.reg
}

func (u *ListRuleSets) AccessPermission() constant.AppPermission {
	return constant.ListRuleSetPermission
}

func (u *ListRuleSets) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	wrapper, err := registry.CreateRequest(u, req)
	includeHistorical, paramExists := req.QueryStringParameters["includesHistorical"]
	if paramExists {
		if includeHistorical == "true" || includeHistorical == "false" {
			wrapper.ListRuleSets.IncludeHistorical, _ = strconv.ParseBool(includeHistorical)
		} else {
			return model.RequestWrapper{}, errors.New("query parameter includeHistorical must either be 'true' or 'false'")
		}
	} else {
		wrapper.ListRuleSets.IncludeHistorical = false
	}
	return wrapper, err
}

func (u *ListRuleSets) ValidateRequest(req model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	return nil
}

func (u *ListRuleSets) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	includeHistorical := req.ListRuleSets.IncludeHistorical
	ruleSets, err := u.reg.Accp.ListIdResRuleSets(includeHistorical)
	if err != nil {
		return model.ResponseWrapper{}, err
	}

	profileMappings, err := u.reg.Accp.GetProfileLevelFields()
	if err != nil {
		return model.ResponseWrapper{}, err
	}

	return model.ResponseWrapper{RuleSets: &ruleSets, ProfileMappings: &profileMappings}, nil
}

func (u *ListRuleSets) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
