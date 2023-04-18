# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#  SPDX-License-Identifier: MIT-0
env=$1
bucket=$2
email=$3
partitionStartDate=$4

echo "**********************************************"
echo "* AWS Unified Customer Profile: Infrastructure in environement '$env' and artifact bucket '$bucket'"
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket> <email> <partitionStartDate>"
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
    cdk --app bin/ucp-infra.js deploy UCPInfraStack$env -c envName=$env -c artifactBucket=$bucket -c partitionStartDate=$partitionStartDate --require-approval never
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
    connectProfileExportBucket=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='connectProfileExportBucket'].OutputValue" --output text)
    kmsKeyProfileDomain=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='kmsKeyProfileDomain'].OutputValue" --output text)
    glueDBname=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='glueDBname'].OutputValue" --output text)
    ucpDataAdminRoleName=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='ucpDataAdminRoleName'].OutputValue" --output text)
    customerJobNameairbooking=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerJobNameairbooking'].OutputValue" --output text)
    customerJobNameclickstream=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerJobNameclickstream'].OutputValue" --output text)
    customerJobNameguestprofile=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerJobNameguestprofile'].OutputValue" --output text)
    customerJobNamehotelbooking=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerJobNamehotelbooking'].OutputValue" --output text)
    customerJobNamehotelstay=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerJobNamehotelstay'].OutputValue" --output text)
    customerJobNamepaxprofile=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerJobNamepaxprofile'].OutputValue" --output text)
    customerBuckethotelbooking=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerBuckethotelbooking'].OutputValue" --output text)
    customerBucketairbooking=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerBucketairbooking'].OutputValue" --output text)
    customerBucketguestprofile=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerBucketguestprofile'].OutputValue" --output text)
    customerBucketpaxprofile=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerBucketpaxprofile'].OutputValue" --output text)
    customerBucketclickstream=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerBucketclickstream'].OutputValue" --output text)
    customerBuckethotelstay=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerBuckethotelstay'].OutputValue" --output text)
    customerTestBuckethotelbooking=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerTestBuckethotelbooking'].OutputValue" --output text)
    customerTestBucketairbooking=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerTestBucketairbooking'].OutputValue" --output text)
    customerTestBucketguestprofile=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerTestBucketguestprofile'].OutputValue" --output text)
    customerTestBucketpaxprofile=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerTestBucketpaxprofile'].OutputValue" --output text)
    customerTestBucketclickstream=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerTestBucketclickstream'].OutputValue" --output text)
    customerTestBuckethotelstay=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='customerTestBuckethotelstay'].OutputValue" --output text)
    connectProfileImportBucketOut=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='connectProfileImportsBucketOut'].OutputValue" --output text)
    connectProfileImportBucketTestOut=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='connectProfileImportBucketTestOut'].OutputValue" --output text)
    testTableNameairbooking=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='testTableNameairbooking'].OutputValue" --output text)
    testTableNameclickstream=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='testTableNameclickstream'].OutputValue" --output text)
    testTableNameguestprofile=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='testTableNameguestprofile'].OutputValue" --output text)
    testTableNamehotelbooking=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='testTableNamehotelbooking'].OutputValue" --output text)
    testTableNamehotelstay=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='testTableNamehotelstay'].OutputValue" --output text)
    testTableNamepaxprofile=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='testTableNamepaxprofile'].OutputValue" --output text)
    lambdaFunctionNameRealTime=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='lambdaFunctionNameRealTime'].OutputValue" --output text)
    kinesisStreamNameRealTime=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='kinesisStreamNameRealTime'].OutputValue" --output text)
    kinesisStreamOutputNameRealTime=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='kinesisStreamOutputNameRealTime'].OutputValue" --output text)
    lambdaFunctionNameRealTimeTest=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='lambdaFunctionNameRealTimeTest'].OutputValue" --output text)
    kinesisStreamNameRealTimeTest=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='kinesisStreamNameRealTimeTest'].OutputValue" --output text)
    kinesisStreamOutputNameRealTimeTest=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='kinesisStreamOutputNameRealTimeTest'].OutputValue" --output text)
    dlgRealTimeGo=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='dlgRealTimeGo'].OutputValue" --output text)
    dldRealTimePython=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='dldRealTimePython'].OutputValue" --output text)
    dlgRealTimeGoTest=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='dlgRealTimeGoTest'].OutputValue" --output text)
    dldRealTimePythonTest=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='dldRealTimePythonTest'].OutputValue" --output text)
    accpDomainErrorQueue=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='accpDomainErrorQueue'].OutputValue" --output text)
    dynamodbConfigTableName=$(aws cloudformation describe-stacks --stack-name UCPInfraStack$env --query "Stacks[0].Outputs[?OutputKey=='dynamodbConfigTableName'].OutputValue" --output text)

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
    --------------------------------------------------------------------------------------------</br>
    | Glue DB Name            |  '$glueDBname'</br>
    --------------------------------------------------------------------------------------------</br>
    | Data Admin Role Name    |  '$ucpDataAdminRoleName'</br>
    --------------------------------------------------------------------------------------------</br>
    | AirBooking Job Name     |  '$customerJobNameairbooking'</br>
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
         "\"connectProfileExportBucket\":\"$connectProfileExportBucket\","\
         "\"kmsKeyProfileDomain\":\"$kmsKeyProfileDomain\","\
         "\"glueDBname\":\"$glueDBname\","\
         "\"ucpDataAdminRoleName\":\"$ucpDataAdminRoleName\","\
         "\"customerJobNameairbooking\":\"$customerJobNameairbooking\","\
         "\"customerJobNameclickstream\":\"$customerJobNameclickstream\","\
         "\"customerJobNameguestprofile\":\"$customerJobNameguestprofile\","\
         "\"customerJobNamehotelbooking\":\"$customerJobNamehotelbooking\","\
         "\"customerJobNamehotelstay\":\"$customerJobNamehotelstay\","\
         "\"customerJobNamepaxprofile\":\"$customerJobNamepaxprofile\","\
         "\"customerBuckethotelbooking\":\"$customerBuckethotelbooking\","\
         "\"customerBucketairbooking\":\"$customerBucketairbooking\","\
         "\"customerBucketguestprofile\":\"$customerBucketguestprofile\","\
         "\"customerBucketpaxprofile\":\"$customerBucketpaxprofile\","\
         "\"customerBucketclickstream\":\"$customerBucketclickstream\","\
         "\"customerBuckethotelstay\":\"$customerBuckethotelstay\","\
         "\"customerTestBuckethotelbooking\":\"$customerTestBuckethotelbooking\","\
         "\"customerTestBucketairbooking\":\"$customerTestBucketairbooking\","\
         "\"customerTestBucketguestprofile\":\"$customerTestBucketguestprofile\","\
         "\"customerTestBucketpaxprofile\":\"$customerTestBucketpaxprofile\","\
         "\"customerTestBucketclickstream\":\"$customerTestBucketclickstream\","\
         "\"customerTestBuckethotelstay\":\"$customerTestBuckethotelstay\","\
         "\"connectProfileImportBucketOut\":\"$connectProfileImportBucketOut\","\
         "\"connectProfileImportBucketTestOut\":\"$connectProfileImportBucketTestOut\","\
         "\"testTableNameairbooking\":\"$testTableNameairbooking\","\
         "\"testTableNameclickstream\":\"$testTableNameclickstream\","\
         "\"testTableNameguestprofile\":\"$testTableNameguestprofile\","\
         "\"testTableNamehotelbooking\":\"$testTableNamehotelbooking\","\
         "\"testTableNamehotelstay\":\"$testTableNamehotelstay\","\
         "\"testTableNamepaxprofile\":\"$testTableNamepaxprofile\","\
         "\"lambdaFunctionNameRealTime\":\"$lambdaFunctionNameRealTime\","\
         "\"kinesisStreamNameRealTime\":\"$kinesisStreamNameRealTime\","\
         "\"kinesisStreamOutputNameRealTime\":\"$kinesisStreamOutputNameRealTime\","\
         "\"lambdaFunctionNameRealTimeTest\":\"$lambdaFunctionNameRealTimeTest\","\
         "\"kinesisStreamNameRealTimeTest\":\"$kinesisStreamNameRealTimeTest\","\
         "\"kinesisStreamOutputNameRealTimeTest\":\"$kinesisStreamOutputNameRealTimeTest\","\
         "\"accpDomainErrorQueue\":\"$accpDomainErrorQueue\","\
         "\"dynamodbConfigTableName\":\"$dynamodbConfigTableName\","\
         "\"region\":\"$OUTRegion\""\
         "}">infra-config-$env.json
    cat infra-config-$env.json
aws s3 cp infra-config-$env.json s3://$bucket/config/ucp-config-$env.json

fi  
