
env=$1
bucket=$2
LOCAL_ENV_NAME=dev


##parse --skip-tests flag
skipTests=false
if [ "$3" = "--skip-tests" ]; then
  echo "Skipping tests"
  skipTests=true
fi
echo "**********************************************"
echo "*  ucp-real-time transformer '$env' "
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


echo "******************************"
echo "* Go Lambda Build*"
echo "******************************"

echo "1-Organizing dependenies"

export GOOS=linux
export GOARCH=arm64 
export GOWORK=off # temporarily turn of since Workspace mode does not support vendoring
go build -mod=vendor -o bootstrap src/main/main.go
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  exit $rc
fi

if [ $skipTests = false ]; then
echo "2-running tests"
if [ $env == $LOCAL_ENV_NAME ]; then
  export GOOS=darwin
fi
export GOARCH=amd64
sh ./test.sh $env $bucket
rc=$?
if [ $rc -ne 0 ]; then
  echo "Existing Build with status $rc" >&2
  rm bootstrap
  rm main.zip
  rm mainAccp.zip
  rm -rf src/tah_lib
  rm tah-core.zip
  rm -rf src/tah-core
  exit $rc
fi
fi

echo "3-Zipping executable"
zip mainAccp.zip bootstrap

echo "4-Moving code to deployment folder"
mkdir -p ../../deployment/regional-s3-assets
cp mainAccp.zip  ../../deployment/regional-s3-assets/ucpRealtimeTransformerAccp.zip

echo "5-Cleaning up"
rm bootstrap
rm mainAccp.zip
rm tah-core.zip
rm -rf src/tah-core
rm -rf src/ucp-common
