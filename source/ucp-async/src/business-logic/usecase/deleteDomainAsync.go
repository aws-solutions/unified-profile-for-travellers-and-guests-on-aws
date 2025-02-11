// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package usecase

import (
	"encoding/json"
	"errors"
	"sync"
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/glue"
	common "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	commonServices "tah/upt/source/ucp-common/src/services/admin"
)

type DeleteDomainUsecase struct {
	name string
}

func InitDeleteDomain() *DeleteDomainUsecase {
	return &DeleteDomainUsecase{
		name: "DeleteDomain",
	}
}

func (u *DeleteDomainUsecase) Name() string {
	return u.name
}

func (u *DeleteDomainUsecase) Execute(payload model.AsyncInvokePayload, services model.Services, tx core.Transaction) error {
	payloadJson, err := json.Marshal(payload.Body)
	if err != nil {
		tx.Error("[%s] Error marshalling payload %v", u.name, err)
		return err
	}

	data := uc.DeleteDomainBody{}
	err = json.Unmarshal(payloadJson, &data)
	if err != nil {
		tx.Error("[%s] Error unmarshalling payload body %v", u.name, err)
		return err
	}
	tx.Debug("[%s] Payload body: %+v", u.name, data)
	return u.deleteDomain(services, data, tx)
}

/*
* Services used:
 1. GlueConfig
 2. ConfigDB
 3. PrivacyResultsDB
 4. CognitoConfig
 5. AccpConfig
*/
func (u *DeleteDomainUsecase) deleteDomain(services model.Services, data uc.DeleteDomainBody, tx core.Transaction) error {
	var wg sync.WaitGroup

	tx.Info("[%s] Deleting Glue tables for domain %s", u.name, data.DomainName)
	var deleteGlueTablesError error
	wg.Add(1)
	go func() {
		defer wg.Done()
		deleteGlueTablesError = deleteGlueTables(data, services.GlueConfig, tx)
	}()

	// Stop data transfer from Industry Connector
	var removeLinkedDomainError error
	wg.Add(1)
	go func() {
		defer wg.Done()
		tx.Info("[%s] Stopping data transfer from Industry Connector", u.name)
		removeLinkedDomainError = removeLinkedDomain(data, services.ConfigDB, tx)
	}()

	var removeGluePartitionsErrors error
	wg.Add(1)
	go func() {
		defer wg.Done()
		tx.Info("[%s] Removing partitions from glue table for domain %s", u.name, data.DomainName)
		removeGluePartitionsErrors = errors.Join(removeGluePartitions(data, services.GlueConfig, services.Env["GLUE_EXPORT_TABLE_NAME"], tx)...)
	}()

	// Delete entries from Privacy Search Table
	var privacySearchError error
	wg.Add(1)
	go func() {
		defer wg.Done()
		tx.Info("[%s] Deleting entries from Privacy Search Table", u.name)
		searchResults := []uc.PrivacySearchResult{}
		privacySearchError = services.PrivacyResultsDB.FindByPk(data.DomainName, &searchResults)
		if privacySearchError != nil {
			tx.Error("[%s] Error querying privacy searches: %s", u.name, privacySearchError.Error())
			return
		}

		deleteRequest := []map[string]string{}
		for _, searchResult := range searchResults {
			deleteRequest = append(deleteRequest, map[string]string{
				services.PrivacyResultsDB.PrimaryKey: data.DomainName,
				services.PrivacyResultsDB.SortKey:    searchResult.ConnectId,
			})
		}
		privacySearchError = services.PrivacyResultsDB.DeleteMany(deleteRequest)
		if privacySearchError != nil {
			tx.Error("[%s] Error deleting privacy searches: %s", u.name, privacySearchError.Error())
		}
	}()

	tx.Info("[%s] Deleting cognito groups for %s", u.name, data.DomainName)
	var deleteGroupError error
	wg.Add(1)
	go func() {
		defer wg.Done()
		deleteGroupError = deleteCognitoGroups(services.CognitoConfig, tx, u.name, data.DomainName)
	}()

	wg.Wait()

	tx.Info("[%s] Deleting domain %s", u.name, data.DomainName)
	// We want to log if Cognito Groups fail to delete. Depending on a customer's deployment configuration, groups may have never been created.
	if deleteGroupError != nil {
		if isPermissionSystemEnabled(services.Env) {
			tx.Error("[%s] Error deleting cognito groups: %s", u.name, deleteGroupError.Error())
		} else {
			tx.Info("[%s] UPT is deployed with Cognito user group permission system disabled. No Cognito user groups were deleted.", u.name)
		}
	}

	services.AccpConfig.SetDomain(data.DomainName)
	deleteDomainError := services.AccpConfig.DeleteDomain()

	// we want to return errors only at the end to let the delete process go all the way through
	return errors.Join(deleteGlueTablesError, deleteDomainError, privacySearchError, removeGluePartitionsErrors, removeLinkedDomainError)
}

func deleteGlueTables(data uc.DeleteDomainBody, glueConfig glue.IConfig, tx core.Transaction) error {
	var lastErr error
	for _, bizObject := range common.BUSINESS_OBJECTS {
		tableName := commonServices.BuildTableName(data.Env, bizObject, data.DomainName)
		err := glueConfig.DeleteTable(tableName)
		if err != nil {
			tx.Error("[DeleteDomain] Error deleting table %s: %v", tableName, err)
			lastErr = err
		}
	}
	return lastErr
}

func removeLinkedDomain(data uc.DeleteDomainBody, configDB db.DBConfig, tx core.Transaction) error {
	err := commonServices.RemoveLinkedDomain(configDB, data.DomainName)
	if err != nil {
		tx.Error("[DeleteDomain] Error removing link to Industry Connector: %v", err)
	}
	return err
}

func removeGluePartitions(data uc.DeleteDomainBody, glueConfig glue.IConfig, glueTableName string, tx core.Transaction) []error {
	errs := glueConfig.RemovePartitionsFromTable(glueTableName, []glue.Partition{{Values: []string{data.DomainName}}})
	if len(errs) > 0 {
		tx.Error("[DeleteDomain] Error removing partitions to glue table %+v", errs)
	}
	return errs
}

func deleteCognitoGroups(cognitoConfig cognito.ICognitoConfig, tx core.Transaction, ucName string, domainName string) error {
	var deleteGroupError error
	appAccessGroupPrefix := buildAppAccessGroupPrefix(domainName)
	for roleName, rolePermission := range appAccessRoleMap {
		groupName := buildAppAccessGroupName(appAccessGroupPrefix, roleName, rolePermission)
		deleteGroupError = errors.Join(deleteGroupError, cognitoConfig.DeleteGroup(groupName))
	}
	deleteGroupError = errors.Join(deleteGroupError, cognitoConfig.DeleteGroup(buildDataAccessGroupName(domainName)))
	if deleteGroupError != nil {
		tx.Error("[%s] Error deleting cognito groups: %s", ucName, deleteGroupError.Error())
	}
	return deleteGroupError
}
