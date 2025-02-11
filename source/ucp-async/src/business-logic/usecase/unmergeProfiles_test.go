// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/ucp-common/src/model/admin"
	"tah/upt/source/ucp-common/src/model/async"
	ucModel "tah/upt/source/ucp-common/src/model/async/usecase"
	"testing"
	"time"
)

func TestUnmergeProfiles(t *testing.T) {

	uc := InitUnmergeProfiles()
	if uc.Name() != "UnmergeProfiles" {
		t.Errorf("Expected UnmergeProfiles, got %v", uc.Name())
	}
	domainName := "my_test_domain_" + time.Now().Format("20060102150405")

	domain := customerprofiles.Domain{Name: domainName}
	domains := []customerprofiles.Domain{domain}
	var accp = customerprofiles.InitMock(&domain, &domains, &profilemodel.Profile{}, &[]profilemodel.Profile{}, &[]customerprofiles.ObjectMapping{})
	solutionsConfig := awssolutions.InitMock()
	services := async.Services{
		AccpConfig:      accp,
		SolutionsConfig: solutionsConfig,
	}

	unmergeRequest := admin.UnmergeRq{
		ToUnmergeConnectID:   "connectId1",
		MergedIntoConnectID:  "connectId2",
		InteractionToUnmerge: "interactionId1",
		InteractionType:      "air_booking",
		UnmergeContext: customerprofiles.ProfileMergeContext{
			Timestamp: time.Now(),
			MergeType: customerprofiles.MergeTypeUnmerge,
		},
	}

	payload := async.AsyncInvokePayload{
		EventID:       "test-event-id",
		Usecase:       async.USECASE_EMPTY_TABLE,
		TransactionID: "test-transaction-id",
		Body: ucModel.UnmergeProfilesBody{
			Domain: domainName,
			Rq:     unmergeRequest,
		},
	}

	err := uc.Execute(payload, services, core.NewTransaction(t.Name(), "", core.LogLevelDebug))
	if err != nil {
		t.Errorf("Error invoking usecase: %v", err)
	}

	if len(accp.Merged) != 4 || accp.Merged[0] != unmergeRequest.ToUnmergeConnectID || accp.Merged[1] != unmergeRequest.MergedIntoConnectID || accp.Merged[2] != unmergeRequest.InteractionToUnmerge || accp.Merged[3] != unmergeRequest.InteractionType {
		t.Error("Could not validate unmerge request")
	}
}
