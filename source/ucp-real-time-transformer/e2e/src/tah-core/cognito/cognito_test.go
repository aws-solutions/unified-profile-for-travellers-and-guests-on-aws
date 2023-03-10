package cognito

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

/**************
* COGNITO TESTING
****************/
func TestCognito(t *testing.T) {
	userPoolName := "CloudrackTestUserPool"
	groupName := "CloudrackTestGroup"
	testUser := "test@gmail.com"
	c := Init("")
	err := c.CreateUserPool(userPoolName)
	if err != nil {
		t.Errorf("[core][Cognito] Error while creating user pool" + err.Error())
	}
	err = c.CreateGroup(groupName, "a test group for automared testing")
	if err != nil {
		t.Errorf("[core][Cognito] Error while creating group" + err.Error())
	}
	err = c.CreateUser(testUser, groupName)
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
}
