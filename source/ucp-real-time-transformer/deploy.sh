
env=$1
bucket=$2
LOCAL_ENV_NAME=dev

echo "**********************************************"
echo "*  UCP Real Time Transformer '$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]; then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
    exit 1
fi
echo "******************************"
echo "* Python lambda Build *"
echo "******************************"

echo "1-removing existing executables"
rm main main.zip

echo "2-Downloading shared python code"
rm -rf src/python/tah_lib
aws s3 cp s3://$bucket/$env/etl/tah_lib.zip tah_lib.zip
unzip -d src/python tah_lib.zip
rm tah_lib.zip

cd src/python
sh test.sh $envName $artifactBucket
rc=$?
if [ $rc -ne 0 ]; then
  echo "error running python unit tests $rc" >&2
  exit $rc
fi
zip -r ../../main.zip tah_lib index.py
cd ../..

echo "******************************"
echo "* Go Lambda Build*"
echo "******************************"

echo "1-Organizing dependenies"
if [ $env == $LOCAL_ENV_NAME ]; then
    echo "1.1-Cleaning unused dependencies (local env only)"
    go mod tidy
    echo "1.1-Vendoring dependencies (local env only)"
    go mod vendor
fi

export GOARCH=amd64
export GOOS=linux
go build -mod=vendor -o main src/main/main.go
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi
echo "2-running tests"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS=darwin
fi
sh ./test.sh $env $bucket
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi
echo "3-Zipping executable"
zip mainAccp.zip main

echo "4-Deploying code to lambdas"
sh push.sh $env $bucket

echo "5-Cleaning up"
rm tah_lib.zip
rm main.zip
rm mainAccp.zip
rm -rf src/tah_lib
rm tah-core.zip
rm -rf src/tah-core

