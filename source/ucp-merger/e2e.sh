env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "Run merger e2e testing"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS="darwin"
fi
export GOWORK=off # prevent issue when running on pipeline, since tah-core is not committed
export UCP_REGION=$(aws configure get region)
go test -v -failfast src/e2e/e2e_test.go
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi