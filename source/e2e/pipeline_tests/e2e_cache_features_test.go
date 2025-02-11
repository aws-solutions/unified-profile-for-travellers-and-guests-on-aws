// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2epipeline

import (
	"strconv"
	"sync"
	csi "tah/upt/schemas/src/tah-common/common"
	util "tah/upt/source/e2e"
	"tah/upt/source/e2e/generators"
	lcs "tah/upt/source/storage"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

func TestTravellerIdSearch(t *testing.T) {
	t.Parallel()
	kinesisRate := 1000
	NUM_RECORDS := 10
	firstName := "John"
	travellerIdPrefix := "travId"
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}

	domainName, uptHandler, _, cpConfig, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	_, err = uptHandler.SaveIdResRuleSet(domainName, []lcs.Rule{
		{
			Index:       0,
			Name:        "sameFirstName",
			Description: "Merge profiles with matching first name",
			Conditions: []lcs.Condition{
				{
					Index:               0,
					ConditionType:       lcs.CONDITION_TYPE_MATCH,
					IncomingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					IncomingObjectField: "FirstName",
					Op:                  lcs.RULE_OP_EQUALS,
					ExistingObjectType:  lcs.PROFILE_OBJECT_TYPE_NAME,
					ExistingObjectField: "FirstName",
					IndexNormalization: lcs.IndexNormalizationSettings{
						Lowercase: true,
						Trim:      true,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Error saving rule set: %v", err)
	}

	_, err = uptHandler.ActivateIdResRuleSetWithRetries(domainName, 5)
	if err != nil {
		t.Fatalf("Error activating rule set: %v", err)
	}

	for i := 0; i < NUM_RECORDS; i++ {
		travellerId := travellerIdPrefix + strconv.Itoa(i)
		csiRecord, err := generators.GenerateSpecificCsiRecord(travellerId, domainName, func(gp *csi.CustomerServiceInteraction) {
			gp.FirstName = firstName
		})
		if err != nil {
			t.Fatalf("Error generating CSI: %v", err)
		}
		util.SendKinesisBatch(t, realTimeStream, kinesisRate, csiRecord)
	}

	err = util.WaitForExpectedCondition(func() bool {
		var wg sync.WaitGroup
		existsChannel := make([]bool, NUM_RECORDS)
		for i := 0; i < NUM_RECORDS; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				travellerId := travellerIdPrefix + strconv.Itoa(i)
				profileOne, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, travellerId, []string{"csi"})
				if !exists {
					existsChannel[i] = false
					return
				}
				//this check fails occasionally, which is a problem
				if len(profileOne.CustomerServiceInteractionRecords) != NUM_RECORDS {
					t.Logf("Expected %v CSI records, got %v", NUM_RECORDS, len(profileOne.CustomerServiceInteractionRecords))
					existsChannel[i] = false
					return
				}
				cpProfile, err := cpConfig.SearchProfiles("traveller_id", []string{travellerId})
				if err != nil {
					t.Logf("Error searching profiles: %v", err)
					existsChannel[i] = false
					return
				}
				if len(cpProfile) != 1 {
					t.Logf("Expected 1 profile, got %v", len(cpProfile))
					existsChannel[i] = false
					return
				}
				if cpProfile[0].FirstName != firstName {
					t.Logf("Expected first name %v, got %v", firstName, cpProfile[0].FirstName)
					existsChannel[i] = false
					return
				}
				if profileOne.FirstName != firstName {
					t.Logf("Expected first name %v, got %v", firstName, profileOne.FirstName)
					existsChannel[i] = false
					return
				}
				existsChannel[i] = true
			}(i)
		}
		wg.Wait()

		for _, exists := range existsChannel {
			if !exists {
				trueCount := 0
				for _, value := range existsChannel {
					if value {
						trueCount++
					}
				}
				percentageTrue := (float64(trueCount) / float64(len(existsChannel))) * 100
				t.Logf("Percentage of true values: %.2f%%\n", percentageTrue)
				return false
			}
		}
		return true
	}, 12, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}

}
