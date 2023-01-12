package customerprofiles

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	appflow "tah/core/appflow"
	core "tah/core/core"
	sqs "tah/core/sqs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	customerProfileSdk "github.com/aws/aws-sdk-go/service/customerprofiles"
)

var OBJECT_TYPE_NAME_ORDER = "_order"
var PROFILE_ID_KEY = "_profileId"

type CustomerProfileConfig struct {
	Client     *customerProfileSdk.CustomerProfiles
	DomainName string
	Region     string
	SQSClient  *sqs.Config
}

type Profile struct {
	ProfileId            string
	AccountNumber        string
	FirstName            string
	MiddleName           string
	LastName             string
	BirthDate            string
	Gender               string
	PhoneNumber          string
	MobilePhoneNumber    string
	BusinessPhoneNumber  string
	PersonalEmailAddress string
	BusinessEmailAddress string
	EmailAddress         string
	Attributes           map[string]string
	Address              Address
	Orders               []Order
	Matches              []Match
}

type Domain struct {
	Name            string
	NObjects        int64
	NProfiles       int64
	MatchingEnabled bool
	Created         time.Time
	LastUpdated     time.Time
}

type ProfileObject struct {
	ID          string
	Type        string
	JSONContent string
	Order       Order
}

type Order struct {
	OrderId               string
	CustomerEmail         string
	CustomerPhone         string
	CreatedDate           string
	UpdatedDate           string
	ProcessedDate         string
	ClosedDate            string
	CancelledDate         string
	CancelReason          string
	Name                  string
	AdditionalInformation string
	Gateway               string
	Status                string
	StatusCode            string
	StatusUrl             string
	CreditCardNumber      string
	CreditCardCompany     string
	FulfillmentStatus     string
	TotalPrice            string
	TotalTax              string
	TotalDiscounts        string
	TotalItemsPrice       string
	TotalShippingPrice    string
	TotalTipReceived      string
	Currency              string
	TotalWeight           string
	BillingAddress        string
	ShippingAddress       string
	OrderItems            string
	Attributes            map[string]string
}

type IngestionError struct {
	Reason  string
	Message string
}

type Address struct {
	Lines     []string `json:"lines"`     //unique identifier of the hotel
	City      string   `json:"city"`      //unique identifier of the hotel
	Country   string   `json:"country"`   //unique identifier of the hotel
	Latitude  float64  `json:"latitude"`  //unique identifier of the hotel
	Longitude float64  `json:"longitude"` //unique identifier of the hotel
	ZipCode   string   `json:"zipcode"`   //unique identifier of the hotel
}

type MatchList struct {
	ConfidenceScore float64
	ProfileIds      []string
}

type Match struct {
	ConfidenceScore float64
	ProfileID       string
	FirstName       string
	LastName        string
	BirthDate       string
	PhoneNumber     string
	EmailAddress    string
}
type ObjectMapping struct {
	Name   string         `json:"name"`
	Fields []FieldMapping `json:"fields"`
}

type FieldMapping struct {
	Type        string   `json:"type"`
	Source      string   `json:"source"`
	Target      string   `json:"target"`
	Indexes     []string `json:"indexes"`
	Searcheable bool     `json:"searchable"`
}

type Integration struct {
	Source         string
	Target         string
	Status         string
	StatusMessage  string
	LastRun        time.Time
	LastRunStatus  string
	LastRunMessage string
	Trigger        string
}

type Task struct {
	ConnectorOperator ConnectorOperator `json:"connectoroperator"`
	DestinationField  string            `json:"destinationfield"`
	SourceFields      []string          `json:"sourcefields"`
	TaskProperties    map[string]string `json:"taskproperties"`
	TaskType          string            `json:"tasktype"`
}

// Note: Only using S3 for this, but these other fields
// are present in the API so I am including them here
type ConnectorOperator struct {
	Marketo    string `json:"marketo"`
	S3         string `json:"s3"`
	Salesforce string `json:"salesforce"`
	ServiceNow string `json:"servicenow"`
	Zendesk    string `json:"zendesk"`
}

// Source Flows
type SourceFlowConfig struct {
	ConnectorType             string
	SourceConnectorProperties SourceConnectorProperties
}

type SourceConnectorProperties struct {
	S3 S3SourceProperties
}

type S3SourceProperties struct {
	BucketName   string
	BucketPrefix string
}

// Trigger Flows
type TriggerConfig struct {
	TriggerProperties TriggerProperties
	TriggerType       string
}

