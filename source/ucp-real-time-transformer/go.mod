module ucp/glue-real-time-transformer

go 1.18

replace tah/core v0.0.0-unpublished => ./src/tah-core

require (
	github.com/aws/aws-lambda-go v1.37.0
	tah/core v0.0.0-unpublished
)

require (
	github.com/aws/aws-sdk-go v1.44.178 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mmcloughlin/geohash v0.10.0 // indirect
)
