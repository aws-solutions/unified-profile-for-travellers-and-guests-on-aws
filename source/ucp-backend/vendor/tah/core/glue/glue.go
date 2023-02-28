package glue

import (
	"errors"
	"fmt"
	"log"
	"tah/core/core"
	"tah/core/iam"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glue"
)

var JOB_RUN_STATUS_NOT_RUNNING = "not_running"
var JOB_RUN_STATUS_UNKNOWN = "unknown"

type Config struct {
	Client *glue.Glue
	Region string
	DbName string
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

type Job struct {
	Name             string
	Description      string
	DefaultArguments map[string]string
	MaxCapacity      float64
	NumberOfWorkers  int64
	WorkerType       string
}

func Init(region, dbName string) Config {
	mySession := session.Must(session.NewSession())
	// Set max retries to 0. This disables any default retry functionality provided
	// by the AWS SDK. https://pkg.go.dev/github.com/aws/aws-sdk-go/aws/request#Retryer
	// This avoid issues with patterns like DeleteIfExists, where a resource
	// that does not exist fails, then is created, finally the failed DeleteIfExists
	// call is retried and unintentionally deletes the new resource
	cfg := aws.NewConfig().WithRegion(region).WithMaxRetries(0)
	svc := glue.New(mySession, cfg)
	return Config{
		Client: svc,
		Region: region,
		DbName: dbName,
	}
}

func (c Config) CreateGlueRole(roleName string, actions []string) (string, string, error) {
	iamClient := iam.Init()
	doc := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect:   "Allow",
				Action:   actions,
				Resource: "*",
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
		S3Target{
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
		time.Sleep(5 * time.Second)
		crawler, err = c.GetCrawler(name)
		if err != nil {
			return "", err
		}
		it += 1
		if it*5 >= timeoutSeconds {
			return "", errors.New(fmt.Sprintf("Crawler wait timed out after %v secconds", it*5))
		}
	}
	log.Printf("Crawler State: %v. Completed", crawler.State)
	return crawler.LastCrawlStatus, err
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
	log.Printf("Waiting for job run to complete")
	status, err := c.GetJobRunStatus(name)
	if err != nil {
		return JOB_RUN_STATUS_UNKNOWN, err
	}
	it := 0
	for status != glue.JobRunStateSucceeded {
		log.Printf("Job run Status: %v Waiting 5 seconds before checking again", status)
		time.Sleep(5 * time.Second)
		status, err = c.GetJobRunStatus(name)
		if err != nil {
			return JOB_RUN_STATUS_UNKNOWN, err
		}
		if status == "FAILED" {
			log.Printf("Job run Status: %v. Completed", status)
			return status, nil
		}
		it += 1
		if it*5 >= timeoutSeconds {
			return JOB_RUN_STATUS_UNKNOWN, errors.New(fmt.Sprintf("Job wait timed out after %v secconds", it*5))
		}
	}
	log.Printf("Job run Status: %v. Completed", status)
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

func (c Config) GetJobRunStatus(jobName string) (string, error) {
	input := &glue.GetJobRunsInput{
		JobName: &jobName,
	}
	output, err := c.Client.GetJobRuns(input)
	if err != nil {
		log.Printf("[GetJob] Error getting job runs: %v", err)
		return JOB_RUN_STATUS_UNKNOWN, err
	}
	lastJobRunTime := time.Time{}
	status := JOB_RUN_STATUS_NOT_RUNNING
	for _, jobRun := range output.JobRuns {
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
			return JOB_RUN_STATUS_UNKNOWN, err
		}
		for _, jobRun := range output.JobRuns {
			if jobRun.StartedOn.After(lastJobRunTime) {
				lastJobRunTime = *jobRun.StartedOn
				status = *jobRun.JobRunState
			}
		}
	}
	return status, nil
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
