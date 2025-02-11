// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"errors"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	commonServices "tah/upt/source/ucp-common/src/services/admin"

	"github.com/aws/aws-lambda-go/events"
)

type UpdatePromptConfig struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewUpdatePromptConfig() *UpdatePromptConfig {
	return &UpdatePromptConfig{name: "NewUpdatePromptConfig"}
}

func (u *UpdatePromptConfig) Name() string {
	return u.name
}
func (u *UpdatePromptConfig) Tx() core.Transaction {
	return *u.tx
}
func (u *UpdatePromptConfig) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *UpdatePromptConfig) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *UpdatePromptConfig) Registry() *registry.Registry {
	return u.reg
}

func (u *UpdatePromptConfig) AccessPermission() constant.AppPermission {
	return constant.ConfigureGenAiPermission
}

func (u *UpdatePromptConfig) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *UpdatePromptConfig) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[UpdatePromptConfig] Validating request ", rq)

	if rq.DomainSetting.PromptConfig.IsActive && rq.DomainSetting.PromptConfig.Value == "" {
		return errors.New("prompt value cannot be empty if summary generation is active")
	}
	return nil
}

func (u *UpdatePromptConfig) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

func (u *UpdatePromptConfig) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	u.tx.Info("[UpdatePromptConfig] Saving prompt configuration ", req.DomainSetting)
	domainName := u.reg.Env["ACCP_DOMAIN_NAME"]
	err := commonServices.UpdateDomainSetting(*u.tx, *u.reg.PortalConfigDB, PORTAL_CONFIG_PROMPT_PK, domainName, req.DomainSetting.PromptConfig.Value, req.DomainSetting.PromptConfig.IsActive)
	if err != nil {
		u.tx.Error("[UpdatePromptConfig] Error saving prompt configuration; error ", err)
		return model.ResponseWrapper{}, err
	}

	u.tx.Info("[UpdatePromptConfig] Saving match configuration")
	err = commonServices.UpdateDomainSetting(*u.tx, *u.reg.PortalConfigDB, PORTAL_CONFIG_MATCH_THRESHOLD_PK, domainName, req.DomainSetting.MatchConfig.Value, req.DomainSetting.MatchConfig.IsActive)
	if err != nil {
		u.tx.Error("[UpdatePromptConfig] Error saving match configuration; error ", err)
		return model.ResponseWrapper{}, err
	}
	return model.ResponseWrapper{}, nil
}
