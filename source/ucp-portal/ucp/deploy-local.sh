envName=$(jq -r .localEnvName ../../env.json)
artifactBucket=$(jq -r .artifactBucket ../../env.json)
aws s3 cp s3://$artifactBucket/config/ucp-config-$envName.json src/app/ucp-config.json
cp src/app/app.constants-local.ts src/app/app.constants.ts
ng serve
