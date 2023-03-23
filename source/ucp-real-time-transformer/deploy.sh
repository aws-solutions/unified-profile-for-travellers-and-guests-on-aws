
env=$1
bucket=$2

echo "**********************************************"
echo "*  UCP Real Time Transformer '$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]; then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
    exit 1
fi
echo "2-removing existing executables"
rm main main.zip

aws s3 cp s3://$bucket/$env/etl/tah_lib.zip tah_lib.zip
unzip -d src/ tah_lib.zip

cd src
zip -r ../main.zip tah_lib index.py
cd ..

echo "3-Organizing dependenies"
if [ $env == $LOCAL_ENV_NAME ]; then
    echo "3.1-Cleaning unused dependencies (local env only)"
    go mod tidy
    echo "3.1-Vendoring dependencies (local env only)"
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

echo "6-Zipping executable"
zip mainAccp.zip main

sh push.sh $env $bucket

rm tah_lib.zip
rm main.zip
rm mainAccp.zip
rm -rf src/tah_lib

