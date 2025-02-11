// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofiles

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	appflow "tah/upt/source/tah-core/appflow"
	core "tah/upt/source/tah-core/core"
	sqs "tah/upt/source/tah-core/sqs"

	"maps"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"

	"github.com/aws/aws-sdk-go-v2/aws"
	mw "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	customerProfileSdk "github.com/aws/aws-sdk-go-v2/service/customerprofiles"
	"github.com/aws/aws-sdk-go-v2/service/customerprofiles/types"
	"github.com/aws/smithy-go/middleware"
)

var PROFILE_ID_KEY = "_profileId"

var MULTIPLE_OBJECT_ERROR = "multiple objects found"
var NO_PROFILE_OBJECT_ERROR = "object not found"

const PROFILE_FIELD_OBJECT_TYPE = "profileFieldOjectType"
const CONNECT_ID_ATTRIBUTE = "external_connect_id"
const UNIQUE_OBJECT_ATTRIBUTE = "external_unique_object_key"
const EXTENDED_DATA_ATTRIBUTE = "extended_data"
const TIMESTAMP_ATTRIBUTE = "timestamp"

var STANDARD_IDENTIFIER_PROFILE = types.StandardIdentifierProfile
var STANDARD_IDENTIFIER_ASSET = types.StandardIdentifierAsset
var STANDARD_IDENTIFIER_CASE = types.StandardIdentifierCase
var STANDARD_IDENTIFIER_UNIQUE = types.StandardIdentifierUnique
var STANDARD_IDENTIFIER_SECONDARY = types.StandardIdentifierSecondary
var STANDARD_IDENTIFIER_LOOKUP_ONLY = types.StandardIdentifierLookupOnly
var STANDARD_IDENTIFIER_NEW_ONLY = types.StandardIdentifierNewOnly
var STANDARD_IDENTIFIER_ORDER = types.StandardIdentifierOrder

type ICustomerProfileConfig interface {
	CreateMapping(name string, description string, fieldMappings []FieldMapping) error
	SetTx(tx core.Transaction)
	CreateDomainWithQueue(
		name string,
		kmsArn string,
		tags map[string]string,
		queueUrl string,
		matchS3Name string,
		options DomainOptions,
	) error
	ListDomains() ([]Domain, error)
	DeleteDomain() error
	GetProfile(id string, objectTypeNames []string, pagination []PaginationOptions) (profilemodel.Profile, error)
	SearchProfiles(key string, values []string) ([]profilemodel.Profile, error)
	DeleteProfile(id string, options ...DeleteProfileOptions) error
	PutIntegration(
		flowNamePrefix, objectName, bucketName, bucketPrefix string,
		fieldMappings []FieldMapping,
		startTime time.Time,
	) error
	GetDomain() (Domain, error)
	GetMappings() ([]ObjectMapping, error)
	GetIntegrations() ([]Integration, error)
	DeleteDomainByName(name string) error
	MergeProfiles(profileId string, profileIdToMerge string) (string, error)
	MergeMany(pairs []ProfilePair) (string, error)
	SetDomain(domain string)
	PutProfileObject(object, objectTypeName string) error
	PutProfileAsObject(p profilemodel.Profile) error
	PutProfileObjectFromLcs(p profilemodel.ProfileObject, lcsId string) error
	IsThrottlingError(err error) bool
	GetProfileId(profileKey, profileId string) (string, error)
	GetSpecificProfileObject(
		objectKey string,
		objectId string,
		profileId string,
		objectTypeName string,
	) (profilemodel.ProfileObject, error)
	CreateEventStream(domainName, streamName, streamArn string) error
	DeleteEventStream(domainName, streamName string) error
}

type CustomerProfileConfig struct {
	Client          *customerProfileSdk.Client
	DomainName      string
	Region          string
	SQSClient       *sqs.Config
	Tx              core.Transaction
	solutionId      string
	solutionVersion string
}

// Creates the mapping for Customer Profiles.
//
// Target fields are currently only supported for profile/order/asset/case
// types. We use custom objects, which just send/receive json strings without
// any mappings. However, we still need to create a mapping for object keys.
// This is done by creating the mapping and setting KeyOnly to true.
// See TestCustomerProfiles for implementation example.

// Provides static check for the interface implementation without allocating memory at runtime
var _ ICustomerProfileConfig = &CustomerProfileConfig{}

func InitWithDomain(domainName, region, solutionId, solutionVersion string) *CustomerProfileConfig {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithAPIOptions(
			[]func(*middleware.Stack) error{mw.AddUserAgentKey("AWSSOLUTION/" + solutionId + "/" + solutionVersion)},
		),
	)
	if err != nil {
		panic("error initializing customerprofiles client")
	}
	return &CustomerProfileConfig{
		Client:          customerProfileSdk.NewFromConfig(cfg),
		DomainName:      domainName,
		Region:          region,
		solutionId:      solutionId,
		solutionVersion: solutionVersion,
	}
}

func Init(region, solutionId, solutionVersion string) *CustomerProfileConfig {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithAPIOptions(
			[]func(*middleware.Stack) error{mw.AddUserAgentKey("AWSSOLUTION/" + solutionId + "/" + solutionVersion)},
		),
	)
	if err != nil {
		panic("error initializing customerprofiles client")
	}
	return &CustomerProfileConfig{
		Client:          customerProfileSdk.NewFromConfig(cfg),
		Region:          region,
		solutionId:      solutionId,
		solutionVersion: solutionVersion,
	}
}

func (c *CustomerProfileConfig) SetTx(tx core.Transaction) {
	tx.LogPrefix = "customerprofiles"
	c.Tx = tx
}

func (c *CustomerProfileConfig) SetDomain(domain string) {
	c.DomainName = domain
}

func (c CustomerProfileConfig) IsThrottlingError(err error) bool {
	var throttlingException *types.ThrottlingException
	return err != nil && errors.As(err, &throttlingException)
}

func (c CustomerProfileConfig) ListDomains() ([]Domain, error) {
	c.Tx.Debug("List customer profile domains")
	input := &customerProfileSdk.ListDomainsInput{}
	out, err := c.Client.ListDomains(context.TODO(), input)
	if err != nil {
		return []Domain{}, err
	}
	domains := []Domain{}
	for _, dom := range out.Items {
		domains = append(domains, Domain{
			Name:        *dom.DomainName,
			Created:     *dom.CreatedAt,
			LastUpdated: *dom.LastUpdatedAt,
			Tags:        dom.Tags,
		})
	}
	return domains, nil
}

func (c *CustomerProfileConfig) CreateDomainWithQueue(
	name string,
	kmsArn string,
	tags map[string]string,
	queueUrl string,
	matchS3Name string,
	options DomainOptions,
) error {

	input := &customerProfileSdk.CreateDomainInput{}
	if options.AiIdResolutionOn {
		input = &customerProfileSdk.CreateDomainInput{
			DomainName: aws.String(name),
			Matching: &types.MatchingRequest{
				Enabled: aws.Bool(options.AiIdResolutionOn),
				ExportingConfig: &types.ExportingConfig{
					S3Exporting: &types.S3ExportingConfig{
						S3BucketName: aws.String(matchS3Name),
						S3KeyName:    aws.String("Match"),
					},
				},
			},
			DefaultExpirationDays: aws.Int32(300),
			DefaultEncryptionKey:  &kmsArn,
			Tags:                  tags,
		}
	} else {
		input = &customerProfileSdk.CreateDomainInput{
			DomainName:            aws.String(name),
			DefaultExpirationDays: aws.Int32(300),
			DefaultEncryptionKey:  &kmsArn,
			Tags:                  tags,
		}
	}

	if queueUrl != "" {
		qClient := sqs.InitWithQueueUrl(queueUrl, c.Region, c.solutionId, c.solutionVersion)
		c.SQSClient = &qClient
		input.DeadLetterQueueUrl = aws.String(queueUrl)
	}

	_, err := c.Client.CreateDomain(context.TODO(), input)
	if err == nil {
		c.DomainName = name
	}
	return err
}

