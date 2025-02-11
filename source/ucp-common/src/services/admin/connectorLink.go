// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"strings"
	"tah/upt/source/tah-core/db"
	constants "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/admin"
	utils "tah/upt/source/ucp-common/src/utils/utils"
)

func GetLinkedDomains(dbClient db.DBConfig) (model.IndustryConnectorDomainList, error) {
	domainList := model.IndustryConnectorDomainList{
		ItemID:   constants.CONFIG_DB_LINKED_CONNECTORS_PK,
		ItemType: constants.CONFIG_DB_LINKED_CONNECTORS_SK,
	}
	err := dbClient.Get(constants.CONFIG_DB_LINKED_CONNECTORS_PK, constants.CONFIG_DB_LINKED_CONNECTORS_SK, &domainList)
	notFoundPrefix := "ResourceNotFoundException"
	if err != nil && strings.HasPrefix(err.Error(), notFoundPrefix) {
		return domainList, nil
	}
	return domainList, err
}

func AddLinkedDomain(dbClient db.DBConfig, domain string) error {
	domainList, err := GetLinkedDomains(dbClient)
	if err != nil {
		return err
	}
	if utils.ContainsString(domainList.DomainList, domain) {
		return nil
	}

	domainList.DomainList = append(domainList.DomainList, domain)
	_, err = dbClient.Save(domainList)
	return err
}

func RemoveLinkedDomain(dbClient db.DBConfig, domain string) error {
	domainList, err := GetLinkedDomains(dbClient)
	if err != nil {
		return err
	}

	updatedDomains := []string{}
	for _, d := range domainList.DomainList {
		if d != domain {
			updatedDomains = append(updatedDomains, d)
		}
	}
	domainList.DomainList = updatedDomains

	_, err = dbClient.Save(domainList)
	return err
}
