envName=$(jq -r .localEnvName ../../env.json)
artifactBucket=$(jq -r .artifactBucket ../../env.json)
sh e2e.sh $envName $artifactBucket