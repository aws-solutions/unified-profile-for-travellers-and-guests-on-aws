// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package entityresolution

import (
	"context"
	"log"
	"strings"
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/iam"
	"tah/upt/source/tah-core/s3"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/entityresolution"
	entityResolutionTypes "github.com/aws/aws-sdk-go-v2/service/entityresolution/types"
	requestsigner "github.com/opensearch-project/opensearch-go/v2/signer/aws"
)

type Config struct {
	Client          *entityresolution.Client
	Tx              core.Transaction
	Region          string
	Signer          requestsigner.Signer
	S3Config        s3.S3Config
	GlueDB          string
	SolutionId      string
	SolutionVersion string
}

type EntityResolutionProfile struct {
	LastUpdated       time.Time
	ConnectId         string
	ProfileId         string
	AccountNumber     string
	Name              string
	FirstName         string
	MiddleName        string
	LastName          string
	Phone             string
	PhoneNumber       string
	EmailAddress      string
	Attributes        map[string]string
	Address           string
	AddressStreet1    string
	AddressStreet2    string
	AddressStreet3    string
	AddressCity       string
	AddressState      string
	AddressCountry    string
	AddressPostalCode string
}

type EntityResolutionFieldMapping struct {
	Source string
	Target string
}

const (
	JobStatusRunning   = "RUNNING"
	JobStatusSucceeded = "SUCCEEDED"
	JobStatusFailed    = "FAILED"
	JobStatusQueued    = "QUEUED"
)

var RESERVED_FIELDS = []string{
	"NAME",
	"NAME_FIRST",
	"NAME_MIDDLE",
	"NAME_LAST",
	"PHONE",
	"PHONE_NUMBER",
	"PHONE_COUNTRYCODE",
	"EMAIL_ADDRESS",
	"ADDRESS",
	"ADDRESS_STREET1",
	"ADDRESS_STREET2",
	"ADDRESS_STREET3",
	"ADDRESS_CITY",
	"ADDRESS_STATE",
	"ADDRESS_COUNTRY",
	"ADDRESS_POSTAL_CODE",
	"UNIQUE_ID",
	"DATE",
	"STRING",
	"PROVIDER_ID",
}

func Init(region string, glueDb string, solutionId, solutionVersion string) Config {
	//httpClient := core.CreateClient(solutionId, solutionVersion)
	// sess := session.Must(session.NewSession(&aws.Config{
	// 	Region: *aws.String(region),
	// }))
	//TODO: Need to figure out a way to modify solution headers in go v2 for this
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return Config{
		Client:          entityresolution.NewFromConfig(cfg),
		Region:          region,
		GlueDB:          glueDb,
		SolutionId:      solutionId,
		SolutionVersion: solutionVersion,
	}
}

func (c *Config) SetTx(tx core.Transaction) {
	tx.LogPrefix = "entityresolution"
	c.Tx = tx
}

func (c Config) CreateMatchingWorkflow(workflowName string, schemaName string, tables []string, roleArn string, bucketOutputName string, uniqueColumn string) (string, error) {
	//TODO: test if normalization improves results
	//can't find any documentation on how normalization actually works
	applyNormalization := false
	glueTableArnList := []string{}
	inputSourceConfigList := []entityResolutionTypes.InputSource{}
	for _, tableName := range tables {
		accountId := extractAccountID(roleArn)
		partition := extractPartition(roleArn)
		tableArn := "arn:" + partition + ":glue:" + c.Region + ":" + accountId + ":table/" + c.GlueDB + "/" + tableName
		glueTableArnList = append(glueTableArnList, tableArn)
	}

	for _, tableArn := range glueTableArnList {
		inputSourceConfig := entityResolutionTypes.InputSource{
			InputSourceARN:     &tableArn,
			SchemaName:         &schemaName,
			ApplyNormalization: &applyNormalization,
		}
		inputSourceConfigList = append(inputSourceConfigList, inputSourceConfig)
	}

	outputAttribute := entityResolutionTypes.OutputAttribute{
		Name: &uniqueColumn,
	}

	outputNameUri := "s3://" + bucketOutputName
	outputSourceConfig := entityResolutionTypes.OutputSource{
		OutputS3Path: &outputNameUri,
		Output:       []entityResolutionTypes.OutputAttribute{outputAttribute},
	}

	resolutionType := entityResolutionTypes.ResolutionTypeMlMatching

	resolution := entityResolutionTypes.ResolutionTechniques{
		ResolutionType: resolutionType,
	}

	input := entityresolution.CreateMatchingWorkflowInput{
		InputSourceConfig:    inputSourceConfigList,
		OutputSourceConfig:   []entityResolutionTypes.OutputSource{outputSourceConfig},
		WorkflowName:         &workflowName,
		ResolutionTechniques: &resolution,
		RoleArn:              &roleArn,
	}

	output, err := c.Client.CreateMatchingWorkflow(context.TODO(), &input)
	if err != nil {
		c.Tx.Error("Error creating matching workflow: %v", err)
		return "", err
	}

	return *output.WorkflowName, nil
}

