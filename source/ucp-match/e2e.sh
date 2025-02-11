env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "1-Getting Stack information"
aws s3 cp s3://$bucket/config/ucp-config-$env.json ./ucp-config.json
cat ./ucp-config.json

export DYNAMO_TABLE=$(jq -r .dynamoTable ./ucp-config.json)
export DYNAMO_PK=$(jq -r .dynamoPk ./ucp-config.json)
export DYNAMO_SK=$(jq -r .dynamoSk ./ucp-config.json)
export MATCH_BUCKET=$(jq -r .matchBucket ./ucp-config.json)

echo "4-run testing"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS="darwin"
fi
export GOWORK=off # prevent issue when running on pipeline, since tah-core is not committed
export UCP_REGION=$(aws configure get region)
go test -timeout 30m -v -failfast e2e/e2e_test.go
rc=$?

rm -rf ./ucp-config.json
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi