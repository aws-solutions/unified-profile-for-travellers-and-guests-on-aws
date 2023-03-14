package admin

import (
	"encoding/json"
	"log"
	"tah/core/core"
	"tah/core/iam"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type LinkIndustryConnector struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewLinkIndustryConnector() *LinkIndustryConnector {
	return &LinkIndustryConnector{name: "LinkIndustryConnector"}
}

func (u *LinkIndustryConnector) Name() string {
	return u.name
}
func (u *LinkIndustryConnector) Tx() core.Transaction {
	return *u.tx
}
func (u *LinkIndustryConnector) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *LinkIndustryConnector) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *LinkIndustryConnector) Registry() *registry.Registry {
	return u.reg
}

func (u *LinkIndustryConnector) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *LinkIndustryConnector) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *LinkIndustryConnector) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	accountId := u.reg.Env["AWS_ACCOUNT_ID"]
	region := u.reg.Region
	datalakeAdminRoleArn := u.reg.Env["DATALAKE_ADMIN_ROLE_ARN"]

	glueRoleArn, _, err := GenerateGlueServiceRole(u, req.LinkIndustryConnectorRq.BucketArn, accountId, region)
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	bucketPolicy, err := GenerateConnectorBucketPolicy(req.LinkIndustryConnectorRq.BucketArn, glueRoleArn, datalakeAdminRoleArn)
	if err != nil {
		return model.ResponseWrapper{}, err
	}

	return model.ResponseWrapper{
		AwsResources: model.AwsResources{
			GlueRoleArn:              glueRoleArn,
			TahConnectorBucketPolicy: bucketPolicy,
		},
	}, err
}

func (u *LinkIndustryConnector) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

func GenerateGlueServiceRole(u *LinkIndustryConnector, bucketArn, accountId, region string) (string, string, error) {
	roleName := "CrawlConnectorServiceRole"
	err := u.reg.Iam.DeleteRoleIfExists(roleName)
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
	roleArn, policyArn, err := u.reg.Iam.CreateRoleWithPolicy(roleName, "glue.amazonaws.com", policyDoc)
	if err != nil {
		log.Printf("[GenerateGlueServiceRole] Error creating role: %v", err)
		return "", "", err
	}
	return roleArn, policyArn, nil
}

//TODO: Replace these custome definitions by tah-core struct and unit-test the policy result
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
