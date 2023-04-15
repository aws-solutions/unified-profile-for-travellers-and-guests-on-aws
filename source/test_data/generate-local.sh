envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)
sh generate.sh $envName $artifactBucket $1
