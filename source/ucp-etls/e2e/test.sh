#!/bin/sh

env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "**********************************************"
echo "* UCP deployement for env '$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]; then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
    exit 1
fi

echo "1-removing existing executables"
rm main main.zip

echo "2-Organizing dependenies"
if [ $env == $LOCAL_ENV_NAME ]; then
    echo "2.1-Cleaning unused dependencies (local env only)"
    go mod tidy
    echo "2.2-Vendoring dependencies (local env only)"
    go mod vendor
fi

echo "3-run testing"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS=darwin
fi
export GOWORK=off # prevent issue when running on pipeline, since tah-core is not committed
export UCP_REGION=$(aws configure get region)
go test -v -failfast  -timeout 30m src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi
