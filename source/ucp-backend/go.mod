module tah/ucp

go 1.18

replace tah/ucp/src/business-logic/usecase v0.0.0-unpublished => ./src/business-logic/usecase

replace tah/ucp/src/business-logic/model v0.0.0-unpublished => ./src/business-logic/model

// Note - ./src/tah-core must be copied from your local S3 bucket
// Use lbuild-local.sh to copy or as an example for commands to use
replace tah/core v0.0.0-unpublished => ./src/tah-core

require (
	github.com/aws/aws-lambda-go v1.37.0
	github.com/pkg/errors v0.9.1
	tah/core v0.0.0-unpublished
)

require (
	github.com/aws/aws-sdk-go v1.44.178 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mmcloughlin/geohash v0.10.0 // indirect
)
