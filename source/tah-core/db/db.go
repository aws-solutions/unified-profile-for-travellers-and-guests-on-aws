// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	core "tah/upt/source/tah-core/core"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	mw "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"
)

type Notification struct {
	MessageType string `json:"messageType"`
}

type DynamoRecord struct {
	Pk string `json:"pk"`
	Sk string `json:"sk"`
}

type DynamoUserConnection struct {
	UserId      string `json:"userId"`
	ConnetionId string `json:"connectionId"`
}

type DynamoEventChange struct {
	NewImage map[string]*types.AttributeValue `json:"NewImage"`
	OldImage map[string]*types.AttributeValue `json:"OldImage"`
}

type DynamoEventRecord struct {
	Change    DynamoEventChange `json:"dynamodb"`
	EventName string            `json:"eventName"`
	EventID   string            `json:"eventID"`
	// ... more fields if needed: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_streams_GetRecords.html
}

type DynamoEvent struct {
	Records []DynamoEventRecord `json:"records"`
}

type IDBConfig interface {
	SetTx(tx core.Transaction) error
	Save(prop interface{}) (interface{}, error)
	FindStartingWith(pk string, value string, data interface{}) error
	DeleteMany(data interface{}) error
	SaveMany(data interface{}) error
	FindAll(data interface{}, lastEvaluatedKey map[string]types.AttributeValue) (map[string]types.AttributeValue, error)
	Get(pk string, sk string, data interface{}) error
}

type DBConfig struct {
	DbService       *dynamodb.Client
	PrimaryKey      string
	SortKey         string
	TableName       string
	LambdaContext   context.Context
	Tx              core.Transaction
	SolutionId      string
	SolutionVersion string
}

type DynamoFilterExpression struct {
	BeginsWith []DynamoFilterCondition
	Contains   []DynamoFilterCondition
	EqualBool  []DynamoBoolFilterCondition
}
type DynamoProjectionExpression struct {
	Projection []string
}
type DynamoIndex struct {
	Name string
	Pk   string
	Sk   string
}

type DynamoFilterCondition struct {
	Key   string
	Value string
}

type DynamoBoolFilterCondition struct {
	Key   string
	Value bool
}

type DynamoPaginationOptions struct {
	Page     int64
	PageSize int32
}

type QueryOptions struct {
	Filter               DynamoFilterExpression
	Index                DynamoIndex
	PaginOptions         DynamoPaginationOptions
	ReverseOrder         bool
	ProjectionExpression DynamoProjectionExpression
	Consistent           bool
	NoObjectNoError      bool
}

type TableOptions struct {
	TTLAttribute  string
	GSIs          []GSI
	StreamEnabled bool
	Tags          map[string]string
}

type GSI struct {
	IndexName      string
	PkName         string
	PkType         string
	SkName         string
	SkType         string
	ProjectionType types.ProjectionType
}

type FindByGsiOptions struct {
	Limit int32
}

func (exp DynamoFilterExpression) HasFilter() bool {
	return exp.HasBeginsWith() || exp.HasContains() || exp.HasEqual()
}
func (exp DynamoFilterExpression) HasBeginsWith() bool {
	if len(exp.BeginsWith) > 0 && exp.BeginsWith[0].Key != "" {
		return true
	}
	return false
}
func (exp DynamoFilterExpression) HasContains() bool {
	if len(exp.Contains) > 0 && exp.Contains[0].Key != "" {
		return true
	}
	return false
}
func (exp DynamoFilterExpression) HasEqual() bool {
	if len(exp.EqualBool) > 0 && exp.EqualBool[0].Key != "" {
		return true
	}
	return false
}

// init setup teh session and define table name, primary key and sort key
func Init(tn string, pk string, sk string, solutionId, solutionVersion string) DBConfig {
	if pk == "" {
		log.Printf("[CORE][DB] WARNING: empty PK provided")
	}
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithAPIOptions([]func(*middleware.Stack) error{mw.AddUserAgentKey("AWSSOLUTION/" + solutionId + "/" + solutionVersion)}),
	)
	if err != nil {
		panic("error initializing dynamodb client")
	}
	// Create DynamoDB client
	return DBConfig{
		DbService:       dynamodb.NewFromConfig(cfg),
		PrimaryKey:      pk,
		SortKey:         sk,
		TableName:       tn,
		LambdaContext:   context.Background(),
		SolutionId:      solutionId,
		SolutionVersion: solutionVersion,
	}

}

func InitWithNewTable(tn string, pk string, sk string, solutionId, solutionVersion string) (DBConfig, error) {
	return InitWithNewTableAndTTL(tn, pk, sk, "", solutionId, solutionVersion)
}

func InitWithNewTableAndTTL(tn string, pk string, sk string, ttlAttr string, solutionId, solutionVersion string) (DBConfig, error) {
	if pk == "" {
		log.Printf("[CORE][DB] WARNING: empty PK provided")
	}
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithAPIOptions([]func(*middleware.Stack) error{mw.AddUserAgentKey("AWSSOLUTION/" + solutionId + "/" + solutionVersion)}),
	)
	if err != nil {
		panic("error initializing dynamodb client")
	}
	// Create DynamoDB client
	config := DBConfig{
		DbService:       dynamodb.NewFromConfig(cfg),
		PrimaryKey:      pk,
		SortKey:         sk,
		TableName:       tn,
		LambdaContext:   context.Background(),
		SolutionId:      solutionId,
		SolutionVersion: solutionVersion,
	}
	options := TableOptions{}
	if ttlAttr != "" {
		options.TTLAttribute = ttlAttr
	}
	err = config.CreateTableWithOptions(tn, pk, sk, options)
	if err != nil {
		log.Printf("[Dynamo][InitWithNewTable] error calling CreateTable: %s", err)
	}
	return config, err
}

// This function returns a handler for the provided table
func (dbc DBConfig) New(tn string, pk string, sk string) DBConfig {
	h := Init(tn, pk, sk, dbc.SolutionId, dbc.SolutionVersion)
	h.SetTx(dbc.Tx)
	return h
}

func (dbc DBConfig) CreateTable(tn string, pk string, sk string) error {
	return dbc.CreateTableWithOptions(tn, pk, sk, TableOptions{})
}

func (dbc DBConfig) CreateTableWithOptions(tn string, pk string, sk string, options TableOptions) error {
	attrsDef := []types.AttributeDefinition{
		{
			AttributeName: aws.String(pk),
			AttributeType: types.ScalarAttributeTypeS,
		},
	}
	keyEl := []types.KeySchemaElement{
		{
			AttributeName: aws.String(pk),
			KeyType:       types.KeyTypeHash,
		},
	}
	if sk != "" {
		attrsDef = append(attrsDef, types.AttributeDefinition{
			AttributeName: aws.String(sk),
			AttributeType: types.ScalarAttributeTypeS,
		})
		keyEl = append(keyEl, types.KeySchemaElement{
			AttributeName: aws.String(sk),
			KeyType:       types.KeyTypeRange,
		})
	}

	gsis := []types.GlobalSecondaryIndex{}
	for _, gsi := range options.GSIs {
		attrsDef = append(attrsDef, types.AttributeDefinition{
			AttributeName: aws.String(gsi.PkName),
			AttributeType: types.ScalarAttributeType(gsi.PkType),
		})
		keySchema := []types.KeySchemaElement{
			{
				AttributeName: aws.String(gsi.PkName),
				KeyType:       types.KeyTypeHash,
			},
		}
		if gsi.SkName != "" {
			keySchema = append(keySchema, types.KeySchemaElement{
				AttributeName: aws.String(gsi.SkName),
				KeyType:       types.KeyTypeRange,
			})
			if gsi.SkName != sk && gsi.PkName != sk {
				attrsDef = append(attrsDef, types.AttributeDefinition{
					AttributeName: aws.String(gsi.SkName),
					AttributeType: types.ScalarAttributeType(gsi.SkType),
				})
			}
		}
		if gsi.ProjectionType == "" {
			gsi.ProjectionType = types.ProjectionTypeAll
		}
		gsis = append(gsis, types.GlobalSecondaryIndex{
			IndexName: aws.String(gsi.IndexName),
			KeySchema: keySchema,
			Projection: &types.Projection{
				ProjectionType: gsi.ProjectionType,
			},
		})
	}

	tags := []types.Tag{}
	for tag, val := range options.Tags {
		tags = append(tags, types.Tag{
			Key:   aws.String(tag),
			Value: aws.String(val),
		})
	}

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: attrsDef,
		KeySchema:            keyEl,
		TableName:            aws.String(tn),
		BillingMode:          types.BillingModePayPerRequest,
		Tags:                 tags,
	}
	if len(gsis) > 0 {
		input.GlobalSecondaryIndexes = gsis
	}
	if options.StreamEnabled {
		input.StreamSpecification = &types.StreamSpecification{
			StreamEnabled:  aws.Bool(true),
			StreamViewType: types.StreamViewTypeNewAndOldImages,
		}
	}

	dbc.Tx.Info("[Dynamo][CreateTable] Create table input:  %+v", input)
	_, err := dbc.DbService.CreateTable(context.TODO(), input)
	if err != nil {
		dbc.Tx.Error("[Dynamo][CreateTable] error calling CreateTable: %s", err)
	}
	if options.TTLAttribute != "" {
		dbc.Tx.Debug("[Dynamo][CreateTable] Enabling time to live")
		dbc.TableName = tn
		dbc.WaitForTableCreation()
		err = dbc.EnabledTTl(options.TTLAttribute)
	}

	return err
}

