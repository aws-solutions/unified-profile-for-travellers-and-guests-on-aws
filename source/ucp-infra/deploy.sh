
# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#  SPDX-License-Identifier: MIT-0
env=$1
bucket=$2

echo "**********************************************"
echo "* AWS Unified Customer Profile: Infrastructure in environement '$env' and artifact bucket '$bucket'"
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket> <contactEmail>"
else
    echo "Checking pre-requiresits'$env' "
    OUTRegion=$(aws configure get region)

    if [ -z "$OUTRegion" ]; then
        echo "The AWS CLI on this envitonement is not configured with a region."
        echo "Usage:"
        echo "aws configure set region <region-name> (ex: us-west-1)"
        exit 1
    fi
    echo "DEPLOYEMENT ACTIONS'$env' "
    echo "***********************************************"
    echo "1.0-Building stack for environement '$env' "
    npm run build

    echo "1.1-Synthesizing CloudFormation template for environement '$env' "
    cdk --app bin/ucp-infra.js synth -c envName=$env -c artifactBucket=$bucket
    echo "1.2-Analyzing changes for environment '$env' "
    cdk --app bin/ucp-infra.js diff -c envName=$env -c artifactBucket=$bucket
    echo "1.3-Deploying infrastructure for environement '$env' "
    cdk --app bin/ucp-infra.js deploy UCPInfraStack$env -c envName=$env -c artifactBucket=$bucket --require-approval never
    rc=$?
    if [ $rc -ne 0 ]; then
      echo "CDK Deploy Failed! Existing Build with status $rc" >&2
      exit $rc
    fi


    echo "Post-Deployment Actions for env '$env' "
    echo "***********************************************"
    echo ""
    echo "Generating Configuration'$env' "
    echo "3.0-Removing existing config filr for '$env' (if exists)"
    rm ucp-infra-config-$env.json
    echo "3.1-Generating new config file for env '$env'"
    apiEndpoint=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='hcmApiUrl'].OutputValue" --output text)
    userPoolId=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='userPoolId'].OutputValue" --output text)
    clientId=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='cognitoAppClientId'].OutputValue" --output text)
    tokenEnpoint=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='tokenEnpoint'].OutputValue" --output text)
    cognitoDomain=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='cognitoDomain'].OutputValue" --output text)
    ucpApiId=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='ucpApiId'].OutputValue" --output text)
    httpApiUrl=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='httpApiUrl'].OutputValue" --output text)
    websiteDistributionId=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='websiteDistributionId'].OutputValue" --output text)
    cloudfrontDomainName=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='websiteDomainName'].OutputValue" --output text)
    websiteBucket=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='websiteBucket'].OutputValue" --output text)
   
    echo "3.2 Creating admin User and getting refresh token"
    RANDOM=$$
    time=$(date +"%Y-%m-%d-%H-%M-%S")
    username="hcm-admin-$time@example.com"
    password="testPassword1\$$RANDOM"
    echo "{\"username\" : \"$username\"}" > userInfo.json
    echo "1. Create user $username for testing"
    aws cognito-idp admin-create-user --user-pool-id $userPoolId --username $username --message-action "SUPPRESS"
    aws cognito-idp admin-set-user-password --user-pool-id $userPoolId --username $username --password $password --permanent 
    aws cognito-idp admin-initiate-auth --user-pool-id $userPoolId\
                                        --client-id $clientId\
                                        --auth-flow ADMIN_USER_PASSWORD_AUTH\
                                        --auth-parameters USERNAME=$username,PASSWORD=$password> loginInfo.json
    cat loginInfo.json
    refershToken=$(jq -r .AuthenticationResult.RefreshToken loginInfo.json)
    rm loginInfo.json

    summary='
    <h3>AWS UCP Configuration </h3></br>
    --------------------------------------------------------------------------------------------</br>
    | Cognito URL             |  '$tokenEnpoint'</br>
    --------------------------------------------------------------------------------------------</br>
    | API Gateway ID         |  '$ucpApiId'</br>
    --------------------------------------------------------------------------------------------</br>
    | API Gateway URL         |  '$httpApiUrl'</br>
    --------------------------------------------------------------------------------------------</br>
    | Client ID               |  '$clientId'</br>
    --------------------------------------------------------------------------------------------</br>
    | Refresh Token           |  '$refershToken'</br>
    --------------------------------------------------------------------------------------------</br>
    | Cognito Domain        |  '$cognitoDomain'</br>
    --------------------------------------------------------------------------------------------</br>
    | Environment             |  '$env'</br>
    --------------------------------------------------------------------------------------------</br>
    | Region                  |  '$OUTRegion'</br>
    --------------------------------------------------------------------------------------------</br>
    | Cognito User Pool ID    |  '$userPoolId'</br>
    --------------------------------------------------------------------------------------------</br>
    | Password                |  '$password'</br>
    --------------------------------------------------------------------------------------------</br>
    | UCP Portal URL          |  '$ucpPortalUrl'</br>
    --------------------------------------------------------------------------------------------</br>
    | Cloufront Distribution  |  '$websiteDistributionId'</br>
    --------------------------------------------------------------------------------------------</br>
    | Portal URL              |  '$cloudfrontDomainName'</br>
    --------------------------------------------------------------------------------------------</br>
    | Website Bucket          |  '$websiteBucket'</br>
    --------------------------------------------------------------------------------------------</br>'

    echo "$summary"
    
   # echo "Sending stack outputs to email: $email"
   #aws ses send-email --from "$email" --destination "ToAddresses=$email" --message "Subject={Data=Your IOT Connectiviity Quickstart deployment Output,Charset=utf8},Body={Text={Data=$summary,Charset=utf8},Html={Data=$summary,Charset=utf8}}"
 
 echo "{"\
         "\"env\" : \"$env\","\
         "\"ucpApiUrl\" : \"$httpApiUrl\","\
         "\"ucpApiId\" : \"$ucpApiId\","\
         "\"cognitoUserPoolId\" : \"$userPoolId\","\
         "\"cognitoClientId\" : \"$clientId\","\
         "\"cognitoRefreshToken\" : \"$refershToken\","\
         "\"cognitoDomain\" : \"$cognitoDomain\","\
         "\"password\" : \"$password\","\
         "\"tokenEnpoint\" : \"$tokenEnpoint\","\
         "\"contentBucket\" : \"$websiteBucket\","\
         "\"cloudfrontDomainName\" : \"$cloudfrontDomainName\","\
         "\"websiteDistributionId\" : \"$websiteDistributionId\","\
         "\"region\":\"$OUTRegion\""\
         "}">infra-config-$env.json
    cat infra-config-$env.json
aws s3 cp infra-config-$env.json s3://$bucket/config/ucp-config-$env.json

fi  
