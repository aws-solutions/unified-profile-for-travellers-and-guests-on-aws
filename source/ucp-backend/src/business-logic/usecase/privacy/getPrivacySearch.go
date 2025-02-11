// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package privacy

import (
	"errors"
	"fmt"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	privacySearchResultModel "tah/upt/source/ucp-common/src/model/async/usecase"
	"tah/upt/source/ucp-common/src/utils/utils"

	"github.com/aws/aws-lambda-go/events"
)

type GetPrivacySearch struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

// Static interface check
var _ registry.Usecase = &GetPrivacySearch{}

func NewGetPrivacySearch() *GetPrivacySearch {
	return &GetPrivacySearch{
		name: "GetPrivacySearch",
	}
}

// CreateRequest implements registry.Usecase.
func (u *GetPrivacySearch) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return model.RequestWrapper{
		GetPrivacySearch: model.GetPrivacySearchRq{
			DomainName: u.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR],
			ConnectId:  req.PathParameters[ID_PATH_PARAMETER],
		},
	}, nil
}

// CreateResponse implements registry.Usecase.
func (u *GetPrivacySearch) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

// Name implements registry.Usecase.
func (u *GetPrivacySearch) Name() string {
	return u.name
}

// Registry implements registry.Usecase.
func (u *GetPrivacySearch) Registry() *registry.Registry {
	return u.reg
}

// Run implements registry.Usecase.
func (u *GetPrivacySearch) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	searchResult := privacySearchResultModel.PrivacySearchResult{}

	err := u.reg.PrivacyDB.Get(req.GetPrivacySearch.DomainName, req.GetPrivacySearch.ConnectId, &searchResult)
	if err != nil {
		u.tx.Error("[%s] Error getting privacy search: %s", u.name, err.Error())
		return model.ResponseWrapper{}, err
	}

	return model.ResponseWrapper{
		PrivacySearchResult: &searchResult,
	}, nil
}

// SetRegistry implements registry.Usecase.
func (u *GetPrivacySearch) SetRegistry(r *registry.Registry) {
	u.reg = r
}

// SetTx implements registry.Usecase.
func (u *GetPrivacySearch) SetTx(tx *core.Transaction) {
	u.tx = tx
}

// Tx implements registry.Usecase.
func (u *GetPrivacySearch) Tx() core.Transaction {
	return *u.tx
}

func (u *GetPrivacySearch) AccessPermission() admin.AppPermission {
	return admin.GetPrivacySearchPermission
}

// ValidateRequest implements registry.Usecase.
func (u *GetPrivacySearch) ValidateRequest(req model.RequestWrapper) error {
	u.tx.Debug("[%s] Validating Request", u.name)

	if req.GetPrivacySearch.ConnectId == "" {
		return fmt.Errorf("[%s] a connectid must be supplied", u.name)
	}
	if !utils.IsUUID(req.GetPrivacySearch.ConnectId) {
		return errors.New("invalid Connect ID format")
	}
	if req.GetPrivacySearch.DomainName == "" {
		return fmt.Errorf("[%s] a domain name is required", u.name)
	}

	return nil
}
