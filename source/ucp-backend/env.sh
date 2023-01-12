#!/bin/sh

env=$1
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=local_dev_env

echo "**********************************************"
echo "* Restaurant Edge Platfrom environment setup '$env' "
echo "***********************************************"
if [ -z "$env" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env>"
    exit 1
else
    export GOPATH=$(pwd)
    if [ $env == $LOCAL_ENV_NAME ]
    then
    echo "0-Getting core package"
        rm -rf src/tah-core
        aws s3api get-object --bucket cloudrack-infra-artifacts --key $env/tah-core.zip tah-core.zip
        unzip tah-core.zip -d src/tah-core/
        rm -rf tah-core.zip
    fi
    echo "1-Getting go package dependencies"
    echo "******************************"
    echo "1.1-Getting GORM"
    go get -u github.com/jinzhu/gorm
    echo "1.2-Getting AWS SDK"
    go get -u github.com/aws/aws-sdk-go/...
    echo "1.3-Getting MySQL Drivers"
    go get -u github.com/go-sql-driver/mysql
    echo "1.4-Getting Lambda Go"
    go get -u github.com/aws/aws-lambda-go/lambda
    echo "1.5-Getting Xray Go"
    go get -u github.com/aws/aws-xray-sdk-go/...
    echo "1.6-Getting GeoHashing lib"
    go get -u github.com/mmcloughlin/geohash
      echo "10-Getting Google UUID lib"
    go get github.com/google/uuid
fi