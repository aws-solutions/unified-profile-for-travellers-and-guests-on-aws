env=$1
stackName=$2
#update this variable to specify the name of your loval env
echo "**********************************************"
echo "* IOT Onboarding Code Pipeline project '$env' "
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
        cdk synth -c envName=$env > iot-onboarding-$env.yml
        echo "1.2-Analyzing changes for environment '$env' "
        cdk diff -c envName=$env
        rc=$?
    if [ $rc -ne 0 ]; then
      echo "CDK Deploy Failed! Existing Build with status $rc" >&2
      exit $rc
    fi
fi
