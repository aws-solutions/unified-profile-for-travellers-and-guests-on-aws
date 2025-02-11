// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package testutils

import (
	"os"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/appregistry"
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/iam"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
)

/*Utility for testing purpose*/

// returns a test registry with dummy service wrapper config
func BuildTestRegistry(region string) registry.Registry {
	var appregistryClient = appregistry.Init(region, "", "")
	var iamClient = iam.Init("", "")
	var glueClient = glue.Init(region, "test_glue_db", "", "")
	var dbConfig = db.Init("TEST_TABLE", "TEST_PK", "TEST_SK", "", "")
	var cognitoClient = cognito.Init("test-user-pool", "", "", core.LogLevelDebug)
	profiles := customerprofiles.InitMockV2()
	return registry.NewRegistry(region, core.LogLevelDebug, registry.ServiceHandlers{
		AppRegistry: &appregistryClient,
		Iam:         &iamClient,
		Glue:        &glueClient,
		Accp:        profiles,
		ErrorDB:     &dbConfig,
		Cognito:     cognitoClient,
	})
}

func GetTestRegion() string {
	//getting region for local testing
	region := os.Getenv("UCP_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}
