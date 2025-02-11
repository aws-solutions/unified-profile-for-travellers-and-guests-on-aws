// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"math"

	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	commonModel "tah/upt/source/ucp-common/src/model/admin"

	"github.com/aws/aws-lambda-go/events"
)

type ListErrors struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewListErrors() *ListErrors {
	return &ListErrors{name: "ListErrors"}
}

func (u *ListErrors) Name() string {
	return u.name
}
func (u *ListErrors) Tx() core.Transaction {
	return *u.tx
}
func (u *ListErrors) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *ListErrors) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *ListErrors) Registry() *registry.Registry {
	return u.reg
}

func (u *ListErrors) AccessPermission() constant.AppPermission {
	return constant.PublicAccessPermission
}

func (u *ListErrors) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	wrapper, err := registry.CreateRequest(u, req)
	wrapper.SearchRq.TravellerID = req.QueryStringParameters["travellerId"]
	return wrapper, err
}

func (u *ListErrors) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	return nil
}

func (u *ListErrors) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	errs := []model.UcpIngestionError{}
	po := commonModel.PaginationOptions{}
	if len(req.Pagination) > 0 {
		po = req.Pagination[0]
	}
	queryOptions := db.QueryOptions{
		ReverseOrder: true,
		PaginOptions: db.DynamoPaginationOptions{
			Page:     int64(po.Page),
			PageSize: int32(po.PageSize),
		}}
	if req.SearchRq.TravellerID != "" {
		queryOptions.Filter = db.DynamoFilterExpression{BeginsWith: []db.DynamoFilterCondition{{Key: "travelerId", Value: req.SearchRq.TravellerID}}}
	}

	err := u.reg.ErrorDB.FindStartingWithAndFilterWithIndex(ERROR_PK, ERROR_SK_PREFIX, &errs, queryOptions)
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	count, err2 := u.reg.ErrorDB.GetItemCount()
	// ItemCount is provided by calling DescribeTable. The count is updated "approximately every six hours",
	// so there are times when the count is less than the actual number of items found.
	// To prevent a weird experience where the count is obviously wrong, we use the displayed error list's length.
	// (e.g. if a domain was recently created and experienced its first errors, the count could display 0 for several
	// hours despite there being a list of errors displayed to the user)
	// https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_TableDescription.html
	estimateErrorCount := float64((po.Page * po.PageSize) + len(errs))
	bestEstimate := int64(math.Max(estimateErrorCount, float64(count)))
	return model.ResponseWrapper{IngestionErrors: &errs, TotalErrors: &bestEstimate}, err2
}

func (u *ListErrors) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
