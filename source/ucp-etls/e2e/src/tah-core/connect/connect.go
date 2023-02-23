package connect

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	connectSdk "github.com/aws/aws-sdk-go/service/connect"
)

type ConnectConfig struct {
	Service           *connectSdk.Connect
	ContactFlowId     string
	InstanceId        string
	SourcePhoneNumber string
}

func Init(instanceId string, contactFlowId string, region string, connectPhone string) ConnectConfig {
	mySession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)},
	))
	// Create a Pinpoint client from just a session.
	svc := connectSdk.New(mySession)
	// Create DynamoDB client
	return ConnectConfig{
		Service:           svc,
		ContactFlowId:     contactFlowId,
		InstanceId:        instanceId,
		SourcePhoneNumber: connectPhone,
	}
}

func (cfg ConnectConfig) Callback(phoneNumber string, attributes map[string]string) error {
	input := &connectSdk.StartOutboundVoiceContactInput{
		ContactFlowId:          aws.String(cfg.ContactFlowId),
		InstanceId:             aws.String(cfg.InstanceId),
		SourcePhoneNumber:      aws.String(cfg.SourcePhoneNumber),
		DestinationPhoneNumber: aws.String(phoneNumber),
		Attributes:             formatAtributes(attributes),
	}
	_, err := cfg.Service.StartOutboundVoiceContact(input)
	if err != nil {
		fmt.Printf("[CORE][CONNECT] StartOutboundVoiceContact error: %+v", err)
		return err
	}
	return nil
}

func formatAtributes(metrics map[string]string) map[string]*string {
	formated := map[string]*string{}
	for key, val := range metrics {
		formated[key] = aws.String(val)
	}
	return formated
}
