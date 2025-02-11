#!/bin/sh

##########################################################################################
# This script retrieves the latest schemas from the schemas directory, runs go mod tidy, and go mod vendor
# to update the vendor folder and setup the solution for local build and tests.
# It also places the relevant static files (like schemas) in the right folder in the solution to allow it to build.
################################################################################################

# set required variables
envName=$(jq -r .localEnvName ./env.json)
artifactBucket=$(jq -r .artifactBucket ./env.json)
if [ -z "$envName" ] || [ -z "$artifactBucket" ]; then
    echo "One or more required variables not set" >&2
    exit 1
fi

pushd ../schemas
    pushd generators/go-to-spec
        # if config file exists, save it temporarily
        if [ -f config.json ]; then
            mv config.json temp.config.json
        fi
        cp config.tpl.json config.json
    popd
    ./deploy.sh "$envName" "$artifactBucket"
    pushd generators/go-to-spec
        rm config.json
        # if temp config file exists, restore it
        if [ -f temp.config.json ]; then
            mv temp.config.json config.json
        fi
    popd
popd

# update tah-common glue schemas
echo "Getting latest tah-common glue schemas"
rm -rf tah-common/tah-common-glue-schemas/
cp -r ../schemas/glue_schemas tah-common/tah-common-glue-schemas

# update tah-common json schemas
echo "Getting latest tah-common json schemas"
rm -rf tah-common/tah-common-json-schemas/
cp -r ../schemas/schemas tah-common/tah-common-json-schemas

## Clean up Go module
echo "Running go mod tidy"
go mod tidy
echo "Running go mod vendor"
go mod vendor
