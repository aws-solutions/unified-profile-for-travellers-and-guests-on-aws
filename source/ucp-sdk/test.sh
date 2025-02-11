#!/bin/sh

env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "**********************************************"
echo "* UPT SDK code testing'$env' "
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

echo "4-Unit testing"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS=darwin
fi
export GOWORK=off # prevent issue when running on pipeline, since tah-core is not committed
export UCP_REGION=$(aws configure get region)
mkdir ../z-coverage/ucp-sdk
go test  -v -failfast -coverprofile ../z-coverage/ucp-sdk/sdk.out src/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with failed status $rc" >&2
  exit $rc
fi
fi
