package usecase

import (
	appregistry "tah/core/appregistry"
	glue "tah/core/glue"
	iam "tah/core/iam"
	model "tah/ucp/src/business-logic/model"

	"encoding/json"
	"log"
	"strings"

	"golang.org/x/exp/maps"
)

const IndustryConnectorPrefix = "travel-and-hospitality-connector"

// https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html
type RolePrincipal struct {
	AWS []string
}

type Statement struct {
	Effect    string
	Principal RolePrincipal
	Action    []string
	Resource  []string
}

type Policy struct {
	Version   string
	Statement Statement
}

func ListIndustryConnectors(cfg appregistry.Config) ([]appregistry.ApplicationSummary, error) {
	connectors := []appregistry.ApplicationSummary{}
	applications, err := cfg.ListApplications()
	if err != nil {
		return connectors, err
	}
	for _, app := range applications {
		// This assumes the naming convention for industry connector solutions will not change
		if strings.HasPrefix(app.Name, IndustryConnectorPrefix) {
			connectors = append(connectors, app)
		}
	}
	return connectors, nil
}

func LinkIndustryConnector(iamClient iam.Config, connectorData model.LinkIndustryConnectorRq, accountId, region, datalakeAdminRoleArn string) (string, string, error) {
	glueRoleArn, _, err := GenerateGlueServiceRole(iamClient, connectorData.BucketArn, accountId, region)
	if err != nil {
		log.Printf("[LinkIndustryConnector] Error generating glue service role: %v", err)
		return "", "", err
	}
	bucketPolicy, err := GenerateConnectorBucketPolicy(connectorData.BucketArn, glueRoleArn, datalakeAdminRoleArn)
	if err != nil {
		log.Printf("[LinkIndustryConnector] Error generating bucket policy for connector: %v", err)
		return "", "", err
	}
	return glueRoleArn, bucketPolicy, nil
}

func GenerateGlueServiceRole(cfg iam.Config, bucketArn, accountId, region string) (string, string, error) {
	roleName := "CrawlConnectorServiceRole"
	err := cfg.DeleteRoleIfExists(roleName)
	if err != nil {
		log.Printf("[GenerateGlueServiceRole] Error deleting role: %v", err)
		return "", "", err
	}
	statement1 := iam.StatementEntry{
		Effect: "Allow",
		Action: []string{
			"s3:PutObject",
			"s3:GetObject",
			"s3:DeleteObject",
		},
		Resource: bucketArn + "/*",
	}
	statement2 := iam.StatementEntry{
		Effect: "Allow",
		Action: []string{
			"s3:ListBucket",
		},
		Resource: "*",
	}
	statement3 := iam.StatementEntry{
		Effect: "Allow",
		Action: []string{
			"glue:GetConnection",
			"glue:GetDatabase",
			"glue:GetTable",
			"glue:CreateTable",
			"glue:BatchGetPartition",
			"glue:BatchCreatePartition",
		},
		Resource: "arn:aws:glue:" + region + ":" + accountId + ":*",
	}
	statement4 := iam.StatementEntry{
		Effect: "Allow",
		Action: []string{
			"logs:PutLogEvents",
		},
		Resource: "arn:aws:logs:" + region + ":" + accountId + ":log-group:/aws-glue/*",
	}
	policyDoc := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			statement1,
			statement2,
			statement3,
			statement4,
		},
	}
	roleArn, policyArn, err := cfg.CreateRoleWithPolicy(roleName, "glue.amazonaws.com", policyDoc)
	if err != nil {
		log.Printf("[GenerateGlueServiceRole] Error creating role: %v", err)
		return "", "", err
	}
	return roleArn, policyArn, nil
}

func GenerateConnectorBucketPolicy(bucketArn, glueRoleArn, datalakeAdminRoleArn string) (string, error) {
	policy := Policy{
		Version: "2012-10-17",
		Statement: Statement{
			Effect: "Allow",
			Principal: RolePrincipal{
				AWS: []string{
					glueRoleArn,
					datalakeAdminRoleArn,
				},
			},
			Action: []string{
				"s3:*",
			},
			Resource: []string{
				bucketArn,
				bucketArn + "/*",
			},
		},
	}
	out, err := json.Marshal(policy)
	if err != nil {
		log.Printf("[GenerateConnectorBucketPolicy] Error generating policy: %v", err)
		return "", err
	}
	return string(out), nil
}

