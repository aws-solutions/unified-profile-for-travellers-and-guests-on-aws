env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

export GLUE_SCHEMA_PATH_ASYNC=$(cd ../tah-common/ && pwd)
export ENV_NAME=$env

echo "1-run testing"
if [[ $env == $LOCAL_ENV_NAME ]]; then
  export GOOS="darwin"
fi
mkdir ../z-coverage/ucp-async
export UCP_REGION=$(aws configure get region)
go test -v -failfast -coverprofile ../z-coverage/ucp-async/cover-main.out src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-async/cover-usecase.out src/business-logic/usecase/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi
