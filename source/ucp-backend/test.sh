env=$1
bucket=$2

export UCP_REGION=$(aws configure get region)
echo "Running solution functional tests in region $UCP_REGION"
echo ""
echo "Setting up environmental variables for use in usecase test"
aws s3 cp s3://$bucket/config/ucp-config-$env.json ucp-config.json
cat ucp-config.json
export CONNECT_PROFILE_SOURCE_BUCKET=$(jq -r .connectProfileExportBucket ucp-config.json)
export KMS_KEY_PROFILE_DOMAIN=$(jq -r .kmsKeyProfileDomain ucp-config.json)
rm ucp-config.json

go test -v -failfast src/business-logic/usecase/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
