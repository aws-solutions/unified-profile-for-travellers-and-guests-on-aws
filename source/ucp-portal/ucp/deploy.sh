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
echo "0-switching constant file to app.constants-env.ts"
cp src/app/app.constants-env.ts src/app/app.constants.ts
cat src/app/app.constants.ts
echo "1-Get infra config for env $env"
aws s3 cp s3://$bucket/config/ucp-config-$env.json src/app/ucp-config.json

echo "2-Build application package for env $env"
rm -rf dist
ng build --prod --aot
echo "3-Get deploy bucket name from infra config (replace this aws cli call by local json parsing)"
OutS3BucketAdmin=$(aws cloudformation describe-stacks --stack-name CloudrackInfraWebStack$env --query "Stacks[0].Outputs[?OutputKey=='s3AdministrationPortal'].OutputValue" --output text)
echo "4-load application to bucket $OutS3BucketAdmin"
aws s3 sync ./dist/cloudrack-fe-admin s3://$OutS3BucketAdmin --delete
echo "5-Change local constant file back for local build (this is to be removed after cicd ready)"
cp src/app/app.constants-local.ts src/app/app.constants.ts
distributionId=$(aws cloudformation describe-stacks --stack-name CloudrackInfraWebStack$env --query "Stacks[0].Outputs[?OutputKey=='cloudrackAdminDistributionId'].OutputValue" --output text)
aws cloudfront create-invalidation --distribution-id $distributionId --path "/*"
fi