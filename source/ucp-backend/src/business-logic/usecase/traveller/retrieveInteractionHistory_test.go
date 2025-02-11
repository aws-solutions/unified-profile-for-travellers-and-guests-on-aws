// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	testutils "tah/upt/source/ucp-common/src/utils/test"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestRetrieveInteractionHistory(t *testing.T) {
	testName := testutils.GetTestName(TestRetrieveInteractionHistory)
	t.Logf("[%s] Initialize Test Resources", testName)

	retrieveInteractionHistory := NewRetrieveInteractionHistory()
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	retrieveInteractionHistory.SetTx(&tx)

	domainName := "domainName"
	domain := customerprofiles.Domain{Name: domainName}
	domains := []customerprofiles.Domain{domain}
	profile := profilemodel.Profile{}
	profiles := []profilemodel.Profile{}
	mappings := []customerprofiles.ObjectMapping{}
	reg := registry.Registry{
		Env: map[string]string{
			constants.ACCP_DOMAIN_NAME_ENV_VAR: domainName,
		},
		Accp: customerprofiles.InitMock(&domain, &domains, &profile, &profiles, &mappings),
	}
	retrieveInteractionHistory.SetRegistry(&reg)

	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			"id": "mock-traveller",
		},
		QueryStringParameters: map[string]string{
			"objectType": "hotel_booking",
			"connectId":  "12345",
		},
		Headers: map[string]string{
			"customer-profiles-domain": domainName,
		},
	}

	reqWrapper, err := retrieveInteractionHistory.CreateRequest(req)
	if err != nil {
		t.Errorf("[TestRetrieveInteractionHistory]  Error creating request: %v", err)
	}

	err = retrieveInteractionHistory.ValidateRequest(reqWrapper)
	if err != nil {
		t.Errorf("[TestRetrieveInteractionHistory] Invalid request: %v", err)
	}

	res, err := retrieveInteractionHistory.Run(reqWrapper)
	if err != nil {
		t.Errorf("[TestRetrieveInteractionHistory] Error running request: %v", err)
	}

	_, err = retrieveInteractionHistory.CreateResponse(res)
	if err != nil {
		t.Fatalf("[%s] Error creating response: %v", testName, err)
	}
}
