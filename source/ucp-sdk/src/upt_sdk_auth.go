// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package upt_sdk

import (
	"fmt"
	"log"
	"tah/upt/source/tah-core/core"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

func (h UptHandler) CreateDomainGroupAndUser(domainName string) (string, string, error) {
	adminGroupName := cognitoGroupNameFromDomain(domainName)
	adminGroupDescription := "*/*"
	adminUserName := adminUserEmailFromDomain(domainName)
	adminPassword := core.RandomPassword(40)
	err := h.CognitoClient.CreateGroup(adminGroupName, adminGroupDescription)
	if err != nil {
		log.Printf("Error creating admin group: %s", err)
		return "", "", err
	}
	err = h.CognitoClient.CreateUser(adminUserName, adminGroupName, true)
	if err != nil {
		log.Printf("Error creating admin user: %s", err)
		return "", "", err
	}
	err = h.CognitoClient.SetUserPassword(adminUserName, adminPassword, true)
	if err != nil {
		log.Printf("Error creating admin user: %s", err)
		return "", "", err
	}
	appAccessGroup := "app-global-custom" + domainName + fmt.Sprintf("/%x", constants.AdminPermission)
	err = h.CognitoClient.CreateGroup(appAccessGroup, "")
	if err != nil {
		log.Println("Error creating app access group")
		return "", "", err
	}
	err = h.CognitoClient.AddUserToGroup(adminUserName, appAccessGroup)
	if err != nil {
		log.Println("Error adding user to group")
		return "", "", err
	}
	return adminUserName, adminPassword, nil
}

func (h UptHandler) Login(adminUserName, adminPassword string) (string, string, error) {
	adminRefreshToken, adminAccessToken, err := h.CognitoClient.AuthenticateWithUsernamePassword(adminUserName, adminPassword)
	if err != nil {
		log.Printf("Error authenticating admin user: %s", err)
		return "", "", err
	}
	return adminRefreshToken, adminAccessToken, nil
}

func (h *UptHandler) SetAuthTokens(adminRefreshToken, adminAccessToken string) {
	h.RefreshToken = adminRefreshToken
	h.AccessToken = adminAccessToken
}
func (h *UptHandler) UnsetAuthTokens() {
	h.RefreshToken = ""
	h.AccessToken = ""
}

func (h UptHandler) DeleteGroupAndUser(domainName string) error {
	adminGroupName := cognitoGroupNameFromDomain(domainName)
	adminUserName := adminUserEmailFromDomain(domainName)
	if _, err := h.CognitoClient.GetGroup(adminGroupName); err == nil {
		err := h.CognitoClient.DeleteGroup(adminGroupName)
		if err != nil {
			log.Printf("Error deleting admin group: %s", err)
		}
	}
	err := h.CognitoClient.DeleteUser(adminUserName)
	if err != nil {
		log.Printf("Error deleting admin user: %s", err)
	}
	return nil
}

func (h UptHandler) createAuthHeader() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + h.AccessToken,
	}
}

func cognitoGroupNameFromDomain(domainName string) string {
	return "ucp-" + domainName + "-admin"
}
func adminUserEmailFromDomain(domainName string) string {
	return domainName + "-admin-@example.com"
}
