package eventbridge

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	core "tah/core/core"
	iam "tah/core/iam"
	sqs "tah/core/sqs"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

var SQS_QUEUE_NAME_PRERFIX = "dlq-eventbridg-api-dest-"

type Config struct {
	Svc    *cloudwatchevents.CloudWatchEvents
	Tx     core.Transaction
	Region string
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

func Init(region string) Config {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create DynamoDB client
	return Config{
		Svc:    cloudwatchevents.New(session),
		Region: region,
	}
}

func (cfg *Config) SetTx(tx core.Transaction) error {
	tx.LogPrefix = "EVENTBRIDGE"
	cfg.Tx = tx
	return nil
}

//nCalls number of parallel calls to make to the enpoint
//ruleName unique rule name (see evenbridge API for coonstraints)
//rateMinutes rate of API call in minutes
//httpMethod HTTP method to use to call API endpointt on schedule
//endpoint API endpoint to call on schedule
//body: static body to be send to api endpoint
//cognitoEndpoint needed for client_credentials grant
//clientID needed for client_credentials grant
//clientSecret needed for client_credentials grant
//oAuthScope needed for client_credentials grant
func (cfg *Config) CreateHttpApiScheduledTrigger(nCalls int, ruleName string, rateMinutes int64, httpMethod string, endpoint string, body interface{}, cognitoEndpoint string, clientID string, clientSecret string, oAuthScope string) error {
	cfg.Tx.Log("CreateHttpApiScheduledTrigger")
	if nCalls == 0 {
		return errors.New("Number of targets must be greater than 0")
	}
	if rateMinutes == 0 {
		return errors.New("rateMinutes must be greater than 0")
	}
	if ruleName == "" {
		return errors.New("ruleName must not be empty")
	}
	cfg.Tx.Log("1-Create Connection is Not Exist(if Not Exists")
	conArn, err := cfg.CognitoConnectionExists(cognitoEndpoint)
	if err != nil {
		cfg.Tx.Log("error checking is cognito connection exists %v", err)
		return err
	}
	if conArn == "" {
		conArn, err = cfg.CreateCognitoConnection(cognitoEndpoint, clientID, clientSecret, oAuthScope)
		if err != nil {
			cfg.Tx.Log("error creating cognito connection %v", err)
			return err
		}
	}
	cfg.Tx.Log("2-Create API Destination (if Not Exists)")
	apiDestArn, err2 := cfg.ApiDestinationExists(endpoint)
	if err2 != nil {
		cfg.Tx.Log("error checking is cognito connection exists %v", err)
		return err2
	}
	if apiDestArn == "" {
		apiDestArn, err2 = cfg.CreateApiDestination(conArn, httpMethod, endpoint)
		if err2 != nil {
			cfg.Tx.Log("error creating cognito connection %v", err2)
			return err2
		}
	}
	cfg.Tx.Log("2-Create Rule")
	ruleArn, err3 := cfg.ApiRuleExist(ruleName)
	if err3 != nil {
		cfg.Tx.Log("error checking is rule exists %v", err3)
		return err3
	}
	if ruleArn == "" {
		err3 = cfg.CreateApiTrigger(nCalls, ruleName, rateMinutes, apiDestArn, body)
		if err3 != nil {
			cfg.Tx.Log("error creating Api destination trigger: %v", err3)
			return err3
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

//eventbridge has a constraings on the name
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
						&cloudwatchevents.ConnectionQueryStringParameter{
							Key:   aws.String("grant_type"),
							Value: aws.String("client_credentials"),
						},
						&cloudwatchevents.ConnectionQueryStringParameter{
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

//eventbridge has a constraings on the name
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
		State:              aws.String("ENABLED"),
	}
	_, err := cfg.Svc.PutRule(input)
	if err != nil {
		return err
	}
	return cfg.AddApiTarget(nCalls, ruleName, apiDestinationArn, bodyObj)
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

//Add 1 or many times the same target to a rule (this allows to create parallel request to the same endpoint)
func (cfg *Config) AddApiTarget(nCalls int, ruleName string, apiDestinationArn string, bodyObj interface{}) error {
	if nCalls == 0 {
		return errors.New("Number of targets must be greater than 0")
	}
	cfg.Tx.Log("Add API Target to event rule")
	cfg.Tx.Log("1-Create DLQ")
	sqsCfg := sqs.Init(cfg.Region)
	_, err := sqsCfg.Create(SQS_QUEUE_NAME_PRERFIX + ruleName)
	if err != nil {
		return err
	}
	cfg.Tx.Log("2-Set policy for Connect profile access")
	err = sqsCfg.SetPolicy("Service", "events.amazonaws.com", []string{"SQS:*"})
	if err != nil {
		cfg.Tx.Log("Could not set policy on dead letter queue")
		return err
	}
	queueArn, err2 := sqsCfg.GetQueueArn()
	if err2 != nil {
		return err2
	}
	cfg.Tx.Log("3-Create body")
	jsonBody, err3 := json.Marshal(bodyObj)
	if err3 != nil {
		return err3
	}
	cfg.Tx.Log("4-Create role (delete exisingrole)")
	iamCfg := iam.Init()
	roleArn := iamCfg.RoleExists(ruleName)
	if roleArn != "" {
		err3 = iamCfg.DeleteRole(ruleName)
		if err3 != nil {
			cfg.Tx.Log("Could not deleet existing role")
			return err3
		}
	}
	roleArn, _, err3 = iamCfg.CreateRoleWithActionResource(ruleName, "events.amazonaws.com", "events:InvokeApiDestination", apiDestinationArn)
	if err3 != nil {
		cfg.Tx.Log("Could not create IAM role")
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
	sqsSvc := sqs.Init(cfg.Region)
	url, err := sqsSvc.GetQueueUrl(SQS_QUEUE_NAME_PRERFIX + ruleName)
	if err != nil {
		log.Printf("Error getting queueu URL from name %v", err)
		return []EventError{}, 0, err
	}
	log.Printf("found dead letter queue %+v", url)
	sqsSvc = sqs.InitWithQueueUrl(url, cfg.Region)
	//TODO: fix this. will not sacle. dump DQD content into Dynamo asyncroniously
	res, err2 := sqsSvc.Get()
	if err2 != nil {
		log.Printf(" Error fetching from SQS queue %s: %v", url, err2)
		return []EventError{}, 0, err2
	}
	log.Printf("SQS response: %+v ", res)
	eventErrors := []EventError{}
	for _, msg := range res.Peek {
		ts, err3 := core.ParseEpochMs(msg.Attributes["SentTimestamp"])
		if err3 != nil {
			log.Printf(" Error parsing timestamp %s: %v", msg.Attributes["SentTimestamp"], err3)
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
		cfg.Tx.Log("[RemoveTargets] No target to remove. returning")
		return nil
	}
	remTarInput := &cloudwatchevents.RemoveTargetsInput{
		Rule: aws.String(ruleName),
		Ids:  ids,
	}
	_, err = cfg.Svc.RemoveTargets(remTarInput)
	if err != nil {
		cfg.Tx.Log("[RemoveTargets] Error removing targets: %v", err)
		return err
	}
	sqsCfg := sqs.Init(cfg.Region)
	err = sqsCfg.DeleteByName(SQS_QUEUE_NAME_PRERFIX + ruleName)
	return err
}
