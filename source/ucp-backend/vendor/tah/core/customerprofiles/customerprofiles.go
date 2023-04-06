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

var PROFILE_ID_KEY = "_profileId"

var STANDARD_IDENTIFIER_PROFILE = customerProfileSdk.StandardIdentifierProfile
var STANDARD_IDENTIFIER_ASSET = customerProfileSdk.StandardIdentifierAsset
var STANDARD_IDENTIFIER_CASE = customerProfileSdk.StandardIdentifierCase
var STANDARD_IDENTIFIER_UNIQUE = customerProfileSdk.StandardIdentifierUnique
var STANDARD_IDENTIFIER_SECONDARY = customerProfileSdk.StandardIdentifierSecondary
var STANDARD_IDENTIFIER_LOOKUP_ONLY = customerProfileSdk.StandardIdentifierLookupOnly
var STANDARD_IDENTIFIER_NEW_ONLY = customerProfileSdk.StandardIdentifierNewOnly
var STANDARD_IDENTIFIER_ORDER = customerProfileSdk.StandardIdentifierOrder

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
	ProfileObjects       []ProfileObject
	Matches              []Match
}

type Domain struct {
	Name                 string
	NObjects             int64
	NProfiles            int64
	MatchingEnabled      bool
	DefaultEncryptionKey string
	Created              time.Time
	LastUpdated          time.Time
	Tags                 map[string]string
}

type ProfileObject struct {
	ID          string
	Type        string
	JSONContent string
	Attributes  map[string]string
}

type IngestionError struct {
	Reason  string
	Message string
}

type Address struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	Address3   string `json:"address3"`
	Address4   string `json:"address4"`
	City       string `json:"city"`
	State      string `json:"state"`
	Province   string `json:"province"`
	PostalCode string `json:"postalcode"`
	Country    string `json:"country"`
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
	KeyOnly     bool     `json:"keyOnly"`
}

type FieldMappings []FieldMapping

type Integration struct {
	Name           string
	Source         string
	Target         string
	Status         string
	StatusMessage  string
	LastRun        time.Time
	LastRunStatus  string
	LastRunMessage string
	Trigger        string
}

func InitWithDomain(domainName string, region string) CustomerProfileConfig {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
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
			Tags:        core.ToMapString(dom.Tags),
		})
	}
	return domains, nil
}

func (c *CustomerProfileConfig) CreateDomain(name string, idResolutionOn bool, kmsArn string, tags map[string]string) error {
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
		DefaultEncryptionKey:  &kmsArn,
		Tags:                  core.ToMapPtString(tags),
	}
	_, err = c.Client.CreateDomain(input)
	if err == nil {
		c.DomainName = name
	}
	return err
}

func (c *CustomerProfileConfig) DeleteDomain() error {
	if c.DomainName == "" {
		return errors.New("customer rrofile client not configured with domain name. use deletedomainbyname or assign domain name")
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
			Type:   aws.StringValue(val.ContentType),
			Source: aws.StringValue(val.Source),
			Target: aws.StringValue(val.Target),
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
			Target:      setTarget(fm),
		}
	}
	return result
}

