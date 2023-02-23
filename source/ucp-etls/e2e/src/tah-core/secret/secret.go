package secrets

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	secretsmanager "github.com/aws/aws-sdk-go/service/secretsmanager"
)

type Config struct {
	SecretArn string
	Mgr       *secretsmanager.SecretsManager
}

func Init(arn string) Config {
	mySession := session.Must(session.NewSession())
	return Config{
		SecretArn: arn,
		Mgr:       secretsmanager.New(mySession),
	}
}
func InitWithRegion(arn string, region string) Config {
	mySession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)}))
	return Config{
		SecretArn: arn,
		Mgr:       secretsmanager.New(mySession),
	}
}

//use this function if the secret is a key / val map
func (c Config) Get(key string) string {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(c.SecretArn),
	}
	out, err := c.Mgr.GetSecretValue(input)
	if err != nil {
		fmt.Println("[SECRET] error while retreiveing secret: %s ", err)
		return ""
	}
	res := map[string]string{}
	json.Unmarshal([]byte(*out.SecretString), &res)
	fmt.Println("[SECRET] Secret retreived from secret manager!")
	return res[key]
}

//use this function if teh secret is a single string value
func (c Config) GetRaw() string {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(c.SecretArn),
	}
	out, err := c.Mgr.GetSecretValue(input)
	if err != nil {
		fmt.Println("[SECRET] error while retreiveing secret: %s ", err)
		return ""
	}
	fmt.Println("[SECRET] Secret retreived from secret manager!")
	return *out.SecretString
}
