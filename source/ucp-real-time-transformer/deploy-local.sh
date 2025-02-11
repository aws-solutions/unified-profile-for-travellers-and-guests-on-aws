#!/bin/bash

export AWS_PAGER=""
envName=$(jq -r .localEnvName ../env.json)
fnNameTransformer=ucpRealtimeTransformer
fnNameIngestor=ucpRealtimeTransformerAccp
fileTransformer=""../../deployment/regional-s3-assets/$fnNameTransformer.zip
fileIngestor=""../../deployment/regional-s3-assets/$fnNameIngestor.zip

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

aws lambda get-function --function-name $fnNameTransformer"$envName"
rc=$?
if [ $rc -ne 0 ]; then
    # assuming if one doesn't exist, neither exist
    echo "Lambda function not found. This is expected if the Lambda has not been created yet."
    echo "Skipping update-function-code"
    exit 0
fi

aws lambda update-function-code --function-name $fnNameTransformer"$envName" --zip-file fileb://$fileTransformer
aws lambda update-function-code --function-name $fnNameIngestor"$envName" --zip-file fileb://$fileIngestor
