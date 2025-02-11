envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)
path=$envName

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
if [ "$skip_tests" = true ]; then
  build_cmd+=" --skip-tests"
fi

# Execute build script
$build_cmd
rc=$?
if [ $rc != 0 ]; then
    echo "Changes have been detected in the transformation code"
    echo "To accept these changes, run sh update-test-data.sh"
    exit 1
fi
region=$(aws configure get region)
aws s3 cp ../../deployment/regional-s3-assets/etl s3://$artifactBucket-$region/$path/etl --recursive
