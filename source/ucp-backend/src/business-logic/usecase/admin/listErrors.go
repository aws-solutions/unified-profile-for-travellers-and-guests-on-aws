package admin

import (
	"tah/core/core"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type ListErrors struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewListErrors() *ListErrors {
	return &ListErrors{name: "ListErrors"}
}

func (u *ListErrors) Name() string {
	return u.name
}
func (u *ListErrors) Tx() core.Transaction {
	return *u.tx
}
func (u *ListErrors) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *ListErrors) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *ListErrors) Registry() *registry.Registry {
	return u.reg
}

func (u *ListErrors) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *ListErrors) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *ListErrors) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	errs, totalErrors, err := u.reg.Accp.GetErrors()
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	ingErrors := []model.IngestionErrors{}
	for _, ingErr := range errs {
		ingErrors = append(ingErrors, model.IngestionErrors{
			Reason:  ingErr.Reason,
			Message: ingErr.Message,
		})
	}
	return model.ResponseWrapper{IngestionErrors: ingErrors, TotalErrors: totalErrors}, nil
}

func (u *ListErrors) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
