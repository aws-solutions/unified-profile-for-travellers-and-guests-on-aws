envName=$(jq -r .localEnvName ../../../env.json)
artifactBucket=$(jq -r .artifactBucket ../../../env.json)
cp protractor.conf-local.js protractor.conf.js
sh test.sh $envName $artifactBucket
rc=$?
cp protractor.conf-pipeline.js protractor.conf.js
if [ $rc -ne 0 ]; then
    echo "Test failed. Existing E2E tests with status $rc" >&2
    exit $rc
fi