envName=$(jq -r .localEnvName ../../env.json)
artifactBucket=$(jq -r .artifactBucket ../../env.json)
if [ "$envName" = "null" -o "$artifactBucket" = "null" ]
then
    echo "localEnvName and artifactBucket fields must be set in env.json"
    exit 1
fi
sh build.sh $envName $artifactBucket $1