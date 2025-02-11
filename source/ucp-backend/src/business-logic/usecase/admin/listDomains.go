// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"

	"github.com/aws/aws-lambda-go/events"
)

type ListUcpDomains struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewListUcpDomains() *ListUcpDomains {
	return &ListUcpDomains{name: "ListUcpDomains"}
}

func (u *ListUcpDomains) Name() string {
	return u.name
}
func (u *ListUcpDomains) Tx() core.Transaction {
	return *u.tx
}
func (u *ListUcpDomains) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *ListUcpDomains) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *ListUcpDomains) Registry() *registry.Registry {
	return u.reg
}

func (u *ListUcpDomains) AccessPermission() constant.AppPermission {
	return constant.PublicAccessPermission
}

func (u *ListUcpDomains) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *ListUcpDomains) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	return nil
}

func (u *ListUcpDomains) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	profileDomains, err := u.reg.Accp.ListDomains()
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	domains := []model.Domain{}
	for _, dom := range profileDomains {
		if u.reg.Env["LAMBDA_ENV"] == dom.Tags[DOMAIN_TAG_ENV_NAME] {
			domains = append(domains, model.Domain{
				Name:        dom.Name,
				Created:     dom.Created,
				LastUpdated: dom.LastUpdated,
			})
		}
	}
	return model.ResponseWrapper{UCPConfig: &model.UCPConfig{Domains: domains}}, err
}

func (u *ListUcpDomains) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
