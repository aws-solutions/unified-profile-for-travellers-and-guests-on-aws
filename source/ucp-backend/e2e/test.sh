#!/bin/sh

env=$1
bucket=$2
echo "**********************************************"
echo "* End To End Testing Industry connector '$env' "
echo "***********************************************"
if [ -z "$env" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh test.sh <env>"
else

echo "1. getting infra configuration for environement $env"
aws s3 cp s3://$bucket/config/ucp-config-$env.json infra-config.json
cat infra-config.json
#URL of teh API gateway endpoint
apiGatewayUrl=$(jq -r .ucpApiUrl infra-config.json)
#Cognito token endpoint to obstain access token
tokenEndpoint=$(jq -r .tokenEndpoint infra-config.json)
#Cognito Client ID for manual users (to be used with refresh token flow)
clientId=$(jq -r .cognitoClientId infra-config.json)
#Cognito client ID for programatic users (to be sued with client credential grant)
#clientIdApps=$(jq -r .cognitoClientIdApps infra-config.json)
#Cognito user pool ID
userPoolId=$(jq -r .cognitoUserPoolId infra-config.json)



echo "2 Creating admin User and getting refresh token"
RANDOM=$$
echo "Random number initialized with seed: $RANDOM"
time=$(date +"%Y-%m-%d-%H-%M-%S")
username="postman-admin$time@example.com"
password="testPassword1\$$RANDOM"
aws cognito-idp admin-create-user --user-pool-id $userPoolId --username $username --message-action "SUPPRESS"
aws cognito-idp admin-set-user-password --user-pool-id $userPoolId --username $username --password $password --permanent 
aws cognito-idp admin-initiate-auth --user-pool-id $userPoolId\
                                    --client-id $clientId\
                                    --auth-flow ADMIN_USER_PASSWORD_AUTH\
                                    --auth-parameters USERNAME=$username,PASSWORD=$password> loginInfo.json
bitSize=32
admin=$(printf "%x\n" $((2 ** $bitSize - 1)))
appAccessGroup="app-global-realtime-e2e/$admin"
aws cognito-idp create-group --user-pool-id $userPoolId \
                             --group-name $appAccessGroup \
                             --description ""
aws cognito-idp admin-add-user-to-group \
    --user-pool-id $userPoolId  \
    --username $username \
    --group-name $appAccessGroup
cat loginInfo.json
refreshToken=$(jq -r .AuthenticationResult.RefreshToken loginInfo.json)
rm loginInfo.json

clientSecret=$(aws cognito-idp describe-user-pool-client --user-pool-id $userPoolId --client-id $clientId --query "UserPoolClient.q" --output text)
domainName="postman_$RANDOM"
echo "Cognito Domain $domainName"

echo "3. Starting test for environment '$env' on api Gateway enpoint:$apiGatewayUrl "
newman run --verbose --env-var apiGatewayUrl=$apiGatewayUrl --env-var clientSecret=$clientSecret --env-var tokenEndpoint=$tokenEndpoint --env-var clientId=$clientId --env-var refreshToken=$refreshToken --env-var domainName=$domainName --env-var environment=$env ucp.postman_collection.json

rc=$?
echo "4. Cleaning up local files and test users"
rm infra-config.json
#we do not delete the user is the test fails to be able to investigate
if [ $rc -ne 0 ]; then
    echo "E2e Tests failed. exiting with status $rc" >&2
    exit $rc
fi
aws cognito-idp admin-delete-user --user-pool-id $userPoolId --username $username
aws cognito-idp delete-group --user-pool-id $userPoolId --group-name $appAccessGroup
rc=$?
if [ $rc -ne 0 ]; then
    echo "Auth endpoint testing failed"
    echo "Auth response: $authResponse"
    exit 1
fi
fi