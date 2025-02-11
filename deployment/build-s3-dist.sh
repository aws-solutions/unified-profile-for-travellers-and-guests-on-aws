#!/bin/bash
#
# This script packages your project into a solution distributable that can be
# used as an input to the solution builder validation pipeline.
#
# Important notes and prereq's:
#   1. The initialize-repo.sh script must have been run in order for this script to
#      function properly.
#   2. This script should be run from the repo's /deployment folder.
#
# This script will perform the following tasks:
#   1. Remove any old dist files from previous runs.
#   2. Build and synthesize your CDK project.
#   3. Organize source code artifacts into the /regional-s3-assets folder.
#   4. Remove any temporary files used for staging.
#
# Parameters:
#  - source-bucket-base-name: Name for the S3 bucket location where the template will source the Lambda
#    code from. The template will append '-[region_name]' to this bucket name.
#    For example: ./build-s3-dist.sh solutions v1.0.0
#    The template will then expect the source code to be located in the solutions-[region_name] bucket
#  - solution-name: name of the solution for consistency
#  - version-code: version of the package
#-----------------------
# Formatting
set -x

bold=$(tput bold)
normal=$(tput sgr0)
#------------------------------------------------------------------------------
# SETTINGS
#------------------------------------------------------------------------------
# Note: should match package.json
template_format="json"
run_helper="false"

# run_helper is false for yaml - not supported
[[ $template_format == "yaml" ]] && {
    run_helper="false"
    echo "${bold}Solution_helper disabled:${normal} template format is yaml"
}

#------------------------------------------------------------------------------
# DISABLE OVERRIDE WARNINGS
#------------------------------------------------------------------------------
# Use with care: disables the warning for overridden properties on
# AWS Solutions Constructs
export overrideWarningsEnabled=false

#------------------------------------------------------------------------------
# Build Functions
#------------------------------------------------------------------------------
# Echo, execute, and check the return code for a command. Exit if rc > 0
# ex. do_cmd npm run build
usage() {
    echo "Usage: $0 bucket solution-name version"
    echo "Please provide the base source bucket name, trademarked solution name, and version."
    echo "For example: ./build-s3-dist.sh mybucket my-solution v1.0.0"
    exit 1
}

do_cmd() {
    echo "------ EXEC $*"
    $*
    rc=$?
    if [ $rc -gt 0 ]; then
        echo "Aborted - rc=$rc"
        exit $rc
    fi
}

sedi() {
    # cross-platform for sed -i
    sed -i $* 2>/dev/null || sed -i "" $*
}

# use sed to perform token replacement
# ex. do_replace myfile.json %%VERSION%% v1.1.1
do_replace() {
    replace="s/$2/$3/g"
    file=$1
    do_cmd sedi $replace $file
}

#------------------------------------------------------------------------------
# INITIALIZATION
#------------------------------------------------------------------------------
# solution_config must exist in the deployment folder (same folder as this
# file) . It is the definitive source for solution ID, name, and trademarked
# name.
#
# Example:
#
# SOLUTION_ID='SO0111'
# SOLUTION_NAME='AWS Security Hub Automated Response & Remediation'
# SOLUTION_TRADEMARKEDNAME='aws-security-hub-automated-response-and-remediation'
# SOLUTION_VERSION='v1.1.1' # optional
if [[ -e './solution_config' ]]; then
    source ./solution_config
else
    echo "solution_config is missing from the solution root."
    exit 1
fi

if [[ -z $SOLUTION_ID ]]; then
    echo "SOLUTION_ID is missing from ../solution_config"
    exit 1
else
    export SOLUTION_ID
fi

if [[ -z $SOLUTION_NAME ]]; then
    echo "SOLUTION_NAME is missing from ../solution_config"
    exit 1
else
    export SOLUTION_NAME
fi

if [[ -z $SOLUTION_TRADEMARKEDNAME ]]; then
    echo "SOLUTION_TRADEMARKEDNAME is missing from ../solution_config"
    exit 1
else
    export SOLUTION_TRADEMARKEDNAME
fi

if [[ ! -z $SOLUTION_VERSION ]]; then
    export SOLUTION_VERSION
fi

#------------------------------------------------------------------------------
# Validate command line parameters
#------------------------------------------------------------------------------
# Validate command line input - must provide bucket
[[ -z $1 ]] && {
    usage
    exit 1
} || { SOLUTION_BUCKET=$1; }

# Environmental variables for use in CDK
export DIST_OUTPUT_BUCKET=$SOLUTION_BUCKET

# Version from the command line is definitive. Otherwise, use, in order of precedence:
# - SOLUTION_VERSION from solution_config
# - version.txt
#
# Note: Solutions Pipeline sends bucket, name, version. Command line expects bucket, version
# if there is a 3rd parm then version is $3, else $2
#
# If confused, use build-s3-dist.sh <bucket> <version>
if [ ! -z $3 ]; then
    version="$3"
elif [ ! -z "$2" ]; then
    version=$2
elif [ ! -z $SOLUTION_VERSION ]; then
    version=$SOLUTION_VERSION
elif [ -e ../source/version.txt ]; then
    version=$(cat ../source/version.txt)
else
    echo "Version not found. Version must be passed as an argument or in version.txt in the format vn.n.n"
    exit 1
fi
SOLUTION_VERSION=$version

# SOLUTION_VERSION should be vn.n.n
if [[ $SOLUTION_VERSION != v* ]]; then
    echo prepend v to $SOLUTION_VERSION
    SOLUTION_VERSION=v${SOLUTION_VERSION}
