module tah/ucp

go 1.16

replace tah/ucp/src/business-logic/usecase v0.0.0-unpublished => ./src/business-logic/usecase

replace tah/ucp/src/business-logic/model v0.0.0-unpublished => ./src/business-logic/model

replace tah/core v0.0.0-unpublished => ./src/tah-core

require (
	github.com/aws/aws-lambda-go v1.37.0
	github.com/pkg/errors v0.9.1
	tah/core v0.0.0-unpublished
)