// Note: this currently only supports STANDARD table class. To support other classes
// like STANDARD_INFREQUENT_ACCESS, createTableInput must be updated.
func (dbc DBConfig) ReplaceTable() error {
	// Get existing table configuration
	dbc.Tx.Debug("[Dynamo][ReplaceTable] Get existing table config")
	getTableInput := &dynamodb.DescribeTableInput{
		TableName: aws.String(dbc.TableName),
	}
	oldTable, err := dbc.DbService.DescribeTable(context.TODO(), getTableInput)
	if err != nil {
		return err
	}
	getTagsInput := &dynamodb.ListTagsOfResourceInput{
		ResourceArn: oldTable.Table.TableArn,
	}
	oldTableTags, err := dbc.DbService.ListTagsOfResource(context.TODO(), getTagsInput)
	if err != nil {
		return err
	}
	getTTLInput := &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(dbc.TableName),
	}
	oldTableTTL, err := dbc.DbService.DescribeTimeToLive(context.TODO(), getTTLInput)
	if err != nil {
		return err
	}
	// Delete existing table
	dbc.Tx.Info("[Dynamo][ReplaceTable] Delete existing table")
	err = dbc.DeleteTable(dbc.TableName)
	if err != nil {
		return err
	}
	// Wait for table to be deleted
	err = dbc.WaitForTableDeletion()
	if err != nil {
		return err
	}
	// Create new table
	dbc.Tx.Info("[Dynamo][ReplaceTable] Create new table")
	createTableInput := &dynamodb.CreateTableInput{
		AttributeDefinitions:   oldTable.Table.AttributeDefinitions,
		BillingMode:            oldTable.Table.BillingModeSummary.BillingMode,
		GlobalSecondaryIndexes: buildReplacementGSI(oldTable),
		KeySchema:              oldTable.Table.KeySchema,
		ProvisionedThroughput:  buildProvisionedThroughput(oldTable),
		LocalSecondaryIndexes:  buildReplacementLSI(oldTable),
		SSESpecification:       buildSSE(oldTable),
		StreamSpecification:    oldTable.Table.StreamSpecification,
		TableClass:             types.TableClassStandard,
		TableName:              oldTable.Table.TableName,
		Tags:                   oldTableTags.Tags,
	}
	_, err = dbc.DbService.CreateTable(context.TODO(), createTableInput)
	if err != nil {
		return err
	}
	// Wait for table to be created
	err = dbc.WaitForTableCreation()
	if err != nil {
		return err
	}
	// Enable TTL (SDK does not support creating a table with TTL enabled)
	// https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/#CreateTableInput
	if oldTableTTL.TimeToLiveDescription != nil &&
		oldTableTTL.TimeToLiveDescription.TimeToLiveStatus == types.TimeToLiveStatusEnabled {
		dbc.Tx.Info("[Dynamo][ReplaceTable] Enabling time to live")
		enableTTLInput := &dynamodb.UpdateTimeToLiveInput{
			TableName: aws.String(dbc.TableName),
			TimeToLiveSpecification: &types.TimeToLiveSpecification{
				AttributeName: oldTableTTL.TimeToLiveDescription.AttributeName,
				Enabled:       aws.Bool(true),
			},
		}
		_, err = dbc.DbService.UpdateTimeToLive(context.TODO(), enableTTLInput)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildReplacementGSI(t *dynamodb.DescribeTableOutput) []types.GlobalSecondaryIndex {
	if len(t.Table.GlobalSecondaryIndexes) == 0 {
		return nil
	}
	indexes := []types.GlobalSecondaryIndex{}
	for _, gsi := range t.Table.GlobalSecondaryIndexes {
		index := types.GlobalSecondaryIndex{
			IndexName:             gsi.IndexName,
			KeySchema:             gsi.KeySchema,
			Projection:            gsi.Projection,
			ProvisionedThroughput: buildProvisionedThroughput(t),
		}
		indexes = append(indexes, index)
	}
	return indexes
}

func buildReplacementLSI(t *dynamodb.DescribeTableOutput) []types.LocalSecondaryIndex {
	if len(t.Table.LocalSecondaryIndexes) == 0 {
		return nil
	}
	indexes := []types.LocalSecondaryIndex{}
	for _, lsi := range t.Table.LocalSecondaryIndexes {
		index := types.LocalSecondaryIndex{
			IndexName:  lsi.IndexName,
			KeySchema:  lsi.KeySchema,
			Projection: lsi.Projection,
		}
		indexes = append(indexes, index)
	}
	return indexes
}

// ProvisionedThroughput is only specified if BillingMode is PAY_PER_REQUEST.
// Otherwise, no value should be provided.
func buildProvisionedThroughput(oldTable *dynamodb.DescribeTableOutput) *types.ProvisionedThroughput {
	if oldTable.Table.BillingModeSummary.BillingMode == types.BillingModePayPerRequest {
		return nil
	}
	return &types.ProvisionedThroughput{
		ReadCapacityUnits:  oldTable.Table.ProvisionedThroughput.ReadCapacityUnits,
		WriteCapacityUnits: oldTable.Table.ProvisionedThroughput.WriteCapacityUnits,
	}
}

// Set server-side encryption to AWS managed key or AWS owned key. If a KMSMasterKeyId is
// provided, we use the KMS managed key. Otherwise, we use the default KMS key.
// https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/#SSESpecification
func buildSSE(oldTable *dynamodb.DescribeTableOutput) *types.SSESpecification {
	if oldTable.Table.SSEDescription != nil && oldTable.Table.SSEDescription.KMSMasterKeyArn != nil {
		return &types.SSESpecification{
			Enabled:        aws.Bool(true),
			KMSMasterKeyId: oldTable.Table.SSEDescription.KMSMasterKeyArn,
			SSEType:        oldTable.Table.SSEDescription.SSEType,
		}
	}
	return &types.SSESpecification{
		Enabled: aws.Bool(false),
	}
}

func (dbc DBConfig) EnabledTTl(attr string) error {
	dbc.Tx.Info("[Dynamo][EnabledTTl] Enabling time to live")
	existingAttr, _ := dbc.TTLAttribute()
	if existingAttr == attr {
		dbc.Tx.Info("[Dynamo][EnabledTTl] TTL is already enabled on attribute: %s, do nothing", existingAttr)
		return nil
	}
	input := &dynamodb.UpdateTimeToLiveInput{
		TableName: aws.String(dbc.TableName),
		TimeToLiveSpecification: &types.TimeToLiveSpecification{
			AttributeName: aws.String(attr),
			Enabled:       aws.Bool(true),
		},
	}
	dbc.Tx.Debug("[Dynamo][EnabledTTl] Create table input:  %+v", input)
	_, err := dbc.DbService.UpdateTimeToLive(context.TODO(), input)
	return err
}

// return the TTL attribute name fo rthe table or an erroor if TTL not enabled
func (dbc DBConfig) TTLAttribute() (string, error) {
	input := &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(dbc.TableName),
	}
	out, err := dbc.DbService.DescribeTimeToLive(context.TODO(), input)
	if err != nil {
		dbc.Tx.Error("DescribeTimeToLive failed with error: %+v", err)
		return "", err
	}
	if out.TimeToLiveDescription != nil && out.TimeToLiveDescription.TimeToLiveStatus == types.TimeToLiveStatusEnabled &&
		out.TimeToLiveDescription.AttributeName != nil {
		return *out.TimeToLiveDescription.AttributeName, nil
	}
	return "", errors.New("TTL Not properly configured")
}

// WaitForTableCreation prevents ResourceNotFoundException when DescribeTable is called right after CreateTable.
// Dynamo documentation states DescribeTable uses eventually consistent reads, so we wait for the table metadata to become available.
func (dbc DBConfig) GetStreamArn() (string, error) {
	err := dbc.WaitForTableCreation()
	if err != nil {
		dbc.Tx.Error("[GetStreamArn] Could not get streamArn, table not yet created: %+v", err)
		return "", err
	}

	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(dbc.TableName),
	}
	out, err := dbc.DbService.DescribeTable(context.TODO(), input)
	if err != nil {
		dbc.Tx.Error("[GetStreamArn] DescribeTable failed with error: %+v", err)
		return "", err
	}
	if out == nil || out.Table == nil {
		return "", nil
	}
	if out.Table.LatestStreamArn != nil {
		return *out.Table.LatestStreamArn, nil
	}
	return "", nil
}

func (dbc DBConfig) WaitForTableCreation() error {
	dbc.Tx.Info("[Dynamo][WaitForTableCreation] waiting for table [%s] to be created", dbc.TableName)
	waiter := dynamodb.NewTableExistsWaiter(dbc.DbService)
	err := waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{TableName: aws.String(dbc.TableName)}, 2*time.Minute)
	if err != nil {
		dbc.Tx.Error("[Dynamo][WaitForTableCreation] error waiting for table creation: %v", err)
	}
	return err
}

func (dbc DBConfig) WaitForTableDeletion() error {
	waiter := dynamodb.NewTableNotExistsWaiter(dbc.DbService)
	err := waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{TableName: aws.String(dbc.TableName)}, 2*time.Minute)
	if err != nil {
		dbc.Tx.Error("[Dynamo][WaitForTableDeletion] error waiting for table deletion: %v", err)
	}
	return err
}

func (dbc DBConfig) GetItemCount() (int64, error) {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(dbc.TableName),
	}
	out, err := dbc.DbService.DescribeTable(context.TODO(), input)
	if out.Table != nil && out.Table.ItemCount != nil {
		return *out.Table.ItemCount, err
	}

	return 0, err
}

func (dbc DBConfig) DeleteTable(tn string) error {
	dbc.Tx.Info("[Dynamo][DeleteTable] tableName=%+v", tn)

	input := &dynamodb.DeleteTableInput{
		TableName: aws.String(tn),
	}
	_, err := dbc.DbService.DeleteTable(context.TODO(), input)
	if err != nil {
		log.Printf("[Dynamo][DeleteTable] error calling Delete: %s", err)
	}
	return err
}

