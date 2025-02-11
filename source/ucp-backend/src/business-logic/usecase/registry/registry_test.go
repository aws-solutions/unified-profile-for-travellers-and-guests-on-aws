// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"testing"

	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/appregistry"
	"tah/upt/source/tah-core/awssolutions"
	"tah/upt/source/tah-core/cognito"
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/iam"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	constant "tah/upt/source/ucp-common/src/constant/admin"

	"github.com/aws/aws-lambda-go/events"
)

const REGION = "eu-central-1"

type GenericUseCase struct {
	name string `default:"genericUseCase"`
	tx   *core.Transaction
	reg  *Registry
}

func (u *GenericUseCase) Name() string {
	return u.name
}
func (u *GenericUseCase) Tx() core.Transaction {
	return *u.tx
}
func (u *GenericUseCase) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *GenericUseCase) ValidateRequest(req model.RequestWrapper) error {
	return nil
}
func (u *GenericUseCase) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return DecodeBody(u, req)
}
func (u *GenericUseCase) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return CreateResponse(u, res)
}
func (u *GenericUseCase) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	return model.ResponseWrapper{}, nil
}
func (u *GenericUseCase) SetRegistry(reg *Registry) {
	u.reg = reg
}
func (u *GenericUseCase) Registry() *Registry {
	return u.reg
}

func (u *GenericUseCase) AccessPermission() constant.AppPermission {
	return constant.PublicAccessPermission
}

func TestGenericUseCase(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	var appregistryClient = appregistry.Init(REGION, "", "")
	var iamClient = iam.Init("", "")
	var glueClient = glue.Init(REGION, "test_glue_db", "", "")
	var dbConfig = db.Init("TEST_TABLE", "TEST_PK", "TEST_SK", "", "")
	domain := "test_domain"
	groupName := "ucp-" + domain + "-admin"
	var cognitoClient = cognito.InitMock(nil, &[]cognito.Group{{Name: groupName, Description: "*/*"}})
	cognitoClient.On("ListGroupsForUser", "test_user")
	profiles := customerprofiles.InitMockV2()
	profiles.On("GetObjectLevelFields").Return(map[string][]string{
		"air_booking": {"to", "timestamp"},
		"email_history": {
			"address",
			"timestamp",
		},
	}, nil)
	solutionUtils := awssolutions.InitMock()

	ucRegistry := NewRegistry(REGION, core.LogLevelDebug, ServiceHandlers{
		AppRegistry:   &appregistryClient,
		Iam:           &iamClient,
		Glue:          &glueClient,
		Accp:          profiles,
		ErrorDB:       &dbConfig,
		Cognito:       cognitoClient,
		SolutionUtils: solutionUtils,
	})
	ucRegistry.AddEnv("ACCP_DOMAIN_NAME", domain)
	ucRegistry.Register("POST", "test/path/{id}", &GenericUseCase{})

	ucRegistry.AddEnv("TEST_ENV_VAR", "TEST_ENV_VAR_VALUE")
	if ucRegistry.Env["TEST_ENV_VAR"] != "TEST_ENV_VAR_VALUE" {
		t.Errorf("Use case registry AddEnv failed  %+v", ucRegistry.Env)
	}

	ucRegistry.SetTx(&tx)
	for i := range ucRegistry.Reg {
		if ucRegistry.Reg[i].Tx().TransactionID != tx.TransactionID {
			t.Errorf("SetTx failed: Transaction not assigned to use case %+v", i)
		}
	}

	res, err := ucRegistry.Run(createApiGtwRq("POST", "test/path/{id}", domain, groupName))
	if err != nil {
		t.Errorf("Use case registry Run returns error for genericUseCase: %+v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("Use case registry shoudl return a 200 HTTP Code. Response:  %+v", res)
	}
	if ucRegistry.DataAccessPermission != "*/*" {
		t.Errorf("Use case registry DataAccessPermission shoudl be  %v and not %v", "*/*", ucRegistry.DataAccessPermission)
	}
	res, err = ucRegistry.Run(createApiGtwRq("POST", "test/path/undefined/{id}", domain, groupName))
	if err != nil {
		t.Errorf("Use case registry Run returns error for genericUseCaseUndefined: %+v", err)
	}
	if res.StatusCode != 404 {
		t.Errorf("Use case registry Run should return a 404 for unregistered use case. Response:  %+v", res)
	}
}

func createApiGtwRq(httpMethod string, path string, domain string, groupName string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Resource:   path,
		Path:       path,
		HTTPMethod: httpMethod,
		Headers: map[string]string{
			"Content-Type":             "application/json",
			"customer-profiles-domain": domain,
		},
		PathParameters: map[string]string{"id": "test_id"},
		Body:           "{}",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"claims": map[string]interface{}{
					"cognito:groups": groupName,
					"username":       "test_user",
				},
			},
		},
	}
}

func TestCreateRequest(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	_, err := CreateRequest(&GenericUseCase{tx: &tx}, events.APIGatewayProxyRequest{})
	if err != nil {
		t.Errorf("CreateRequest failed: %+v", err)
	}
}

func TestCreateResponse(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	_, err := CreateResponse(&GenericUseCase{tx: &tx}, model.ResponseWrapper{})
	if err != nil {
		t.Errorf("CreateRequest failed: %+v", err)
	}
}
