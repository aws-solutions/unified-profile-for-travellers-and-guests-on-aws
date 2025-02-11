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

type GetPortalConfig struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewGetPortalConfig() *GetPortalConfig {
	return &GetPortalConfig{name: "GetPortalConfig"}
}

func (u *GetPortalConfig) Name() string {
	return u.name
}
func (u *GetPortalConfig) Tx() core.Transaction {
	return *u.tx
}
func (u *GetPortalConfig) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *GetPortalConfig) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *GetPortalConfig) Registry() *registry.Registry {
	return u.reg
}

func (u *GetPortalConfig) AccessPermission() constant.AppPermission {
	return constant.PublicAccessPermission
}

func (u *GetPortalConfig) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *GetPortalConfig) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	return nil
}

func (u *GetPortalConfig) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	dynMappings := []model.DynamoHyperlinkMapping{}
	mappings := []model.HyperlinkMapping{}

	err := u.reg.PortalConfigDB.FindStartingWith(PORTAL_CONFIG_HYPERLINKS_PK, PORTAL_CONFIG_HYPERLINKS_SK_PREFIX, &dynMappings)
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	for _, dyn := range dynMappings {
		mappings = append(mappings, model.HyperlinkMapping{
			AccpObject: dyn.AccpObject,
			FieldName:  dyn.FieldName,
			Template:   dyn.Template,
		})
	}
	return model.ResponseWrapper{PortalConfig: &model.PortalConfig{HyperlinkMappings: mappings}}, err
}

func (u *GetPortalConfig) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