// add lambda execution context to db connection structto be able to be used for tracing
// this is not done in the init function to be able to initialize the connectino
// outside the lambda function handler
func (dbc *DBConfig) AddContext(ctx context.Context) error {
	dbc.LambdaContext = ctx
	return nil
}

func (dbc *DBConfig) SetTx(tx core.Transaction) error {
	tx.LogPrefix = "DYNAMODB"
	dbc.Tx = tx
	return nil
}

func ValidateConfig(dbc DBConfig) error {
	if dbc.LambdaContext == nil {
		return errors.New("lambda Context is Empty. Please call AddContext function after Init")
	}
	if dbc.PrimaryKey == "" {
		return errors.New("cannot have an empty PK in DB config")
	}
	return nil
}

func (dbc DBConfig) Save(prop interface{}) (interface{}, error) {
	dbc.Tx.Info("[Save] Saving Item to dynamo")
	err := ValidateConfig(dbc)
	if err != nil {
		return nil, err
	}
	av, err := attributevalue.MarshalMapWithOptions(prop, func(encoderOptions *attributevalue.EncoderOptions) {
		encoderOptions.TagKey = "json"
	})
	if err != nil {
		dbc.Tx.Error("Got error marshalling new property item:")
		dbc.Tx.Error(err.Error())
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(dbc.TableName),
	}

	_, err = dbc.DbService.PutItem(dbc.LambdaContext, input)
	if err != nil {
		dbc.Tx.Error("Got error calling PutItem:")
		dbc.Tx.Error(err.Error())
	}
	return prop, err
}

func (dbc DBConfig) Lock(pk string, sk string, lockKey string) error {
	return dbc.LockWithTTL(pk, sk, lockKey, 0)
}

// Lock item if not locked
func (dbc DBConfig) LockWithTTL(pk string, sk string, lockKey string, ttlSeconds int64) error {
	dbc.Tx.Info("[Lock] Locking item %s, %s on key %s", pk, sk, lockKey)

	//Managing TTL (used to expire the loock after a while)
	now := time.Now()
	expTime := now.Add(time.Second * time.Duration(ttlSeconds))
	ttlUpdateExpr := ""
	ttlCondition := ""
	ttlAttr := ""
	var err error
	if ttlSeconds > 0 {
		ttlAttr, err = dbc.TTLAttribute()
		if err != nil || ttlAttr == "" {
			dbc.Tx.Error("[Lock] Cannot set TTL on non TTL enabled table (%v)", err)
			return err
		}
		dbc.Tx.Debug("[Lock] creating lock with TTL %v", ttlSeconds)
		ttlUpdateExpr = ", " + ttlAttr + " = :ttlVal"
		//if TTL target date is in the past we allow the lock regardless of lock value
		ttlCondition = " OR " + ttlAttr + " < :nowVal"
	}

	err = ValidateConfig(dbc)
	if err != nil {
		return err
	}
	pkv, err := attributevalue.Marshal(pk)
	if err != nil {
		dbc.Tx.Error("[Lock] error marshalling pk: \"%s\": %w", pk, err)
		return err
	}
	key := map[string]types.AttributeValue{dbc.PrimaryKey: pkv}
	if dbc.SortKey != "" {
		skv, err := attributevalue.Marshal(sk)
		if err != nil {
			dbc.Tx.Error("[Lock] error marshalling sk: \"%s\": %w", sk, err)
			return err
		}
		key[dbc.SortKey] = skv
	}
	valueToSub, err := attributevalue.Marshal(true)
	if err != nil {
		dbc.Tx.Error("[Lock] error marshalling bool: \"%b\": %w", true, err)
		return err
	}
	value := map[string]types.AttributeValue{
		":valueToSub": valueToSub,
	}
	if ttlSeconds > 0 {
		ttlVal, err := attributevalue.Marshal(expTime.Unix())
		if err != nil {
			dbc.Tx.Error("[Lock] error marshalling expiration time: %w:", err)
			return err
		}
		value[":ttlVal"] = ttlVal
		nowVal, err := attributevalue.Marshal(now.Unix())
		if err != nil {
			dbc.Tx.Error("[Lock] error marshalling current time: %w:", err)
			return err
		}
		value[":nowVal"] = nowVal
	}

	dbc.Tx.Info("[Lock] Locking item %s, %s on key %s", pk, sk, lockKey)
	input := &dynamodb.UpdateItemInput{
		Key:                       key,
		TableName:                 aws.String(dbc.TableName),
		ConditionExpression:       aws.String(lockKey + " <> :valueToSub" + ttlCondition),
		UpdateExpression:          aws.String("SET " + lockKey + " = :valueToSub" + ttlUpdateExpr),
		ExpressionAttributeValues: value,
	}
	dbc.Tx.Debug("[Lock] UpdateItemInput=%+v", input)
	_, err = dbc.DbService.UpdateItem(dbc.LambdaContext, input)
	if err != nil {
		dbc.Tx.Error("Got error calling UpdateWithCondition:")
		dbc.Tx.Error(err.Error())
	}
	return err
}

func (dbc DBConfig) Unlock(pk string, sk string, lockKey string) error {
	err := ValidateConfig(dbc)
	if err != nil {
		return err
	}
	pkv, err := attributevalue.Marshal(pk)
	if err != nil {
		dbc.Tx.Error("[Unlock] error marshalling pk: \"%s\": %w:", pk, err)
		return err
	}
	key := map[string]types.AttributeValue{
		dbc.PrimaryKey: pkv,
	}
	if dbc.SortKey != "" {
		skv, err := attributevalue.Marshal(sk)
		if err != nil {
			dbc.Tx.Error("[Unlock] error marshalling sk: \"%s\": %w:", sk, err)
			return err
		}
		key[dbc.SortKey] = skv
	}
	valueToSub, err := attributevalue.Marshal(false)
	if err != nil {
		dbc.Tx.Error("[Unlock] error marshalling bool: \"%b\": %w:", false, err)
		return err
	}
	value := map[string]types.AttributeValue{
		":valueToSub": valueToSub,
	}
	input := &dynamodb.UpdateItemInput{
		Key:                       key,
		TableName:                 aws.String(dbc.TableName),
		ConditionExpression:       aws.String("attribute_exists(" + lockKey + ") and (" + lockKey + " <> :valueToSub)"),
		UpdateExpression:          aws.String("SET " + lockKey + " = :valueToSub"),
		ExpressionAttributeValues: value,
	}
	dbc.Tx.Debug("[Unlock] UpdateItemInput=%+v", input)
	_, err = dbc.DbService.UpdateItem(dbc.LambdaContext, input)
	if err != nil {
		dbc.Tx.Error("Got error calling UpdateWithCondition:")
		dbc.Tx.Error(err.Error())
	}
	return err
}

