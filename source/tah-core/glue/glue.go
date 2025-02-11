// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package glue

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/iam"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glue"
)

var JOB_RUN_STATUS_NOT_RUNNING = "not_running"
var JOB_RUN_STATUS_UNKNOWN = "unknown"
var WAIT_DELAY = 5

// Glue allows 100 partitions in batch create
var GLUE_PARTITION_BATCH_SIZE = 100

// Glue allows 25 partitions in batch delete
var GLUE_PARTITION_DELETE_BATCH_SIZE = 25

// Static check for interface implementation
var _ IConfig = &Config{}

type IConfig interface {
	CreateTable(name string, bucketName string, partitionKeys []PartitionKey, schema Schema) error
	DeleteTable(name string) error
	RunJob(name string, args map[string]string) error
	AddPartitionsToTable(tableName string, partitions []Partition) []error
	AddParquetPartitionsToTable(tableName string, partitions []Partition) []error
	GetPartitionsFromTable(tableName string) ([]Partition, error)
	RemovePartitionsFromTable(tableName string, partitions []Partition) []error
}

type Config struct {
	Client          *glue.Glue
	Region          string
	DbName          string
	SolutionId      string
	SolutionVersion string
}

type S3Target struct {
	ConnectionName string
	Path           string
	SampleSize     int64
}

type Crawler struct {
	Name            string
	State           string
	LastCrawlStatus string
}

type Partition struct {
	Values   []string
	Location string
}

type Table struct {
	Name       string
	Location   string
	ColumnList []string
}

type Job struct {
	Name             string
	Description      string
	DefaultArguments map[string]string
	MaxCapacity      float64
	NumberOfWorkers  int64
	WorkerType       string
}

type PartitionKey struct {
	Name string
	Type string
}

type Schema struct {
	Columns []struct {
		Name string `json:"name"`
		Type struct {
			IsPrimitive bool   `json:"isPrimitive"`
			InputString string `json:"inputString"`
		} `json:"type"`
	} `json:"columns"`
}

func Init(region, dbName string, solutionId, solutionVersion string) Config {
	mySession := session.Must(session.NewSession())
	client := core.CreateClient(solutionId, solutionVersion)
	// Set max retries to 0. This disables any default retry functionality provided
	// by the AWS SDK. https://pkg.go.dev/github.com/aws/aws-sdk-go/aws/request#Retryer
	// This avoid issues with patterns like DeleteIfExists, where a resource
	// that does not exist fails, then is created, finally the failed DeleteIfExists
	// call is retried and unintentionally deletes the new resource
	cfg := aws.NewConfig().WithRegion(region).WithHTTPClient(client).WithMaxRetries(0)
	svc := glue.New(mySession, cfg)
	return Config{
		Client:          svc,
		Region:          region,
		DbName:          dbName,
		SolutionId:      solutionId,
		SolutionVersion: solutionVersion,
	}
}

func InitWithRetries(region, dbName string, solutionId, solutionVersion string) Config {
	mySession := session.Must(session.NewSession())
	client := core.CreateClient(solutionId, solutionVersion)
	cfg := aws.NewConfig().WithHTTPClient(client).WithRegion(region)
	svc := glue.New(mySession, cfg)
	return Config{
		Client:          svc,
		Region:          region,
		DbName:          dbName,
		SolutionId:      solutionId,
		SolutionVersion: solutionVersion,
	}
}

func (c Config) CreateGlueRole(roleName string, actions []string) (string, string, error) {
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
	_, policyArn, err := iamClient.CreateRoleWithPolicy(roleName, "glue.amazonaws.com", doc)
	return roleName, policyArn, err
}

func (c Config) CreateDatabase(name string) error {
	databaseInput := &glue.DatabaseInput{
		Name: &name,
	}
	input := &glue.CreateDatabaseInput{
		DatabaseInput: databaseInput,
	}
	_, err := c.Client.CreateDatabase(input)
	if err != nil {
		log.Printf("[CreateDatabase] Error: %v", err)
	}
	return err
}

