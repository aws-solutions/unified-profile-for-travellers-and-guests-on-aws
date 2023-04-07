module tah/ucp-common

go 1.18

replace tah/ucp-common/src/ v0.0.0-unpublished => ./src/

// Note - ./src/tah-core must be copied from your local S3 bucket
// Use lbuild-local.sh to copy or as an example for commands to use
replace tah/core v0.0.0-unpublished => ./src/tah-core

require tah/core v0.0.0-unpublished

require (
	github.com/aws/aws-sdk-go v1.44.178 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mmcloughlin/geohash v0.10.0 // indirect
)
