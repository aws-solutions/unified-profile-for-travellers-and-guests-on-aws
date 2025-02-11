// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package eventbridge

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	core "tah/upt/source/tah-core/core"
	iam "tah/upt/source/tah-core/iam"
	sqs "tah/upt/source/tah-core/sqs"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/eventbridge"
)

var SQS_QUEUE_NAME_PRERFIX = "dlq-eventbridg-api-dest-"

type Config struct {
	Svc             *cloudwatchevents.CloudWatchEvents
	EbSvc           *eventbridge.EventBridge
	Tx              core.Transaction
	Region          string
	EventBus        string
	solutionID      string
	solutionVersion string
}

type IConfig interface {
	SendMessage(source string, description string, data interface{}) error
}

type ApiRule struct {
	Schedule    string
	State       string
	Name        string
	Arn         string
	Description string
	Targets     []string
}

type EventError struct {
	ErrorCode    string
	ErrorMessage string
	Rule         string
	Target       string
	Request      string
	Sent         time.Time
}

func Init(region string, solutionId, solutionVersion string) Config {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := core.CreateClient(solutionId, solutionVersion)
	// Create DynamoDB client
	return Config{
		Svc:             cloudwatchevents.New(session, &aws.Config{HTTPClient: client}),
		EbSvc:           eventbridge.New(session),
		Region:          region,
		solutionID:      solutionId,
		solutionVersion: solutionVersion,
	}
}

func InitWithBus(region string, busName string, solutionId, solutionVersion string) Config {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := core.CreateClient(solutionId, solutionVersion)
	// Create DynamoDB client
	return Config{
		Svc:             cloudwatchevents.New(session, &aws.Config{HTTPClient: client}),
		EbSvc:           eventbridge.New(session),
		Region:          region,
		EventBus:        busName,
		solutionID:      solutionId,
		solutionVersion: solutionVersion,
	}
}

func InitWithRandomBus(region string, prefix string, solutionId, solutionVersion string) (Config, error) {
	//random busname
	busName := prefix + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	cfg := Init(region, solutionId, solutionVersion)
	_, err := (&cfg).CreateBus(busName)
	return InitWithBus(region, busName, solutionId, solutionVersion), err
}

func (cfg *Config) SetTx(tx core.Transaction) error {
	tx.LogPrefix = "EVENTBRIDGE"
	cfg.Tx = tx
	return nil
}

// nCalls number of parallel calls to make to the enpoint
// ruleName unique rule name (see evenbridge API for coonstraints)
// rateMinutes rate of API call in minutes
// httpMethod HTTP method to use to call API endpointt on schedule
// endpoint API endpoint to call on schedule
// body: static body to be send to api endpoint
// cognitoEndpoint needed for client_credentials grant
// clientID needed for client_credentials grant
// clientSecret needed for client_credentials grant
// oAuthScope needed for client_credentials grant
func (cfg *Config) CreateHttpApiScheduledTrigger(nCalls int, ruleName string, rateMinutes int64, httpMethod string, endpoint string, body interface{}, cognitoEndpoint string, clientID string, clientSecret string, oAuthScope string) error {
	cfg.Tx.Info("CreateHttpApiScheduledTrigger")
	if nCalls == 0 {
		return errors.New("Number of targets must be greater than 0")
	}
	if rateMinutes == 0 {
		return errors.New("rateMinutes must be greater than 0")
	}
	if ruleName == "" {
		return errors.New("ruleName must not be empty")
	}
	cfg.Tx.Debug("1-Create Connection is Not Exist(if Not Exists")
	conArn, err := cfg.CognitoConnectionExists(cognitoEndpoint)
	if err != nil {
		cfg.Tx.Error("error checking is cognito connection exists %v", err)
		return err
	}
	if conArn == "" {
		conArn, err = cfg.CreateCognitoConnection(cognitoEndpoint, clientID, clientSecret, oAuthScope)
		if err != nil {
			cfg.Tx.Error("error creating cognito connection %v", err)
			return err
		}
	}
	cfg.Tx.Debug("2-Create API Destination (if Not Exists)")
	apiDestArn, err := cfg.ApiDestinationExists(endpoint)
	if err != nil {
		cfg.Tx.Error("error checking is cognito connection exists %v", err)
		return err
	}
	if apiDestArn == "" {
		apiDestArn, err = cfg.CreateApiDestination(conArn, httpMethod, endpoint)
		if err != nil {
			cfg.Tx.Error("error creating cognito connection %v", err)
			return err
		}
	}
	cfg.Tx.Debug("2-Create Rule")
	ruleArn, err := cfg.ApiRuleExist(ruleName)
	if err != nil {
		cfg.Tx.Error("error checking is rule exists %v", err)
		return err
	}
	if ruleArn == "" {
		err = cfg.CreateApiTrigger(nCalls, ruleName, rateMinutes, apiDestArn, body)
		if err != nil {
			cfg.Tx.Error("error creating Api destination trigger: %v", err)
			return err
		}
	}
	return nil
}

