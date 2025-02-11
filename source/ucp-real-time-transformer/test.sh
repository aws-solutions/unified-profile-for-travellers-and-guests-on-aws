env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "4-run testing"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS="darwin"
fi
export UCP_REGION=$(aws configure get region)
export GOWORK=off # prevent issue when running on pipeline, since tah-core is not committed
mkdir ../z-coverage/ucp-real-time-transformer

go test -v -failfast -coverprofile ../z-coverage/ucp-real-time-transformer/cover-main.out src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi