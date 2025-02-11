envName=$(jq -r .localEnvName ../../env.json)
artifactBucket=$(jq -r .artifactBucket ../../env.json)
aws s3 cp s3://$artifactBucket/config/ucp-config-$envName.json public/assets/ucp-config.json
npm run dev