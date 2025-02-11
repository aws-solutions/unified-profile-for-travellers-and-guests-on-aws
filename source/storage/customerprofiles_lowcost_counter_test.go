// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	"encoding/json"
	"log"
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/kms"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicCounting(t *testing.T) {
	t.Parallel()
	aurora, err := SetupAurora(t)
	if err != nil {
		log.Fatalf("error setting up aurora: %v", err)
	}
	dh := DataHandler{AurSvc: aurora}

	// initial table setup
	domainName := "basic_count_" + strings.ToLower(core.GenerateUniqueId())
	err = dh.CreateRecordCountTable(domainName)
	if err != nil {
		t.Fatalf("error creating table: %v", err)
	}
	err = dh.InitializeCount(domainName, []string{countTableProfileObject})
	if err != nil {
		t.Fatalf("error initializing count: %v", err)
	}
	count, err := dh.GetCount(domainName, countTableProfileObject)
	if err != nil {
		t.Errorf("error getting count: %v", err)
	}
	t.Cleanup(func() {
		err = dh.DeleteRecordCountTable(domainName)
		if err != nil {
			t.Errorf("error deleting table: %v", err)
		}
	})
	assert.Equal(t, int64(0), count)

	con, err := dh.AurSvc.AcquireConnection(dh.Tx)
	if err != nil {
		t.Errorf("Error acquiring connection: %v", err)
	}
	defer con.Release()

	// increment
	err = dh.UpdateCount(con, domainName, countTableProfileObject, 5)
	if err != nil {
		t.Errorf("error incrementing count: %v", err)
	}
	count, err = dh.GetCount(domainName, countTableProfileObject)
	if err != nil {
		t.Errorf("error getting count: %v", err)
	}
	assert.Equal(t, int64(5), count)

	// decrement
	err = dh.UpdateCount(con, domainName, countTableProfileObject, -3)
	if err != nil {
		t.Errorf("error decrementing count: %v", err)
	}
	count, err = dh.GetCount(domainName, countTableProfileObject)
	if err != nil {
		t.Errorf("error getting count: %v", err)
	}
	assert.Equal(t, int64(2), count)

	// get all object counts
	counts, err := dh.GetAllObjectCounts(domainName)
	if err != nil {
		t.Errorf("error getting all object counts: %v", err)
	}
	assert.Equal(t, 1, len(counts))
	assert.Equal(t, int64(2), counts[countTableProfileObject])
}

