#!/bin/sh

env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "**********************************************"
echo "* UCP common code testing '$env' "
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
mkdir ../z-coverage/ucp-common
go test -v -failfast -coverprofile ../z-coverage/ucp-common/adapter.out src/adapter/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with failed status $rc" >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-common/service-admin.out src/services/admin/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with failed status $rc" >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-common/service-traveler.out src/services/traveller/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with failed status $rc" >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-common/utils-test.out src/utils/test/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with failed status $rc" >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-common/utils-utils.out src/utils/utils/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with failed status $rc" >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-common/utils-config.out src/utils/config/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with failed status $rc" >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-common/constant-admin.out src/constant/admin/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with failed status $rc" >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-common/cover-mappings.out src/model/accp-mappings/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
fi