type TriggerProperties struct {
	Scheduled ScheduledTriggerProperties
}

type ScheduledTriggerProperties struct {
	ScheduleExpression string
	DataPullMode       string
	ScheduleStartTime  time.Time
}

type FlowDefinition struct {
	FlowName         string
	Description      string
	KmsArn           string
	SourceFlowConfig SourceFlowConfig
	TriggerConfig    TriggerConfig
	Tasks            []Task
}

func InitWithDomain(domainName string, region string) CustomerProfileConfig {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create DynamoDB client
	return CustomerProfileConfig{
		Client:     customerProfileSdk.New(session),
		DomainName: domainName,
		Region:     region,
	}

}

func Init(region string) CustomerProfileConfig {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create DynamoDB client
	return CustomerProfileConfig{
		Client: customerProfileSdk.New(session),
		Region: region,
	}

}

func (c CustomerProfileConfig) ListDomains() ([]Domain, error) {
	log.Printf("[core][customerProfiles] list customer profile domains")
	input := &customerProfileSdk.ListDomainsInput{}
	out, err := c.Client.ListDomains(input)
	if err != nil {
		return []Domain{}, err
	}
	domains := []Domain{}
	for _, dom := range out.Items {
		domains = append(domains, Domain{
			Name:        *dom.DomainName,
			Created:     *dom.CreatedAt,
			LastUpdated: *dom.LastUpdatedAt,
		})
	}
	return domains, nil

}

func (c *CustomerProfileConfig) CreateDomain(name string, idResolutionOn bool) error {
	sqsSvc := sqs.Init(c.Region)
	log.Printf("[core][customerProfiles] Creating new customer profile domain")
	log.Printf("[core][customerProfiles] 1-Creating SQS Queue")
	queueUrl, err := sqsSvc.Create("connect-profile-dlq-" + name)
	if err != nil {
		log.Printf("[core][customerProfiles] could not create dead letter queue")
		return err
	}
	log.Printf("[core][customerProfiles] 1-Set policy for Connect profile access")
	err = sqsSvc.SetPolicy("Service", "profile.amazonaws.com", []string{"SQS:*"})
	if err != nil {
		log.Printf("[core][customerProfiles] could not set policy on dead letter queue")
		return err
	}
	c.SQSClient = &sqsSvc
	input := &customerProfileSdk.CreateDomainInput{
		DomainName: aws.String(name),
		Matching: &customerProfileSdk.MatchingRequest{
			Enabled: aws.Bool(idResolutionOn),
		},
		DeadLetterQueueUrl:    aws.String(queueUrl),
		DefaultExpirationDays: aws.Int64(300),
	}
	_, err = c.Client.CreateDomain(input)
	if err == nil {
		c.DomainName = name
	}
	return err
}

func (c *CustomerProfileConfig) DeleteDomain() error {
	if c.DomainName == "" {
		return errors.New("Customer Profile Client not configured with domain name. Use DeleteDomainByName or assign domain name")
	}
	err := c.DeleteDomainByName(c.DomainName)
	if err == nil {
		c.DomainName = ""
	}
	return err
}

func (c CustomerProfileConfig) DeleteDomainByName(name string) error {
	log.Printf("[core][customerProfiles] Deleteing new customer profile domain")
	log.Printf("[core][customerProfiles] 1-deleting SQS Queue")
	var sqsClient sqs.Config
	if c.SQSClient == nil {
		out, err := c.Client.GetDomain(&customerProfileSdk.GetDomainInput{
			DomainName: aws.String(name),
		})
		if err != nil {
			log.Printf("[core][customerProfiles] Error getting domain for deletion")
			return err
		}
		if out.DeadLetterQueueUrl == nil {
			log.Printf("No Dead Letter Queue configured with Profile Domain. Nothing to delete")
		} else {
			sqsClient = sqs.InitWithQueueUrl(*out.DeadLetterQueueUrl, c.Region)
		}
	} else {
		sqsClient = *c.SQSClient
	}
	err := sqsClient.Delete()
	if err != nil {
		log.Printf("[core][customerProfiles] could not delete dead letter queue")
		return err
	}
	input := &customerProfileSdk.DeleteDomainInput{
		DomainName: aws.String(name),
	}
	_, err = c.Client.DeleteDomain(input)

	return err
}