func (c Config) CreateSchema(fieldMappings []EntityResolutionFieldMapping, schemaName string) (string, error) {
	mappedInputFields := make([](entityResolutionTypes.SchemaInputAttribute), 0)
	for _, fieldMapping := range fieldMappings {
		c.Tx.Debug("Mapping %v to %v", fieldMapping.Source, fieldMapping.Target)
		//first group uses built in entity resolution fields, second group custom strings
		if core.SliceContains(fieldMapping.Target, RESERVED_FIELDS) {
			c.Tx.Info("Creating schema input attribute %v", fieldMapping.Target)
			schemaInputAttribute := c.CreateSchemaInputAttribute(fieldMapping.Source, fieldMapping.Target, "", fieldMapping.Source, "")
			mappedInputFields = append(mappedInputFields, schemaInputAttribute)
		} else {
			c.Tx.Info("Creating schema input attribute %v", fieldMapping.Target)
			schemaInputAttribute := c.CreateSchemaInputAttribute(fieldMapping.Source, "STRING", "", fieldMapping.Source, fieldMapping.Target)
			mappedInputFields = append(mappedInputFields, schemaInputAttribute)
		}
	}

	input := entityresolution.CreateSchemaMappingInput{
		SchemaName:        &schemaName,
		MappedInputFields: mappedInputFields,
	}

	output, err := c.Client.CreateSchemaMapping(context.Background(), &input)
	if err != nil {
		c.Tx.Error("Error creating schema mapping %v", err)
		return "", err
	}

	return *output.SchemaName, nil
}

func (c Config) DeleteSchema(schemaName string) (string, error) {
	input := entityresolution.DeleteSchemaMappingInput{
		SchemaName: &schemaName,
	}

	output, err := c.Client.DeleteSchemaMapping(context.Background(), &input)
	if err != nil {
		c.Tx.Error("Error deleting schema %v", err)
		return "", err
	}

	return *output.Message, nil
}

func (c Config) DeleteWorkflow(workflowName string) (string, error) {
	input := entityresolution.DeleteMatchingWorkflowInput{
		WorkflowName: &workflowName,
	}

	output, err := c.Client.DeleteMatchingWorkflow(context.Background(), &input)
	if err != nil {
		c.Tx.Error("Error deleting workflow %v", err)
		return "", err
	}

	return *output.Message, nil
}

func (c Config) CreateSchemaInputAttribute(fieldName string, attributeType string, groupName string, matchKey string, subType string) entityResolutionTypes.SchemaInputAttribute {

	inputType := entityResolutionTypes.SchemaAttributeType(attributeType)

	attribute := entityResolutionTypes.SchemaInputAttribute{
		FieldName: &fieldName,
		Type:      inputType,
		GroupName: &groupName,
		MatchKey:  &matchKey,
	}

	return attribute
}

func (c Config) StartMatchingJob(workflowName string) (string, error) {
	input := entityresolution.StartMatchingJobInput{
		WorkflowName: &workflowName,
	}

	output, err := c.Client.StartMatchingJob(context.TODO(), &input)
	if err != nil {
		c.Tx.Error("Error starting matching job %v", err)
		return "", err
	}

	return *output.JobId, nil
}

func (c Config) CreateRole(roleName string, actions []string) (string, string, error) {
	iamClient := iam.Init(c.SolutionId, c.SolutionVersion)
	doc := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect:   "Allow",
				Action:   actions,
				Resource: []string{"*"},
			},
		},
	}
	_, policyArn, err := iamClient.CreateRoleWithPolicy(roleName, "entityresolution.amazonaws.com", doc)
	return roleName, policyArn, err
}

func (c Config) GetMatchingJobStatus(jobId string, workflowName string) (string, error) {
	input := entityresolution.GetMatchingJobInput{
		JobId:        &jobId,
		WorkflowName: &workflowName,
	}

	output, err := c.Client.GetMatchingJob(context.Background(), &input)
	if err != nil {
		return "", err
	}
	return string(output.Status), nil
}

func (c Config) GetMatchingWorkflow(workflowName string) (string, error) {
	c.Tx.Info("Retrieving matching workflow %v", workflowName)
	input := entityresolution.GetMatchingWorkflowInput{
		WorkflowName: &workflowName,
	}
	output, err := c.Client.GetMatchingWorkflow(context.Background(), &input)
	if err != nil {
		c.Tx.Error("Error getting matching workflow %v", err)
		return "", err
	}
	return *output.WorkflowArn, nil
}

func extractAccountID(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) >= 5 {
		return parts[4]
	}
	return ""
}

func extractPartition(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) >= 5 {
		return parts[1]
	}
	return ""
}

func MapStringStringToEntityResolutionFieldMapping(mapStringString map[string]string) []EntityResolutionFieldMapping {
	fieldMappings := make([]EntityResolutionFieldMapping, 0)
	for source, target := range mapStringString {
		fieldMappings = append(fieldMappings, EntityResolutionFieldMapping{
			Source: source,
			Target: target,
		})
	}
	return fieldMappings
}
