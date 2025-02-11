// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/ucp-common/src/utils/config"
	"tah/upt/source/ucp-cp-writer/src/usecase"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestCustomerProfilesIntegration(t *testing.T) {
	t.Parallel()

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}
	streamName := "test-cp-output-stream" + core.GenerateUniqueId()
	cfg, err := kinesis.InitAndCreate(streamName, envCfg.Region, "", "", core.LogLevelDebug)
	if err != nil {
		t.Fatalf("error setting up event stream: %v", err)
	}
	t.Cleanup(func() {
		err = cfg.Delete(streamName)
		if err != nil {
			t.Errorf("[%v] DeleteTable failed: %s", t.Name(), err)
		}
	})
	_, err = cfg.WaitForStreamCreation(20)
	if err != nil {
		t.Fatalf("error waiting for stream creation: %v", err)
	}
	stream, err := cfg.Describe()
	if err != nil {
		t.Fatalf("error describing event stream: %v", err)
	}

	// LCS setup
	domainOptions := DomainOptions{
		CustomerProfileMode:    aws.Bool(true),
		DynamoMode:             aws.Bool(true),
		CpExportDestinationArn: stream.Arn,
	}
	lcs, _ := setupAndRegisterCleanup(t, setupOptions{domainOptions: domainOptions, domainNamePrefix: "cp_integration"})

	// Validate control plane
	domain, err := lcs.CustomerProfileConfig.GetDomain()
	if err != nil {
		t.Fatalf("[%s] error getting cp domain %v", t.Name(), err)
	}
	if domain.Name != lcs.DomainName {
		t.Fatalf("[%s] expected domain name %s, got %s", t.Name(), lcs.DomainName, domain.Name)
	}
	mappings, err := lcs.CustomerProfileConfig.GetMappings()
	if err != nil {
		t.Fatalf("[%s] error getting cp mappings %v", t.Name(), err)
	}
	if len(mappings) != 2 { // profile + booking
		t.Fatalf("[%s] expected 2 mappings, got %d", t.Name(), len(mappings))
	}

	// Validate cache rule set is not cached
	_, found := lcs.Config.retrieveFromCache(lcs.DomainName, controlPlaneCacheType_ActiveCacheRuleSet)
	if found {
		t.Fatalf("[%s] expected active rule set to not be cached", t.Name())
	}
	// Retrieve rule set (should be empty at this point)
	_, err = lcs.Config.getRuleSet(lcs.DomainName, RULE_SET_TYPE_ACTIVE, RULE_SK_CACHE_PREFIX)
	if err != nil {
		t.Fatalf("[%s] error getting active rule set %v", t.Name(), err)
	}
	// Validate empty cache rule set is cached
	cachedCacheRuleSet, found := lcs.Config.retrieveFromCache(lcs.DomainName, controlPlaneCacheType_ActiveCacheRuleSet)
	if !found {
		t.Fatalf("[%s] expected active rule set to be cached", t.Name())
	}
	if casted, ok := cachedCacheRuleSet.(RuleSet); ok {
		if len(casted.Rules) != 0 {
			t.Fatalf("[%s] expected active rule set to be empty, got %v", t.Name(), casted.Rules)
		}
	} else {
		t.Fatalf("[%s] expected active rule set to be empty, got %v", t.Name(), cachedCacheRuleSet)
	}

	// Add skip condition
	ruleSet1 := RuleSet{
		Rules: []Rule{
			{
				Index:       0,
				Name:        "rule1",
				Description: "skip condition",
				Conditions: []Condition{
					{
						Index:               0,
						ConditionType:       CONDITION_TYPE_SKIP,
						IncomingObjectType:  PROFILE_OBJECT_TYPE_NAME,
						IncomingObjectField: RESERVED_FIELD_LAST_NAME,
						Op:                  RULE_OP_EQUALS_VALUE,
						SkipConditionValue:  Value{ValueType: CONDITION_VALUE_TYPE_STRING, StringValue: ""},
						IndexNormalization:  NORMALIZATION_SETTINGS_DEFAULT,
					},
				},
			},
		},
	}
	err = lcs.SaveCacheRuleSet(ruleSet1.Rules)
	if err != nil {
		t.Fatalf("[%s] error saving ruleset: %v", t.Name(), err)
	}
	_, err = lcs.ListCacheRuleSets(true)
	if err != nil {
		t.Fatalf("[%s] error listing ruleset: %v", t.Name(), err)
	}
	err = lcs.ActivateCacheRuleSet()
	if err != nil {
		t.Fatalf("[%s] error activating ruleset: %v", t.Name(), err)
	}

	// Insert profile objects
	lastUpdated := time.Now().UTC()
	objects := []genericTestProfileObject{
		{
			ObjectID:    "obj_1",
			TravelerID:  "traveler_1",
			LastName:    "Smith",
			Email:       "cpintegration@example.com",
			LastUpdated: lastUpdated,
		},
		{
			ObjectID:    "obj_2",
			TravelerID:  "traveler_2",
			LastName:    "Brown",
			LastUpdated: lastUpdated,
		},
		{
			ObjectID:    "obj_3",
			TravelerID:  "anonymous",
			LastUpdated: lastUpdated,
			// no last name, should be skipped
		},
	}
	for _, obj := range objects {
		objJson, _ := json.Marshal(obj)
		err = lcs.PutProfileObject(string(objJson), genericTestObjectType_Booking)
		if err != nil {
			t.Fatalf("[%s] error putting profile object %v", t.Name(), err)
		}
	}

	// Validate cache rules are now available in the cache
	cachedCacheRuleSet, found = lcs.Config.retrieveFromCache(lcs.DomainName, controlPlaneCacheType_ActiveCacheRuleSet)
	if !found {
		t.Fatalf("[%s] expected active rule set to be cached", t.Name())
	}
	if casted, ok := cachedCacheRuleSet.(RuleSet); ok {
		if len(casted.Rules) != len(ruleSet1.Rules) {
			t.Fatalf("[%s] expected active rule set to have 1 rule, got %v", t.Name(), casted.Rules)
		}
	} else {
		t.Fatalf("[%s] expected active rule set to be RuleSet, got %v", t.Name(), cachedCacheRuleSet)
	}

	// Validate objects were properly sent to CP Writer queue
	uptId1, err := WaitForProfileCreationByProfileId("traveler_1", "profile_id", &lcs)
	if err != nil {
		t.Fatalf("[%s] error waiting for profile creation %v", t.Name(), err)
	}
	uptId2, err := WaitForProfileCreationByProfileId("traveler_2", "profile_id", &lcs)
	if err != nil {
		t.Fatalf("[%s] error waiting for profile creation %v", t.Name(), err)
	}
	expectedMessages := map[int]sqs.Message{
		1: {
			Body: `{"profile":{"LastUpdated":"` + lastUpdated.Format(time.RFC3339Nano) + `","domain":"` + lcs.DomainName + `","ProfileId":"` + uptId1 + `","AccountNumber":"","FirstName":"","MiddleName":"","LastName":"Smith","BirthDate":"","Gender":"","PhoneNumber":"","MobilePhoneNumber":"","HomePhoneNumber":"","BusinessPhoneNumber":"","PersonalEmailAddress":"cpintegration@example.com","BusinessEmailAddress":"","EmailAddress":"","Attributes":{"profile_id":"traveler_1","profile_id_obj_type":"test_booking"},"Address":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"MailingAddress":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"BillingAddress":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"ShippingAddress":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"BusinessName":"","ProfileObjects":[],"Matches":null}}`,
			MessageAttributes: map[string]string{
				"request_type": usecase.CPWriterRequestType_PutProfile,
				"domain":       lcs.DomainName,
				"txid":         lcs.Tx.TransactionID,
			},
		},
		2: {
			Body: `{"profileId":"` + uptId1 + `","profileObject":{"ID":"obj_1","Type":"test_booking","JSONContent":"","Attributes":{"accp_object_id":"obj_1","amount":"","booker_id":"","booking_id":"","email":"cpintegration@example.com","first_name":"","last_name":"Smith","last_updated":"` + lastUpdated.Format(time.RFC3339Nano) + `","loyalty_id":"","profile_id":"traveler_1","traveler_id":"traveler_1"},"AttributesInterface":{"accp_object_id":"obj_1","amount":"","booker_id":"","booking_id":"","email":"cpintegration@example.com","first_name":"","last_name":"Smith","last_updated":"` + lastUpdated.Format(time.RFC3339Nano) + `","loyalty_id":"","profile_id":"traveler_1","traveler_id":"traveler_1"},"TotalCount":0}}`,
			MessageAttributes: map[string]string{
				"request_type": usecase.CPWriterRequestType_PutProfileObject,
				"domain":       lcs.DomainName,
				"txid":         lcs.Tx.TransactionID,
			},
		},
		3: {
			Body: `{"profile":{"LastUpdated":"` + lastUpdated.Format(time.RFC3339Nano) + `","domain":"` + lcs.DomainName + `","ProfileId":"` + uptId2 + `","AccountNumber":"","FirstName":"","MiddleName":"","LastName":"Brown","BirthDate":"","Gender":"","PhoneNumber":"","MobilePhoneNumber":"","HomePhoneNumber":"","BusinessPhoneNumber":"","PersonalEmailAddress":"","BusinessEmailAddress":"","EmailAddress":"","Attributes":{"profile_id":"traveler_2","profile_id_obj_type":"test_booking"},"Address":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"MailingAddress":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"BillingAddress":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"ShippingAddress":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"BusinessName":"","ProfileObjects":[],"Matches":null}}`,
			MessageAttributes: map[string]string{
				"request_type": usecase.CPWriterRequestType_PutProfile,
				"domain":       lcs.DomainName,
				"txid":         lcs.Tx.TransactionID,
			},
		},
		4: {
			Body: `{"profileId":"` + uptId2 + `","profileObject":{"ID":"obj_2","Type":"test_booking","JSONContent":"","Attributes":{"accp_object_id":"obj_2","amount":"","booker_id":"","booking_id":"","email":"","first_name":"","last_name":"Brown","last_updated":"` + lastUpdated.Format(time.RFC3339Nano) + `","loyalty_id":"","profile_id":"traveler_2","traveler_id":"traveler_2"},"AttributesInterface":{"accp_object_id":"obj_2","amount":"","booker_id":"","booking_id":"","email":"","first_name":"","last_name":"Brown","last_updated":"` + lastUpdated.Format(time.RFC3339Nano) + `","loyalty_id":"","profile_id":"traveler_2","traveler_id":"traveler_2"},"TotalCount":0}}`,
			MessageAttributes: map[string]string{
				"request_type": usecase.CPWriterRequestType_PutProfileObject,
				"domain":       lcs.DomainName,
				"txid":         lcs.Tx.TransactionID,
			},
		},
	}

	// Validate messages
	actualMessages, err := fetchMessages(lcs.CPWriterClient, 4, 5, 6)
	if err != nil {
		t.Fatalf("[%s] error fetching messages %v", t.Name(), err)
	}
	if len(actualMessages) != len(expectedMessages) {
		t.Fatalf("[%s] expected %d messages, got %d", t.Name(), len(expectedMessages), len(actualMessages))
	}
	for _, actual := range actualMessages {
		for i, expected := range expectedMessages {
			index := strings.Index(expected.Body, "LastUpdated")
			actualIndex := strings.Index(actual.Body, "LastUpdated")
			expectedSubstring := expected.Body
			actualSubstring := actual.Body
			if index != -1 && actualIndex != -1 {
				startIndex := index + len("LastUpdated") + 2
				actualStartIndex := actualIndex + len("LastUpdated") + 2
				// Find the first comma in the sliced string
				commaIndexExpected := startIndex + strings.Index(expectedSubstring[startIndex:], ",\"")
				commaIndexActual := actualStartIndex + strings.Index(actualSubstring[actualStartIndex:], ",\"")
				if commaIndexExpected != -1 && commaIndexActual != -1 {
					expectedSubstring = expected.Body[:startIndex] + expected.Body[commaIndexExpected:]
					actualSubstring = actual.Body[:actualStartIndex] + actual.Body[commaIndexActual:]
				}
			}
			if expectedSubstring == actualSubstring && reflect.DeepEqual(actual.MessageAttributes, expected.MessageAttributes) {
				delete(expectedMessages, i)
			}
		}
	}
	if len(expectedMessages) > 0 {
		t.Fatalf("[%s] expected messages not found %v within %+v", t.Name(), expectedMessages, actualMessages)
	}

	// Send merge request
	ts := time.Now()
	_, err = lcs.MergeProfiles(
		uptId2,
		uptId1,
		ProfileMergeContext{Timestamp: ts, MergeType: MergeTypeManual, MergeIntoConnectId: uptId2, ToMergeConnectId: uptId1},
	)
	if err != nil {
		t.Fatalf("[%s] error merging profiles %v", t.Name(), err)
	}

	// Validate objects were properly sent to CP Writer queue
	expectedMessages = map[int]sqs.Message{
		1: {
			Body: `{"profile":{"LastUpdated":"0001-01-01T00:00:00Z","domain":"` + lcs.DomainName + `","ProfileId":"` + uptId2 + `","AccountNumber":"","FirstName":"","MiddleName":"","LastName":"Brown","BirthDate":"","Gender":"","PhoneNumber":"","MobilePhoneNumber":"","HomePhoneNumber":"","BusinessPhoneNumber":"","PersonalEmailAddress":"cpintegration@example.com","BusinessEmailAddress":"","EmailAddress":"","Attributes":{"profile_id":"traveler_2"},"Address":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"MailingAddress":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"BillingAddress":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"ShippingAddress":{"Address1":"","Address2":"","Address3":"","Address4":"","City":"","State":"","Province":"","PostalCode":"","Country":""},"BusinessName":"","ProfileObjects":[],"Matches":null}}`,
			MessageAttributes: map[string]string{
				"request_type": usecase.CPWriterRequestType_PutProfile,
				"domain":       lcs.DomainName,
				"txid":         lcs.Tx.TransactionID,
			},
		},
		2: {
			Body: `{\"profileId\":\"` + uptId2 + `",\"profileObject\":{\"ID\":\"obj_1\",\"Type\":\"test_booking\",\"JSONContent\":\"\",\"Attributes\":{\"_object_type_name\":\"test_booking\",\"accp_object_id\":\"obj_1\",\"amount\":\"\",\"booker_id\":\"\",\"booking_id\":\"\",\"connect_id\":\"` + uptId2 + `",\"email\":\"cpintegration@example.com\",\"first_name\":\"\",\"last_name\":\"Smith\",\"loyalty_id\":\"\",\"overall_confidence_score\":\"0\",\"profile_id\":\"traveler_1\",\"timestamp\":\"2024-07-08 17:53:06.677609 +0000 UTC\",\"traveler_id\":\"traveler_1\"},\"AttributesInterface\":{\"_object_type_name\":\"test_booking\",\"accp_object_id\":\"obj_1\",\"amount\":\"\",\"booker_id\":\"\",\"booking_id\":\"\",\"connect_id\":\"` + uptId2 + `",\"email\":\"cpintegration@example.com\",\"first_name\":\"\",\"last_name\":\"Smith\",\"loyalty_id\":\"\",\"overall_confidence_score\":0,\"profile_id\":\"traveler_1\",\"timestamp\":\"2024-07-08T17:53:06.677609Z\",\"traveler_id\":\"traveler_1\"}}}`,
			MessageAttributes: map[string]string{
				"request_type": usecase.CPWriterRequestType_PutProfileObject,
				"domain":       lcs.DomainName,
				"txid":         lcs.Tx.TransactionID,
			},
		},
		3: {
			Body: `{"profileId":"` + uptId2 + `","profileObject":{"ID":"obj_2","Type":"test_booking","JSONContent":"","Attributes":{"accp_object_id":"obj_2","amount":"","booker_id":"","booking_id":"","email":"","first_name":"","last_name":"Brown","loyalty_id":"","profile_id":"traveler_2","traveler_id":"traveler_2"},"AttributesInterface":{"accp_object_id":"obj_2","amount":"","booker_id":"","booking_id":"","email":"","first_name":"","last_name":"Brown","loyalty_id":"","profile_id":"traveler_2","traveler_id":"traveler_2"}}}`,
			MessageAttributes: map[string]string{
				"request_type": usecase.CPWriterRequestType_PutProfileObject,
				"domain":       lcs.DomainName,
				"txid":         lcs.Tx.TransactionID,
			},
		},
		4: {
			Body: `{"profileId":"` + uptId1 + `"}`,
			MessageAttributes: map[string]string{
				"request_type": usecase.CPWriterRequestType_DeleteProfile,
				"domain":       lcs.DomainName,
				"txid":         lcs.Tx.TransactionID,
			},
		},
	}

	actualMessages, err = fetchMessages(lcs.CPWriterClient, 4, 5, 6)
	if err != nil {
		t.Fatalf("[%s] error fetching messages %v", t.Name(), err)
	}

	for _, actual := range actualMessages {
		for i, expected := range expectedMessages {
			if reflect.DeepEqual(actual.MessageAttributes, expected.MessageAttributes) {
				delete(expectedMessages, i)
				break
			}
		}
	}
	if len(expectedMessages) > 0 {
		t.Fatalf("[%s] expected messages not found %v", t.Name(), expectedMessages)
	}
}