func (c *CustomerProfileConfig) GetMappings() ([]ObjectMapping, error) {
	log.Printf("[core][customerProfiles][GetMappings] listing mappings")
	input := &customerProfileSdk.ListProfileObjectTypesInput{
		DomainName: aws.String(c.DomainName),
	}
	out, err := c.Client.ListProfileObjectTypes(input)
	if err != nil {
		return []ObjectMapping{}, err
	}
	mapping := []ObjectMapping{}
	log.Printf("[core][customerProfiles][GetMappings] ListProfileObjectTypes result: %+v", out)

	//TODO: parallelize this.
	for _, val := range out.Items {
		fMap, err2 := c.GetMapping(*val.ObjectTypeName)
		if err2 != nil {
			return []ObjectMapping{}, err2
		}
		mapping = append(mapping, fMap)
	}
	log.Printf("[core][customerProfiles][GetMappings] result: %+v", mapping)
	return mapping, err
}

func (c *CustomerProfileConfig) GetMapping(name string) (ObjectMapping, error) {
	input := &customerProfileSdk.GetProfileObjectTypeInput{
		DomainName:     aws.String(c.DomainName),
		ObjectTypeName: aws.String(name),
	}
	out, err := c.Client.GetProfileObjectType(input)
	if err != nil {
		return ObjectMapping{}, err
	}
	log.Printf("[core][customerProfiles][GetMapping] result from service: %+v", out)
	mapping := ObjectMapping{
		Name:   name,
		Fields: []FieldMapping{},
	}
	for _, val := range out.Fields {
		mapping.Fields = append(mapping.Fields, FieldMapping{
			Type:   *val.ContentType,
			Source: *val.Source,
			Target: *val.Target,
		})
	}
	return mapping, nil
}

func (c *CustomerProfileConfig) CreateMapping(name string, description string, fieldMappings []FieldMapping) error {
	input := &customerProfileSdk.PutProfileObjectTypeInput{
		AllowProfileCreation: aws.Bool(true),
		DomainName:           aws.String(c.DomainName),
		Description:          aws.String(description),
		ObjectTypeName:       aws.String(name),
		Fields:               toObjectTypeFieldMap(fieldMappings),
		Keys:                 toObjectTypeKeyMap(fieldMappings),
	}
	log.Printf("[core][customerProfiles][CreateMapping] Mapping creation request: %+v", input)
	_, err := c.Client.PutProfileObjectType(input)
	return err
}

func (c *CustomerProfileConfig) WaitForMappingCreation(name string) error {
	maxTries := 10
	try := 0
	for try < maxTries {
		mapping, err := c.GetMapping(name)
		if err == nil && mapping.Name == name {
			log.Printf("[core][customerProfiles][WaitForMappingCreation] Mapping creation successful")
			return nil
		}
		log.Printf("[core][customerProfiles][WaitForMappingCreation] Mapping not ready waiting 5 s")
		time.Sleep(5000)
		try += 1
	}
	return errors.New("Mapping creation seem to have failed")
}

func (c *CustomerProfileConfig) DeleteMapping(name string) error {
	input := &customerProfileSdk.DeleteProfileObjectTypeInput{
		DomainName:     aws.String(c.DomainName),
		ObjectTypeName: aws.String(name),
	}
	_, err := c.Client.DeleteProfileObjectType(input)
	return err
}

func toObjectTypeFieldMap(fieldMappings []FieldMapping) map[string]*customerProfileSdk.ObjectTypeField {
	result := map[string]*customerProfileSdk.ObjectTypeField{}
	for _, fm := range fieldMappings {
		keyName := buildMappingKey(fm)
		result[keyName] = &customerProfileSdk.ObjectTypeField{
			ContentType: aws.String(fm.Type),
			Source:      aws.String(fm.Source),
			Target:      aws.String(fm.Target),
		}
	}
	return result
}

func buildMappingKey(fieldMapping FieldMapping) string {
	str := strings.Replace(fieldMapping.Target, "_", "", -1)
	str = strings.Replace(str, ".", "", -1)
	return str
}
func buildSearchKey(fieldMapping FieldMapping) string {
	split := strings.Split(fieldMapping.Target, ".")
	return split[len(split)-1]
}

