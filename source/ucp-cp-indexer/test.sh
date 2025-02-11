#!/bin/bash

env=$1
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "1-run testing"
if [[ $env == $LOCAL_ENV_NAME ]]; then
  export GOOS="darwin"
fi
mkdir ../z-coverage/ucp-cp-indexer

go test -v -failfast -coverprofile ../z-coverage/ucp-cp-indexer/cover-main.out src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi