package kendra

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	sfn "github.com/aws/aws-sdk-go/service/sfn"
	"github.com/google/uuid"
)

type SfnCfg struct {
	Client          *sfn.SFN
	StateMachineArn string
}

func Init(arn string) SfnCfg {
	mySession := session.Must(session.NewSession())
	dbSession := sfn.New(mySession, aws.NewConfig())
	// Create DynamoDB client
	return SfnCfg{
		Client:          dbSession,
		StateMachineArn: arn,
	}

}

func (s SfnCfg) Start() error {
	input := &sfn.StartExecutionInput{
		Name:            aws.String(uuid.New().String()),
		StateMachineArn: aws.String(s.StateMachineArn),
	}
	log.Printf("[stepfunction][start] input %+v", input)
	_, err := s.Client.StartExecution(input)
	return err
}