func (dbc DBConfig) Delete(prop interface{}) (interface{}, error) {
	av, err := attributevalue.MarshalMapWithOptions(prop, func(encoderOptions *attributevalue.EncoderOptions) {
		encoderOptions.TagKey = "json"
	})
	if err != nil {
		dbc.Tx.Error("[Delete] Error marshalling new property item: %v", err.Error())
	}
	input := &dynamodb.DeleteItemInput{
		Key:       av,
		TableName: aws.String(dbc.TableName),
	}

	_, err = dbc.DbService.DeleteItem(dbc.LambdaContext, input)
	if err != nil {
		dbc.Tx.Error("Got error calling DeetItem:")
		dbc.Tx.Error(err.Error())
	}
	return prop, err
}

// TODO: handle tables that only use a partition key, without sort key
func (dbc DBConfig) DeleteByKey(pkValue, skValue string) error {
	pkv, err := attributevalue.Marshal(pkValue)
	if err != nil {
		dbc.Tx.Error("[DeleteByKey] error marshalling pk: \"%s\": %w:", pkValue, err)
		return err
	}
	skv, err := attributevalue.Marshal(skValue)
	if err != nil {
		dbc.Tx.Error("[DeleteByKey] error marshalling sk: \"%s\": %w:", skValue, err)
		return err
	}
	input := dynamodb.DeleteItemInput{
		TableName: &dbc.TableName,
		Key: map[string]types.AttributeValue{
			dbc.PrimaryKey: pkv,
			dbc.SortKey:    skv,
		},
	}
	_, err = dbc.DbService.DeleteItem(context.TODO(), &input)
	if err != nil {
		log.Printf("[DeleteByKey] Error: %v", err)
	}
	return err
}

// TODO: Batch the FindAll
func (dbc DBConfig) DeleteAll() error {
	lastEvalKey := map[string]types.AttributeValue{}
	i := 0
	for len(lastEvalKey) > 0 || i == 0 {
		i++
		input := &dynamodb.ScanInput{
			TableName:            aws.String(dbc.TableName),
			ProjectionExpression: aws.String(dbc.PrimaryKey + "," + dbc.SortKey),
		}
		out, err := dbc.DbService.Scan(context.TODO(), input)
		if err != nil {
			dbc.Tx.Error("[DeleteAll] error colling dynamo %v", err)
			return err
		}
		lastEvalKey = out.LastEvaluatedKey
		data := []map[string]interface{}{}
		err = attributevalue.UnmarshalListOfMapsWithOptions(out.Items, &data, func(decoderOptions *attributevalue.DecoderOptions) {
			decoderOptions.TagKey = "json"
		})
		if err != nil {
			dbc.Tx.Error("[DeleteAll] Failed to unmarshall records %+v with error: %v", data, err)
		}
		toDelete := []map[string]interface{}{}
		for _, el := range data {
			toDelete = append(toDelete, map[string]interface{}{
				dbc.PrimaryKey: el[dbc.PrimaryKey],
				dbc.SortKey:    el[dbc.SortKey],
			})
		}
		err = dbc.DeleteMany(toDelete)
		if err != nil {
			dbc.Tx.Error("Error deleting all items %v", err.Error())
			return err
		}
	}
	return nil
}

// Writtes many items to a single table
func (dbc DBConfig) SaveMany(data interface{}) error {
	if dbc.TableName == "" {
		return errors.New("dynamo DB handler is missing a table name")
	}
	//Dynamo db currently limits batches to 25 items
	batches := core.Chunk(core.InterfaceSlice(data), 25)
	for i, dataArray := range batches {
		dbc.Tx.Debug("DB> Batch %i inserting %d items", i, len(dataArray))
		items := make([]types.WriteRequest, len(dataArray))
		for i, item := range dataArray {
			av, err := attributevalue.MarshalMapWithOptions(item, func(encoderOptions *attributevalue.EncoderOptions) {
				encoderOptions.TagKey = "json"
			})
			if err != nil {
				dbc.Tx.Error("[SaveMany] Error marshalling new property item: %v", err.Error())
			}
			items[i] = types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: av,
				},
			}
		}
		bwii := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				dbc.TableName: items,
			},
		}
		_, err := dbc.DbService.BatchWriteItem(context.TODO(), bwii)
		if err != nil {
			dbc.LogError(err)
			return err
		}
	}
	return nil
}

// Logging a morre descriptive error
func (dbc DBConfig) LogError(err error) {
	var provisionedThroughputExceededException *types.ProvisionedThroughputExceededException
	if errors.As(err, &provisionedThroughputExceededException) {
		dbc.Tx.Error(provisionedThroughputExceededException.Error())
		return
	}
	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		dbc.Tx.Error(resourceNotFoundException.Error())
		return
	}
	var itemCollectionSizeLimitExceededException *types.ItemCollectionSizeLimitExceededException
	if errors.As(err, &itemCollectionSizeLimitExceededException) {
		dbc.Tx.Error(itemCollectionSizeLimitExceededException.Error())
		return
	}
	var requestLimitExceeded *types.RequestLimitExceeded
	if errors.As(err, &requestLimitExceeded) {
		dbc.Tx.Error(requestLimitExceeded.Error())
		return
	}
	var internalServerError *types.InternalServerError
	if errors.As(err, &internalServerError) {
		dbc.Tx.Error(internalServerError.Error())
		return
	}

	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		dbc.Tx.Error("%s: %s", apiError.ErrorCode(), apiError.ErrorMessage())
		return
	}

	dbc.Tx.Error(err.Error())
}

func (dbc DBConfig) FindAll(data interface{}, lastEvaluatedKey map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(dbc.TableName),
	}
	if lastEvaluatedKey != nil {
		input.ExclusiveStartKey = lastEvaluatedKey
	}
	out, err := dbc.DbService.Scan(context.TODO(), input)
	if err != nil {
		dbc.Tx.Error("[FindAll] error calling dynamo %v", err)
		return nil, err
	}
	err = attributevalue.UnmarshalListOfMapsWithOptions(out.Items, data, func(decoderOptions *attributevalue.DecoderOptions) {
		decoderOptions.TagKey = "json"
	})
	if err != nil {
		dbc.Tx.Error("[FindAll] Failed to unmarshall records %+v with error: %v", data, err)
	}
	return out.LastEvaluatedKey, err
}

