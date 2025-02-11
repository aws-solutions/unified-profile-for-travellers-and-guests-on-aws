// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"fmt"
	"log"
	"strings"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/lambda"

	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	limits "tah/upt/source/ucp-common/src/constant/limits"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func TestDomainCreationDeletion(t *testing.T) {
	t.Parallel()
	createUc, deleteUc, listUc := InitUseCase(t)

	log.Printf("Testing domain creation")
	testDomain := "ucp_component_" + strings.ToLower(core.GenerateUniqueId())
	badDomain := "UCPcomponent-test-domain-bad-" + time.Now().Format("15-04-05")
	goodReq := events.APIGatewayProxyRequest{
		Body: "{\"domain\": { \"customerProfileDomain\": \"" + testDomain + "\"}}",
	}
	wrapper, err := createUc.CreateRequest(goodReq)
	if err != nil {
		t.Errorf("Error creating UCP Request: %v", err)
	}
	err = createUc.ValidateRequest(wrapper)
	if err != nil {
		t.Errorf("Error validating UCP Request: %v", err)
	}
	badReq := events.APIGatewayProxyRequest{
		Body: "{\"domain\": { \"customerProfileDomain\": \"" + badDomain + "\"}}",
	}
	log.Printf("Validating bad domain doesn't create")
	badWrapper, err := createUc.CreateRequest(badReq)
	if err != nil {
		t.Errorf("Error creating UCP Request: %v", err)
	}
	err = createUc.ValidateRequest(badWrapper)
	if err == nil {
		t.Errorf("Error validating bad UCP Request, expected error")
	}

	_, err = createUc.Run(wrapper)
	if err != nil {
		t.Errorf("Error creating UCP domain: %v", err)
	}
	_, err = listUc.Run(wrapper)
	if err != nil {
		t.Errorf("Error listing UCP domains: %v", err)
	}
	log.Printf("Testing domain deletions")
	_, err = deleteUc.Run(wrapper)
	if err != nil {
		t.Errorf("Error deleting UCP domain: %v", err)
	}
}

type ValidationTests struct {
	Name            string
	ExistingDomains []string
	Input           model.RequestWrapper
	ExpectedError   string
}

var VALIDATION_TESTS = []ValidationTests{
	{
		Name: "long domain name",
		Input: model.RequestWrapper{
			Domain: model.Domain{Name: "domainname_domainname_domainname_domainname_domainname_domainname"},
		},
		ExpectedError: fmt.Sprintf("Domain name length should be between %d and %d chars", customerprofiles.MIN_DOMAIN_NAME_LENGTH, customerprofiles.MAX_DOMAIN_NAME_LENGTH),
	},
	{
		Name: "short domain name",
		Input: model.RequestWrapper{
			Domain: model.Domain{Name: "do"},
		},
		ExpectedError: fmt.Sprintf("Domain name length should be between %d and %d chars", customerprofiles.MIN_DOMAIN_NAME_LENGTH, customerprofiles.MAX_DOMAIN_NAME_LENGTH),
	},
	{
		Name: "number of domains exeeded",
		Input: model.RequestWrapper{
			Domain: model.Domain{Name: "valid_domain_name"},
		},
		ExistingDomains: []string{"domain1",
			"domain2",
			"domain3",
			"domain4",
			"domain5",
			"domain6",
			"domain7",
			"domain8",
			"domain9",
			"domain10",
			"domain11",
			"domain12",
			"domain13",
			"domain14",
			"domain15",
		},
		ExpectedError: fmt.Sprintf("limit of %d domains reached", limits.UPT_LIMIT_MAX_DOMAINS),
	},
	{
		Name: "Domain already exists",
		Input: model.RequestWrapper{
			Domain: model.Domain{Name: "valid_domain_name"},
		},
		ExistingDomains: []string{"valid_domain_name"},
		ExpectedError:   fmt.Sprintf("domain %s already exists", "valid_domain_name"),
	},
}