func (c Config) DeleteDatabase(name string) error {
	input := &glue.DeleteDatabaseInput{
		Name: &name,
	}
	_, err := c.Client.DeleteDatabase(input)
	if err != nil {
		log.Printf("[DeleteDatabase] Error: %v", err)
	}
	return err
}

func (c Config) CreateSimpleS3Crawler(name, role, schedule, s3Bucket string) error {
	s3Targets := []S3Target{
		{
			Path:       "s3://" + s3Bucket,
			SampleSize: 50,
		},
	}
	return c.CreateS3Crawler(name, role, schedule, "", "", s3Targets)
}

func (c Config) CreateS3Crawler(name, role, schedule, queueArn, dlqArn string, s3Targets []S3Target) error {
	err1 := c.DeleteCrawlerIfExists(name)
	if err1 != nil {
		log.Printf("[CreateS3Crawler] Error deleting existing crawler: %v", err1)
		return err1
	}
	targets := &glue.CrawlerTargets{
		S3Targets: ConvertS3Targets(s3Targets, queueArn, dlqArn),
	}
	input := &glue.CreateCrawlerInput{
		DatabaseName: &c.DbName,
		Name:         &name,
		Role:         &role,
		Schedule:     &schedule,
		Targets:      targets,
	}
	_, err2 := c.Client.CreateCrawler(input)
	if err2 != nil {
		log.Printf("[CreateS3Crawler] Error creating crawler: %v", err2)
	}
	return err2
}

func (c Config) GetCrawler(name string) (Crawler, error) {
	output, err := c.Client.GetCrawler(&glue.GetCrawlerInput{
		Name: &name,
	})
	if err != nil {
		return Crawler{}, err
	}
	crawler := Crawler{
		Name:  *output.Crawler.Name,
		State: *output.Crawler.State,
	}
	if output.Crawler.LastCrawl != nil {
		crawler.LastCrawlStatus = *output.Crawler.LastCrawl.Status
	}
	return crawler, nil
}

func (c Config) DeleteCrawlerIfExists(name string) error {
	input := &glue.DeleteCrawlerInput{
		Name: &name,
	}
	_, err := c.Client.DeleteCrawler(input)
	if isNoSuchEntityError(err) {
		return nil
	}
	if err != nil {
		log.Printf("[DeleteCrawler] Error: %v", err)
		return err
	}
	return nil
}

func (c Config) RunCrawler(name string) error {
	input := &glue.StartCrawlerInput{
		Name: &name,
	}
	_, err := c.Client.StartCrawler(input)
	return err
}

func (c Config) WaitForCrawlerRun(name string, timeoutSeconds int) (string, error) {
	log.Printf("Waiting for crawler run to complete")
	crawler, err := c.GetCrawler(name)
	if err != nil {
		return "", err
	}
	it := 0
	for crawler.State != glue.CrawlerStateReady {
		log.Printf("Crawler State: %v Waiting 5 seconds before checking again", crawler.State)
		time.Sleep(time.Duration(WAIT_DELAY) * time.Second)
		crawler, err = c.GetCrawler(name)
		if err != nil {
			return "", err
		}
		it += 1
		if it*WAIT_DELAY >= timeoutSeconds {
			return "", fmt.Errorf("Crawler wait timed out after %v seconds", it*5)
		}
	}
	log.Printf("Crawler State: %v. Completed", crawler.State)
	return crawler.LastCrawlStatus, err
}

