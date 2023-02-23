package glue

import (
	"log"
	"os"
	"tah/core/iam"
	"tah/core/s3"
	"tah/core/sqs"
	"testing"
	"time"
)

var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")
var TAH_CORE_ACCOUNT = os.Getenv("TAH_CORE_ACCOUNT")

func TestSimpleCrawler(t *testing.T) {
	dbName := "test-glue-db"
	crawlerName := "test-crawler"
	glueClient := Init(TAH_CORE_REGION, dbName)
	roleName := "simple-crawler-test-role"
	schedule := "cron(15 12 * * ? *)"
	iamClient := iam.Init()
	s3c := s3.InitRegion(TAH_CORE_REGION)
	bucketName, _ := s3c.CreateRandomBucket("tah-core-glue-test-bucket")

	err0 := glueClient.CreateDatabase(dbName)
	if err0 != nil {
		t.Errorf("[TestGlue] error creating database: %v", err0)
	}
	_, policyArn, err := glueClient.CreateGlueRole(roleName, []string{"glue:*", "sts:AssumeRole"})
	if err != nil {
		t.Errorf("[TestGlue] error deleting crawler: %v", err)
	}
	log.Printf("[TestGlue] Waiting for crawler role policy to propagate")
	time.Sleep(15 * time.Second)
	err = glueClient.CreateSimpleS3Crawler(crawlerName, roleName, schedule, "s3://"+bucketName)
	if err != nil {
		t.Errorf("[TestGlue] error creating simple crawler: %v", err)
	}
	err = glueClient.RunCrawler(crawlerName)
	if err != nil {
		t.Errorf("[TestGlue] error running crawler: %v", err)
	}
	status, err2 := glueClient.WaitForCrawlerRun(crawlerName, 300)
	if err2 != nil {
		t.Errorf("[TestGlue] error running crawler: %v", err2)
	}
	log.Printf("[TestGlue] Crawler run complete with status %s", status)

	err = glueClient.DeleteCrawlerIfExists(crawlerName)
	if err != nil {
		t.Errorf("[TestGlue] error deleting crawler: %v", err)
	}
	err = iamClient.DeleteRole(roleName)
	if err != nil {
		t.Errorf("[TestGlue] error deleting role: %v", err)
	}
	err = iamClient.DeletePolicy(policyArn)
	if err != nil {
		t.Errorf("[TestGlue] error deleting policy: %v", err)
	}
	err = glueClient.DeleteDatabase(dbName)
	if err != nil {
		t.Errorf("[TestGlue] error deleting database: %v", err)
	}
	err = s3c.DeleteBucket(bucketName)
	if err != nil {
		t.Errorf("[TestGlue] error deleting database: %v", err)
	}
}

