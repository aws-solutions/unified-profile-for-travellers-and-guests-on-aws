#!/bin/bash

n_days=$1


echo "***********************************************************************************"
echo "* AWS T&H Demo platfrom Flight Generator                                       *"
echo "**************************************************************************************"
echo "*  This script populate the demo platfrom flight database the next %v days       *"
echo "**********************************************************************************"

envName=$(jq -r .localEnvName ../../env.json)
artifactBucket=$(jq -r .artifactBucket ../../env.json)
awsprofile=$(jq -r .awsProfile ../../env.json)

echo "Getting tah dependencies"
tahCoreVersion=latest

echo "Getting tah-core version $tahCoreVersion"
rm -rf src/tah-core
rm main

aws s3api get-object --bucket "$artifactBucket" --key "$envName"/$tahCoreVersion/tah-core.zip tah-core.zip
  rc=$?
  if [ $rc -ne 0 ]; then
    echo "Could not find tah-core with version $tahCoreVersion rc" >&2
    exit $rc
  fi
unzip tah-core.zip -d src/tah-core/


echo "3-Copy Source Code"
cp -r ../../src/tah-common src/.

echo "4-Organze dependencies"
go mod tidy -e

echo "5-Generate schema and Examples"
go build -o main src/main/main.go
rc=$?
if [ $rc -ne 0 ]; then
    echo "Go mod tidy: Existing Build with status $rc" >&2
    exit $rc
fi

if [ -z "$awsprofile" ] || [ "$awsprofile" = "null" ]; then
  echo "AWS Profile not provided: running default CLI profile"
else 
  export AWS_PROFILE=$awsprofile
  echo "AWS Profile provided: running profile $AWS_PROFILE"
fi
export DEMO_PLATFORM_SHOPPING_ENDPOINT=$(jq -r .shoppingApiEndpoint ../../env.json)

chmod +x main
./main "$n_days"

rm tah-core.zip



