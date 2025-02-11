env=$1
bucket=$2

export UCP_REGION=$(aws configure get region)
export GLUE_SCHEMA_PATH=$(cd ../tah-common/ && pwd)
echo "Running solution functional tests in region $UCP_REGION"
echo ""
mkdir ../z-coverage/ucp-backend
go test -v -failfast -coverprofile ../z-coverage/ucp-backend/cover-usecase.out src/business-logic/usecase/admin/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-backend/cover-registry.out src/business-logic/usecase/registry/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-backend/cover-usecase-traveller.out src/business-logic/usecase/traveller/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-backend/cover-filter.out src/business-logic/usecase/task/filter/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-backend/cover-usecase-privacy.out src/business-logic/usecase/privacy/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile ../z-coverage/ucp-backend/cover-main.out src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
