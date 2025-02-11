env=$1
bucket=$2

export UCP_REGION=$(aws configure get region)
export GOWORK=off # prevent issue when running on pipeline, since tah-core is not committed
echo "Running solution functional tests in region $UCP_REGION"
echo ""
mkdir ../z-coverage/ucp-merger
go test -v -failfast -coverprofile ../z-coverage/ucp-merger/cover-main.out src/main/*
rc=$?
if [ $rc -ne 0 ]; then
echo "GO Unit Testing failed." >&2
exit $rc
fi