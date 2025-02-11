#!/bin/bash

env=$1
bucket=$2
LOCAL_ENV_NAME=dev
FUNCTION_NAME=ucpCpWriter

skipTests=false
if [ "$3" = "--skip-tests" ]; then
  echo "Skipping tests"
  skipTests=true
fi

echo "**********************************************"
echo "* ucp-cp-writer deployment for env '$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]; then
    echo "Environment must not be empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
    exit 1
fi

echo "1-removing existing executables"
rm bootstrap main.zip
echo "2-Organizing dependenies"

echo "3-Building Executable in Vendor Mode"
export GOOS=linux
export GOARCH=arm64
export GOWORK=off # temporarily turn of since Workspace mode does not support vendoring
go build -mod=vendor -o bootstrap src/main/main.go
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi

if [ $skipTests = false ]; then
  echo "4-Unit testing"
  if [ "$env" == $LOCAL_ENV_NAME ]; then
    export GOOS=darwin
  fi
  export GOARCH=amd64
  sh ./test.sh "$env"
  rc=$?
  if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
  fi
fi

echo "5-Zipping executable"
zip main.zip bootstrap
mkdir -p ../../deployment/regional-s3-assets
cp main.zip  ../../deployment/regional-s3-assets/$FUNCTION_NAME.zip
rm -r main.zip bootstrap