func (c *CustomerProfileConfig) CreateDomain(
	name string,
	kmsArn string,
	tags map[string]string,
	matchS3Name string,
	options DomainOptions,
) error {
	c.Tx.Info(" Creating new customer profile domain")
	sqsSvc := sqs.Init(c.Region, c.solutionId, c.solutionVersion)
	c.Tx.Info(" 1-Creating SQS Queue")
	queueUrl, err := sqsSvc.Create("connect-profile-dlq-" + name)
	if err != nil {
		c.Tx.Error(" could not create dead letter queue")
		return err
	}
	c.Tx.Info(" 1-Set policy for Connect profile access")
	err = sqsSvc.SetPolicy("Service", "profile.amazonaws.com", []string{"SQS:*"})
	if err != nil {
		c.Tx.Error(" could not set policy on dead letter queue")
		return err
	}
	return c.CreateDomainWithQueue(name, kmsArn, tags, queueUrl, matchS3Name, options)
}

func (c *CustomerProfileConfig) DeleteDomain() error {
	if c.DomainName == "" {
		return errors.New(
			"customer profile client not configured with domain name. use deletedomainbyname or assign domain name",
		)
	}
	err := c.DeleteDomainByName(c.DomainName)
	if err == nil {
		c.DomainName = ""
	}
	return err
}

func (c CustomerProfileConfig) DeleteDomainByName(name string) error {
	c.Tx.Info(" Deleteing new customer profile domain")
	input := &customerProfileSdk.DeleteDomainInput{
		DomainName: aws.String(name),
	}
	_, err := c.Client.DeleteDomain(context.TODO(), input)
	return err
}

func (c CustomerProfileConfig) DeleteDomainAndQueueByName(name string) error {
	c.Tx.Info(" Deleteing new customer profile domain")
	c.Tx.Info(" 1-deleting SQS Queue")
	var sqsClient sqs.Config
	if c.SQSClient == nil {
		out, err := c.Client.GetDomain(context.TODO(), &customerProfileSdk.GetDomainInput{
			DomainName: aws.String(name),
		})
		if err != nil {
			c.Tx.Error(" Error getting domain for deletion")
			return err
		}
		if out.DeadLetterQueueUrl != nil {
			sqsClient = sqs.InitWithQueueUrl(*out.DeadLetterQueueUrl, c.Region, c.solutionId, c.solutionVersion)
			err := sqsClient.Delete()
			if err != nil {
				c.Tx.Error(" could not delete dead letter queue")
				return err
			}
		} else {
			c.Tx.Info("No Dead Letter Queue configured with Profile Domain. Nothing to delete")
		}
	} else {
		err := (*c.SQSClient).Delete()
		if err != nil {
			c.Tx.Error(" could not delete dead letter queue")
			return err
		}
	}
	input := &customerProfileSdk.DeleteDomainInput{
		DomainName: aws.String(name),
	}
	_, err := c.Client.DeleteDomain(context.TODO(), input)
	return err
}

func (c *CustomerProfileConfig) GetMappings() ([]ObjectMapping, error) {
	c.Tx.Debug("[GetMappings] listing mappings")
	input := &customerProfileSdk.ListProfileObjectTypesInput{
		DomainName: aws.String(c.DomainName),
	}
	out, err := c.Client.ListProfileObjectTypes(context.TODO(), input)
	if err != nil {
		return []ObjectMapping{}, err
	}
	mapping := []ObjectMapping{}
	c.Tx.Debug("[GetMappings] ListProfileObjectTypes result: %+v", out)

	//TODO: parallelize this.
	for _, val := range out.Items {
		fMap, err2 := c.GetMapping(*val.ObjectTypeName)
		if err2 != nil {
			return []ObjectMapping{}, err2
		}
		mapping = append(mapping, fMap)
	}
	c.Tx.Info("[GetMappings] result: %+v", mapping)
	return mapping, err
}

func (c *CustomerProfileConfig) GetMapping(name string) (ObjectMapping, error) {
	input := &customerProfileSdk.GetProfileObjectTypeInput{
		DomainName:     aws.String(c.DomainName),
		ObjectTypeName: aws.String(name),
	}
	out, err := c.Client.GetProfileObjectType(context.TODO(), input)
	if err != nil {
		return ObjectMapping{}, err
	}
	c.Tx.Debug("[GetMapping] found mapping with %v fields", len(out.Fields))
	mapping := ObjectMapping{
		Name:   name,
		Fields: []FieldMapping{},
	}
	c.Tx.Debug("[GetMapping] profile output: %+v", out)
	for _, val := range out.Fields {
		fMapping := FieldMapping{
			Type:   string(val.ContentType),
			Source: aws.ToString(val.Source),
			Target: aws.ToString(val.Target),
		}

		keys := out.Keys[buildSearchKey(fMapping)]
		if len(keys) == 0 {
			//key only mapping
			keys = out.Keys[buildSearchString(fMapping.Source)]
		}
		if len(keys) > 0 {
			c.Tx.Debug("[GetMapping] found key %v", keys)
			if len(keys[0].StandardIdentifiers) > 0 {
				for _, standardIdentifier := range keys[0].StandardIdentifiers {
					fMapping.Indexes = append(fMapping.Indexes, string(standardIdentifier))
				}
			} else {
				fMapping.Searcheable = true
			}
		}

		if fMapping.Target == "" {
			fMapping.KeyOnly = true
		}
		mapping.Fields = append(mapping.Fields, fMapping)
	}
	return mapping, nil
}

func (c *CustomerProfileConfig) CreateMapping(name string, description string, fieldMappings []FieldMapping) error {
	c.Tx.Info(
		"[CreateMapping] Creating mapping  %s (%s) for domain %s and feilds %+v",
		name,
		description,
		c.DomainName,
		fieldMappings,
	)
	timeStampFormat := "EPOCHMILLI"
	fieldMap := toObjectTypeFieldMap(fieldMappings)
	keyMap := toObjectTypeKeyMap(fieldMappings)
	input := &customerProfileSdk.PutProfileObjectTypeInput{
		AllowProfileCreation:             true,
		DomainName:                       aws.String(c.DomainName),
		Description:                      aws.String(description),
		ObjectTypeName:                   aws.String(name),
		SourceLastUpdatedTimestampFormat: &timeStampFormat,
		Fields:                           fieldMap,
		Keys:                             keyMap,
	}
	c.Tx.Debug("[CreateMapping] PutProfileObjectTypeInput: %+v", input)
	_, err := c.Client.PutProfileObjectType(context.TODO(), input, func(options *customerProfileSdk.Options) {
		// as a control plane operation, we are willing to wait longer for this operation to succeed
		if options != nil && options.RetryMaxAttempts < 10 {
			options.RetryMaxAttempts = 10
		}
	})
	return err
}

func (c *CustomerProfileConfig) DeleteMapping(name string) error {
	input := &customerProfileSdk.DeleteProfileObjectTypeInput{
		DomainName:     aws.String(c.DomainName),
		ObjectTypeName: aws.String(name),
	}
	_, err := c.Client.DeleteProfileObjectType(context.TODO(), input)
	return err
}

