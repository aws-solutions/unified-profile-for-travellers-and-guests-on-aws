// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package sqs

import (
	"encoding/json"
	"log"
	"strconv"
	core "tah/upt/source/tah-core/core"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type IConfig interface {
	Create(queueName string) (string, error)
	CreateRandom(prefix string) (string, error)
	GetAttributes() (map[string]string, error)
	GetQueueArn() (string, error)
	GetQueueUrl(name string) (string, error)
	SetPolicy(principalType string, principal string, actions []string) error
	SetAttribute(key string, value string) error
	Delete() error
	DeleteByName(queueName string) error
	SendWithStringAttributes(msg string, msgAttr map[string]string) error
	SendWithStringAttributesAndDelay(msg string, msgAttr map[string]string, delaySeconds int) error
	Get(options GetMessageOptions) (QueuePeekContent, error)
}

var _ IConfig = &Config{}

type Config struct {
	Client   *sqs.SQS
	QueueUrl string
}

type QueuePolicy struct {
	Version   string
	Id        string
	Statement []Statement
}

type Message struct {
	Body              string
	MessageAttributes map[string]string
	Attributes        map[string]string
}

type QueuePeekContent struct {
	NMessages int64
	Peek      []Message
}

type Statement struct {
	Sid       string
	Effect    string
	Principal map[string]string
	Action    []string
	Resource  string
}

// Options for receiving messages from a queue. A default value is used for any
// parameters that are not provided.
type GetMessageOptions struct {
	// The maximum number of messages to return. Amazon SQS never returns more
	// messages than this value (however, fewer messages might be returned).
	// Valid values: 1 to 10. Defaults to 10 messages.
	MaxNumberOfMessages int

	// The duration (in seconds) that the received messages are hidden from subsequent
	// retrieve requests after being retrieved by a ReceiveMessage request.
	// Defaults to 1 second.
	VisibilityTimeout int

	// The duration (in seconds) for which the call waits for a message to arrive in
	// the queue before returning. If a message is available, the call returns sooner
	// than WaitTimeSeconds. If no messages are available and the wait time expires,
	// the call returns successfully with an empty list of messages.
	// Defaults to 1 second.
	WaitTimeSeconds int
}

func Init(region, solutionId, solutionVersion string) Config {
	mySession := session.Must(session.NewSession())
	httpClient := core.CreateClient(solutionId, solutionVersion)
	svc := sqs.New(mySession, aws.NewConfig().WithRegion(region).WithHTTPClient(httpClient))
	return Config{
		Client: svc,
	}
}

func InitWithQueueUrl(queueUrl, region, solutionId, solutionVersion string) Config {
	mySession := session.Must(session.NewSession())
	httpClient := core.CreateClient(solutionId, solutionVersion)
	svc := sqs.New(mySession, aws.NewConfig().WithRegion(region).WithHTTPClient(httpClient))
	return Config{
		Client:   svc,
		QueueUrl: queueUrl,
	}
}

func (c *Config) Create(queueName string) (string, error) {
	input := &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	}
	out, err := c.Client.CreateQueue(input)
	if err != nil {
		return "", err
	}
	c.QueueUrl = *out.QueueUrl
	return *out.QueueUrl, nil
}

// creates a random queue with a prefix
func (c *Config) CreateRandom(prefix string) (string, error) {
	qName := prefix + core.GenerateUniqueId()
	return c.Create(qName)
}

func (c Config) GetAttributes() (map[string]string, error) {
	input := &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(c.QueueUrl),
		AttributeNames: []*string{aws.String("All")},
	}
	out, err := c.Client.GetQueueAttributes(input)
	if err != nil {
		return map[string]string{}, err
	}
	stringMap := core.ToMapString(out.Attributes)
	log.Printf("[core][sqs][GetAttributes] Attributes %+v", stringMap)
	return stringMap, nil
}

func (c Config) GetQueueArn() (string, error) {
	attrs, err := c.GetAttributes()
	if err != nil {
		return "", err
	}
	return attrs["QueueArn"], nil
}

func (c Config) GetQueueUrl(name string) (string, error) {
	input := &sqs.GetQueueUrlInput{QueueName: aws.String(name)}
	out, err := c.Client.GetQueueUrl(input)
	if err != nil {
		return "", err
	}
	return *out.QueueUrl, nil
}

func (c Config) SetPolicy(principalType string, principal string, actions []string) error {
	arn, err := c.GetQueueArn()
	if err != nil {
		return err
	}
	policy := QueuePolicy{
		Version: "2012-10-17",
		Statement: []Statement{
			{
				Effect: "Allow",
				Principal: map[string]string{
					principalType: principal,
				},
				Action:   actions,
				Resource: arn,
			},
		},
	}
	marshaledPolicy, err2 := json.Marshal(policy)
	if err2 != nil {
		return err2
	}

	err2 = c.SetAttribute("Policy", string(marshaledPolicy))
	return err2
}

