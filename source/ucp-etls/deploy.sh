
# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#  SPDX-License-Identifier: MIT-0

env=$1
bucket=$2

echo "**********************************************"
echo "*  UCP Glue ETL '$env' "
echo "***********************************************"
if [ -z "$env" || -z "$bucket"]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
else
    aws s3 cp bookingToUcp.py s3://$bucket/$env/etl/bookingToUcp.py
    aws s3 cp clickstreamToUcp.py s3://$bucket/$env/etl/clickstreamToUcp.py
    aws s3 cp loyaltyToUcp.py s3://$bucket/$env/etl/loyaltyToUcp.py
    aws s3 cp connectProfileToAmperity.py s3://$bucket/$env/etl/connectProfileToAmperity.py
fi