/*
func TestCPIntegrationControl(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[customerprofiles] Error setting up aurora: %+v ", err)
	}
	dynamoTableName := "customer-profiles-integration-test-" + core.GenerateUniqueId()
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error init with new table: %v", err)
	}
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error waiting for table creation: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[TestUpdateProfile] DeleteTable failed: %s", err)
		}
	})

	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)

	kmsc := kms.Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	keyArn, err := kmsc.CreateKey("tah-unit-test-key")
	if err != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err)
	}
	t.Cleanup(func() {
		err = kmsc.DeleteKey(keyArn)
		if err != nil {
			t.Errorf("[TestUpdateProfile] DeleteTable failed: %s", err)
		}
	})

	// Init low cost config
	mergeQueue := buildTestMergeQueue(t, envCfg.Region)
	cpWriterQueue := buildTestCPWriterQueue(t, envCfg.Region)
	initOptions := CustomerProfileInitOptions{MergeQueueClient: mergeQueue}
	lc := InitLowCost(
		envCfg.Region,
		auroraCfg,
		&dynamoCfg,
		kinesisMockCfg,
		cpWriterQueue,
		"tah/upt/source/tah-core/customerprofile",
		core.TEST_SOLUTION_ID,
		core.TEST_SOLUTION_VERSION,
		&initOptions,
	)

	//Tests that cp domain is created but dynamo is not
	t.Logf("Creating domain with CP and no dynamo")
	CPMode := true
	DynamoMode := false

	domainOptionsCP := DomainOptions{
		CustomerProfileMode: &CPMode,
		DynamoMode:          &DynamoMode,
	}

	domainWithCp := randomDomain("with_cp_domain")
	err = lc.CreateDomainWithQueue(domainWithCp, keyArn, map[string]string{DOMAIN_CONFIG_ENV_KEY: "test"}, "", "", domainOptionsCP)
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error creating domain: %v", err)
	}

	testCfg := customerprofiles.InitWithDomain(domainWithCp, envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	dom, err := testCfg.GetDomain()
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error getting domain: %v", err)
	}
	if dom.Name != domainWithCp {
		t.Errorf("[TestControlPlaneIntegration] wrong domain: %v", err)
	}

	err = lc.DeleteDomainByName(domainWithCp)
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error deleting domain: %v", err)
	}
	_, err = testCfg.GetDomain()
	if err == nil {
		t.Errorf("[TestControlPlaneIntegration] domain should not exist")
	}

	//Tests that cp domain is created and dynamo is
	t.Logf("Creating domain with both")
	CPMode = true
	DynamoMode = true

	DomainOptionsBoth := DomainOptions{
		CustomerProfileMode: &CPMode,
		DynamoMode:          &DynamoMode,
	}

	domainWithBoth := randomDomain("with_both")
	err = lc.CreateDomainWithQueue(domainWithBoth, keyArn, map[string]string{DOMAIN_CONFIG_ENV_KEY: "test"}, "", "", DomainOptionsBoth)
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error creating domain: %v", err)
	}

	testCfg.DomainName = domainWithBoth
	dom, err = testCfg.GetDomain()
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error getting domain: %v", err)
	}
	if dom.Name != domainWithBoth {
		t.Errorf("[TestControlPlaneIntegration] wrong domain: %v", err)
	}

	err = lc.DeleteDomainByName(domainWithBoth)
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error deleting domain: %v", err)
	}
	_, err = testCfg.GetDomain()
	if err == nil {
		t.Errorf("[TestControlPlaneIntegration] domain should not exist")
	}

	t.Logf("Creating domain with dynamo and no CP")
	CPMode = false
	DynamoMode = true

	DomainOptionsDynamo := DomainOptions{
		CustomerProfileMode: &CPMode,
		DynamoMode:          &DynamoMode,
	}

	domainWithDynamo := randomDomain("with_dynamo")
	err = lc.CreateDomainWithQueue(domainWithDynamo, keyArn, map[string]string{DOMAIN_CONFIG_ENV_KEY: "test"}, "", "", DomainOptionsDynamo)
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error creating domain: %v", err)
	}

	testCfg.DomainName = domainWithDynamo
	_, err = testCfg.GetDomain()
	if err == nil {
		t.Errorf("[TestControlPlaneIntegration] domain should not exist")
	}

	err = lc.DeleteDomainByName(domainWithDynamo)
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error deleting domain: %v", err)
	}
}

func TestCPIntegrationControlMapping(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Errorf("[customerprofiles] Error setting up aurora: %+v ", err)
	}
	dynamoTableName := "customer-profiles-integration-mapping-test-" + core.GenerateUniqueId()
	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error init with new table: %v", err)
	}
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error waiting for table creation: %v", err)
	}
	err = dynamoCfg.WaitForTableCreation()
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error waiting for table creation: %v", err)
	}
	t.Cleanup(func() {
		err = dynamoCfg.DeleteTable(dynamoTableName)
		if err != nil {
			t.Errorf("[TestControlPlaneIntegration] DeleteTable failed: %s", err)
		}
	})

	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)

	kmsc := kms.Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	keyArn, err := kmsc.CreateKey("tah-unit-test-key-mapping")
	if err != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err)
	}
	t.Cleanup(func() {
		err = kmsc.DeleteKey(keyArn)
		if err != nil {
			t.Errorf("[TestUpdateProfile] DeleteKey failed: %s", err)
		}
	})

	// Init low cost config
	mergeQueue := buildTestMergeQueue(t, envCfg.Region)
	cpWriterQueue := buildTestCPWriterQueue(t, envCfg.Region)
	initOptions := CustomerProfileInitOptions{MergeQueueClient: mergeQueue}
	lc := InitLowCost(
		envCfg.Region,
		auroraCfg,
		&dynamoCfg,
		kinesisMockCfg,
		cpWriterQueue,
		"tah/upt/source/tah-core/customerprofile",
		core.TEST_SOLUTION_ID,
		core.TEST_SOLUTION_VERSION,
		&initOptions,
	)

	CPMode := true
	DynamoMode := false

	domainOptionsCP := DomainOptions{
		CustomerProfileMode: &CPMode,
		DynamoMode:          &DynamoMode,
	}

	domainWithCp := randomDomain("mapping_test")
	err = lc.CreateDomainWithQueue(domainWithCp, keyArn, map[string]string{DOMAIN_CONFIG_ENV_KEY: "test"}, "", "", domainOptionsCP)
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error creating domain: %v", err)
	}

	err = lc.CreateMapping("air_booking", "air booking mapping", []FieldMapping{
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.traveler_id",
			Target:  "_profile.Attributes.traveler_id",
			Indexes: []string{"PROFILE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.test_field",
			Target:  "test_field",
			KeyOnly: true,
		},
	})
	if err != nil {
		t.Errorf("[TestMergeManyLC] error creating mapping: %v", err)
	}

	cfg := customerprofiles.InitWithDomain(domainWithCp, envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	mapp, err := cfg.GetMapping("air_booking")
	if err != nil {
		t.Errorf("[TestMergeManyLC] error getting mapping for this object type: %v", err)
	}
	if len(mapp.Fields) != 3 {
		t.Errorf("[TestMergeManyLC] wrong number of fields: %v", err)
	}

	err = lc.DeleteDomainByName(domainWithCp)
	if err != nil {
		t.Errorf("[TestControlPlaneIntegration] error deleting domain: %v", err)
	}
}

func TestCPIntegrationProfileInsertion(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	domainWithCp := randomDomain("insertion_test")
	lc, err := setupCPIntegrationTest(t, "customer-profiles-insertion-test", domainWithCp)
	if err != nil {
		t.Errorf("[TestCpIntegrationProfileInsertion] error setting up integration test: %v", err)
	}

	clickstreamMappings := []FieldMapping{
		// Profile Object Unique Key
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.session_id",
			Target:  "session_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.num_pax_children",
			Target:  "num_pax_children",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.traveler_id",
			Target:  "_profile.Attributes.profile_id",
			Indexes: []string{"PROFILE"},
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.last_name",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
	}

	clickstreamType := "clickstream"
	err = lc.CreateMapping(clickstreamType, "clickstream mapping", clickstreamMappings)
	if err != nil {
		t.Errorf("[TestProfileInsertionCP] error creating mapping: %v", err)
	}

	profileTwoTravelerID := "profileTwoTravellerId"
	obj, _ := json.Marshal(map[string]string{
		"traveler_id":      profileTwoTravelerID,
		"accp_object_id":   "test_accp_object_id",
		"session_id":       "test_session_id",
		"num_pax_children": "test_num_pax_children",
		"email":            "clickstreamEmail@example.com",
		"last_name":        "Smith",
	})
	err = lc.PutProfileObject(string(obj), clickstreamType)
	if err != nil {
		t.Errorf("[TestProfileInsertionCP] error creating profile object 1: %v", err)
	}
	anotherObj, _ := json.Marshal(map[string]string{
		"traveler_id":      profileTwoTravelerID,
		"accp_object_id":   "test_accp_object_id_new",
		"session_id":       "test_session_id",
		"num_pax_children": "test_num_pax_children",
		"last_name":        "Brown",
	})
	err = lc.PutProfileObject(string(anotherObj), clickstreamType)
	if err != nil {
		t.Errorf("[TestProfileInsertionCP] error creating profile object 1: %v", err)
	}
	cfg := customerprofiles.InitWithDomain(domainWithCp, envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	ourConnectId, err := WaitForProfileCreationByProfileId(profileTwoTravelerID, "profile_id", lc)
	if err != nil {
		t.Errorf("[TestProfileInsertionCP] error waiting for profile creation: %v", err)
	}
	log.Printf("[TestProfileInsertionCP] connect id: %+v", ourConnectId)
	cpConnectId, err := WaitForProfileCreationInCPByProfileId(ourConnectId, cfg)
	if err != nil {
		t.Errorf("[TestProfileInsertionCP] error waiting for profile creation in CP: %v", err)
	}
	connectId, err := cfg.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, ourConnectId)
	if err != nil {
		t.Errorf("[TestProfileInsertionCP] failed to get profile id")
	}
	log.Printf("[TestProfileInsertionCP] connect id: %+v", connectId)
	if connectId != cpConnectId {
		t.Errorf("[TestProfileInsertionCP] wrong connect id: %v", connectId)
	}
	err = WaitForExpectedCondition(func() bool {
		profile, err := cfg.GetProfile(
			connectId,
			[]string{customerprofiles.PROFILE_FIELD_OBJECT_TYPE, clickstreamType},
			[]customerprofiles.PaginationOptions{},
		)
		if err != nil {
			log.Printf("[TestProfileInsertionCP] failed to get profile: %v", err)
			return false
		}
		if len(profile.ProfileObjects) != 3 {
			log.Printf("[TestProfileInsertionCP] wrong number of profile objects: %v", len(profile.ProfileObjects))
			return false
		}
		//should just be Smith, but CPs problem and they are fixing it
		if (profile.LastName != "Brown") && (profile.LastName != "Smith") {
			log.Printf("[TestProfileInsertionCP] wrong last name: %v", profile.LastName)
			return false
		}
		if profile.PersonalEmailAddress != "clickstreamEmail@example.com" {
			log.Printf("[TestProfileInsertionCP] wrong email: %v", profile.PersonalEmailAddress)
			return false
		}
		domain, err := cfg.GetDomain()
		if err != nil {
			return false
		}
		if domain.NProfiles != 1 {
			log.Printf("[TestProfileInsertionCP] number of profiles is wrong")
			return false
		}
		return true
	}, 24, 5*time.Second)
	if err != nil {
		t.Errorf("[TestProfileInsertionCP] error waiting for profile creation: %v", err)
	}
}

func TestCPAirBookingLoad(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	domainWithCp := randomDomain("air_load_test")
	lc, err := setupCPIntegrationTest(t, "customer-profiles-air-booking-load-test", domainWithCp)
	if err != nil {
		t.Errorf("[TestCpIntegrationProfileInsertion] error setting up integration test: %v", err)
	}

	// Create mapping
	airBookingType := "air_booking"
	testMappings := []FieldMapping{
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.traveler_id",
			Target:  "_profile.Attributes.profile_id",
			Indexes: []string{"PROFILE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.segment_id",
			Target:  "segment_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.from",
			Target:  "from",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.arrival_date",
			Target:  "arrival_date",
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
	}
	err = lc.CreateMapping(airBookingType, "Mapping for "+airBookingType, testMappings)
	if err != nil {
		t.Errorf("[TestHistory] error creating mapping: %v", err)
	}

	for i := 0; i < 100; i++ {
		profileOneTravelerID := "test_traveler_id" + strconv.Itoa(i)
		obj1, _ := json.Marshal(map[string]string{
			"traveler_id":    profileOneTravelerID,
			"accp_object_id": "test_accp_object_id" + strconv.Itoa(i),
			"segment_id":     "test_segment_id",
			"from":           "test_from" + strconv.Itoa(i),
			"arrival_date":   "test_arrival_date",
			"email":          "airBookingEmail" + strconv.Itoa(i) + "@example.com",
		})
		err = lc.PutProfileObject(string(obj1), airBookingType)
		if err != nil {
			t.Errorf("[TestHistory] error creating profile object 1: %v", err)
		}
	}
	cfg := customerprofiles.InitWithDomain(domainWithCp, envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)

	err = WaitForExpectedCondition(func() bool {
		domain, err := cfg.GetDomain()
		if err != nil {
			return false
		}
		if domain.NProfiles != 100 {
			t.Logf("[TestProfileInsertionCP] wrong number of profiles: %v", domain.NProfiles)
			return false
		}
		return true
	}, 24, 5*time.Second)
	if err != nil {
		t.Errorf("[TestProfileInsertionCP] error waiting for profile creation: %v", err)
	}
}

func TestCPPushAndGetWithAttributesCpIntegration(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	domainWithCp := "domain_profile_push_attributes" + fmt.Sprint(rand.Int())
	dynamoTableName := randomTable("customer-profiles-push-and-get-with-attributes")
	lc, err := setupCPIntegrationTest(t, dynamoTableName, domainWithCp)
	if err != nil {
		t.Errorf("[TestCpIntegrationProfileAttributes] error setting up integration test: %v", err)
	}

	airBookingType := "air_booking"
	testMappings := []FieldMapping{
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.traveler_id",
			Target:  "_profile.Attributes.profile_id",
			Indexes: []string{"PROFILE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.segment_id",
			Target:  "segment_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.from",
			Target:  "from",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.arrival_date",
			Target:  "arrival_date",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.credit_card",
			Target:  "_profile.Attributes.credit_card",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.street",
			Target:  "_profile.Address.City",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_destination",
			Target:  "_profile.Attributes.last_destination",
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.firstName",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.lastName",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
	}
	err = lc.CreateMapping(airBookingType, "Mapping for "+airBookingType, testMappings)
	if err != nil {
		t.Errorf("[TestCpIntegrationProfileAttributes] error creating mapping: %v", err)
	}

	//not ideal, but no way to synchronize these events
	time.Sleep(5 * time.Second)

	profileOneTravelerID := "test_traveler_id" + strconv.Itoa(1)
	obj1, _ := json.Marshal(map[string]string{
		"traveler_id":      profileOneTravelerID,
		"accp_object_id":   "test_accp_object_id" + strconv.Itoa(1),
		"segment_id":       "test_segment_id",
		"from":             "test_from" + strconv.Itoa(1),
		"arrival_date":     "test_arrival_date",
		"email":            "airBookingEmail" + strconv.Itoa(1) + "@example.com",
		"credit_card":      "test_credit_card",
		"last_destination": "test_last_destination",
		"firstName":        "test_firstName",
		"lastName":         "test_lastName",
		"street":           "test_street",
	})
	err = lc.PutProfileObject(string(obj1), airBookingType)
	if err != nil {
		t.Errorf("[TestCpIntegrationProfileAttributes] error creating profile object 1: %v", err)
	}
	profId, err := lc.GetProfileId("profile_id", profileOneTravelerID)
	if err != nil {
		t.Errorf("[TestCpIntegrationProfileAttributes] error getting profile id: %v", err)
	}
	cfg := customerprofiles.InitWithDomain(domainWithCp, envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	err = WaitForExpectedCondition(func() bool {
		connectId, err := cfg.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, profId)
		if err != nil {
			t.Logf("[TestCpIntegrationProfileAttributes] failed to get profile id")
			return false
		}
		prof, err := cfg.GetProfile(connectId, []string{airBookingType}, []customerprofiles.PaginationOptions{})
		if err != nil {
			t.Logf("[TestCpIntegrationProfileAttributes] error getting profile: %v", err)
			return false
		}
		//fails sometimes, because the mapping has not updated before
		//the object gets sent. Another race condition here. Also addressing with team
		if prof.PersonalEmailAddress == "" {
			t.Logf(
				"[TestCpIntegrationProfileAttributes] Known race condition, updating mappings and putprofileobject creates profile blank fields",
			)
			return true
		}
		if prof.Attributes["credit_card"] != "test_credit_card" {
			t.Logf("[TestCpIntegrationProfileAttributes] wrong credit card value: %v", prof.Attributes["credit_card"])
			return false
		}
		if prof.Attributes["last_destination"] != "test_last_destination" {
			t.Logf("[TestCpIntegrationProfileAttributes] wrong last destination value: %v", prof.Attributes["last_destination"])
			return false
		}
		if prof.PersonalEmailAddress != "airBookingEmail1@example.com" {
			t.Logf("[TestCpIntegrationProfileAttributes] wrong email value: %v", prof.PersonalEmailAddress)
			return false
		}
		if prof.FirstName != "test_firstName" {
			t.Logf("[TestCpIntegrationProfileAttributes] wrong first name value: %v", prof.FirstName)
			return false
		}
		if prof.LastName != "test_lastName" {
			t.Logf("[TestCpIntegrationProfileAttributes] wrong last name value: %v", prof.LastName)
			return false
		}
		if prof.Address.City != "test_street" {
			t.Logf("[TestCpIntegrationProfileAttributes] wrong street value: %v", prof.Address.City)
			return false
		}
		return true
	}, 12, 5*time.Second)
	if err != nil {
		t.Errorf("[TestCpIntegrationProfileAttributes] error waiting for profile creation: %v", err)
	}
}

func TestCPMergeCPIntegration(t *testing.T) {
	t.Parallel()
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	domainWithCp := randomDomain("merge_test")
	dynamoTableName := "customer-profiles-merge-" + core.GenerateUniqueId()
	lc, err := setupCPIntegrationTest(t, dynamoTableName, domainWithCp)
	if err != nil {
		t.Errorf("[TestMergeCPIntegration] error setting up integration test: %v", err)
	}

	airBookingType := "air_booking"
	testMappings := []FieldMapping{
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.traveler_id",
			Target:  "_profile.Attributes.profile_id",
			Indexes: []string{"PROFILE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.segment_id",
			Target:  "segment_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.from",
			Target:  "from",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.arrival_date",
			Target:  "arrival_date",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.credit_card",
			Target:  "_profile.Attributes.credit_card",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_destination",
			Target:  "_profile.Attributes.last_destination",
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.firstName",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.lastName",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
	}
	clickstreamType := "clickstream"
	clickstreamMappings := []FieldMapping{
		// Profile Object Unique Key
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.session_id",
			Target:  "session_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.num_pax_children",
			Target:  "num_pax_children",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.traveler_id",
			Target:  "_profile.Attributes.profile_id",
			Indexes: []string{"PROFILE"},
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.last_name",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
	}
	err = lc.CreateMapping(airBookingType, "Mapping for "+airBookingType, testMappings)
	if err != nil {
		t.Errorf("[TestHistory] error creating mapping: %v", err)
	}
	err = lc.CreateMapping(clickstreamType, "Mapping for "+airBookingType, clickstreamMappings)
	if err != nil {
		t.Errorf("[TestHistory] error creating mapping: %v", err)
	}

	profileOneTravelerId := "test_traveler_id" + strconv.Itoa(1)
	obj1, _ := json.Marshal(map[string]string{
		"traveler_id":      profileOneTravelerId,
		"accp_object_id":   "test_accp_object_id" + strconv.Itoa(1),
		"segment_id":       "test_segment_id",
		"from":             "test_from" + strconv.Itoa(1),
		"arrival_date":     "test_arrival_date",
		"email":            "airBookingEmail" + strconv.Itoa(1) + "@example.com",
		"credit_card":      "test_credit_card",
		"last_destination": "test_last_destination",
		"firstName":        "test_firstName",
		"lastName":         "test_lastName",
	})
	err = lc.PutProfileObject(string(obj1), airBookingType)
	if err != nil {
		t.Errorf("[TestHistory] error creating profile object 1: %v", err)
	}
	profileTwoTravelerId := "test_traveler_id" + strconv.Itoa(2)
	obj, _ := json.Marshal(map[string]string{
		"traveler_id":      profileTwoTravelerId,
		"accp_object_id":   "test_accp_object_id",
		"session_id":       "test_session_id",
		"num_pax_children": "test_num_pax_children",
		"email":            "clickstreamEmail@example.com",
		"last_name":        "Smith",
	})
	err = lc.PutProfileObject(string(obj), clickstreamType)
	if err != nil {
		t.Errorf("[TestMergeCPIntegration] error creating profile object 1: %v", err)
	}

	connectId1, err := WaitForProfileCreationByProfileId(profileOneTravelerId, "profile_id", lc)
	if err != nil {
		t.Errorf("[TestMergeCPIntegration] error waiting for profile one creation: %v", err)
	}
	cfg := customerprofiles.InitWithDomain(domainWithCp, envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	_, err = WaitForProfileCreationInCPByProfileId(connectId1, cfg)
	if err != nil {
		t.Errorf("[TestMergeCPIntegration] error waiting for profile one creation in CP: %v", err)
	}
	connectId2, err := WaitForProfileCreationByProfileId(profileTwoTravelerId, "profile_id", lc)
	if err != nil {
		t.Errorf("[TestMergeCPIntegration] error waiting for profile two creation: %v", err)
	}
	_, err = WaitForProfileCreationInCPByProfileId(connectId2, cfg)
	if err != nil {
		t.Errorf("[TestMergeCPIntegration] error waiting for profile two creation in CP: %v", err)
	}

	status, err := lc.MergeProfiles(connectId1, connectId2, ProfileMergeContext{})
	if err != nil {
		t.Errorf("[TestMergeCPIntegration] error merging profiles: %v", err)
	}
	if status != "SUCCESS" {
		t.Errorf("[TestMergeCPIntegration] error merging profiles, incorrect status: %v", status)
	}
	err = WaitForExpectedCondition(func() bool {
		boo := true
		_, err := cfg.GetDomain()
		if err != nil {
			boo = false
		}
		connectId, err := cfg.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, connectId1)
		if err != nil {
			t.Logf("[TestMergeCPIntegration] error, failed to get profile id")
			boo = false
		}
		prof, err := cfg.GetProfile(
			connectId,
			[]string{airBookingType, clickstreamType, customerprofiles.PROFILE_FIELD_OBJECT_TYPE},
			[]customerprofiles.PaginationOptions{},
		)
		if err != nil {
			t.Logf("[TestMergeCPIntegration] error getting profile: %v", err)
			boo = false
		}
		if prof.FirstName != "test_firstName" {
			t.Logf("[TestMergeCPIntegration] error, wrong first name value: %v", prof.FirstName)
			boo = false
		}
		if prof.LastName != "Smith" {
			t.Logf("[TestMergeCPIntegration] error, wrong last name value: %v", prof.LastName)
			boo = false
		}
		if len(prof.ProfileObjects) != 3 {
			t.Logf("[TestMergeCPIntegration] error, wrong number of profile objects: %v", len(prof.ProfileObjects))
			boo = false
		}
		//cp team working on a fix for this, will uncomment when fixed
		// connectId, err = cfg.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, connectId2)
		// if connectId != "" {
		// 	t.Logf("[TestMergeCPIntegration] error, profile should not exist: %v", err)
		// 	boo = false
		// }
		// prof2, err := cfg.GetProfile(connectId, []string{airBookingType, clickstreamType, customerprofiles.PROFILE_FIELD_OBJECT_TYPE}, []customerprofiles.PaginationOptions{})
		// if err == nil {
		// 	t.Logf("[TestMergeCPIntegration] error getting profile: %v", err)
		// 	return false
		// }
		//log.Printf("[TestMergeCPIntegration] profile: %v", prof2)
		// if domain.NProfiles != 1 {
		// 	t.Logf("[TestMergeCPIntegration] error, wrong number of profiles: %v", domain.NProfiles)
		// 	boo = false
		// }
		return boo
	}, 24, 5*time.Second)
	if err != nil {
		t.Errorf("[TestMergeCPIntegration] error waiting for profile merge: %v", err)
	}

	obj2, _ := json.Marshal(map[string]string{
		"traveler_id":      profileOneTravelerId,
		"accp_object_id":   "test_accp_object_id2",
		"session_id":       "test_session_id",
		"num_pax_children": "test_num_pax_children",
		"email":            "clickstreamEmail@example.com",
		"last_name":        "new_last_name",
	})
	err = lc.PutProfileObject(string(obj2), clickstreamType)
	if err != nil {
		t.Errorf("[TestMergeCPIntegration] error creating profile object 1: %v", err)
	}
	err = WaitForExpectedCondition(func() bool {
		connectId, err := cfg.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, connectId1)
		if err != nil {
			t.Logf("[TestMergeCPIntegration] error, failed to get profile id")
			return false
		}
		_, err = cfg.GetProfile(connectId, []string{airBookingType, clickstreamType}, []customerprofiles.PaginationOptions{})
		if err != nil {
			t.Logf("[TestMergeCPIntegration] error getting profile: %v", err)
			return false
		}
		// if prof.LastName != "new_last_name" {
		// 	t.Logf("[TestMergeCPIntegration] error, wrong last name value: %v", prof.LastName)
		// 	return false
		// }
		// if len(prof.ProfileObjects) != 3 {
		// 	t.Logf("[TestMergeCPIntegration] error, wrong number of profile objects: %v", len(prof.ProfileObjects))
		// 	return false
		// }
		return true
	}, 24, 5*time.Second)
	if err != nil {
		t.Errorf("[TestMergeCPIntegration] error waiting for profile merge: %v", err)
	}
}

func TestSkipConditionCache(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	domainWithCp := randomDomain("profile_skip_test")
	dynamoTableName := "customer-profiles-skip-" + core.GenerateUniqueId()
	lc, err := setupCPIntegrationTest(t, dynamoTableName, domainWithCp)
	if err != nil {
		t.Fatalf("[TestSkipCPIntegration] error setting up integration test: %v", err)
	}
	airBookingType := "air_booking"
	testMappings := []FieldMapping{
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.traveler_id",
			Target:  "_profile.Attributes.profile_id",
			Indexes: []string{"PROFILE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.segment_id",
			Target:  "segment_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.credit_card",
			Target:  "_profile.Attributes.credit_card",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_destination",
			Target:  "_profile.Attributes.last_destination",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.address",
			Target:  "_profile.Address.Address1",
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.firstName",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.lastName",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.businessEmail",
			Target:      "_profile.BusinessEmailAddress",
			Searcheable: true,
		},
	}
	clickstreamType := "clickstream"
	clickstreamMappings := []FieldMapping{
		// Profile Object Unique Key
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.session_id",
			Target:  "session_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.traveler_id",
			Target:  "_profile.Attributes.profile_id",
			Indexes: []string{"PROFILE"},
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.businessEmail",
			Target:      "_profile.BusinessEmailAddress",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.cc",
			Target:      "_profile.Attributes.cc",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.last_name",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
	}
	err = lc.CreateMapping(airBookingType, "Mapping for "+airBookingType, testMappings)
	if err != nil {
		t.Fatalf("[TestHistory] error creating mapping: %v", err)
	}
	err = lc.CreateMapping(clickstreamType, "Mapping for "+airBookingType, clickstreamMappings)
	if err != nil {
		t.Fatalf("[TestHistory] error creating mapping: %v", err)
	}

	ruleSet1 := RuleSet{
		Name: "ruleset1",
		Rules: []Rule{
			{
				Index:       0,
				Name:        "rule1",
				Description: "Two skip conditions",
				Conditions: []Condition{
					{
						Index:                0,
						ConditionType:        CONDITION_TYPE_SKIP,
						IncomingObjectType:   PROFILE_OBJECT_TYPE_NAME,
						IncomingObjectField:  "FirstName",
						Op:                   RULE_OP_EQUALS,
						IncomingObjectType2:  PROFILE_OBJECT_TYPE_NAME,
						IncomingObjectField2: "Attributes.last_destination",
						IndexNormalization:   NORMALIZATION_SETTINGS_DEFAULT,
					},
					{
						Index:               1,
						ConditionType:       CONDITION_TYPE_SKIP,
						IncomingObjectType:  PROFILE_OBJECT_TYPE_NAME,
						IncomingObjectField: "BusinessEmailAddress",
						Op:                  RULE_OP_EQUALS_VALUE,
						SkipConditionValue:  StringVal("bad data"),
						IndexNormalization:  NORMALIZATION_SETTINGS_DEFAULT,
					},
				},
			},
			{
				Index:       1,
				Name:        "rule1",
				Description: "Two skip conditions",
				Conditions: []Condition{
					{
						Index:                0,
						ConditionType:        CONDITION_TYPE_SKIP,
						IncomingObjectType:   PROFILE_OBJECT_TYPE_NAME,
						IncomingObjectField:  "Attributes.cc",
						Op:                   RULE_OP_EQUALS,
						IncomingObjectType2:  PROFILE_OBJECT_TYPE_NAME,
						IncomingObjectField2: "PersonalEmailAddress",
						IndexNormalization:   NORMALIZATION_SETTINGS_DEFAULT,
					},
					{
						Index:               1,
						ConditionType:       CONDITION_TYPE_SKIP,
						IncomingObjectType:  PROFILE_OBJECT_TYPE_NAME,
						IncomingObjectField: "Attributes.cc",
						Op:                  RULE_OP_NOT_EQUALS_VALUE,
						SkipConditionValue:  StringVal(""),
						IndexNormalization:  NORMALIZATION_SETTINGS_DEFAULT,
					},
				},
			},
		},
	}

	err = lc.SaveCacheRuleSet(ruleSet1.Rules)
	if err != nil {
		t.Fatalf("[TestHistory] error saving ruleset: %v", err)
	}
	ruleSet, err := lc.ListCacheRuleSets(true)
	if err != nil {
		t.Fatalf("[TestHistory] error listing ruleset: %v", err)
	}
	log.Printf("[TestHistory] ruleset: %v", ruleSet)
	err = lc.ActivateCacheRuleSet()
	if err != nil {
		t.Fatalf("[TestHistory] error activating ruleset: %v", err)
	}

	//object 1: only one condition met, need both, goes into cache
	profileOneTravelerID := "test_traveler_id" + strconv.Itoa(1)
	obj1, _ := json.Marshal(map[string]string{
		"traveler_id":      profileOneTravelerID,
		"accp_object_id":   "test_accp_object_id" + strconv.Itoa(1),
		"segment_id":       "test_segment_id",
		"last_destination": "last_destination" + strconv.Itoa(1),
		"firstName":        "last_destination" + strconv.Itoa(1),
	})
	err = lc.PutProfileObject(string(obj1), airBookingType)
	if err != nil {
		t.Fatalf("[TestSkipCPIntegration] error creating profile object 1: %v", err)
	}

	//object 2: both conditions met, does not go into cache
	profileTwoTravelerID := "test_traveler_id" + strconv.Itoa(2)
	obj2, _ := json.Marshal(map[string]string{
		"traveler_id":      profileTwoTravelerID,
		"accp_object_id":   "test_accp_object_id" + strconv.Itoa(2),
		"segment_id":       "test_segment_id",
		"last_destination": "last_destination" + strconv.Itoa(2),
		"firstName":        "last_destination" + strconv.Itoa(2),
		"businessEmail":    "bad data",
	})
	err = lc.PutProfileObject(string(obj2), airBookingType)
	if err != nil {
		t.Fatalf("[TestSkipCPIntegration] error creating profile object 2: %v", err)
	}

	//object 3: Different rule is an or
	//also, we compare across objects for this rule
	//clickstream and air booking set different profile values, and we need both
	//this is potentially unexpected, as the profile from first object will go through,
	//the second won't there are use cases for this, but something to keep in mind
	//both objects map to same profile
	profileThreeTravelerID := "test_traveler_id" + strconv.Itoa(3)
	obj3click, _ := json.Marshal(map[string]string{
		"traveler_id":    profileThreeTravelerID,
		"accp_object_id": "test_accp_object_id_air" + strconv.Itoa(3),
		"last_name":      "Bob",
		"cc":             "something",
	})
	err = lc.PutProfileObject(string(obj3click), clickstreamType)
	if err != nil {
		t.Fatalf("[TestSkipCPIntegration] error creating profile object 3: %v", err)
	}
	obj3air, _ := json.Marshal(map[string]string{
		"traveler_id":    profileThreeTravelerID,
		"accp_object_id": "test_accp_object_id_click" + strconv.Itoa(3),
		"segment_id":     "test_segment_id",
		"email":          "something",
		"firstName":      "Michael",
	})
	err = lc.PutProfileObject(string(obj3air), airBookingType)
	if err != nil {
		t.Fatalf("[TestSkipCPIntegration] error creating profile object 3: %v", err)
	}

	//object 4: no conditions met, goes into cache
	profileFourTravelerID := "test_traveler_id" + strconv.Itoa(4)
	obj4, _ := json.Marshal(map[string]string{
		"traveler_id":      profileFourTravelerID,
		"accp_object_id":   "test_accp_object_id" + strconv.Itoa(4),
		"segment_id":       "test_segment_id",
		"last_destination": "destination" + strconv.Itoa(4),
	})
	err = lc.PutProfileObject(string(obj4), airBookingType)
	if err != nil {
		t.Fatalf("[TestSkipCPIntegration] error creating profile object 4: %v", err)
	}

	connectId1, err := WaitForProfileCreationByProfileId(profileOneTravelerID, "profile_id", lc)
	if err != nil {
		t.Fatalf("[TestSkipCPIntegration] error waiting for profile one creation: %v", err)
	}
	connectId2, err := WaitForProfileCreationByProfileId(profileTwoTravelerID, "profile_id", lc)
	if err != nil {
		t.Fatalf("[TestSkipCPIntegration] error waiting for profile two creation: %v", err)
	}
	connectId4, err := WaitForProfileCreationByProfileId(profileFourTravelerID, "profile_id", lc)
	if err != nil {
		t.Fatalf("[TestSkipCPIntegration] error waiting for profile four creation: %v", err)
	}
	cfg := customerprofiles.InitWithDomain(domainWithCp, envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	err = WaitForExpectedCondition(func() bool {
		cpId, err := cfg.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, connectId4)
		if err != nil {
			t.Logf("[TestSkipCPIntegration] error, failed to get profile id")
			return false
		}
		if cpId == "" {
			t.Logf("[TestSkipCPIntegration] profile should not exist in cache")
			return false
		}
		cpId, err = cfg.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, connectId2)
		if err != nil {
			t.Logf("[TestSkipCPIntegration] profile should not exist in cache")
			return false
		}
		if cpId != "" {
			t.Logf("[TestSkipCPIntegration] profile should not exist in cache")
			return false
		}
		cpId, err = cfg.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, connectId1)
		if err != nil {
			t.Logf("[TestSkipCPIntegration] error, failed to get profile id")
			return false
		}
		if cpId == "" {
			t.Logf("[TestSkipCPIntegration] profile should not exist in cache")
			return false
		}
		connectId3, err := lc.GetProfileId("profile_id", profileThreeTravelerID)
		if err != nil {
			t.Logf("[TestSkipCPIntegration] error, failed to get profile id")
			return false
		}
		cpId, err = cfg.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, connectId3)
		if err != nil {
			t.Logf("[TestSkipCPIntegration] error, failed to get profile id")
			return false
		}
		prof, err := cfg.GetProfile(cpId, []string{airBookingType, clickstreamType}, []customerprofiles.PaginationOptions{})
		if err != nil {
			t.Logf("[TestSkipCPIntegration] error getting profile: %v", err)
			return false
		}
		if prof.FirstName == "Michael" {
			//if this is wrong, it will never resolve itself
			t.Fatalf("[TestSkipCPIntegration] wrong first name sent")
			return false
		}
		return true
	}, 24, 5*time.Second)
	if err != nil {
		t.Fatalf("[TestSkipCPIntegration] error waiting for expected condition: %v", err)
	}
}
*/