func (cfg *Config) CognitoConnectionExists(endpoint string) (string, error) {
	input := &cloudwatchevents.DescribeConnectionInput{
		Name: aws.String(buildCognitoConnectionName(endpoint)),
	}
	out, err := cfg.Svc.DescribeConnection(input)
	if err != nil {
		return "", nil
	}
	if out.ConnectionArn != nil && *out.ConnectionArn != "" {
		return *out.ConnectionArn, nil
	}
	return "", nil
}

// eventbridge has a constraings on the name
func buildCognitoConnectionName(endpoint string) string {
	return "connection-" + strings.Split(strings.Split(endpoint, ".")[0], "//")[1]
}

func (cfg *Config) DeleteCognitoConnection(endpoint string) error {
	input := &cloudwatchevents.DeleteConnectionInput{
		Name: aws.String(buildCognitoConnectionName(endpoint)),
	}
	_, err := cfg.Svc.DeleteConnection(input)
	return err
}

func (cfg *Config) CreateCognitoConnection(endpoint string, clientID string, clientSecret string, oAuthScope string) (string, error) {
	input := &cloudwatchevents.CreateConnectionInput{
		AuthorizationType: aws.String(cloudwatchevents.ConnectionAuthorizationTypeOauthClientCredentials),
		Description:       aws.String("Connection to cognito endpoint " + endpoint),
		Name:              aws.String(buildCognitoConnectionName(endpoint)),
		AuthParameters: &cloudwatchevents.CreateConnectionAuthRequestParameters{
			OAuthParameters: &cloudwatchevents.CreateConnectionOAuthRequestParameters{
				AuthorizationEndpoint: aws.String(endpoint),
				HttpMethod:            aws.String("POST"),
				ClientParameters: &cloudwatchevents.CreateConnectionOAuthClientRequestParameters{
					ClientID:     aws.String(clientID),
					ClientSecret: aws.String(clientSecret),
				},
				OAuthHttpParameters: &cloudwatchevents.ConnectionHttpParameters{
					QueryStringParameters: []*cloudwatchevents.ConnectionQueryStringParameter{
						{
							Key:   aws.String("grant_type"),
							Value: aws.String("client_credentials"),
						},
						{
							Key:   aws.String("scope"),
							Value: aws.String(oAuthScope),
						},
					},
				},
			},
		},
	}
	out, err := cfg.Svc.CreateConnection(input)
	if out.ConnectionArn != nil {
		return *out.ConnectionArn, err
	}
	return "", err
}

// eventbridge has a constraings on the name
func buildCognitoEndpointName(endpoint string) string {
	return "cognito-api-destination-" + strings.Split(strings.Split(endpoint, ".")[0], "//")[1]
}

func (cfg *Config) ApiDestinationExists(apiEndpoint string) (string, error) {
	input := &cloudwatchevents.DescribeApiDestinationInput{
		Name: aws.String(buildCognitoEndpointName(apiEndpoint)),
	}
	out, err := cfg.Svc.DescribeApiDestination(input)
	if err != nil {
		return "", nil
	}
	if out.ApiDestinationArn != nil && *out.ApiDestinationArn != "" {
		return *out.ApiDestinationArn, nil
	}
	return "", nil
}