// Deletes many items from a single table
func (dbc DBConfig) DeleteMany(data interface{}) error {
	//Dynamo db currently limits batches to 25 items
	batches := core.Chunk(core.InterfaceSlice(data), 25)
	for i, dataArray := range batches {

		dbc.Tx.Info("Batch %d deleting: %+v", i, dataArray)
		items := make([]types.WriteRequest, len(dataArray))
		for i, item := range dataArray {
			av, err := attributevalue.MarshalMapWithOptions(item, func(encoderOptions *attributevalue.EncoderOptions) {
				encoderOptions.TagKey = "json"
			})
			if err != nil {
				dbc.Tx.Error("[DeleteMany] Error marshalling new property item: %v", err.Error())
			}
			items[i] = types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: av,
				},
			}
		}
		bwii := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				dbc.TableName: items,
			},
		}

		_, err := dbc.DbService.BatchWriteItem(context.TODO(), bwii)
		if err != nil {
			dbc.LogError(err)
			return err
		}
	}
	return nil
}

func (dbc DBConfig) GetMany(records []DynamoRecord, data interface{}) error {
	//Dynamo db currently limits read batches to 100 items OR up to 16 MB of data,
	//https: //docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/#DynamoDB.BatchGetItem
	allItems := []map[string]types.AttributeValue{}
	batches := core.Chunk(core.InterfaceSlice(records), 100)
	for i, dataArray := range batches {

		dbc.Tx.Info("DB> Batch %v fetching: %v items", i, len(dataArray))
		items := make([]map[string]types.AttributeValue, len(dataArray))
		for i, item := range dataArray {
			rec := item.(DynamoRecord)
			pkv, err := attributevalue.Marshal(rec.Pk)
			if err != nil {
				dbc.Tx.Error("[GetMany] error marshalling pk: \"%s\": %w:", rec.Pk, err)
				return err
			}
			av := map[string]types.AttributeValue{
				dbc.PrimaryKey: pkv,
			}
			if rec.Sk != "" {
				skv, err := attributevalue.Marshal(rec.Sk)
				if err != nil {
					dbc.Tx.Error("[GetMany] error marshalling sk: \"%s\": %w:", rec.Sk, err)
					return err
				}
				av[dbc.SortKey] = skv
			}
			items[i] = av
		}

		bgii := &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				dbc.TableName: {
					Keys: items,
				},
			},
		}

		bgio, err := dbc.DbService.BatchGetItem(context.TODO(), bgii)
		if err != nil {
			dbc.LogError(err)
			return err
		}
		dbc.Tx.Debug("DB> Batch %v found: %v items", i, len(bgio.Responses[dbc.TableName]))
		allItems = append(allItems, bgio.Responses[dbc.TableName]...)
	}
	dbc.Tx.Info("DB> Found %v items in total", len(allItems))
	err := attributevalue.UnmarshalListOfMapsWithOptions(allItems, data, func(decoderOptions *attributevalue.DecoderOptions) {
		decoderOptions.TagKey = "json"
	})
	if err != nil {
		dbc.Tx.Error("DB:GetMany> Error while unmarshalling")
		dbc.Tx.Error(err.Error())
		return err
	}
	return nil
}

func (dbc DBConfig) GetWithOptions(pk string, sk string, data interface{}, getOps QueryOptions) error {
	err := ValidateConfig(dbc)
	if err != nil {
		return err
	}
	pkv, err := attributevalue.Marshal(pk)
	if err != nil {
		dbc.Tx.Error("[Get] error marshalling pk: \"%s\": %w:", pk, err)
		return err
	}
	av := map[string]types.AttributeValue{
		dbc.PrimaryKey: pkv,
	}
	if sk != "" {
		skv, err := attributevalue.Marshal(sk)
		if err != nil {
			dbc.Tx.Error("[Get] error marshalling sk: \"%s\": %w:", sk, err)
			return err
		}
		av[dbc.SortKey] = skv
	}

	gii := &dynamodb.GetItemInput{
		TableName: aws.String(dbc.TableName),
		Key:       av,
	}
	if getOps.Consistent {
		gii.ConsistentRead = aws.Bool(true)
	}
	dbc.Tx.Debug("DB> DynamoDB  Get rq %+v", gii)
	result, err := dbc.DbService.GetItem(dbc.LambdaContext, gii)
	if err != nil {
		dbc.Tx.Error("[dynamo][get] GetItem error: %v", err)
		return err
	}

	if getOps.NoObjectNoError {
		if result.Item == nil {
			return nil
		}
	}

	err = attributevalue.UnmarshalMapWithOptions(result.Item, data, func(decoderOptions *attributevalue.DecoderOptions) {
		decoderOptions.TagKey = "json"
	})
	if err != nil {
		dbc.Tx.Error("[dynamo][get]  Failed to unmarshal Record: %v", err)
	}
	return err
}

func (dbc DBConfig) ConsistentGet(pk string, sk string, data interface{}) error {
	return dbc.GetWithOptions(pk, sk, data, QueryOptions{Consistent: true})
}

func (dbc DBConfig) ConsistentSafeGet(pk string, sk string, data interface{}) error {
	return dbc.GetWithOptions(pk, sk, data, QueryOptions{Consistent: true, NoObjectNoError: true})
}

func (dbc DBConfig) Get(pk string, sk string, data interface{}) error {
	return dbc.GetWithOptions(pk, sk, data, QueryOptions{})
}

func (dbc DBConfig) FindStartingWith(pk string, value string, data interface{}) error {
	return dbc.FindStartingWithAndFilterWithIndex(pk, value, data, QueryOptions{})
}

func (dbc DBConfig) FindStartingWithAndFilter(pk string, value string, data interface{}, filter DynamoFilterExpression) error {
	return dbc.FindStartingWithAndFilterWithIndex(pk, value, data, QueryOptions{Filter: filter})
}
func (dbc DBConfig) FindStartingWithByGsi(pk string, value string, data interface{}, index DynamoIndex) error {
	return dbc.FindStartingWithAndFilterWithIndex(pk, value, data, QueryOptions{Index: index})
}

