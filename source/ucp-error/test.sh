env=$1
bucket=$2

export UCP_REGION=$(aws configure get region)
echo "Running solution functional tests in region $UCP_REGION"
echo ""
mkdir ../z-coverage/ucp-error
go test -v -failfast -coverprofile ../z-coverage/ucp-error/cover-main.out src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
