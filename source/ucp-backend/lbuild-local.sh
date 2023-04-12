envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)

echo "Getting tah dependencies"
tahCoreVersion=$(jq -r '."tah-core"' ../../tah.json)

echo "Getting tah-core version $tahCoreVersion"
rm -rf src/tah-core
aws s3api get-object --bucket $artifactBucket --key $envName/$tahCoreVersion/tah-core.zip tah-core.zip
    rc=$?
    if [ $rc -ne 0 ]; then
      echo "Could not find tah-core with version $tahCoreVersion rc" >&2
      exit $rc
    fi

unzip tah-core.zip -d src/tah-core/
rm -rf tah-core.zip

echo "Getting tah dependencies version"
tahCdkCommonVersion=$(jq -r '."tah-cdk-common"' ../../tah.json)
tahCommonVersion=$(jq -r '."tah-common"' ../../tah.json)

rm -rf ./tah-common-glue-schemas/*
aws s3api get-object --bucket $artifactBucket --key $envName/$tahCommonVersion/tah-common-glue-schemas.zip tah-common-glue-schemas.zip
rc=$?
if [ $rc -ne 0 ]; then
    echo "Could not find tah-common with version $tahCommonVersion $rc" >&2
    exit $rc
fi
unzip tah-common-glue-schemas.zip -d ./tah-common-glue-schemas
rm tah-common-glue-schemas.zip

echo "Copying ucp-common"
cp -r ../ucp-common src/.

sh lbuild.sh $envName $artifactBucket
