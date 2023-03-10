package testutils

import (
	"os"
	"tah/core/appregistry"
	"tah/core/core"
	"tah/core/customerprofiles"
	"tah/core/glue"
	"tah/core/iam"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

/*Utility for testing purpose*/

type GenericUseCase struct {
	name string `default:"genericUseCase"`
	tx   *core.Transaction
	reg  *registry.Registry
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
	return registry.DecodeBody(u, req)
}
func (u *GenericUseCase) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
func (u *GenericUseCase) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	return model.ResponseWrapper{}, nil
}
func (u *GenericUseCase) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *GenericUseCase) Registry() *registry.Registry {
	return u.reg
}

func BuildGenericUsecase(region string) GenericUseCase {
	tx := core.NewTransaction("test_validator", "")
	reg := BuildTestRegistry(region)
	genUc := GenericUseCase{}
	genUc.SetTx(&tx)
	genUc.SetRegistry(&reg)
	return genUc
}

//returns a test registry with dummy service wrapper config
func BuildTestRegistry(region string) registry.Registry {
	var appregistryClient = appregistry.Init(region)
	var iamClient = iam.Init()
	var glueClient = glue.Init(region, "test_glue_db")
	profiles := customerprofiles.InitWithDomain("test-domain", region)
	return registry.NewRegistry(region, &appregistryClient, &iamClient, &glueClient, &profiles)
}

func GetTestRegion() string {
	//getting region for local testing
	region := os.Getenv("UCP_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}
