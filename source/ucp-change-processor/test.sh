env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "4-run testing"
if [[ $env == $LOCAL_ENV_NAME ]]; then
  export GOOS="darwin"
fi
export UCP_REGION=$(aws configure get region)
export GLUE_SCHEMA_PATH=$(cd ../tah-common/ && pwd)
mkdir ../z-coverage/ucp-change-processor

go test -v -short -timeout 30m -failfast -coverprofile ../z-coverage/ucp-change-processor/cover-main.out src/main/*

rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi