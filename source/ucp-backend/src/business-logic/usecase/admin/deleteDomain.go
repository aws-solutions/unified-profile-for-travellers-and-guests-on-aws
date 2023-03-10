package admin

import (
	"tah/core/core"
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
	return registry.CreateRequest(u, req)
}

func (u *DeleteDomain) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *DeleteDomain) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	err := u.reg.Accp.DeleteDomain()
	return model.ResponseWrapper{}, err
}

func (u *DeleteDomain) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
