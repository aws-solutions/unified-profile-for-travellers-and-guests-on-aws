// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"tah/core/appregistry"
	core "tah/core/core"
	"tah/core/customerprofiles"
	"tah/core/db"
	"tah/core/glue"
	"tah/core/iam"
	model "tah/ucp/src/business-logic/model/common"
	"testing"

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

func TestGenericUseCase(t *testing.T) {
	tx := core.NewTransaction("ucp_test", "")
	var appregistryClient = appregistry.Init(REGION)
	var iamClient = iam.Init()
	var glueClient = glue.Init(REGION, "test_glue_db")
	var dbConfig = db.Init("TEST_TABLE", "TEST_PK", "TEST_SK")
	profiles := customerprofiles.InitWithDomain("test-domain", REGION)

	ucRegistry := NewRegistry(REGION, ServiceHandlers{AppRegistry: &appregistryClient, Iam: &iamClient, Glue: &glueClient, Accp: &profiles, ErrorDB: &dbConfig})
	ucRegistry.Register("POST", "test/path/{id}", &GenericUseCase{})

	ucRegistry.AddEnv("TEST_ENV_VAR", "TEST_ENV_VAR_VALUE")
	if ucRegistry.Env["TEST_ENV_VAR"] != "TEST_ENV_VAR_VALUE" {
		t.Errorf("Use case registry AddEnv failed  %+v", ucRegistry.Env)
	}

	ucRegistry.SetTx(&tx)
	for i, _ := range ucRegistry.Reg {
		if ucRegistry.Reg[i].Tx().TransactionID != tx.TransactionID {
			t.Errorf("SetTx failed: Transaction not assigned to use case %+v", i)
		}
	}

	res, err := ucRegistry.Run(createApiGtwRq("POST", "test/path/{id}"))
	if err != nil {
		t.Errorf("Use case registry Run returns error for genericUseCase: %+v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("Use case registry shoudl return a 200 HTTP Code. Response:  %+v", res)
	}
	res, err = ucRegistry.Run(createApiGtwRq("POST", "test/path/undefined/{id}"))
	if err != nil {
		t.Errorf("Use case registry Run returns error for genericUseCaseUndefined: %+v", err)
	}
	if res.StatusCode != 400 {
		t.Errorf("Use case registry Run should return a 400 for unregistered use case. Response:  %+v", res)
	}
}

func createApiGtwRq(httpMethod string, path string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Resource:   path,
		Path:       path,
		HTTPMethod: httpMethod,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		PathParameters: map[string]string{"id": "test_id"},
		Body:           "{}",
	}
}