func (c Config) CreateTable(name string, bucketName string, partitionKeys []PartitionKey, schema Schema) error {
	pKeys := []*glue.Column{}
	for _, key := range partitionKeys {
		pKeys = append(pKeys, &glue.Column{
			Name: aws.String(key.Name),
			Type: aws.String(key.Type),
		})
	}

	glueColumnList := []*glue.Column{}

	columnList := schema.Columns
	for _, column := range columnList {
		columnName := column.Name
		columnType := column.Type.InputString
		glueColumn := glue.Column{}

		glueColumn.Name = &columnName
		glueColumn.Type = &columnType
		glueColumnList = append(glueColumnList, &glueColumn)
	}

	_, err := c.Client.CreateTable(&glue.CreateTableInput{
		DatabaseName: aws.String(c.DbName),
		TableInput: &glue.TableInput{
			Name:        aws.String(name),
			Description: aws.String("Table created by Go SDK"),
			StorageDescriptor: &glue.StorageDescriptor{
				Location:     aws.String("s3://" + bucketName),
				Columns:      glueColumnList,
				InputFormat:  aws.String("org.apache.hadoop.mapred.TextInputFormat"),
				OutputFormat: aws.String("org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat"),
				SerdeInfo: &glue.SerDeInfo{
					SerializationLibrary: aws.String("org.openx.data.jsonserde.JsonSerDe"),
				},
			},
			Parameters: map[string]*string{
				"classification": aws.String("json"),
			},
			PartitionKeys: pKeys,
			TableType:     aws.String("EXTERNAL_TABLE"),
		},
	})
	return err
}

func (c Config) CreateParquetTable(name string, bucketName string, partitionKeys []PartitionKey, schema Schema) error {
	pKeys := []*glue.Column{}
	for _, key := range partitionKeys {
		pKeys = append(pKeys, &glue.Column{
			Name: aws.String(key.Name),
			Type: aws.String(key.Type),
		})
	}

	glueColumnList := []*glue.Column{}

	columnList := schema.Columns
	for _, column := range columnList {
		columnName := column.Name
		columnType := column.Type.InputString
		glueColumn := glue.Column{}

		glueColumn.Name = &columnName
		glueColumn.Type = &columnType
		glueColumnList = append(glueColumnList, &glueColumn)
	}

	_, err := c.Client.CreateTable(&glue.CreateTableInput{
		DatabaseName: aws.String(c.DbName),
		TableInput: &glue.TableInput{
			Name:        aws.String(name),
			Description: aws.String("Table created by Go SDK"),
			StorageDescriptor: &glue.StorageDescriptor{
				Location:     aws.String("s3://" + bucketName),
				Columns:      glueColumnList,
				InputFormat:  aws.String("org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"),
				OutputFormat: aws.String("org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"),
				SerdeInfo: &glue.SerDeInfo{
					SerializationLibrary: aws.String("org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"),
				},
			},
			Parameters: map[string]*string{
				"classification": aws.String("parquet"),
			},
			PartitionKeys: pKeys,
			TableType:     aws.String("EXTERNAL_TABLE"),
		},
	})
	return err
}

func ParseSchema(schemaData string) (Schema, error) {
	var typeScriptSchema Schema

	err1 := json.Unmarshal([]byte(schemaData), &typeScriptSchema)
	if err1 != nil {
		return Schema{}, err1
	}
	return typeScriptSchema, nil
}

func (c Config) AddPartitionsToTable(tableName string, partitions []Partition) []error {
	return c.addPartitions(tableName, partitions, &glue.StorageDescriptor{
		InputFormat:  aws.String("org.apache.hadoop.mapred.TextInputFormat"),
		OutputFormat: aws.String("org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat"),
		Compressed:   aws.Bool(false),
		SerdeInfo: &glue.SerDeInfo{
			SerializationLibrary: aws.String("org.openx.data.jsonserde.JsonSerDe"),
		},
	})
}

func (c Config) AddParquetPartitionsToTable(tableName string, partitions []Partition) []error {
	return c.addPartitions(tableName, partitions, &glue.StorageDescriptor{
		InputFormat:  aws.String("org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"),
		OutputFormat: aws.String("org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"),
		Compressed:   aws.Bool(true),
		Parameters: map[string]*string{
			"parquet.compression": aws.String("GZIP"),
		},
		SerdeInfo: &glue.SerDeInfo{
			SerializationLibrary: aws.String("org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"),
		},
	})
}

