package admin

import (
	"tah/core/core"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type DeleteError struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewDeleteError() *DeleteError {
	return &DeleteError{name: "DeleteError"}
}

func (u *DeleteError) Name() string {
	return u.name
}
func (u *DeleteError) Tx() core.Transaction {
	return *u.tx
}
func (u *DeleteError) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *DeleteError) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *DeleteError) Registry() *registry.Registry {
	return u.reg
}

func (u *DeleteError) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return model.RequestWrapper{
		UcpErrorToDelete: model.UcpIngestionErrorToDelete{
			Type: ERROR_PK,
			ID:   req.PathParameters["id"],
		},
	}, nil
}

func (u *DeleteError) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *DeleteError) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	_, err := u.reg.ErrorDB.Delete(req.UcpErrorToDelete)
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	return model.ResponseWrapper{}, nil
}

func (u *DeleteError) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
