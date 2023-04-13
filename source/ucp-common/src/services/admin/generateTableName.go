package admin

import (
	model "tah/ucp-common/src/model/admin"
)

func BuildTableName(env string, bo model.BusinessObject, domainName string) string {
	return "ucp_" + env + "_" + bo.Name + "_" + domainName
}