func (c Config) addPartitions(tableName string, partitions []Partition, storageDescriptor *glue.StorageDescriptor) []error {
	errs := []error{}
	log.Printf("Total Number of Partitions to create: %v", len(partitions))
	batchedPartitions := core.Chunk(core.InterfaceSlice(partitions), GLUE_PARTITION_BATCH_SIZE)
	for i, batch := range batchedPartitions {
		partitionInputs := []*glue.PartitionInput{}
		for _, ifce := range batch {
			part := ifce.(Partition)
			partitionInputs = append(partitionInputs, &glue.PartitionInput{
				Values: aws.StringSlice(part.Values),
				StorageDescriptor: &glue.StorageDescriptor{
					Location:     aws.String(part.Location),
					InputFormat:  storageDescriptor.InputFormat,
					OutputFormat: storageDescriptor.OutputFormat,
					Compressed:   storageDescriptor.Compressed,
					Parameters:   storageDescriptor.Parameters,
					SerdeInfo:    storageDescriptor.SerdeInfo,
				},
			})
		}
		log.Printf("[batch-%d] Number of Partitions to create: %v", i, len(partitionInputs))
		out, err := c.Client.BatchCreatePartition(&glue.BatchCreatePartitionInput{
			DatabaseName:       aws.String(c.DbName),
			TableName:          aws.String(tableName),
			PartitionInputList: partitionInputs,
		})
		if err != nil {
			errs = append(errs, err)
		}
		for _, creationErr := range out.Errors {
			if creationErr.ErrorDetail != nil && creationErr.ErrorDetail.ErrorMessage != nil {
				part := ""
				for _, val := range creationErr.PartitionValues {
					part += *val + "/"
				}
				errs = append(errs, fmt.Errorf("[partition_create_error][%s] %v", part, *creationErr.ErrorDetail.ErrorMessage))
			} else {
				errs = append(errs, errors.New("unknown partition creation error"))
			}
		}
	}
	return errs
}

func (c Config) GetPartitionsFromTable(tableName string) ([]Partition, error) {
	return c.GetPartitionsFromTableWithPredicate(tableName, "")
}

func (c Config) GetPartitionsFromTableWithPredicate(tableName string, predicate string) ([]Partition, error) {
	input := &glue.GetPartitionsInput{
		DatabaseName: aws.String(c.DbName),
		TableName:    aws.String(tableName),
	}
	if predicate != "" {
		input.Expression = aws.String(predicate)
	}
	out, err := c.Client.GetPartitions(input)
	if err != nil {
		return []Partition{}, err
	}
	partitions := []Partition{}
	for _, partition := range out.Partitions {
		if partition != nil {
			values := []string{}
			for _, v := range (*partition).Values {
				values = append(values, *v)
			}
			location := ""
			if partition.StorageDescriptor != nil && partition.StorageDescriptor.Location != nil {
				location = *partition.StorageDescriptor.Location
			}
			partitions = append(partitions, Partition{
				Values:   values,
				Location: location,
			})
		}
	}
	return partitions, nil
}

func (c Config) RemovePartitionsFromTable(tableName string, partitions []Partition) []error {
	errs := []error{}
	log.Printf("Total Number of Partitions to delete: %v", len(partitions))
	batchedPartitions := core.Chunk(core.InterfaceSlice(partitions), GLUE_PARTITION_DELETE_BATCH_SIZE)
	for i, batch := range batchedPartitions {
		partitionInputs := []*glue.PartitionValueList{}
		for _, ifce := range batch {
			part := ifce.(Partition)
			partitionInputs = append(partitionInputs, &glue.PartitionValueList{
				Values: aws.StringSlice(part.Values),
			})
		}
		log.Printf("[batch-%d] Number of Partitions to delete: %v", i, len(partitionInputs))
		out, err := c.Client.BatchDeletePartition(&glue.BatchDeletePartitionInput{
			DatabaseName:       aws.String(c.DbName),
			TableName:          aws.String(tableName),
			PartitionsToDelete: partitionInputs,
		})
		if err != nil {
			errs = append(errs, err)
		}
		for _, creationErr := range out.Errors {
			if creationErr.ErrorDetail != nil && creationErr.ErrorDetail.ErrorMessage != nil {
				part := ""
				for _, val := range creationErr.PartitionValues {
					part += *val + "/"
				}
				errs = append(errs, fmt.Errorf("[partition_delete_error][%s] %v", part, *creationErr.ErrorDetail.ErrorMessage))
			} else {
				errs = append(errs, errors.New("unknown partition deletion error"))
			}
		}
	}
	return errs
}

