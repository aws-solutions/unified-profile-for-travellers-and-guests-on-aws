// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"errors"
	"log"
	"strconv"
	"strings"
	customerprofiles "tah/upt/source/storage"
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	kms "tah/upt/source/tah-core/kms"
	"tah/upt/source/tah-core/sqs"
	model "tah/upt/source/ucp-common/src/model/admin"
	"tah/upt/source/ucp-common/src/utils/config"
	"tah/upt/source/ucp-common/src/utils/test"
	"testing"
	"time"
)

var fieldMappings = customerprofiles.FieldMappings{
	{
		Type:        "STRING",
		Source:      "_source.firstName",
		Target:      "_profile.FirstName",
		Indexes:     []string{},
		Searcheable: true,
	},
	{
		Type:        "STRING",
		Source:      "_source.lastName",
		Target:      "_profile.LastName",
		Indexes:     []string{},
		Searcheable: true,
	},
	{
		Type:        "STRING",
		Source:      "_source.traveller_id",
		Target:      "_profile.Attributes.profile_id",
		Indexes:     []string{"PROFILE"},
		Searcheable: true,
	},
	{
		Type:    "STRING",
		Source:  "_source.booking_id",
		Target:  "booking_id",
		Indexes: []string{"UNIQUE"},
		KeyOnly: true,
	},
	{
		Type:        customerprofiles.MappingTypeString,
		Source:      "_source." + customerprofiles.LAST_UPDATED_FIELD,
		Target:      customerprofiles.LAST_UPDATED_FIELD,
		Searcheable: true,
	},
}

