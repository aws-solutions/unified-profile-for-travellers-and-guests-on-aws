#!/bin/sh

echo "**********************************************"
echo "* E2E Test: Real time flow for env: $env"
echo "**********************************************"

env=$1
bucket=$2
if [ -z "$env" ] || [ -z "$bucket" ]; then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh test.sh <env> <bucket>"
    exit 1
fi

echo "Getting infra config"
aws s3 cp s3://$bucket/config/ucp-config-$env.json infra-config.json
cat infra-config.json
echo "Setting env variables for e2e test"
export UCP_REGION=$(aws configure get region)
export API_GATEWAY_URL=$(jq -r .ucpApiUrl infra-config.json)
export USER_POOL_ID=$(jq -r .cognitoUserPoolId infra-config.json)
export CLIENT_ID=$(jq -r .cognitoClientId infra-config.json)
export TOKEN_ENDPOINT=$(jq -r .tokenEndpoint infra-config.json)
export KINESIS_STREAM=$(jq -r .kinesisStreamNameRealTime infra-config.json)
export EVENT_BUS=$(jq -r .travellerEventBus infra-config.json)
export OUTPUT_BUCKET=$(jq -r .outputBucket infra-config.json)
export BACKUP_BUCKET=$(jq -r .realTimeFeedBucket infra-config.json)
export REALTIME_BACKUP_ENABLED=$(jq -r .realTimeBackupEnabled infra-config.json)
export GLUE_DB="ucp_db_$env"
export TRAVELER_TABLE="ucp_traveler_$env"
export ENV_NAME=$env
export GOWORK=off # prevent issue when running on pipeline, since tah-core is not committed

echo "Running e2e test"
go test -v -failfast -timeout 30m
rc=$?
rm infra-config.json
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi
