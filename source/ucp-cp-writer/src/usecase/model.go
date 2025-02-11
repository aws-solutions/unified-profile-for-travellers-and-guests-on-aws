// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/sqs"
)

const (
	CPWriterRequestType_PutProfile       string = "put_profile"
	CPWriterRequestType_PutProfileObject string = "put_profile_object"
	CPWriterRequestType_DeleteProfile    string = "delete_profile"
)

type ServiceRegistry struct {
	Tx              core.Transaction
	CPConfig        customerprofiles.ICustomerProfileConfig
	DLQConfig       sqs.IConfig
	SolutionsConfig awssolutions.IConfig
	DbConfig        db.DBConfig
	CpQueue         sqs.IConfig
}

type PutProfileRq struct {
	Profile profilemodel.Profile `json:"profile"`
}

type PutProfileObjectRq struct {
	ProfileID     string                     `json:"profileId"`
	ProfileObject profilemodel.ProfileObject `json:"profileObject"`
}

type DeleteProfileRq struct {
	ProfileID  string `json:"profileId"`
	NumRetries int    `json:"numRetries"`
}
