env=$1
bucket=$2

echo "**********************************************"
echo "* Deploying Adminitration Portal '$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
else
echo "1-Get infra config for env $env"
aws s3 cp s3://$bucket/config/ucp-config-$env.json public/assets/ucp-config.json
cat public/assets/ucp-config.json
contentBucket=$(jq -r .contentBucket public/assets/ucp-config.json)
distributionId=$(jq -r .websiteDistributionId public/assets/ucp-config.json)

echo "2-Build application package for env $env"
rm -rf dist
npm run build
echo "3-Restructure dist to match custom resource output"
echo "4-load application to bucket $contentBucket"
# loading the dist to artifact S3 bucket to allow the infar script to run
# s3://<artifact-bucket>/<envName>/ui/dist.zip
zip -r dist.zip dist
region=$(aws configure get region)
aws s3 cp dist.zip s3://$bucket-$region/$env/ui/dist.zip
rm dist.zip
aws s3 sync ./dist s3://$contentBucket --delete
echo "5-Invalidate cloudfront distribution $distributionId"
aws cloudfront create-invalidation --distribution-id $distributionId --path "/*" > /dev/null
fi