func (cfg *Config) CreateApiDestination(connectionArn string, httpMethod string, apiEndpoint string) (string, error) {
	input := &cloudwatchevents.CreateApiDestinationInput{
		ConnectionArn:      aws.String(connectionArn),
		Description:        aws.String("Api Destination for cognito endpoint" + apiEndpoint),
		Name:               aws.String(buildCognitoEndpointName(apiEndpoint)),
		HttpMethod:         aws.String(httpMethod),
		InvocationEndpoint: aws.String(apiEndpoint),
	}
	out, err := cfg.Svc.CreateApiDestination(input)
	if out.ApiDestinationArn != nil {
		return *out.ApiDestinationArn, err
	}
	return "", err
}

func (cfg *Config) DeleteApiDestination(apiEndpoint string) error {
	input := &cloudwatchevents.DeleteApiDestinationInput{
		Name: aws.String(buildCognitoEndpointName(apiEndpoint)),
	}
	_, err := cfg.Svc.DeleteApiDestination(input)
	return err
}

func (cfg *Config) CreateApiTrigger(nCalls int, ruleName string, rateMinutes int64, apiDestinationArn string, bodyObj interface{}) error {
	rate := buildScheduleExpression(rateMinutes)
	input := &cloudwatchevents.PutRuleInput{
		Name:               aws.String(ruleName),
		Description:        aws.String("Trigger for " + apiDestinationArn + " every " + rate),
		ScheduleExpression: aws.String(rate),
		State:              aws.String(eventbridge.RuleStateEnabled),
	}
	_, err := cfg.Svc.PutRule(input)
	if err != nil {
		return err
	}
	return cfg.AddApiTarget(nCalls, ruleName, apiDestinationArn, bodyObj)
}

func (cfg *Config) CreateSQSTarget(name, eventBusName, queueUrl, eventPattern string) error {
	// Allow messages to be sent from EventBridge to SQS queue
	sqsCfg := sqs.InitWithQueueUrl(queueUrl, cfg.Region, cfg.solutionID, cfg.solutionVersion)
	err := sqsCfg.SetPolicy("Service", "events.amazonaws.com", []string{"SQS:*"})
	if err != nil {
		cfg.Tx.Error("[CreateSQSTarget] Could not set policy: %v", err)
		return err
	}
	queueArn, err := sqsCfg.GetQueueArn()
	if err != nil {
		cfg.Tx.Error("[CreateSQSTarget] Could not get queue arn: %v", err)
		return err
	}

	ruleInput := &cloudwatchevents.PutRuleInput{
		Name:         aws.String(name),
		Description:  aws.String("Rule for " + queueArn),
		EventBusName: aws.String(eventBusName),
		EventPattern: aws.String(eventPattern),
		State:        aws.String(eventbridge.RuleStateEnabled),
	}
	_, err = cfg.Svc.PutRule(ruleInput)
	if err != nil {
		cfg.Tx.Error("[CreateSQSTarget] Could not create rule: %v", err)
		return err
	}

	targetInput := &cloudwatchevents.PutTargetsInput{
		EventBusName: aws.String(eventBusName),
		Rule:         aws.String(name),
		Targets: []*cloudwatchevents.Target{
			{
				Arn: aws.String(queueArn),
				Id:  aws.String(name),
			},
		},
	}
	_, err = cfg.Svc.PutTargets(targetInput)
	if err != nil {
		cfg.Tx.Error("[CreateSQSTarget] Could not create target %v", err)
		return err
	}

	return nil
}

func buildScheduleExpression(rateMin int64) string {
	return "rate(" + strconv.FormatInt(rateMin, 10) + " minutes)"
}

func (cfg *Config) ApiRuleExist(ruleName string) (string, error) {
	input := &cloudwatchevents.DescribeRuleInput{
		Name: aws.String(ruleName),
	}
	out, err := cfg.Svc.DescribeRule(input)
	if err != nil {
		return "", nil
	}
	if out.Arn != nil && *out.Arn != "" {
		return *out.Arn, nil
	}
	return "", nil
}

