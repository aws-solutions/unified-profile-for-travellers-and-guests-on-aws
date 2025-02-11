// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package glue

import (
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/iam"
	"tah/upt/source/tah-core/s3"
	"tah/upt/source/tah-core/sqs"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

func TestTableCrud(t *testing.T) {
	t.Parallel()

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	dbName := "tah_corer_test_db"
	tableName := "tah_core_test_table"
	bucket := "any_bucket_name"
	schemaRawData := `{
		"columns": [
			{
				"name": "model_num",
				"type": {
					"isPrimitive": true,
					"inputString": "int"
				}
			},
			{
				"name": "id_num",
				"type": {
					"isPrimitive": true,
					"inputString": "int"
				}
			},
			{
				"name": "profile_name",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			},
			{
				"name": "sample",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			}
		]
	}`
	glueClient := Init(envCfg.Region, dbName, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	err = glueClient.CreateDatabase(dbName)
	if err != nil {
		t.Fatalf("[TestGlue] error creating database: %v", err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteDatabase(dbName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting database: %v", err)
		}
	})
	schemaStruct, err := ParseSchema(schemaRawData)
	if err != nil {
		t.Errorf("[TestGlue] error parsing schema string: %v", err)
	}
	err = glueClient.CreateTable(
		tableName,
		bucket,
		[]PartitionKey{{Name: "year", Type: "int"}, {Name: "month", Type: "int"}, {Name: "day", Type: "int"}},
		schemaStruct,
	)
	if err != nil {
		t.Fatalf("[TestGlue] error creating table: %v", err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteTable(tableName)
		if err != nil {
			t.Errorf("[TestGlue] error creating table: %v", err)
		}
	})
	table, err := glueClient.GetTable(tableName)
	if err != nil {
		t.Fatalf("[TestGlue] error geting table: %v", err)
	}
	if table.Name != tableName {
		t.Errorf("[TestGlue] invalid table table. %v should be %v", table.Name, tableName)
	}
	expectedColumns := map[string]bool{"model_num": true, "id_num": true, "profile_name": true, "sample": true}
	tableColumnNames := table.ColumnList
	if len(expectedColumns) != len(tableColumnNames) {
		t.Errorf("[TestGlue] Column list has incorrect length")
	}
	for _, column := range tableColumnNames {
		val, ok := expectedColumns[column]
		if !ok || !val {
			t.Errorf("[TestGlue] Column list does not match expected column list: %v unexpected", column)
		}
		expectedColumns[column] = false
	}
	partitions := []Partition{}
	now := time.Now()
	for i := 0; i < 155; i++ {
		t := now.AddDate(0, 0, i)
		y := t.Format("2006")
		m := t.Format("01")
		d := t.Format("02")
		partitions = append(partitions, Partition{
			Values:   []string{y, m, d},
			Location: "s3://" + bucket + "/" + y + "/" + m + "/" + d,
		})
	}

	errs := glueClient.AddPartitionsToTable(tableName, partitions)
	if len(errs) > 0 {
		t.Fatalf("[TestGlue] error adding partitions to table: %v", errs)
	}
	t.Cleanup(func() {
		errs = glueClient.RemovePartitionsFromTable(tableName, partitions)
		if len(errs) > 0 {
			t.Errorf("[TestGlue] error removing partitions to table: %v", errs)
		}
	})

	parts, err := glueClient.GetPartitionsFromTable(tableName)
	if err != nil {
		t.Fatalf("[TestGlue] error getting partitions from table: %v", err)
	}
	t.Logf("[TestGlue] partitions: %v", parts)
	if len(parts) != len(partitions) {
		t.Errorf("[TestGlue] table %s should have %v partitions, not %v", tableName, len(partitions), len(parts))
	}

	newDate := now.AddDate(0, 0, 154)
	y := newDate.Format("2006")
	m := newDate.Format("01")
	d := newDate.Format("02")
	is_later_year := "(year > '" + y + "')"
	is_later_month_same_year := "(year = '" + y + "' AND month > '" + m + "')"
	is_same_or_later_day := "(year = '" + y + "' AND month = '" + m + "' AND day >= '" + d + "')"
	cpp := is_later_year + " OR " + is_later_month_same_year + " OR " + is_same_or_later_day
	t.Logf("[TestGlue]  predicate: %v", cpp)
	parts, err = glueClient.GetPartitionsFromTableWithPredicate(tableName, cpp)
	if err != nil {
		t.Fatalf("[TestGlue] error getting partitions from table with predicate: %v", err)
	}
	t.Logf("[TestGlue] partitions with predicate: %v", parts)
	if len(parts) == 0 {
		t.Logf("[TestGlue] retry predicate: %v", cpp)
		parts, err = glueClient.GetPartitionsFromTableWithPredicate(tableName, cpp)
		t.Logf("[TestGlue] partitions with predicate: %v", parts)
		if err != nil {
			t.Fatalf("[TestGlue] error getting partitions from table with predicate: %v", err)
		}
	}
	if len(parts) != 1 {
		t.Errorf("[TestGlue] table %s should have %v partitions, not %v", tableName, 2, len(parts))
	}
}

func TestParquetTableCrud(t *testing.T) {
	t.Parallel()

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	testPostFix := strings.ToLower(core.GenerateUniqueId())
	dbName := "tah_core_test_parquet_db" + testPostFix
	tableName := "tah_core_test_parquet_table" + testPostFix
	bucket := "tah_core_test_parquet_bucket_name" + testPostFix
	schemaRawData := `{
		"columns": [
			{
				"name": "model_num",
				"type": {
					"isPrimitive": true,
					"inputString": "int"
				}
			},
			{
				"name": "id_num",
				"type": {
					"isPrimitive": true,
					"inputString": "int"
				}
			},
			{
				"name": "profile_name",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			},
			{
				"name": "sample",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			}
		]
	}`

	glueClient := Init(envCfg.Region, dbName, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	err = glueClient.CreateDatabase(dbName)
	if err != nil {
		t.Fatalf("[%s] error creating database: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteDatabase(dbName)
		if err != nil {
			t.Errorf("[%s] Error deleting db: %v", t.Name(), err)
		}
	})

	schemaStruct, err := ParseSchema(schemaRawData)
	if err != nil {
		t.Fatalf("[%s] error parsing schema string: %v", t.Name(), err)
	}

	err = glueClient.CreateParquetTable(
		tableName,
		bucket,
		[]PartitionKey{
			{Name: "year", Type: "int"},
			{Name: "month", Type: "int"},
			{Name: "day", Type: "int"},
		},
		schemaStruct,
	)
	if err != nil {
		t.Fatalf("[%s] error creating table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteTable(tableName)
		if err != nil {
			t.Errorf("[%s] Error deleting table: %v", t.Name(), err)
		}
	})

	table, err := glueClient.GetTable(tableName)
	if err != nil {
		t.Fatalf("[%s] error getting table: %v", t.Name(), err)
	}
	if table.Name != tableName {
		t.Fatalf("[%s] invalid table table. %v should be %v", t.Name(), table.Name, tableName)
	}

	expectedColumns := map[string]bool{"model_num": true, "id_num": true, "profile_name": true, "sample": true}
	tableColumnNames := table.ColumnList
	if len(expectedColumns) != len(tableColumnNames) {
		t.Fatalf("[%s] Column list has incorrect length", t.Name())
	}

	for _, column := range tableColumnNames {
		val, ok := expectedColumns[column]
		if !ok || !val {
			t.Fatalf("[%s] Column list does not match expected column list: %v unexpected", t.Name(), column)
		}
		expectedColumns[column] = false
	}

	partitions := []Partition{}
	now := time.Now()
	for i := 0; i < 155; i++ {
		t := now.AddDate(0, 0, i)
		y := t.Format("2006")
		m := t.Format("01")
		d := t.Format("02")
		partitions = append(partitions, Partition{
			Values:   []string{y, m, d},
			Location: "s3://" + bucket + "/" + y + "/" + m + "/" + d,
		})
	}

	errs := glueClient.AddParquetPartitionsToTable(tableName, partitions)
	if len(errs) > 0 {
		t.Fatalf("[%s] error adding partitions to table: %v", t.Name(), errs)
	}
	t.Cleanup(func() {
		errs := glueClient.RemovePartitionsFromTable(tableName, partitions)
		if len(errs) > 0 {
			t.Errorf("[%s] Error removing partitions: %v", t.Name(), errs)
		}
	})

	parts, err := glueClient.GetPartitionsFromTable(tableName)
	if err != nil {
		t.Fatalf("[%s] error getting partitions from table: %v", t.Name(), err)
	}
	if len(parts) != len(partitions) {
		t.Fatalf("[%s] table %s should have %v partitions, not %v", t.Name(), tableName, len(partitions), len(parts))
	}

	newDate := now.AddDate(0, 0, 154)
	y := newDate.Format("2006")
	m := newDate.Format("01")
	d := newDate.Format("02")
	is_later_year := "(year > '" + y + "')"
	is_later_month_same_year := "(year = '" + y + "' AND month > '" + m + "')"
	is_same_or_later_day := "(year = '" + y + "' AND month = '" + m + "' AND day >= '" + d + "')"
	cpp := is_later_year + " OR " + is_later_month_same_year + " OR " + is_same_or_later_day
	t.Logf("[TestGlue]  predicate: %v", cpp)
	parts, err = glueClient.GetPartitionsFromTableWithPredicate(tableName, cpp)
	if err != nil {
		t.Fatalf("[TestGlue] error getting partitions from table with predicate: %v", err)
	}
	t.Logf("[TestGlue] partitions with predicate: %v", parts)
	if len(parts) == 0 {
		t.Logf("[TestGlue] retry predicate: %v", cpp)
		parts, err = glueClient.GetPartitionsFromTableWithPredicate(tableName, cpp)
		t.Logf("[TestGlue] partitions with predicate: %v", parts)
		if err != nil {
			t.Fatalf("[TestGlue] error getting partitions from table with predicate: %v", err)
		}
	}
	if len(parts) != 1 {
		t.Errorf("[TestGlue] table %s should have %v partitions, not %v", tableName, 1, len(parts))
	}
}

func TestSimpleCrawler(t *testing.T) {
	t.Parallel()

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	dbName := "test-glue-db-" + core.GenerateUniqueId()
	crawlerName := "test-crawler-" + strings.ToLower(core.GenerateUniqueId())
	glueClient := Init(envCfg.Region, dbName, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	roleName := "simple-crawler-test-role-" + strings.ToLower(core.GenerateUniqueId())
	schedule := "cron(15 12 * * ? *)"
	iamClient := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	s3c := s3.InitRegion(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	bucketName, err := s3c.CreateRandomBucket("tah-core-glue-test-bucket")
	if err != nil {
		t.Fatalf("[%s] Error creating bucket: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = s3c.DeleteBucket(bucketName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting database: %v", err)
		}
	})

	err = glueClient.CreateDatabase(dbName)
	if err != nil {
		t.Fatalf("[TestGlue] error creating database: %v", err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteDatabase(dbName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting database: %v", err)
		}
	})
	_, policyArn, err := glueClient.CreateGlueRole(roleName, []string{"glue:*", "sts:AssumeRole"})
	if err != nil {
		t.Fatalf("[TestGlue] error creating Glue role: %v", err)
	}
	t.Cleanup(func() {
		err = iamClient.DeletePolicy(policyArn)
		if err != nil {
			t.Errorf("[TestGlue] error deleting policy: %v", err)
		}
	})
	t.Cleanup(func() {
		err = iamClient.DeleteRole(roleName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting role: %v", err)
		}
	})
	t.Logf("[TestGlue] Waiting for crawler role policy to propagate")
	time.Sleep(15 * time.Second)
	err = glueClient.CreateSimpleS3Crawler(crawlerName, roleName, schedule, bucketName)
	if err != nil {
		t.Fatalf("[TestGlue] error creating simple crawler: %v", err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteCrawlerIfExists(crawlerName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting crawler: %v", err)
		}
	})
	err = glueClient.RunCrawler(crawlerName)
	if err != nil {
		t.Fatalf("[TestGlue] error running crawler: %v", err)
	}
	status, err := glueClient.WaitForCrawlerRun(crawlerName, 300)
	if err != nil {
		t.Errorf("[TestGlue] error running crawler: %v", err)
	}
	t.Logf("[TestGlue] Crawler run complete with status %s", status)
}

func TestGlue(t *testing.T) {
	t.Parallel()

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	// Prepare required resources
	sqsClient := sqs.Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	queueName := "glue-sqs-test-queue-" + strings.ToLower(core.GenerateUniqueId())
	dlqName := "glue-sqs-test-dlq-" + strings.ToLower(core.GenerateUniqueId())
	queue, err := sqsClient.Create(queueName)
	if err != nil {
		t.Fatalf("[TestGlue] error creating queue: %v", err)
	}
	t.Cleanup(func() {
		err = sqsClient.DeleteByName(queueName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting queue: %v", err)
		}
	})
	dlq, err := sqsClient.Create(dlqName)
	if err != nil {
		t.Fatalf("[TestGlue] error creating dlq: %v", err)
	}
	t.Cleanup(func() {
		err = sqsClient.DeleteByName(dlqName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting dlq: %v", err)
		}
	})
	doc := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"glue:*",
					"sts:AssumeRole",
				},
				Resource: []string{"*"},
			},
		},
	}
	iamClient := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	roleName := "test-glue-role-" + strings.ToLower(core.GenerateUniqueId())
	roleArn, policyArn, err := iamClient.CreateRoleWithPolicy(roleName, "glue.amazonaws.com", doc)
	if err != nil {
		t.Fatalf("[TestGlue] error creating role: %v", err)
	}
	t.Cleanup(func() {
		err = iamClient.DeletePolicy(policyArn)
		if err != nil {
			t.Errorf("[TestGlue] error deleting policy: %v", err)
		}
	})
	t.Cleanup(func() {
		err = iamClient.DeleteRole(roleName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting role: %v", err)
		}
	})
	t.Logf("Sleeping to allow role to propagate")
	time.Sleep(5 * time.Second)
	s3c := s3.InitRegion(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	bucketName, err := s3c.CreateRandomBucket("tah-core-glue-test-bucket")
	if err != nil {
		t.Fatalf("[%s] Error creating bucket: %v", t.Name(), err)
	}
	s3c.Bucket = bucketName
	t.Cleanup(func() {
		err = s3c.EmptyAndDelete()
		if err != nil {
			t.Errorf("[TestGlue] error deleting bucket: %v", err)
		}
	})
	t.Logf("Sleeping to allow bucket to propagate")
	time.Sleep(5 * time.Second)
	bucketArn := "arn:aws:s3:::" + bucketName
	resources := []string{bucketArn, bucketArn + "/*"}
	actions := []string{"s3:GetObject"}
	principal := map[string][]string{"AWS": {roleArn}}
	err = s3c.AddPolicy(bucketName, resources, actions, principal)
	if err != nil {
		t.Fatalf("[%s] Error adding bucket policy: %v", t.Name(), err)
	}
	scriptName := "glue.py"
	scriptPath := "test"
	s3c.Path = scriptPath
	err = s3c.UploadFile(scriptName, "../../test_data/ssi_core/glue.py")
	if err != nil {
		t.Fatalf("[TestParseCSV] Failed to upload CSV file: %+v", err)
	}
	s3Target := S3Target{
		ConnectionName: "test",
		Path:           "s3://" + bucketName,
		SampleSize:     1,
	}
	s3Targets := []S3Target{
		s3Target,
	}

	// Test Glue code
	dbName := "test-glue-db-" + core.GenerateUniqueId()
	crawlerName := "test-crawler-" + strings.ToLower(core.GenerateUniqueId())
	jobName := "test-job"
	triggerName := "test-trigger"
	scriptLocation := "s3://" + bucketName + "/" + scriptPath + "/" + scriptName
	glueClient := Init(envCfg.Region, dbName, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	err = glueClient.CreateDatabase(dbName)
	if err != nil {
		t.Fatalf("[TestGlue] error creating database: %v", err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteDatabase(dbName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting database: %v", err)
		}
	})
	t.Logf("[TestGlue] Waiting for crawler role policy to propagate")
	time.Sleep(15 * time.Second)
	err = glueClient.CreateS3Crawler(crawlerName, roleName, "cron(15 12 * * ? *)", queue, dlq, s3Targets)
	if err != nil {
		t.Fatalf("[TestGlue] error creating crawler: %v", err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteCrawlerIfExists(crawlerName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting crawler: %v", err)
		}
	})
	crawler := Crawler{}
	crawler, err = glueClient.GetCrawler(crawlerName)
	if err != nil {
		t.Fatalf("[TestGlue] error finding newly created crawler: %v", err)
	}
	if crawler.Name != crawlerName {
		t.Errorf("[TestGlue] error finding newly created crawler: %v", err)
	}
	t.Logf("Crawler %v", crawler)
	err = glueClient.CreateS3Crawler(crawlerName, roleName, "cron(15 12 * * ? *)", queue, dlq, s3Targets)
	if err != nil {
		t.Fatalf("[TestGlue] error creating crawler: %v", err)
	}
	err = glueClient.CreateSparkETLJob(jobName, scriptLocation, roleName)
	if err != nil {
		t.Fatalf("[TestGlue] error creating Spark ETL job: %v", err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteJob(jobName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting job: %v", err)
		}
	})
	// Test update with nil DefaultArguments map
	err = glueClient.UpdateJobArgument(jobName, "test-key", "test-val")
	if err != nil {
		t.Fatalf("[TestGlue] error updating job argument: %v", err)
	}
	job, err := glueClient.GetJob(jobName)
	if err != nil {
		t.Fatalf("[TestGlue] error getting job: %v", err)
	}
	if job.DefaultArguments["test-key"] != "test-val" {
		t.Errorf("[TestGlue] updated job argument not found")
	}
	// Test update with non-nil DefaultArguments
	err = glueClient.UpdateJobArgument(jobName, "test-key-two", "test-val-two")
	if err != nil {
		t.Fatalf("[TestGlue] error updating job argument: %v", err)
	}
	job, err = glueClient.GetJob(jobName)
	if err != nil {
		t.Fatalf("[TestGlue] error getting job: %v", err)
	}
	if job.DefaultArguments["test-key-two"] != "test-val-two" {
		t.Errorf("[TestGlue] updated job argument not found")
	}
	err = glueClient.RunJob(jobName, map[string]string{"test-key-two": "test-val-two"})
	if err != nil {
		t.Fatalf("[TestGlue] error running job: %v", err)
	}
	status, err := glueClient.WaitForJobRun(jobName, 400)
	if err != nil {
		t.Fatalf("[TestGlue] error waiting for job completion: %v", err)
	}
	if status != "SUCCEEDED" {
		t.Fatalf("[%s] Expected job to succeed: %s", t.Name(), status)
	}
	jobNames := []string{jobName}
	err = glueClient.CreateCrawlerSucceededTrigger(triggerName, crawlerName, jobNames)
	if err != nil {
		t.Fatalf("[TestGlue] error creating trigger: %v", err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteTrigger(triggerName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting trigger: %v", err)
		}
	})
	updatedJobNames := []string{jobName}
	err = glueClient.ReplaceTriggerJobs(triggerName, updatedJobNames)
	if err != nil {
		t.Fatalf("[TestGlue] error updating trigger actions")
	}
}

func TestGlueTags(t *testing.T) {
	t.Parallel()

	// Setup
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	glueClient := Init(envCfg.Region, "", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	dbName := "test-tags-" + core.GenerateUniqueId()
	err = glueClient.CreateDatabase(dbName)
	if err != nil {
		t.Fatalf("[%s] Error creating DB: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteDatabase(dbName)
		if err != nil {
			t.Errorf("[%s] Error deleting DB: %v", t.Name(), err)
		}
	})
	dbArn := "arn:aws:glue:" + envCfg.Region + ":" + envCfg.AccountID + ":database/" + dbName
	tagsToAdd := make(map[string]string)
	tagsToAdd["key1"] = "val1"
	tagsToRemove := []string{
		"key1",
	}
	tagsToReplace := make(map[string]string)
	tagsToReplace["key2"] = "val2"

	// Tests
	// Check baseline tags for resource
	defaultTags, err := glueClient.GetTags(dbArn)
	if err != nil {
		t.Fatalf("[TestGlue] error getting tags: %v", err)
	}
	t.Logf("default tags %v", defaultTags)
	if len(defaultTags) > 0 {
		t.Errorf("[TestGlue] length of original tags is not 0")
	}

	// Tag resource
	err = glueClient.TagResource(dbArn, tagsToAdd)
	if err != nil {
		t.Fatalf("[TestGlue] error tagging resource: %v", err)
	}
	t.Cleanup(func() {
		// tags can carry over when recreating a resource with the same name
		// so remaining tag is manually deleted
		err = glueClient.UntagResource(dbArn, []string{"key1"})
		if err != nil {
			t.Errorf("[%s] Error untagging resource: %v", t.Name(), err)
		}
	})

	// Verify new tags
	var newTags map[string]string
	newTags, err = glueClient.GetTags(dbArn)
	if err != nil {
		t.Fatalf("[TestGlue] error getting tags: %v", err)
	}
	t.Logf("new tags %v", newTags)

	// Untag resource, apply different tag
	err = glueClient.UntagResource(dbArn, tagsToRemove)
	if err != nil {
		t.Fatalf("[TestGlue] error untagging resource: %v", err)
	}
	err = glueClient.TagResource(dbArn, tagsToReplace)
	if err != nil {
		t.Fatalf("[TestGlue] error tagging resource: %v", err)
	}
	t.Cleanup(func() {
		// tags can carry over when recreating a resource with the same name
		// so remaining tag is manually deleted
		err = glueClient.UntagResource(dbArn, []string{"key2"})
		if err != nil {
			t.Errorf("[%s] Error untagging resource: %v", t.Name(), err)
		}
	})

	// Verify changes to tags
	newTags, err = glueClient.GetTags(dbArn)
	if err != nil {
		t.Fatalf("[TestGlue] error getting tags: %v", err)
	}
	t.Logf("replacement tags %v", newTags)
	if newTags["key1"] != "" || newTags["key2"] != "val2" || len(newTags) != 1 {
		t.Errorf("[TestGlue] unexpected tags")
	}
}