func TestDomainValidator(t *testing.T) {
	t.Parallel()
	createUc, _, _ := InitUseCase(t)
	log.Printf("Testing domain creation")

	for _, test := range VALIDATION_TESTS {
		log.Printf("Testing %v", test.Name)
		domains := []customerprofiles.Domain{}
		for _, domain := range test.ExistingDomains {
			domains = append(domains, customerprofiles.Domain{Name: domain, Tags: map[string]string{constant.DOMAIN_TAG_ENV_NAME: "e2e-env"}})
		}
		createUc.reg.Accp.(*customerprofiles.MockCustomerProfileLowCostConfig).Domains = &domains
		createUc.reg.SetAppAccessPermission(constant.CreateDomainPermission)
		err := createUc.ValidateRequest(test.Input)
		log.Printf("Validation Error: %v", err)
		if err == nil {
			t.Errorf("Error validating UCP Request, expected error")
			return
		}
		if err.Error() != test.ExpectedError {
			t.Errorf("Error validating UCP Request, expected error: %v, got: %v", test.ExpectedError, err.Error())
		}
	}
}

func InitUseCase(t *testing.T) (*CreateDomain, *DeleteDomain, *ListUcpDomains) {
	// Set up resources
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("unable to load configs: %v", err)
	}
	rootDir, err := config.AbsolutePathToProjectRootDir()
	if err != nil {
		t.Fatalf("unable to determine root dir")
	}
	glueSchemaPath := rootDir + "source/tah-common/"
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	asyncLambdaConfig := lambda.InitMock("ucpAsyncEnv")
	cognitoConfig := cognito.InitMock(nil, nil)

	domains := []customerprofiles.Domain{}
	accpConfig := customerprofiles.InitMock(nil, &domains, nil, nil, nil)
	configTableName := "dom_crud-test_" + strings.ToLower(core.GenerateUniqueId())
	configDbClient, err := db.InitWithNewTable(configTableName, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("[%s] error creating table: %v", t.Name(), err)
	}
	t.Cleanup(func() { // Clean up resources
		err = configDbClient.DeleteTable(configTableName)
		if err != nil {
			t.Errorf("Error deleting table %v", err)
		}
	})
	configDbClient.WaitForTableCreation()
	glueClient := glue.Init(envCfg.Region, "traveller_test", "", "")

	if err != nil {
		t.Errorf("Could not create config db client %v", err)
	}
	err = configDbClient.WaitForTableCreation()
	if err != nil {
		t.Errorf("Could not wait for table creation %v", err)
	}
	services := registry.ServiceHandlers{
		AsyncLambda: asyncLambdaConfig,
		Accp:        accpConfig,
		ConfigDB:    &configDbClient,
		Cognito:     cognitoConfig,
		Glue:        &glueClient,
	}
	reg := registry.NewRegistry(envCfg.Region, core.LogLevelDebug, services)
	reg.Env["KMS_KEY_PROFILE_DOMAIN"] = "kms_profile"
	reg.Env["ACCP_DOMAIN_DLQ"] = "accp_domain"
	reg.Env["LAMBDA_ENV"] = "e2e-env"
	reg.Env["MATCH_BUCKET_NAME"] = "match-bucket-name"
	reg.Env["CONNECT_PROFILE_SOURCE_BUCKET"] = "connect-source-bucket"
	reg.AddEnv("GLUE_SCHEMA_PATH", glueSchemaPath)

	createUc := NewCreateDomain()
	createUc.SetRegistry(&reg)
	createUc.SetTx(&tx)
	createUc.Tx()
	deleteUc := NewDeleteDomain()
	deleteUc.SetRegistry(&reg)
	deleteUc.SetTx(&tx)
	deleteUc.Tx()
	listUc := NewListUcpDomains()
	listUc.SetRegistry(&reg)
	listUc.SetTx(&tx)

	return createUc, deleteUc, listUc
}
