// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"encoding/json"
	"testing"

	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"

	"github.com/aws/aws-lambda-go/events"
)

func TestPromptConfig(t *testing.T) {
	t.Parallel()
	portalConfigTableName := "ucp-test-prompt-confi-" + core.GenerateUniqueId()

	portalCfg, err := db.InitWithNewTable(portalConfigTableName, "config_item", "config_item_category", "", "")
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}

	t.Cleanup(func() {
		err = portalCfg.DeleteTable(portalConfigTableName)
		if err != nil {
			t.Fatalf("[%s] Error deleting domain", t.Name())
		}
	})
	portalCfg.WaitForTableCreation()

	reg := registry.Registry{
		Env:            map[string]string{"ACCP_DOMAIN_NAME": "test_domain"},
		PortalConfigDB: &portalCfg,
	}

	validReqBody := model.RequestWrapper{
		DomainSetting: model.DomainSetting{
			PromptConfig: model.DomainConfig{
				Value:    "dummy prompt",
				IsActive: true,
			},
		},
	}
	invalidReqBody := model.RequestWrapper{
		DomainSetting: model.DomainSetting{
			PromptConfig: model.DomainConfig{
				Value:    "",
				IsActive: true,
			},
		},
	}

	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)

	savePromptConfig(t, reg, tx, validReqBody)
	getPromptConfig(t, reg, tx, validReqBody)
	invalidPromptConfig(t, reg, tx, invalidReqBody)
}

func savePromptConfig(t *testing.T, reg registry.Registry, tx core.Transaction, reqBody model.RequestWrapper) {
	updateuc := NewUpdatePromptConfig()
	updateuc.SetRegistry(&reg)
	updateuc.SetTx(&tx)
	updateuc.Tx()
	updateuc.Registry()

	if updateuc.Name() != "NewUpdatePromptConfig" {
		t.Fatalf("[%s] Error: expected request name NewUpdatePromptConfig but received %v", t.Name(), updateuc.Name())
	}

	rqBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("[%s] Error marshalling request", t.Name())
	}

	updateuc.reg.SetAppAccessPermission(constant.ConfigureGenAiPermission)
	saveWrapper, err := updateuc.CreateRequest(events.APIGatewayProxyRequest{Body: string(rqBody)})
	if err != nil {
		t.Fatalf("[%s] Error creating request", t.Name())
	}

	err = updateuc.ValidateRequest(saveWrapper)
	if err != nil {
		t.Fatalf("[%s] Request %v is invalid", t.Name(), saveWrapper)
	}

	res, err := updateuc.Run(saveWrapper)
	if err != nil {
		t.Fatalf("[%s] Could not run request: %v", t.Name(), err)
	}

	response, err := updateuc.CreateResponse(res)

	if err != nil {
		t.Fatalf("[%s] Error while creating response: %v", t.Name(), err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("[%s] Expected status code 200, got %d", t.Name(), response.StatusCode)
	}
}

func getPromptConfig(t *testing.T, reg registry.Registry, tx core.Transaction, reqBody model.RequestWrapper) {
	getuc := NewGetPromptConfig()

	getuc.SetRegistry(&reg)
	getuc.SetTx(&tx)
	getuc.Tx()
	getuc.Registry()

	if getuc.Name() != "GetPromptConfig" {
		t.Fatalf("[%s] Error: expected request name NewGetPromptConfig but received %v", t.Name(), getuc.Name())
	}

	getWrapper, err := getuc.CreateRequest(events.APIGatewayProxyRequest{})
	if err != nil {
		t.Fatalf("[%s] Error creating request", t.Name())
	}

	err = getuc.ValidateRequest(getWrapper)
	if err != nil {
		t.Fatalf("[%s] Request %v is invalid", t.Name(), getWrapper)
	}

	res, err := getuc.Run(getWrapper)
	if err != nil {
		t.Fatalf("[%s] error run GET request: %v", t.Name(), err)
	}

	// Validate
	if res.DomainSetting.PromptConfig.Value != reqBody.DomainSetting.PromptConfig.Value || res.DomainSetting.PromptConfig.IsActive != reqBody.DomainSetting.PromptConfig.IsActive {
		t.Fatalf("[%s] invalid Get request", t.Name())
	}

	response, err := getuc.CreateResponse(res)

	if err != nil {
		t.Fatalf("[%s] Error while creating response: %v", t.Name(), err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("[%s] Expected status code 200, got %d", t.Name(), response.StatusCode)
	}
}

func invalidPromptConfig(t *testing.T, reg registry.Registry, tx core.Transaction, reqBody model.RequestWrapper) {
	invaliduc := NewUpdatePromptConfig()
	invaliduc.SetRegistry(&reg)
	invaliduc.SetTx(&tx)

	rqBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("[%s] Error marshalling request", t.Name())
	}

	saveWrapper, err := invaliduc.CreateRequest(events.APIGatewayProxyRequest{
		Body: string(rqBody),
	})
	if err != nil {
		t.Fatalf("[%s] Error creating request", t.Name())
	}

	err = invaliduc.ValidateRequest(saveWrapper)
	if err == nil {
		t.Fatalf("[%s] Request %v is invalid", t.Name(), saveWrapper)
	}
}
