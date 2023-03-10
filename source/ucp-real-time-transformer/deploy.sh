
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
unzip -d src/ tah_lib.zip

cd e2e/src
zip -r ../../main.zip tah_lib index.py
cd ../..

sh push.sh $env $bucket

rm tah_lib.zip
rm main.zip

