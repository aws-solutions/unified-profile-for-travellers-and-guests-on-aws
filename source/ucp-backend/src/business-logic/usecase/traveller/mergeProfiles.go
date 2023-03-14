package traveller

import (
	"tah/core/core"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type MergeProfile struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewMergeProfile() *MergeProfile {
	return &MergeProfile{name: "MergeProfile"}
}

func (u *MergeProfile) Name() string {
	return u.name
}
func (u *MergeProfile) Tx() core.Transaction {
	return *u.tx
}
func (u *MergeProfile) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *MergeProfile) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *MergeProfile) Registry() *registry.Registry {
	return u.reg
}

func (u *MergeProfile) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return model.RequestWrapper{}, nil
}

func (u *MergeProfile) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *MergeProfile) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	return model.ResponseWrapper{}, nil
}

func (u *MergeProfile) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
