# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0
env=$1
bucket=$2
path=$3
errorTTL=$4
ecrRepository=$5
ecrPublishRoleArn=$6

# set -eoux pipefail

##parse --skip-tests flag
skipTests=false
if [ "$7" = "--skip-tests" -o "$8" = "--skip-tests" ]; then
    echo "Skipping tests"
    skipTests=true
fi

echo "**********************************************"
echo "* ucp-infra build '$env' and artifact bucket 's3://$bucket/$path'"
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]; then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh build.sh <env> <bucket>"
    exit 1
fi

npx prettier --check ./**/*.ts || echo Problems found
npx eslint --max-warnings=0 . || echo Problems found

# Clean out any generated files as we no longer compile ts files into js, but
# having old js files can cause an issue where it's used instead of updated ts.
npx tsc --build --clean

rc=$?
if [ $skipTests = false ]; then
    echo "2-Running Tests"
    npm test
    rc=$?
    cp -r coverage/ ../z-coverage/ucp-infra
    rm -r coverage
fi

if [ $rc -ne 0 ]; then
    echo "CDK Deploy Failed! Existing Build with status $rc" >&2
    exit $rc
fi

echo "1. Building custom resource for environment '$env' "
cd custom_resource || exit
if [ $skipTests = false ]; then
    npm run package
    rc=$?
    cp -r coverage/ ../../z-coverage/ucp-infra/custom_resource
    rm -r coverage
else
    npm run package:notest
    rc=$?
fi

if [ $rc -ne 0 ]; then
    echo "CDK Deploy Failed! Existing Build with status $rc" >&2
    exit $rc
fi

cd ..
echo "Building UPT stack for environment '$env' "
rm -rf cdk.out/*
mkdir -p ../../deployment/global-s3-assets
mkdir -p ../../deployment/regional-s3-assets
standardTemplate=cdk.out/UCPInfraStack"$env".template.json
byoVpcTemplate=cdk.out/UCPInfraStackByoVpc"$env".template.json

echo "Synthesizing CloudFormation templates for environment '$env'"
npm run cdk synth -- -c envName="$env" -c artifactBucket="$bucket" -c artifactBucketPath="$path" -c errorTTL="$errorTTL" -c ecrRepository="$ecrRepository" -c ecrPublishRoleArn="$ecrPublishRoleArn"
rc=$?
if [ $rc -ne 0 ]; then
    echo "CDK synth failed for the standard template with status: $rc" >&2
    exit $rc
fi
if [ $skipTests = false ]; then
    cfn_nag_scan --input-path "$standardTemplate"
    rc=$?
    if [ $rc -ne 0 ]; then
        echo "cfn nag failed for the standard template with status: $rc" >&2
        exit $rc
    fi
fi

if [ $skipTests = false ]; then
    cfn_nag_scan --input-path "$byoVpcTemplate"
    rc=$?
    if [ $rc -ne 0 ]; then
        echo "cfn nag failed for BYO VPC with status $rc" >&2
        exit $rc
    fi
fi

cp "$standardTemplate" ../../deployment/global-s3-assets/ucp.template.json
cp "$byoVpcTemplate" ../../deployment/global-s3-assets/ucp-byo-vpc.template.json
cp ./custom_resource/dist/package.zip ../../deployment/regional-s3-assets/custom-resource.zip