func TestMergeAndUnmergeProfiles(t *testing.T) {
	// Set up resources
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] unable to load configs: %v", err)
	}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	kmsc := kms.Init(envCfg.Region, "", "")
	log.Printf("[TestMergeAndUnmergeProfiles] Setup test environment")
	name := "lcs-test-merges-" + core.GenerateUniqueId()
	lcsConfigTable, err := db.InitWithNewTable(name, "pk", "sk", "", "")
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] error creating lcs config table: %v", err)
	}
	t.Cleanup(func() {
		err = lcsConfigTable.DeleteTable(lcsConfigTable.TableName)
		if err != nil {
			t.Errorf("[TestMergeAndUnmergeProfiles] error deleting lcs config table: %v", err)
		}
	})
	err = lcsConfigTable.WaitForTableCreation()
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] error waiting for lcs config table: %v", err)
	}
	lcsKinesisStream, err := kinesis.InitAndCreate(name, envCfg.Region, "", "", core.LogLevelDebug)
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] error creating lcs kinesis stream: %v", err)
	}
	t.Cleanup(func() {
		err = lcsKinesisStream.Delete(lcsKinesisStream.Stream)
		if err != nil {
			t.Errorf("[TestMergeAndUnmergeProfiles] error deleting lcs kinesis stream: %v", err)
		}
	})
	_, err = lcsKinesisStream.WaitForStreamCreation(300)
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] error waiting for lcs kinesis stream: %v", err)
	}
	mergeQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = mergeQueueClient.CreateRandom("mergerMergeProfilesTest")
	if err != nil {
		t.Fatalf("error creating merge queue: %v", err)
	}
	t.Cleanup(func() {
		err = mergeQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting merge queue: %v", err)
		}
	})
	cpWriterQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = cpWriterQueueClient.CreateRandom("cpWriterMergeProfilesTest")
	if err != nil {
		t.Fatalf("error creating cp writer queue: %v", err)
	}
	t.Cleanup(func() {
		err = cpWriterQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting cp writer queue: %v", err)
		}
	})
	initParams := test.ProfileStorageParams{
		LcsConfigTable:      &lcsConfigTable,
		LcsKinesisStream:    lcsKinesisStream,
		MergeQueueClient:    &mergeQueueClient,
		CPWriterQueueClient: &mergeQueueClient,
	}
	profileStorage, err := test.InitProfileStorage(initParams)
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] Error initializing profile storage %v", err)
	}
	keyArn, err := kmsc.CreateKey("tah-unit-test-key")
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] Could not create KMS key to unit test UCP %v", err)
	}
	t.Cleanup(func() {
		err = kmsc.DeleteKey(keyArn)
		if err != nil {
			t.Errorf("[TestMergeAndUnmergeProfiles] Could not delete KMS key to unit test UCP %v", err)
		}
	})

	// Admin Creation Tasks
	domainName := "test_domain_2" + strings.ToLower(core.GenerateUniqueId())
	options := customerprofiles.DomainOptions{
		AiIdResolutionOn:       false,
		RuleBaseIdResolutionOn: true,
	}
	err = profileStorage.CreateDomainWithQueue(domainName, keyArn, map[string]string{"envName": "test"}, "", "", options)
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] Error creating domain %v", err)
	}
	t.Cleanup(func() {
		err = profileStorage.DeleteDomain()
		if err != nil {
			t.Errorf("[TestMergeAndUnmergeProfiles] Error deleting domain %v", err)
		}
	})

	objectName := "sample_object"
	err = profileStorage.CreateMapping(objectName, "description of the test mapping", fieldMappings)
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] Error creating mapping %v", err)
	}

	testSize := 20

	// Profile Management
	log.Printf("[TestMergeAndUnmergeProfiles] Creating %v Profiles", testSize)
	for i := 0; i < testSize; i++ {
		var profileObject1 = `{"firstName": "First", "lastName": "Last", "traveller_id": "source_id_` + strconv.Itoa(i) + `", "booking_id": "booking_source` + strconv.Itoa(i) + `", "last_updated": "` + time.Now().UTC().Format(customerprofiles.INGEST_TIMESTAMP_FORMAT) + `"}`
		var profileObject2 = `{"firstName": "First", "lastName": "Last", "traveller_id": "target_id_` + strconv.Itoa(i) + `", "booking_id": "booking_target` + strconv.Itoa(i) + `", "last_updated": "` + time.Now().UTC().Format(customerprofiles.INGEST_TIMESTAMP_FORMAT) + `"}`
		err = profileStorage.PutProfileObject(profileObject1, objectName)
		if err != nil {
			t.Fatalf("Error putting profile object %v", err)
		}
		profileStorage.PutProfileObject(profileObject2, objectName)
		if err != nil {
			t.Fatalf("Error putting profile object %v", err)
		}
	}

	for i := testSize - 1; i >= 0; i-- {
		_, err := WaitForProfileCreationByProfileId("source_id_"+strconv.Itoa(i), profileStorage)
		if err != nil {
			t.Fatalf("[TestMergeAndUnmergeProfiles] Error waiting for source profiles creation %v", err)
		}
		_, err = WaitForProfileCreationByProfileId("target_id_"+strconv.Itoa(i), profileStorage)
		if err != nil {
			t.Fatalf("[TestMergeAndUnmergeProfiles] Error waiting for target profiles creation %v", err)
		}
	}

	profilePairs := []model.MergeRq{}
	profileIdMapping := map[string]string{}
	for i := 0; i < testSize; i++ {
		accpProfileId1, err := profileStorage.GetProfileId("profile_id", "source_id_"+strconv.Itoa(i))
		if err != nil {
			t.Fatalf("[TestMergeAndUnmergeProfiles] Error getting profile id %v", err)
		}
		if accpProfileId1 == "" {
			_, err := WaitForProfileCreationByProfileId("source_id_"+strconv.Itoa(i), profileStorage)
			if err != nil {
				t.Fatalf("[TestMergeAndUnmergeProfiles] Error waiting for source profiles creation %v", err)
			}
			accpProfileId1, _ = profileStorage.GetProfileId("profile_id", "source_id_"+strconv.Itoa(i))
		}

		accpProfileId2, err := profileStorage.GetProfileId("profile_id", "target_id_"+strconv.Itoa(i))
		if err != nil {
			t.Fatalf("[TestMergeAndUnmergeProfiles] Error getting profile id %v", err)
		}
		if accpProfileId2 == "" {
			_, err := WaitForProfileCreationByProfileId("target_id_"+strconv.Itoa(i), profileStorage)
			if err != nil {
				t.Fatalf("[TestMergeAndUnmergeProfiles] Error waiting for target_id_ profiles creation %v", err)
			}
			accpProfileId2, _ = profileStorage.GetProfileId("profile_id", "target_id_"+strconv.Itoa(i))
		}
		log.Printf("[TestMergeAndUnmergeProfiles] Adding  profile pair (%s,%s) to merge", accpProfileId1, accpProfileId2)
		profilePairs = append(profilePairs, model.MergeRq{SourceProfileID: accpProfileId1, TargetProfileID: accpProfileId2})
		profileIdMapping["source_id_"+strconv.Itoa(i)] = accpProfileId1
		profileIdMapping["target_id_"+strconv.Itoa(i)] = accpProfileId2
	}
	log.Printf("[TestMergeAndUnmergeProfiles] Merging %v profile pairs", testSize)
	err = MergeProfiles(tx, profileStorage, domainName, profilePairs)
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] Error merging %v profile pairs %v", testSize, err)
	}

	success, failure, err := WaitForProfileMergeCompletion(tx, profileStorage, domainName, profilePairs, 60)
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] Error waiting for merge completion %v", err)
	}
	if len(failure) > 0 {
		t.Fatalf("[TestMergeAndUnmergeProfiles] Failed to merge %v profile pairs", len(failure))
		for _, f := range failure {
			t.Error("Error merging profile pair: ", f)
		}
	}
	if len(success) != testSize {
		t.Fatalf(
			"[TestMergeAndUnmergeProfiles]  Successfully merged profile pairs should be of size %v  and not %v",
			testSize,
			len(success),
		)
	}

	// Unmerging a previously merged profile pair
	err = UnmergeProfiles(tx, profileStorage, domainName, model.UnmergeRq{
		MergedIntoConnectID: profilePairs[0].SourceProfileID,
		ToUnmergeConnectID:  profilePairs[0].TargetProfileID,
		InteractionType:     objectName,
	})
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] Error performing unmerge %v", err)
	}

	// This checks if unmerged profile is recreated in dynamo and aurora
	err = ValidateProfileUnmergeCompletion(tx, profileStorage, domainName, model.UnmergeRq{
		MergedIntoConnectID: profilePairs[0].SourceProfileID,
		ToUnmergeConnectID:  profilePairs[0].TargetProfileID,
		InteractionType:     objectName,
	}, 60)
	if err != nil {
		t.Fatalf("[TestMergeAndUnmergeProfiles] Error performing unmerge %v", err)
	}
}

func WaitForProfileCreationByProfileId(profileId string, profileStorage customerprofiles.ICustomerProfileLowCostConfig) (string, error) {
	i, max, duration := 0, 12, 5*time.Second
	for ; i < max; i++ {
		accpProfileId, _ := profileStorage.GetProfileId("profile_id", profileId)
		if accpProfileId != "" {
			log.Printf("Successfully found Customer's Profile ID: %v", accpProfileId)
			return accpProfileId, nil
		}
		log.Printf("[TestCustomerProfiles] Profile is not available yet, waiting 5 seconds")
		time.Sleep(duration)
	}
	return "", errors.New("Profile not found after timeout")
}

func WaitForMerge(accpProfileId string, profileId2 string, profileStorage customerprofiles.ICustomerProfileLowCostConfig) error {
	log.Printf("[TestCustomerProfiles] Waiting for profile to be deleted")
	i, max, duration := 0, 12, 5*time.Second
	for ; i < max; i++ {
		accpProfileId2, _ := profileStorage.GetProfileId("profile_id", profileId2)

		if accpProfileId == accpProfileId2 {
			log.Printf("Target profile %s returns connectID %s: Merge successfull", profileId2, accpProfileId)
			return nil
		}
		log.Printf(
			"[TestCustomerProfiles] Both profiles %s and %s have different connect IDs. Merge not yet completed",
			accpProfileId,
			accpProfileId2,
		)
		time.Sleep(duration)
	}
	return errors.New("Profile merge failed after timeout")
}
