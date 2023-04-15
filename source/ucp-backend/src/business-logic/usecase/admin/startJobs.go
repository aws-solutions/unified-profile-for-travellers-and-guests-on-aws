package admin

import (
	"tah/core/core"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type StartJobs struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewStartJobs() *StartJobs {
	return &StartJobs{name: "StartJobs"}
}

func (u *StartJobs) Name() string {
	return u.name
}
func (u *StartJobs) Tx() core.Transaction {
	return *u.tx
}
func (u *StartJobs) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *StartJobs) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *StartJobs) Registry() *registry.Registry {
	return u.reg
}

func (u *StartJobs) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *StartJobs) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *StartJobs) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	u.tx.Log("Starting glue jobs")
	_, err := u.reg.SyncLambda.InvokeAsync(events.CloudWatchEvent{})
	return model.ResponseWrapper{}, err
}

func (u *StartJobs) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
