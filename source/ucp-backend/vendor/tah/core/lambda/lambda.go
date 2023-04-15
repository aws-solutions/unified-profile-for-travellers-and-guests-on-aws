package lambda

import (
	"encoding/json"
	"log"
	"strings"
	core "tah/core/core"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsLambda "github.com/aws/aws-sdk-go/service/lambda"
)

type DummyPayload struct{}

func (p DummyPayload) Decode(dec json.Decoder) (error, core.JSONObject) {
	return dec.Decode(&p), p
}

type Config struct {
	Client       *awsLambda.Lambda //sdk client to make call to the AWS API
	FunctionName string
}

func Init(function string) Config {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return Config{
		Client:       awsLambda.New(sess),
		FunctionName: function,
	}
}

func (c Config) InvokeAsync(payload interface{}) (interface{}, error) {
	return c.Invoke(payload, DummyPayload{}, true)
}

func (c Config) Invoke(payload interface{}, res core.JSONObject, async bool) (interface{}, error) {
	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[LAMBDA] Error marshaling request for lambda invoke: %+v", err)
		return res, err
	}
	invokType := "RequestResponse"
	if async {
		invokType = "Event"
	}
	input := &awsLambda.InvokeInput{
		Payload:        marshalledPayload,
		FunctionName:   aws.String(c.FunctionName),
		InvocationType: aws.String(invokType),
	}
	log.Printf("[LAMBDA] Lambda Request: %+v", input)
	output, err2 := c.Client.Invoke(input)
	if err2 != nil {
		log.Printf("[LAMBDA] Error invoking lambda: %+v", err2)
		return res, err2
	}
	log.Printf("[LAMBDA] Lambda Response: %+v", output)
	//Async lambda returns empty payload
	if !async {
		err, res = res.Decode(*json.NewDecoder(strings.NewReader(string(output.Payload))))
		if err != nil {
			log.Printf("[LAMBDA] Error unmarshaling response for lambda invoke: %+v", err)
			return res, err
		}
	}
	log.Printf("[LAMBDA] DecodedLambda Response: %+v", res)
	return res, err
}