func (c Config) GetTable(name string) (Table, error) {
	output, err := c.Client.GetTable(&glue.GetTableInput{
		DatabaseName: aws.String(c.DbName),
		Name:         &name,
	})
	columns := output.Table.StorageDescriptor.Columns
	columnStringList := make([]string, 0)
	for _, column := range columns {
		columnStringList = append(columnStringList, *column.Name)
	}

	if err != nil {
		return Table{}, err
	}
	table := Table{
		Name:       *output.Table.Name,
		ColumnList: columnStringList,
	}
	return table, nil
}

func (c Config) ListTables() ([]Table, error) {
	output, err := c.Client.GetTables(&glue.GetTablesInput{
		DatabaseName: aws.String(c.DbName),
	})
	if err != nil {
		return []Table{}, err
	}

	tableList := make([]Table, 0)
	for _, tableItem := range output.TableList {
		tableName := tableItem.Name
		location := tableItem.StorageDescriptor.Location
		table := Table{
			Name:     *tableName,
			Location: aws.StringValue(location),
		}
		tableList = append(tableList, table)
	}

	return tableList, nil
}

func (c Config) DeleteTable(name string) error {
	_, err := c.Client.DeleteTable(&glue.DeleteTableInput{
		DatabaseName: aws.String(c.DbName),
		Name:         &name,
	})
	return err
}

func ConvertS3Targets(targets []S3Target, queueArn, dlqArn string) []*glue.S3Target {
	glueTargets := []*glue.S3Target{}
	for _, target := range targets {
		glueTarget := glue.S3Target{
			ConnectionName: &target.ConnectionName,
			Path:           &target.Path,
			SampleSize:     &target.SampleSize,
		}
		if dlqArn != "" {
			glueTarget.DlqEventQueueArn = &dlqArn
		}
		if queueArn != "" {
			glueTarget.EventQueueArn = &queueArn
		}
		glueTargets = append(glueTargets, &glueTarget)
	}
	return glueTargets
}

func (c Config) CreateSparkETLJob(jobName, scriptLocation, role string) error {
	commandName := "glueetl"
	command := glue.JobCommand{
		Name:           &commandName,
		ScriptLocation: &scriptLocation,
		PythonVersion:  aws.String("3"),
	}
	input := glue.CreateJobInput{
		Command: &command,
		Name:    &jobName,
		Role:    &role,
	}

	_, err := c.Client.CreateJob(&input)
	if err != nil {
		log.Printf("[CreateSparkETLJob] Error: %v", err)
	}
	return err
}

func (c Config) RunJob(name string, args map[string]string) error {
	input := &glue.StartJobRunInput{
		JobName:   &name,
		Arguments: map[string]*string{},
	}
	for key, value := range args {
		input.Arguments[key] = aws.String(value)
	}
	_, err := c.Client.StartJobRun(input)
	return err
}

