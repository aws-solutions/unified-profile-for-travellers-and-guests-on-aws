package cognito

import (
	"errors"
	"log"
	"strings"
	core "tah/core/core"

	b64 "encoding/base64"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognitoSdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

//Interfaces to abstract the cognito.CognitoConfig from tah/core object allowing mock-testing
type ICognitoConfig interface {
	GetAccessTokenFromClientCredentials(clientID string, clientSecret string, scope string) (string, error)
	SetTokenUrl(url string)
}

type CognitoConfig struct {
	Client     *cognitoSdk.CognitoIdentityProvider
	UserPoolID string
	ClientID   string
	TokenUrl   string //cognitoo token exchange URL
}

type User struct {
	Email  string
	Status string
}
type Group struct {
	Name string
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

func Init(userPoolID string) CognitoConfig {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create DynamoDB client
	return CognitoConfig{
		Client:     cognitoSdk.New(session),
		UserPoolID: userPoolID,
	}

}
func InitWithClientID(userPoolID string, clientID string) CognitoConfig {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return CognitoConfig{
		Client:     cognitoSdk.New(session),
		UserPoolID: userPoolID,
		ClientID:   clientID,
	}
}

func InitWithTokenUrl(userPoolID string, url string) CognitoConfig {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return CognitoConfig{
		Client:     cognitoSdk.New(session),
		UserPoolID: userPoolID,
		TokenUrl:   url,
	}
}

func (c *CognitoConfig) SetTokenUrl(url string) {
	c.TokenUrl = url
}

func (c CognitoConfig) RefreshToken(refreshToken string) (string, error) {
	log.Printf("[core][cognito] Refresh Access Token ")
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
		return "", errors.New("Cognito client must be initialized with token enpoint")
	}
	log.Printf("[core][cognito] perform authnetication using cognito endpoint")
	httpSvc := core.HttpInit(c.TokenUrl + "?grant_type=client_credentials")
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

func (c CognitoConfig) ClientSecret(clientID string) (string, error) {
	log.Printf("[core][cognito] Refresh Access Token ")
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
	log.Printf("[core][cognito] Creating user pool %v", name)
	input := &cognitoSdk.CreateUserPoolInput{
		PoolName: aws.String(name),
	}
	out, err := c.Client.CreateUserPool(input)
	c.UserPoolID = *out.UserPool.Id
	return err
}
func (c CognitoConfig) DeleteUserPool() error {
	log.Printf("[core][cognito] Deleting user pool %v", c.UserPoolID)
	input := &cognitoSdk.DeleteUserPoolInput{
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.DeleteUserPool(input)
	c.UserPoolID = ""
	return err
}

func (c CognitoConfig) CreateGroup(name string, description string) error {
	log.Printf("[core][cognito] Creating Group %v", name)
	input := &cognitoSdk.CreateGroupInput{
		GroupName:   aws.String(name),
		Description: aws.String(description),
		UserPoolId:  aws.String(c.UserPoolID),
	}
	_, err := c.Client.CreateGroup(input)
	return err
}
func (c CognitoConfig) DeleteGroup(name string) error {
	log.Printf("[core][cognito] Deleting Group %v", name)
	input := &cognitoSdk.DeleteGroupInput{
		GroupName:  aws.String(name),
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.DeleteGroup(input)
	return err
}
func (c CognitoConfig) CreateUser(email string, groupName string) error {
	log.Printf("[core][cognito] Creating User %v in group %v", email, groupName)
	input := &cognitoSdk.AdminCreateUserInput{
		Username: aws.String(email),
		DesiredDeliveryMediums: []*string{
			aws.String("EMAIL"),
		},
		UserAttributes: []*cognitoSdk.AttributeType{
			&cognitoSdk.AttributeType{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
			&cognitoSdk.AttributeType{
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
		},
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.AdminCreateUser(input)
	if err != nil {
		return err
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
	log.Printf("[core][cognito] Checkinh existence of User %v", email)
	user, err := c.GetUser(email)
	if err != nil {
		log.Printf("[core][cognito] Error while Checkinh existence of User %v: %v", email, err)
		return false
	}
	return user.Email == email
}

func (c CognitoConfig) GetUser(email string) (User, error) {
	log.Printf("[core][cognito] Retreiving User %v", email)
	input := &cognitoSdk.AdminGetUserInput{
		Username:   aws.String(email),
		UserPoolId: aws.String(c.UserPoolID),
	}
	out, err := c.Client.AdminGetUser(input)
	if err != nil {
		log.Printf("[core][cognito] Error while Retreiving User %v: %v", email, err)
		return User{}, err
	}
	user := User{Status: *out.UserStatus}
	for _, attr := range out.UserAttributes {
		if attr.Name != nil && *attr.Name == "email" {
			user.Email = *attr.Value
		}
	}
	log.Printf("[core][cognito] found user: %+v", user)
	return user, nil
}

func (c CognitoConfig) AddUserToGroup(email string, groupName string) error {
	log.Printf("[core][cognito] Adding User %v to group %v", email, groupName)
	input := &cognitoSdk.AdminAddUserToGroupInput{
		Username:   aws.String(email),
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.AdminAddUserToGroup(input)
	return err
}

func (c CognitoConfig) ReserUserPassword(email string) error {
	user, err := c.GetUser(email)
	if err != nil {
		log.Printf("[core][cognito] Error while resetting PWD for User %v: %v", email, err)
		return err
	}
	if user.Status == "FORCE_CHANGE_PASSWORD" {
		log.Printf("[core][cognito] User has not changed temporary password yet. resending invitation")
		return c.ResetUserInitialPassword(email)
	}
	log.Printf("[core][cognito] User has already changed temporary password. resetting password")
	return c.ResetUserChangedPassword(email)
}

func (c CognitoConfig) ResetUserChangedPassword(email string) error {
	log.Printf("[core][cognito] resetting password for User %v", email)
	input := &cognitoSdk.AdminResetUserPasswordInput{
		Username:   aws.String(email),
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.AdminResetUserPassword(input)
	return err
}

//to use is use has never used the autogenerated password
func (c CognitoConfig) ResetUserInitialPassword(email string) error {
	log.Printf("[core][cognito] resetting password for User %v", email)
	input := &cognitoSdk.AdminCreateUserInput{
		Username: aws.String(email),
		DesiredDeliveryMediums: []*string{
			aws.String("EMAIL"),
		},
		MessageAction: aws.String("RESEND"),
		UserAttributes: []*cognitoSdk.AttributeType{
			&cognitoSdk.AttributeType{
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
	log.Printf("[core][cognito] Deleting User %v", email)
	input := &cognitoSdk.AdminDeleteUserInput{
		Username:   aws.String(email),
		UserPoolId: aws.String(c.UserPoolID),
	}
	_, err := c.Client.AdminDeleteUser(input)
	return err
}

//Todo; implement pagination
func (c CognitoConfig) UsersInGroup(groupName string) ([]User, error) {
	log.Printf("[core][cognito] Search users in group %v", groupName)
	input := &cognitoSdk.ListUsersInGroupInput{
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(c.UserPoolID),
	}
	out, err := c.Client.ListUsersInGroup(input)
	log.Printf("[core][cognito] user search response %+v", out)
	if err != nil {
		return []User{}, err
	}
	users := []User{}
	for _, user := range out.Users {
		if user != nil {
			for _, attr := range user.Attributes {
				if attr.Name != nil && *attr.Name == "email" {
					users = append(users, User{
						Email: *attr.Value,
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
		stringValue := (req.RequestContext.Authorizer["claims"].(map[string]interface{}))["cognito:groups"].(string)
		return ParseGroupsFromString(stringValue, " ")
	}
	return []string{}
}

func ParseUserFromLambdaRq(req events.APIGatewayProxyRequest) string {
	if req.RequestContext.Authorizer["claims"] != nil {
		return (req.RequestContext.Authorizer["claims"].(map[string]interface{}))["username"].(string)
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