func (cfg *Config) RetreiveApiRule(ruleName string) (ApiRule, error) {
	input := &cloudwatchevents.DescribeRuleInput{
		Name: aws.String(ruleName),
	}
	out, err := cfg.Svc.DescribeRule(input)
	if err != nil {
		return ApiRule{}, err
	}

	listInput := &cloudwatchevents.ListTargetsByRuleInput{
		Rule: aws.String(ruleName),
	}
	listTrargerRes, err2 := cfg.Svc.ListTargetsByRule(listInput)
	if err2 != nil {
		return ApiRule{}, nil
	}
	targets := []string{}
	for _, target := range listTrargerRes.Targets {
		targets = append(targets, *target.Arn)
	}
	return ApiRule{
		Schedule:    *out.ScheduleExpression,
		State:       *out.State,
		Name:        *out.Name,
		Arn:         *out.Arn,
		Description: *out.Description,
		Targets:     targets,
	}, nil
}

func (cfg *Config) DeleteApiRule(ruleName string) error {
	err := cfg.RemoveTargets(ruleName)
	if err != nil {
		return err
	}
	input := &cloudwatchevents.DeleteRuleInput{
		Name: aws.String(ruleName),
	}
	_, err = cfg.Svc.DeleteRule(input)
	return err
}

func (cfg *Config) DeleteRuleByName(ruleName string) error {
	cfg.Tx.Info("Deleting Rule %v", ruleName)
	err := cfg.RemoveTargets(ruleName)
	if err != nil {
		cfg.Tx.Error("error removing targets for rule %v: %v", ruleName, err)
		return err
	}
	cfg.Tx.Debug("Successfully removed target for rule %v", ruleName)
	input := &cloudwatchevents.DeleteRuleInput{
		Name: aws.String(ruleName),
	}
	cfg.Tx.Debug("Delete input: %+v", input)
	_, err = cfg.Svc.DeleteRule(input)
	if err != nil {
		cfg.Tx.Error("error deleting rule %v: %v", ruleName, err)
	}
	return err
}

// Add 1 or many times the same target to a rule (this allows to create parallel request to the same endpoint)
func (cfg *Config) AddApiTarget(nCalls int, ruleName string, apiDestinationArn string, bodyObj interface{}) error {
	if nCalls == 0 {
		return errors.New("Number of targets must be greater than 0")
	}
	cfg.Tx.Info("Add API Target to event rule")
	cfg.Tx.Debug("1-Create DLQ")
	sqsCfg := sqs.Init(cfg.Region, cfg.solutionID, cfg.solutionVersion)
	_, err := sqsCfg.Create(SQS_QUEUE_NAME_PRERFIX + ruleName)
	if err != nil {
		return err
	}
	cfg.Tx.Debug("2-Set policy for Connect profile access")
	err = sqsCfg.SetPolicy("Service", "events.amazonaws.com", []string{"SQS:*"})
	if err != nil {
		cfg.Tx.Error("Could not set policy on dead letter queue")
		return err
	}
	queueArn, err2 := sqsCfg.GetQueueArn()
	if err2 != nil {
		return err2
	}
	cfg.Tx.Debug("3-Create body")
	jsonBody, err3 := json.Marshal(bodyObj)
	if err3 != nil {
		return err3
	}
	cfg.Tx.Debug("4-Create role (delete exisingrole)")
	iamCfg := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	roleArn := iamCfg.RoleExists(ruleName)
	if roleArn != "" {
		err3 = iamCfg.DeleteRole(ruleName)
		if err3 != nil {
			cfg.Tx.Error("Could not deleet existing role")
			return err3
		}
	}
	roleArn, _, err3 = iamCfg.CreateRoleWithActionResource(ruleName, "events.amazonaws.com", "events:InvokeApiDestination", apiDestinationArn)
	if err3 != nil {
		cfg.Tx.Error("Could not create IAM role")
		return err3
	}
	input := &cloudwatchevents.PutTargetsInput{
		Rule:    aws.String(ruleName),
		Targets: []*cloudwatchevents.Target{},
	}
	for i := 0; i < nCalls; i++ {
		input.Targets = append(input.Targets, &cloudwatchevents.Target{
			Id:  aws.String(ruleName + "-" + strconv.Itoa(i)),
			Arn: aws.String(apiDestinationArn),
			DeadLetterConfig: &cloudwatchevents.DeadLetterConfig{
				Arn: aws.String(queueArn),
			},
			Input:   aws.String(string(jsonBody)),
			RoleArn: aws.String(roleArn),
		})
	}
	_, err = cfg.Svc.PutTargets(input)
	return err
}