func (c Config) WaitForJobRun(name string, timeoutSeconds int) (string, error) {
	log.Printf("Waiting for job %s run to complete", name)
	_, status, err := c.GetJobRunStatus(name)
	if err != nil {
		return JOB_RUN_STATUS_UNKNOWN, err
	}
	it := 0
	for status != glue.JobRunStateSucceeded {
		log.Printf("Job %s run Status: %v Waiting 5 seconds before checking again", name, status)
		time.Sleep(time.Duration(WAIT_DELAY) * time.Second)
		_, status, err = c.GetJobRunStatus(name)
		if err != nil {
			return JOB_RUN_STATUS_UNKNOWN, err
		}
		if status == "FAILED" {
			log.Printf("Job %s run Status: %v. Completed", name, status)
			return status, nil
		}
		it += 1
		if it*WAIT_DELAY >= timeoutSeconds {
			return JOB_RUN_STATUS_UNKNOWN, fmt.Errorf("Job %s wait timed out after %v seconds", name, it*5)
		}
	}
	log.Printf("Job %s run Status: %v. Completed", name, status)
	return status, err
}

func (c Config) GetJob(jobName string) (Job, error) {
	output, err := c.Client.GetJob(&glue.GetJobInput{
		JobName: &jobName,
	})
	if err != nil {
		log.Printf("[GetJob] Error getting job: %v", err)
		return Job{}, err
	}
	job := Job{
		Name:             jobName,
		Description:      core.PtToString(output.Job.Description),
		MaxCapacity:      core.PtToFloat64(output.Job.MaxCapacity),
		NumberOfWorkers:  core.PtToInt64(output.Job.NumberOfWorkers),
		WorkerType:       core.PtToString(output.Job.WorkerType),
		DefaultArguments: core.ToMapString(output.Job.DefaultArguments),
	}
	return job, nil
}

func (c Config) GetJobRunStatus(jobName string) (time.Time, string, error) {
	return c.GetJobRunStatusWithFilter(jobName, map[string]string{})
}

func (c Config) GetJobRunStatusWithFilter(jobName string, parametters map[string]string) (time.Time, string, error) {
	input := &glue.GetJobRunsInput{
		JobName: &jobName,
	}
	output, err := c.Client.GetJobRuns(input)
	if err != nil {
		log.Printf("[GetJob] Error getting job runs: %v", err)
		return time.Time{}, JOB_RUN_STATUS_UNKNOWN, err
	}
	lastJobRunTime := time.Time{}
	status := JOB_RUN_STATUS_NOT_RUNNING
	jobRuns := filterByParametters(output.JobRuns, parametters)
	for _, jobRun := range jobRuns {
		if jobRun.StartedOn.After(lastJobRunTime) {
			lastJobRunTime = *jobRun.StartedOn
			status = *jobRun.JobRunState
		}
	}
	if output.NextToken != nil {
		input.NextToken = output.NextToken
		output, err = c.Client.GetJobRuns(input)
		if err != nil {
			log.Printf("[GetJob] Error getting job runs: %v", err)
			return lastJobRunTime, JOB_RUN_STATUS_UNKNOWN, err
		}
		jobRuns = filterByParametters(output.JobRuns, parametters)
		for _, jobRun := range jobRuns {
			if jobRun.StartedOn.After(lastJobRunTime) {
				lastJobRunTime = *jobRun.StartedOn
				status = *jobRun.JobRunState
			}
		}
	}
	return lastJobRunTime, status, nil
}

func filterByParametters(jobRuns []*glue.JobRun, parameters map[string]string) []*glue.JobRun {
	if len(parameters) == 0 {
		return jobRuns
	}
	filtered := []*glue.JobRun{}
	for _, jobRun := range jobRuns {
		for key, value := range parameters {
			if jobRun.Arguments[key] != nil && *jobRun.Arguments[key] == value {
				filtered = append(filtered, jobRun)
			}
		}
	}
	return filtered

}

func (c Config) DeleteJob(jobName string) error {
	input := glue.DeleteJobInput{
		JobName: &jobName,
	}
	_, err := c.Client.DeleteJob(&input)
	if err != nil {
		log.Printf("[DeleteJob] Error: %v", err)
	}
	return err
}