func toObjectTypeKeyMap(fieldMappings []FieldMapping) map[string][]*customerProfileSdk.ObjectTypeKey {
	result := map[string][]*customerProfileSdk.ObjectTypeKey{}
	for _, fm := range fieldMappings {
		keyName := buildMappingKey(fm)
		searchKeyName := buildSearchKey(fm)

		if len(fm.Indexes) > 0 {
			result[searchKeyName] = []*customerProfileSdk.ObjectTypeKey{
				&customerProfileSdk.ObjectTypeKey{
					FieldNames:          []*string{aws.String(keyName)},
					StandardIdentifiers: toPointerArray(fm.Indexes),
				},
			}
		} else if fm.Searcheable {
			result[searchKeyName] = []*customerProfileSdk.ObjectTypeKey{
				&customerProfileSdk.ObjectTypeKey{
					FieldNames: []*string{aws.String(keyName)},
				},
			}
		}
	}
	return result
}

func (c CustomerProfileConfig) GetDomain() (Domain, error) {
	log.Printf("[core][customerProfiles] Get Domain metadata")
	log.Printf("[core][customerProfiles] 1-find SQS queue")
	input := &customerProfileSdk.GetDomainInput{
		DomainName: aws.String(c.DomainName),
	}
	out, err := c.Client.GetDomain(input)
	if err != nil {
		log.Printf("[core][customerProfiles] error getting domain")
		return Domain{}, err
	}
	dom := Domain{
		Name: *out.DomainName,
	}
	if out.Matching != nil {
		dom.MatchingEnabled = toBool(out.Matching.Enabled)
	}
	if out.Stats != nil {
		dom.NProfiles = toInt64(out.Stats.ProfileCount)
		dom.NObjects = toInt64(out.Stats.ObjectCount)
	}
	return dom, nil

}

func (c CustomerProfileConfig) GetMatchesById(profileID string) ([]Match, error) {
	log.Printf("[core][customerProfiles][GetMatchesById] Getting matchine for ID %v", profileID)
	log.Printf("[core][customerProfiles][GetMatchesById] 1-getting all matches indexed by ID")
	index, err := c.IndexMatchesById()
	if err != nil {
		log.Printf("[core][customerProfiles] error indexing matches")
		return []Match{}, err
	}
	log.Printf("[core][customerProfiles][GetMatchesById] Found index: %v", index)
	log.Printf("[core][customerProfiles][GetMatchesById] 2-getting matches for ID %v from index", profileID)
	if matches, ok := index[profileID]; ok {
		log.Printf("[core][customerProfiles][GetMatchesById] found matches %v", matches)
		return matches, nil
	}
	return []Match{}, nil
}

// use only for POC. will need to ingest matches into dynamo page by page
func (c CustomerProfileConfig) IndexMatchesById() (map[string][]Match, error) {
	log.Printf("[core][customerProfiles][IndexMatchesById] 0-get matches")
	matches, err := c.GetMatches()
	if err != nil {
		log.Printf("[core][customerProfiles] error getting matches")
		return map[string][]Match{}, err
	}
	index := map[string][]Match{}
	log.Printf("[core][customerProfiles][IndexMatchesById] 1-indexing matches by ID")
	for _, matchList := range matches {
		for _, id := range matchList.ProfileIds {
			if _, ok := index[id]; !ok {
				index[id] = []Match{}
			}
			for _, id2 := range matchList.ProfileIds {
				if id2 != id && notIn(id2, index[id]) {
					index[id] = append(index[id], Match{
						ConfidenceScore: matchList.ConfidenceScore,
						ProfileID:       id2,
					})
				}
			}
		}
	}
	log.Printf("[core][customerProfiles][IndexMatchesById] 2- created index: %+v", index)
	return index, nil
}

func notIn(id string, matches []Match) bool {
	for _, match := range matches {
		if id == match.ProfileID {
			return false
		}
	}
	return true
}

func (c CustomerProfileConfig) GetIntegrations() ([]Integration, error) {
	log.Printf("[core][customerProfiles] Getting integrations for domain %s", c.DomainName)
	input := &customerProfileSdk.ListIntegrationsInput{
		DomainName: aws.String(c.DomainName),
	}
	log.Printf("[core][customerProfiles] 1-List integration")
	out, err := c.Client.ListIntegrations(input)
	if err != nil {
		log.Printf("[core][customerProfiles] error getting integration")
		return []Integration{}, err
	}
	log.Printf("[core][customerProfiles] List integration response: %+v", out)
	flowNames := []string{}
	for _, integration := range out.Items {
		if strings.HasPrefix(*integration.Uri, "arn:aws:appflow") {
			flowNames = append(flowNames, getFlowNameFromUri(*integration.Uri))
		}
	}
	apflowCfg := appflow.Init()
	log.Printf("[core][customerProfiles]2- Batch describe the following flows from AppFlow: %v", flowNames)
	flows, err2 := apflowCfg.GetFlows(flowNames)
	if err2 != nil {
		log.Printf("[core][customerProfiles] error getting flows")
		return []Integration{}, err2
	}
	log.Printf("[core][customerProfiles] Appflow response: %+v", flows)
	integrations := []Integration{}
	for _, flow := range flows {
		integrations = append(integrations, Integration{
			Source:         flow.SouceDetails,
			Target:         flow.TargetDetails,
			Status:         flow.Status,
			StatusMessage:  flow.StatusMessage,
			LastRun:        flow.LastRun,
			LastRunStatus:  flow.LastRunStatus,
			LastRunMessage: flow.LastRunMessage,
			Trigger:        flow.Trigger,
		})
	}
	log.Printf("[core][customerProfiles] Final integration response: %+v", integrations)
	return integrations, err
}

