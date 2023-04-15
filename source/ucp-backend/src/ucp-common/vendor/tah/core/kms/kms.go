package kms

import (
	"errors"
	"log"
	"strings"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awskms "github.com/aws/aws-sdk-go/service/kms"
)

type Config struct {
	Svc    *awskms.KMS
	Region string
}

func Init(region string) Config {
	cfg := aws.NewConfig().WithRegion(region)
	svc := awskms.New(session.New(), cfg)
	return Config{
		Svc:    svc,
		Region: region,
	}
}

func (kmsc Config) CreateKey(descritpion string) (string, error) {
	input := &awskms.CreateKeyInput{
		Description: aws.String(descritpion)}
	out, err := kmsc.Svc.CreateKey(input)
	if err != nil {
		log.Printf("Error creating key: %v", err)
		return "", err
	}
	log.Printf("Key Creation Response: : %v", out)
	return *(out.KeyMetadata).Arn, nil
}

func (kmsc Config) DeleteKey(keyArn string) error {
	keyId, err := arnToId(keyArn)
	if err != nil {
		return err
	}
	input := &awskms.ScheduleKeyDeletionInput{
		KeyId: aws.String(keyId),
	}
	_, err = kmsc.Svc.ScheduleKeyDeletion(input)
	if err != nil {
		log.Printf("Error scheduling key deletion: %v", err)
	}
	return err
}

func arnToId(keyArn string) (string, error) {
	res := strings.Split(keyArn, "/")
	if len(res) < 2 {
		return "", errors.New(fmt.Sprintf("Invalid Key Arn format: %s", keyArn))
	}
	return res[1], nil
}
