#!/bin/bash
export UCP_REGION=$(aws configure get region)
echo "Running solution functional tests in region $UCP_REGION"
echo ""
mkdir ../z-coverage/ucp-industry-connector-transfer
go test -v -failfast -coverprofile ../z-coverage/ucp-industry-connector-transfer/cover-main.out src/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
