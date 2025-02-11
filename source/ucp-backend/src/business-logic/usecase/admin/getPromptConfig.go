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

type GetPromptConfig struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewGetPromptConfig() *GetPromptConfig {
	return &GetPromptConfig{name: "GetPromptConfig"}
}

func (u *GetPromptConfig) Name() string {
	return u.name
}
func (u *GetPromptConfig) Tx() core.Transaction {
	return *u.tx
}
func (u *GetPromptConfig) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *GetPromptConfig) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *GetPromptConfig) Registry() *registry.Registry {
	return u.reg
}

func (u *GetPromptConfig) AccessPermission() constant.AppPermission {
	return constant.PublicAccessPermission
}

func (u *GetPromptConfig) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *GetPromptConfig) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	return nil
}

func (u *GetPromptConfig) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

func (u *GetPromptConfig) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	u.tx.Info("[GetPromptConfig] getting prompt configuration")
	promptConfig, promptConfigErr := u.getConfig(PORTAL_CONFIG_PROMPT_PK)
	matchConfig, matchConfigErr := u.getConfig(PORTAL_CONFIG_MATCH_THRESHOLD_PK)

	allError := errors.Join(promptConfigErr, matchConfigErr)
	if allError != nil {
		u.tx.Error("error getting configuration: %s", allError)
	}

	return model.ResponseWrapper{DomainSetting: &model.DomainSetting{
		PromptConfig: promptConfig,
		MatchConfig:  matchConfig,
	}}, allError
}

func (u *GetPromptConfig) getConfig(key string) (model.DomainConfig, error) {
	config := model.DynamoDomainConfig{}
	value := ""
	isActive := false

	err := u.reg.PortalConfigDB.Get(key, u.reg.Env["ACCP_DOMAIN_NAME"], &config)
	if err != nil {
		u.tx.Error("error getting configuration: %s", err)
	} else {
		value = config.Value
		isActive = config.IsActive
	}

	return model.DomainConfig{
		Value:    value,
		IsActive: isActive,
	}, err
}
