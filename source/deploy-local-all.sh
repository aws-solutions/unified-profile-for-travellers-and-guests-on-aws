#!/bin/bash
[[ $DEBUG ]] && set -x
set -eo pipefail

# Disable AWS Pager while running the script, allowing it
# to run without requiring any user interaction
# https://docs.aws.amazon.com/cli/latest/userguide/cli-usage-pagination.html#cli-usage-pagination-clientside
export AWS_PAGER=""

# Validate Docker is running, otherwise infra deploy will fail
if ! docker info >/dev/null 2>&1; then
    echo "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Validate Python is installed
py_cmd=python
if ! command -v python &> /dev/null; then
    echo "Python is not installed. Checking Python3."
    if ! command -v python3 &> /dev/null; then
        echo "Python3 is not installed. Please install Python and try again."
        exit 1
    fi
    echo "Python3 is installed. Using Python3."
    py_cmd=python3
fi

# Check for venv, set up and activate if necessary
if [ -z "$VIRTUAL_ENV" ]; then
    echo "There is not an active venv"
    pushd ..
    if [ ! -d ".venv" ]; then
        echo "No venv found, creating new venv"
        $py_cmd -m venv .venv
    fi
    echo "Activating venv"
    # shellcheck disable=SC1091 # disable warning for not finding .venv/bin/activate
    source .venv/bin/activate
    pip3 install --upgrade pip boto3 pytest coverage requests --break-system-packages 
    popd
fi

# Parse command-line options
skip_tests=false
skip_errors=false
while [[ "$#" -gt 0 ]]; do
  case $1 in
    --skip-tests)
      skip_tests=true
      ;;
    --skip-errors)
      skip_errors=true
      ;;
    *)
      echo "Unrecognized option: $1"
      exit 1
      ;;
  esac
  shift
done

# Update Private Packages
sh update-private-packages.sh
rc=$?
if [ "$rc" -ne 0 ]; then
  echo "******************************************************"
  echo "Update private packages failed with with status $rc" >&2
  echo "******************************************************"
  if [ "$skip_errors" == false ]; then
    exit "$rc"
  else
    echo
    echo "...Ignoring errors. Continuing anyway."
    echo
  fi
fi

# run tah_lib tests
pushd tah_lib
$py_cmd -m pip install tox
$py_cmd -m tox
popd

# Update build command
deploy_cmd="sh deploy-local.sh"
if [ "$skip_tests" = true ]; then
  deploy_cmd+=" --skip-tests"
fi
deploy_cmd+=" --skip-update-deps"
echo "********************************"
echo "Skip tests: $skip_tests"
echo "Skip errors: $skip_errors"
echo "********************************"

# Deploy Front End dist.zip
echo "********************************"
echo "* Deploying React Front End"
echo "********************************"

envName=$(jq -r .localEnvName env.json)
artifactBucket=$(jq -r .artifactBucket env.json)
region=$(aws configure get region)
PREV_PWD=$PWD

cd ucp-portal/ucp-react || exit
if [ "$skip_tests" = true ]; then
  sh build-local.sh --skip-tests
else
  sh build-local.sh
fi
rc=$?
# Check if the front end build command was successful
if [ "$rc" -ne 0 ]; then
  echo "******************************************************"
  echo "Deploy front end failed with with status $rc" >&2
  echo "******************************************************"
  if [ "$skip_errors" == false ]; then
    exit "$rc"
  else
    echo
    echo "...Ignoring errors. Continuing anyway."
    echo
  fi
fi
aws s3 cp ../../../deployment/regional-s3-assets/ui/dist.zip s3://"${artifactBucket}"-"${region}"/"${envName}"/ui/dist.zip
cd "$PREV_PWD" || exit

# Create an array with all folder names to iteratively call $deploy_cmd
folders=("ucp-etls" "ucp-backend" "ucp-sync" "ucp-change-processor" "ucp-error" "ucp-match" "ucp-merger" "ucp-real-time-transformer" "ucp-industry-connector-transfer" "ucp-retry" "ucp-async" "ucp-s3-excise-queue-processor" "ucp-batch" "ucp-cp-writer" "ucp-cp-indexer" "ucp-infra")

# Loop over the folders array and call $deploy_cmd for each folder
for folder in "${folders[@]}"; do
  echo "********************************"
  echo "* Deploying $folder"
  echo "********************************"

  # cd into the folder
  PREV_PWD=$PWD
  cd "$folder" || exit

  # Call $deploy_cmd
  $deploy_cmd

  rc=$?

  # Check if the deploy command was successful
  if [ "$rc" -ne 0 ]; then
    echo "******************************************************"
    echo "Deploy $folder failed with with status $rc" >&2
    echo "******************************************************"
    if [ "$skip_errors" == false ]; then
      exit "$rc"
    else
      echo
      echo "...Ignoring errors. Continuing anyway."
      echo
    fi
  fi

  # cd back to the parent folder
  cd "$PREV_PWD" || exit
done

# Invalidate Cloudfront Distribution
aws s3 cp s3://"$artifactBucket"/config/ucp-config-"$envName".json ucp-config.json
distributionId=$(jq -r .websiteDistributionId ucp-config.json)
echo "Invalidating Cloudfront Distribution: $distributionId"
aws cloudfront create-invalidation --distribution-id "$distributionId" --path "/*" > /dev/null
rm ucp-config.json
