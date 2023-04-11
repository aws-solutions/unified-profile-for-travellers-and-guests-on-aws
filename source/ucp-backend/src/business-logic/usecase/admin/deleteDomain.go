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
	env := u.reg.Env["LAMBDA_ENV"]
	err0 := u.reg.Accp.DeleteDomain()
	if err0 != nil {
		return model.ResponseWrapper{}, err0
	}

	if u.reg.Glue != nil {
		businessObjectList := []string{HOTEL_BOOKING, HOTEL_STAY_REVENUE, AIR_BOOKING, CLICKSTREAM, GUEST_PROFILE, PASSENGER_PROFILE}
		for _, bizObject := range businessObjectList {
			err := u.reg.Glue.DeleteTable("ucp_" + env + "_" + req.Domain.Name + "_" + bizObject)
			if err != nil {
				return model.ResponseWrapper{}, err
			}
		}
	} else {
		u.tx.Log("[DeleteUcpDomain][warning] No tables in domain as glueClient unavailable, no tables deleted")
	}
	return model.ResponseWrapper{}, nil
}

func (u *DeleteDomain) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
