#!/bin/sh
env=$1
bucket=$2

echo "**********************************************"
echo "* UCP back end deployment for environment '$env' and bucket $bucket"
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]
then
    echo "Environment and Bucket Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
else
echo "Publishing new function code"
aws s3 cp main.zip s3://$bucket/$env/ucpRealTimeTransformer/main.zip
aws lambda update-function-code --function-name ucpRealTimeTransformer$env --zip-file fileb://main.zip
aws lambda update-function-code --function-name ucpRealTimeTransformerTest$env --zip-file fileb://main.zip

aws s3 cp mainAccp.zip s3://$bucket/$env/ucpRealTimeTransformerAccp/mainAccp.zip
aws lambda update-function-code --function-name ucpRealTimeTransformerAccp$env --zip-file fileb://mainAccp.zip
aws lambda update-function-code --function-name ucpRealTimeTransformerAccpTest$env --zip-file fileb://mainAccp.zip
fi