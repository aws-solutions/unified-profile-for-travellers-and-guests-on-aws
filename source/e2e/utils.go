// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	lcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/ecs"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/tah-core/ssm"
	"tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/config"
	"tah/upt/source/ucp-common/src/utils/utils"
	uptSdk "tah/upt/source/ucp-sdk/src"
	"testing"
	"time"
)

type ProfileChangeEvent struct {
	Version    string                   `json:"version"`
	ID         string                   `json:"id"`
	DetailType string                   `json:"detail-type"`
	Source     string                   `json:"source"`
	Detail     ProfileChangeEventDetail `json:"detail"`
}
type ProfileChangeEventDetail struct {
	TravellerID string        `json:"travelerID"`
	Domain      string        `json:"domain"`
	Events      []DetailEvent `json:"events"`
}
type DetailEvent struct {
	EventType  string          `json:"eventType"`
	ObjectType string          `json:"objectType"`
	Data       DetailEventData `json:"data"`
}
type DetailEventData struct {
	ModelVersion      string       `json:"modelVersion"`
	TravellerID       string       `json:"travellerId"`
	AirBookingRecords []AirBooking `json:"airBookingRecords"`
	// ...additional fields available
}

type AirBooking struct {
	TravellerID string `json:"travellerId"`
}

type ConfigRecord struct {
	Pk string `json:"pk"`
	Sk string `json:"sk"`
}

func GetDynamoCacheCounts(t *testing.T, dynamoCache db.IDBConfig) int {
	data := []map[string]interface{}{}
	totalCount := 0
	lastEval, err := dynamoCache.FindAll(&data, nil)
	if err != nil {
		t.Fatalf("[%s] Error scanning Dynamo Cache: %v", t.Name(), err)
	}
	totalCount += len(data)
	for lastEval != nil {
		lastEval, err = dynamoCache.FindAll(&data, lastEval)
		if err != nil {
			t.Fatalf("[%s] Error scanning Dynamo Cache: %v", t.Name(), err)
		}
		totalCount += len(data)
	}
	return totalCount
}

func GetCpCacheCounts(t *testing.T, cpCache customerprofiles.ICustomerProfileConfig) (int64, int64) {
	dom, err := cpCache.GetDomain()
	if err != nil {
		t.Fatalf("[%s] Error getting CP domain: %v", t.Name(), err)
	}
	return dom.NProfiles, dom.NObjects
}

func WaitForExpectedCondition(action func() bool, maxAttempts int, duration time.Duration) error {
	for i := 0; i < maxAttempts; i++ {
		log.Printf("[WaitForExpectedCondition] attempt %v, waiting %v", i+1, duration)
		time.Sleep(duration)
		if action() {
			return nil
		}
		log.Printf("[WaitForExpectedCondition] attempt %v, condition not met", i+1)
	}
	return fmt.Errorf("[WaitForExpectedCondition] condition not met after maximum attempts")
}

func SendKinesisBatch(t *testing.T, stream *kinesis.Config, kinesisRate int, records []kinesis.Record) {
	numSent := 0
	for i := 0; i < len(records); i += kinesisRate {
		var batch []kinesis.Record
		if i+kinesisRate > len(records) {
			batch = records[i:]
		} else {
			batch = records[i : i+kinesisRate]
		}
		err, errs := stream.PutRecords(batch)
		if err != nil || len(errs) != 0 {
			t.Fatalf("[%v] error sending data to stream: %v. %+v", t.Name(), err, errs)
		}
		numSent += len(batch)
		t.Logf("[%s] sent batch %v, total of %v records sent", t.Name(), i/kinesisRate+1, numSent)
	}
}

func RandomDomain(prefix string) string {
	return fmt.Sprintf("%s_%s", strings.ToLower(prefix), strings.ToLower(core.GenerateUniqueId()))
}

