module tah/upt

go 1.23.0

require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go v1.55.5
	github.com/aws/aws-sdk-go-v2 v1.32.4
	github.com/aws/aws-sdk-go-v2/config v1.27.31
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.15.0
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression v1.7.35
	github.com/aws/aws-sdk-go-v2/service/customerprofiles v1.39.4
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.34.6
	github.com/aws/aws-sdk-go-v2/service/ecs v1.45.2
	github.com/aws/aws-sdk-go-v2/service/entityresolution v1.13.1
	github.com/aws/aws-sdk-go-v2/service/ssm v1.52.6
	github.com/aws/smithy-go v1.22.0
	github.com/brianvoe/gofakeit/v6 v6.28.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.1
	github.com/mmcloughlin/geohash v0.10.0
	github.com/opensearch-project/opensearch-go/v2 v2.3.0
	github.com/stretchr/testify v1.10.0
	github.com/xitongsys/parquet-go v1.6.2
	github.com/xitongsys/parquet-go-source v0.0.0-20240122235623-d6294584ab18
	golang.org/x/sync v0.11.0
)

replace nhooyr.io/websocket v1.8.7 => github.com/coder/websocket v1.8.7

require github.com/patrickmn/go-cache v2.1.0+incompatible

require (
	github.com/apache/arrow/go/arrow v0.0.0-20200730104253-651201b0f516 // indirect
	github.com/apache/thrift v0.14.2 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.30 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.55.5
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.22.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.9.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.22.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.26.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.30.5 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
