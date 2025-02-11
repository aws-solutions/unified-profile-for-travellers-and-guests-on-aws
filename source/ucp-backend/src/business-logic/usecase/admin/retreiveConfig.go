// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"errors"

	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	common "tah/upt/source/ucp-common/src/constant/admin"
	accpmappings "tah/upt/source/ucp-common/src/model/accp-mappings"
	admin "tah/upt/source/ucp-common/src/services/admin"

	"github.com/aws/aws-lambda-go/events"
)

type RetreiveConfig struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewRetreiveConfig() *RetreiveConfig {
	return &RetreiveConfig{name: "RetreiveConfig"}
}

func (u *RetreiveConfig) Name() string {
	return u.name
}
func (u *RetreiveConfig) Tx() core.Transaction {
	return *u.tx
}
func (u *RetreiveConfig) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *RetreiveConfig) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *RetreiveConfig) Registry() *registry.Registry {
	return u.reg
}

func (u *RetreiveConfig) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *RetreiveConfig) AccessPermission() common.AppPermission {
	return common.PublicAccessPermission
}

func (u *RetreiveConfig) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	return nil
}

func (u *RetreiveConfig) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	dom, err := u.reg.Accp.GetDomain()
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	domain := model.Domain{
		Name:             dom.Name,
		NObjects:         dom.NObjects,
		NProfiles:        dom.NProfiles,
		MatchingEnabled:  dom.MatchingEnabled,
		IsLowCostEnabled: dom.IsLowCostEnabled,
		CacheMode:        dom.CacheMode,
	}
	mappings, err2 := u.reg.Accp.GetMappings()
	if err2 != nil {
		return model.ResponseWrapper{}, err2
	}
	domain.Mappings = admin.ParseMappings(mappings)

	//effective mappings are the mapping in place in the domain (as oposed to latest mapping which are the latest mapping in the solution version)
	u.tx.Debug("Checking for outdated mapppings %s")
	effectiveMappingsByObjectType := admin.OrganizeByObjectType(domain.Mappings)
	for _, businessObject := range common.ACCP_RECORDS {
		accpRecName := businessObject.Name
		effectiveMapping := effectiveMappingsByObjectType[accpRecName]
		latestMapping := accpmappings.ACCP_OBJECT_MAPPINGS[accpRecName]()
		u.tx.Debug("Checking if mapping is outdated for %s", accpRecName)
		u.tx.Debug("Effective mapping %+v", effectiveMapping)
		u.tx.Debug("Latest mapping %+v", admin.ParseMapping(latestMapping))
		isOutdated, err := admin.HasOutdatedMappings(u.Tx(), effectiveMapping, admin.ParseMapping(latestMapping))
		if err != nil {
			return model.ResponseWrapper{}, err
		}
		if isOutdated {
			domain.NeedsMappingUpdate = isOutdated
		}
	}
	u.tx.Debug("Looking for missing event source mapping for retry lambda")
	hasEsn, err := admin.HasEventSourceMapping(u.reg.RetryLambda, *u.reg.ErrorDB, *u.tx)
	if err != nil {
		return model.ResponseWrapper{}, errors.New("Could not find event source mapping: " + err.Error())
	}
	if !hasEsn {
		u.tx.Info("Missing event source mapping for retry lambda. Prompting user to refresh the domain")
		domain.NeedsMappingUpdate = true
	}

	return model.ResponseWrapper{UCPConfig: &model.UCPConfig{Domains: []model.Domain{domain}}}, err
}

func (u *RetreiveConfig) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
