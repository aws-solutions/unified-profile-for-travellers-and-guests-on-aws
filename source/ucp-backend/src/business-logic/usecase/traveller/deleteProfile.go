// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"errors"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	utils "tah/upt/source/ucp-common/src/utils/utils"

	"github.com/aws/aws-lambda-go/events"
)

type DeleteProfile struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewDeleteProfile() *DeleteProfile {
	return &DeleteProfile{name: "DeleteProfile"}
}

func (u *DeleteProfile) Name() string {
	return u.name
}
func (u *DeleteProfile) Tx() core.Transaction {
	return *u.tx
}
func (u *DeleteProfile) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *DeleteProfile) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *DeleteProfile) Registry() *registry.Registry {
	return u.reg
}

func (u *DeleteProfile) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	rw := model.RequestWrapper{
		ID: req.PathParameters["id"],
	}
	return rw, nil
}

func (u *DeleteProfile) AccessPermission() constant.AppPermission {
	return constant.DeleteProfilePermission
}

func (u *DeleteProfile) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	if u.reg.DataAccessPermission != constant.DATA_ACCESS_PERMISSION_ALL {
		return errors.New("only administrator can delete profiles")
	}
	if !utils.IsUUID(rq.ID) {
		return errors.New("invalid Connect ID format")
	}
	return nil
}

func (u *DeleteProfile) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	err := u.reg.Accp.DeleteProfile(req.ID)
	return model.ResponseWrapper{}, err
}

func (u *DeleteProfile) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
