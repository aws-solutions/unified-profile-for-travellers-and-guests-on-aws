#!/bin/bash

envName=$(jq -r .localEnvName env.json)
artifactBucket=$(jq -r .artifactBucket env.json)

echo "1-Identifing version"
branch=$(git rev-parse --abbrev-ref HEAD)
regexp="release\/(.*)"
echo "current Git branch: $branch"
[[ $branch =~ $regexp ]]
current_version=$(echo "${BASH_REMATCH[1]}")
echo "current version: $current_version"

sh ./deploy.sh "$envName" "$artifactBucket"