func TestGlue(t *testing.T) {
	// Prepare required resources
	sqsClient := sqs.Init(TAH_CORE_REGION)
	queueName := "glue-sqs-test-queue"
	dlqName := "glue-sqs-test-dlq"
	queue, err1 := sqsClient.Create(queueName)
	if err1 != nil {
		t.Errorf("[TestGlue] error creating queue: %v", err1)
	}
	dlq, err2 := sqsClient.Create(dlqName)
	if err1 != nil {
		t.Errorf("[TestGlue] error creating dlq: %v", err2)
	}
	doc := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"glue:*",
					"sts:AssumeRole",
				},
				Resource: "*",
			},
		},
	}
	iamClient := iam.Init()
	roleName := "test-glue-role"
	_, policyArn, err := iamClient.CreateRoleWithPolicy(roleName, "glue.amazonaws.com", doc)
	if err != nil {
		t.Errorf("[TestGlue] error creating role: %v", err)
	}
	s3c := s3.InitRegion(TAH_CORE_REGION)
	bucketName, _ := s3c.CreateRandomBucket("tah-core-glue-test-bucket")
	scriptName := "glue.py"
	scriptPath := "test"
	s3c.Bucket = bucketName
	s3c.Path = scriptPath
	err = s3c.UploadFile(scriptName, "../../test_assets/glue.py")
	if err != nil {
		t.Errorf("[TestParseCSV] Failed to upload CSV file: %+v", err)
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
	dbName := "test-glue-db"
	crawlerName := "test-crawler"
	jobName := "test-job"
	triggerName := "test-trigger"
	scriptLocation := bucketName + "/" + scriptPath + "/" + scriptName
	glueClient := Init(TAH_CORE_REGION, dbName)
	err = glueClient.CreateDatabase(dbName)
	if err != nil {
		t.Errorf("[TestGlue] error creating database: %v", err)
	}
	log.Printf("[TestGlue] Waiting for crawler role policy to propagate")
	time.Sleep(15 * time.Second)
	err = glueClient.CreateS3Crawler(crawlerName, roleName, "cron(15 12 * * ? *)", queue, dlq, s3Targets)
	if err != nil {
		t.Errorf("[TestGlue] error creating crawler: %v", err)
	}
	crawler := Crawler{}
	crawler, err = glueClient.GetCrawler(crawlerName)
	if err != nil {
		t.Errorf("[TestGlue] error finding newly created crawler: %v", err)
	}
	if crawler.Name != crawlerName {
		t.Errorf("[TestGlue] error finding newly created crawler: %v", err)
	}
	log.Printf("Crawler %v", crawler)
	err = glueClient.CreateS3Crawler(crawlerName, roleName, "cron(15 12 * * ? *)", queue, dlq, s3Targets)
	if err != nil {
		t.Errorf("[TestGlue] error creating crawler: %v", err)
	}
	err = glueClient.CreateSparkETLJob(jobName, scriptLocation, roleName)
	if err != nil {
		t.Errorf("[TestGlue] error creating Spark ETL job: %v", err)
	}
	// Test update with nil DefaultArguments map
	err = glueClient.UpdateJobArgument(jobName, "test-key", "test-val")
	if err != nil {
		t.Errorf("[TestGlue] error updating job argument: %v", err)
	}
	job, err2 := glueClient.GetJob(jobName)
	if err2 != nil {
		t.Errorf("[TestGlue] error getting job: %v", err2)
	}
	if job.DefaultArguments["test-key"] != "test-val" {
		t.Errorf("[TestGlue] updated job argument not found")
	}
	// Test update with non-nil DefaultArguments
	err = glueClient.UpdateJobArgument(jobName, "test-key-two", "test-val-two")
	if err != nil {
		t.Errorf("[TestGlue] error updating job argument: %v", err)
	}
	job, err2 = glueClient.GetJob(jobName)
	if err2 != nil {
		t.Errorf("[TestGlue] error getting job: %v", err2)
	}
	if job.DefaultArguments["test-key-two"] != "test-val-two" {
		t.Errorf("[TestGlue] updated job argument not found")
	}
	err2 = glueClient.RunJob(jobName, map[string]string{"test-key-two": "test-val-two"})
	if err2 != nil {
		t.Errorf("[TestGlue] error running job: %v", err2)
	}
	_, err2 = glueClient.WaitForJobRun(jobName, 400)
	if err2 != nil {
		t.Errorf("[TestGlue] error waiting for job completion: %v", err2)
	}
	jobNames := []string{jobName}
	err = glueClient.CreateCrawlerSucceededTrigger(triggerName, crawlerName, jobNames)
	if err != nil {
		t.Errorf("[TestGlue] error creating trigger: %v", err)
	}
	updatedJobNames := []string{jobName}
	err = glueClient.ReplaceTriggerJobs(triggerName, updatedJobNames)
	if err != nil {
		t.Errorf("[TestGlue] error updating trigger actions")
	}
	err = glueClient.DeleteTrigger(triggerName)
	if err != nil {
		t.Errorf("[TestGlue] error deleting trigger: %v", err)
	}
	err = glueClient.DeleteJob(jobName)
	if err != nil {
		t.Errorf("[TestGlue] error deleting job: %v", err)
	}
	err = glueClient.DeleteTriggerIfExists("does-not-exist")
	if err != nil {
		t.Errorf("[TestGlue] error deleting job: %v", err)
	}
	err = glueClient.DeleteCrawlerIfExists(crawlerName)
	if err != nil {
		t.Errorf("[TestGlue] error deleting crawler: %v", err)
	}
	err = glueClient.DeleteCrawlerIfExists("does-not-exist")
	if err != nil {
		t.Errorf("[TestGlue] error deleting crawler: %v", err)
	}
	err = glueClient.DeleteDatabase(dbName)
	if err != nil {
		t.Errorf("[TestGlue] error deleting database: %v", err)
	}

	// Tear down resources
	err = iamClient.DeleteRole(roleName)
	if err != nil {
		t.Errorf("[TestGlue] error deleting role: %v", err)
	}
	err = iamClient.DeletePolicy(policyArn)
	if err != nil {
		t.Errorf("[TestGlue] error deleting policy: %v", err)
	}
	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestGlue] error deleting bucket: %v", err)
	}
	err = sqsClient.DeleteByName(queueName)
	if err != nil {
		t.Errorf("[TestGlue] error deleting queue: %v", err)
	}
	err = sqsClient.DeleteByName(dlqName)
	if err != nil {
		t.Errorf("[TestGlue] error deleting dlq: %v", err)
	}
}

func TestGlueTags(t *testing.T) {
	// Setup
	glueClient := Init(TAH_CORE_REGION, "")
	dbName := "test-tags"
	glueClient.CreateDatabase(dbName)
	dbArn := "arn:aws:glue:" + TAH_CORE_REGION + ":" + TAH_CORE_ACCOUNT + ":database/" + dbName
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
		t.Errorf("[TestGlue] error getting tags: %v", err)
	}
	log.Printf("default tags %v", defaultTags)
	if len(defaultTags) > 0 {
		t.Errorf("[TestGlue] length of original tags is not 0")
	}

	// Tag resource
	err = glueClient.TagResource(dbArn, tagsToAdd)
	if err != nil {
		t.Errorf("[TestGlue] error tagging resource: %v", err)
	}

	// Verify new tags
	var newTags map[string]string
	newTags, err = glueClient.GetTags(dbArn)
	if err != nil {
		t.Errorf("[TestGlue] error getting tags: %v", err)
	}
	log.Printf("new tags %v", newTags)

	// Untag resource, apply different tag
	err = glueClient.UntagResource(dbArn, tagsToRemove)
	if err != nil {
		t.Errorf("[TestGlue] error untagging resource: %v", err)
	}
	err = glueClient.TagResource(dbArn, tagsToReplace)
	if err != nil {
		t.Errorf("[TestGlue] error tagging resource: %v", err)
	}

	// Verify changes to tags
	newTags, err = glueClient.GetTags(dbArn)
	if err != nil {
		t.Errorf("[TestGlue] error getting tags: %v", err)
	}
	log.Printf("replacement tags %v", newTags)
	if newTags["key1"] != "" || newTags["key2"] != "val2" || len(newTags) != 1 {
		t.Errorf("[TestGlue] unexpected tags")
	}

	// Cleanup
	// tags can carry over when recreating a resource with the same name
	// so remaining tag is manually deleted
	glueClient.UntagResource(dbArn, []string{"key2"})
	glueClient.DeleteDatabase(dbName)
}
