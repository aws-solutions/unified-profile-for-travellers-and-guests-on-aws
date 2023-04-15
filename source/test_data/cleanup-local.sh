envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)
sh cleanup.sh $envName $artifactBucket $1