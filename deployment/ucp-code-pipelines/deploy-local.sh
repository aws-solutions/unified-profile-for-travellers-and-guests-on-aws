envName=$(jq -r .localEnvName ../env.json)
email=$(jq -r .email ../env.json)
token=$(jq -r .token ../env.json)
githubUsername=$(jq -r .githubUsername ../env.json)
solutionEnvName=$(jq -r .solutionEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)
buildFromUpstream=$(jq -r .buildFromUpstream ../env.json)
branch=$(jq -r .branch ../env.json)
partitionStartDate=$(jq -r .partitionStartDate ../env.json)

echo "Getting tah dependencies version"
tahCdkCommonVersion=$(jq -r '."tah-cdk-common"' ../../tah.json)

echo "Downloading shared cdk code"
echo "Getting tah-cdk-common version $tahCdkCommonVersion"
rm -r tah-cdk-common
aws s3api get-object --bucket $artifactBucket --key $envName/$tahCdkCommonVersion/tah-cdk-common.zip tah-cdk-common.zip
rc=$?
if [ $rc -ne 0 ]; then
    echo "Could not find tah-cdk-common with version $tahCdkCommonVersion rc" >&2
    exit $rc
fi
unzip tah-cdk-common.zip -d ./tah-cdk-common
rm tah-cdk-common.zip
rm tah-cdk-common/glue.ts

echo "Deploying infrastrure"
sh deploy.sh $envName $email $token $solutionEnvName $githubUsername $buildFromUpstream $branch $partitionStartDate