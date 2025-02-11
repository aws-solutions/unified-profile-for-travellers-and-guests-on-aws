env=$1
bucket=$2

echo "******************************************************"
echo "* Unified Traveller Profile Front end End to end test "
echo "******************************************************"
if [ -z "$env" ] || [ -z "$bucket" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh e2e.sh <env> <bucket>"
else
    echo "1-Get infra config for env $env"
    aws s3 cp s3://$bucket/config/ucp-config-$env.json cypress/ucp-config.json
    rc=$?
    if [ $rc -ne 0 ]; then
        echo "Failed to retreive infra config. Exiting E2E tests with status $rc" >&2
        exit $rc
    fi
    userPoolId=$(jq -r .cognitoUserPoolId cypress/ucp-config.json)

    echo "2 Creating front end user"
    RANDOM=$$
    time=$(date +"%Y-%m-%d-%H-%M-%S")
    username="user-$time@example.com"
    password="testPassword1\$$RANDOM"
    echo "Creating user $username for testing"
    aws cognito-idp admin-create-user --user-pool-id $userPoolId --username $username --message-action "SUPPRESS" --user-attributes Name="email_verified",Value="true" Name="email",Value=$username
    aws cognito-idp admin-set-user-password --user-pool-id $userPoolId --username $username --password $password --no-permanent
    bitSize=32
    admin=$(printf "%x\n" $((2 ** $bitSize - 1)))
    appAccessGroup="app-global-react-e2e/$admin"
    aws cognito-idp create-group --user-pool-id $userPoolId --group-name $appAccessGroup --description ""
    aws cognito-idp admin-add-user-to-group --user-pool-id $userPoolId  --username $username --group-name $appAccessGroup
    echo "3 Generating cypress.env.json file"
    cat cypress/ucp-config.json | 
        jq "{cloudfrontDomainName, ucpApiUrl, cognitoUserPoolId, cognitoClientId}" | 
        jq ". += { "email": \"$username\", "pwd": \"$password\", "runPasswordChangeLogin": "true"}" > cypress.env.json
    echo "4 Clearing Config file"
    rm cypress/ucp-config.json

    echo "5 Running Cypress E2E tests"
    npm run cypress:run
    rc=$?
    aws s3 sync cypress/videos s3://$bucket/$env/e2e --delete

    echo "6 Deleting front end user and cleaning up config files"
    aws cognito-idp delete-group --user-pool-id $userPoolId --group-name $appAccessGroup
    aws cognito-idp admin-delete-user --user-pool-id $userPoolId --username $username
    if [ $rc -ne 0 ]; then
        echo "Test failed. Exiting E2E tests with status $rc" >&2
        exit $rc
    fi
fi