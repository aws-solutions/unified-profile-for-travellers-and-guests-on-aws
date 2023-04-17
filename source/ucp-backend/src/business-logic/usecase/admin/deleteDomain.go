package admin

import (
	"errors"
	"tah/core/core"
	common "tah/ucp-common/src/constant/admin"
	services "tah/ucp-common/src/services/admin"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type DeleteDomain struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewDeleteDomain() *DeleteDomain {
	return &DeleteDomain{name: "DeleteDomain"}
}

func (u *DeleteDomain) Name() string {
	return u.name
}
func (u *DeleteDomain) Tx() core.Transaction {
	return *u.tx
}
func (u *DeleteDomain) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *DeleteDomain) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *DeleteDomain) Registry() *registry.Registry {
	return u.reg
}

func (u *DeleteDomain) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	rq := model.RequestWrapper{
		Domain: model.Domain{Name: req.PathParameters["id"]},
	}
	return rq, nil
}

func (u *DeleteDomain) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	if rq.Domain.Name == "" {
		return errors.New("Domain name cannot be empty")
	}
	return nil
}

func (u *DeleteDomain) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	env := u.reg.Env["LAMBDA_ENV"]
	u.tx.Log("Deleting tables for domain %s", req.Domain.Name)
	for _, bizObject := range common.BUSINESS_OBJECTS {
		tableName := services.BuildTableName(env, bizObject, req.Domain.Name)
		err := u.reg.Glue.DeleteTable(tableName)
		if err != nil {
			return model.ResponseWrapper{}, err
		}
	}
	u.tx.Log("Deleting domain %s", req.Domain.Name)
	err0 := u.reg.Accp.DeleteDomain()
	if err0 != nil {
		return model.ResponseWrapper{}, err0
	}

	return model.ResponseWrapper{}, nil
}

func (u *DeleteDomain) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