func WaitForProfileCreationByProfileId(profileId string, keyName string, profilesSvc2 *CustomerProfileConfigLowCost) (string, error) {
	i, max, duration := 0, 2, 5*time.Second
	for ; i < max; i++ {
		accpProfileId, _ := profilesSvc2.GetProfileId(keyName, profileId)
		if accpProfileId != "" {
			log.Printf("Successfully found Customer's Profile ID: %v", accpProfileId)
			return accpProfileId, nil
		}
		log.Printf("[TestCustomerProfiles] Profile is not available yet, waiting 5 seconds")
		time.Sleep(duration)
	}
	return "", errors.New("Profile not found after timeout")
}

func WaitForProfileCreationInCPByProfileId(profileId string, profilesSvc2 *customerprofiles.CustomerProfileConfig) (string, error) {
	i, max, duration := 0, 24, 5*time.Second
	for ; i < max; i++ {
		accpProfileId, _ := profilesSvc2.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, profileId)
		if accpProfileId != "" {
			log.Printf("Successfully found Customer's Profile ID: %v", accpProfileId)
			return accpProfileId, nil
		}
		log.Printf("[TestCustomerProfiles] Profile is not available yet, waiting 5 seconds")
		time.Sleep(duration)
	}
	return "", errors.New("Profile not found after timeout")
}

