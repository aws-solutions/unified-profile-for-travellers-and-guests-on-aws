envName=$(jq -r .localEnvName ../../env.json)
artifactBucket=$(jq -r .artifactBucket ../../env.json)

# Parse command-line options
skip_update_deps=false
while [[ "$#" -gt 0 ]]; do
  case $1 in
    --skip-update-deps)
      skip_update_deps=true
      ;;
    *)
      echo "Unrecognized option: $1"
      exit 1
      ;;
  esac
  shift
done

# Update deps
if [ "$skip_update_deps" = false ]; then
    echo "Updating dependencies"
    current_directory=$(pwd)
    cd ../
    sh update-private-packages.sh
    cd "$current_directory" || exit
fi

sh test.sh "$envName" "$artifactBucket"