func (c Config) CreateCrawlerSucceededTrigger(triggerName, crawlerName string, jobNames []string) error {
	err := c.DeleteCrawlerIfExists(crawlerName)
	if err != nil {
		log.Printf("[CreateTrigger] Error deleting existing trigger: %v", err)
	}
	actions := []*glue.Action{}
	for _, jobName := range jobNames {
		actions = append(actions, &glue.Action{
			JobName: &jobName,
		})
	}
	crawlState := glue.CrawlStateSucceeded
	logicalOperator := glue.LogicalOperatorEquals
	predicate := glue.Predicate{
		Conditions: []*glue.Condition{
			{
				CrawlState:      &crawlState,
				CrawlerName:     &crawlerName,
				LogicalOperator: &logicalOperator,
			},
		},
	}
	triggerType := glue.TriggerTypeConditional
	startOnCreation := true
	description := fmt.Sprintf("Fires after %v successfully runs", crawlerName)
	input := glue.CreateTriggerInput{
		Actions:         actions,
		Description:     &description,
		Name:            &triggerName,
		Predicate:       &predicate,
		StartOnCreation: &startOnCreation,
		Type:            &triggerType,
	}

	_, err = c.Client.CreateTrigger(&input)
	if err != nil {
		log.Printf("[CreateTrigger] Error: %v", err)
	}
	return err
}

// Replace the Trigger's actions with the provided Glue jobs
func (c Config) ReplaceTriggerJobs(triggerName string, jobNames []string) error {
	// Get existing trigger data
	getTriggerInput := glue.GetTriggerInput{Name: &triggerName}
	trigger, err := c.Client.GetTrigger(&getTriggerInput)
	if err != nil {
		log.Printf("[ReplaceTriggerJobs] Error: %v", err)
		return err
	}

	// Create new list of actions
	var actionInput []*glue.Action
	for _, v := range jobNames {
		action := glue.Action{
			JobName: &v,
		}
		actionInput = append(actionInput, &action)
	}

	// Update trigger with new actions
	// This updates the previous trigger definition by overwriting it completely
	// https://docs.aws.amazon.com/sdk-for-go/api/service/glue/#TriggerUpdate
	triggerUpdate := glue.TriggerUpdate{
		Actions:                actionInput,
		Description:            trigger.Trigger.Description,
		EventBatchingCondition: trigger.Trigger.EventBatchingCondition,
		Predicate:              trigger.Trigger.Predicate,
		Schedule:               trigger.Trigger.Schedule,
	}
	updateTriggerInput := glue.UpdateTriggerInput{
		Name:          &triggerName,
		TriggerUpdate: &triggerUpdate,
	}
	_, err = c.Client.UpdateTrigger(&updateTriggerInput)
	if err != nil {
		log.Printf("[UpdateTriggerActions] Error: %v", err)
	}
	return err
}

func (c Config) DeleteTrigger(triggerName string) error {
	input := glue.DeleteTriggerInput{
		Name: &triggerName,
	}
	_, err := c.Client.DeleteTrigger(&input)
	if err != nil {
		log.Printf("[DeleteTrigger] Error: %v", err)
	}
	return err
}

func (c Config) DeleteTriggerIfExists(triggerName string) error {
	err := c.DeleteTrigger(triggerName)
	if err != nil && !isNoSuchEntityError(err) {
		return err
	}
	return nil
}