func InitializeEndToEndTestEnvironment(t *testing.T, infraConfig config.InfraConfig, envConfig config.EnvConfig) (string, *uptSdk.UptHandler, *db.DBConfig, *customerprofiles.CustomerProfileConfig, *kinesis.Config, error) {
	//as determined by testing, our limit for number of domains is 19, so making this just the a character + uuid
	//future story: look to lower the length of the "customer_service_interaction_object_history" table, and enforce the
	//domain limit
	domainName := RandomDomain("t")
	t.Cleanup(func() {
		configTable := infraConfig.LcsConfigTableName
		dbConfig := db.Init(configTable, "pk", "sk", "", "")
		var items []ConfigRecord
		err := dbConfig.FindByPk(domainName, &items)
		if err != nil {
			t.Errorf("Error finding items to delete: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("All storage config items for this domain should be deleted, but some remain")
			err = dbConfig.DeleteMany(items)
			if err != nil {
				t.Errorf("Error deleting items: %v", err)
			}
		}
	})

	// Upt Sdk
	uptHandler, err := uptSdk.Init(infraConfig)
	if err != nil {
		return "", nil, nil, nil, nil, fmt.Errorf("error initializing SDK: %v", err)
	}

	err = uptHandler.SetupTestAuth(domainName)
	if err != nil {
		return "", nil, nil, nil, nil, fmt.Errorf("error setting up auth: %v", err)
	}

	t.Cleanup(func() {
		err := uptHandler.CleanupAuth(domainName)
		if err != nil {
			t.Errorf("Error cleaning up auth: %v", err)
		}
	})

	// Create Domain
	err = uptHandler.CreateDomainAndWait(domainName, 300)
	if err != nil {
		return "", nil, nil, nil, nil, fmt.Errorf("error creating domain: %v", err)
	}

	t.Cleanup(func() {
		err := uptHandler.DeleteDomainAndWait(domainName)
		if err != nil {
			t.Errorf("Error deleting domain: %v", err)
		}
	})

	// Init Cache
	dynamoCache := db.Init("ucp_domain_"+domainName, "connect_id", "record_type", "", "")
	cpCache := customerprofiles.InitWithDomain(domainName, envConfig.Region, "", "")

	// Init Kinesis Stream
	realTimeStream := kinesis.Init(infraConfig.RealTimeIngestorStreamName, envConfig.Region, "", "", core.LogLevelDebug)

	return domainName, &uptHandler, &dynamoCache, cpCache, realTimeStream, nil
}

func CamelToSnakeCase(s string) string {
	var result strings.Builder
	for i, ch := range s {
		if i > 0 && ch >= 'A' && ch <= 'Z' {
			result.WriteByte('_')
		}
		if ch >= 'A' && ch <= 'Z' {
			result.WriteByte(byte(ch + 32))
		} else {
			result.WriteByte(byte(ch))
		}
	}
	return result.String()
}

func GetEventsFromSqs(sqsConfig *sqs.Config, sqsTimeout int) ([]ProfileChangeEvent, error) {
	log.Printf("[TestChangeEvent] Validating event export")
	allMessages := make(map[string]ProfileChangeEvent)
	i, max, duration := 0, sqsTimeout/5, 5*time.Second
	for i < max {
		messages, err := sqsConfig.Get(sqs.GetMessageOptions{})
		if err != nil {
			return []ProfileChangeEvent{}, err
		}
		for _, peek := range messages.Peek {
			var event ProfileChangeEvent
			err := json.Unmarshal([]byte(peek.Body), &event)
			if err != nil {
				return []ProfileChangeEvent{}, err
			}
			allMessages[event.ID] = event
		}
		if len(allMessages) == int(messages.NMessages) {
			break
		}
		log.Println("[TestChangeEvent] Waiting 5 seconds for events to propagate")
		time.Sleep(duration)
		i++
	}

	log.Printf("[TestChangeEvent] Received %d messages", len(allMessages))
	ebEvents := make([]ProfileChangeEvent, 0, len(allMessages))
	// Create event array
	for _, value := range allMessages {
		ebEvents = append(ebEvents, value)
	}
	if i == max {
		return ebEvents, errors.New("timed out waiting for events to propagate")
	}
	return ebEvents, nil
}

func GetProfileWithTravellerId(t *testing.T, uptHandler *uptSdk.UptHandler, domainName string, travellerId string, objectTypeNames []string) (traveler.Traveller, bool) {
	connectIdOne, err := uptHandler.GetProfileId(domainName, travellerId)
	if err != nil {
		t.Logf("Error getting profile id from traveler id: %v", err)
		return traveler.Traveller{}, false
	}
	profileResp, err := uptHandler.RetreiveProfile(domainName, connectIdOne, objectTypeNames)
	if err != nil {
		t.Logf("Error retrieving profile: %v", err)
		return traveler.Traveller{}, false
	}
	profiles := utils.SafeDereference(profileResp.Profiles)
	if len(profiles) != 1 {
		t.Logf("Expected 1 profile, got %d", len(profiles))
		return traveler.Traveller{}, false
	}
	profileOne := profiles[0]
	return profileOne, true
}

func GetProfileFromAccp(t *testing.T, cpConfig *customerprofiles.CustomerProfileConfig, domainName string, lcsId string, objectTypeNames []string) (string, profilemodel.Profile, bool) {
	cpId, err := cpConfig.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, lcsId)
	if err != nil {
		t.Logf("Error getting accp id: %v", err)
		return "", profilemodel.Profile{}, false
	}
	if cpId == "" {
		t.Logf("No accp id found for lcs id %s", lcsId)
		return "", profilemodel.Profile{}, false
	}
	profileCp, err := cpConfig.GetProfile(cpId, objectTypeNames, []customerprofiles.PaginationOptions{})
	if err != nil {
		t.Logf("Error retrieving profile from CP: %v", err)
		return "", profilemodel.Profile{}, false
	}
	return cpId, profileCp, true
}

