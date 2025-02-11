#!/bin/bash

if ! docker info >/dev/null 2>&1; then
    echo "Docker is not running. Please start Docker and try again."
    exit 1
fi

export AWS_PAGER=""
envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)
path=$envName

# set -eoux pipefail

# install deps
npm ci
pushd ./custom_resource || exit
npm ci
popd || exit

sh build-local.sh "$1"
rc=$?
if [ $rc -ne 0 ]; then
echo "CDK Deploy Failed! Existing Build with status $rc" >&2
exit $rc
fi
region=$(aws configure get region)
aws s3 cp ../../deployment/regional-s3-assets/custom-resource.zip s3://"$artifactBucket"-"$region"/"$path"/custom-resource.zip

assetsFile=./cdk.out/UCPInfraStack"$envName".assets.json
dockerImages=$(jq -r '.dockerImages | keys | join(" ")' "$assetsFile")
pushd ../../deployment/cdk-solution-helper || exit
npm ci
npx ts-node ./index ../../source/ucp-infra/cdk.out ../regional-s3-assets
pushd ../../source/ucp-infra/cdk.out || exit
for pythonAsset in $(ls asset.*.py); do
    cp "$pythonAsset" ../../../deployment/regional-s3-assets/"${pythonAsset#asset.}"
done
popd || exit
for image in $dockerImages; do rm ../regional-s3-assets/"$image".zip; done
aws s3 cp --recursive ../regional-s3-assets s3://"$artifactBucket"-"$region"/"$path"
popd || exit

for image in $dockerImages; do npx cdk-assets -p "$assetsFile" publish "$image"; done

pushd ../../deployment/regional-s3-assets || exit
zip -r allArtifacts.zip *
mv allArtifacts.zip ../global-s3-assets
aws s3 cp ../global-s3-assets/allArtifacts.zip s3://"$artifactBucket"/"$path"/allArtifacts.zip
popd || exit

aws cloudformation deploy \
    --force-upload \
    --s3-prefix "$path" \
    --template-file ../../deployment/global-s3-assets/ucp.template.json \
    --stack-name UCPInfraStack"$envName" \
    --s3-bucket  "$artifactBucket" \
    --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM \
    --parameter-overrides logLevel=DEBUG

rc=$?
if [ $rc -ne 0 ]; then
    echo "Failed to deploy Stack $rc" >&2
    echo "Go to https://console.aws.amazon.com/cloudformation/home?#/stacks/events and check stack UCPInfraStack$envName?"
    exit $rc
fi
sh generate-config.sh "$envName" "$artifactBucket"
