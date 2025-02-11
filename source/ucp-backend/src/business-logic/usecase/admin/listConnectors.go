// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"strings"

	"tah/upt/source/tah-core/appregistry"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	adminConst "tah/upt/source/ucp-common/src/constant/admin"
	adminModel "tah/upt/source/ucp-common/src/model/admin"

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

func (u *ListConnectors) AccessPermission() adminConst.AppPermission {
	return adminConst.PublicAccessPermission
}

func (u *ListConnectors) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *ListConnectors) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	return nil
}

func (u *ListConnectors) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	domainList := adminModel.IndustryConnectorDomainList{}
	err := u.reg.ConfigDB.Get(adminConst.CONFIG_DB_LINKED_CONNECTORS_PK, adminConst.CONFIG_DB_LINKED_CONNECTORS_SK, &domainList)
	if err != nil {
		u.tx.Error("[ListConnectors] Error retrieving linked domains: %v", err)
		return model.ResponseWrapper{}, err
	}
	applications, err := u.reg.AppRegistry.ListApplications()
	if err != nil {
		u.tx.Error("[ListConnectors] Error retrieving deployed Industry Connectors: %v", err)
		return model.ResponseWrapper{}, err
	}
	connectors := u.applicationsToConnectors(applications, domainList.DomainList)
	return model.ResponseWrapper{
		Connectors: &connectors,
	}, nil
}

func (u *ListConnectors) applicationsToConnectors(applications []appregistry.ApplicationSummary, domainNames []string) []model.Connector {
	connectors := []model.Connector{}
	for _, app := range applications {
		// This assumes the naming convention for industry connector solutions will not change
		if strings.HasPrefix(app.Name, INDUSTRY_CONNECTOR_PREFIX) {
			var status string
			if u.reg.Env["IS_PLACEHOLDER_CONNECTOR_BUCKET"] == "true" {
				status = model.ConnectorStatus_DeployedWithoutBucket
			} else {
				status = model.ConnectorStatus_DeployedWithBucket
			}
			conn := model.Connector{
				Name:          app.Name,
				Status:        status,
				LinkedDomains: domainNames,
			}
			connectors = append(connectors, conn)
		}
	}
	return connectors
}

func (u *ListConnectors) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