func (dbc DBConfig) FindStartingWithAndFilterWithIndex(pk string, value string, data interface{}, queryOptions QueryOptions) error {
	dbc.Tx.Debug("[FindStartingWithAndFilterWithIndex] Pk: %v, Sk start with: %s and queryOptions %+v", pk, value, queryOptions)
	filter := queryOptions.Filter
	index := queryOptions.Index
	reverseOrder := queryOptions.ReverseOrder
	paginOptions := queryOptions.PaginOptions
	projExpr := strings.Join(queryOptions.ProjectionExpression.Projection, ", ")

	pkName := dbc.PrimaryKey
	skName := dbc.SortKey
	//If search done per index, we change pk and sk to the index ones
	if index.Name != "" {
		pkName = index.Pk
		skName = index.Sk
	}
	pkv, err := attributevalue.Marshal(pk)
	if err != nil {
		dbc.Tx.Error("[FindStartingWithAndFilterWithIndex] error marshalling pk: \"%s\": %w:", pk, err)
		return err
	}
	val, err := attributevalue.Marshal(value)
	if err != nil {
		dbc.Tx.Error("[FindStartingWithAndFilterWithIndex] error marshalling value: \"%s\": %w:", value, err)
		return err
	}
	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String(dbc.TableName),
		KeyConditions: map[string]types.Condition{
			pkName: {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{pkv},
			},
			skName: {
				ComparisonOperator: types.ComparisonOperatorBeginsWith,
				AttributeValueList: []types.AttributeValue{val},
			},
		},
	}

	//Optional index to search in
	if index.Name != "" {
		queryInput.IndexName = aws.String(index.Name)
	}

	if reverseOrder {
		queryInput.ScanIndexForward = aws.Bool(false)
	}

	if projExpr != "" {
		queryInput.ProjectionExpression = aws.String(projExpr)
	}

	//Building Filter expression
	if filter.HasFilter() {
		expr, err := dbc.BuildQueryFilter(filter)
		if err != nil {
			dbc.Tx.Error("DB:FindStartingWithAndFilter> Error in filter expression: %+v", err)
			return err
		}
		queryInput.ExpressionAttributeNames = expr.Names()
		queryInput.ExpressionAttributeValues = expr.Values()
		queryInput.FilterExpression = expr.Filter()
	}
	//Run query with support for pagination
	allItems, err := dbc.RunQuery(queryInput, paginOptions)
	if err != nil {
		dbc.Tx.Error("[FindStartingWithAndFilterWithIndex] Run query failed: %v", err)
		return err
	}
	//Unparshal result
	err = attributevalue.UnmarshalListOfMapsWithOptions(allItems, data, func(decoderOptions *attributevalue.DecoderOptions) {
		decoderOptions.TagKey = "json"
	})
	if err != nil {
		dbc.Tx.Error("[FindStartingWithAndFilterWithIndex] Failed to unmarshal Record: %v", err)
	}
	return err
}

func (dbc DBConfig) BuildQueryFilter(filter DynamoFilterExpression) (expression.Expression, error) {
	builder := expression.NewBuilder()
	if filter.HasBeginsWith() {
		dbc.Tx.Debug("[BuildQueryFilter] Adding Filter expression: %+v", filter.BeginsWith[0])
		condBuilder := expression.Name(filter.BeginsWith[0].Key).BeginsWith(filter.BeginsWith[0].Value)
		builder = builder.WithFilter(condBuilder)
	}
	if filter.HasContains() {
		dbc.Tx.Debug("[BuildQueryFilter] Adding Filter expression: %+v", filter.Contains[0])
		condBuilder := expression.Name(filter.Contains[0].Key).Contains(filter.Contains[0].Value)
		builder = builder.WithFilter(condBuilder)
	}
	if filter.HasEqual() {
		dbc.Tx.Debug("[BuildQueryFilter] Adding Filter expression: %+v", filter.EqualBool[0])
		condBuilder := expression.Name(filter.EqualBool[0].Key).Equal(expression.Value(filter.EqualBool[0].Value))
		builder = builder.WithFilter(condBuilder)
	}
	expr, err := builder.Build()
	return expr, err
}

func (dbc DBConfig) FindGreaterThanByGsi(pk string, value int64, data interface{}, queryOptions QueryOptions) error {
	return dbc.FindGreaterThanAndFilterWithIndex(pk, value, data, queryOptions)
}

func (dbc DBConfig) FindGreaterThan(pk string, value int64, data interface{}, queryOptions QueryOptions) error {
	return dbc.FindGreaterThanAndFilterWithIndex(pk, value, data, queryOptions)
}

func (dbc DBConfig) FindGreaterThanAndFilterWithIndex(pk string, value int64, data interface{}, queryOptions QueryOptions) error {
	dbc.Tx.Debug("[FindGreaterThanAndFilterWithIndex] Pk: %v, Sk greater than: %d with QueryOptions: %+v", pk, value, queryOptions)
	filter := queryOptions.Filter
	index := queryOptions.Index
	paginOptions := queryOptions.PaginOptions
	reverseOrder := queryOptions.ReverseOrder
	projExpr := strings.Join(queryOptions.ProjectionExpression.Projection, ", ")

	pkName := dbc.PrimaryKey
	skName := dbc.SortKey
	//If search done per index, we change pk and sk to the index ones
	if index.Name != "" {
		pkName = index.Pk
		skName = index.Sk
	}
	pkv, err := attributevalue.Marshal(pk)
	if err != nil {
		dbc.Tx.Error("[FindGreaterThanAndFilterWithIndex] error marshalling pk: \"%s\": %w:", pk, err)
		return err
	}
	val, err := attributevalue.Marshal(value)
	if err != nil {
		dbc.Tx.Error("[FindGreaterThanAndFilterWithIndex] error marshalling value: \"%s\": %w:", value, err)
		return err
	}
	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String(dbc.TableName),
		KeyConditions: map[string]types.Condition{
			pkName: {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{pkv},
			},
			skName: {
				ComparisonOperator: types.ComparisonOperatorGt,
				AttributeValueList: []types.AttributeValue{val},
			},
		},
	}

	//Optional index to search in
	if index.Name != "" {
		queryInput.IndexName = aws.String(index.Name)
	}

	if reverseOrder {
		queryInput.ScanIndexForward = aws.Bool(false)
	}

	if projExpr != "" {
		queryInput.ProjectionExpression = aws.String(projExpr)
	}

	//Building Filter expression
	if filter.HasFilter() {
		expr, err := dbc.BuildQueryFilter(filter)
		if err != nil {
			dbc.Tx.Error("[FindGreaterThanAndFilterWithIndex] Error in filter expression: %+v", err)
			return err
		}
		queryInput.ExpressionAttributeNames = expr.Names()
		queryInput.ExpressionAttributeValues = expr.Values()
		queryInput.FilterExpression = expr.Filter()
	}

	dbc.Tx.Debug("[FindGreaterThanAndFilterWithIndex] rq: %+v", queryInput)
	//run query
	allItems, err := dbc.RunQuery(queryInput, paginOptions)
	if err != nil {
		dbc.Tx.Error("[FindGreaterThanAndFilterWithIndex] Run query failed: %v", err)
		return err
	}
	//unmarshal data
	err = attributevalue.UnmarshalListOfMapsWithOptions(allItems, data, func(decoderOptions *attributevalue.DecoderOptions) {
		decoderOptions.TagKey = "json"
	})
	if err != nil {
		dbc.Tx.Error("[FindGreaterThanAndFilterWithIndex] Failed to unmarshal Record: %v", err)
	}
	return err
}

