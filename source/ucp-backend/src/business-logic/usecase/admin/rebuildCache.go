// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	async "tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

// The max value should change when more cache types are added
var MIN_CACHE_BIT customerprofileslcs.CacheMode

type RebuildCache struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewRebuildCache() *RebuildCache {
	return &RebuildCache{name: "RebuildCache"}
}

func (u *RebuildCache) Name() string {
	return u.name
}
func (u *RebuildCache) Tx() core.Transaction {
	return *u.tx
}
func (u *RebuildCache) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *RebuildCache) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *RebuildCache) Registry() *registry.Registry {
	return u.reg
}

func (u *RebuildCache) AccessPermission() constant.AppPermission {
	return constant.RebuildCachePermission
}

func (u *RebuildCache) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	wrapper, err := registry.CreateRequest(u, req)
	cacheType, exists := req.QueryStringParameters["type"]
	if exists {
		typeInt, err := strconv.ParseUint(cacheType, 10, binary.Size(customerprofileslcs.CacheMode(0))*8)
		if err != nil {
			return model.RequestWrapper{}, fmt.Errorf("error parsing query parameter type: %v", err)
		}
		wrapper.RebuildCache.CacheMode = customerprofileslcs.CacheMode(typeInt)
	} else {
		return model.RequestWrapper{}, fmt.Errorf("query parameter 'type' is required")
	}
	return wrapper, err
}

func (u *RebuildCache) ValidateRequest(req model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	cacheMode := req.RebuildCache.CacheMode
	if cacheMode < MIN_CACHE_BIT || cacheMode > customerprofileslcs.MAX_CACHE_BIT {
		return fmt.Errorf("cache mode not recognized")
	}
	return nil
}

func (u *RebuildCache) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	body := uc.RebuildCacheBody{
		MatchBucketName: u.reg.Env["MATCH_BUCKET_NAME"],
		Env:             u.reg.Env["LAMBDA_ENV"],
		DomainName:      u.reg.Env["ACCP_DOMAIN_NAME"],
		CacheMode:       req.RebuildCache.CacheMode,
	}

	eventId := uuid.NewString()
	payload := async.AsyncInvokePayload{
		EventID:       eventId,
		Usecase:       async.USECASE_REBUILD_CACHE,
		TransactionID: u.tx.TransactionID,
		Body:          body,
	}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		u.tx.Error("[%s] Error marshalling async payload: %v", u.Name(), err)
		return model.ResponseWrapper{}, err
	}

	_, err = u.reg.AsyncLambda.InvokeAsync(payloadJson)
	if err != nil {
		u.tx.Error("[%s] Error invoking async lambda: %v", u.Name(), err)
		return model.ResponseWrapper{}, err
	}

	asyncEvent := async.AsyncEvent{
		EventID:     eventId,
		Usecase:     async.USECASE_REBUILD_CACHE,
		Status:      async.EVENT_STATUS_INVOKED,
		LastUpdated: time.Now(),
	}
	_, err = u.reg.ConfigDB.Save(asyncEvent)
	if err != nil {
		return model.ResponseWrapper{}, err
	}

	res := model.ResponseWrapper{
		AsyncEvent: &asyncEvent,
	}
	return res, nil
}

func (u *RebuildCache) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
