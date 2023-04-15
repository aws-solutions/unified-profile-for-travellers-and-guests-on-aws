package db

import (
	"context"
	"errors"
	"log"
	"strconv"
	core "tah/core/core"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
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
	NewImage map[string]*dynamodb.AttributeValue `json:"NewImage"`
	OldImage map[string]*dynamodb.AttributeValue `json:"OldImage"`
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
	Get(pk string, sk string, data interface{}) error
}

type DBConfig struct {
	DbService     *dynamodb.DynamoDB
	PrimaryKey    string
	SortKey       string
	TableName     string
	LambdaContext context.Context
	Tx            core.Transaction
}

type DynamoFilterExpression struct {
	BeginsWith []DynamoFilterCondition
	Contains   []DynamoFilterCondition
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

type DynamoPaginationOptions struct {
	Page     int64
	PageSize int64
}

type QueryOptions struct {
	Filter       DynamoFilterExpression
	Index        DynamoIndex
	PaginOptions DynamoPaginationOptions
	ReverseOrder bool
}

type TableOptions struct {
	TTLAttribute string
}

func (exp DynamoFilterExpression) HasFilter() bool {
	return exp.HasBeginsWith() || exp.HasContains()
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

//init setup teh session and define table name, primary key and sort key
func Init(tn string, pk string, sk string) DBConfig {
	if pk == "" {
		log.Printf("[CORE][DB] WARNING: empty PK provided")
	}
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	dbSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create DynamoDB client
	return DBConfig{
		DbService:     dynamodb.New(dbSession),
		PrimaryKey:    pk,
		SortKey:       sk,
		TableName:     tn,
		LambdaContext: context.Background(),
	}

}

func InitWithNewTable(tn string, pk string, sk string) (DBConfig, error) {
	return InitWithNewTableAndTTL(tn, pk, sk, "")
}

func InitWithNewTableAndTTL(tn string, pk string, sk string, ttlAttr string) (DBConfig, error) {
	if pk == "" {
		log.Printf("[CORE][DB] WARNING: empty PK provided")
	}
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	dbSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create DynamoDB client
	cfg := DBConfig{
		DbService:     dynamodb.New(dbSession),
		PrimaryKey:    pk,
		SortKey:       sk,
		TableName:     tn,
		LambdaContext: context.Background(),
	}
	options := TableOptions{}
	if ttlAttr != "" {
		options.TTLAttribute = ttlAttr
	}
	err := cfg.CreateTableWithOptions(tn, pk, sk, options)
	if err != nil {
		log.Printf("[Dynamo][InitWithNewTable] error calling CreateTable: %s", err)
	}
	return cfg, err
}

func (dbc DBConfig) CreateTable(tn string, pk string, sk string) error {
	return dbc.CreateTableWithOptions(tn, pk, sk, TableOptions{})
}

func (dbc DBConfig) CreateTableWithOptions(tn string, pk string, sk string, options TableOptions) error {
	attrsDef := []*dynamodb.AttributeDefinition{
		{
			AttributeName: aws.String(pk),
			AttributeType: aws.String("S"),
		},
	}
	keyEl := []*dynamodb.KeySchemaElement{
		{
			AttributeName: aws.String(pk),
			KeyType:       aws.String("HASH"),
		},
	}
	if sk != "" {
		attrsDef = append(attrsDef, &dynamodb.AttributeDefinition{
			AttributeName: aws.String(sk),
			AttributeType: aws.String("S"),
		})
		keyEl = append(keyEl, &dynamodb.KeySchemaElement{
			AttributeName: aws.String(sk),
			KeyType:       aws.String("RANGE"),
		})
	}
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: attrsDef,
		KeySchema:            keyEl,
		TableName:            aws.String(tn),
		BillingMode:          aws.String("PAY_PER_REQUEST"),
	}
	dbc.Tx.Log("[Dynamo][CreateTable] Create table input:  %+v", input)
	_, err := dbc.DbService.CreateTable(input)
	if err != nil {
		dbc.Tx.Log("[Dynamo][CreateTable] error calling CreateTable: %s", err)
	}
	if options.TTLAttribute != "" {
		dbc.Tx.Log("[Dynamo][CreateTable] Enabling time to live")
		dbc.WaitForTableCreation()
		dbc.TableName = tn
		err = dbc.EnabledTTl(options.TTLAttribute)
	}

	return err
}

func (dbc DBConfig) EnabledTTl(attr string) error {
	dbc.Tx.Log("[Dynamo][EnabledTTl] Enabling time to live")
	existingAttr, _ := dbc.TTLAttribute()
	if existingAttr == attr {
		dbc.Tx.Log("[Dynamo][EnabledTTl] TTL is already enabled on attribute: %s, do nothing", existingAttr)
		return nil
	}
	input := &dynamodb.UpdateTimeToLiveInput{
		TableName: aws.String(dbc.TableName),
		TimeToLiveSpecification: &dynamodb.TimeToLiveSpecification{
			AttributeName: aws.String(attr),
			Enabled:       aws.Bool(true),
		},
	}
	dbc.Tx.Log("[Dynamo][EnabledTTl] Create table input:  %+v", input)
	_, err := dbc.DbService.UpdateTimeToLive(input)
	return err
}

//return the TTL attribute name fo rthe table or an erroor if TTL not enabled
func (dbc DBConfig) TTLAttribute() (string, error) {
	input := &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(dbc.TableName),
	}
	out, err := dbc.DbService.DescribeTimeToLive(input)
	if err != nil {
		dbc.Tx.Log("DescribeTimeToLive failed with error: %+v", err)
		return "", err
	}
	if out.TimeToLiveDescription != nil && *out.TimeToLiveDescription.TimeToLiveStatus == "ENABLED" && out.TimeToLiveDescription.AttributeName != nil {
		return *out.TimeToLiveDescription.AttributeName, nil
	}
	return "", errors.New("TTL Not properly configured")
}