func CreateConnectorCrawler(glueClient glue.Config, glueRoleArn, bucketName, env, queueArn, dlqArn string) error {
	cronSchedule := "cron(0 * * * ? *)"
	s3Targets := []glue.S3Target{
		{
			Path:       "s3://" + bucketName,
			SampleSize: 10,
		},
	}
	crawlerName := "ucp-connector-crawler-" + env
	err := glueClient.CreateS3Crawler(crawlerName, glueRoleArn, cronSchedule, queueArn, dlqArn, s3Targets)
	if err != nil {
		log.Printf("[CreateConnectorCrawler] Error creating crawler: %v", err)
	}
	return err
}

func CreateConnectorJobTrigger(glueClient glue.Config, accountId, region, connectorId, crawlerName string) ([]string, error) {
	triggerName := "ucp-connector-crawl-success-trigger"
	triggerArn := "arn:aws:glue:" + region + ":" + accountId + ":trigger/" + triggerName
	tagKey := "connectors"
	entityNotFoundPrefix := "EntityNotFoundException"

	// Get current tags
	currentTags, err := glueClient.GetTags(triggerArn)
	taggedConnectors := currentTags[tagKey]
	triggerExists := true
	if err != nil {
		if strings.HasPrefix(err.Error(), entityNotFoundPrefix) {
			triggerExists = false
		} else {
			return []string{}, err
		}
	}

	// Create updated tags with new connector's ID
	updatedTag, updateRequired := updateTaggedConnectors(taggedConnectors, connectorId)
	updatedTags := map[string]string{
		tagKey: updatedTag,
	}

	// Create updated list of jobs to trigger
	supportedBusinessObjects := tagToObjects(updatedTag)
	jobNames := GetJobNamesForObjects(supportedBusinessObjects)

	if !updateRequired {
		return jobNames, nil
	}

	if triggerExists {
		// Update jobs based on tagged connectors
		err := glueClient.ReplaceTriggerJobs(triggerName, jobNames)
		if err != nil {
			log.Printf("[CreateConnectorCrawler] Error updating trigger: %v", err)
			return []string{}, err
		}
	} else {
		// Create trigger
		err := glueClient.CreateCrawlerSucceededTrigger(triggerName, crawlerName, jobNames)
		if err != nil {
			log.Printf("[CreateConnectorCrawler] Error creating trigger: %v", err)
			return []string{}, err
		}
	}

	// Set/update trigger tags
	err = glueClient.TagResource(triggerArn, updatedTags)
	return jobNames, err
}

func updateTaggedConnectors(taggedConnectorString, newConnector string) (string, bool) {
	if len(taggedConnectorString) == 0 {
		return newConnector, true
	}
	existingConnectors := strings.Split(taggedConnectorString, ":")
	for _, connector := range existingConnectors {
		if connector == newConnector {
			return taggedConnectorString, false
		}
	}
	return taggedConnectorString + ":" + newConnector, true
}

// Convert colon separated list of connectors to an array of business objects
func tagToObjects(tag string) []string {
	taggedConnectors := strings.Split(tag, ":")
	objects := make(map[string]string)
	for _, v := range taggedConnectors {
		connectorObjects := connectorMap[v]
		for _, v := range connectorObjects {
			objects[v] = ""
		}
	}

	return maps.Keys(objects)
}

func AddConnectorBucketToJobs(glueClient glue.Config, bucketName string, jobNames []string) error {
	key := "--SOURCE_TABLE"
	val := strings.Replace(bucketName, "-", "_", -1) // translate bucket name to Glue table name
	for _, job := range jobNames {
		err := glueClient.UpdateJobArgument(job, key, val)
		if err != nil {
			log.Printf("[AddConnectorBucketToJobs] Error setting Connector table name for Glue job %v", err)
			return err
		}
	}
	return nil
}
