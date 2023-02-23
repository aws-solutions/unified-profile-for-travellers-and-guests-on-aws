envName=$(jq -r .localEnvName ../../env.json)
artifactBucket=$(jq -r .artifactBucket ../../env.json)

echo "Getting tah dependencies"
tahCoreVersion=$(jq -r '."tah-core"' ../../../tah.json)

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

sh test.sh $envName $artifactBucket
