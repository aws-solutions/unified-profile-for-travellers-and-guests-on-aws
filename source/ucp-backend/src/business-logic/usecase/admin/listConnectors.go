package admin

import (
	"strings"
	"tah/core/core"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

const INDUSTRY_CONNECTOR_PREFIX = "travel-and-hospitality-connector"

type ListConnectors struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewListConnectors() *ListConnectors {
	return &ListConnectors{name: "ListConnectors"}
}

func (u *ListConnectors) Name() string {
	return u.name
}
func (u *ListConnectors) Tx() core.Transaction {
	return *u.tx
}
func (u *ListConnectors) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *ListConnectors) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *ListConnectors) Registry() *registry.Registry {
	return u.reg
}

func (u *ListConnectors) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *ListConnectors) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *ListConnectors) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	connectors := []model.Connector{}
	applications, err := u.reg.AppRegistry.ListApplications()
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	for _, app := range applications {
		// This assumes the naming convention for industry connector solutions will not change
		if strings.HasPrefix(app.Name, INDUSTRY_CONNECTOR_PREFIX) {
			connectors = append(connectors, model.Connector{Name: app.Name})
		}
	}
	return model.ResponseWrapper{
		Connectors: connectors,
	}, nil

}

func (u *ListConnectors) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
