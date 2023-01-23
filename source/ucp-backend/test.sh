export UCP_REGION=$(aws configure get region)
echo "Running solution functional tests in region $UCP_REGION"
echo ""
echo "Setting up environmental variables for use in usecase test"
export CONNECT_PROFILE_EXPORT_BUCKET=$(jq -r .connectProfileExportBucket ucp-config.json)
export KMS_KEY_PROFILE_DOMAIN=$(jq -r .kmsKeyProfileDomain ucp-config.json)

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