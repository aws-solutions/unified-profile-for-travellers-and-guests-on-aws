#!/bin/sh

env=$1
bucket=$2
LOCAL_ENV_NAME=dev
PY_FUNCTION_NAME=ucpS3ExciseQueueProcessor

##parse --skip-tests flag
skipTests=false
if [ "$3" = "--skip-tests" ]; then
  echo "Skipping tests"
  skipTests=true
fi

echo "***********************************************************"
echo "* ucp-s3-excise-queue-processor deployment for env '$env' *"
echo "***********************************************************"
if [ -z "$env" ] || [ -z "$bucket" ]; then
    echo "Environment must not be empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
    exit 1
fi
echo "*************************************************"
echo "* Python S3 Excise Queue Processor Lambda Build *"
echo "*************************************************"
if [ $skipTests = false ]; then
  echo "Running tests"
  sh test.sh $envName $artifactBucket
  rc=$?
  if [ $rc -ne 0 ]; then
    echo "Tests failed"
    exit $rc
  fi
fi

zip -r mainPy.zip ucp_s3_excise_queue_processor/index.py ucp_s3_excise_queue_processor/mutex ucp_s3_excise_queue_processor/__init__.py

rc=$?
if [ $rc -ne 0 ]; then
  echo "Python Lambda failed Build with status $rc" >&2
  exit $rc
fi
mkdir -p ../../deployment/regional-s3-assets
cp mainPy.zip ../../deployment/regional-s3-assets/$PY_FUNCTION_NAME.zip
rm -r mainPy.zip