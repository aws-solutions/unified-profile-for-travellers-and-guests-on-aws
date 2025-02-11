envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)
errorTTL=$(jq -r .errorTTL ../env.json)
ecrRepository=$(jq -r .ecrRepository ../env.json)
ecrPublishRoleArn=$(jq -r .ecrPublishRoleArn ../env.json)

if [ "$errorTTL" == "null" ]; then
    errorTTL=7
fi
sh build.sh $envName $artifactBucket $envName $errorTTL $ecrRepository $ecrPublishRoleArn $1
