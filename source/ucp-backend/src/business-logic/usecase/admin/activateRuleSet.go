// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"encoding/json"
	"errors"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	async "tah/upt/source/ucp-common/src/model/async"
	"tah/upt/source/ucp-common/src/model/async/usecase"
	commonAdmin "tah/upt/source/ucp-common/src/services/admin"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/google/uuid"
)

type ActivateRuleSet struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewActivateIdResRuleSet() *ActivateRuleSet {
	return &ActivateRuleSet{name: "ActivateRuleSet"}
}

func (u *ActivateRuleSet) Name() string {
	return u.name
}
func (u *ActivateRuleSet) Tx() core.Transaction {
	return *u.tx
}
func (u *ActivateRuleSet) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *ActivateRuleSet) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *ActivateRuleSet) Registry() *registry.Registry {
	return u.reg
}

func (u *ActivateRuleSet) AccessPermission() constant.AppPermission {
	return constant.ActivateRuleSetPermission
}

func (u *ActivateRuleSet) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *ActivateRuleSet) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())

	if u.reg.Env["ACCP_DOMAIN_NAME"] == "" {
		return errors.New("no domain found, cannot activate rule set")
	}
	return nil
}

func (u *ActivateRuleSet) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	params, err := u.reg.Ssm.GetParametersByPath(context.TODO(), u.reg.Env["SSM_PARAM_NAMESPACE"])
	if err != nil {
		u.tx.Error("error querying SSM parameters: %v", err)
		return model.ResponseWrapper{}, err
	}
	taskConfig, err := commonAdmin.GetTaskConfigForIdRes(params)
	if err != nil {
		u.tx.Error("error loading ECS task config for batch ID res: %v", err)
		return model.ResponseWrapper{}, err
	}
	tasks, err := u.reg.Ecs.ListTasks(context.TODO(), taskConfig, nil, aws.String(string(awsecstypes.DesiredStatusRunning)))
	if err != nil {
		u.tx.Error("error listing batch ID res tasks: %v", err)
		return model.ResponseWrapper{}, err
	}
	if len(tasks) > 0 {
		return model.ResponseWrapper{}, errors.New("unable to activate rule set while batch ID res is running")
	}

	err = u.reg.Accp.ActivateIdResRuleSet()
	if err != nil {
		return model.ResponseWrapper{}, err
	}

	domainName := u.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR]

	eventId := uuid.NewString()
	payload := async.AsyncInvokePayload{
		EventID:       eventId,
		Usecase:       async.USECASE_START_BATCH_ID_RES,
		TransactionID: u.tx.TransactionID,
		Body:          usecase.StartBatchIdResBody{Domain: domainName},
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		u.tx.Error("[%s] Error marshalling aysnc payload: %v", u.Name(), err)
		return model.ResponseWrapper{}, err
	}

	if u.reg.AsyncLambda == nil {
		err = errors.New("no async Lambda client")
		u.tx.Error("[%s] Error invoking async Lambda: %v", u.Name(), err)
		return model.ResponseWrapper{}, err
	}
	_, err = u.reg.AsyncLambda.InvokeAsync(payloadJson)
	if err != nil {
		u.tx.Error("[%s] Error invoking async Lambda: %v", u.Name(), err)
		return model.ResponseWrapper{}, err
	}

	if u.reg.ConfigDB == nil {
		err = errors.New("no config DB client")
		u.tx.Error("[%s] Error saving async event to config DB: %v", u.Name(), err)
		return model.ResponseWrapper{}, err
	}

	asyncEvent := async.AsyncEvent{
		EventID:     eventId,
		Usecase:     async.USECASE_START_BATCH_ID_RES,
		Status:      async.EVENT_STATUS_INVOKED,
		LastUpdated: time.Now(),
	}
	_, err = u.reg.ConfigDB.Save(asyncEvent)
	if err != nil {
		u.tx.Error("[%s] Error saving async event to config DB: %v", u.Name(), err)
		return model.ResponseWrapper{}, err
	}

	response := model.ResponseWrapper{
		AsyncEvent: &asyncEvent,
	}
	return response, nil
}

func (u *ActivateRuleSet) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