func toObjectTypeFieldMap(fieldMappings []FieldMapping) map[string]types.ObjectTypeField {
	result := map[string]types.ObjectTypeField{}
	for _, fm := range fieldMappings {
		keyName := buildMappingKey(fm)
		result[keyName] = types.ObjectTypeField{
			ContentType: typeToAccpType(fm.Type),
			Source:      aws.String(fm.Source),
			Target:      setTarget(fm),
		}
	}
	return result
}

// LCS may support additional data types not available with ACCP. This converts unsupported field types
// back into ACCP field types.
//
// Example: LCS "string" defaults to VARCHAR(255), LCS "text" uses TEXT instead.
// For ACCP, both would just use "string".
func typeToAccpType(contentType string) types.FieldContentType {
	if contentType == MappingTypeText {
		contentType = MappingTypeString
	}
	return types.FieldContentType(contentType)
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
	return buildMappingString(fieldMapping.Target)
}
func buildMappingString(inStr string) string {
	str := strings.Replace(inStr, "_", "", -1)
	str = strings.Replace(str, ".", "", -1)
	return str
}
func buildSearchKey(fieldMapping FieldMapping) string {
	return buildSearchString(fieldMapping.Target)
}
func buildSearchString(inStr string) string {
	split := strings.Split(inStr, ".")
	return split[len(split)-1]
}

func toObjectTypeKeyMap(fieldMappings []FieldMapping) map[string][]types.ObjectTypeKey {
	result := map[string][]types.ObjectTypeKey{}
	for _, fm := range fieldMappings {
		keyName := buildMappingKey(fm)
		searchKeyName := buildSearchKey(fm)

		var indexes []types.StandardIdentifier
		for _, index := range fm.Indexes {
			indexes = append(indexes, types.StandardIdentifier(index))
		}

		if len(fm.Indexes) > 0 {
			result[searchKeyName] = []types.ObjectTypeKey{
				{
					FieldNames:          []string{keyName},
					StandardIdentifiers: indexes,
				},
			}
		} else if fm.Searcheable {
			result[searchKeyName] = []types.ObjectTypeKey{
				{
					FieldNames: []string{keyName},
				},
			}
		}
	}
	return result
}

func (c CustomerProfileConfig) DomainExists(domainName string) (bool, error) {
	input := &customerProfileSdk.GetDomainInput{
		DomainName: aws.String(domainName),
	}
	out, err := c.Client.GetDomain(context.TODO(), input)
	if err != nil {
		c.Tx.Error(" error getting domain")
		return false, err
	}
	if out.DomainName != nil && *out.DomainName == domainName {
		return true, nil
	}
	return false, nil
}

func (c CustomerProfileConfig) GetDomain() (Domain, error) {
	c.Tx.Info(" Get Domain metadata")
	c.Tx.Debug(" 1-find SQS queue")
	input := &customerProfileSdk.GetDomainInput{
		DomainName: aws.String(c.DomainName),
	}
	out, err := c.Client.GetDomain(context.TODO(), input)
	if err != nil {
		c.Tx.Error(" error getting domain")
		return Domain{}, err
	}
	dom := Domain{
		Name:                 *out.DomainName,
		DefaultEncryptionKey: *out.DefaultEncryptionKey,
		Tags:                 out.Tags,
	}
	if out.Matching != nil {
		dom.MatchingEnabled = aws.ToBool(out.Matching.Enabled)
	}
	if out.Stats != nil {
		dom.NProfiles = out.Stats.ProfileCount
		dom.NObjects = out.Stats.ObjectCount
	}
	dom.IsLowCostEnabled = false
	return dom, nil
}

func (c CustomerProfileConfig) GetMatchesById(profileID string) ([]profilemodel.Match, error) {
	c.Tx.Info("[GetMatchesById] Getting matchine for ID %v", profileID)
	c.Tx.Debug("[GetMatchesById] 1-getting all matches indexed by ID")
	index, err := c.IndexMatchesById()
	if err != nil {
		c.Tx.Error(" error indexing matches")
		return []profilemodel.Match{}, err
	}
	c.Tx.Debug("[GetMatchesById] Found index: %v", index)
	c.Tx.Debug("[GetMatchesById] 2-getting matches for ID %v from index", profileID)
	if matches, ok := index[profileID]; ok {
		c.Tx.Info("[GetMatchesById] found matches")
		return matches, nil
	}
	return []profilemodel.Match{}, nil
}

// use only for POC. will need to ingest matches into dynamo page by page
func (c CustomerProfileConfig) IndexMatchesById() (map[string][]profilemodel.Match, error) {
	c.Tx.Debug("[IndexMatchesById] 0-get matches")
	matches, err := c.GetMatches()
	if err != nil {
		c.Tx.Error(" error getting matches")
		return map[string][]profilemodel.Match{}, err
	}
	index := map[string][]profilemodel.Match{}
	c.Tx.Debug("[IndexMatchesById] 1-indexing matches by ID")
	for _, matchList := range matches {
		for _, id := range matchList.ProfileIds {
			if _, ok := index[id]; !ok {
				index[id] = []profilemodel.Match{}
			}
			for _, id2 := range matchList.ProfileIds {
				if id2 != id && notIn(id2, index[id]) {
					index[id] = append(index[id], profilemodel.Match{
						ConfidenceScore: matchList.ConfidenceScore,
						ProfileID:       id2,
					})
				}
			}
		}
	}
	c.Tx.Debug("[IndexMatchesById] 2- created index: %+v", index)
	return index, nil
}

