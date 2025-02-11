// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	testutils "tah/upt/source/ucp-common/src/utils/test"

	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestUnmergeProfiles(t *testing.T) {
	testName := testutils.GetTestName(TestUnmergeProfiles)
	t.Logf("[%s] Initialize Test Resources", testName)

	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)

	domainName := "domainName"
	domain := customerprofiles.Domain{Name: domainName}
	domains := []customerprofiles.Domain{domain}

	configTableName := "create-unmerge-test-" + core.GenerateUniqueId()
	configDbClient, err := db.InitWithNewTable(configTableName, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("Error initializing config db: %s", err)
	}
	t.Cleanup(func() { configDbClient.DeleteTable(configTableName) })
	configDbClient.WaitForTableCreation()

	reg := registry.Registry{
		Env: map[string]string{
			constants.ACCP_DOMAIN_NAME_ENV_VAR: domainName,
		},
		Accp:        customerprofiles.InitMock(&domain, &domains, &profilemodel.Profile{}, &[]profilemodel.Profile{}, &[]customerprofiles.ObjectMapping{}),
		AsyncLambda: lambda.InitMock("ucpAsync"),
		ConfigDB:    &configDbClient,
	}

	uc := NewUnmergeProfile()
	uc.SetRegistry(&reg)
	uc.SetTx(&tx)
	req := events.APIGatewayProxyRequest{
		Body: `{
			"unmergeRq": {
				"mergedIntoConnectID": "mergedIntoConnectID",
				"toUnmergeConnectID": "toUnmergeConnectID",
				"interactionToUnmerge": "interactionToUnmerge",
				"objectType": "hotel_booking"
			}
		}`,
	}

	uc.reg.DataAccessPermission = "*/*"
	reqWrapper, err := uc.CreateRequest(req)
	if err != nil {
		t.Errorf("[%s]  Error creating request: %v", testName, err)
	}
	uc.reg.SetAppAccessPermission(constant.UnmergeProfilePermission)
	err = uc.ValidateRequest(reqWrapper)
	if err != nil {
		t.Errorf("[%s] Invalid request: %v", testName, err)
	}

	res, err := uc.Run(reqWrapper)
	if err != nil {
		t.Errorf("[%s] Error running request: %v", testName, err)
	}

	_, err = uc.CreateResponse(res)
	if err != nil {
		t.Fatalf("[%s] Error creating response: %v", testName, err)
	}
}
