// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package privacy

import (
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"

	"github.com/aws/aws-lambda-go/events"
)

type GetPurgeStatus struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

// Static interface check
var _ registry.Usecase = (*GetPurgeStatus)(nil)

func NewGetPurgeStatus() *GetPurgeStatus {
	return &GetPurgeStatus{
		name: "GetPurgeStatus",
	}
}

// CreateRequest implements registry.Usecase.
func (u *GetPurgeStatus) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return model.RequestWrapper{
		GetPurgeStatus: model.GetPurgeStatusRq{
			DomainName: u.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR],
		},
	}, nil
}

// CreateResponse implements registry.Usecase.
func (u *GetPurgeStatus) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

func (u *GetPurgeStatus) AccessPermission() admin.AppPermission {
	return admin.PrivacyDataPurgePermission
}

// Name implements registry.Usecase.
func (u *GetPurgeStatus) Name() string {
	return u.name
}

// Registry implements registry.Usecase.
func (u *GetPurgeStatus) Registry() *registry.Registry {
	return u.reg
}

// Run implements registry.Usecase.
func (u *GetPurgeStatus) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	u.tx.Info("[%s] Getting Is Purge Running for Domain: %s", u.name, req.GetPurgeStatus.DomainName)

	searchResults := []uc.PrivacySearchResult{}
	err := u.reg.PrivacyDB.FindByPk(req.GetPurgeStatus.DomainName, &searchResults)
	if err != nil {
		u.tx.Error("[%s] Error getting records from PrivacySearch table: %s", u.name, err.Error())
		return model.ResponseWrapper{}, err
	}

	for _, result := range searchResults {
		//	Limit checks to privacy records that have had a purge initiated
		//	These are denoted by a pipe (|) in the connectId
		if strings.Contains(result.ConnectId, "|") {
			if result.Status == uc.PRIVACY_STATUS_PURGE_INVOKED || result.Status == uc.PRIVACY_STATUS_PURGE_RUNNING {
				return model.ResponseWrapper{
					PrivacyPurgeStatus: &uc.PurgeStatusResult{
						DomainName:     req.GetPurgeStatus.DomainName,
						IsPurgeRunning: true,
					},
				}, nil
			}
		}
	}

	return model.ResponseWrapper{
		PrivacyPurgeStatus: &uc.PurgeStatusResult{
			DomainName:     req.GetPurgeStatus.DomainName,
			IsPurgeRunning: false,
		},
	}, nil
}

// SetRegistry implements registry.Usecase.
func (u *GetPurgeStatus) SetRegistry(r *registry.Registry) {
	u.reg = r
}

// SetTx implements registry.Usecase.
func (u *GetPurgeStatus) SetTx(tx *core.Transaction) {
	u.tx = tx
}

// Tx implements registry.Usecase.
func (u *GetPurgeStatus) Tx() core.Transaction {
	return *u.tx
}

// ValidateRequest implements registry.Usecase.
func (u *GetPurgeStatus) ValidateRequest(req model.RequestWrapper) error {
	u.tx.Debug("[%s] Validating request", u.name)
	return nil
}
