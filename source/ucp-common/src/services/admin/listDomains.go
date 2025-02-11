// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/admin"
)

func SearchDomain(tx core.Transaction, accp customerprofiles.ICustomerProfileLowCostConfig, envName string) ([]model.Domain, error) {
	domains := []model.Domain{}
	profileDomains, err := accp.ListDomains()
	if err != nil {
		return domains, err
	}
	for _, dom := range profileDomains {
		if envName == dom.Tags[constant.DOMAIN_TAG_ENV_NAME] {
			domains = append(domains, model.Domain{
				Name:        dom.Name,
				Created:     dom.Created,
				LastUpdated: dom.LastUpdated,
			})
		}
	}
	return domains, nil
}
