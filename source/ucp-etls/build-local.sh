envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)
sh build.sh $envName $artifactBucket $1
rc=$?
if [ $rc != 0 ]; then
    echo "Changes have been detected in the transformation code"
    echo "To accept these changes, run sh update-test-data.sh"
    exit 1
fi