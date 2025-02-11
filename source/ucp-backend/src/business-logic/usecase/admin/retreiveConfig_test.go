// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"log"
	"testing"
	"time"

	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constants "tah/upt/source/ucp-common/src/constant/admin"

	"github.com/aws/aws-lambda-go/events"
)

func TestRetreiveConfig(t *testing.T) {
	log.Printf("TestRetreiveConfig: initialize test resources")
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	domain := customerprofiles.Domain{Name: "test_domain"}
	domains := []customerprofiles.Domain{domain}
	profile := profilemodel.Profile{}
	profiles := []profilemodel.Profile{}
	mappings := []customerprofiles.ObjectMapping{
		{Name: constants.ACCP_RECORD_AIR_BOOKING},
		{Name: constants.ACCP_RECORD_EMAIL_HISTORY},
		{Name: constants.ACCP_RECORD_PHONE_HISTORY},
		{Name: constants.ACCP_RECORD_AIR_LOYALTY},
		{Name: constants.ACCP_RECORD_LOYALTY_TX},
		{Name: constants.ACCP_RECORD_CLICKSTREAM},
		{Name: constants.ACCP_RECORD_GUEST_PROFILE},
		{Name: constants.ACCP_RECORD_HOTEL_LOYALTY},
		{Name: constants.ACCP_RECORD_HOTEL_BOOKING},
		{Name: constants.ACCP_RECORD_PAX_PROFILE},
		{Name: constants.ACCP_RECORD_HOTEL_STAY_MAPPING},
		{Name: constants.ACCP_RECORD_CSI},
		{Name: constants.ACCP_RECORD_ANCILLARY},
	}
	// Async services setup
	errorTableName := "update-domain-test-" + time.Now().Format("2006-01-02-15-04-05")
	dbClient := db.Init(errorTableName, "pk", "sk", "", "")
	err := dbClient.CreateTableWithOptions(errorTableName, "pk", "sk", db.TableOptions{
		StreamEnabled: true,
	})
	if err != nil {
		t.Errorf("Could not create config db client %v", err)
	}
	err = dbClient.WaitForTableCreation()
	if err != nil {
		t.Errorf("Error waiting for config table creation: %v", err)
	}

	var accp = customerprofiles.InitMock(&domain, &domains, &profile, &profiles, &mappings)
	reg := registry.Registry{Accp: accp,
		RetryLambda: lambda.InitMock("ucpRetryEnv"),
		ErrorDB:     &dbClient}
	uc := NewRetreiveConfig()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)
	rq := events.APIGatewayProxyRequest{}
	tx.Debug("Api Gateway request", rq)
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "RetreiveConfig", err0)
	}
	err0 = uc.ValidateRequest(wrapper)
	if err0 != nil {
		t.Errorf("[%s] Error validating request request %v", "RetreiveConfig", err0)
	}
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] Error running use case: %v", "RetreiveConfig", err)
	}
	if len(rs.UCPConfig.Domains) != 1 {
		t.Errorf("[%s] Retrieve config should return one domain and not %v", "RetreiveConfig", len(rs.UCPConfig.Domains))
	}
	if !rs.UCPConfig.Domains[0].NeedsMappingUpdate {
		t.Errorf("[%s] Domain should have NeedsMappingUpdate set to true", "RetreiveConfig")
	}

	apiRes, err2 := uc.CreateResponse(rs)
	if err2 != nil {
		t.Errorf("[%s] Error creating response %v", "RetreiveConfig", err2)
	}
	tx.Debug("Api Gateway response", apiRes)

	log.Printf("TestRetreiveConfig: deleting test resources")
	err = dbClient.DeleteTable(errorTableName)
	if err != nil {
		t.Errorf("Error deleting table %v", err)
	}
}
