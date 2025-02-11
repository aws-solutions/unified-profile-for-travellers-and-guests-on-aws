
# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#  SPDX-License-Identifier: MIT-0

env=$1
bucket=$2


##parse --skip-tests flag
skipTests=false
if [ "$3" = "--skip-tests" ]; then
  echo "Skipping tests"
  skipTests=true
fi
echo "**********************************************"
echo "*  ucp-etls '$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]; then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
    exit 1
fi

if [ $skipTests = false ]; then
    echo "Running unit tests"
    mkdir ../z-coverage/tests
    mkdir ../z-coverage/tests/coverage-reports
    coverage_report_path=../z-coverage/tests/coverage-reports/etls-coverage.coverage.xml
    source_dir="$(cd $PWD/..; pwd -P)"
    python3 -m pip install -e ../tah_lib --break-system-packages
    rc=$?
    if [ $rc != 0 ]; then
        echo "tah_lib install failed"
        exit 1
    fi
    python3 -m coverage run -m unittest discover
    rc=$?
    python3 -m coverage xml
    cp coverage.xml $coverage_report_path
    sed -i -e "s,<source>$source_dir,<source>source,g" ../z-coverage/tests/coverage-reports/etls-coverage.coverage.xml
    sed -i -e "s,filename=\"$source_dir,filename=\"source,g" ../z-coverage/tests/coverage-reports/etls-coverage.coverage.xml
    rm coverage.xml
    if [ $rc != 0 ]; then
        echo "Changes have been detected in the transformation code"
        echo "To accept these changes, run sh update-test-data.sh"
        exit 1
    fi
fi

echo "Creating directory is exist"
mkdir -p ../../deployment/regional-s3-assets/etl
rm  ../../deployment/regional-s3-assets/etl/*
cp etls/* ../../deployment/regional-s3-assets/etl/
