envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)

echo "Downloading shared cdk code"
rm -r tah-cdk-common
aws s3api get-object --bucket $artifactBucket --key $envName/tah-cdk-common.zip tah-cdk-common.zip
unzip tah-cdk-common.zip -d ./tah-cdk-common
rm tah-cdk-common.zip

sh deploy.sh $envName $artifactBucket