func (dbc DBConfig) FindByPk(pk string, data interface{}) error {
	pkv, err := attributevalue.Marshal(pk)
	if err != nil {
		dbc.Tx.Error("[FindByPk] error marshalling pk: \"%s\": %w:", pk, err)
		return err
	}
	results, err := dbc.RunQuery(&dynamodb.QueryInput{
		TableName: aws.String(dbc.TableName),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": pkv,
		},
		KeyConditionExpression: aws.String(fmt.Sprintf("%s=:pk", dbc.PrimaryKey)),
	}, DynamoPaginationOptions{})
	if err != nil {
		return err
	}

	err = attributevalue.UnmarshalListOfMapsWithOptions(results, &data, func(decoderOptions *attributevalue.DecoderOptions) {
		decoderOptions.TagKey = "json"
	})
	if err != nil {
		return err
	}
	return nil
}

func (dbc DBConfig) RunQuery(
	queryInput *dynamodb.QueryInput,
	paginOptions DynamoPaginationOptions,
) ([]map[string]types.AttributeValue, error) {
	dbc.Tx.Debug("[RunQuery] Request: %+v", queryInput)

	//Optional pagination
	if paginOptions.PageSize > 0 {
		queryInput.Limit = aws.Int32(paginOptions.PageSize)
	}

	allItems := []map[string]types.AttributeValue{}
	var page int64 = 0
	var result, err = dbc.DbService.Query(dbc.LambdaContext, queryInput)
	if err != nil {
		dbc.Tx.Error("[RunQuery] error running Query: %v", err)
		return allItems, err
	}
	dbc.Tx.Debug("[RunQuery] Response contains %v items", len(result.Items))
	for _, item := range result.Items {
		if paginOptions.PageSize == 0 || page == paginOptions.Page {
			allItems = append(allItems, item)
		}
	}
	//if paginated response
	for result.LastEvaluatedKey != nil && (paginOptions.PageSize == 0 || page < paginOptions.Page) {
		page += 1
		var lastEvalKey map[string]string
		err = attributevalue.UnmarshalMap(result.LastEvaluatedKey, &lastEvalKey)
		if err != nil {
			dbc.Tx.Error("[RunQuery] error unmarshalling last evaluated key: %v", err)
			return allItems, err
		}
		dbc.Tx.Debug(
			"[RunQuery] Response is paginated (requested page: %v, current page: %v,pageSize:%v)). LastEvaluatedKey: %v",
			paginOptions.Page,
			page,
			paginOptions.PageSize,
			lastEvalKey,
		)
		queryInput.ExclusiveStartKey = result.LastEvaluatedKey
		result, err = dbc.DbService.Query(dbc.LambdaContext, queryInput)
		dbc.Tx.Debug("[RunQuery] Response contains %v items", len(result.Items))
		if err != nil {
			dbc.Tx.Error("[RunQuery] error running query: %v", err)
			return allItems, err
		}
		for _, item := range result.Items {
			if paginOptions.PageSize == 0 || page == paginOptions.Page {
				allItems = append(allItems, item)
			}
		}
	}
	return allItems, nil
}

func (dbc DBConfig) FindByGsi(value string, indexName string, indexPk string, data interface{}, options FindByGsiOptions) error {
	val, err := attributevalue.Marshal(value)
	if err != nil {
		dbc.Tx.Error("[FindByGsi] error marshalling value: \"%s\": %w:", value, err)
		return err
	}
	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String(dbc.TableName),
		IndexName: aws.String(indexName),
		KeyConditions: map[string]types.Condition{
			indexPk: {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{val},
			},
		},
	}
	if options.Limit != 0 {
		queryInput.Limit = aws.Int32(options.Limit)
	}
	dbc.Tx.Debug("DB:FindByGsi Request: %+v", queryInput)

	result, err := dbc.DbService.Query(dbc.LambdaContext, queryInput)
	if err != nil {
		dbc.Tx.Error("NOT FOUND")
		dbc.Tx.Error(err.Error())
		return err
	}
	dbc.Tx.Debug("DB:FindByGsi Result: %+v", result)

	err = attributevalue.UnmarshalListOfMapsWithOptions(result.Items, data, func(decoderOptions *attributevalue.DecoderOptions) {
		decoderOptions.TagKey = "json"
	})
	if err != nil {
		dbc.Tx.Error("[FindByGsi] Failed to unmarshal Record: %v", err)
	}
	return err
}

func (dbc DBConfig) UpdateItems(pk string, sk string, indextovalue map[string]interface{}) error {
	pkv, err := attributevalue.Marshal(pk)
	if err != nil {
		dbc.Tx.Error("[UpdateItems] error marshalling pk: \"%s\": %w:", pk, err)
		return err
	}
	skv, err := attributevalue.Marshal(sk)
	if err != nil {
		dbc.Tx.Error("[UpdateItems] error marshalling sk: \"%s\": %w:", sk, err)
		return err
	}
	key := map[string]types.AttributeValue{
		dbc.PrimaryKey: pkv,
		dbc.SortKey:    skv,
	}

	setExpressions := []string{}
	expressionAttributeValues := map[string]types.AttributeValue{}
	expressionAttributeNames := map[string]string{}

	for key, value := range indextovalue {
		placeholder := fmt.Sprintf(":%v", key)
		expressionAttributeNames[fmt.Sprintf("#%s", key)] = key
		setExpressions = append(setExpressions, fmt.Sprintf("#%v=%v", key, placeholder))

		val, err := attributevalue.MarshalWithOptions(value, func(encoderOptions *attributevalue.EncoderOptions) {
			encoderOptions.TagKey = "json"
		})
		if err != nil {
			dbc.Tx.Error("[UpdateItems] error marshalling value: \"%v\": %w:", value, err)
			return err
		}
		expressionAttributeValues[placeholder] = val
	}

	_, err = dbc.DbService.UpdateItem(dbc.LambdaContext, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(dbc.TableName),
		Key:                       key,
		UpdateExpression:          aws.String(fmt.Sprintf("SET %v", strings.Join(setExpressions, ","))),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
	})
	return err
}

func IsItemNotFoundError(err error) bool {
	var indexNotFoundException *types.IndexNotFoundException
	return errors.As(err, &indexNotFoundException)
}
