// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/bedrock"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestGetProfileSummary(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("[%s] error getting config: %v", t.Name(), err)
	}

	portalConfigTableName := "ucp-test-profile-summary-" + core.GenerateUniqueId()
	portalCfg, err := db.InitWithNewTable(portalConfigTableName, "config_item", "config_item_category", "", "")
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	portalCfg.WaitForTableCreation()
	t.Cleanup(func() {
		err = portalCfg.DeleteTable(portalConfigTableName)
		if err != nil {
			t.Fatalf("[%s] Error deleting table", t.Name())
		}
	})

	profile := profilemodel.Profile{
		ProfileId:      "mock-traveller",
		FirstName:      "First",
		LastName:       "Last",
		ProfileObjects: []profilemodel.ProfileObject{},
	}
	accpService := customerprofiles.InitMock(nil, nil, &profile, nil, &[]customerprofiles.ObjectMapping{
		{Name: "air_booking"},
		{Name: "air_loyalty"},
		{Name: "hotel_loyalty"},
		{Name: "hotel_booking"},
		{Name: "hotel_stay_revenue_items"},
		{Name: "customer_service_interaction"},
		{Name: "ancillary_service"},
		{Name: "alternate_profile_id"},
		{Name: "guest_profile"},
		{Name: "pax_profile"},
		{Name: "email_history"},
		{Name: "phone_history"},
		{Name: "loyalty_transaction"},
		{Name: "clickstream"},
	})

	bedrockClient := bedrock.Init(envCfg.Region, "", "")

	reg := registry.Registry{
		Env:            map[string]string{"ACCP_DOMAIN_NAME": "test_domain"},
		PortalConfigDB: &portalCfg,
		Bedrock:        &bedrockClient,
		Accp:           accpService,
	}

	dynamoRecord := model.DynamoDomainConfig{
		Pk:       "ai_prompt",
		Sk:       "test_domain",
		Value:    "promptValue",
		IsActive: true,
	}
	portalCfg.Save(dynamoRecord)

	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)

	getProfileSummary := NewGetProfileSummary()
	getProfileSummary.SetTx(&tx)
	getProfileSummary.SetRegistry(&reg)
	getProfileSummary.Tx()
	getProfileSummary.Registry()
	actualName := getProfileSummary.Name()

	if getProfileSummary.Name() != "GetProfileSummary" {
		t.Fatalf("[%s] Error: expected request name GetProfileSummary but received %v", t.Name(), actualName)
	}

	// invalid request
	err = getProfileSummary.ValidateRequest(model.RequestWrapper{})
	if err == nil {
		t.Fatalf("[%s] ID should be missing: %v", t.Name(), err)
	}

	// valid request
	validReq := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			"id": "550e8400-e29b-41d4-a716-446655440000",
		},
	}

	reqWrapper, err := getProfileSummary.CreateRequest(validReq)
	if err != nil {
		t.Fatalf("[%s] Error creating request: %v", t.Name(), err)
	}

	err = getProfileSummary.ValidateRequest(reqWrapper)
	if err != nil {
		t.Fatalf("[%s] Invalid request: %v", t.Name(), err)
	}

	res, err := getProfileSummary.Run(reqWrapper)
	if err != nil {
		t.Fatalf("[%s] Error running request: %v", t.Name(), err)
	}

	response, err := getProfileSummary.CreateResponse(res)
	if err != nil {
		t.Fatalf("[%s] Error while creating response: %v", t.Name(), err)
	}

	if response.StatusCode != 200 {
		t.Fatalf("[%s] Expected status code 200, got %d", t.Name(), response.StatusCode)
	}
}
