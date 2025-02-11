// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lambda

import (
	"encoding/json"
	"log"
	"strings"
	core "tah/upt/source/tah-core/core"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsLambda "github.com/aws/aws-sdk-go/service/lambda"
)

var RUNTIME_GO = "go1.x"
var RUNTIME_PYTHON = "python3.7"
var RUNTIME_RUBY = "ruby2.7"
var RUNTIME_JAVA = "java8"
var RUNTIME_DOTNET = "dotnetcore3.1"
var RUNTIME_JS = "nodejs18.x"

type DummyPayload struct{}

func (p DummyPayload) Decode(dec json.Decoder) (error, core.JSONObject) {
	return dec.Decode(&p), p
}

type IConfig interface {
	Invoke(payload interface{}, res core.JSONObject, async bool) (interface{}, error)
	InvokeAsync(payload interface{}) (interface{}, error)
	CreateFunction(name string, roleArn string, s3Bucket string, s3Key string, options LambdaOptions) error
	DeleteFunction(name string) error
	CreateDynamoEventSourceMapping(dynamoDbStreamArn string, functionName string) error
	SearchDynamoEventSourceMappings(dynamoDbStreamArn string, functionName string) ([]EventSourceMapping, error)
	DeleteDynamoEventSourceMapping(uuid string) error
	Get(attr string) string
	ListTags() (map[string]string, error)
}

type Config struct {
	Client       *awsLambda.Lambda //sdk client to make call to the AWS API
	FunctionName string
}

type EventSourceMapping struct {
	StreamArn    string
	FunctionName string
	UUID         string
}

type LambdaOptions struct {
	Handler string
	Runtime string
}

func Init(function, solutionId, solutionVersion string) *Config {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	httpClient := core.CreateClient(solutionId, solutionVersion)
	return &Config{
		Client:       awsLambda.New(sess, &aws.Config{HTTPClient: httpClient}),
		FunctionName: function,
	}
}

func (c *Config) Get(str string) string {
	if str == "FunctionName" {
		return c.FunctionName
	}
	return ""
}

func (c *Config) InvokeAsync(payload interface{}) (interface{}, error) {
	return c.Invoke(payload, DummyPayload{}, true)
}

func (c *Config) Invoke(payload interface{}, res core.JSONObject, async bool) (interface{}, error) {
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

func (c *Config) CreateFunction(name string, roleArn string, s3Bucket string, s3Key string, options LambdaOptions) error {
	log.Printf("[lambda] creating Lambda function %v", name)
	runtime := RUNTIME_JS
	if options.Runtime != "" {
		runtime = options.Runtime
	}
	handler := "main"
	if options.Handler != "" {
		handler = options.Handler
	}

	input := &awsLambda.CreateFunctionInput{
		Code: &awsLambda.FunctionCode{
			S3Bucket: aws.String(s3Bucket),
			S3Key:    aws.String(s3Key),
		},
		FunctionName: aws.String(name),
		Runtime:      aws.String(runtime),
		Handler:      aws.String(handler),
		Role:         aws.String(roleArn),
	}
	_, err := c.Client.CreateFunction(input)
	if err != nil {
		log.Printf("[lambda] Error creating lambda function: %+v", err)
		return err
	}
	log.Printf("[lambda] Lambda function successfully created: %+v", name)
	return nil
}

func (c *Config) DeleteFunction(name string) error {
	log.Printf("[lambda] deleting Lambda function %v", name)
	input := &awsLambda.DeleteFunctionInput{
		FunctionName: aws.String(name),
	}
	_, err := c.Client.DeleteFunction(input)
	if err != nil {
		log.Printf("[lambda] Error deleting lambda function: %+v", err)
		return err
	}
	log.Printf("[lambda] Lambda function deleted: %+v", name)
	return nil
}

func (c *Config) CreateDynamoEventSourceMapping(dynamoDbStreamArn string, functionName string) error {
	log.Printf("[lambda] Creating event source mapping between DynamoDB stream %v and function  %+v", dynamoDbStreamArn, functionName)
	input := &awsLambda.CreateEventSourceMappingInput{
		EventSourceArn:   aws.String(dynamoDbStreamArn),
		FunctionName:     aws.String(functionName),
		StartingPosition: aws.String("TRIM_HORIZON"),
		BatchSize:        aws.Int64(10),
	}
	_, err := c.Client.CreateEventSourceMapping(input)
	if err != nil {
		log.Printf("[lambda] Error creating event source mapping: %+v", err)
	}
	return err
}

func (c *Config) SearchDynamoEventSourceMappings(dynamoDbStreamArn string, functionName string) ([]EventSourceMapping, error) {
	log.Printf("[lambda] Searching for event source mapping between DynamoDB stream %v and function  %+v", dynamoDbStreamArn, functionName)
	input := &awsLambda.ListEventSourceMappingsInput{
		EventSourceArn: aws.String(dynamoDbStreamArn),
		FunctionName:   aws.String(functionName),
	}
	out, err := c.Client.ListEventSourceMappings(input)
	if err != nil {
		log.Printf("[LAMBDA] Error creating event source mapping: %+v", err)
	}
	mappings := []EventSourceMapping{}
	for _, esm := range out.EventSourceMappings {
		mappings = append(mappings, EventSourceMapping{
			StreamArn:    *esm.EventSourceArn,
			FunctionName: arnToName(*esm.FunctionArn),
			UUID:         *esm.UUID,
		})
	}
	return mappings, err
}

func (c *Config) DeleteDynamoEventSourceMapping(uuid string) error {
	log.Printf("[lambda] Deleting event source mapping with uuid %v", uuid)
	input := &awsLambda.DeleteEventSourceMappingInput{
		UUID: aws.String(uuid),
	}
	_, err := c.Client.DeleteEventSourceMapping(input)
	if err != nil {
		log.Printf("[lambda] Error deleting event source mapping: %+v", err)
	}
	return err
}

func (c *Config) ListTags() (map[string]string, error) {
	getFunctionInput := &awsLambda.GetFunctionInput{
		FunctionName: aws.String(c.FunctionName),
	}
	function, err := c.Client.GetFunction(getFunctionInput)
	if err != nil {
		return make(map[string]string), err
	}
	return aws.StringValueMap(function.Tags), nil
}

func arnToName(arn string) string {
	segs := strings.Split(arn, ":")
	if len(segs) > 0 {
		return segs[len(segs)-1]
	}
	return ""
}

func arnToStreamId(arn string) string {
	segs := strings.Split(arn, "/")
	if len(segs) > 0 {
		return segs[len(segs)-1]
	}
	return ""
}