func (c CustomerProfileConfig) GetSourceTargetNames(objectname string) ([]string, []string, error) {

	mappings, err := c.GetMapping(objectname)
	if err != nil {
		log.Printf("Error getting mapping: %s", objectname)
		return []string{}, []string{}, err
	}
	objectFields := mappings.Fields
	fieldLength := len(objectFields)
	sourceNames := make([]string, fieldLength)
	targetNames := make([]string, fieldLength)
	for i, field := range objectFields {
		//source list
		source := field.Source
		sourceSplit := strings.Split(source, ".")
		sName := sourceSplit[len(sourceSplit)-1]
		sourceNames[i] = sName
		//target list
		target := field.Target
		targetSplit := strings.Split(target, ".")
		tName := targetSplit[len(targetSplit)-1]
		targetNames[i] = tName
	}
	return sourceNames, targetNames, err
}

func (c CustomerProfileConfig) GenerateTaskList(objectname string) ([]Task, error) {
	taskSlice := make([]Task, 0)
	conOpProj := ConnectorOperator{
		S3: "Projection",
	}
	sourceNames, targetNames, err := c.GetSourceTargetNames(objectname)
	if err != nil {
		log.Printf("Error getting Source and Target Lists for %s", objectname)
		return taskSlice, err
	}
	//Filter Task
	filTask := Task{
		TaskType:          "Filter",
		SourceFields:      sourceNames,
		ConnectorOperator: conOpProj,
	}
	taskSlice = append(taskSlice, filTask)
	conOpMap := ConnectorOperator{
		S3: "NO_OP",
	}
	//Mapping Tasks
	for i, source := range sourceNames {
		mapTask := Task{
			TaskType:          "Map",
			SourceFields:      []string{source},
			ConnectorOperator: conOpMap,
			DestinationField:  targetNames[i],
		}
		taskSlice = append(taskSlice, mapTask)
	}
	return taskSlice, err
}

func (c CustomerProfileConfig) GenerateFlowDefinition(objectname string, bucketname string, kmsarn string) (FlowDefinition, error) {
	taskList, err := c.GenerateTaskList(objectname)
	if err != nil {
		log.Printf("Error generating flow definition for %s", objectname)
		return FlowDefinition{}, err
	}
	sourceConfig := SourceFlowConfig{
		ConnectorType: "S3",
		SourceConnectorProperties: SourceConnectorProperties{
			S3: S3SourceProperties{
				BucketName:   bucketname,
				BucketPrefix: objectname,
			},
		},
	}
	trigConfig := TriggerConfig{
		TriggerProperties: TriggerProperties{
			Scheduled: ScheduledTriggerProperties{
				ScheduleExpression: "rate(1hours)",
				DataPullMode:       "Incremental",
				ScheduleStartTime:  time.Now(),
			},
		},
		TriggerType: "Scheduled",
	}
	flowDef := FlowDefinition{
		FlowName:         "Customer_Profile_Integration_Flowdefinition_" + objectname,
		Description:      "Customer Profile FlowDefinition for " + objectname,
		KmsArn:           kmsarn,
		TriggerConfig:    trigConfig,
		SourceFlowConfig: sourceConfig,
		Tasks:            taskList,
	}
	return flowDef, nil
}

func getFlowNameFromUri(uri string) string {
	split := strings.Split(uri, "/")
	return split[len(split)-1]
}

