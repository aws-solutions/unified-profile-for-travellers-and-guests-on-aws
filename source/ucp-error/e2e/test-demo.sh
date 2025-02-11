export AWS_PROFILE=ucp-demo
envName=$(jq -r .localEnvName ../../env-demo.json)
artifactBucket=$(jq -r .artifactBucket ../../env-demo.json)
sh test.sh $envName $artifactBucket