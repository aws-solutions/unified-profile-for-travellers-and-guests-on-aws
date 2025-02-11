#!/bin/sh

env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev


##parse --skip-tests flag
skipTests=false
if [ "$3" = "--skip-tests" ]; then
  echo "Skipping tests"
  skipTests=true
fi
echo "**********************************************"
echo "* ucp-change-processor deployement for env '$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
else

echo "2-removing existing executables"
rm bootstrap main.zip

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
echo "5-Unit testing"
  if [ $env = $LOCAL_ENV_NAME ]; then
    export GOOS=darwin
  fi
  export GOARCH=amd64
  sh ./test.sh
  rc=$?
  if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
  fi
fi


echo "6-Zipping executable"
zip main.zip bootstrap
mkdir -p ../../deployment/regional-s3-assets
cp main.zip  ../../deployment/regional-s3-assets/ucpChangeProcessor.zip
rm main.zip bootstrap
fi