func TestRecordCount(t *testing.T) {
	t.Parallel()

	// Domain setup
	lcs, _ := setupAndRegisterCleanup(t, setupOptions{domainOptions: DomainOptions{RuleBaseIdResolutionOn: true}, domainNamePrefix: "record_count"})

	// Ingest records
	objects := []genericTestProfileObject{
		// new travelers
		{"obj_id_1", "traveler_1", "", "", "", "", "", "", "", utcNow},
		{"obj_id_2", "traveler_2", "", "", "", "", "", "", "", utcNow},
		{"obj_id_3", "traveler_3", "", "", "", "", "", "", "", utcNow},
		{"obj_id_4", "traveler_4", "", "", "", "", "", "", "", utcNow},
		{"obj_id_5", "traveler_5", "", "", "", "", "", "", "", utcNow},
		{"obj_id_6", "traveler_6", "", "", "", "", "", "", "", utcNow},
		{"obj_id_7", "traveler_7", "", "", "", "", "", "", "", utcNow},
		{"obj_id_8", "traveler_8", "", "", "", "", "", "", "", utcNow},
		{"obj_id_9", "traveler_9", "", "", "", "", "", "", "", utcNow},
		{"obj_id_10", "traveler_10", "", "", "", "", "", "", "", utcNow},
		// existing travelers
		{"obj_id_11", "traveler_1", "", "", "", "", "", "", "", utcNow},
		{"obj_id_12", "traveler_1", "", "", "", "", "", "", "", utcNow},
		{"obj_id_13", "traveler_2", "", "", "", "", "", "", "", utcNow},
		{"obj_id_14", "traveler_3", "", "", "", "", "", "", "", utcNow},
	}
	var jsonRecs []string
	for _, obj := range objects {
		jsonRec, err := json.Marshal(obj)
		if err != nil {
			t.Fatalf("error marshalling json: %v", err)
		}
		jsonRecs = append(jsonRecs, string(jsonRec))
	}
	ingestTestRecords(t, &lcs, jsonRecs)

	// Check counts
	count, err := lcs.Data.GetCount(lcs.DomainName, countTableProfileObject)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	assert.Equal(t, int64(10), count)

	// Perform a merge
	id1, err := lcs.GetProfileId(LC_PROFILE_ID_KEY, "traveler_1")
	if err != nil {
		t.Fatalf("error getting profile id 1: %v", err)
	}
	id2, err := lcs.GetProfileId(LC_PROFILE_ID_KEY, "traveler_2")
	if err != nil {
		t.Fatalf("error getting profile id 2: %v", err)
	}
	_, err = lcs.MergeProfiles(id1, id2, ProfileMergeContext{MergeType: MergeTypeManual})
	if err != nil {
		t.Fatalf("error merging profiles: %v", err)
	}

	// Check counts
	count, err = lcs.Data.GetCount(lcs.DomainName, countTableProfileObject)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	assert.Equal(t, int64(9), count)

	// Delete two profiles
	id3, err := lcs.GetProfileId(LC_PROFILE_ID_KEY, "traveler_3")
	if err != nil {
		t.Fatalf("error getting profile id 3: %v", err)
	}
	id4, err := lcs.GetProfileId(LC_PROFILE_ID_KEY, "traveler_4")
	if err != nil {
		t.Fatalf("error getting profile id 4: %v", err)
	}
	err = lcs.DeleteProfile(id3)
	if err != nil {
		t.Fatalf("error deleting profile 3: %v", err)
	}
	err = lcs.DeleteProfile(id4)
	if err != nil {
		t.Fatalf("error deleting profile 4: %v", err)
	}

	// Check counts
	count, err = lcs.Data.GetCount(lcs.DomainName, countTableProfileObject)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	assert.Equal(t, int64(7), count)

	// Unmerge profile
	_, err = lcs.UnmergeProfiles(id2, id1, "", "")
	if err != nil {
		t.Fatalf("error unmerging profiles: %v", err)
	}

	// Check counts
	count, err = lcs.Data.GetCount(lcs.DomainName, countTableProfileObject)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	assert.Equal(t, int64(8), count)

	// Validate all counts
	counts, err := lcs.Data.GetAllObjectCounts(lcs.DomainName)
	if err != nil {
		t.Fatalf("error getting all object counts: %v", err)
	}
	assert.Equal(t, counts[countTableProfileObject], int64(8))
	assert.Equal(t, counts[genericTestObjectType_Booking], int64(12))
}

type setupOptions struct {
	// Create a mock event stream by default.
	// If true, a real event stream will be created.
	shouldCreateEventStream bool
	// By default, providing a nil slice will result in generateTestMappings()
	// being used to create a single mapping for testing with.
	customObjectMappings []ObjectMapping
	// Optionally provide a custom domain options struct.
	domainOptions DomainOptions
	// domain name prefix for created domain
	domainNamePrefix string
}

type setupOutput struct {
	mappingNames []string
	kmsKeyArn    string
}