fi

export SOLUTION_VERSION=$version

PUBLIC_ECR_REGISTRY=$4
if [ -z $PUBLIC_ECR_REGISTRY ]; then
    echo "No ECR registry provided"
    exit 1
fi
export PUBLIC_ECR_REGISTRY=$PUBLIC_ECR_REGISTRY

PUBLIC_ECR_TAG=$5
if [ -z $PUBLIC_ECR_TAG ]; then
    echo "No ECR tag provided"
    exit 1
fi
export PUBLIC_ECR_TAG=$PUBLIC_ECR_TAG

#-----------------------------------------------------------------------------------
# Get reference for all important folders
#-----------------------------------------------------------------------------------
template_dir="$PWD"
staging_dist_dir="$template_dir/staging"
template_dist_dir="$template_dir/global-s3-assets"
build_dist_dir="$template_dir/regional-s3-assets"
source_dir="$template_dir/../source"

echo "------------------------------------------------------------------------------"
echo "${bold}[Init] Remove any old dist files from previous runs${normal}"
echo "------------------------------------------------------------------------------"

do_cmd rm -rf $template_dist_dir
do_cmd mkdir -p $template_dist_dir
do_cmd rm -rf $build_dist_dir
do_cmd mkdir -p $build_dist_dir
do_cmd rm -rf $staging_dist_dir
do_cmd mkdir -p $staging_dist_dir
rm -rf ./ecr

echo "------------------------------------------------------------------------------"
echo "${bold}[Create] Templates${normal}"
echo "------------------------------------------------------------------------------"

envname=prod
# Install the global aws-cdk package
# Note: do not install using global (-g) option. This makes build-s3-dist.sh difficult
# for customers and developers to use, as it globally changes their environment.
do_cmd cd $source_dir/ucp-infra
do_cmd npm ci
do_cmd sh build.sh $envname $SOLUTION_BUCKET $SOLUTION_TRADEMARKEDNAME/$SOLUTION_VERSION 7 true public no-role --skip-tests
do_cmd mv $template_dist_dir/ucp.template.json $template_dist_dir/ucp.template
do_cmd mv $template_dist_dir/ucp-byo-vpc.template.json $template_dist_dir/ucp-byo-vpc.template
# Find and replace bucket_name, solution_name, and version
echo "Find and replace bucket_name, solution_name, and version"
cd $template_dist_dir
do_replace "*.template" %%BUCKET_NAME%% ${SOLUTION_BUCKET}
do_replace "*.template" %%SOLUTION_NAME%% ${SOLUTION_TRADEMARKEDNAME}
do_replace "*.template" %%VERSION%% ${SOLUTION_VERSION}

cd ..
rm -rf ../../staging
mkdir -p ../../staging/upt-id-res
cp -r .. ../../staging/upt-id-res
mkdir -p ../../staging/upt-fe
cp -r ../source/Dockerfile.frontend ../../staging/upt-fe/Dockerfile
cp -r ../source/appPermissions.json ../../staging/upt-fe
cp -r ../source/ucp-portal ../../staging/upt-fe
mkdir -p ./ecr
mv ../../staging/upt-id-res ./ecr/
mv ../../staging/upt-fe ./ecr/
ls -al ./ecr/upt-id-res
ls -al ./ecr/upt-fe

echo "------------------------------------------------------------------------------"
echo "${bold}[Packing] Source code artifacts${normal}"
echo "------------------------------------------------------------------------------"
#adding a dummy file in case there is no regional assets
do_cmd cd $source_dir/ucp-etls
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/tah_lib
do_cmd python -m pip install tox
do_cmd python -m tox
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-change-processor
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-error
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-industry-connector-transfer
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-match
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-merger
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-portal/ucp-react
do_cmd npm ci
# --skip-tests flag is removed since UI unit tests do not use real resources
do_cmd sh build.sh $envname $SOLUTION_BUCKET
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-real-time-transformer
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-sync
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-backend
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-retry
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-async
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-s3-excise-queue-processor
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
do_cmd cd $source_dir/ucp-batch
do_cmd sh build.sh $envname $SOLUTION_BUCKET --skip-tests
rc=$?
if [ $rc -ne 0 ]; then
    echo "Existing Build with status $rc" >&2
    exit $rc
fi
# Return to original directory from when we started the build
cd $template_dir

header() {
    declare text=$1
    echo "------------------------------------------------------------------------------"
    echo "$bold$text$normal"
    echo "------------------------------------------------------------------------------"
}

header "[CDK-Helper] Copy the Lambda Assets"
pushd $source_dir/ucp-infra
assetsFile=./cdk.out/UCPInfraStack"$envName".assets.json
dockerImages=$(jq -r '.dockerImages | keys | join(" ")' "$assetsFile")
popd
pushd "$template_dir"/cdk-solution-helper
cdk_out_dir="$source_dir"/ucp-infra/cdk.out
regional_dist_dir="$build_dist_dir"
npm ci
npx ts-node ./index "$cdk_out_dir" "$regional_dist_dir"
for image in $dockerImages; do rm ../regional-s3-assets/"$image".zip; done
pushd "$cdk_out_dir"
for pythonAsset in $(ls asset.*.py); do
    cp "$pythonAsset" ../../../deployment/regional-s3-assets/${pythonAsset#asset.}
done
popd
popd

pushd $template_dir/regional-s3-assets
zip -r allArtifacts.zip *
mv allArtifacts.zip $template_dir/global-s3-assets
popd
