#!/bin/sh

env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "**********************************************"
echo "* UCP coommon code testing'$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
else
echo "1-Setting up environement"
#export GOPATH=$(pwd)
#echo "GOPATH set to $GOPATH"

echo "2-removing existing executables"
rm main main.zip
echo "3-Organizing dependenies"
if [ $env == $LOCAL_ENV_NAME ]
    then
    echo "3.1-Cleaning unused dependencies (local env only)"
    go mod tidy
    echo "3.1-Vendoring dependencies (local env only)"
    go mod vendor
    rm -rf src/tah-core
fi
echo "4-Unit testing"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS=darwin
fi
export UCP_REGION=$(aws configure get region)
mkdir test_coverage
go test  -v -failfast -coverprofile test_coverage/service-admin.out src/services/admin/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with failed status $rc" >&2
  exit $rc
fi
go test  -v -failfast -coverprofile test_coverage/utils-test.out src/utils/test/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with failed status $rc" >&2
  exit $rc
fi
fi
