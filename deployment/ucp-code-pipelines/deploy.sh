env=$1
email=$2
token=$3
solutionEnvName=$4
githubUsername=$5
buildFromUpstream=$6
branch=$7
echo "env=$1"
echo "email=$2"
echo "token=$3"
echo "solutionEnvName=$4"
echo "githubUsername=$5"
echo "buildFromUpstream=$6"
echo "branch=$7"

#update this variable to specify the name of your loval env
echo "**********************************************"
echo "* Unified Profile for travellers and guests on AWS Pipeline '$env' "
echo "***********************************************"
if [ -z "$env" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env>"
    exit 1
else  
    echo "Checking pre-requisits'$env' "
    OUTRegion=$(aws configure get region)
    if [ -z "$OUTRegion" ]; then
        echo "The AWS CLI on this envitonement is not configured with a region."
        echo "Usage:"
        echo "aws configure set region <region-name> (ex: us-west-1)"
        exit 1
    fi
    echo "Deployment Actions for env: '$env' "
    echo "***********************************************"
    echo "1.0-Building stack for environement '$env' "
    npm run build
    #This section allows user to deploy only 1 stack or all stacks
    echo "1.1-Synthesizing CloudFormation template for environement '$env' "
    cdk synth -c envName=$env
    cfn_nag_scan --input-path cdk.out/UCPCodePipelinesStack.template.json
    echo "1.2-Analyzing changes for environment '$env' "
    cdk diff -c envName=$env
    echo "1.3-Deploying infrastructure for environement '$env' "
    cdk deploy  -c envName=$env --format=json --require-approval never --parameters contactEmail=$email --parameters gitHubUserName=$githubUsername --parameters githubtoken=$token --parameters buildFromUpstream=$buildFromUpstream --parameters branch=$branch
    rc=$?
    if [ $rc -ne 0 ]; then
      echo "CDK Deploy Failed! Existing Build with status $rc" >&2
      exit $rc
    fi
fi