func (dbc DBConfig) WaitForTableCreation() error {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(dbc.TableName),
	}
	i := 0
	for {
		out, err := dbc.DbService.DescribeTable(input)
		if err != nil {
			return err
		}
		if *out.Table.TableStatus == "ACTIVE" || i >= 24 {
			break
		} else {
			log.Printf("[Dynamo][CreateTable] table status is : %s, waiting 5 seconds", *out.Table.TableStatus)
		}
		time.Sleep(5 * time.Second)
		i += 1
	}
	return nil
}

func (dbc DBConfig) DeleteTable(tn string) error {
	dbc.Tx.Log("[Dynamo][DeleteTable] tableNaem=%+v", tn)

	input := &dynamodb.DeleteTableInput{
		TableName: aws.String(tn),
	}
	_, err := dbc.DbService.DeleteTable(input)
	if err != nil {
		log.Printf("[Dynamo][DeleteTable] error calling Delete: %s", err)
	}
	return err
}

//add lambda execution context to db connection structto be able to be used for tracing
//this is not done in the init function to be able to initialize the connectino
//outside the lambda function handler
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
		return errors.New("Lambda Context is Empty. Please call AddContext function after Init")
	}
	if dbc.PrimaryKey == "" {
		return errors.New("Cannot have an empty PK in DB config")
	}
	return nil
}

func (dbc DBConfig) Save(prop interface{}) (interface{}, error) {
	dbc.Tx.Log("[Save] Saving Item to dynamo")
	err := ValidateConfig(dbc)
	if err != nil {
		return nil, err
	}
	av, err := dynamodbattribute.MarshalMap(prop)
	if err != nil {
		dbc.Tx.Log("Got error marshall	ing new property item:")
		dbc.Tx.Log(err.Error())
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(dbc.TableName),
	}

	_, err = dbc.DbService.PutItemWithContext(dbc.LambdaContext, input)
	if err != nil {
		dbc.Tx.Log("Got error calling PutItem:")
		dbc.Tx.Log(err.Error())
	}
	return prop, err
}

func (dbc DBConfig) Lock(pk string, sk string, lockKey string) error {
	return dbc.LockWithTTL(pk, sk, lockKey, 0)
}