// use only for POC. will need to ingest matches into dynamo page by page
func (c CustomerProfileConfig) GetMatches() ([]MatchList, error) {
	log.Printf("[core][customerProfiles] Get Profiles matches from Identity resolution")
	input := &customerProfileSdk.GetMatchesInput{
		DomainName: aws.String(c.DomainName),
	}
	out, err := c.Client.GetMatches(input)
	if err != nil {
		log.Printf("[core][customerProfiles] error getting matches")
		return []MatchList{}, err
	}
	matches := []MatchList{}
	for _, item := range out.Matches {
		matches = append(matches, toMatchList(item))
	}
	for out.NextToken != nil {
		log.Printf("[core][customerProfiles] response is paginated. getting next bacth from token: %+v", out.NextToken)
		input.NextToken = out.NextToken
		out, err := c.Client.GetMatches(input)
		if err != nil {
			return matches, err
		}
		for _, item := range out.Matches {
			matches = append(matches, toMatchList(item))
		}
	}
	log.Printf("[core][customerProfiles] matches : %+v", matches)
	return matches, nil
}

func (c CustomerProfileConfig) GetErrors() ([]IngestionError, int64, error) {
	log.Printf("[core][customerProfiles] Get Errors from queue")
	log.Printf("[core][customerProfiles] 1-find SQS queue")
	input := &customerProfileSdk.GetDomainInput{
		DomainName: aws.String(c.DomainName),
	}
	out, err := c.Client.GetDomain(input)
	if err != nil {
		log.Printf("[core][customerProfiles] error getting domain")
		return []IngestionError{}, 0, err
	}
	if out.DeadLetterQueueUrl == nil {
		return []IngestionError{}, 0, errors.New("No Dead Letter Queue configured with Profile Domain " + c.DomainName)
	}
	log.Printf("[core][customerProfiles] found dead letter queue %+v", *out.DeadLetterQueueUrl)
	sqsSvc := sqs.InitWithQueueUrl(*out.DeadLetterQueueUrl, c.Region)
	//TODO: fix this. will not sacle. dump DQD content into Dynamo asyncroniously
	res, err2 := sqsSvc.Get()
	if err2 != nil {
		log.Printf("[core][customerProfiles] Error fetching from SQS queue %s: %v", *out.DeadLetterQueueUrl, err2)
		return []IngestionError{}, 0, err2
	}
	log.Printf("[core][customerProfiles] SQS response: %+v ", res)
	ingestionErrors := []IngestionError{}
	for _, msg := range res.Peek {
		ingestionErrors = append(ingestionErrors, IngestionError{
			Reason:  msg.MessageAttributes["Message"],
			Message: msg.Body,
		})
	}
	return ingestionErrors, res.NMessages, nil

}

// TODO: to decomision when service will support mor e than one value
// TODO: make this thread safe
func (c CustomerProfileConfig) SearchMultipleProfiles(key string, values []string) ([]Profile, error) {
	log.Printf("[core][customerProfiles][SearchMultipleProfiles] Parallel search for multiple profiles for %v in %v", key, values)
	var wg sync.WaitGroup
	wg.Add(len(values))
	allProfiles := []Profile{}
	var errs []error
	for i, value := range values {
		go func(val string, index int) {
			log.Printf("[core][customerProfiles[SearchMultipleProfiles]] Search %d started for value %v", index, val)
			profiles, err := c.SearchProfiles(key, []string{val})
			if err != nil {
				errs = append(errs, err)
			}
			for _, pro := range profiles {
				allProfiles = append(allProfiles, pro)
			}
			wg.Done()
		}(value, i)
	}
	wg.Wait()
	if len(errs) > 0 {
		log.Printf("[core][customerProfiles[SearchMultipleProfiles] At least one error occured during one of the parralel searches %v", errs)
		return []Profile{}, errs[0]
	}
	log.Printf("[core][customerProfiles[SearchMultipleProfiles]] Final search results %+v", allProfiles)
	return allProfiles, nil
}

func (c CustomerProfileConfig) SearchProfiles(key string, values []string) ([]Profile, error) {
	if len(values) > 1 {
		return []Profile{}, errors.New("Service only support one vaule for now")
	}
	log.Printf("[core][customerProfiles] Search profile for %v in %v", key, values)
	input := &customerProfileSdk.SearchProfilesInput{
		DomainName: aws.String(c.DomainName),
		KeyName:    aws.String(key),
		Values:     toPointerArray(values),
	}
	out, err := c.Client.SearchProfiles(input)
	log.Printf("[core][customerProfiles] Search response: %+v", out)
	profiles := []Profile{}
	if err != nil {
		return profiles, err
	}
	for _, item := range out.Items {
		profiles = append(profiles, toProfile(item))
	}
	for out.NextToken != nil {
		log.Printf("[core][customerProfiles] response is paginated. getting next bacth from token: %+v", out.NextToken)
		input.NextToken = out.NextToken
		out, err = c.Client.SearchProfiles(input)
		if err != nil {
			return profiles, err
		}
		for _, item := range out.Items {
			profiles = append(profiles, toProfile(item))
		}
	}
	log.Printf("[core][customerProfiles] Search Response after mapping : %+v", out)
	return profiles, nil
}

