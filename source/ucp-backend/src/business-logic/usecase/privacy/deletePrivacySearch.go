// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package privacy

import (
	"encoding/json"
	"fmt"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/model/async/usecase"

	"github.com/aws/aws-lambda-go/events"
)

type DeletePrivacySearch struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

// Static interface check
var _ registry.Usecase = &DeletePrivacySearch{}

func NewDeletePrivacySearch() *DeletePrivacySearch {
	return &DeletePrivacySearch{
		name: "DeletePrivacySearch",
	}
}

// CreateRequest implements registry.Usecase.
func (u *DeletePrivacySearch) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	if req.Body == "" {
		return model.RequestWrapper{}, nil
	}

	var connectIds []string
	err := json.Unmarshal([]byte(req.Body), &connectIds)
	if err != nil {
		u.tx.Error("error parsing connect ids with request body %v: %v", req.Body, err)
		return model.RequestWrapper{}, fmt.Errorf("error parsing connect ids")
	}
	return model.RequestWrapper{
		DeletePrivacySearches: model.DeletePrivacySearchesRq{
			DomainName: u.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR],
			ConnectIds: connectIds,
		},
	}, nil
}

// CreateResponse implements registry.Usecase.
func (u *DeletePrivacySearch) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

// Name implements registry.Usecase.
func (u *DeletePrivacySearch) Name() string {
	return u.name
}

// Registry implements registry.Usecase.
func (u *DeletePrivacySearch) Registry() *registry.Registry {
	return u.reg
}

// Run implements registry.Usecase.
func (u *DeletePrivacySearch) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	deleteRequest := []map[string]string{}
	for _, connectId := range req.DeletePrivacySearches.ConnectIds {
		var connectIdLocationResults []usecase.PrivacySearchResult
		err := u.reg.PrivacyDB.FindStartingWith(req.DeletePrivacySearches.DomainName, connectId, &connectIdLocationResults)
		if err != nil {
			u.tx.Error("[%s] Error finding privacy search: %v", u.name, err)
			return model.ResponseWrapper{}, err
		}
		for _, result := range connectIdLocationResults {
			deleteRequest = append(deleteRequest, map[string]string{u.reg.PrivacyDB.PrimaryKey: result.DomainName, u.reg.PrivacyDB.SortKey: result.ConnectId})
		}
	}
	u.tx.Info("[%s] Deleting privacy searches: %v", u.name, deleteRequest)
	err := u.reg.PrivacyDB.DeleteMany(deleteRequest)
	if err != nil {
		u.tx.Error("[%s] Error deleting privacy search: %v", u.name, err)
		return model.ResponseWrapper{}, err
	}

	return model.ResponseWrapper{}, nil
}

// SetRegistry implements registry.Usecase.
func (u *DeletePrivacySearch) SetRegistry(r *registry.Registry) {
	u.reg = r
}

// SetTx implements registry.Usecase.
func (u *DeletePrivacySearch) SetTx(tx *core.Transaction) {
	u.tx = tx
}

// Tx implements registry.Usecase.
func (u *DeletePrivacySearch) Tx() core.Transaction {
	return *u.tx
}

func (u *DeletePrivacySearch) AccessPermission() admin.AppPermission {
	return admin.DeletePrivacySearchPermission
}

// ValidateRequest implements registry.Usecase.
func (u *DeletePrivacySearch) ValidateRequest(req model.RequestWrapper) error {
	u.tx.Debug("[%s] Validating Request", u.name)

	if len(req.DeletePrivacySearches.ConnectIds) == 0 {
		return fmt.Errorf("[%s] a connect id is required", u.name)
	}

	return nil
}
