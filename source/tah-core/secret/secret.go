// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"encoding/json"
	"fmt"
	"log"

	"tah/upt/source/tah-core/core"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	secretsmanager "github.com/aws/aws-sdk-go/service/secretsmanager"
)

type Config struct {
	SecretArn string
	Mgr       *secretsmanager.SecretsManager
}

type Logininfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Init(arn, solutionId, solutionVersion string) Config {
	mySession := session.Must(session.NewSession())
	httpClient := core.CreateClient(solutionId, solutionVersion)
	return Config{
		SecretArn: arn,
		Mgr:       secretsmanager.New(mySession, &aws.Config{HTTPClient: httpClient}),
	}
}
func InitWithRegion(arn, region, solutionId, solutionVersion string) Config {
	httpClient := core.CreateClient(solutionId, solutionVersion)
	mySession := session.Must(session.NewSession(&aws.Config{
		Region:     aws.String(region),
		HTTPClient: httpClient,
	}))
	return Config{
		SecretArn: arn,
		Mgr:       secretsmanager.New(mySession),
	}
}

// use this function if the secret is a key / val map
func (c Config) Get(key string) string {
	log.Printf("Retreiving key %v from secret %v manager", key, c.SecretArn)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(c.SecretArn),
	}
	out, err := c.Mgr.GetSecretValue(input)
	if err != nil {
		fmt.Printf("[secret] error while retreiveing secret: %s ", err)
		return ""
	}
	res := map[string]string{}
	json.Unmarshal([]byte(*out.SecretString), &res)
	return res[key]
}

func (c Config) GetWithArnkey(key string, arn string) string {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(arn),
	}
	out, err := c.Mgr.GetSecretValue(input)
	if err != nil {
		fmt.Printf("[SECRET] error while retreiveing secret: %s ", err)
		return ""
	}
	res := map[string]string{}
	json.Unmarshal([]byte(*out.SecretString), &res)
	fmt.Println("[SECRET] Secret retreived from secret manager!")
	return res[key]
}

// use this function if teh secret is a single string value
func (c Config) GetRaw() string {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(c.SecretArn),
	}
	out, err := c.Mgr.GetSecretValue(input)
	if err != nil {
		fmt.Printf("[SECRET] error while retreiveing secret: %s ", err)
		return ""
	}
	fmt.Println("[SECRET] Secret retreived from secret manager!")
	return *out.SecretString
}

func (c Config) GenerateRandomPassword() (secretsmanager.GetRandomPasswordOutput, error) {
	password := &secretsmanager.GetRandomPasswordInput{
		PasswordLength: aws.Int64(64),
		ExcludeCharacters: aws.String(
			" %+~`#$&*()|[]{}:;<>?!'/@\\",
		),
		ExcludePunctuation: aws.Bool(true),
	}

	out, err := c.Mgr.GetRandomPassword(password)
	if err != nil {
		fmt.Printf("[SECRET] error while retreiveing secret: %s ", err)
		return secretsmanager.GetRandomPasswordOutput{}, err

	}
	return *out, nil
}

func (c *Config) CreateAuroraSecret(secretName string, username string, pwd string) (string, error) {

	loginInfo := Logininfo{
		Username: username,
		Password: pwd,
	}

	bytes, err := json.Marshal(loginInfo)
	if err != nil {
		log.Printf("[secret] error while marshalling secret: %s ", err)
		return "", err
	}

	input := &secretsmanager.CreateSecretInput{
		Name:         &secretName,
		SecretString: aws.String(string(bytes)),
	}

	secretOutput, err := c.Mgr.CreateSecret(input)
	if err != nil {
		log.Printf("[secret] error while creating secret: %s ", err)
		return "", err
	}
	arn := *secretOutput.ARN
	c.SecretArn = arn
	return arn, nil
}

func (c Config) CreateSecret(secretName string, username string) (secretsmanager.CreateSecretOutput, string, error) {

	rand, err := c.GenerateRandomPassword()
	if err != nil {
		fmt.Printf("[SECRET] error while retrieving secret: %s ", err)
		return secretsmanager.CreateSecretOutput{}, "", err
	}
	password := *rand.RandomPassword

	loginInfo := Logininfo{
		Username: username,
		Password: password,
	}

	bytes, err := json.Marshal(loginInfo)
	if err != nil {
		fmt.Printf("[SECRET] error while marshalling secret: %s ", err)
		return secretsmanager.CreateSecretOutput{}, "", err
	}

	input := &secretsmanager.CreateSecretInput{
		Name:         &secretName,
		SecretString: aws.String(string(bytes)),
	}

	secretOutput, err := c.Mgr.CreateSecret(input)
	if err != nil {
		fmt.Printf("[SECRET] error while creating secret: %s ", err)
		return secretsmanager.CreateSecretOutput{}, "", err
	}

	secretArn := *secretOutput.ARN
	fmt.Printf("[SECRET] Secret created with ARN: %s\n", secretArn)
	c.SecretArn = secretArn
	fmt.Printf("[SECRET] Secret created with ARN: %s\n", c.SecretArn)
	return *secretOutput, secretArn, nil
}

func (c Config) DeleteSecret(name string, forceDelete bool) error {
	input := &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(name),
		ForceDeleteWithoutRecovery: aws.Bool(forceDelete),
	}
	_, err := c.Mgr.DeleteSecret(input)
	if err != nil {
		fmt.Printf("[SECRET] error while deleting secret: %s ", err)
	}
	return nil
}

func (c Config) GetSecretArn(name string) (string, error) {
	input := &secretsmanager.ListSecretsInput{
		Filters: []*secretsmanager.Filter{{
			Key:    aws.String("name"),
			Values: []*string{aws.String(name)},
		}},
	}
	out, err := c.Mgr.ListSecrets(input)
	if err != nil {
		fmt.Printf("[secret] error while deleting secret: %s ", err)
	}
	if len(out.SecretList) == 0 {
		return "", fmt.Errorf("secret not found")
	}
	return *out.SecretList[0].ARN, nil
}
