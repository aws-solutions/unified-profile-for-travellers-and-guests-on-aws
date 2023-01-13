env=$1
bucket=$2

echo "******************************************************"
echo "* Unified Traveller Profile Front end End to end test "
echo "******************************************************"
if [ -z "$env" ] || [ -z "$bucket" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
else
echo "1-Get infra config for env $env"
aws s3 cp s3://$bucket/config/ucp-config-$env.json ../src/app/ucp-config.json
cat ../src/app/ucp-config.json
cloudfrontDomainName=$(jq -r .cloudfrontDomainName ../src/app/ucp-config.json)
userPoolId=$(jq -r .cognitoUserPoolId ../src/app/ucp-config.json)
clientId=$(jq -r .cognitoClientId ../src/app/ucp-config.json)

echo "2 Creating front end user"
RANDOM=$$
time=$(date +"%Y-%m-%d-%H-%M-%S")
username="user-$time@example.com"
password="testPassword1\$$RANDOM"
echo "{\"email\" : \"$username\", \"pwd\" : \"$password\"}" > src/creds.json
echo "1. Create user $username for testing"
aws cognito-idp admin-create-user --user-pool-id $userPoolId --username $username --message-action "SUPPRESS"
aws cognito-idp admin-set-user-password --user-pool-id $userPoolId --username $username --password $password --permanent 

echo "3-Run protractor tests $env"
ng e2e
rc=$?
echo "4 Deleting front end user and cleaning up config files"
aws cognito-idp admin-delete-user --user-pool-id $userPoolId --username $username
rm src/creds.json
if [ $rc -ne 0 ]; then
echo "Test failed. Existing E2E tests with status $rc" >&2
exit $rc
fi
fi