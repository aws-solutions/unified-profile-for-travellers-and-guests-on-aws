#!/bin/bash

n_profiles=$1
businesObject=$2
traffic=$3
domain=$4

echo "***************************************************"
echo "*              UPT Traffic Generator              *"
echo "***************************************************"

echo "Validating input parameters"
if [ -z "$n_profiles" ]; then
  echo "Number of profiles not provided."
  exit 1
fi
if [ -z "$businesObject" ]; then
  echo "Business object not provided."
  exit 1
fi
if [ -z "$traffic" ]; then
  echo "Traffic type not provided."
  exit 1
fi
if [ -z "$domain" ]; then
  echo "Domain not provided."
  exit 1
fi

echo "Configuring environment"
envJson="./env.json"
awsprofile=$(jq -r .awsProfile $envJson)
echo "AWS Profile: $awsprofile"
if [ -z "$awsprofile" ] || [ "$awsprofile" == "null" ]; then
  echo "AWS Profile not provided: running default CLI profile"
else 
  export AWS_PROFILE=$awsprofile
  echo "AWS Profile provided: running profile $AWS_PROFILE"
fi
TAH_COMMON_STREAM_NAME=$(jq -r .stream $envJson)
if [ -z "$TAH_COMMON_STREAM_NAME" ] || [ "$TAH_COMMON_STREAM_NAME" == "null" ]; then
  echo "Stream name not provided."
  exit 1
fi
export TAH_COMMON_STREAM_NAME=$TAH_COMMON_STREAM_NAME
TAH_COMMON_REGION=$(jq -r .region $envJson)
if [ -z "$TAH_COMMON_REGION" ] || [ "$TAH_COMMON_REGION" == "null" ]; then
  echo "Region not provided."
  exit 1
fi
export TAH_COMMON_REGION=$TAH_COMMON_REGION

echo "Building generator executable"
go build -o main src/main/main.go
rc=$?
if [ "$rc" -ne 0 ]; then
    echo "Go mod tidy: Existing Build with status $rc" >&2
    exit "$rc"
fi

echo "Generating live traffic to stream $TAH_COMMON_STREAM_NAME in region $TAH_COMMON_REGION"
./main "$n_profiles" "$businesObject" "$traffic" "$domain"
