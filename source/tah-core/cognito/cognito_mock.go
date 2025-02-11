// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package cognito

import (
	"github.com/stretchr/testify/mock"
)

type MockCognitoConfig struct {
	Token  *string
	Groups *[]Group
	mock.Mock
}

func InitMock(token *string, groups *[]Group) *MockCognitoConfig {
	return &MockCognitoConfig{
		Token:  token,
		Groups: groups,
		Mock:   mock.Mock{},
	}
}

func (c *MockCognitoConfig) SetTokenUrl(url string) {
	c.Called(url)
}

func (c *MockCognitoConfig) GetAccessTokenFromClientCredentials(clientID string, clientSecret string, scope string) (string, error) {
	c.Called(clientID, clientSecret, scope)
	return *c.Token, nil
}

func (c *MockCognitoConfig) CreateGroup(name string, description string) error {
	c.Called(name, description)
	if c.Groups == nil {
		c.Groups = &[]Group{{Name: name, Description: description}}
	} else {
		*c.Groups = append(*c.Groups, Group{Name: name, Description: description})
	}

	return nil
}

func (c *MockCognitoConfig) ListGroupsForUser(username string) ([]Group, error) {
	c.Called(username)
	return *c.Groups, nil
}

func (c *MockCognitoConfig) IsGroupExistsException(err error) bool {
	c.Called(err)
	return false
}

func (c *MockCognitoConfig) DeleteGroup(name string) error {
	c.Called(name)
	groupsNew := []Group{}
	for _, grp := range *c.Groups {
		if grp.Name != name {
			groupsNew = append(groupsNew, grp)
		}
	}
	*c.Groups = groupsNew
	return nil
}