//Lock item if not locked
func (dbc DBConfig) LockWithTTL(pk string, sk string, lockKey string, ttlSeconds int64) error {
	dbc.Tx.Log("[Lock] Locking item %s, %s on key %s", pk, sk, lockKey)

	//Managing TTL (used to expire the loock after a while)
	now := time.Now()
	expTime := now.Add(time.Second * time.Duration(ttlSeconds))
	ttlUpdateExpr := ""
	ttlCondition := ""
	ttlAttr := ""
	var err0 error
	if ttlSeconds > 0 {
		ttlAttr, err0 = dbc.TTLAttribute()
		if err0 != nil || ttlAttr == "" {
			dbc.Tx.Log("[Lock] Cannot set TTL on non TTL enabled table (%v)", err0)
			return err0
		}
		dbc.Tx.Log("[Lock] creating lock with TTL %v", ttlSeconds)
		ttlUpdateExpr = ", " + ttlAttr + " = :ttlVal"
		//if TTL target date is in the past we allow the lock regardless of lock value
		ttlCondition = " OR " + ttlAttr + " < :nowVal"
	}

	err := ValidateConfig(dbc)
	if err != nil {
		return err
	}
	key := map[string]*dynamodb.AttributeValue{
		dbc.PrimaryKey: {
			S: aws.String(pk),
		},
	}
	if dbc.SortKey != "" {
		key[dbc.SortKey] = &dynamodb.AttributeValue{
			S: aws.String(sk),
		}
	}
	value := map[string]*dynamodb.AttributeValue{
		":valueToSub": {
			BOOL: aws.Bool(true),
		},
	}
	if ttlSeconds > 0 {
		value[":ttlVal"] = &dynamodb.AttributeValue{
			N: aws.String(strconv.FormatInt(expTime.Unix(), 10)),
		}
		value[":nowVal"] = &dynamodb.AttributeValue{
			N: aws.String(strconv.FormatInt(now.Unix(), 10)),
		}
	}

	dbc.Tx.Log("[Lock] Locking item %s, %s on key %s", pk, sk, lockKey)
	input := &dynamodb.UpdateItemInput{
		Key:       key,
		TableName: aws.String(dbc.TableName),
		//ConditionExpression:       aws.String("attribute_exists(" + lockKey + ") and (" + lockKey + " <> :valueToSub)"),
		ConditionExpression:       aws.String(lockKey + " <> :valueToSub" + ttlCondition),
		UpdateExpression:          aws.String("SET " + lockKey + " = :valueToSub" + ttlUpdateExpr),
		ExpressionAttributeValues: value,
	}
	dbc.Tx.Log("[Lock] UpdateItemInput=%+v", input)
	_, err = dbc.DbService.UpdateItemWithContext(dbc.LambdaContext, input)
	if err != nil {
		dbc.Tx.Log("Got error calling UpdateWithCondition:")
		dbc.Tx.Log(err.Error())
	}
	return err
}

func (dbc DBConfig) Unlock(pk string, sk string, lockKey string) error {
	err := ValidateConfig(dbc)
	if err != nil {
		return err
	}
	key := map[string]*dynamodb.AttributeValue{
		dbc.PrimaryKey: {
			S: aws.String(pk),
		},
	}
	if dbc.SortKey != "" {
		key[dbc.SortKey] = &dynamodb.AttributeValue{
			S: aws.String(sk),
		}
	}
	value := map[string]*dynamodb.AttributeValue{
		":valueToSub": {
			BOOL: aws.Bool(false),
		},
	}
	input := &dynamodb.UpdateItemInput{
		Key:                       key,
		TableName:                 aws.String(dbc.TableName),
		ConditionExpression:       aws.String("attribute_exists(" + lockKey + ") and (" + lockKey + " <> :valueToSub)"),
		UpdateExpression:          aws.String("SET " + lockKey + " = :valueToSub"),
		ExpressionAttributeValues: value,
	}
	dbc.Tx.Log("[Lock] UpdateItemInput=%+v", input)
	_, err = dbc.DbService.UpdateItemWithContext(dbc.LambdaContext, input)
	if err != nil {
		dbc.Tx.Log("Got error calling UpdateWithCondition:")
		dbc.Tx.Log(err.Error())
	}
	return err
}

func (dbc DBConfig) Delete(prop interface{}) (interface{}, error) {
	av, err := dynamodbattribute.MarshalMap(prop)
	if err != nil {
		dbc.Tx.Log("[Delete] Error marshalling new property item: %v", err.Error())
	}
	input := &dynamodb.DeleteItemInput{
		Key:       av,
		TableName: aws.String(dbc.TableName),
	}

	_, err = dbc.DbService.DeleteItemWithContext(dbc.LambdaContext, input)
	if err != nil {
		dbc.Tx.Log("Got error calling DeetItem:")
		dbc.Tx.Log(err.Error())
	}
	return prop, err
}

