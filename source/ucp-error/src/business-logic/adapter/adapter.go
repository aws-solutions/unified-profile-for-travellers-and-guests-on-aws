package adapter

import (
	"tah/ucp-sync/src/business-logic/model"
)

func AccpErrorToUcpError(accp model.ACCPIngestionError) model.UcpIngestionError {
	ucp := model.UcpIngestionError{}

	return ucp
}
