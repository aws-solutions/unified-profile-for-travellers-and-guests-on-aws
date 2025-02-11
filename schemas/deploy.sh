#!/bin/sh
env=$1
bucket=$2
version=$3

#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev
export TAH_CORE_ENV=$env

echo "**********************************************"
echo "* AWS Travel and Hospitality Common libs Deployment  '$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
else
echo ""
echo "1-Publishing Common Libraries for Environment '$env'"
cd src
echo "1.0-Organizing dependencies"
go mod tidy
rc=$?
if [ $rc -ne 0 ]; then
    echo "Go mod tidy: Existing Build with status $rc" >&2
    exit $rc
fi
echo "1.1-running tests"
sh test.sh $env
rc=$?
if [ $rc -ne 0 ]; then
    echo "Unit Test Failed: Existing Build with status $rc" >&2
    exit $rc
fi
echo "1.2-creating archive"
cd tah-common
if [ -z "$version" ]; then
    version="latest"
fi
echo "1.3 Uploading Common Libraries version $version for Environement '$env'"
zip -r tah-common.zip .
aws s3 cp tah-common.zip s3://$bucket/$env/$version/tah-common.zip
rm tah-common.zip
echo ""
echo "2-Publishing Json schema files"
echo "2.1-generating json schemas and examples"
cd ../../generators/go-to-spec
sh generate.sh
rc=$?
if [ $rc -ne 0 ]; then
    echo "Go mod tidy: Existing Build with status $rc" >&2
    exit $rc
fi
echo "3-Copying generator executable"
aws s3 cp main s3://$bucket/$env/$version/main

echo "4-generating glue schema files"
cd ../spec-to-glue
sh generate.sh
fi
