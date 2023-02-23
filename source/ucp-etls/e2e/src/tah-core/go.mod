module tah/core

go 1.18

replace tah/core v0.0.0-unpublished => ./tah-core

require (
	github.com/aws/aws-lambda-go v1.37.0
	github.com/aws/aws-sdk-go v1.44.178
	github.com/go-sql-driver/mysql v1.7.0
	github.com/google/uuid v1.3.0
	github.com/jinzhu/gorm v1.9.16
	github.com/mmcloughlin/geohash v0.10.0
	golang.org/x/net v0.5.0
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
)
