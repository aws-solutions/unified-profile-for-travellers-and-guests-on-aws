
# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#  SPDX-License-Identifier: MIT-0

env=$1
bucket=$2

echo "**********************************************"
echo "*  UCP Glue ETL '$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]; then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
    exit 1
fi

echo "Running unit tests"
python3 -m unittest discover
if [ $? != 0 ]; then
    exit 1
fi

aws s3 cp clickstream/clickstreamToUcp.py s3://$bucket/$env/etl/clickstreamToUcp.py
aws s3 cp clickstream/clickstreamTransform.py s3://$bucket/$env/etl/clickstreamTransform.py

aws s3 cp air_booking/air_bookingToUcp.py s3://$bucket/$env/etl/air_bookingToUcp.py
aws s3 cp air_booking/air_bookingTransform.py s3://$bucket/$env/etl/air_bookingTransform.py

aws s3 cp hotel_booking/hotel_bookingToUcp.py s3://$bucket/$env/etl/hotel_bookingToUcp.py
aws s3 cp hotel_booking/hotel_bookingTransform.py s3://$bucket/$env/etl/hotel_bookingTransform.py

aws s3 cp guest_profile/guest_profileToUcp.py s3://$bucket/$env/etl/guest_profileToUcp.py
aws s3 cp guest_profile/guest_profileTransform.py s3://$bucket/$env/etl/guest_profileTransform.py

aws s3 cp autoFlatten.py s3://$bucket/$env/etl/autoFlatten.py

