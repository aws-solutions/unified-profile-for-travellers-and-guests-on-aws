// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-common/src/model/async"
)

type Usecase interface {
	Name() string
	Execute(payload model.AsyncInvokePayload, services model.Services, tx core.Transaction) error
}

func GetUsecases() map[string]Usecase {
	usecases := map[string]Usecase{
		model.USECASE_CREATE_DOMAIN:         InitCreateDomain(),
		model.USECASE_DELETE_DOMAIN:         InitDeleteDomain(),
		model.USECASE_EMPTY_TABLE:           InitEmptyDynamoTable(),
		model.USECASE_MERGE_PROFILES:        InitMergeProfiles(),
		model.USECASE_UNMERGE_PROFILES:      InitUnmergeProfiles(),
		model.USECASE_CREATE_PRIVACY_SEARCH: InitCreatePrivacySearch(),
		model.USECASE_PURGE_PROFILE_DATA:    InitPurgeProfileData(),
		model.USECASE_START_BATCH_ID_RES:    InitStartBatchIdRes(),
		model.USECASE_REBUILD_CACHE:         InitRebuildCache(),
	}

	return usecases
}
