// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"errors"

	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	common "tah/upt/source/ucp-common/src/constant/admin"
	accpmappings "tah/upt/source/ucp-common/src/model/accp-mappings"
	adminModel "tah/upt/source/ucp-common/src/model/admin"
	admin "tah/upt/source/ucp-common/src/services/admin"

	"github.com/aws/aws-lambda-go/events"
)

type UpdateDomain struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewUpdateDomain() *UpdateDomain {
	return &UpdateDomain{name: "UpdateDomain"}
}

func (u *UpdateDomain) Name() string {
	return u.name
}
func (u *UpdateDomain) Tx() core.Transaction {
	return *u.tx
}
func (u *UpdateDomain) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *UpdateDomain) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *UpdateDomain) Registry() *registry.Registry {
	return u.reg
}

func (u *UpdateDomain) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	u.tx.Debug("Creating request wrapper")
	return registry.CreateRequest(u, req)
}

func (u *UpdateDomain) AccessPermission() common.AppPermission {
	return common.PublicAccessPermission
}

func (u *UpdateDomain) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	// Validate against existing domains
	_, err := u.reg.Accp.GetDomain()
	if err != nil {
		return errors.New("error retrieving domain or domain does not exist")
	}
	return nil
}

func (u *UpdateDomain) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	u.tx.Info("Running UpdateDomain")
	res := model.ResponseWrapper{}
	mappings, err2 := u.reg.Accp.GetMappings()
	if err2 != nil {
		return model.ResponseWrapper{}, err2
	}

	//effective mappings are the mapping in place in the domain (as oposed to latest mapping which are the latest mapping in the solution version)
	effectiveMappingsByObjectType := admin.OrganizeByObjectType(admin.ParseMappings(mappings))
	for _, businessObject := range common.ACCP_RECORDS {
		accpRecName := businessObject.Name
		effectiveMapping := effectiveMappingsByObjectType[accpRecName]
		latestMapping := accpmappings.ACCP_OBJECT_MAPPINGS[accpRecName]()
		_, err := u.UpdateMappings(accpRecName, effectiveMapping, latestMapping)
		if err != nil {
			return res, err
		}
	}
	//We need this check in case the customer has created a domain with v1.0.0 which did not support
	//Have this check at domain creation
	u.tx.Debug("Checking if EventSourceMapping exists for RetryLambda and creating it if not %s")
	_, err := admin.CreateRetryLambdaEventSourceMapping(u.reg.RetryLambda, *u.reg.ErrorDB, *u.tx)
	if err != nil {
		u.tx.Error("Error creating RetryLambda EventSourceMapping")
		return res, err
	}
	return model.ResponseWrapper{}, nil
}

func (u *UpdateDomain) UpdateMappings(accpRecName string, effectiveMapping adminModel.ObjectMapping, latestMapping customerprofiles.ObjectMapping) (bool, error) {
	isOutdated, err := admin.HasOutdatedMappings(u.Tx(), effectiveMapping, admin.ParseMapping(latestMapping))
	if err != nil {
		return false, err
	}
	if isOutdated {
		u.tx.Info("[UpdateUcpDomain] Updating mapping for %s", accpRecName)
		err := u.reg.Accp.CreateMapping(accpRecName,
			"Primary Mapping for the "+accpRecName+" object (v"+latestMapping.Version+")", latestMapping.Fields)
		if err != nil {
			u.tx.Error("Error updating mapping for %s", accpRecName)
			return false, err
		}
		return true, nil
	} else {
		u.tx.Info("[UpdateUcpDomain] No mapping to be updated for %s", accpRecName)
	}
	return false, nil
}

func (u *UpdateDomain) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