func WaitForExpectedCondition(action func() bool, maxAttempts int, duration time.Duration) error {
	for i := 0; i < maxAttempts; i++ {
		if action() {
			return nil
		}
		log.Printf("[WaitForExpectedCondition] condition not met, waiting %v", duration)
		time.Sleep(duration)
	}
	return errors.New("[WaitForExpectedCondition] condition not met after maximum attempts")
}

// Poll SQS queue until expected number of messages have been received or function times out.
//
// Messages that have been fetched will not be seen again for 1800 seconds, allowing subsequent
// batches of messages to be sent and tested without conflicts.
func fetchMessages(q sqs.IConfig, expectedCount int, interval int, max int) ([]sqs.Message, error) {
	var actualMessages []sqs.Message

	for i := 0; i < max; i++ {
		messages, err := q.Get(sqs.GetMessageOptions{VisibilityTimeout: 1800})
		if err != nil {
			return nil, fmt.Errorf("error getting messages from queue: %v", err)
		}
		actualMessages = append(actualMessages, messages.Peek...)

		if len(actualMessages) == expectedCount {
			return actualMessages, nil
		}

		log.Printf("Waiting %d seconds for %d remaining message(s)", interval, expectedCount-len(actualMessages))
		time.Sleep(time.Duration(interval) * time.Second)
	}

	return nil, fmt.Errorf("timed out waiting for messages, received %d of %d", len(actualMessages), expectedCount)
}

func TestPrepForCP(t *testing.T) {
	a := assert.New(t)
	// VALID CASE
	// Arrange
	conversation := `{"items": [{"from": "Agent", "to": "Traveler", "content": "Message body", "start_time": "2024-09-06T00:00:00.000000Z", "end_time": "2024-09-06T00:00:00.000000Z", "sentiment": "0.9"}]}`
	encoded := base64.StdEncoding.EncodeToString([]byte(conversation))
	csiObj := profilemodel.ProfileObject{Type: "customer_service_interaction", Attributes: map[string]string{"conversation": encoded}, AttributesInterface: map[string]interface{}{"conversation": conversation}}

	// Act
	prepForCP(&csiObj)

	// Assert
	a.Equal(conversation, csiObj.AttributesInterface["conversation"])

	// EMPTY CONVERSATION
	// Arrange
	conversation = ""
	csiObj = profilemodel.ProfileObject{Type: "customer_service_interaction", Attributes: map[string]string{"conversation": conversation}, AttributesInterface: map[string]interface{}{"conversation": conversation}}

	// Act
	prepForCP(&csiObj)

	// Assert
	a.Equal(conversation, csiObj.AttributesInterface["conversation"])
}
