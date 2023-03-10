package admin

import (
	"strings"
	"tah/core/core"
	"tah/core/glue"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type CreateConnectorCrawler struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewCreateConnectorCrawler() *CreateConnectorCrawler {
	return &CreateConnectorCrawler{name: "CreateConnectorCrawler"}
}

func (u *CreateConnectorCrawler) Name() string {
	return u.name
}
func (u *CreateConnectorCrawler) Tx() core.Transaction {
	return *u.tx
}
func (u *CreateConnectorCrawler) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *CreateConnectorCrawler) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *CreateConnectorCrawler) Registry() *registry.Registry {
	return u.reg
}

func (u *CreateConnectorCrawler) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *CreateConnectorCrawler) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *CreateConnectorCrawler) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {

	err := CreateCrawler(u, req)
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	var jobNames = []string{}
	jobNames, err = CreateConnectorJobTrigger(u, req)
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	err = AddConnectorBucketToJobs(u, req, jobNames)
	return model.ResponseWrapper{}, err
}

func (u *CreateConnectorCrawler) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

func CreateCrawler(u *CreateConnectorCrawler, req model.RequestWrapper) error {
	env := u.reg.Env["LAMBDA_ENV"]
	queueArn := u.reg.Env["CONNECTOR_CRAWLER_QUEUE"]
	dlqArn := u.reg.Env["CONNECTOR_CRAWLER_DLQ"]

	glueRoleArn := req.CreateConnectorCrawlerRq.GlueRoleArn
	split := strings.Split(req.CreateConnectorCrawlerRq.BucketPath, ":")
	bucketName := split[len(split)-1]
	cronSchedule := "cron(0 * * * ? *)"

	s3Targets := []glue.S3Target{
		{
			Path:       "s3://" + bucketName,
			SampleSize: 10,
		},
	}
	crawlerName := "ucp-connector-crawler-" + env
	err := u.reg.Glue.CreateS3Crawler(crawlerName, glueRoleArn, cronSchedule, queueArn, dlqArn, s3Targets)
	return err
}

func AddConnectorBucketToJobs(u *CreateConnectorCrawler, req model.RequestWrapper, jobNames []string) error {
	key := "--SOURCE_TABLE"
	split := strings.Split(req.CreateConnectorCrawlerRq.BucketPath, ":")
	bucketName := split[len(split)-1]
	val := strings.Replace(bucketName, "-", "_", -1) // translate bucket name to Glue table name
	for _, job := range jobNames {
		err := u.reg.Glue.UpdateJobArgument(job, key, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func getKeys(m map[string]struct{}) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func CreateConnectorJobTrigger(u *CreateConnectorCrawler, req model.RequestWrapper) ([]string, error) {
	env := u.reg.Env["LAMBDA_ENV"]
	region := u.reg.Region
	accountId := u.reg.Env["AWS_ACCOUNT_ID"]
	crawlerName := "ucp-connector-crawler-" + env
	connectorId := req.CreateConnectorCrawlerRq.ConnectorId

	triggerName := "ucp-connector-crawl-success-trigger"
	triggerArn := "arn:aws:glue:" + region + ":" + accountId + ":trigger/" + triggerName
	tagKey := "connectors"
	entityNotFoundPrefix := "EntityNotFoundException"

	// Get current tags
	currentTags, err := u.reg.Glue.GetTags(triggerArn)
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
	//supportedBusinessObjects := tagToObjects(updatedTag)
	jobNames := []string{
		u.reg.Env["HOTEL_BOOKING_JOB_NAME"],
		u.reg.Env["AIR_BOOKING_JOB_NAME"],
		u.reg.Env["GUEST_PROFILE_JOB_NAME"],
		u.reg.Env["PAX_PROFILE_JOB_NAME"],
		u.reg.Env["CLICKSTREAM_JOB_NAME"],
		u.reg.Env["HOTEL_STAY_JOB_NAME"],
	}

	if !updateRequired {
		return jobNames, nil
	}

	if triggerExists {
		// Update jobs based on tagged connectors
		err := u.reg.Glue.ReplaceTriggerJobs(triggerName, jobNames)
		if err != nil {
			u.tx.Log("[CreateConnectorCrawler] Error updating trigger: %v", err)
			return []string{}, err
		}
	} else {
		// Create trigger
		err := u.reg.Glue.CreateCrawlerSucceededTrigger(triggerName, crawlerName, jobNames)
		if err != nil {
			u.tx.Log("[CreateConnectorCrawler] Error creating trigger: %v", err)
			return []string{}, err
		}
	}

	// Set/update trigger tags
	err = u.reg.Glue.TagResource(triggerArn, updatedTags)
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
	objects := make(map[string]struct{})
	for _, v := range taggedConnectors {
		connectorObjects := CONNECTORS_MAP[v]
		for _, v := range connectorObjects {
			objects[v] = struct{}{}
		}
	}
	return getKeys(objects)
}