func (c Config) UpdateJobArgument(jobName, updateKey, updateVal string) error {
	job, err1 := c.Client.GetJob(&glue.GetJobInput{
		JobName: &jobName,
	})
	if err1 != nil {
		log.Printf("[UpdateJobArgument] Error getting existing job: %v", err1)
		return err1
	}
	// Create new map to update arguments, check for existing args to carry over
	updatedArgs := make(map[string]*string)
	if job.Job.DefaultArguments != nil {
		updatedArgs = job.Job.DefaultArguments
	}
	updatedArgs[updateKey] = &updateVal
	// Create updated input based on original values, adding updated arguments
	updateInput := glue.UpdateJobInput{
		JobName: &jobName,
		JobUpdate: &glue.JobUpdate{
			CodeGenConfigurationNodes: job.Job.CodeGenConfigurationNodes,
			Command:                   job.Job.Command,
			Connections:               job.Job.Connections,
			DefaultArguments:          updatedArgs,
			Description:               job.Job.Description,
			ExecutionClass:            job.Job.ExecutionClass,
			ExecutionProperty:         job.Job.ExecutionProperty,
			GlueVersion:               job.Job.GlueVersion,
			LogUri:                    job.Job.LogUri,
			MaxRetries:                job.Job.MaxRetries,
			NonOverridableArguments:   job.Job.NonOverridableArguments,
			NotificationProperty:      job.Job.NotificationProperty,
			Role:                      job.Job.Role,
			SecurityConfiguration:     job.Job.SecurityConfiguration,
			SourceControlDetails:      job.Job.SourceControlDetails,
			Timeout:                   job.Job.Timeout,
		},
	}
	// Input requires either MaxCapacity OR NumberOfWorkers and WorkerType
	// Determine which is set for the job, then add to updateInput accordingly
	if core.PtToFloat64(job.Job.MaxCapacity) > 0 {
		updateInput.JobUpdate.MaxCapacity = job.Job.MaxCapacity
	} else {
		updateInput.JobUpdate.NumberOfWorkers = job.Job.NumberOfWorkers
		updateInput.JobUpdate.WorkerType = job.Job.WorkerType
	}
	_, err2 := c.Client.UpdateJob(&updateInput)
	if err2 != nil {
		log.Printf("[UpdateJobArguments] Error updating job: %v", err2)
		return err2
	}
	return nil
}

func (c Config) GetTags(arn string) (map[string]string, error) {
	tags := make(map[string]string)
	input := glue.GetTagsInput{
		ResourceArn: &arn,
	}
	output, err := c.Client.GetTags(&input)
	if err != nil {
		log.Printf("[GetTags] Error: %v", err)
		return tags, err
	}

	for k, v := range output.Tags {
		tags[k] = *v
	}
	return tags, nil
}

func (c Config) TagResource(arn string, tags map[string]string) error {
	input := glue.TagResourceInput{
		ResourceArn: &arn,
		TagsToAdd:   core.ToMapPtString(tags),
	}
	_, err := c.Client.TagResource(&input)
	if err != nil {
		log.Printf("[TagResource] Error: %v", err)
	}
	return err
}

func (c Config) UntagResource(arn string, tags []string) error {
	var tagPointers []*string
	for _, v := range tags {
		tagPointers = append(tagPointers, &v)
	}
	input := glue.UntagResourceInput{
		ResourceArn:  &arn,
		TagsToRemove: tagPointers,
	}
	_, err := c.Client.UntagResource(&input)
	if err != nil {
		log.Printf("[UntagResource] Error: %v ", err)
	}
	return err
}

func CreateSchemaFromStringArray(stringArray []string) Schema {
	var columns []struct {
		Name string `json:"name"`
		Type struct {
			IsPrimitive bool   `json:"isPrimitive"`
			InputString string `json:"inputString"`
		} `json:"type"`
	}

	for _, columnName := range stringArray {
		column := struct {
			Name string `json:"name"`
			Type struct {
				IsPrimitive bool   `json:"isPrimitive"`
				InputString string `json:"inputString"`
			} `json:"type"`
		}{
			Name: columnName,
			Type: struct {
				IsPrimitive bool   `json:"isPrimitive"`
				InputString string `json:"inputString"`
			}{
				IsPrimitive: true,
				InputString: "string",
			},
		}
		columns = append(columns, column)
	}

	return Schema{Columns: columns}
}

// More information on handling AWS errors: https://pkg.go.dev/github.com/aws/aws-sdk-go/aws/awserr
func isNoSuchEntityError(err error) bool {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case glue.ErrCodeEntityNotFoundException:
			return true
		default:
			return false
		}
	}
	return false
}