// Create (and clean up) all necessary resources to set up a domain with mappings. Using t, we also handle
// any failures along the way by immediately failing the test.
func setupAndRegisterCleanup(t *testing.T, options setupOptions) (CustomerProfileConfigLowCost, setupOutput) {
	// General setup
	domainName := strings.ToLower(options.domainNamePrefix + "_" + core.GenerateUniqueId())
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	// Config table
	configTableConfig, err := db.InitWithNewTable(domainName, configTablePk, configTableSk, "", "")
	if err != nil {
		t.Fatalf("error setting up config table: %v", err)
	}
	t.Cleanup(func() {
		err = configTableConfig.DeleteTable(domainName)
		if err != nil {
			t.Errorf("error deleting config table: %v", err)
		}
	})
	// KMS key
	kmsClient := kms.Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	keyArn, err := kmsClient.CreateKey("lcs-unit-test")
	if err != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err)
	}
	t.Cleanup(func() {
		err = kmsClient.DeleteKey(keyArn)
		if err != nil {
			t.Errorf("[TestUpdateProfile] DeleteTable failed: %s", err)
		}
	})
	// Merge queue
	mergeQueueConfig := sqs.Init(envCfg.Region, "", "")
	_, err = mergeQueueConfig.Create(domainName)
	if err != nil {
		t.Fatalf("error creating merge queue: %v", err)
	}
	t.Cleanup(func() {
		err = mergeQueueConfig.Delete()
		if err != nil {
			t.Errorf("error deleting merge queue: %v", err)
		}
	})
	// CP Writer Queue
	cpWriterQueueConfig := sqs.Init(envCfg.Region, "", "")
	_, err = cpWriterQueueConfig.Create(domainName + "CPWriter")
	if err != nil {
		t.Fatalf("error creating CP writer queue: %v", err)
	}
	t.Cleanup(func() {
		err = cpWriterQueueConfig.Delete()
		if err != nil {
			t.Errorf("error deleting CP writer queue: %v", err)
		}
	})
	// Customer Profiles DLQ
	cpDlqConfig := sqs.Init(envCfg.Region, "", "")
	_, err = cpDlqConfig.Create(domainName + "DLQ")
	if err != nil {
		t.Fatalf("error creating CP DLQ: %v", err)
	}
	t.Cleanup(func() {
		err = cpDlqConfig.Delete()
		if err != nil {
			t.Errorf("error deleting CP DLQ: %v", err)
		}
	})
	err = cpDlqConfig.SetPolicy("Service", "profile.amazonaws.com", []string{"SQS:*"})
	if err != nil {
		t.Fatalf("error setting policy on CP DLQ: %v", err)
	}
	// Aurora
	auroraConfig, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("error setting up aurora: %v", err)
	}
	// Event stream
	var eventStreamConfig kinesis.IConfig
	if options.shouldCreateEventStream {
		cfg, err := kinesis.InitAndCreate(domainName, envCfg.Region, "", "", core.LogLevelDebug)
		if err != nil {
			t.Fatalf("error setting up event stream: %v", err)
		}
		eventStreamConfig = cfg
	} else {
		eventStreamConfig = kinesis.InitMock(nil, nil, nil, nil)
	}
	// Wait for resource creation
	err = configTableConfig.WaitForTableCreation()
	if err != nil {
		t.Fatalf("error waiting for config table to be created: %v", err)
	}
	// Init LCS
	initOptions := CustomerProfileInitOptions{
		MergeQueueClient: &mergeQueueConfig,
	}
	lcs := InitLowCost(
		envCfg.Region,
		auroraConfig,
		&configTableConfig,
		eventStreamConfig,
		&cpWriterQueueConfig,
		"tah/upt/source/tah-core/customerprofiles/test",
		"",
		"",
		core.LogLevelDebug,
		&initOptions,
		InitLcsCache(),
	)
	// Create domain
	err = lcs.CreateDomainWithQueue(
		domainName,
		keyArn,
		map[string]string{DOMAIN_CONFIG_ENV_KEY: t.Name()},
		cpDlqConfig.QueueUrl,
		"",
		options.domainOptions,
	)
	if err != nil {
		t.Fatalf("error creating domain: %v", err)
	}
	t.Cleanup(func() {
		err = lcs.DeleteDomain()
		if err != nil {
			t.Errorf("error deleting domain: %v", err)
		}
	})
	lcs.SetDomain(domainName)
	// Create mapping
	if options.customObjectMappings != nil {
		for _, mapping := range options.customObjectMappings {
			err = lcs.CreateMapping(mapping.Name, mapping.Name+" "+mapping.Version, mapping.Fields)
			if err != nil {
				t.Fatalf("error creating mapping: %v", err)
			}
		}
	} else {
		err = lcs.CreateMapping(genericTestObjectType_Booking, "test booking object", generateTestMappings())
		if err != nil {
			t.Fatalf("error creating mapping: %v", err)
		}
	}

	output := setupOutput{
		kmsKeyArn: keyArn,
	}
	if len(options.customObjectMappings) > 0 {
		output.mappingNames = make([]string, len(options.customObjectMappings))
		for i, mapping := range options.customObjectMappings {
			output.mappingNames[i] = mapping.Name
		}
	} else {
		output.mappingNames = []string{customerprofiles.PROFILE_FIELD_OBJECT_TYPE, genericTestObjectType_Booking}
	}
	return *lcs, output
}
