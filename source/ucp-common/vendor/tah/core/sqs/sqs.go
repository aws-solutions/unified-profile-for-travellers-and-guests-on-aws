package sqs

import (
	"encoding/json"
	"log"
	"strconv"
	core "tah/core/core"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

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

func Init(region string) Config {
	mySession := session.Must(session.NewSession())
	svc := sqs.New(mySession, aws.NewConfig().WithRegion(region))
	return Config{
		Client: svc,
	}
}

func InitWithQueueUrl(queueUrl string, region string) Config {
	mySession := session.Must(session.NewSession())
	svc := sqs.New(mySession, aws.NewConfig().WithRegion(region))
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

func (c Config) GetAttributes() (map[string]string, error) {
	input := &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(c.QueueUrl),
		AttributeNames: []*string{aws.String("All")},
	}
	out, err := c.Client.GetQueueAttributes(input)
	if err != nil {
		return map[string]string{}, err
	}
	log.Printf("[core][sqs][GetAttributes] Attributes %+v", out.Attributes)
	return core.ToMapString(out.Attributes), nil
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
			Statement{
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

func (c Config) Get() (QueuePeekContent, error) {
	attrs, err := c.GetAttributes()
	if err != nil {
		return QueuePeekContent{}, err
	}
	NMessages, err2 := strconv.ParseInt(attrs["ApproximateNumberOfMessages"], 10, 64)
	if err2 != nil {
		return QueuePeekContent{}, err
	}
	input := &sqs.ReceiveMessageInput{
		MaxNumberOfMessages:   aws.Int64(10),
		VisibilityTimeout:     aws.Int64(1),
		WaitTimeSeconds:       aws.Int64(1),
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