// TODO: parallelize the calls
func (c CustomerProfileConfig) GetProfile(id string) (Profile, error) {
	log.Printf("[core][customerProfiles][GetProfile] 0-retreiving profile with ID : %+v", id)
	log.Printf("[core][customerProfiles][GetProfile] 1-Search profile")
	res, err := c.SearchProfiles(PROFILE_ID_KEY, []string{id})
	if err != nil {
		return Profile{}, err
	}
	if len(res) == 0 {
		return Profile{}, errors.New("Profile with id " + id + "not found ")
	}
	p := res[0]
	log.Printf("[core][customerProfiles][GetProfile] 2-Get profile objects")
	objects, err2 := c.GetProfileObject(OBJECT_TYPE_NAME_ORDER, id)
	if err2 != nil {
		log.Printf("[core][customerProfiles] Error getting orders for profile %+v", id)
		return Profile{}, err2
	}
	for _, obj := range objects {
		p.Orders = append(p.Orders, obj.Order)
	}
	log.Printf("[core][customerProfiles][GetProfile] 3-Get profile matching")
	matches, err3 := c.GetMatchesById(id)
	if err3 != nil {
		//TODO: only continue if the erroorr code is ResourceNotFoundException: GetMatches result not found (mening the ID resolutioon has not run yet)
		log.Printf("[core][customerProfiles] Error getting profile matches for profile %+v: %+v", id, err3)
	} else {
		profileIds := []string{}
		for _, match := range matches {
			profileIds = append(profileIds, match.ProfileID)
		}
		log.Printf("[core][customerProfiles][GetProfile] 4-Retreive data for match profiles")
		profiles, err4 := c.SearchMultipleProfiles(PROFILE_ID_KEY, profileIds)
		if err4 != nil {
			log.Printf("[core][customerProfiles] Error getting matching profile details: %+v", err4)
			return Profile{}, err4
		}
		log.Printf("[core][customerProfiles][GetProfile] 5-Organize by ID")
		profilesById := map[string]Profile{}
		for _, profile := range profiles {
			profilesById[profile.ProfileId] = profile
		}
		log.Printf("[core][customerProfiles][GetProfile] 6-Enrich match")
		for i, match := range matches {
			profile := profilesById[match.ProfileID]
			matches[i].ProfileID = profile.ProfileId
			matches[i].FirstName = profile.FirstName
			matches[i].LastName = profile.LastName
			matches[i].BirthDate = profile.BirthDate
			matches[i].PhoneNumber = profile.PhoneNumber
			matches[i].EmailAddress = profile.EmailAddress
		}
		p.Matches = matches
	}
	log.Printf("[core][customerProfiles][GetProfile] 4-Final Profile: %+v", p)
	return p, nil
}

func (c CustomerProfileConfig) DeleteProfile(id string) error {
	log.Printf("[core][customerProfiles] Delete profile %v", id)
	input := &customerProfileSdk.DeleteProfileInput{
		DomainName: aws.String(c.DomainName),
		ProfileId:  aws.String(id),
	}
	out, err := c.Client.DeleteProfile(input)
	log.Printf("[core][customerProfiles] Delete profile response: %+v", out)
	return err
}