// When using custom object types, we need to create keys that don't map to
// a target field. This allows us to supply the source without trying to map.
func setTarget(fm FieldMapping) *string {
	if fm.KeyOnly {
		return nil
	} else {
		return aws.String(fm.Target)
	}
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
				{
					FieldNames:          []*string{aws.String(keyName)},
					StandardIdentifiers: aws.StringSlice(fm.Indexes),
				},
			}
		} else if fm.Searcheable {
			result[searchKeyName] = []*customerProfileSdk.ObjectTypeKey{
				{
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
		Name:                 *out.DomainName,
		DefaultEncryptionKey: *out.DefaultEncryptionKey,
		Tags:                 core.ToMapString(out.Tags),
	}
	if out.Matching != nil {
		dom.MatchingEnabled = aws.BoolValue(out.Matching.Enabled)
	}
	if out.Stats != nil {
		dom.NProfiles = aws.Int64Value(out.Stats.ProfileCount)
		dom.NObjects = aws.Int64Value(out.Stats.ObjectCount)
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
		log.Printf("[core][customerProfiles][GetMatchesById] found matches")
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

// Create AppFlow -> Amazon Connect integrations.
//
// There are two versions created:
//
// 1. Scheduled integration named with the prefix + "_Scheduled"
//
// 2. On demand integration named with the prefix + "_OnDemand"
//
// These integrations send all fields to Amazon Connect Customer Profile,
// then Customer Profile handles mapping the data to a given Object Type.
func (c CustomerProfileConfig) PutIntegration(flowNamePrefix, objectName, bucketName string, fieldMappings []FieldMapping) error {
	domain, err := c.GetDomain()
	if err != nil {
		return err
	}

	triggerConfig := customerProfileSdk.TriggerConfig{
		TriggerProperties: &customerProfileSdk.TriggerProperties{
			Scheduled: &customerProfileSdk.ScheduledTriggerProperties{
				ScheduleExpression: aws.String("rate(1hours)"),
				DataPullMode:       aws.String(customerProfileSdk.DataPullModeIncremental),
				ScheduleStartTime:  aws.Time(time.Now()),
			},
		},
		TriggerType: aws.String(customerProfileSdk.TriggerTypeScheduled),
	}
	sourceFlowConfig := customerProfileSdk.SourceFlowConfig{
		ConnectorType: aws.String(customerProfileSdk.SourceConnectorTypeS3),
		SourceConnectorProperties: &customerProfileSdk.SourceConnectorProperties{
			S3: &customerProfileSdk.S3SourceProperties{
				BucketName:   &bucketName,
				BucketPrefix: &objectName,
			},
		},
	}
	flowNameScheduled := flowNamePrefix + "_Scheduled"
	flowDefinition := customerProfileSdk.FlowDefinition{
		FlowName:         &flowNameScheduled,
		Description:      aws.String("Flow definition for " + objectName),
		KmsArn:           &domain.DefaultEncryptionKey,
		TriggerConfig:    &triggerConfig,
		SourceFlowConfig: &sourceFlowConfig,
		Tasks:            generateTaskList(fieldMappings),
	}
	scheduledInput := customerProfileSdk.PutIntegrationInput{
		DomainName:     &c.DomainName,
		FlowDefinition: &flowDefinition,
		ObjectTypeName: &objectName,
	}

	// Create scheduled flow
	_, err = c.Client.PutIntegration(&scheduledInput)
	if err != nil {
		return err
	}

	// Create identical on-demand flow, changing only the name and trigger type.
	flowNameOnDemand := flowNamePrefix + "_OnDemand"
	flowDefinition.SetFlowName(flowNameOnDemand)
	flowDefinition.SetTriggerConfig(&customerProfileSdk.TriggerConfig{
		TriggerType: aws.String(customerProfileSdk.TriggerTypeOnDemand),
	})
	onDemandInput := customerProfileSdk.PutIntegrationInput{
		DomainName:     &c.DomainName,
		FlowDefinition: &flowDefinition,
		ObjectTypeName: &objectName,
	}
	_, err = c.Client.PutIntegration(&onDemandInput)
	if err != nil {
		return err
	}

	// Start scheduled flow
	err = startFlow(flowNameScheduled)
	return err
}

// There are two types of tasks that must be created:
//
// 1 - Filter all source fields to be added to the AppFlow flow, meaning each
// field will be sent to Customer Profile
//
// 2 - Mapping specification for AppFlow. Since we handle the actual mapping in
// Customer Profile, we do a no op mapping here.
//
// See more: https://docs.aws.amazon.com/connect/latest/adminguide/customerprofiles-s3-integration.html
func generateTaskList(fieldMappings []FieldMapping) []*customerProfileSdk.Task {
	sourceFields := []string{}
	for _, v := range fieldMappings {
		// Field mapping source may have a structure like "_source.unique_id"
		// and we only want to take actual field name (eg. "unique_id")
		source := strings.Split(v.Source, ".")
		sourceFields = append(sourceFields, source[len(source)-1])
	}
	filterTask := customerProfileSdk.Task{
		ConnectorOperator: &customerProfileSdk.ConnectorOperator{
			S3: aws.String(customerProfileSdk.S3ConnectorOperatorProjection),
		},
		SourceFields: aws.StringSlice(sourceFields),
		TaskType:     aws.String(customerProfileSdk.TaskTypeFilter),
	}
	mapTasks := []customerProfileSdk.Task{}
	for _, v := range sourceFields {
		task := customerProfileSdk.Task{
			ConnectorOperator: &customerProfileSdk.ConnectorOperator{
				S3: aws.String(customerProfileSdk.S3ConnectorOperatorNoOp),
			},
			SourceFields:     []*string{aws.String(v)},
			DestinationField: aws.String(v),
			TaskType:         aws.String(customerProfileSdk.TaskTypeMap),
		}
		mapTasks = append(mapTasks, task)
	}
	tasks := []customerProfileSdk.Task{}
	tasks = append(tasks, filterTask)
	tasks = append(tasks, mapTasks...)
	taskPointers := make([]*customerProfileSdk.Task, len(tasks))
	for i := range tasks {
		taskPointers[i] = &tasks[i]
	}
	return taskPointers
}

func startFlow(flowName string) error {
	af := appflow.Init()
	_, err := af.StartFlow(flowName)
	return err
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
			Name:           flow.Name,
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

func (c CustomerProfileConfig) DeleteIntegration(uri string) (customerProfileSdk.DeleteIntegrationOutput, error) {
	OutIntegration, err := c.Client.DeleteIntegration(&customerProfileSdk.DeleteIntegrationInput{
		DomainName: &c.DomainName,
		Uri:        &uri,
	})
	if err != nil {
		log.Printf("Error deleting integration %s", err)
	}

	return *OutIntegration, err
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
			allProfiles = append(allProfiles, profiles...)
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
		return []Profile{}, errors.New("service only suppors one value for now")
	}
	log.Printf("[core][customerProfiles] Search profile for %v in %v", key, values)
	input := &customerProfileSdk.SearchProfilesInput{
		DomainName: aws.String(c.DomainName),
		KeyName:    aws.String(key),
		Values:     aws.StringSlice(values),
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
// Search for a profile by ProfileID, and return data for specified object types.
func (c CustomerProfileConfig) GetProfile(id string, objectTypeNames []string) (Profile, error) {
	log.Printf("[core][customerProfiles][GetProfile] 0-retreiving profile with ID : %+v", id)
	log.Printf("[core][customerProfiles][GetProfile] 1-Search profile")
	res, err := c.SearchProfiles(PROFILE_ID_KEY, []string{id})
	if err != nil {
		return Profile{}, err
	}
	if len(res) == 0 {
		return Profile{}, errors.New("Profile with id " + id + " not found ")
	}
	if len(res) > 1 {
		return Profile{}, errors.New("Multiple profiles found for ID " + id)
	}
	p := res[0]
	log.Printf("[core][customerProfiles][GetProfile] 2-Get profile objects")

	// Concurrently get data for all profile object types
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	errs := []error{}
	wg.Add(len(objectTypeNames))
	for _, objectType := range objectTypeNames {
		go func(objectType, id string) {
			obj, err := c.GetProfileObject(objectType, id)
			mu.Lock()
			if err != nil {
				errs = append(errs, err)
			} else {
				p.ProfileObjects = append(p.ProfileObjects, obj...)
			}
			mu.Unlock()
			wg.Done()
		}(objectType, id)
	}
	wg.Wait()
	// TODO: decide on error handling strategy
	if len(errs) > 0 {
		log.Printf("[core][customerprofiles][GetProfile] Errors while fetching profile objects: %v", errs)
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
		log.Printf("[core][customerProfiles] Parse object json body %+v", order.JSONContent)
		objects[i].Attributes, err = parseProfileObject(order.JSONContent)
		if err != nil {
			log.Printf("[core][customerProfiles] Error Pasring order object: %+v", err)
			return []ProfileObject{}, err
		}
	}
	log.Printf("[core][customerProfiles] Final Response : %+v", out)
	return objects, nil
}

func parseProfileObject(jsonObject string) (map[string]string, error) {
	attributes := make(map[string]string)
	err := json.Unmarshal([]byte(jsonObject), &attributes)
	return attributes, err
}

func toMatchList(item *customerProfileSdk.MatchItem) MatchList {
	return MatchList{
		ConfidenceScore: aws.Float64Value(item.ConfidenceScore),
		ProfileIds:      aws.StringValueSlice(item.ProfileIds),
	}
}

func toProfile(item *customerProfileSdk.Profile) Profile {
	profile := Profile{
		ProfileId:            aws.StringValue(item.ProfileId),
		AccountNumber:        aws.StringValue(item.AccountNumber),
		FirstName:            aws.StringValue(item.FirstName),
		MiddleName:           aws.StringValue(item.MiddleName),
		LastName:             aws.StringValue(item.LastName),
		BirthDate:            aws.StringValue(item.BirthDate),
		Gender:               aws.StringValue(item.Gender),
		PhoneNumber:          aws.StringValue(item.PhoneNumber),
		MobilePhoneNumber:    aws.StringValue(item.MobilePhoneNumber),
		BusinessPhoneNumber:  aws.StringValue(item.BusinessPhoneNumber),
		EmailAddress:         aws.StringValue(item.EmailAddress),
		PersonalEmailAddress: aws.StringValue(item.PersonalEmailAddress),
		BusinessEmailAddress: aws.StringValue(item.BusinessEmailAddress),
		Attributes:           core.ToMapString(item.Attributes),
	}
	// Safely access fields if Address is not nil
	if item.Address != nil {
		profile.Address.Address1 = aws.StringValue(item.Address.Address1)
		profile.Address.Address2 = aws.StringValue(item.Address.Address2)
		profile.Address.Address3 = aws.StringValue(item.Address.Address3)
		profile.Address.Address4 = aws.StringValue(item.Address.Address4)
		profile.Address.City = aws.StringValue(item.Address.City)
		profile.Address.State = aws.StringValue(item.Address.State)
		profile.Address.Province = aws.StringValue(item.Address.State)
		profile.Address.PostalCode = aws.StringValue(item.Address.PostalCode)
		profile.Address.Country = aws.StringValue(item.Address.Country)
	}
	return profile
}

func toProfileObject(item *customerProfileSdk.ListProfileObjectsItem) ProfileObject {
	return ProfileObject{
		ID:          aws.StringValue(item.ProfileObjectUniqueKey),
		Type:        aws.StringValue(item.ObjectTypeName),
		JSONContent: aws.StringValue(item.Object),
	}
}

func (c CustomerProfileConfig) PutProfileObject(object, objectTypeName string) error {
	log.Printf("[customerprofiles] Putting object type %v in domain %v", objectTypeName, c.DomainName)
	input := customerProfileSdk.PutProfileObjectInput{
		DomainName:     &c.DomainName,
		Object:         &object,
		ObjectTypeName: &objectTypeName,
	}
	out, err := c.Client.PutProfileObject(&input)
	if err != nil {
		log.Printf("[customerprofiles] Error putting object: %v", err)
		return err
	}
	log.Printf("[customerprofiles] Successfully put object of type %v. Unique key: %v", objectTypeName, out.ProfileObjectUniqueKey)
	return nil
}

func (c *CustomerProfileConfig) WaitForMappingCreation(name string) error {
	maxTries := 10
	try := 0
	for try < maxTries {
		mapping, err := c.GetMapping(name)
		if err == nil && mapping.Name == name {
			log.Printf("[CustomerProfiles][WaitForMappingCreation] Mapping creation successful")
			return nil
		}
		log.Printf("[CustomerProfiles][WaitForMappingCreation] Mapping not ready waiting 5 s")
		time.Sleep(5000)
		try += 1
	}
	return errors.New("creating profile object failed or is taking longer than usual")
}

func (c *CustomerProfileConfig) WaitForIntegrationCreation(name string) error {
	maxTries := 10
	try := 0
	for try < maxTries {
		integrations, err := c.GetIntegrations()
		if err == nil && containsIntegration(integrations, name) {
			log.Printf("[CustomerProfiles][WaitForIntegrationCreation] Integration creation successful")
			return nil
		}
		log.Printf("[CustomerProfiles][WaitForIntegrationCreation] Integration not ready waiting 5 s")
		time.Sleep(5000)
		try += 1
	}
	return errors.New("creating integration failed or is taking longer than usual")
}

func containsIntegration(integrations []Integration, expectedName string) bool {
	for _, v := range integrations {
		if v.Name == expectedName {
			return true
		}
	}
	return false
}

// Check if a profile exists
// TODO: to be refactored in an Exist function
// TODO: remove the hardcoded key for profile ID.
func (c *CustomerProfileConfig) GetProfileId(profileId string) (string, error) {
	log.Printf("[customerprofiles][GetProfileId] Searching profile with ID %v in domain %v", profileId, c.DomainName)
	input := customerProfileSdk.SearchProfilesInput{
		DomainName: &c.DomainName,
		KeyName:    aws.String("profile_id"),
		Values:     aws.StringSlice([]string{profileId}),
	}
	output, err := c.Client.SearchProfiles(&input)
	if err != nil {
		log.Printf("[customerprofiles][GetProfileId] Error searching for profile %v: %v", profileId, err)
		return "", err
	}
	if len(output.Items) > 1 {
		log.Printf("[customerprofiles][GetProfileId] Error: Found multiple profiles with same ID")
		return "", errors.New("multiple profiles found with same id")
	}
	if len(output.Items) == 0 {
		log.Printf("[customerprofiles][GetProfileId] No profile found with ID %v", profileId)
		return "", nil
	}
	log.Printf("[customerprofiles][GetProfileId] successfully found profile with ID %v", profileId)
	return *output.Items[0].ProfileId, nil
}

func (fms FieldMappings) GetSourceNames() []string {
	names := []string{}
	for _, v := range fms {
		split := strings.Split(v.Source, ".")
		names = append(names, split[len(split)-1])
	}
	return names
}