func GetProfileFromAccpWithTravellerId(t *testing.T, cpConfig *customerprofiles.CustomerProfileConfig, domainName string, travellerId string) (profilemodel.Profile, bool) {
	cpProfile, err := cpConfig.SearchProfiles("traveller_id", []string{travellerId})
	if err != nil {
		t.Logf("Error searching for profile with traveller id: %v", err)
		return profilemodel.Profile{}, false
	}
	if len(cpProfile) != 1 {
		t.Logf("Expected 1 profile, got %d", len(cpProfile))
		return profilemodel.Profile{}, false
	}
	return cpProfile[0], true
}

func GetProfileRecordsFromDynamo(t *testing.T, dynamoConfig *db.DBConfig, domainName string, lcsId string, objectTypeNames []string) (lcs.DynamoProfileRecord, map[string][]lcs.DynamoProfileRecord, bool) {
	profileRecordsMain := []lcs.DynamoProfileRecord{}
	err := dynamoConfig.FindStartingWith(lcsId, lcs.PROFILE_OBJECT_TYPE_PREFIX+"main", &profileRecordsMain)
	if err != nil {
		t.Logf("Error getting dynamo records: %v", err)
		return lcs.DynamoProfileRecord{}, map[string][]lcs.DynamoProfileRecord{}, false
	}
	if len(profileRecordsMain) != 1 {
		t.Logf("Expected 1 profile record, got %d", len(profileRecordsMain))
		return lcs.DynamoProfileRecord{}, map[string][]lcs.DynamoProfileRecord{}, false
	}
	profile := profileRecordsMain[0]
	profileObjects := map[string][]lcs.DynamoProfileRecord{}
	for _, objectName := range objectTypeNames {
		profileRecords := []lcs.DynamoProfileRecord{}
		err = dynamoConfig.FindStartingWith(lcsId, lcs.PROFILE_OBJECT_TYPE_PREFIX+objectName, &profileRecords)
		if err != nil {
			t.Logf("Error getting dynamo records: %v", err)
			return lcs.DynamoProfileRecord{}, map[string][]lcs.DynamoProfileRecord{}, false
		}
		profileObjects[objectName] = profileRecords
	}
	return profile, profileObjects, true
}

func InitEcsConfig(t *testing.T, infraConfig config.InfraConfig, envConfig config.EnvConfig) (*ecs.EcsConfig, ecs.TaskConfig) {
	// Init ECS
	ssmConfig, err := ssm.InitSsm(context.TODO(), envConfig.Region, "", "")
	if err != nil {
		t.Fatalf("[%v] Error initializing ssm: %v", t.Name(), err)
	}
	namespace := infraConfig.SsmParamNamespace
	params, err := ssmConfig.GetParametersByPath(context.TODO(), namespace)
	if err != nil {
		t.Fatalf("[%s] Error getting ssm params %v", t.Name(), err)
	}
	clusterArn, ok := params["RebuildCacheClusterArn"]
	if !ok {
		t.Fatalf("[%v] Error getting cluster arn from ssm", t.Name())
	}
	taskDefinition, ok := params["RebuildCacheTaskDefinitionArn"]
	if !ok {
		t.Fatalf("[%v] Error getting task definition arn from ssm", t.Name())
	}
	containerName, ok := params["RebuildCacheContainerName"]
	if !ok {
		t.Fatalf("[%v] Error getting container name from ssm", t.Name())
	}
	subnet, ok := params["VpcSubnets"]
	if !ok {
		t.Fatalf("[%v] Error getting subnets from ssm", t.Name())
	}
	subnetIds := strings.Split(subnet, ",")
	securityGroup, ok := params["VpcSecurityGroup"]
	if !ok {
		t.Fatalf("[%v] Error getting security groups from ssm", t.Name())
	}
	securityGroups := []string{securityGroup}

	ecsConfig, err := ecs.InitEcs(context.TODO(), envConfig.Region, "", "")
	if err != nil {
		t.Fatalf("[%v] Error initializing ecs: %v", t.Name(), err)
	}

	taskConfig := ecs.TaskConfig{
		ClusterArn:        clusterArn,
		TaskDefinitionArn: taskDefinition,
		ContainerName:     containerName,
		SubnetIds:         subnetIds,
		SecurityGroupIds:  securityGroups,
	}

	return ecsConfig, taskConfig
}
