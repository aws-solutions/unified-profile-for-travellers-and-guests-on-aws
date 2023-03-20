
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

aws s3 cp s3://$bucket/$env/etl/tah_lib.zip tah_lib.zip
unzip -d e2e/src/ tah_lib.zip

cd e2e/src
zip -r ../../main.zip tah_lib index.py
cd ../..

cd e2e

echo "2-removing existing executables"
rm main main.zip
echo "3-Organizing dependenies"
if [ $env == $LOCAL_ENV_NAME ]; then
    echo "3.1-Cleaning unused dependencies (local env only)"
    go mod tidy
    echo "3.1-Vendoring dependencies (local env only)"
    go mod vendor
fi
GOARCH=amd64
GOOS=linux
go build -mod=vendor -o main src/main/main.go

echo "6-Zipping executable"
zip ../mainAccp.zip main

cd ..

sh push.sh $env $bucket

rm tah_lib.zip
rm main.zip
#rm mainAccp.zip
rm -rf e2e/src/tah_lib