func (c CustomerProfileConfig) GetProfileObject(objectTypeName string, profileID string) ([]ProfileObject, error) {
	log.Printf("[core][customerProfiles] List objects of type %s, for profile %v", objectTypeName, profileID)
	input := &customerProfileSdk.ListProfileObjectsInput{
		DomainName:     aws.String(c.DomainName),
		ObjectTypeName: aws.String(objectTypeName),
		ProfileId:      aws.String(profileID),
	}
	out, err := c.Client.ListProfileObjects(input)
	log.Printf("[core][customerProfiles] Objects Search response: %+v", out)
	objects := []ProfileObject{}
	if err != nil {
		return objects, err
	}
	for _, item := range out.Items {
		objects = append(objects, toProfileObject(item))
	}
	for out.NextToken != nil {
		log.Printf("[core][customerProfiles] response is paginated. getting next batch from token: %+v", out.NextToken)
		input.NextToken = out.NextToken
		out, err = c.Client.ListProfileObjects(input)
		if err != nil {
			return objects, err
		}
		for _, item := range out.Items {
			objects = append(objects, toProfileObject(item))
		}
	}
	for i, order := range objects {
		//{\"OrderId\":\"464669299a74418a85f0e7b02cef7548\",\"CustomerEmail\":null,\"CustomerPhone\":null,\"CreatedDate\":null,\"UpdatedDate\":null,\"ProcessedDate\":null,\"ClosedDate\":null,\"CancelledDate\":null,\"CancelReason\":null,\"Name\":\"cloudrack_simulator\",\"AdditionalInformation\":null,\"Gateway\":null,\"Status\":\"confirmed\",\"StatusCode\":null,\"StatusUrl\":null,\"CreditCardNumber\":null,\"CreditCardCompany\":null,\"FulfillmentStatus\":null,\"TotalPrice\":\"700.0\",\"TotalTax\":null,\"TotalDiscounts\":null,\"TotalItemsPrice\":null,\"TotalShippingPrice\":null,\"TotalTipReceived\":null,\"Currency\":null,\"TotalWeight\":null,\"BillingAddress\":null,\"ShippingAddress\":null,\"OrderItems\":null,\"Attributes\":{\"confirmationNumber\":\"3P0SPWP4GL\",\"nNights\":\"2\",\"hotelCode\":\"2273445359\",\"nGuests\":\"1\",\"startDate\":\"20220815\",\"products\":\"DBL_SUITE-\"}}",
		if objectTypeName == OBJECT_TYPE_NAME_ORDER {
			log.Printf("[core][customerProfiles] Parse object json body %+v", order.JSONContent)
			objects[i].Order, err = parseOrderObject(order.JSONContent)
			if err != nil {
				log.Printf("[core][customerProfiles] Error Pasring order object: %+v", err)
				return []ProfileObject{}, err
			}
		}
	}
	log.Printf("[core][customerProfiles] Final Response : %+v", out)
	return objects, nil
}

func parseOrderObject(jsonObject string) (Order, error) {
	order := Order{}
	err := json.Unmarshal([]byte(jsonObject), &order)
	return order, err
}

func toMatchList(item *customerProfileSdk.MatchItem) MatchList {
	return MatchList{
		ConfidenceScore: tofloat64(item.ConfidenceScore),
		ProfileIds:      toStringArray(item.ProfileIds),
	}
}

func toProfile(item *customerProfileSdk.Profile) Profile {
	return Profile{
		ProfileId:            toString(item.ProfileId),
		AccountNumber:        toString(item.AccountNumber),
		FirstName:            toString(item.FirstName),
		MiddleName:           toString(item.MiddleName),
		LastName:             toString(item.LastName),
		BirthDate:            toString(item.BirthDate),
		Gender:               toString(item.Gender),
		PhoneNumber:          toString(item.PhoneNumber),
		MobilePhoneNumber:    toString(item.MobilePhoneNumber),
		BusinessPhoneNumber:  toString(item.BusinessPhoneNumber),
		EmailAddress:         toString(item.EmailAddress),
		PersonalEmailAddress: toString(item.PersonalEmailAddress),
		BusinessEmailAddress: toString(item.BusinessEmailAddress),
		Attributes:           core.ToMapString(item.Attributes),
	}
}

func toProfileObject(item *customerProfileSdk.ListProfileObjectsItem) ProfileObject {
	return ProfileObject{
		ID:          toString(item.ProfileObjectUniqueKey),
		Type:        toString(item.ObjectTypeName),
		JSONContent: toString(item.Object),
	}
}

func toString(in *string) string {
	if in != nil {
		return *in
	}
	return ""
}

func toInt64(in *int64) int64 {
	if in != nil {
		return *in
	}
	return 0
}
func tofloat64(in *float64) float64 {
	if in != nil {
		return *in
	}
	return float64(0.0)
}
func toBool(in *bool) bool {
	if in != nil {
		return *in
	}
	return false
}

func toPointerArray(in []string) []*string {
	out := []*string{}
	for _, val := range in {
		out = append(out, aws.String(val))
	}
	return out
}
func toStringArray(in []*string) []string {
	out := []string{}
	for _, val := range in {
		if val != nil {
			out = append(out, *val)
		}
	}
	return out
}
