env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "1-run testing"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS="darwin"
fi
export UCP_REGION=$(aws configure get region)
mkdir ../z-coverage/ucp-batch

go test -v -failfast -coverprofile ../z-coverage/ucp-batch/cover-main.out src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi
