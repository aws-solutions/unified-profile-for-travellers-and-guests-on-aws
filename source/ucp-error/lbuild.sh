#!/bin/sh

env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "**********************************************"
echo "* UCP Errror aggregator deployement for env '$env' "
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
echo "4-Building Executable in Vendor Mode"
export GOOS=linux
export GOARCH=amd64
go build -mod=vendor -o main src/main/main.go
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi
echo "5-Unit testing"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS=darwin
fi
sh ./test.sh $env $bucket
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi
echo "6-Zipping executable"
zip main.zip main
echo "7-Deploying to Lambda"
sh push.sh $env $bucket
if [ $env == $LOCAL_ENV_NAME ]; then
  echo "8-e2e testing (local Only, e2e are run in a separate pipeline step for staging and prod)"
  cd e2e && sh test.sh $env $bucket && cd ..
fi
fi
