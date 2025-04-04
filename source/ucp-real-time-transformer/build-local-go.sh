envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)

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

# Update Go deps
if [ "$skip_update_deps" = false ]; then
    echo "Updating Golang dependencies"
    current_directory=$(pwd)
    cd ../
    sh update-private-packages.sh
    cd "$current_directory"
fi

# Build
build_cmd="sh build-go.sh $envName $artifactBucket"
if [ "$skip_tests" = true ]; then
    build_cmd+=" --skip-tests"
fi

# Execute build script
$build_cmd
