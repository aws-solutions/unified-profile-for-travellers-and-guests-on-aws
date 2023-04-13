package admin

import (
	"tah/core/core"
	"tah/core/customerprofiles"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type RetreiveConfig struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewRetreiveConfig() *RetreiveConfig {
	return &RetreiveConfig{name: "RetreiveConfig"}
}

func (u *RetreiveConfig) Name() string {
	return u.name
}
func (u *RetreiveConfig) Tx() core.Transaction {
	return *u.tx
}
func (u *RetreiveConfig) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *RetreiveConfig) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *RetreiveConfig) Registry() *registry.Registry {
	return u.reg
}

func (u *RetreiveConfig) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *RetreiveConfig) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *RetreiveConfig) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	dom, err := u.reg.Accp.GetDomain()
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	domain := model.Domain{
		Name:            dom.Name,
		NObjects:        dom.NObjects,
		NProfiles:       dom.NProfiles,
		MatchingEnabled: dom.MatchingEnabled}
	mappings, err2 := u.reg.Accp.GetMappings()
	if err2 != nil {
		return model.ResponseWrapper{}, err2
	}
	domain.Mappings = parseMappings(mappings)

	integration, err3 := u.reg.Accp.GetIntegrations()
	if err3 != nil {
		return model.ResponseWrapper{}, err3
	}
	domain.Integrations = parseIntegrations(integration)

	return model.ResponseWrapper{UCPConfig: model.UCPConfig{Domains: []model.Domain{domain}}}, err
}

func (u *RetreiveConfig) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

func parseIntegrations(profileIntegrations []customerprofiles.Integration) []model.Integration {
	integrations := []model.Integration{}
	for _, pi := range profileIntegrations {
		integrations = append(integrations, model.Integration{
			Source:         pi.Source,
			Target:         pi.Target,
			Status:         pi.Status,
			StatusMessage:  pi.StatusMessage,
			LastRun:        pi.LastRun,
			LastRunStatus:  pi.LastRunStatus,
			LastRunMessage: pi.LastRunMessage,
			Trigger:        pi.Trigger,
			FlowName:       pi.Name,
		})
	}
	return integrations
}

func parseMappings(profileMappings []customerprofiles.ObjectMapping) []model.ObjectMapping {
	modelMappings := []model.ObjectMapping{}
	for _, profileMapping := range profileMappings {
		modelMappings = append(modelMappings, parseMapping(profileMapping))
	}
	return modelMappings
}
func parseMapping(profileMapping customerprofiles.ObjectMapping) model.ObjectMapping {
	return model.ObjectMapping{
		Name:   profileMapping.Name,
		Fields: parseFieldMappings(profileMapping.Fields),
	}
}
func parseFieldMappings(profileMappings []customerprofiles.FieldMapping) []model.FieldMapping {
	modelMappings := []model.FieldMapping{}
	for _, profileMapping := range profileMappings {
		modelMappings = append(modelMappings, parseFieldMapping(profileMapping))
	}
	return modelMappings
}
func parseFieldMapping(profileMappings customerprofiles.FieldMapping) model.FieldMapping {
	return model.FieldMapping{
		Type:   profileMappings.Type,
		Source: profileMappings.Source,
		Target: profileMappings.Target,
	}
}
