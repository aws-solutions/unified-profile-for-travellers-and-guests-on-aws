# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#  SPDX-License-Identifier: MIT-0
env=$1
bucket=$2
echo "env=$1"
echo "bucket=$2"

echo "Post-Deployment Actions for env '$env' "
    echo "***********************************************"
    echo ""
    echo "Generating Configuration'$env' "
    echo "3.0-Removing existing config filr for '$env' (if exists)"
    rm ucp-infra-config-"$env".json
    echo "3.1-Generating new config file for env '$env'"
    stack_outputs=$(aws cloudformation describe-stacks --stack-name UCPInfraStack"$env" --query "Stacks[0].Outputs" --output json)
    parameters_json=$(aws cloudformation describe-stacks --stack-name UCPInfraStack"$env" --query "Stacks[0].Parameters" --output json)

    apiEndpoint=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="hcmApiUrl") | .OutputValue')
    userPoolId=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="userPoolId") | .OutputValue')
    clientId=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="cognitoAppClientId") | .OutputValue')
    tokenEndpoint=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="tokenEndpoint") | .OutputValue')
    cognitoDomain=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="cognitoDomain") | .OutputValue')
    ucpApiId=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="ucpApiId") | .OutputValue')
    httpApiUrl=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="httpApiUrl") | .OutputValue')
    websiteDistributionId=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="websiteDistributionId") | .OutputValue')
    cloudfrontDomainName=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="websiteDomainName") | .OutputValue')
    websiteBucket=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="websiteBucket") | .OutputValue')
    connectProfileExportBucket=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="connectProfileExportBucket") | .OutputValue')
    kmsKeyProfileDomain=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="kmsKeyProfileDomain") | .OutputValue')
    glueDBname=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="glueDBname") | .OutputValue')
    ucpDataAdminRoleName=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="ucpDataAdminRoleName") | .OutputValue')
    customerJobNameairbooking=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerJobNameairbooking") | .OutputValue')
    customerJobNameclickstream=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerJobNameclickstream") | .OutputValue')
    customerJobNameguestprofile=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerJobNameguestprofile") | .OutputValue')
    customerJobNamehotelbooking=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerJobNamehotelbooking") | .OutputValue')
    customerJobNamehotelstay=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerJobNamehotelstay") | .OutputValue')
    customerJobNamecustomerserviceinteraction=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerJobNamecustomerserviceinteraction") | .OutputValue')
    customerJobNamepaxprofile=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerJobNamepaxprofile") | .OutputValue')
    customerBuckethotelbooking=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerBuckethotelbooking") | .OutputValue')
    customerBucketairbooking=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerBucketairbooking") | .OutputValue')
    customerBucketguestprofile=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerBucketguestprofile") | .OutputValue')
    customerBucketpaxprofile=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerBucketpaxprofile") | .OutputValue')
    customerBucketclickstream=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerBucketclickstream") | .OutputValue')
    customerBuckethotelstay=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerBuckethotelstay") | .OutputValue')
    customerBucketcustomerserviceinteraction=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerBucketcustomerserviceinteraction") | .OutputValue')
    customerTestBuckethotelbooking=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerTestBuckethotelbooking") | .OutputValue')
    customerTestBucketairbooking=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerTestBucketairbooking") | .OutputValue')
    customerTestBucketguestprofile=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerTestBucketguestprofile") | .OutputValue')
    customerTestBucketpaxprofile=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerTestBucketpaxprofile") | .OutputValue')
    customerTestBucketclickstream=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerTestBucketclickstream") | .OutputValue')
    customerTestBuckethotelstay=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerTestBuckethotelstay") | .OutputValue')
    customerTestBucketcustomerserviceinteraction=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="customerTestBucketcustomerserviceinteraction") | .OutputValue')
    connectProfileImportBucketOut=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="connectProfileImportBucketOut") | .OutputValue')
    connectProfileImportBucketTestOut=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="connectProfileImportBucketTestOut") | .OutputValue')
    lambdaFunctionNameRealTime=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="lambdaFunctionNameRealTime") | .OutputValue')
    kinesisStreamNameRealTime=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="kinesisStreamNameRealTime") | .OutputValue')
    kinesisStreamOutputNameRealTime=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="kinesisStreamOutputNameRealTime") | .OutputValue')
    dlgRealTimeGo=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="dlgRealTimeGo") | .OutputValue')
    dldRealTimePython=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="dldRealTimePython") | .OutputValue')
    dlgRealTimeGoTest=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="dlgRealTimeGoTest") | .OutputValue')
    dldRealTimePythonTest=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="dldRealTimePythonTest") | .OutputValue')
    accpDomainErrorQueue=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="accpDomainErrorQueue") | .OutputValue')
    accpDomainErrorQueueTest=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="accpDomainErrorQueueTest") | .OutputValue')
    dynamodbConfigTableName=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="dynamodbConfigTableName") | .OutputValue')
    dynamodbPortalConfigTableName=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="dynamodbPortalConfigTableName") | .OutputValue')
    realTimeFeedBucket=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="realTimeFeedBucket") | .OutputValue')
    realTimeBackupEnabled=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="realTimeBackupEnabled") | .OutputValue')
    dynamoTable=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="dynamoTable") | .OutputValue')
    matchBucket=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="matchBucket") | .OutputValue')
    travellerEventBus=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="travellerEventBus") | .OutputValue')
    outputBucket=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="outputBucket") | .OutputValue')
    changeProcessorStreamName=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="kinesisStreamOutputNameChangeProcessor") | .OutputValue')
    changeProcessorStreamArn=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="kinesisStreamOutputNameChangeProcessorArn") | .OutputValue')
    lcsAuroraProxyEndpoint=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="lcsAuroraProxyEndpoint") | .OutputValue')
    lcsAuroraDbName=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="lcsAuroraDbName") | .OutputValue')
    lcsAuroraDbSecretArn=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="lcsAuroraDbSecretArn") | .OutputValue')
    lcsConfigTableName=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="lcsConfigTableName") | .OutputValue')
    lcsConfigTablePk=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="lcsConfigTablePk") | .OutputValue')
    lcsConfigTableSk=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="lcsConfigTableSk") | .OutputValue')
    mergeQueueUrl=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="mergeQueueUrl") | .OutputValue')
    mergerLambdaName=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="mergerLambdaName") | .OutputValue')
    ssmParamNamespace=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="ssmParamNamespace") | .OutputValue')
    lcsCpWriterQueueUrl=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="lcsCpWriterQueueUrl") | .OutputValue')
    cpExportStream=$(echo "$stack_outputs" | jq -r '.[] | select(.OutputKey=="cpExportStream") | .OutputValue')

    customerProfileStorageMode=$(echo "$parameters_json" | jq -r '.[] | select(.ParameterKey=="customerProfileStorageMode") | .ParameterValue')
    dynamoStorageMode=$(echo "$parameters_json" | jq -r '.[] | select(.ParameterKey=="dynamoStorageMode") | .ParameterValue')
    usePermissionSystem=$(echo "$parameters_json" | jq -r '.[] | select(.ParameterKey=="usePermissionSystem") | .ParameterValue')
    
    summary='
    <h3>AWS UCP Configuration </h3></br>
    --------------------------------------------------------------------------------------------</br>
    | Cognito URL             |  '$tokenEndpoint'</br>
    --------------------------------------------------------------------------------------------</br>
    | API Gateway ID          |  '$ucpApiId'</br>
    --------------------------------------------------------------------------------------------</br>
    | API Gateway URL         |  '$httpApiUrl'</br>
    --------------------------------------------------------------------------------------------</br>
    | Client ID               |  '$clientId'</br>
    --------------------------------------------------------------------------------------------</br>
    | Cognito Domain          |  '$cognitoDomain'</br>
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
    
 echo "{"\
         "\"env\" : \"$env\","\
         "\"ucpApiUrl\" : \"$httpApiUrl\","\
         "\"ucpApiId\" : \"$ucpApiId\","\
         "\"cognitoUserPoolId\" : \"$userPoolId\","\
         "\"cognitoClientId\" : \"$clientId\","\
         "\"cognitoRefreshToken\" : \"$refershToken\","\
         "\"cognitoDomain\" : \"$cognitoDomain\","\
         "\"password\" : \"$password\","\
         "\"tokenEndpoint\" : \"$tokenEndpoint\","\
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
         "\"customerJobNamecustomerserviceinteraction\":\"$customerJobNamecustomerserviceinteraction\","\
         "\"customerJobNamepaxprofile\":\"$customerJobNamepaxprofile\","\
         "\"customerBuckethotelbooking\":\"$customerBuckethotelbooking\","\
         "\"customerBucketairbooking\":\"$customerBucketairbooking\","\
         "\"customerBucketguestprofile\":\"$customerBucketguestprofile\","\
         "\"customerBucketpaxprofile\":\"$customerBucketpaxprofile\","\
         "\"customerBucketclickstream\":\"$customerBucketclickstream\","\
         "\"customerBuckethotelstay\":\"$customerBuckethotelstay\","\
         "\"customerBucketcustomerserviceinteraction\":\"$customerBucketcustomerserviceinteraction\","\
         "\"customerTestBuckethotelbooking\":\"$customerTestBuckethotelbooking\","\
         "\"customerTestBucketairbooking\":\"$customerTestBucketairbooking\","\
         "\"customerTestBucketguestprofile\":\"$customerTestBucketguestprofile\","\
         "\"customerTestBucketpaxprofile\":\"$customerTestBucketpaxprofile\","\
         "\"customerTestBucketclickstream\":\"$customerTestBucketclickstream\","\
         "\"customerTestBuckethotelstay\":\"$customerTestBuckethotelstay\","\
         "\"customerTestBucketcustomerserviceinteraction\":\"$customerTestBucketcustomerserviceinteraction\","\
         "\"connectProfileImportBucketOut\":\"$connectProfileImportBucketOut\","\
         "\"connectProfileImportBucketTestOut\":\"$connectProfileImportBucketTestOut\","\
         "\"lambdaFunctionNameRealTime\":\"$lambdaFunctionNameRealTime\","\
         "\"kinesisStreamNameRealTime\":\"$kinesisStreamNameRealTime\","\
         "\"kinesisStreamOutputNameRealTime\":\"$kinesisStreamOutputNameRealTime\","\
         "\"lambdaFunctionNameRealTimeTest\":\"$lambdaFunctionNameRealTimeTest\","\
         "\"kinesisStreamNameRealTimeTest\":\"$kinesisStreamNameRealTimeTest\","\
         "\"kinesisStreamOutputNameRealTimeTest\":\"$kinesisStreamOutputNameRealTimeTest\","\
         "\"accpDomainErrorQueue\":\"$accpDomainErrorQueue\","\
         "\"dynamodbConfigTableName\":\"$dynamodbConfigTableName\","\
         "\"dynamodbPortalConfigTableName\":\"$dynamodbPortalConfigTableName\","\
         "\"realTimeFeedBucket\":\"$realTimeFeedBucket\","\
         "\"realTimeBackupEnabled\":\"$realTimeBackupEnabled\","\
         "\"realTimeFeedBucketTest\":\"$realTimeFeedBucketTest\","\
         "\"dynamoTable\":\"$dynamoTable\","\
         "\"matchBucket\":\"$matchBucket\","\
         "\"travellerEventBus\":\"$travellerEventBus\","\
         "\"outputBucket\":\"$outputBucket\","\
         "\"changeProcessorStreamName\":\"$changeProcessorStreamName\","\
         "\"lcsAuroraProxyEndpoint\":\"$lcsAuroraProxyEndpoint\","\
         "\"lcsAuroraDbName\":\"$lcsAuroraDbName\","\
         "\"lcsAuroraDbSecretArn\":\"$lcsAuroraDbSecretArn\","\
         "\"lcsConfigTableName\":\"$lcsConfigTableName\","\
         "\"lcsConfigTablePk\":\"$lcsConfigTablePk\","\
         "\"lcsConfigTableSk\":\"$lcsConfigTableSk\","\
         "\"retryLambdaName\":\"$retryLambdaName\","\
         "\"mergeQueueUrl\":\"$mergeQueueUrl\","\
         "\"mergerLambdaName\":\"$mergerLambdaName\","\
         "\"ssmParamNamespace\":\"$ssmParamNamespace\","\
         "\"customerProfileStorageMode\":\"$customerProfileStorageMode\","\
         "\"dynamoStorageMode\":\"$dynamoStorageMode\","\
         "\"usePermissionSystem\":\"$usePermissionSystem\","\
         "\"lcsCpWriterQueueUrl\":\"$lcsCpWriterQueueUrl\","\
         "\"cpExportStream\":\"$cpExportStream\","\
         "\"region\":\"$OUTRegion\""\
         "}">infra-config-"$env".json
    cat infra-config-"$env".json
aws s3 cp infra-config-"$env".json s3://"$bucket"/config/ucp-config-"$env".json