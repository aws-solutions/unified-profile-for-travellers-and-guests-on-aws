envName=$(jq -r .localEnvName ../../../env.json)
artifactBucket=$(jq -r .artifactBucket ../../../env.json)
sh test.sh $envName $artifactBucket
