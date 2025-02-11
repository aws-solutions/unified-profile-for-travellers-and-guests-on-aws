// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package traveller

import (
	"errors"
	"sync"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-common/src/model/admin"
	"time"
)

func MergeProfiles(tx core.Transaction, accp customerprofiles.ICustomerProfileLowCostConfig, domain string, rq []model.MergeRq) error {
	tx.Debug("Setting domain to ACCP config to %s", domain)
	accp.SetDomain(domain)
	tx.Info("Merging %v profiles for domain %s", len(rq), domain)
	profilePairs := []customerprofiles.ProfilePair{}
	for _, r := range rq {
		profilePairs = append(profilePairs, customerprofiles.ProfilePair{
			SourceID:     r.SourceProfileID,
			TargetID:     r.TargetProfileID,
			MergeContext: r.MergeContext,
		})
	}
	_, err := accp.MergeMany(profilePairs)
	return err
}

func WaitForProfileMergeCompletion(
	tx core.Transaction,
	accp customerprofiles.ICustomerProfileLowCostConfig,
	domain string,
	rq []model.MergeRq,
	timeoutSeconds int,
) (success []model.MergeRq, failure []model.MergeRq, err error) {
	tx.Debug("Running WaitForProfileMergeCompletion with timeout %v (%v iterations)", timeoutSeconds, timeoutSeconds/5)
	tx.Debug("Setting domain to ACCP config to %s", domain)
	accp.SetDomain(domain)
	tx.Debug("Waiting for profile to be deleted")
	duration := 5 * time.Second
	max := timeoutSeconds / 5

	var wg sync.WaitGroup
	mu := &sync.Mutex{}

	for index, mergeRq := range rq {
		wg.Add(1)
		go func(j int, mergRqToCheck model.MergeRq) {
			defer wg.Done()

			for i := 0; i < max; i++ {
				accpProfileId, _ := accp.GetProfileId("_profileId", mergRqToCheck.SourceProfileID)
				accpProfileId2, _ := accp.GetProfileId("_profileId", mergRqToCheck.TargetProfileID)

				if accpProfileId == accpProfileId2 {
					tx.Info("[mergeRq-%v] Target and source proifles return same connectID %s: Merge successfull", j, accpProfileId)
					mu.Lock()
					defer mu.Unlock()
					success = append(success, mergRqToCheck)
					return
				}
				tx.Debug(
					"[mergeRq-%v] Both profiles %s and %s have different connect IDs. Merge not yet completed. waiting %v",
					j,
					accpProfileId,
					accpProfileId2,
					duration,
				)
				time.Sleep(duration)
			}
			tx.Error("[mergeRq-%v] Timeout while waiting for profile merge. Merged failed", j)
			mu.Lock()
			defer mu.Unlock()
			failure = append(failure, mergRqToCheck)
		}(index, mergeRq)
	}
	wg.Wait()
	if len(failure) > 0 {
		err = errors.New("some profile pairs failed to merge")
	}
	return success, failure, err
}
