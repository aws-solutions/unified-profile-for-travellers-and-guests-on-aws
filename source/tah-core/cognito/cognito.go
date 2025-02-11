// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cognito

import (
	"errors"
	"log"
	"strings"
	core "tah/upt/source/tah-core/core"
	"time"

	b64 "encoding/base64"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cognitoSdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

// Interfaces to abstract the cognito.CognitoConfig from tah/core object allowing mock-testing
type ICognitoConfig interface {
	GetAccessTokenFromClientCredentials(clientID string, clientSecret string, scope string) (string, error)
	SetTokenUrl(url string)
	CreateGroup(name string, description string) error
	ListGroupsForUser(username string) ([]Group, error)
	IsGroupExistsException(err error) bool
	DeleteGroup(name string) error
}

type CognitoConfig struct {
	Client     *cognitoSdk.CognitoIdentityProvider
	UserPoolID string
	ClientID   string
	TokenUrl   string //cognitoo token exchange URL
	Tx         core.Transaction
}

type User struct {
	Username string
	Email    string
	Status   string
}

type Group struct {
	Name        string
	Description string
}

type CognitoClientCredentialRq struct {
	GrantType string `json:"grant_type"`
	ClientID  string `json:"client_id"`
	Scope     string `json:"scope"`
}

type CognitoClientCredentialRs struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (p CognitoClientCredentialRs) Decode(dec json.Decoder) (error, core.JSONObject) {
	return dec.Decode(&p), p
}

func Init(userPoolID string, solutionId, solutionVersion string, logLevel core.LogLevel) *CognitoConfig {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := core.CreateClient(solutionId, solutionVersion)
	// Create DynamoDB client
	return &CognitoConfig{
		Client:     cognitoSdk.New(session, &aws.Config{HTTPClient: client}),
		UserPoolID: userPoolID,
		Tx:         core.NewTransaction("Cognito", "", logLevel),
	}

}
func InitWithClientID(userPoolID string, clientID string, solutionId, solutionVersion string, logLevel core.LogLevel) *CognitoConfig {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := core.CreateClient(solutionId, solutionVersion)
	return &CognitoConfig{
		Client:     cognitoSdk.New(session, &aws.Config{HTTPClient: client}),
		UserPoolID: userPoolID,
		ClientID:   clientID,
		Tx:         core.NewTransaction("Cognito", "", logLevel),
	}
}

func InitWithTokenUrl(userPoolID string, url string, solutionId, solutionVersion string, logLevel core.LogLevel) *CognitoConfig {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := core.CreateClient(solutionId, solutionVersion)
	return &CognitoConfig{
		Client:     cognitoSdk.New(session, &aws.Config{HTTPClient: client}),
		UserPoolID: userPoolID,
		TokenUrl:   url,
		Tx:         core.NewTransaction("Cognito", "", logLevel),
	}
}

func (c *CognitoConfig) SetTokenUrl(url string) {
	c.TokenUrl = url
}

func (c CognitoConfig) RefreshToken(refreshToken string) (string, error) {
	c.Tx.Info("[core][cognito] Refresh Access Token ")
	input := &cognitoSdk.AdminInitiateAuthInput{
		UserPoolId: aws.String(c.UserPoolID),
		AuthFlow:   aws.String("REFRESH_TOKEN_AUTH"),
		ClientId:   aws.String(c.ClientID),
		AuthParameters: map[string]*string{
			"REFRESH_TOKEN": aws.String(refreshToken),
		},
	}
	out, err := c.Client.AdminInitiateAuth(input)
	if out.AuthenticationResult != nil && out.AuthenticationResult.AccessToken != nil {
		return *out.AuthenticationResult.AccessToken, err
	}
	return "", err
}

func (c CognitoConfig) GetAccessTokenFromClientCredentials(clientID string, clientSecret string, scope string) (string, error) {
	if c.TokenUrl == "" {
		return "", errors.New("cognito client must be initialized with token enpoint")
	}
	c.Tx.Info("[core][cognito] perform authentication using cognito endpoint")
	httpSvc := core.HttpInit(c.TokenUrl+"?grant_type=client_credentials", c.Tx.LogLevel)
	rq := CognitoClientCredentialRq{
		GrantType: "client_credentials",
		ClientID:  clientID,
		Scope:     scope,
	}
	options := core.RestOptions{
		Headers: map[string]string{
			"Authorization": "Basic " + b64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)),
			"Content-Type":  "application/x-www-form-urlencoded",
		},
	}
	rs := CognitoClientCredentialRs{}
	res, err := httpSvc.HttpPost(rq, rs, options)
	rs = res.(CognitoClientCredentialRs)
	if err != nil {
		return "", err
	}
	return rs.AccessToken, err
}

