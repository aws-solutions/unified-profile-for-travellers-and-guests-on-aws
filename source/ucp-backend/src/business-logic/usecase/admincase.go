package usecase

import (
	appregistry "tah/core/appregistry"
	glue "tah/core/glue"
	iam "tah/core/iam"
	model "tah/ucp/src/business-logic/model"

	"encoding/json"
	"log"
	"strings"
)

const IndustryConnectorPrefix = "travel-and-hospitality-connector"

// https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html
type RolePrincipal struct {
	AWS string
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

func LinkIndustryConnector(iamClient iam.Config, connectorData model.LinkIndustryConnectorRq, accountId, region string) (string, string, error) {
	glueRoleArn, _, err := GenerateGlueServiceRole(iamClient, connectorData.BucketArn, accountId, region)
	if err != nil {
		log.Printf("[LinkIndustryConnector] Error generating glue service role: %v", err)
		return "", "", err
	}
	bucketPolicy, err := GenerateConnectorBucketPolicy(connectorData.BucketArn, glueRoleArn)
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

func GenerateConnectorBucketPolicy(bucketArn, glueRoleArn string) (string, error) {
	policy := Policy{
		Version: "2012-10-17",
		Statement: Statement{
			Effect: "Allow",
			Principal: RolePrincipal{
				AWS: glueRoleArn,
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

func CreateConnectorJobTrigger(glueClient glue.Config, businessObject, crawlerName, jobName string) error {
	err := glueClient.CreateCrawlerSucceededTrigger("ucp-crawl-success-trigger-"+businessObject, crawlerName, jobName)
	if err != nil {
		log.Printf("[CreateConnectorCrawler] Error creating trigger for %v: %v", businessObject, err)
	}
	return err
}
