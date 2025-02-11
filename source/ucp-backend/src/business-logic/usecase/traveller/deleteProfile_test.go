// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"log"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestDeleteProfile(t *testing.T) {
	log.Printf("TestDeleteProfile: initialize test resources")
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	domain := customerprofiles.Domain{Name: "test_domain"}
	domains := []customerprofiles.Domain{domain}
	profile := profilemodel.Profile{}
	profiles := []profilemodel.Profile{}
	mappings := []customerprofiles.ObjectMapping{}
	var accp = customerprofiles.InitMock(&domain, &domains, &profile, &profiles, &mappings)
	reg := registry.Registry{Accp: accp}
	uc := NewDeleteProfile()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)
	rq := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			"id": "550e8400-e29b-41d4-a716-446655440000",
		},
	}
	invaldRq := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			"id": "invalid_id",
		},
	}
	tx.Debug("Api Gateway request", rq)
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "DeleteProfile", err0)
	}
	uc.reg.SetAppAccessPermission(constant.DeleteProfilePermission)

	err0 = uc.ValidateRequest(wrapper)
	if err0 == nil {
		t.Errorf("[%s] request should be denied if user does not have the right permissin group", "MergeProfile")
	}

	uc.reg.DataAccessPermission = "*/*"
	tx.Debug("Api Gateway invalid request", invaldRq)
	wrapper, err0 = uc.CreateRequest(invaldRq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "MergeProfile", err0)
	}
	err0 = uc.ValidateRequest(wrapper)
	if err0 == nil {
		t.Errorf("[%s] request should be rejected for invalid connect_id", "MergeProfile")
	}

	tx.Debug("Api Gateway request", rq)
	wrapper, err0 = uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "MergeProfile", err0)
	}
	err0 = uc.ValidateRequest(wrapper)
	if err0 != nil {
		t.Errorf("[%s] request should be accepted for authorized user and valid connect_id", "MergeProfile")
	}
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] Error running use case: %v", "DeleteProfile", err)
	}
	apiRes, err2 := uc.CreateResponse(rs)
	if err2 != nil {
		t.Errorf("[%s] Error creating response %v", "DeleteProfile", err2)
	}
	tx.Debug("Api Gateway response", apiRes)

	log.Printf("TestDeleteProfile: deleting test resources")
}
