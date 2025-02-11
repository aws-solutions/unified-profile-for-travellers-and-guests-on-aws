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

func UnmergeProfiles(tx core.Transaction, accp customerprofiles.ICustomerProfileLowCostConfig, domain string, rq model.UnmergeRq) error {
	tx.Debug("Setting domain to ACCP config to %s", domain)
	accp.SetDomain(domain)

	_, err := accp.UnmergeProfiles(rq.ToUnmergeConnectID, rq.MergedIntoConnectID, rq.InteractionToUnmerge, rq.InteractionType)
	return err
}

// After unmerge is performed, this function periodically checks aurora and dynamo to ensure unmerge changes have been propagated to both aurora and dynamo
func ValidateProfileUnmergeCompletion(tx core.Transaction, accp customerprofiles.ICustomerProfileLowCostConfig, domain string, rq model.UnmergeRq, timeoutSeconds int) error {
	tx.Debug("[ValidateProfileUnmergeCompletion] running with timeout %v (%v iterations)", timeoutSeconds, timeoutSeconds/5)

	tx.Debug("Setting domain to ACCP config to %s", domain)
	accp.SetDomain(domain)

	duration := 5 * time.Second
	max := timeoutSeconds / 5
	success := false

	var wg sync.WaitGroup
	wg.Add(1)

	go func(unmergeRqToCheck model.UnmergeRq) {
		for i := 0; i < max; i++ {

			// check aurora master table
			rdsCheck := validateAuroraProfiles(tx, accp, unmergeRqToCheck, domain)
			// check dynamodb for profile creation
			dynamoCheck := validateDynamoProfiles(accp, unmergeRqToCheck)
			if rdsCheck && dynamoCheck {
				tx.Info("[ValidateProfileUnmergeCompletion] Unmerge successful, profile %s unmerged from %s", unmergeRqToCheck.ToUnmergeConnectID, unmergeRqToCheck.MergedIntoConnectID)
				success = true
				wg.Done()
				return
			}
			tx.Debug("[ValidateProfileUnmergeCompletion] Both profiles %s and %s have same connect IDs. Unmerge not yet completed. waiting %v", unmergeRqToCheck.ToUnmergeConnectID, unmergeRqToCheck.MergedIntoConnectID, duration)
			time.Sleep(duration)
		}
		tx.Info("[ValidateProfileUnmergeCompletion] Timeout while waiting for profile unmerge")
		wg.Done()
	}(rq)

	wg.Wait()
	if !success {
		return errors.New("[ValidateProfileUnmergeCompletion] failed to unmerge profiles")
	}
	return nil
}

func validateDynamoProfiles(accp customerprofiles.ICustomerProfileLowCostConfig, unmergeRqToCheck model.UnmergeRq) bool {
	// GetProfile api returns an error if
	// 1. multiple profiles are found with the same connectId, or
	// 2. no profile is found with the specifed connectId

	// If unmerge is unsuccessful, both queries below will return an error.
	// If no error returned, that implies unmerge was successful
	_, err := accp.GetProfile(unmergeRqToCheck.MergedIntoConnectID, []string{}, []customerprofiles.PaginationOptions{})
	if err != nil {
		return false
	}

	_, err = accp.GetProfile(unmergeRqToCheck.ToUnmergeConnectID, []string{}, []customerprofiles.PaginationOptions{})
	return err == nil
}

// Unmerge operation reverts the connectId for a profile to the same as its original_connectId. This function queries rds to confirm the same
func validateAuroraProfiles(tx core.Transaction, accp customerprofiles.ICustomerProfileLowCostConfig, unmergeRqToCheck model.UnmergeRq, domain string) bool {
	parentOfToUnmerge, err := accp.FindCurrentParentOfProfile(domain, unmergeRqToCheck.ToUnmergeConnectID)
	if err != nil || parentOfToUnmerge != unmergeRqToCheck.ToUnmergeConnectID {
		tx.Error("[validateAuroraProfiles] Profile was not unmerged. Connect Id %s does not match its original connect ID %s ", parentOfToUnmerge, unmergeRqToCheck.ToUnmergeConnectID)
		return false
	}

	parentOfMergedInto, err := accp.FindCurrentParentOfProfile(domain, unmergeRqToCheck.MergedIntoConnectID)
	if err != nil || parentOfMergedInto != unmergeRqToCheck.MergedIntoConnectID {
		tx.Error("[validateAuroraProfiles] Connect Id %s does not match its original connect ID %s", parentOfMergedInto, unmergeRqToCheck.MergedIntoConnectID)
		return false
	}

	return true
}
