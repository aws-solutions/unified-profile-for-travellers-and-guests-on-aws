#!/bin/sh

env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "**********************************************"
echo "* UCP deployement for env '$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
else
echo "1-Getting Stack information"
aws s3 cp s3://$bucket/config/ucp-config-$env.json ./ucp-config.json
cat ./ucp-config.json

export GLUE_DB_NAME=$(jq -r .glueDBname ./ucp-config.json)
export GLUE_ROLE_NAME=$(jq -r .ucpDataAdminRoleName ./ucp-config.json)
export GLUE_JOB_NAME_AIR_BOOKING=$(jq -r .customerJobNameairbooking ./ucp-config.json)
export TEST_BUCKET_AIR_BOOKING=$(jq -r .customerTestBucketairbooking ./ucp-config.json)
export TEST_BUCKET_ACCP_IMPORT=$(jq -r .connectProfileImportBucketTestOut ./ucp-config.json)
echo "GLUE_ROLE_NAME: $GLUE_ROLE_NAME"
echo "GLUE_DB_NAME: $GLUE_DB_NAME"
echo "GLUE_JOB_NAME_AIR_BOOKING: $GLUE_JOB_NAME_AIR_BOOKING"
echo "TEST_BUCKET_AIR_BOOKING: $TEST_BUCKET_AIR_BOOKING"
echo "TEST_BUCKET_ACCP_IMPORT: $TEST_BUCKET_ACCP_IMPORT"


echo "2-removing existing executables"
rm main main.zip
echo "3-Organizing dependenies"
if [ $env == $LOCAL_ENV_NAME ]
    then
    echo "3.1-Cleaning unused dependencies (local env only)"
    go mod tidy
    echo "3.1-Vendoring dependencies (local env only)"
    go mod vendor
fi
echo "4-run testing"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS=darwin
fi
export UCP_REGION=$(aws configure get region)
go test -v -failfast src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi
fi