func (c CognitoConfig) AuthenticateWithUsernamePassword(username string, password string) (string, string, error) {
	c.Tx.Info("[core][cognito] Authenticating with username and password")
	input := &cognitoSdk.AdminInitiateAuthInput{
		AuthFlow: aws.String(cognitoSdk.AuthFlowTypeAdminUserPasswordAuth),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(username),
			"PASSWORD": aws.String(password),
		},
		ClientId:   aws.String(c.ClientID),
		UserPoolId: aws.String(c.UserPoolID),
	}
	output, err := c.Client.AdminInitiateAuth(input)
	if err != nil {
		return "", "", err
	}
	return aws.StringValue(output.AuthenticationResult.RefreshToken), aws.StringValue(output.AuthenticationResult.AccessToken), nil
}

func (c CognitoConfig) ClientSecret(clientID string) (string, error) {
	c.Tx.Info("[core][cognito] Refresh Access Token ")
	input := &cognitoSdk.DescribeUserPoolClientInput{
		UserPoolId: aws.String(c.UserPoolID),
		ClientId:   aws.String(clientID),
	}
	out, err := c.Client.DescribeUserPoolClient(input)
	if out.UserPoolClient != nil && out.UserPoolClient.ClientSecret != nil {
		return *out.UserPoolClient.ClientSecret, err
	}
	return "", err
}

func (c *CognitoConfig) CreateUserPool(name string) error {
	c.Tx.Info("[core][cognito] Creating user pool %v", name)
	input := &cognitoSdk.CreateUserPoolInput{
		PoolName: aws.String(name),
	}
	out, err := c.Client.CreateUserPool(input)
	c.UserPoolID = *out.UserPool.Id
	return err
}
func (c *CognitoConfig) DeleteUserPool() error {
	c.Tx.Info("[core][cognito] Deleting user pool %v", c.UserPoolID)
	input := &cognitoSdk.DeleteUserPoolInput{
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.DeleteUserPool(input)
	c.UserPoolID = ""
	return err
}

func (c CognitoConfig) CreateGroup(name string, description string) error {
	c.Tx.Info("[core][cognito] Creating Group %v", name)
	input := &cognitoSdk.CreateGroupInput{
		GroupName:   aws.String(name),
		Description: aws.String(description),
		UserPoolId:  aws.String(c.UserPoolID),
	}
	_, err := c.Client.CreateGroup(input)
	return err
}
func (c CognitoConfig) DeleteGroup(name string) error {
	c.Tx.Info("[core][cognito] Deleting Group %v", name)
	input := &cognitoSdk.DeleteGroupInput{
		GroupName:  aws.String(name),
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.DeleteGroup(input)
	return err
}
func (c CognitoConfig) CreateUser(email string, groupName string, suppressEmail bool) error {
	c.Tx.Info("[core][cognito] Creating User %v in group %v", email, groupName)
	input := &cognitoSdk.AdminCreateUserInput{
		Username: aws.String(email),
		DesiredDeliveryMediums: []*string{
			aws.String("EMAIL"),
		},
		UserAttributes: []*cognitoSdk.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
		},
		UserPoolId: aws.String(c.UserPoolID),
	}
	if suppressEmail {
		input.MessageAction = aws.String("SUPPRESS")
	}
	_, err := c.Client.AdminCreateUser(input)
	if err != nil {
		return err
	}

	if groupName == "" {
		// No group name provided, return early
		return nil
	}
	err = c.AddUserToGroup(email, groupName)
	if err != nil {
		err2 := c.DeleteUser(email)
		if err2 != nil {
			return err2
		}
		return errors.New("Error while adding user to group. user was NOT Created")
	}
	return err
}

func (c CognitoConfig) UserExists(email string) bool {
	c.Tx.Info("[core][cognito] Checkinh existence of User %v", email)
	user, err := c.GetUser(email)
	if err != nil {
		c.Tx.Error("[core][cognito] Error while Checkinh existence of User %v: %v", email, err)
		return false
	}
	return user.Email == email
}

func (c CognitoConfig) GetUser(email string) (User, error) {
	c.Tx.Info("[core][cognito] Retreiving User %v", email)
	input := &cognitoSdk.AdminGetUserInput{
		Username:   aws.String(email),
		UserPoolId: aws.String(c.UserPoolID),
	}
	out, err := c.Client.AdminGetUser(input)
	if err != nil {
		c.Tx.Error("[core][cognito] Error while Retreiving User %v: %v", email, err)
		return User{}, err
	}
	user := User{
		Username: aws.StringValue(out.Username),
		Status:   aws.StringValue(out.UserStatus),
	}
	for _, attr := range out.UserAttributes {
		if attr.Name != nil && *attr.Name == "email" {
			user.Email = *attr.Value
		}
	}
	c.Tx.Info("[core][cognito] found user: %+v", user)
	return user, nil
}

func (c CognitoConfig) GetGroup(name string) (Group, error) {
	c.Tx.Info("[core][cognito] Getting group %v", name)
	input := cognitoSdk.GetGroupInput{
		GroupName:  &name,
		UserPoolId: &c.UserPoolID,
	}
	out, err := c.Client.GetGroup(&input)
	if err != nil {
		return Group{}, err
	}
	group := Group{
		Name:        *out.Group.GroupName,
		Description: *out.Group.Description,
	}
	return group, nil
}

func (c CognitoConfig) ListGroupsForUser(username string) ([]Group, error) {
	c.Tx.Info("[core][cognito] Getting groups for user %v", username)
	input := cognitoSdk.AdminListGroupsForUserInput{
		UserPoolId: &c.UserPoolID,
		Username:   &username,
	}
	out, err := c.Client.AdminListGroupsForUser(&input)
	if err != nil {
		c.Tx.Error("[core][cognito] Error while getting groups for user %v: %v", username, err)
		awserr, ok := err.(awserr.Error)
		if ok && awserr.Code() == "TooManyRequestsException" {
			// Cognito is throttled, sleeping and retrying once
			c.Tx.Error("[core][cognito] Encountered TooManyRequestsException, retrying in 10 seconds")
			time.Sleep(10 * time.Second)
			out, err = c.Client.AdminListGroupsForUser(&input)
		}
		// check err (whether its the first or second attempt)
		if err != nil {
			return nil, err
		}
	}
	groups := []Group{}
	for _, g := range out.Groups {
		group := Group{
			Name:        aws.StringValue(g.GroupName),
			Description: aws.StringValue(g.Description),
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (c CognitoConfig) AddUserToGroup(email string, groupName string) error {
	c.Tx.Info("[core][cognito] Adding User %v to group %v", email, groupName)
	input := &cognitoSdk.AdminAddUserToGroupInput{
		Username:   aws.String(email),
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.AdminAddUserToGroup(input)
	return err
}

func (c CognitoConfig) SetUserPassword(email, password string, isPermanent bool) error {
	c.Tx.Info("[core][cognito] Setting User %v password", email)
	input := &cognitoSdk.AdminSetUserPasswordInput{
		Username:   aws.String(email),
		Password:   aws.String(password),
		UserPoolId: aws.String(c.UserPoolID),
		Permanent:  aws.Bool(isPermanent),
	}
	_, err := c.Client.AdminSetUserPassword(input)
	return err
}

func (c CognitoConfig) ReserUserPassword(email string) error {
	user, err := c.GetUser(email)
	if err != nil {
		c.Tx.Error("[core][cognito] Error while resetting PWD for User %v: %v", email, err)
		return err
	}
	if user.Status == "FORCE_CHANGE_PASSWORD" {
		c.Tx.Info("[core][cognito] User has not changed temporary password yet. resending invitation")
		return c.ResetUserInitialPassword(email)
	}
	c.Tx.Info("[core][cognito] User has already changed temporary password. resetting password")
	return c.ResetUserChangedPassword(email)
}

func (c CognitoConfig) ResetUserChangedPassword(email string) error {
	c.Tx.Info("[core][cognito] resetting password for User %v", email)
	input := &cognitoSdk.AdminResetUserPasswordInput{
		Username:   aws.String(email),
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.AdminResetUserPassword(input)
	return err
}

// to use is use has never used the autogenerated password
func (c CognitoConfig) ResetUserInitialPassword(email string) error {
	c.Tx.Info("[core][cognito] resetting password for User %v", email)
	input := &cognitoSdk.AdminCreateUserInput{
		Username: aws.String(email),
		DesiredDeliveryMediums: []*string{
			aws.String("EMAIL"),
		},
		MessageAction: aws.String("RESEND"),
		UserAttributes: []*cognitoSdk.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
		},
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.AdminCreateUser(input)
	return err
}

func (c CognitoConfig) DeleteUser(email string) error {
	c.Tx.Info("[core][cognito] Deleting User %v", email)
	input := &cognitoSdk.AdminDeleteUserInput{
		Username:   aws.String(email),
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.AdminDeleteUser(input)
	return err
}

// Todo; implement pagination
func (c CognitoConfig) UsersInGroup(groupName string) ([]User, error) {
	c.Tx.Info("[core][cognito] Search users in group %v", groupName)
	input := &cognitoSdk.ListUsersInGroupInput{
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(c.UserPoolID),
	}
	out, err := c.Client.ListUsersInGroup(input)
	c.Tx.Info("[core][cognito] user search response %+v", out)
	if err != nil {
		return []User{}, err
	}
	users := []User{}
	for _, user := range out.Users {
		if user != nil {
			for _, attr := range user.Attributes {
				if attr.Name != nil && *attr.Name == "email" {
					users = append(users, User{
						Username: aws.StringValue(user.Username),
						Email:    aws.StringValue(attr.Value),
					})
				}
			}
		}
	}
	return users, nil
}

func ParseGroupsFromLambdaRq(req events.APIGatewayProxyRequest) []string {
	log.Printf("[core][cognito] ParseGroupsFromLambdaRq: %+v", req)

	if req.RequestContext.Authorizer["claims"] != nil {
		claim, ok := req.RequestContext.Authorizer["claims"].(map[string]interface{})
		if ok {
			stringValue := claim["cognito:groups"].(string)
			return ParseGroupsFromString(stringValue, " ")
		}
	}
	return []string{}
}

func ParseUserFromLambdaRq(req events.APIGatewayProxyRequest) string {
	if req.RequestContext.Authorizer["claims"] != nil {
		claim, ok := (req.RequestContext.Authorizer["claims"].(map[string]interface{}))
		if ok {
			if claim["username"] != nil {
				return claim["username"].(string)
			}
			if claim["cognito:username"] != nil {
				return claim["cognito:username"].(string)
			}
		}
	}
	return ""
}

func ParseGroupsFromString(groupString string, separator string) []string {
	stringValue := groupString
	stringValue = strings.Replace(stringValue, "[", "", -1)
	stringValue = strings.Replace(stringValue, "]", "", -1)
	groups := strings.Split(stringValue, separator)
	for i, val := range groups {
		groups[i] = strings.TrimSpace(val)
	}
	return groups
}

// Used only for tests that require a valid user pool client,
// not to be used in production.
func (c *CognitoConfig) createTestClient(name string) error {
	input := cognitoSdk.CreateUserPoolClientInput{
		ClientName:        aws.String(name),
		ExplicitAuthFlows: []*string{aws.String(cognitoSdk.AuthFlowTypeAdminNoSrpAuth)},
		UserPoolId:        aws.String(c.UserPoolID),
	}
	output, err := c.Client.CreateUserPoolClient(&input)
	if err != nil {
		return err
	}
	c.ClientID = aws.StringValue(output.UserPoolClient.ClientId)
	return err
}

func (c CognitoConfig) IsGroupExistsException(err error) bool {
	if aerr, ok := err.(awserr.Error); ok {
		return aerr.Code() == cognitoSdk.ErrCodeGroupExistsException
	}
	return false
}
