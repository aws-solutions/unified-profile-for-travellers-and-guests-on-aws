package traveller

import (
	"tah/core/core"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type DeleteProfile struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewDeleteProfile() *DeleteProfile {
	return &DeleteProfile{name: "DeleteProfile"}
}

func (u *DeleteProfile) Name() string {
	return u.name
}
func (u *DeleteProfile) Tx() core.Transaction {
	return *u.tx
}
func (u *DeleteProfile) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *DeleteProfile) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *DeleteProfile) Registry() *registry.Registry {
	return u.reg
}

func (u *DeleteProfile) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	rw := model.RequestWrapper{
		ID: req.PathParameters["id"],
	}
	return rw, nil
}

func (u *DeleteProfile) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *DeleteProfile) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	err := u.reg.Accp.DeleteProfile(req.ID)
	return model.ResponseWrapper{}, err
}

func (u *DeleteProfile) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
