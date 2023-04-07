env=$1
bucket=$2

aws s3 cp s3://$bucket/config/ucp-config-$env.json ./ucp-config.json
cat ./ucp-config.json

export GLUE_JOB_NAME_AIR_BOOKING=$(jq -r .customerJobNameairbooking ./ucp-config.json)
export GLUE_JOB_NAME_HOTEL_BOOKINGS=$(jq -r .customerJobNamehotelbooking ./ucp-config.json)
export GLUE_JOB_NAME_PAX_PROFILES=$(jq -r .customerJobNamepaxprofile ./ucp-config.json)
export GLUE_JOB_NAME_GUEST_PROFILES=$(jq -r .customerJobNameguestprofile ./ucp-config.json)
export GLUE_JOB_NAME_STAY_REVENUE=$(jq -r .customerJobNamehotelstay ./ucp-config.json)
export GLUE_JOB_NAME_CLICKSTREAM=$(jq -r .customerJobNameclickstream ./ucp-config.json)

export UCP_REGION=$(aws configure get region)
echo "Running solution functional tests in region $UCP_REGION"
echo ""
mkdir test_coverage
go test -v -failfast -coverprofile test_coverage/cover-main.out src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile test_coverage/cover-maintainGluePartitions.out src/business-logic/usecase/maintainGluePartitions/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi