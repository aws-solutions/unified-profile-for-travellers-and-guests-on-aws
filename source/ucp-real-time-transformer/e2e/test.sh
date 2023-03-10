env=$1
bucket=$2
#update this variable to specify the name of your loval env
LOCAL_ENV_NAME=dev

echo "1-Getting Stack information"
aws s3 cp s3://$bucket/config/ucp-config-$env.json ./ucp-config.json
cat ./ucp-config.json

export LAMBDA_NAME_REAL_TIME=$(jq -r .lambdaFunctionNameRealTimeTest ./ucp-config.json)
export KINESIS_NAME_REAL_TIME=$(jq -r .kinesisStreamNameRealTimeTest ./ucp-config.json)
export KINESIS_NAME_OUTPUT_REAL_TIME=$(jq -r .kinesisStreamOutputNameRealTimeTest ./ucp-config.json)

echo "2-removing existing executables"
rm main main.zip
echo "3-Organizing dependenies"
if [ $env == $LOCAL_ENV_NAME ]; then
    echo "3.1-Cleaning unused dependencies (local env only)"
    go mod tidy
    echo "3.1-Vendoring dependencies (local env only)"
    go mod vendor
fi
echo "4-run testing"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS=darwin
fi
export UCP_REGION=$(aws configure get region)
go test -v -failfast src/kinesis/*
rc=$?
rm -rf src/tah_core
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi
fi