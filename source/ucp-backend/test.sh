env=$1
bucket=$2

export UCP_REGION=$(aws configure get region)
echo "Running solution functional tests in region $UCP_REGION"
echo ""
mkdir test_coverage
go test -v -failfast -coverprofile test_coverage/cover-registry.out src/business-logic/usecase/registry/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile test_coverage/cover-mappings.out src/business-logic/model/accp-mappings/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile test_coverage/cover-utils.out src/business-logic/utils/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile test_coverage/cover-validator.out src/business-logic/validator/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile test_coverage/cover-usecase.out src/business-logic/usecase/admin/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile test_coverage/cover-usecase.out src/business-logic/usecase/traveller/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile test_coverage/cover-usecase.out src/business-logic/usecase/registry/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast -coverprofile test_coverage/cover-main.out src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
