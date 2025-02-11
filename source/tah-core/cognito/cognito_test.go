// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cognito

import (
	"testing"

	"tah/upt/source/tah-core/core"

	"github.com/aws/aws-lambda-go/events"
)

/**************
* COGNITO TESTING
****************/
func TestCognito(t *testing.T) {
	userPoolName := "CloudrackTestUserPool"
	groupName := "CloudrackTestGroup"
	groupDescription := "a test group for automared testing"
	testUser := "test@example.com"
	c := Init("", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION, core.LogLevelDebug)
	err := c.CreateUserPool(userPoolName)
	if err != nil {
		t.Errorf("[core][Cognito] Error while creating user pool" + err.Error())
	}
	err = c.createTestClient("cognito-test-client")
	if err != nil {
		t.Errorf("[core][Cognito] Error while creating test client" + err.Error())
	}
	err = c.CreateGroup(groupName, groupDescription)
	if err != nil {
		t.Errorf("[core][Cognito] Error while creating group" + err.Error())
	}
	err = c.CreateUser(testUser, groupName, true)
	if err != nil {
		t.Errorf("[core][Cognito] Error while creating user" + err.Error())
	}
	if !c.UserExists(testUser) {
		t.Errorf("[core][Cognito] UserExists functiono should return TRUE")
	}
	user, err3 := c.GetUser(testUser)
	if err3 != nil {
		t.Errorf("[core][Cognito] Error while retreiving user" + err3.Error())
	}
	if user.Email != testUser {
		t.Errorf("[core][Cognito] User retreived should have email %s and not %s", testUser, user.Email)
	}
	password := core.RandomPassword(40)
	err = c.SetUserPassword(testUser, password, true)
	if err != nil {
		t.Errorf("[core][Cognito] Error while setting user password" + err.Error())
	}
	refreshToken, accessToken, err := c.AuthenticateWithUsernamePassword(testUser, password)
	if err != nil {
		t.Errorf("[core][Cognito] Error while getting access token from username and password" + err.Error())
	}
	if accessToken == "" {
		t.Errorf("[core][Cognito] Access token should not be empty")
	}
	if refreshToken == "" {
		t.Errorf("[core][Cognito] Refresh token should not be empty")
	}
	err = c.ReserUserPassword(testUser)
	if err != nil {
		t.Errorf("[core][Cognito] Error while reseting user password" + err.Error())
	}
	users, err2 := c.UsersInGroup(groupName)
	if err2 != nil {
		t.Errorf("[core][Cognito] Error while searching users in group (after user create)" + err2.Error())
	}
	if len(users) != 1 {
		t.Errorf("[core][Cognito]There should be 1 user in group and not %d", len(users))
	}
	if users[0].Email != testUser {
		t.Errorf("[core][Cognito] user email should be %s and not %v", testUser, users[0].Email)
	}
	group, err := c.GetGroup(groupName)
	if err != nil {
		t.Errorf("[core][Cognito] Error while getting group: %v", err)
	}
	if group.Name != groupName || group.Description != groupDescription {
		t.Errorf("[core][Cognito] Group does not match expected result")
	}
	groups, err := c.ListGroupsForUser(user.Username)
	if err != nil {
		t.Errorf("[core][Cognito] Error while getting groups for user: %v", err)
	}
	if len(groups) != 1 || groups[0].Name != groupName {
		t.Errorf("[core][Cognito] User should be in one group: %v", groupName)
	}
	err = c.DeleteUser(testUser)
	if err != nil {
		t.Errorf("[core][Cognito] Error while deleting user" + err.Error())
	}
	if c.UserExists(testUser) {
		t.Errorf("[core][Cognito] UserExists functiono should return FALSE")
	}
	users, err = c.UsersInGroup(groupName)
	if err != nil {
		t.Errorf("[core][Cognito] Error while searching users in group (after user delete)" + err.Error())
	}
	if len(users) != 0 {
		t.Errorf("[core][Cognito]There should be 0 user in group and not %d", len(users))
	}
	err = c.DeleteGroup(groupName)
	if err != nil {
		t.Errorf("[core][Cognito] Error while deleting group" + err.Error())
	}
	err = c.DeleteUserPool()
	if err != nil {
		t.Errorf("[core][Cognito] Error while deleting user Pool" + err.Error())
	}
}

func TestCognitoGroupParser(t *testing.T) {
	//authorizer := map[string]string{
	//	"cognito:groups": "[group1, group2]",
	//}
	//mapToParse := authorizer.(map[string]interface{})
	req := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"claims": map[string]interface{}{
					"cognito:groups": "[group1 group2]",
				},
			},
		},
	}
	//req.RequestContext.Authorizer["cognito:groups"] = "[group1, group2]"
	groups := ParseGroupsFromLambdaRq(req)
	if len(groups) != 2 {
		t.Errorf("[core][Cognito] Group length should be 2 and not %v", len(groups))
	}
	if groups[0] != "group1" || groups[1] != "group2" {
		t.Errorf("[core][Cognito] Group should be [group1,group2] and not [%s,%s]", groups[0], groups[1])
	}

	tests := []string{"[group1, group2]", "[group1 , group2]", "[group1,group2]"}
	for _, testString := range tests {
		groups = ParseGroupsFromString(testString, ",")
		if len(groups) != 2 {
			t.Errorf("[core][Cognito] Group length should be 2 and not %v", len(groups))
		}
		groups = ParseGroupsFromString(testString, ",")
		if groups[0] != "group1" || groups[1] != "group2" {
			t.Errorf("[core][Cognito] Group should be [group1,group2] and not [%s,%s]", groups[0], groups[1])
		}
	}
	groups = ParseGroupsFromString("[group1 ]", ",")
	if groups[0] != "group1" {
		t.Errorf("[core][Cognito] Group length should be 2 and not %v", len(groups))
	}

}

func TestCognitoUsernameParser(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"claims": map[string]interface{}{
					"username": "grollat",
				},
			},
		},
	}
	username := ParseUserFromLambdaRq(req)
	if username != "grollat" {
		t.Errorf("[core][Cognito] username shouuld be grollat and noot %v", username)
	}

	req2 := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"claims": map[string]interface{}{
					"cognito:username": "grollat",
				},
			},
		},
	}
	username = ParseUserFromLambdaRq(req2)
	if username != "grollat" {
		t.Errorf("[core][Cognito] username shouuld be grollat and noot %v", username)
	}
}

func TestCognitoMocks(t *testing.T) {
	testUser := "test@example.com"
	groupName := "MockTestGroup"
	groupDescription := "a test group for mock testing"

	// Set up real resources
	config := Init("", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION, core.LogLevelDebug)
	err := config.CreateUserPool("testUserPool")
	if err != nil {
		t.Errorf("[core][Cognito] Error while creating user pool %v", err)
	}
	err = config.CreateGroup(groupName, groupDescription)
	if err != nil {
		t.Errorf("[core][Cognito] Error while creating group %v", err)
	}
	err = config.CreateUser(testUser, groupName, true)
	if err != nil {
		t.Errorf("[core][Cognito] Error while creating user %v", err)
	}
	user, err := config.GetUser(testUser)
	if err != nil {
		t.Errorf("[core][Cognito] Error while retreiving user %v", err)
	}
	// Set up mock
	mockGroups := []Group{
		{Name: groupName, Description: groupDescription},
	}
	mockConfig := InitMock(nil, &mockGroups)

	// Test real resources
	groups, err := injectedFunc(user.Username, config)
	if err != nil {
		t.Errorf("[core][Cognito] Error while getting groups for user: %v", err)
	}
	if len(groups) != 1 || groups[0].Name != groupName {
		t.Errorf("[core][Cognito] User should be in one group: %v", groupName)
	}
	// Test mock
	mockConfig.On("ListGroupsForUser", testUser)
	groups, err = injectedFunc(user.Username, mockConfig)
	if err != nil {
		t.Errorf("[core][Cognito] Error while getting mock groups for user: %v", err)
	}
	if len(groups) != 1 || groups[0].Name != groupName {
		t.Errorf("[core][Cognito] User should be in one mock group: %v", groupName)
	}

	// Clean up
	err = config.DeleteUser(testUser)
	if err != nil {
		t.Errorf("[core][Cognito] Error while deleting user %v", err)
	}
	if config.UserExists(testUser) {
		t.Errorf("[core][Cognito] UserExists functiono should return FALSE")
	}
	err = config.DeleteGroup(groupName)
	if err != nil {
		t.Errorf("[core][Cognito] Error while deleting group %v", err)
	}
	err = config.DeleteUserPool()
	if err != nil {
		t.Errorf("[core][Cognito] Error while deleting user Pool %v", err)
	}
}

func injectedFunc(username string, config ICognitoConfig) ([]Group, error) {
	groups, err := config.ListGroupsForUser(username)
	return groups, err
}
