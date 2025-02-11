// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"

	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/ecs"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/tah-core/sqs"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	common "tah/upt/source/ucp-common/src/constant/admin"
	accpMappings "tah/upt/source/ucp-common/src/model/accp-mappings"
	"tah/upt/source/ucp-common/src/utils/config"
	testutils "tah/upt/source/ucp-common/src/utils/test"
	"tah/upt/source/ucp-common/src/utils/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/stretchr/testify/mock"
)

type MockSsm struct {
	mock.Mock
}

func (m *MockSsm) GetParametersByPath(ctx context.Context, namespace string) (map[string]string, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(map[string]string), args.Error(1)
}

type MockEcs struct {
	mock.Mock
}

func (m *MockEcs) ListTasks(ctx context.Context, taskConfig ecs.TaskConfig, startedBy, desiredStatus *string) ([]string, error) {
	args := m.Called(ctx, taskConfig, startedBy, desiredStatus)
	return args.Get(0).([]string), args.Error(1)
}

func TestRuleSets(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	testName := "TestRuleSets"
	log.Printf("[%s] initialize test resources", testName)
	dynamoConfigTableName := "rule-set-" + strings.ToLower(core.GenerateUniqueId())
	domain := "rulesets_data_" + strings.ToLower(core.GenerateUniqueId())

	dynamo_pk := "pk"
	dynamo_sk := "sk"
	dynamoCfg, err := db.InitWithNewTable(dynamoConfigTableName, dynamo_pk, dynamo_sk, "", "")
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoConfigTableName)
		if err != nil {
			t.Errorf("Error deleting table %v", err)
		}
	})
	dynamoCfg.WaitForTableCreation()

	configTableName := "config-" + strings.ToLower(core.GenerateUniqueId())
	configDbClient, err := db.InitWithNewTable(configTableName, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = configDbClient.DeleteTable(configTableName)
		if err != nil {
			t.Errorf("Error deleting table %v", err)
		}
	})
	configDbClient.WaitForTableCreation()

	kinesisConfig, err := kinesis.InitAndCreate(domain, envCfg.Region, "", "", core.LogLevelDebug)
	if err != nil {
		t.Fatalf("[%s] Error creating kinesis config %v", testName, err)
	}
	t.Cleanup(func() {
		err = kinesisConfig.Delete(domain)
		if err != nil {
			t.Errorf("[%s] Error deleting stream %v", testName, err)
		}
	})
	mergeQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = mergeQueueClient.CreateRandom("mergerRuleSetsTest")
	if err != nil {
		t.Errorf("error creating merge queue: %v", err)
	}
	t.Cleanup(func() {
		err = mergeQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting merge queue: %v", err)
		}
	})
	cpWriterQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = cpWriterQueueClient.CreateRandom("cpWriterRuleSetsTest")
	if err != nil {
		t.Errorf("error creating cp writer queue: %v", err)
	}
	t.Cleanup(func() {
		err = cpWriterQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting cp writer queue: %v", err)
		}
	})
	initParams := testutils.ProfileStorageParams{
		LcsConfigTable:      &dynamoCfg,
		LcsKinesisStream:    kinesisConfig,
		MergeQueueClient:    &mergeQueueClient,
		CPWriterQueueClient: &cpWriterQueueClient,
	}
	cfg, err := testutils.InitProfileStorage(initParams)
	if err != nil {
		t.Errorf("Error creating customer profiles %v", err)
	}

	// Create domain
	options := customerprofiles.DomainOptions{
		AiIdResolutionOn:       false,
		RuleBaseIdResolutionOn: true,
	}
	err = cfg.CreateDomainWithQueue(domain, "", map[string]string{"envName": "travelerTest"}, "", "", options)
	if err != nil {
		t.Fatalf("[%s] Error creating domain %v", t.Name(), err)
	}

	// Defer deleting domain
	t.Cleanup(func() {
		err = cfg.DeleteDomain()
		if err != nil {
			t.Errorf("[%s] Error deleting domain", t.Name())
		}
	})
	cfg.SetDomain(domain)
	tx := core.NewTransaction(testName, "", core.LogLevelDebug)
	cfg.SetTx(tx)
	err = createMappings(cfg, tx)
	if err != nil {
		t.Errorf("[%s] Error creating mappings %v", t.Name(), err)
	}

	mockSsm := MockSsm{}
	defer mockSsm.AssertExpectations(t)
	mockEcs := MockEcs{}
	defer mockEcs.AssertExpectations(t)

	ssmParamNamespace := "my-namespace"
	mockSsm.On("GetParametersByPath", context.TODO(), ssmParamNamespace).Return(map[string]string{
		"VpcSubnets":             "",
		"VpcSecurityGroup":       "",
		"IdResClusterArn":        "",
		"IdResTaskDefinitionArn": "",
		"IdResContainerName":     "",
	}, nil)
	var startedBy *string = nil
	mockEcs.On("ListTasks", context.TODO(), mock.Anything, startedBy, aws.String(string(awsecstypes.DesiredStatusRunning))).
		Return([]string{}, nil)

	// Create rule sets
	reg := registry.Registry{
		Accp:        cfg,
		AsyncLambda: lambda.InitMock("async"),
		ConfigDB:    &configDbClient,
		Ssm:         &mockSsm,
		Ecs:         &mockEcs,
		Env:         map[string]string{"ACCP_DOMAIN_NAME": "test_domain", "SSM_PARAM_NAMESPACE": ssmParamNamespace},
	}

	saveUc := NewSaveIdResRuleSet()
	saveUc.SetTx(&tx)
	saveUc.SetRegistry(&reg)

	activateUc := NewActivateIdResRuleSet()
	activateUc.SetTx(&tx)
	activateUc.SetRegistry(&reg)

	listUc := NewListIdResRuleSets()
	listUc.SetTx(&tx)
	listUc.SetRegistry(&reg)

	testMatchCondition := customerprofiles.Condition{
		Index:               1,
		ConditionType:       customerprofiles.CONDITION_TYPE_MATCH,
		IncomingObjectType:  "air_booking",
		IncomingObjectField: "to",
		ExistingObjectType:  "email_history",
		ExistingObjectField: "address",
		IndexNormalization: customerprofiles.IndexNormalizationSettings{
			Lowercase: true,
			Trim:      true,
		},
	}

	testRules := []customerprofiles.Rule{
		{
			Index:       1,
			Name:        "test_rule_1",
			Description: "test_rule_1",
			Conditions:  []customerprofiles.Condition{testMatchCondition},
		},
		{
			Index:       2,
			Name:        "test_rule_2",
			Description: "test_rule_2",
			Conditions:  []customerprofiles.Condition{testMatchCondition},
		},
		{
			Index:       3,
			Name:        "test_rule_3",
			Description: "test_rule_3",
			Conditions:  []customerprofiles.Condition{testMatchCondition},
		},
	}

	testSaveRuleSetBody, err := json.Marshal(
		model.RequestWrapper{
			SaveRuleSet: model.SaveRuleSetRq{
				Rules: testRules,
			},
		},
	)
	if err != nil {
		t.Errorf("[%s] Error Marshalling Request: %v", t.Name(), err)
	}

	// Save Request
	saveRq := events.APIGatewayProxyRequest{
		Body: string(testSaveRuleSetBody),
	}
	tx.Debug("Save Rule Set Request", saveRq)

	saveWrapper, err := saveUc.CreateRequest(saveRq)
	if err != nil {
		t.Errorf("[%s] Error creating save rule set request %v", t.Name(), err)
	}
	saveUc.reg.SetAppAccessPermission(common.SaveRuleSetPermission)

	err = saveUc.ValidateRequest(saveWrapper)
	if err != nil {
		t.Errorf("[%s] Error validating save rule set request %v", t.Name(), err)
	}
	saveResponseWrapper, err := saveUc.Run(saveWrapper)
	if err != nil {
		t.Errorf("[%s] Error running save rule set use case: %v", t.Name(), err)
	}
	saveResponse, err := saveUc.CreateResponse(saveResponseWrapper)
	if err != nil {
		t.Errorf("[%s] Error creating save rule set response %v", t.Name(), err)
	}
	tx.Debug("Save Rule Set Response", saveResponse)

	// List Request (Draft)
	emptyRq := events.APIGatewayProxyRequest{}
	listWrapper, err := listUc.CreateRequest(emptyRq)
	if err != nil {
		t.Errorf("[%s] Error creating list rule set request %v", t.Name(), err)
	}
	err = listUc.ValidateRequest(listWrapper)
	if err != nil {
		t.Errorf("[%s] Error validating list rule set request %v", t.Name(), err)
	}
	listResponseWrapper, err := listUc.Run(listWrapper)
	if err != nil {
		t.Errorf("[%s] Error running list rule set use case: %v", t.Name(), err)
	}
	listResponse, err := listUc.CreateResponse(listResponseWrapper)
	if err != nil {
		t.Errorf("[%s] Error creating list rule set response %v", t.Name(), err)
	}
	tx.Debug("List Rule Set Response", listResponse)

	// Validate List Response
	parsedListResponse := model.ResponseWrapper{}
	err = json.Unmarshal([]byte(listResponse.Body), &parsedListResponse)
	if err != nil {
		t.Errorf("[%s] Error parsing list rule set response %v", t.Name(), err)
	}
	listRuleSets := utils.SafeDereference(parsedListResponse.RuleSets)
	if len(listRuleSets) != 2 {
		t.Errorf("[%s] Returned rule set should have length of 2", t.Name())
	}
	if listRuleSets[0].Name != "active" || listRuleSets[1].Name != "draft" {
		t.Errorf("[%s] returned response is in the wrong order", t.Name())
	}
	if len(listRuleSets[0].Rules) != 0 || len(listRuleSets[1].Rules) != 3 {
		t.Errorf("[%s] returned response has the wrong number of rules", t.Name())
	}

	// Activate Request
	activateWrapper, err := activateUc.CreateRequest(emptyRq)
	if err != nil {
		t.Errorf("[%s] Error creating activate rule set request %v", t.Name(), err)
	}
	activateUc.reg.SetAppAccessPermission(common.ActivateRuleSetPermission)
	err = activateUc.ValidateRequest(activateWrapper)
	if err != nil {
		t.Errorf("[%s] Error validating activate rule set request %v", t.Name(), err)
	}
	activateResponseWrapper, err := activateUc.Run(activateWrapper)
	if err != nil {
		t.Errorf("[%s] Error running activate rule set use case: %v", t.Name(), err)
	}
	activateResponse, err := activateUc.CreateResponse(activateResponseWrapper)
	if err != nil {
		t.Errorf("[%s] Error creating activate rule set response %v", t.Name(), err)
	}
	tx.Debug("Activate Rule Set Response", activateResponse)

	// List Request (Active)
	listWrapper1, err := listUc.CreateRequest(emptyRq)
	if err != nil {
		t.Errorf("[%s] Error creating list rule set request %v", t.Name(), err)
	}

	err = listUc.ValidateRequest(listWrapper1)
	if err != nil {
		t.Errorf("[%s] Error validating list rule set request %v", t.Name(), err)
	}
	listResponseWrapper1, err := listUc.Run(listWrapper1)
	if err != nil {
		t.Errorf("[%s] Error running list rule set use case: %v", t.Name(), err)
	}
	listResponse1, err := listUc.CreateResponse(listResponseWrapper1)
	if err != nil {
		t.Errorf("[%s] Error creating list rule set response %v", t.Name(), err)
	}
	tx.Debug("List Rule Set Response", listResponse1)

	// Validate Response
	parsedListResponse1 := model.ResponseWrapper{}
	err = json.Unmarshal([]byte(listResponse1.Body), &parsedListResponse1)
	if err != nil {
		t.Errorf("[%s] Error parsing list rule set response %v", t.Name(), err)
	}
	listRuleSets1 := utils.SafeDereference(parsedListResponse1.RuleSets)

	if len(listRuleSets1) != 2 {
		t.Errorf("[%s] Returned rule set should have length of 2", t.Name())
	}
	if listRuleSets1[0].Name != "active" || listRuleSets1[1].Name != "draft" {
		t.Errorf("[%s] returned response is in the wrong order", t.Name())
	}
	if len(listRuleSets1[0].Rules) != 3 || len(listRuleSets1[1].Rules) != 0 {
		t.Errorf("[%s] returned response has the wrong number of rules", t.Name())
	}
}

