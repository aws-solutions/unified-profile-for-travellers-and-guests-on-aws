#!/bin/bash

envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)
path=$envName
fnName=ucpSync
file=""../../deployment/regional-s3-assets/$fnName.zip

# Parse command-line options
skip_update_deps=false
skip_tests=false
while [[ "$#" -gt 0 ]]; do
  case $1 in
    --skip-update-deps)
      skip_update_deps=true
      ;;
    --skip-tests)
      skip_tests=true
      ;;
    *)
      echo "Unrecognized option: $1"
      exit 1
      ;;
  esac
  shift
done

# Update build command
build_cmd="sh build-local.sh"
if [ "$skip_update_deps" = true ]; then
  build_cmd+=" --skip-update-deps"
fi
if [ "$skip_tests" = true ]; then
  build_cmd+=" --skip-tests"
fi

# Execute build script
$build_cmd
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi

# Deploy updated function code
export AWS_PAGER=""
region=$(aws configure get region)
aws s3 cp $file s3://"$artifactBucket"-"$region"/"$path"/$fnName.zip
rc=$?
if [ $rc -ne 0 ]; then
  echo "Error uploading $fnName.zip to S3" >&2
  exit $rc
fi

aws lambda get-function --function-name $fnName"$envName"
rc=$?
if [ $rc -ne 0 ]; then
    echo "Lambda function not found. This is expected if the Lambda has not been created yet."
    echo "Skipping update-function-code"
    exit 0
fi

aws lambda update-function-code --function-name $fnName"$envName" --s3-bucket "$artifactBucket"-"$region" --s3-key "$path"/$fnName.zip