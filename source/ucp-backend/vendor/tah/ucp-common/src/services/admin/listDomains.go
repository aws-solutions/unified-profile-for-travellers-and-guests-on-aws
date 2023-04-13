package admin

import (
	"tah/core/core"
	customerprofiles "tah/core/customerprofiles"
	constant "tah/ucp-common/src/constant/admin"
	model "tah/ucp-common/src/model/admin"
)

func SearchDomain(tx core.Transaction, accp customerprofiles.CustomerProfileConfig, envName string) ([]model.Domain, error) {
	domains := []model.Domain{}
	profileDomains, err := accp.ListDomains()
	if err != nil {
		return domains, err
	}
	for _, dom := range profileDomains {
		if envName == dom.Tags[constant.DOMAIN_TAG_ENV_NAME] {
			domains = append(domains, model.Domain{
				Name:        dom.Name,
				Created:     dom.Created,
				LastUpdated: dom.LastUpdated,
			})
		}
	}
	return domains, nil
}