func TestRuleSetsCache(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	testName := "TestRuleSets"
	log.Printf("[%s] initialize test resources", testName)
	dynamoConfigTableName := "cache-rule-set-" + strings.ToLower(core.GenerateUniqueId())
	domain := "rulesets_data_" + strings.ToLower(core.GenerateUniqueId())

	dynamo_pk := "pk"
	dynamo_sk := "sk"
	dynamoCfg, err := db.InitWithNewTable(dynamoConfigTableName, dynamo_pk, dynamo_sk, "", "")
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoConfigTableName)
		if err != nil {
			t.Errorf("Error deleting table %v", err)
		}
	})
	dynamoCfg.WaitForTableCreation()

	kinesisConfig, err := kinesis.InitAndCreate(domain, envCfg.Region, "", "", core.LogLevelDebug)
	if err != nil {
		t.Fatalf("[%s] Error creating kinesis config %v", testName, err)
	}
	t.Cleanup(func() {
		err = kinesisConfig.Delete(domain)
		if err != nil {
			t.Errorf("[%s] Error deleting stream %v", testName, err)
		}
	})
	mergeQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = mergeQueueClient.CreateRandom("mergerRuleSetsTest")
	if err != nil {
		t.Errorf("error creating merge queue: %v", err)
	}
	t.Cleanup(func() {
		err = mergeQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting merge queue: %v", err)
		}
	})
	cpWriterQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = cpWriterQueueClient.CreateRandom("cpWriterRuleSetsTest")
	if err != nil {
		t.Errorf("error creating cp writer queue: %v", err)
	}
	t.Cleanup(func() {
		err = cpWriterQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting cp writer queue: %v", err)
		}
	})
	initParams := testutils.ProfileStorageParams{
		LcsConfigTable:      &dynamoCfg,
		LcsKinesisStream:    kinesisConfig,
		MergeQueueClient:    &mergeQueueClient,
		CPWriterQueueClient: &cpWriterQueueClient,
	}
	cfg, err := testutils.InitProfileStorage(initParams)
	if err != nil {
		t.Errorf("Error creating customer profiles %v", err)
	}

	// Create domain
	options := customerprofiles.DomainOptions{
		AiIdResolutionOn:       false,
		RuleBaseIdResolutionOn: true,
	}
	err = cfg.CreateDomainWithQueue(domain, "", map[string]string{"envName": "travelerTest"}, "", "", options)
	if err != nil {
		t.Errorf("[%s] Error creating domain %v", t.Name(), err)
	}

	// Defer deleting domain
	t.Cleanup(func() {
		err = cfg.DeleteDomain()
		if err != nil {
			t.Errorf("[%s] Error deleting domain", t.Name())
		}
	})
	cfg.SetDomain(domain)
	tx := core.NewTransaction(testName, "", core.LogLevelDebug)
	cfg.SetTx(tx)
	err = createMappings(cfg, tx)
	if err != nil {
		t.Errorf("[%s] Error creating mappings %v", t.Name(), err)
	}

	// Create rule sets
	reg := registry.Registry{Accp: cfg, Env: map[string]string{"ACCP_DOMAIN_NAME": "test_domain"}}

	saveUc := NewSaveCacheRuleSet()
	saveUc.SetTx(&tx)
	saveUc.SetRegistry(&reg)

	activateUc := NewActivateCacheRuleSet()
	activateUc.SetTx(&tx)
	activateUc.SetRegistry(&reg)

	listUc := NewListCacheRuleSets()
	listUc.SetTx(&tx)
	listUc.SetRegistry(&reg)

	testSkipCondition := customerprofiles.Condition{
		Index:                1,
		ConditionType:        customerprofiles.CONDITION_TYPE_SKIP,
		IncomingObjectType:   "air_booking",
		IncomingObjectField:  "to",
		IncomingObjectType2:  "email_history",
		IncomingObjectField2: "address",
		Op:                   customerprofiles.RULE_OP_EQUALS,
		IndexNormalization: customerprofiles.IndexNormalizationSettings{
			Lowercase: true,
			Trim:      true,
		},
	}

	testRules := []customerprofiles.Rule{
		{
			Index:       1,
			Name:        "test_rule_1",
			Description: "test_rule_1",
			Conditions:  []customerprofiles.Condition{testSkipCondition},
		},
		{
			Index:       2,
			Name:        "test_rule_2",
			Description: "test_rule_2",
			Conditions:  []customerprofiles.Condition{testSkipCondition},
		},
		{
			Index:       3,
			Name:        "test_rule_3",
			Description: "test_rule_3",
			Conditions:  []customerprofiles.Condition{testSkipCondition},
		},
	}

	testSaveRuleSetBody, err := json.Marshal(
		model.RequestWrapper{
			SaveRuleSet: model.SaveRuleSetRq{
				Rules: testRules,
			},
		},
	)
	if err != nil {
		t.Errorf("[%s] Error Marshalling Request: %v", t.Name(), err)
	}

	// Save Request
	saveRq := events.APIGatewayProxyRequest{
		Body: string(testSaveRuleSetBody),
	}
	tx.Debug("Save Rule Set Request", saveRq)

	saveWrapper, err := saveUc.CreateRequest(saveRq)
	if err != nil {
		t.Errorf("[%s] Error creating save rule set request %v", t.Name(), err)
	}

	saveUc.reg.SetAppAccessPermission(common.SaveRuleSetPermission)
	err = saveUc.ValidateRequest(saveWrapper)
	if err != nil {
		t.Errorf("[%s] Error validating save rule set request %v", t.Name(), err)
	}
	saveResponseWrapper, err := saveUc.Run(saveWrapper)
	if err != nil {
		t.Errorf("[%s] Error running save rule set use case: %v", t.Name(), err)
	}
	saveResponse, err := saveUc.CreateResponse(saveResponseWrapper)
	if err != nil {
		t.Errorf("[%s] Error creating save rule set response %v", t.Name(), err)
	}
	tx.Debug("Save Rule Set Response", saveResponse)

	// List Request (Draft)
	emptyRq := events.APIGatewayProxyRequest{}
	listWrapper, err := listUc.CreateRequest(emptyRq)
	if err != nil {
		t.Errorf("[%s] Error creating list rule set request %v", t.Name(), err)
	}
	err = listUc.ValidateRequest(listWrapper)
	if err != nil {
		t.Errorf("[%s] Error validating list rule set request %v", t.Name(), err)
	}
	listResponseWrapper, err := listUc.Run(listWrapper)
	if err != nil {
		t.Errorf("[%s] Error running list rule set use case: %v", t.Name(), err)
	}
	listResponse, err := listUc.CreateResponse(listResponseWrapper)
	if err != nil {
		t.Errorf("[%s] Error creating list rule set response %v", t.Name(), err)
	}
	tx.Debug("List Rule Set Response", listResponse)

	// Validate List Response
	parsedListResponse := model.ResponseWrapper{}
	err = json.Unmarshal([]byte(listResponse.Body), &parsedListResponse)
	if err != nil {
		t.Errorf("[%s] Error parsing list rule set response %v", t.Name(), err)
	}
	listRuleSets := utils.SafeDereference(parsedListResponse.RuleSets)
	if len(listRuleSets) != 2 {
		t.Errorf("[%s] Returned rule set should have length of 2", t.Name())
	}
	if listRuleSets[0].Name != "active" || listRuleSets[1].Name != "draft" {
		t.Errorf("[%s] returned response is in the wrong order", t.Name())
	}
	if len(listRuleSets[0].Rules) != 0 || len(listRuleSets[1].Rules) != 3 {
		t.Errorf("[%s] returned response has the wrong number of rules", t.Name())
	}

	// Activate Request
	activateWrapper, err := activateUc.CreateRequest(emptyRq)
	if err != nil {
		t.Errorf("[%s] Error creating activate rule set request %v", t.Name(), err)
	}
	err = activateUc.ValidateRequest(activateWrapper)
	if err != nil {
		t.Errorf("[%s] Error validating activate rule set request %v", t.Name(), err)
	}
	activateResponseWrapper, err := activateUc.Run(activateWrapper)
	if err != nil {
		t.Errorf("[%s] Error running activate rule set use case: %v", t.Name(), err)
	}
	activateResponse, err := activateUc.CreateResponse(activateResponseWrapper)
	if err != nil {
		t.Errorf("[%s] Error creating activate rule set response %v", t.Name(), err)
	}
	tx.Debug("Activate Rule Set Response", activateResponse)

	// List Request (Active)
	listWrapper1, err := listUc.CreateRequest(emptyRq)
	if err != nil {
		t.Errorf("[%s] Error creating list rule set request %v", t.Name(), err)
	}
	err = listUc.ValidateRequest(listWrapper1)
	if err != nil {
		t.Errorf("[%s] Error validating list rule set request %v", t.Name(), err)
	}
	listResponseWrapper1, err := listUc.Run(listWrapper1)
	if err != nil {
		t.Errorf("[%s] Error running list rule set use case: %v", t.Name(), err)
	}
	listResponse1, err := listUc.CreateResponse(listResponseWrapper1)
	if err != nil {
		t.Errorf("[%s] Error creating list rule set response %v", t.Name(), err)
	}
	tx.Debug("List Rule Set Response", listResponse1)

	// Validate Response
	parsedListResponse1 := model.ResponseWrapper{}
	err = json.Unmarshal([]byte(listResponse1.Body), &parsedListResponse1)
	if err != nil {
		t.Errorf("[%s] Error parsing list rule set response %v", t.Name(), err)
	}
	listRuleSets1 := utils.SafeDereference(parsedListResponse1.RuleSets)
	if len(listRuleSets1) != 2 {
		t.Errorf("[%s] Returned rule set should have length of 2", t.Name())
	}
	if listRuleSets1[0].Name != "active" || listRuleSets1[1].Name != "draft" {
		t.Errorf("[%s] returned response is in the wrong order", t.Name())
	}
	if len(listRuleSets1[0].Rules) != 3 || len(listRuleSets1[1].Rules) != 0 {
		t.Errorf("[%s] returned response has the wrong number of rules", t.Name())
	}
}

func createMappings(services customerprofiles.ICustomerProfileLowCostConfig, tx core.Transaction) error {
	for _, businessObject := range common.ACCP_RECORDS {
		accpRecName := businessObject.Name
		tx.Debug("[CreateUcpDomain] Creating mapping for %s", accpRecName)
		objMapping := accpMappings.ACCP_OBJECT_MAPPINGS[accpRecName]()
		err := services.CreateMapping(accpRecName, "Primary Mapping for the "+accpRecName+" object", objMapping.Fields)
		if err != nil {
			tx.Error("[CreateUcpDomain] Error creating Mapping: %s. deleting domain", err)
			err2 := services.DeleteDomain()
			if err2 != nil {
				tx.Error("[CreateUcpDomain][warning] Error cleaning up domain after failed mapping creation %v", err2)
			}
			return err
		}
	}
	return nil
}