//Writtes many items to a single table
func (dbc DBConfig) SaveMany(data interface{}) error {
	//Dynamo db currently limits batches to 25 items
	batches := core.Chunk(core.InterfaceSlice(data), 25)
	for i, dataArray := range batches {
		dbc.Tx.Log("DB> Batch %i inserting: %+v", i, dataArray)
		items := make([]*dynamodb.WriteRequest, len(dataArray), len(dataArray))
		for i, item := range dataArray {
			av, err := dynamodbattribute.MarshalMap(item)
			if err != nil {
				dbc.Tx.Log("[SaveMany] Error marshalling new property item: %v", err.Error())
			}
			items[i] = &dynamodb.WriteRequest{
				PutRequest: &dynamodb.PutRequest{
					Item: av,
				},
			}
		}
		bwii := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]*dynamodb.WriteRequest{
				dbc.TableName: items,
			},
		}
		_, err := dbc.DbService.BatchWriteItem(bwii)
		if err != nil {
			dbc.LogError(err)
			return err
		}
	}
	return nil
}

//Logging a morre descriptive error
func (dbc DBConfig) LogError(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeProvisionedThroughputExceededException:
			dbc.Tx.Log(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
		case dynamodb.ErrCodeResourceNotFoundException:
			dbc.Tx.Log(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
		case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
			dbc.Tx.Log(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
		case dynamodb.ErrCodeRequestLimitExceeded:
			dbc.Tx.Log(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
		case dynamodb.ErrCodeInternalServerError:
			dbc.Tx.Log(dynamodb.ErrCodeInternalServerError, aerr.Error())
		default:
			dbc.Tx.Log(aerr.Error())
		}
	} else {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		dbc.Tx.Log(err.Error())
	}
}

func (dbc DBConfig) FindAll(data interface{}) error {
	input := &dynamodb.ScanInput{
		TableName: aws.String(dbc.TableName),
	}
	out, err := dbc.DbService.Scan(input)
	if err != nil {
		dbc.Tx.Log("[FindAll] error colling dynamo %v", err)
		return err
	}
	err = dynamodbattribute.UnmarshalListOfMaps(out.Items, data)
	if err != nil {
		dbc.Tx.Log("[FindAll] Failed to unmarshall records %+v with error: %v", data, err)
	}
	return err
}

//Deletes many items to a single table
func (dbc DBConfig) DeleteMany(data interface{}) error {
	//Dynamo db currently limits batches to 25 items
	batches := core.Chunk(core.InterfaceSlice(data), 25)
	for i, dataArray := range batches {

		dbc.Tx.Log("Batch %d deleting: %+v", i, dataArray)
		items := make([]*dynamodb.WriteRequest, len(dataArray), len(dataArray))
		for i, item := range dataArray {
			av, err := dynamodbattribute.MarshalMap(item)
			if err != nil {
				dbc.Tx.Log("[DeleteMany] Error marshalling new property item: %v", err.Error())
			}
			items[i] = &dynamodb.WriteRequest{
				DeleteRequest: &dynamodb.DeleteRequest{
					Key: av,
				},
			}
		}
		bwii := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]*dynamodb.WriteRequest{
				dbc.TableName: items,
			},
		}

		_, err := dbc.DbService.BatchWriteItem(bwii)
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
	allItems := []map[string]*dynamodb.AttributeValue{}
	batches := core.Chunk(core.InterfaceSlice(records), 100)
	for i, dataArray := range batches {

		dbc.Tx.Log("DB> Batch %v fetching: %v items", i, len(dataArray))
		items := make([]map[string]*dynamodb.AttributeValue, len(dataArray), len(dataArray))
		for i, item := range dataArray {
			rec := item.(DynamoRecord)
			av := map[string]*dynamodb.AttributeValue{
				dbc.PrimaryKey: {
					S: aws.String(rec.Pk),
				},
			}
			if rec.Sk != "" {
				av[dbc.SortKey] = &dynamodb.AttributeValue{
					S: aws.String(rec.Sk),
				}
			}
			items[i] = av
		}

		bgii := &dynamodb.BatchGetItemInput{
			RequestItems: map[string]*dynamodb.KeysAndAttributes{
				dbc.TableName: &dynamodb.KeysAndAttributes{
					Keys: items,
				},
			},
		}
		//dbc.Tx.Log("DB> DynamoDB Batch Get rq composed of %v items", bgii)

		bgio, err := dbc.DbService.BatchGetItem(bgii)
		if err != nil {
			dbc.LogError(err)
			return err
		}
		//dbc.Tx.Log("DB> DynamoDB Batch Get rs %+v", bgio)
		dbc.Tx.Log("DB> Batch %v found: %v items", i, len(bgio.Responses[dbc.TableName]))
		for _, item := range bgio.Responses[dbc.TableName] {
			allItems = append(allItems, item)
		}
	}
	dbc.Tx.Log("DB> Found %v items in total", len(allItems))
	err := dynamodbattribute.UnmarshalListOfMaps(allItems, data)
	if err != nil {
		dbc.Tx.Log("DB:GetMany> Error while unmarshalling")
		dbc.Tx.Log(err.Error())
		return err
	}
	//dbc.Tx.Log("DB> DynamoDB Batch Get unmarchalled rs %+v", data)
	return nil
}

func (dbc DBConfig) Get(pk string, sk string, data interface{}) error {
	err := ValidateConfig(dbc)
	if err != nil {
		return err
	}
	av := map[string]*dynamodb.AttributeValue{
		dbc.PrimaryKey: {
			S: aws.String(pk),
		},
	}
	if sk != "" {
		av[dbc.SortKey] = &dynamodb.AttributeValue{
			S: aws.String(sk),
		}
	}

	gii := &dynamodb.GetItemInput{
		TableName: aws.String(dbc.TableName),
		Key:       av,
	}
	dbc.Tx.Log("DB> DynamoDB  Get rq %+v", gii)
	result, err := dbc.DbService.GetItemWithContext(dbc.LambdaContext, gii)
	if err != nil {
		dbc.Tx.Log("NOT FOUND")
		dbc.Tx.Log(err.Error())
		return err
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, data)
	if err != nil {
		dbc.Tx.Log("[Get] Failed to unmarshal Record: %v", err)
	}
	return err
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
	dbc.Tx.Log("[FindStartingWithAndFilterWithIndex] Pk: %v, Sk start with: %s and queryOptions %+v", pk, value, queryOptions)
	filter := queryOptions.Filter
	index := queryOptions.Index
	reverseOrder := queryOptions.ReverseOrder
	paginOptions := queryOptions.PaginOptions

	pkName := dbc.PrimaryKey
	skName := dbc.SortKey
	//If search done per index, we change pk and sk to the index ones
	if index.Name != "" {
		pkName = index.Pk
		skName = index.Sk
	}
	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String(dbc.TableName),
		KeyConditions: map[string]*dynamodb.Condition{
			pkName: {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(pk),
					},
				},
			},
			skName: {
				ComparisonOperator: aws.String("BEGINS_WITH"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(value),
					},
				},
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

	//Building Filter expression
	if filter.HasFilter() {
		expr, err := dbc.BuildQueryFilter(filter)
		if err != nil {
			dbc.Tx.Log("DB:FindStartingWithAndFilter> Error in filter expression: %+v", err)
			return err
		}
		queryInput.ExpressionAttributeNames = expr.Names()
		queryInput.ExpressionAttributeValues = expr.Values()
		queryInput.FilterExpression = expr.Filter()
	}
	//Run query with support for pagination
	allItems, err := dbc.RunQuery(queryInput, paginOptions)
	if err != nil {
		dbc.Tx.Log("[FindStartingWithAndFilterWithIndex] Run query failed: %v", err)
		return err
	}
	//Unparshal result
	err = dynamodbattribute.UnmarshalListOfMaps(allItems, data)
	if err != nil {
		dbc.Tx.Log("[FindStartingWithAndFilterWithIndex] Failed to unmarshal Record: %v", err)
	}
	return err
}