func (cfg *Config) GetErrors(ruleName string) ([]EventError, int64, error) {
	sqsSvc := sqs.Init(cfg.Region, cfg.solutionID, cfg.solutionVersion)
	url, err := sqsSvc.GetQueueUrl(SQS_QUEUE_NAME_PRERFIX + ruleName)
	if err != nil {
		cfg.Tx.Error("Error getting queueu URL from name %v", err)
		return []EventError{}, 0, err
	}
	cfg.Tx.Debug("found dead letter queue %+v", url)
	sqsSvc = sqs.InitWithQueueUrl(url, cfg.Region, cfg.solutionID, cfg.solutionVersion)
	//TODO: fix this. will not sacle. dump DQD content into Dynamo asyncroniously
	res, err2 := sqsSvc.Get(sqs.GetMessageOptions{})
	if err2 != nil {
		cfg.Tx.Error(" Error fetching from SQS queue %s: %v", url, err2)
		return []EventError{}, 0, err2
	}
	cfg.Tx.Debug("SQS response: %+v ", res)
	eventErrors := []EventError{}
	for _, msg := range res.Peek {
		ts, err3 := core.ParseEpochMs(msg.Attributes["SentTimestamp"])
		if err3 != nil {
			cfg.Tx.Error(" Error parsing timestamp %s: %v", msg.Attributes["SentTimestamp"], err3)
		}
		eventErrors = append(eventErrors, EventError{
			ErrorCode:    msg.MessageAttributes["ERROR_CODE"],
			ErrorMessage: msg.MessageAttributes["ERROR_MESSAGE"],
			Rule:         msg.MessageAttributes["RULE_ARN"],
			Target:       msg.MessageAttributes["TARGET_ARN"],
			Sent:         ts,
			Request:      msg.Body,
		})
	}
	return eventErrors, res.NMessages, nil
}

func (cfg *Config) RemoveTargets(ruleName string) error {
	cfg.Tx.Info("Removing targets from Rule %v", ruleName)
	input := &cloudwatchevents.ListTargetsByRuleInput{
		Rule: aws.String(ruleName),
	}
	out, err := cfg.Svc.ListTargetsByRule(input)
	if err != nil {
		return nil
	}
	ids := []*string{}
	for _, target := range out.Targets {
		ids = append(ids, target.Id)
	}
	if len(ids) == 0 {
		cfg.Tx.Info("[RemoveTargets] No target to remove. returning")
		return nil
	}
	remTarInput := &cloudwatchevents.RemoveTargetsInput{
		Rule: aws.String(ruleName),
		Ids:  ids,
	}
	_, err = cfg.Svc.RemoveTargets(remTarInput)
	if err != nil {
		cfg.Tx.Error("[RemoveTargets] Error removing targets: %v", err)
		return err
	}
	sqsCfg := sqs.Init(cfg.Region, cfg.solutionID, cfg.solutionVersion)
	err = sqsCfg.DeleteByName(SQS_QUEUE_NAME_PRERFIX + ruleName)
	return err
}

func (cfg *Config) CreateBus(name string) (string, error) {
	input := &eventbridge.CreateEventBusInput{
		Name: aws.String(name),
	}
	out, err := cfg.EbSvc.CreateEventBus(input)
	if err != nil {
		return "", err
	}
	return *out.EventBusArn, err
}

func (cfg *Config) DeleteBus(name string) error {
	input := &eventbridge.DeleteEventBusInput{
		Name: aws.String(name),
	}
	_, err := cfg.EbSvc.DeleteEventBus(input)
	return err
}

func (cfg *Config) SendMessage(source string, description string, data interface{}) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}
	input := eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{
			{
				Detail:       aws.String(string(msg)),
				DetailType:   aws.String(description),
				Source:       aws.String(source),
				EventBusName: aws.String(cfg.EventBus),
				Resources:    []*string{},
			},
		},
	}
	_, err = cfg.EbSvc.PutEvents(&input)
	return err
}
