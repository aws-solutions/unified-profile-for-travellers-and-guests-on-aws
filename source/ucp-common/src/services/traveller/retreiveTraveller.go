// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/adapter"
	model "tah/upt/source/ucp-common/src/model/traveler"
)

func RetreiveTraveller(tx core.Transaction, accp customerprofiles.ICustomerProfileLowCostConfig, id string, pagination []customerprofiles.PaginationOptions) (model.Traveller, error) {
	tx.Debug("Fetching mappings from LCS")
	mappings, err := accp.GetMappings()
	if err != nil {
		return model.Traveller{}, err
	}
	var objectTypeNames []string
	for _, mapping := range mappings {
		objectTypeNames = append(objectTypeNames, mapping.Name)
	}

	tx.Info("Retreiving traveller with id %s and pagination %+v", id, pagination)
	profile, err := accp.GetProfile(id, objectTypeNames, pagination)
	if err != nil {
		tx.Error("Error retrieving traveller with id %s: %v", id, err)
		return model.Traveller{}, err
	}
	return adapter.ProfileToTraveller(tx, profile), nil
}