func (c Config) SetAttribute(key string, value string) error {

	input := &sqs.SetQueueAttributesInput{
		QueueUrl: aws.String(c.QueueUrl),
		Attributes: map[string]*string{
			key: aws.String(value),
		},
	}
	_, err := c.Client.SetQueueAttributes(input)
	if err != nil {
		return err
	}
	return nil
}

func (c Config) Delete() error {
	log.Printf("[core][sqs][Delete] Deleting queue %s", c.QueueUrl)
	input := &sqs.DeleteQueueInput{
		QueueUrl: aws.String(c.QueueUrl),
	}
	_, err := c.Client.DeleteQueue(input)
	if err != nil {
		return err
	}
	return nil
}

func (c Config) DeleteByName(queueName string) error {
	url, err := c.GetQueueUrl(queueName)
	if err != nil {
		return err
	}
	input := &sqs.DeleteQueueInput{
		QueueUrl: aws.String(url),
	}
	_, err = c.Client.DeleteQueue(input)
	return err
}

func (c Config) Send(msg string) error {
	return c.SendWithStringAttributes(msg, map[string]string{})
}

func (c Config) SendWithStringAttributes(msg string, msgAttr map[string]string) error {
	input := &sqs.SendMessageInput{
		MessageBody: aws.String(msg),
		QueueUrl:    aws.String(c.QueueUrl),
	}
	if len(msgAttr) > 0 {
		input.MessageAttributes = toMessageAttributes(msgAttr)
	}
	_, err := c.Client.SendMessage(input)
	return err
}

func (c Config) SendWithStringAttributesAndDelay(msg string, msgAttr map[string]string, delaySeconds int) error {
	del := int64(delaySeconds)
	input := &sqs.SendMessageInput{
		MessageBody:  aws.String(msg),
		QueueUrl:     aws.String(c.QueueUrl),
		DelaySeconds: &del,
	}
	if len(msgAttr) > 0 {
		input.MessageAttributes = toMessageAttributes(msgAttr)
	}
	_, err := c.Client.SendMessage(input)
	return err
}

func toMessageAttributes(msgAttr map[string]string) map[string]*sqs.MessageAttributeValue {
	res := map[string]*sqs.MessageAttributeValue{}
	for key, val := range msgAttr {
		res[key] = &sqs.MessageAttributeValue{
			StringValue: aws.String(val),
			DataType:    aws.String("String"),
		}
	}
	return res
}

func (c Config) Get(options GetMessageOptions) (QueuePeekContent, error) {
	// Set default options if zero value provided
	if options.MaxNumberOfMessages == 0 {
		options.MaxNumberOfMessages = 10
	}
	if options.VisibilityTimeout == 0 {
		options.VisibilityTimeout = 1
	}
	if options.WaitTimeSeconds == 0 {
		options.WaitTimeSeconds = 1
	}
	// Get messages
	attrs, err := c.GetAttributes()
	if err != nil {
		return QueuePeekContent{}, err
	}
	NMessages, err := strconv.ParseInt(attrs["ApproximateNumberOfMessages"], 10, 64)
	if err != nil {
		return QueuePeekContent{}, err
	}
	input := &sqs.ReceiveMessageInput{
		MaxNumberOfMessages:   aws.Int64(int64(options.MaxNumberOfMessages)),
		VisibilityTimeout:     aws.Int64(int64(options.VisibilityTimeout)),
		WaitTimeSeconds:       aws.Int64(int64(options.WaitTimeSeconds)),
		QueueUrl:              aws.String(c.QueueUrl),
		MessageAttributeNames: []*string{aws.String("All")},
	}
	res, err := c.Client.ReceiveMessage(input)
	if err != nil {
		return QueuePeekContent{}, err
	}
	response := []Message{}
	for _, msg := range res.Messages {
		response = append(response, Message{Body: *msg.Body, MessageAttributes: msgAttributesToMapString(msg.MessageAttributes), Attributes: attributesToMapString(msg.Attributes)})
	}
	return QueuePeekContent{NMessages: NMessages, Peek: response}, err
}

func msgAttributesToMapString(mavs map[string]*sqs.MessageAttributeValue) map[string]string {
	res := map[string]string{}
	for key, mav := range mavs {
		if mav != nil && mav.StringValue != nil {
			res[key] = *mav.StringValue
		}
	}
	return res
}

func attributesToMapString(avs map[string]*string) map[string]string {
	res := map[string]string{}
	for key, av := range avs {
		if av != nil {
			res[key] = *av
		}
	}
	return res
}
