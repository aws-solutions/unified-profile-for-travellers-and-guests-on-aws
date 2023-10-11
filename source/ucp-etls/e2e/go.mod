module ucp/glue-etl

go 1.18

replace tah/core v0.0.0-unpublished => ./src/tah-core

require tah/core v0.0.0-unpublished

require (
	github.com/aws/aws-sdk-go v1.44.178 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mmcloughlin/geohash v0.10.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/text v0.13.0 // indirect
)