func (dbc DBConfig) BuildQueryFilter(filter DynamoFilterExpression) (expression.Expression, error) {
	builder := expression.NewBuilder()
	if filter.HasBeginsWith() {
		dbc.Tx.Log("[BuildQueryFilter] Adding Filter expression: %+v", filter.BeginsWith[0])
		condBuilder := expression.Name(filter.BeginsWith[0].Key).BeginsWith(filter.BeginsWith[0].Value)
		builder = builder.WithFilter(condBuilder)
	}
	if filter.HasContains() {
		dbc.Tx.Log("[BuildQueryFilter] Adding Filter expression: %+v", filter.Contains[0])
		condBuilder := expression.Name(filter.Contains[0].Key).Contains(filter.Contains[0].Value)
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
	dbc.Tx.Log("[FindGreaterThanAndFilterWithIndex] Pk: %v, Sk greater than: %d with QueryOptions: %+v", pk, value, queryOptions)
	filter := queryOptions.Filter
	index := queryOptions.Index
	paginOptions := queryOptions.PaginOptions
	reverseOrder := queryOptions.ReverseOrder

	pkName := dbc.PrimaryKey
	skName := dbc.SortKey
	//If search done per index, we change pk and sk to the index ones
	if index.Name != "" {
		pkName = index.Pk
		skName = index.Sk
	}
	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String(dbc.TableName),
		KeyConditions: map[string]*dynamodb.Condition{
			pkName: {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(pk),
					},
				},
			},
			skName: {
				ComparisonOperator: aws.String("GT"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						N: aws.String(strconv.FormatInt(value, 10)),
					},
				},
			},
		},
	}

	//Optioonal index to search in
	if index.Name != "" {
		queryInput.IndexName = aws.String(index.Name)
	}

	if reverseOrder {
		queryInput.ScanIndexForward = aws.Bool(false)
	}

	//Building Filter expression
	if filter.HasFilter() {
		expr, err := dbc.BuildQueryFilter(filter)
		if err != nil {
			dbc.Tx.Log("[FindGreaterThanAndFilterWithIndex] Error in filter expression: %+v", err)
			return err
		}
		queryInput.ExpressionAttributeNames = expr.Names()
		queryInput.ExpressionAttributeValues = expr.Values()
		queryInput.FilterExpression = expr.Filter()
	}

	dbc.Tx.Log("[FindGreaterThanAndFilterWithIndex] rq: %+v", queryInput)
	//run query
	allItems, err := dbc.RunQuery(queryInput, paginOptions)
	if err != nil {
		dbc.Tx.Log("[FindGreaterThanAndFilterWithIndex] Run query failed: %v", err)
		return err
	}
	//unmarshal data
	err = dynamodbattribute.UnmarshalListOfMaps(allItems, data)
	if err != nil {
		dbc.Tx.Log("[FindGreaterThanAndFilterWithIndex] Failed to unmarshal Record: %v", err)
	}
	return err
}