func notIn(id string, matches []profilemodel.Match) bool {
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
func (c CustomerProfileConfig) PutIntegration(
	flowNamePrefix, objectName, bucketName, bucketPrefix string,
	fieldMappings []FieldMapping,
	startTime time.Time,
) error {
	domain, err := c.GetDomain()
	if err != nil {
		c.Tx.Error("[PutIntegration] Error getting domain: %v", err)
		return err
	}

	triggerConfig := types.TriggerConfig{
		TriggerProperties: &types.TriggerProperties{
			Scheduled: &types.ScheduledTriggerProperties{
				ScheduleExpression: aws.String("rate(1hours)"),
				DataPullMode:       types.DataPullModeIncremental,
				ScheduleStartTime:  aws.Time(startTime),
			},
		},
		TriggerType: types.TriggerTypeScheduled,
	}
	sourceFlowConfig := types.SourceFlowConfig{
		ConnectorType: types.SourceConnectorTypeS3,
		SourceConnectorProperties: &types.SourceConnectorProperties{
			S3: &types.S3SourceProperties{
				BucketName: aws.String(bucketName),
			},
		},
	}
	if bucketPrefix != "" {
		sourceFlowConfig.SourceConnectorProperties.S3.BucketPrefix = aws.String(bucketPrefix)
	}
	flowNameScheduled := flowNamePrefix + "_Scheduled"
	flowDefinition := types.FlowDefinition{
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
	_, err = c.Client.PutIntegration(context.TODO(), &scheduledInput)
	if err != nil {
		c.Tx.Error("[PutIntegration] Error Create scheduled flow: %v", err)
		return err
	}

	// Create identical on-demand flow, changing only the name and trigger type.
	flowNameOnDemand := flowNamePrefix + "_OnDemand"
	flowDefinition.FlowName = &flowNameOnDemand
	flowDefinition.TriggerConfig = &types.TriggerConfig{
		TriggerType: types.TriggerTypeOndemand,
	}
	onDemandInput := customerProfileSdk.PutIntegrationInput{
		DomainName:     &c.DomainName,
		FlowDefinition: &flowDefinition,
		ObjectTypeName: &objectName,
	}
	_, err = c.Client.PutIntegration(context.TODO(), &onDemandInput)
	if err != nil {
		c.Tx.Error("[PutIntegration] Error Create identical on-demand flow: %v", err)
		return err
	}

	// Start scheduled flow
	err = c.startFlow(flowNameScheduled)
	if err != nil {
		c.Tx.Error("[PutIntegration] Error Start scheduled flow %v", err)
		return err
	}
	return nil
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
func generateTaskList(fieldMappings []FieldMapping) []types.Task {
	sourceFields := []string{}
	for _, v := range fieldMappings {
		// Field mapping source may have a structure like "_source.unique_id"
		// and we only want to take actual field name (eg. "unique_id")
		source := strings.Split(v.Source, ".")
		sourceFields = append(sourceFields, source[len(source)-1])
	}
	filterTask := types.Task{
		ConnectorOperator: &types.ConnectorOperator{
			S3: types.S3ConnectorOperatorProjection,
		},
		SourceFields: sourceFields,
		TaskType:     types.TaskTypeFilter,
	}
	mapTasks := []types.Task{}
	for _, v := range sourceFields {
		task := types.Task{
			ConnectorOperator: &types.ConnectorOperator{
				S3: types.S3ConnectorOperatorNoOp,
			},
			SourceFields:     []string{v},
			DestinationField: aws.String(v),
			TaskType:         types.TaskTypeMap,
		}
		mapTasks = append(mapTasks, task)
	}
	tasks := []types.Task{}
	tasks = append(tasks, filterTask)
	tasks = append(tasks, mapTasks...)
	return tasks
}

func (c CustomerProfileConfig) startFlow(flowName string) error {
	af := appflow.Init(c.solutionId, c.solutionVersion)
	_, err := af.StartFlow(flowName)
	return err
}

func (c CustomerProfileConfig) GetIntegrations() ([]Integration, error) {
	c.Tx.Info("[GetIntegrations] Getting integrations for domain %s", c.DomainName)
	input := &customerProfileSdk.ListIntegrationsInput{
		DomainName: aws.String(c.DomainName),
	}
	c.Tx.Debug("[GetIntegrations] 1-List integration")
	out, err := c.Client.ListIntegrations(context.TODO(), input)
	if err != nil {
		c.Tx.Error(" error getting integration")
		return []Integration{}, err
	}
	c.Tx.Debug("[GetIntegrations] List integration found %v integrations", len(out.Items))
	flowNames := []string{}
	for _, integration := range out.Items {
		if strings.HasPrefix(*integration.Uri, "arn:aws:appflow") {
			flowNames = append(flowNames, getFlowNameFromUri(*integration.Uri))
		}
	}
	apflowCfg := appflow.Init(c.solutionId, c.solutionVersion)
	c.Tx.Debug("[GetIntegrations] 2- Batch describe the following flows from AppFlow: %+v", flowNames)
	flows, err := apflowCfg.GetFlows(flowNames)
	if err != nil {
		c.Tx.Error(" error getting flows")
		return []Integration{}, err
	}
	c.Tx.Debug(" Appflow response found %v flows", len(flows))
	integrations := []Integration{}
	for _, flow := range flows {
		integrations = append(integrations, Integration{
			Name:           flow.Name,
			Source:         flow.SourceDetails,
			Target:         flow.TargetDetails,
			Status:         flow.Status,
			StatusMessage:  flow.StatusMessage,
			LastRun:        flow.LastRun,
			LastRunStatus:  flow.LastRunStatus,
			LastRunMessage: flow.LastRunMessage,
			Trigger:        flow.Trigger,
		})
	}
	return integrations, err
}

func (c CustomerProfileConfig) DeleteIntegration(uri string) (customerProfileSdk.DeleteIntegrationOutput, error) {
	OutIntegration, err := c.Client.DeleteIntegration(context.TODO(), &customerProfileSdk.DeleteIntegrationInput{
		DomainName: &c.DomainName,
		Uri:        &uri,
	})
	if err != nil {
		c.Tx.Error("Error deleting integration %s", err)
	}

	return *OutIntegration, err
}

func getFlowNameFromUri(uri string) string {
	split := strings.Split(uri, "/")
	return split[len(split)-1]
}

// use only for POC. will need to ingest matches into dynamo page by page
func (c CustomerProfileConfig) GetMatches() ([]MatchList, error) {
	c.Tx.Info(" Get Profiles matches from Identity resolution")
	input := &customerProfileSdk.GetMatchesInput{
		DomainName: aws.String(c.DomainName),
	}
	out, err := c.Client.GetMatches(context.TODO(), input)
	if err != nil {
		c.Tx.Error(" error getting matches")
		return []MatchList{}, err
	}
	matches := []MatchList{}
	for _, item := range out.Matches {
		matches = append(matches, toMatchList(&item))
	}
	for out.NextToken != nil {
		c.Tx.Debug(" response is paginated. getting next bacth from token: %+v", out.NextToken)
		input.NextToken = out.NextToken
		out, err := c.Client.GetMatches(context.TODO(), input)
		if err != nil {
			return matches, err
		}
		for _, item := range out.Matches {
			matches = append(matches, toMatchList(&item))
		}
	}
	c.Tx.Info(" matches : %+v", matches)
	return matches, nil
}

func (c CustomerProfileConfig) GetErrors() ([]IngestionError, int64, error) {
	c.Tx.Info(" Get Errors from queue")
	c.Tx.Debug(" 1-find SQS queue")
	input := &customerProfileSdk.GetDomainInput{
		DomainName: aws.String(c.DomainName),
	}
	out, err := c.Client.GetDomain(context.TODO(), input)
	if err != nil {
		c.Tx.Error(" error getting domain")
		return []IngestionError{}, 0, err
	}
	if out.DeadLetterQueueUrl == nil {
		return []IngestionError{}, 0, errors.New("No Dead Letter Queue configured with Profile Domain " + c.DomainName)
	}
	c.Tx.Debug(" found dead letter queue %+v", *out.DeadLetterQueueUrl)
	sqsSvc := sqs.InitWithQueueUrl(*out.DeadLetterQueueUrl, c.Region, c.solutionId, c.solutionVersion)
	//TODO: fix this. will not sacle. dump DQD content into Dynamo asyncroniously
	res, err := sqsSvc.Get(sqs.GetMessageOptions{})
	if err != nil {
		c.Tx.Error(" Error fetching from SQS queue %s: %v", *out.DeadLetterQueueUrl, err)
		return []IngestionError{}, 0, err
	}
	c.Tx.Debug(" SQS response: %+v ", res)
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
func (c CustomerProfileConfig) SearchMultipleProfiles(key string, values []string) ([]profilemodel.Profile, error) {
	c.Tx.Info("[SearchMultipleProfiles] Parallel search for multiple profiles for %v in %v", key, values)
	var wg sync.WaitGroup
	wg.Add(len(values))
	allProfiles := []profilemodel.Profile{}
	var errs []error
	for i, value := range values {
		go func(val string, index int) {
			c.Tx.Debug("[core][customerProfiles[SearchMultipleProfiles]] Search %d started for value %v", index, val)
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
		c.Tx.Debug(
			"[core][customerProfiles[SearchMultipleProfiles] At least one error occured during one of the parralel searches %v",
			errs,
		)
		return []profilemodel.Profile{}, errs[0]
	}
	c.Tx.Info("[core][customerProfiles[SearchMultipleProfiles]] Final search results %+v", allProfiles)
	return allProfiles, nil
}

func (c CustomerProfileConfig) SearchProfiles(key string, values []string) ([]profilemodel.Profile, error) {
	if len(values) > 1 {
		return []profilemodel.Profile{}, errors.New("service only suppors one value for now")
	}
	c.Tx.Info("[SearchProfiles] Search profile for %v in %v", key, values)
	input := &customerProfileSdk.SearchProfilesInput{
		DomainName: aws.String(c.DomainName),
		KeyName:    aws.String(key),
		Values:     values,
	}
	out, err := c.Client.SearchProfiles(context.TODO(), input)
	profiles := []profilemodel.Profile{}
	if err != nil {
		return profiles, err
	}
	for _, item := range out.Items {
		p := toProfile(&item)
		p.Domain = c.DomainName
		profiles = append(profiles, p)
	}
	for out.NextToken != nil {
		c.Tx.Debug("[SearchProfiles] response is paginated. getting next bacth from token: %+v", out.NextToken)
		input.NextToken = out.NextToken
		out, err = c.Client.SearchProfiles(context.TODO(), input)
		if err != nil {
			return profiles, err
		}
		for _, item := range out.Items {
			p := toProfile(&item)
			p.Domain = c.DomainName
			profiles = append(profiles, p)
		}
	}
	c.Tx.Info("[SearchProfiles] found %v profiles", len(profiles))
	return profiles, nil
}

// TODO: parallelize the calls
// Search for a profile by ProfileID, and return data for specified object types.
func (c CustomerProfileConfig) GetProfile(
	id string,
	objectTypeNames []string,
	pagination []PaginationOptions,
) (profilemodel.Profile, error) {
	c.Tx.Info("[GetProfile] 0-retreiving profile with ID : %+v", id)
	//putting pagination options in a map to allow for qiuck access
	poMap := map[string]PaginationOptions{}
	for _, po := range pagination {
		poMap[po.ObjectType] = po
	}
	c.Tx.Debug("[GetProfile] 0-building pagination option map: %v", poMap)

	c.Tx.Debug("[GetProfile] 1-Search profile")
	res, err := c.SearchProfiles(PROFILE_ID_KEY, []string{id})
	if err != nil {
		return profilemodel.Profile{}, err
	}
	profiles := []profilemodel.Profile{}
	//filter to retain only responose with peofile id
	for _, profile := range res {
		if profile.ProfileId == id {
			profiles = append(profiles, profile)
		}
	}
	if len(profiles) == 0 {
		return profilemodel.Profile{}, errors.New("Profile with id " + id + " not found ")
	}
	if len(profiles) > 1 {
		return profilemodel.Profile{}, errors.New("Multiple profiles found for ID " + id)
	}
	p := profiles[0]
	c.Tx.Debug("[GetProfile] 2-Get profile objects")

	// Concurrently get data for all profile object types
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	errs := []error{}
	wg.Add(len(objectTypeNames))
	for _, objectType := range objectTypeNames {
		go func(objectType, id string) {
			obj, err := c.GetProfileObject(objectType, id, poMap[objectType])
			mu.Lock()
			if err != nil {
				if strings.Contains(err.Error(), "ResourceNotFoundException") {
					// Ignore ResourceNotFoundException errors when fetching profile objects.
					// This error occurs when a request is made to get object type(s) that do not exist
					// in an ACCP domain's mapping - likely as a result of a customer using a version
					// of UPT with new objects and a domain that has not migrated to the latest mapping.
					c.Tx.Warn(
						"[GetProfile] WARNING: ResourceNotFoundException error is being ignored. This often occurs when there is a new object type available, but not mapped to a specific Customer Profiles domain.",
					)
				} else {
					errs = append(errs, err)
				}
			} else {
				p.ProfileObjects = append(p.ProfileObjects, obj...)
			}
			mu.Unlock()
			wg.Done()
		}(objectType, id)
	}
	wg.Wait()
	if len(errs) > 0 {
		c.Tx.Error("[GetProfile] Errors while fetching profile objects: %v", errs)
		return p, core.ErrorsToError(errs)
	}
	return p, nil
}

func (c CustomerProfileConfig) DeleteProfile(id string, options ...DeleteProfileOptions) error {
	c.Tx.Info(" Delete profile %v", id)
	input := &customerProfileSdk.DeleteProfileInput{
		DomainName: aws.String(c.DomainName),
		ProfileId:  aws.String(id),
	}
	out, err := c.Client.DeleteProfile(context.TODO(), input)
	c.Tx.Debug(" Delete profile response: %+v", out)
	return err
}

func (c CustomerProfileConfig) GetProfileObject(
	objectTypeName string,
	profileID string,
	pagination PaginationOptions,
) ([]profilemodel.ProfileObject, error) {
	// c.Tx.Log("[GetProfileObject] List objects of type %s, for profile %v and pagination options: %+v", objectTypeName, profileID, pagination)
	input := &customerProfileSdk.ListProfileObjectsInput{
		DomainName:     aws.String(c.DomainName),
		ObjectTypeName: aws.String(objectTypeName),
		ProfileId:      aws.String(profileID),
		MaxResults:     aws.Int32(100),
	}
	//If PageSize==0, we assume that no pagination is provided
	if pagination.PageSize > 0 {
		input.MaxResults = aws.Int32(int32(pagination.PageSize))
	}
	page := 0
	out, err := c.Client.ListProfileObjects(context.TODO(), input)
	//c.Tx.Log(" Objects Search response: %+v", out)
	objects := []profilemodel.ProfileObject{}
	if err != nil {
		return objects, err
	}
	for _, item := range out.Items {
		objects = append(objects, toProfileObject(&item))
	}
	for out.NextToken != nil && (pagination.PageSize == 0 || pagination.Page > page) {
		page++
		// c.Tx.Log("[GetProfileObject] response is paginated. getting next batch from token: %+v", out.NextToken)
		input.NextToken = out.NextToken
		out, err = c.Client.ListProfileObjects(context.TODO(), input)
		if err != nil {
			return objects, err
		}
		for _, item := range out.Items {
			objects = append(objects, toProfileObject(&item))
		}
	}
	for i, order := range objects {
		objects[i].Attributes, err = parseProfileObject(order.JSONContent)
		if err != nil {
			c.Tx.Error("[GetProfileObject] Error Pasring order object: %+v", err)
			return []profilemodel.ProfileObject{}, err
		}
		objects[i].AttributesInterface, err = ParseProfileObjectInterface(order.JSONContent)
		if err != nil {
			c.Tx.Error("[GetProfileObject] Error Pasring order object: %+v", err)
			return []profilemodel.ProfileObject{}, err
		}
	}
	if pagination.PageSize > 0 {
		objects = objects[pagination.Page*pagination.PageSize:]
	}
	return objects, nil
}

func MakeNoProfileObjectErrorString(objectId, profileId string) string {
	return fmt.Sprintf("no object with id %s exists for traveller %s", objectId, profileId)
}

func (c CustomerProfileConfig) GetSpecificProfileObject(
	objectKey string,
	objectId string,
	profileId string,
	objectTypeName string,
) (profilemodel.ProfileObject, error) {
	paginate := PaginationOptions{
		Page:       0,
		PageSize:   0,
		ObjectType: objectTypeName,
	}
	profileObjectList, err := c.GetProfileObject(objectTypeName, profileId, paginate)
	if err != nil {
		return profilemodel.ProfileObject{}, err
	}

	newProfileObjectList := []profilemodel.ProfileObject{}
	for _, object := range profileObjectList {
		val, ok := object.Attributes[objectKey]
		if ok && val == objectId {
			newProfileObjectList = append(newProfileObjectList, object)
		}
	}
	if len(newProfileObjectList) == 0 {
		return profilemodel.ProfileObject{}, fmt.Errorf(MakeNoProfileObjectErrorString(objectId, profileId))
	}
	if len(newProfileObjectList) > 1 {
		return profilemodel.ProfileObject{}, errors.New(MULTIPLE_OBJECT_ERROR)
	}
	return newProfileObjectList[0], nil
}

func (c CustomerProfileConfig) DeleteProfileObject(
	objectKey string,
	objectId string,
	profileId string,
	objectTypeName string,
) (string, error) {
	object, err := c.GetSpecificProfileObject(objectKey, objectId, profileId, objectTypeName)
	if err != nil {
		return "", err
	}
	input := &customerProfileSdk.DeleteProfileObjectInput{
		DomainName:             aws.String(c.DomainName),
		ObjectTypeName:         aws.String(objectTypeName),
		ProfileId:              aws.String(profileId),
		ProfileObjectUniqueKey: &object.ID,
	}
	out, err := c.Client.DeleteProfileObject(context.TODO(), input)
	return *out.Message, err
}

func parseProfileObject(jsonObject string) (map[string]string, error) {
	attributes := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonObject), &attributes)
	return core.MapStringInterfaceToMapStringString(attributes), err
}

func ParseProfileObjectInterface(jsonObject string) (map[string]interface{}, error) {
	attributes := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonObject), &attributes)
	return attributes, err
}

func toMatchList(item *types.MatchItem) MatchList {
	return MatchList{
		ConfidenceScore: aws.ToFloat64(item.ConfidenceScore),
		ProfileIds:      item.ProfileIds,
	}
}

func toProfile(item *types.Profile) profilemodel.Profile {
	profile := profilemodel.Profile{
		ProfileId:            aws.ToString(item.ProfileId),
		AccountNumber:        aws.ToString(item.AccountNumber),
		FirstName:            aws.ToString(item.FirstName),
		MiddleName:           aws.ToString(item.MiddleName),
		LastName:             aws.ToString(item.LastName),
		BirthDate:            aws.ToString(item.BirthDate),
		Gender:               aws.ToString(item.GenderString),
		PhoneNumber:          aws.ToString(item.PhoneNumber),
		MobilePhoneNumber:    aws.ToString(item.MobilePhoneNumber),
		HomePhoneNumber:      aws.ToString(item.HomePhoneNumber),
		BusinessPhoneNumber:  aws.ToString(item.BusinessPhoneNumber),
		EmailAddress:         aws.ToString(item.EmailAddress),
		PersonalEmailAddress: aws.ToString(item.PersonalEmailAddress),
		BusinessEmailAddress: aws.ToString(item.BusinessEmailAddress),
		BusinessName:         aws.ToString(item.BusinessName),
		Attributes:           item.Attributes,
	}
	// Safely access fields if Address is not nil
	if item.Address != nil {
		profile.Address.Address1 = aws.ToString(item.Address.Address1)
		profile.Address.Address2 = aws.ToString(item.Address.Address2)
		profile.Address.Address3 = aws.ToString(item.Address.Address3)
		profile.Address.Address4 = aws.ToString(item.Address.Address4)
		profile.Address.City = aws.ToString(item.Address.City)
		profile.Address.State = aws.ToString(item.Address.State)
		profile.Address.Province = aws.ToString(item.Address.Province)
		profile.Address.PostalCode = aws.ToString(item.Address.PostalCode)
		profile.Address.Country = aws.ToString(item.Address.Country)
	}
	if item.ShippingAddress != nil {
		profile.ShippingAddress.Address1 = aws.ToString(item.ShippingAddress.Address1)
		profile.ShippingAddress.Address2 = aws.ToString(item.ShippingAddress.Address2)
		profile.ShippingAddress.Address3 = aws.ToString(item.ShippingAddress.Address3)
		profile.ShippingAddress.Address4 = aws.ToString(item.ShippingAddress.Address4)
		profile.ShippingAddress.City = aws.ToString(item.ShippingAddress.City)
		profile.ShippingAddress.State = aws.ToString(item.ShippingAddress.State)
		profile.ShippingAddress.Province = aws.ToString(item.ShippingAddress.Province)
		profile.ShippingAddress.PostalCode = aws.ToString(item.ShippingAddress.PostalCode)
		profile.ShippingAddress.Country = aws.ToString(item.ShippingAddress.Country)
	}
	if item.MailingAddress != nil {
		profile.MailingAddress.Address1 = aws.ToString(item.MailingAddress.Address1)
		profile.MailingAddress.Address2 = aws.ToString(item.MailingAddress.Address2)
		profile.MailingAddress.Address3 = aws.ToString(item.MailingAddress.Address3)
		profile.MailingAddress.Address4 = aws.ToString(item.MailingAddress.Address4)
		profile.MailingAddress.City = aws.ToString(item.MailingAddress.City)
		profile.MailingAddress.State = aws.ToString(item.MailingAddress.State)
		profile.MailingAddress.Province = aws.ToString(item.MailingAddress.Province)
		profile.MailingAddress.PostalCode = aws.ToString(item.MailingAddress.PostalCode)
		profile.MailingAddress.Country = aws.ToString(item.MailingAddress.Country)
	}
	if item.BillingAddress != nil {
		profile.BillingAddress.Address1 = aws.ToString(item.BillingAddress.Address1)
		profile.BillingAddress.Address2 = aws.ToString(item.BillingAddress.Address2)
		profile.BillingAddress.Address3 = aws.ToString(item.BillingAddress.Address3)
		profile.BillingAddress.Address4 = aws.ToString(item.BillingAddress.Address4)
		profile.BillingAddress.City = aws.ToString(item.BillingAddress.City)
		profile.BillingAddress.State = aws.ToString(item.BillingAddress.State)
		profile.BillingAddress.Province = aws.ToString(item.BillingAddress.Province)
		profile.BillingAddress.PostalCode = aws.ToString(item.BillingAddress.PostalCode)
		profile.BillingAddress.Country = aws.ToString(item.BillingAddress.Country)
	}
	return profile
}

func toProfileObject(item *types.ListProfileObjectsItem) profilemodel.ProfileObject {
	return profilemodel.ProfileObject{
		ID:          aws.ToString(item.ProfileObjectUniqueKey),
		Type:        aws.ToString(item.ObjectTypeName),
		JSONContent: aws.ToString(item.Object),
	}
}

func (c CustomerProfileConfig) PutProfileObject(object, objectTypeName string) error {
	c.Tx.Info("[PutProfileObject] Putting object type %v in domain %v", objectTypeName, c.DomainName)
	input := customerProfileSdk.PutProfileObjectInput{
		DomainName:     &c.DomainName,
		Object:         &object,
		ObjectTypeName: &objectTypeName,
	}
	out, err := c.Client.PutProfileObject(context.TODO(), &input)
	if err != nil {
		c.Tx.Error("[PutProfileObject] Error putting object: %v", err)
		return err
	}
	c.Tx.Info(
		"[PutProfileObject] Successfully put object of type %v. Unique key: %v",
		objectTypeName,
		out.ProfileObjectUniqueKey,
	)
	return nil
}

func (c *CustomerProfileConfig) WaitForMappingCreation(name string) error {
	maxTries := 10
	try := 0
	for try < maxTries {
		mapping, err := c.GetMapping(name)
		if err == nil && mapping.Name == name {
			c.Tx.Info("[WaitForMappingCreation] Mapping creation successful")
			return nil
		}
		c.Tx.Debug("[WaitForMappingCreation] Mapping not ready waiting 5 s")
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
			c.Tx.Info("[WaitForIntegrationCreation] Integration creation successful")
			return nil
		}
		c.Tx.Debug("[WaitForIntegrationCreation] Integration not ready waiting 5 s")
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

func (c *CustomerProfileConfig) GetProfileId(profileKey, profileId string) (string, error) {
	c.Tx.Info("[GetProfileId] Searching profile with ID %v in domain %v", profileId, c.DomainName)
	input := customerProfileSdk.SearchProfilesInput{
		DomainName: &c.DomainName,
		KeyName:    aws.String(profileKey),
		Values:     []string{profileId},
	}
	c.Tx.Debug("[GetProfileId] SearchProfiles rq: %+v", input)
	output, err := c.Client.SearchProfiles(context.TODO(), &input)
	c.Tx.Debug("[GetProfileId] SearchProfiles rs: %+v", output)
	if err != nil {
		c.Tx.Error("[GetProfileId] Error searching for profile %v: %v", profileId, err)
		return "", err
	}
	if len(output.Items) > 1 {
		c.Tx.Error("[GetProfileId] Error: Found multiple profiles with same ID")
		return "", errors.New("multiple profiles found with same id")
	}
	if len(output.Items) == 0 {
		c.Tx.Info("[GetProfileId] No profile found with ID %v", profileId)
		return "", nil
	}
	c.Tx.Info("[GetProfileId] successfully found profile with ID %v", profileId)
	return *output.Items[0].ProfileId, nil
}

func (c *CustomerProfileConfig) Exists(profileIDValue string) (bool, error) {
	c.Tx.Info("[Exists] Searching profile with ID %s=>%s in domain%v", PROFILE_ID_KEY, profileIDValue, c.DomainName)
	input := customerProfileSdk.SearchProfilesInput{
		DomainName: &c.DomainName,
		KeyName:    aws.String(PROFILE_ID_KEY),
		Values:     []string{profileIDValue},
	}
	output, err := c.Client.SearchProfiles(context.TODO(), &input)
	if err != nil {
		c.Tx.Error("[GetProfileId] Error searching for profile %s=>%s: %v", PROFILE_ID_KEY, profileIDValue, err)
		return false, err
	}
	for _, item := range output.Items {
		if *item.ProfileId == profileIDValue {
			return true, nil
		}
	}
	return false, nil
}

func (c *CustomerProfileConfig) MergeProfiles(profileId string, profileIdToMerge string) (string, error) {
	c.Tx.Info("[MergeProfiles] Merging profile %v with profile %v", profileId, profileIdToMerge)
	mergeArray := []string{profileIdToMerge}
	input := customerProfileSdk.MergeProfilesInput{
		MainProfileId:        &profileId,
		ProfileIdsToBeMerged: mergeArray,
		DomainName:           &c.DomainName,
	}

	out, err := c.Client.MergeProfiles(context.TODO(), &input)
	if err != nil {
		c.Tx.Error("[MergeProfiles] Error merging profiles: %v", err)
		return "", err
	}

	message := out.Message
	c.Tx.Info("[MergeProfiles] Merge profile complete with message %v", *message)
	return *message, nil
}

func (c *CustomerProfileConfig) MergeMany(pairs []ProfilePair) (string, error) {
	c.Tx.Info("[MergeProfiles] Merging %d profile pairs", len(pairs))

	profilePaisBySourceID := organizeProfilePairsBySourceID(pairs)
	errs := []error{}
	inputs := []*customerProfileSdk.MergeProfilesInput{}
	for sourceID, profilePairs := range profilePaisBySourceID {
		targets := []string{}
		for _, pair := range profilePairs {
			targets = append(targets, pair.TargetID)
		}
		inputs = append(inputs, &customerProfileSdk.MergeProfilesInput{
			MainProfileId:        aws.String(sourceID),
			ProfileIdsToBeMerged: targets,
			DomainName:           &c.DomainName,
		})

	}
	batchSize := 10
	batcnInputs := core.Chunk(core.InterfaceSlice(inputs), batchSize)
	for _, inputs := range batcnInputs {
		var wg sync.WaitGroup
		mu := &sync.Mutex{}
		wg.Add(len(inputs))
		for _, input := range inputs {
			go func(mergeInput *customerProfileSdk.MergeProfilesInput) {
				c.Tx.Debug("[MergeProfiles] ACCP Merge request: %+v", mergeInput)
				_, err := c.Client.MergeProfiles(context.TODO(), mergeInput)
				if err != nil {
					c.Tx.Error("[MergeProfiles] Error merging profiles: %v", err)
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
				wg.Done()
			}(input.(*customerProfileSdk.MergeProfilesInput))
		}
		wg.Wait()
	}

	status := "SUCCESS"
	if len(errs) > 0 {
		status = "FAILURE"
	}
	c.Tx.Info("[MergeProfiles] Merge profile complete with status %v", status)
	return status, core.ErrorsToError(errs)
}

func organizeProfilePairsBySourceID(pairs []ProfilePair) map[string][]ProfilePair {
	profilePaisBySourceID := map[string][]ProfilePair{}
	for _, v := range pairs {
		profilePaisBySourceID[v.SourceID] = append(profilePaisBySourceID[v.SourceID], v)
	}
	return profilePaisBySourceID
}

func (c *CustomerProfileConfig) CreateEventStream(domainName, streamName, streamArn string) error {
	input := customerProfileSdk.CreateEventStreamInput{
		DomainName:      &domainName,
		EventStreamName: &streamName,
		Uri:             &streamArn,
	}
	_, err := c.Client.CreateEventStream(context.TODO(), &input)
	if err != nil {
		c.Tx.Error("[CreateEventStream] Error creating event stream: %v", err)
		return err
	}

	c.Tx.Info("[CreateEventStream] Event stream created")
	return nil
}

// Note - event stream is automatically deleted when domain is deleted
func (c *CustomerProfileConfig) DeleteEventStream(domainName, streamName string) error {
	input := customerProfileSdk.DeleteEventStreamInput{
		DomainName:      &domainName,
		EventStreamName: &streamName,
	}
	_, err := c.Client.DeleteEventStream(context.TODO(), &input)
	if err != nil {
		c.Tx.Error("[DeleteEventStream] Error deleting event stream: %v", err)
		return err
	}

	c.Tx.Info("[DeleteEventStream] Event stream deleted")
	return nil
}

func (fms FieldMappings) GetSourceNames() []string {
	names := []string{}
	for _, v := range fms {
		split := strings.Split(v.Source, ".")
		names = append(names, split[len(split)-1])
	}
	return names
}

func (fms FieldMappings) GetTypeNames() []string {
	names := []string{}
	for _, v := range fms {
		split := strings.Split(v.Type, ".")
		names = append(names, split[len(split)-1])
	}
	return names
}

// this function is distinct here because ACCP lives in tah-core
// feature set version knowledge does not belong in tah-core
// this information should be pulled up into UPT common or LCS
// however, the tests currently depend on it
func BuildGenericIntegrationFieldMappingWithTravelerId() FieldMappings {
	return FieldMappings{
		{
			Type:    "STRING",
			Source:  "_source.timestamp",
			Target:  "timestamp",
			Indexes: []string{},
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.traveller_id",
			Target:      "traveller_id",
			Indexes:     []string{},
			Searcheable: true,
			KeyOnly:     true,
		},
		{
			Type:        "STRING",
			Source:      "_source." + CONNECT_ID_ATTRIBUTE,
			Target:      "_profile.Attributes." + CONNECT_ID_ATTRIBUTE,
			Indexes:     []string{"PROFILE"},
			Searcheable: true,
		},
		{
			Type:    "STRING",
			Source:  "_source." + UNIQUE_OBJECT_ATTRIBUTE,
			Target:  UNIQUE_OBJECT_ATTRIBUTE,
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
	}
}

func BuildGenericIntegrationFieldMapping() FieldMappings {
	return FieldMappings{
		{
			Type:    "STRING",
			Source:  "_source.timestamp",
			Target:  "timestamp",
			Indexes: []string{},
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source." + CONNECT_ID_ATTRIBUTE,
			Target:      "_profile.Attributes." + CONNECT_ID_ATTRIBUTE,
			Indexes:     []string{"PROFILE"},
			Searcheable: true,
		},
		{
			Type:    "STRING",
			Source:  "_source." + UNIQUE_OBJECT_ATTRIBUTE,
			Target:  UNIQUE_OBJECT_ATTRIBUTE,
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
	}
}

func BuildProfileFieldMapping() FieldMappings {
	return FieldMappings{
		{
			Type:   "STRING",
			Source: "_source.Attributes.profile_id",
			Target: "_profile.Attributes.profile_id",
			//Indexes:     []string{"UNIQUE"},
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.Attributes." + CONNECT_ID_ATTRIBUTE,
			Target:      "_profile.Attributes." + CONNECT_ID_ATTRIBUTE,
			Indexes:     []string{"PROFILE", "UNIQUE"},
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.Attributes." + TIMESTAMP_ATTRIBUTE,
			Target: "_profile.SourceLastUpdatedTimestamp",
		},
		{
			Type:   "STRING",
			Source: "_source.FirstName",
			Target: "_profile.FirstName",
		},
		{
			Type:   "STRING",
			Source: "_source.MiddleName",
			Target: "_profile.MiddleName",
		},
		{
			Type:   "STRING",
			Source: "_source.LastName",
			Target: "_profile.LastName",
		},
		{
			Type:   "STRING",
			Source: "_source.Gender",
			Target: "_profile.Gender",
		},
		{
			Type:   "STRING",
			Source: "_source.PhoneNumber",
			Target: "_profile.PhoneNumber",
		},
		{
			Type:   "STRING",
			Source: "_source.MobilePhoneNumber",
			Target: "_profile.MobilePhoneNumber",
		},
		{
			Type:   "STRING",
			Source: "_source.BuisnessPhoneNumber",
			Target: "_profile.BusinessPhoneNumber",
		},
		{
			Type:   "STRING",
			Source: "_source.HomePhoneNumber",
			Target: "_profile.HomePhoneNumber",
		},
		{
			Type:   "STRING",
			Source: "_source.BusinessName",
			Target: "_profile.BusinessName",
		},
		{
			Type:   "STRING",
			Source: "_source.BirthDate",
			Target: "_profile.BirthDate",
		},
		{
			Type:   "STRING",
			Source: "_source.EmailAddress",
			Target: "_profile.EmailAddress",
		},
		{
			Type:   "STRING",
			Source: "_source.PersonalEmailAddress",
			Target: "_profile.PersonalEmailAddress",
		},
		{
			Type:   "STRING",
			Source: "_source.BusinessEmailAddress",
			Target: "_profile.BusinessEmailAddress",
		},
		//Addresses
		{
			Type:   "STRING",
			Source: "_source.Address",
			Target: "_profile.Address",
		},
		//Billing Address
		{
			Type:   "STRING",
			Source: "_source.BillingAddress",
			Target: "_profile.BillingAddress",
		},
		//Mailing Address
		{
			Type:   "STRING",
			Source: "_source.MailingAddress",
			Target: "_profile.MailingAddress",
		},
		//Shipping Address
		{
			Type:   "STRING",
			Source: "_source.ShippingAddress",
			Target: "_profile.ShippingAddress",
		},
	}
}

func (c *CustomerProfileConfig) PutProfileAsObject(p profilemodel.Profile) error {
	copyP := p
	copyP.Attributes = maps.Clone(p.Attributes)
	copyP.Attributes[CONNECT_ID_ATTRIBUTE] = p.ProfileId
	milli := p.LastUpdated.UnixMilli()
	copyP.Attributes[TIMESTAMP_ATTRIBUTE] = fmt.Sprintf("%d", milli)
	profileSerialized, err := json.Marshal(copyP)
	if err != nil {
		c.Tx.Error("Error marshalling profile %v", err)
		return err
	}
	profileString := string(profileSerialized)
	err = c.PutProfileObject(profileString, PROFILE_FIELD_OBJECT_TYPE)
	if err != nil {
		c.Tx.Error("Error putting profile %v", err)
		return err
	}
	return nil
}

func (c *CustomerProfileConfig) PutProfileObjectFromLcs(p profilemodel.ProfileObject, lcsId string) error {
	p.AttributesInterface[CONNECT_ID_ATTRIBUTE] = lcsId
	p.AttributesInterface[UNIQUE_OBJECT_ATTRIBUTE] = p.ID + lcsId
	encodedExtendedData, ok := p.Attributes[EXTENDED_DATA_ATTRIBUTE]
	if ok {
		if !slices.Contains([]string{"", "nil", "<nil>", "None"}, encodedExtendedData) {
			decoded, err := base64.StdEncoding.DecodeString(encodedExtendedData)
			if err != nil {
				c.Tx.Error("Error decoding extended data %v", err)
				return err
			}
			var marshalled map[string]interface{}
			err = json.Unmarshal(decoded, &marshalled)
			if err != nil {
				c.Tx.Error("Error unmarshalling extended data %v", err)
				return err
			}
			p.AttributesInterface[EXTENDED_DATA_ATTRIBUTE] = marshalled
		}
	}
	profileSerialized, err := json.Marshal(p.AttributesInterface)
	if err != nil {
		c.Tx.Error("Error marshalling profile object %v", err)
		return err
	}
	profileString := string(profileSerialized)
	err = c.PutProfileObject(profileString, p.Type)
	if err != nil {
		c.Tx.Error("Error putting profile object %v", err)
		return err
	}
	return nil
}

func (c *CustomerProfileConfig) UpdateProfileAttributesFieldMapping(
	objectTypeName string,
	description string,
	additionalFields []string,
) error {
	objectMapping, err := c.GetMapping(objectTypeName)
	if err != nil {
		c.Tx.Error("Error retrieving mapping %v", err)
		return err
	}
	fieldMapping := objectMapping.Fields
	var targetValues []string
	for _, mapping := range fieldMapping {
		if strings.HasPrefix(mapping.Target, "_profile.") {
			targetValue := strings.TrimPrefix(mapping.Target, "_profile.")
			targetValues = append(targetValues, targetValue)
		}
	}

	additionalFieldMappings := []FieldMapping{}
	for _, field := range additionalFields {
		if stringInSlice(field, targetValues) {
			continue
		}
		newFieldMapping := FieldMapping{
			Type:   "STRING",
			Source: "_source." + field,
			Target: "_profile." + field,
		}
		additionalFieldMappings = append(additionalFieldMappings, newFieldMapping)
	}
	fieldMapping = append(fieldMapping, additionalFieldMappings...)
	err = c.CreateMapping(objectTypeName, description, fieldMapping)
	if err != nil {
		c.Tx.Error("Error creating mapping %v", err)
		return err
	}
	return nil
}

// Function to check if a string is present in a slice of strings
func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
