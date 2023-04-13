package admin

import (
	"tah/core/core"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type StartFlow struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewStartFlow() *StartFlow {
	return &StartFlow{name: "StartFlow"}
}

func (u *StartFlow) Name() string {
	return u.name
}
func (u *StartFlow) Tx() core.Transaction {
	return *u.tx
}
func (u *StartFlow) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *StartFlow) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *StartFlow) Registry() *registry.Registry {
	return u.reg
}

func (u *StartFlow) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *StartFlow) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *StartFlow) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {

	return model.ResponseWrapper{}, nil
}

func (u *StartFlow) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