func (dbc DBConfig) RunQuery(queryInput *dynamodb.QueryInput, paginOptions DynamoPaginationOptions) ([]map[string]*dynamodb.AttributeValue, error) {
	dbc.Tx.Log("[RunQuery] Request: %+v", queryInput)

	//Optional pagination
	if paginOptions.PageSize > 0 {
		queryInput.Limit = aws.Int64(paginOptions.PageSize)
	}

	allItems := []map[string]*dynamodb.AttributeValue{}
	var page int64 = 0
	var result, err = dbc.DbService.QueryWithContext(dbc.LambdaContext, queryInput)
	if err != nil {
		dbc.Tx.Log("[RunQuery] error running QueryWithContext: %v", err)
		return allItems, err
	}
	for _, item := range result.Items {
		if paginOptions.PageSize == 0 || page == paginOptions.Page {
			allItems = append(allItems, item)
		}
	}
	//if paginated response
	for result.LastEvaluatedKey != nil && (paginOptions.PageSize == 0 || page < paginOptions.Page) {
		page += 1
		dbc.Tx.Log("[RunQuery] Response is paginated. LastEvaluatedKey: %v", result.LastEvaluatedKey)
		queryInput.ExclusiveStartKey = result.LastEvaluatedKey
		result, err = dbc.DbService.QueryWithContext(dbc.LambdaContext, queryInput)
		if err != nil {
			dbc.Tx.Log("[RunQuery] error running query: %v", err)
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

func (dbc DBConfig) FindByGsi(value string, indexName string, indexPk string, data interface{}) error {
	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String(dbc.TableName),
		IndexName: aws.String(indexName),
		KeyConditions: map[string]*dynamodb.Condition{
			indexPk: {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(value),
					},
				},
			},
		},
	}
	dbc.Tx.Log("DB:FindByGsi Request: %+v", queryInput)

	var result, err = dbc.DbService.QueryWithContext(dbc.LambdaContext, queryInput)
	if err != nil {
		dbc.Tx.Log("NOT FOUND")
		dbc.Tx.Log(err.Error())
		return err
	}
	dbc.Tx.Log("DB:FindByGsi Result: %+v", result)

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, data)
	if err != nil {
		dbc.Tx.Log("[FindByGsi] Failed to unmarshal Record: %v", err)
	}
	return err
}
