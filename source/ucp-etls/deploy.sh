
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

aws s3 cp etls/clickstreamToUcp.py s3://$bucket/$env/etl/clickstreamToUcp.py
aws s3 cp etls/air_bookingToUcp.py s3://$bucket/$env/etl/air_bookingToUcp.py
aws s3 cp etls/hotel_bookingToUcp.py s3://$bucket/$env/etl/hotel_bookingToUcp.py
aws s3 cp etls/guest_profileToUcp.py s3://$bucket/$env/etl/guest_profileToUcp.py
aws s3 cp etls/hotel_stayToUcp.py s3://$bucket/$env/etl/hotel_stayToUcp.py
aws s3 cp etls/pax_profileToUcp.py s3://$bucket/$env/etl/pax_profileToUcp.py

echo "zipping transforms and common code into lib"
zip -r tah_lib.zip tah_lib/* -x tah_lib/__pycache__/**\* -x tah_lib/__pycache__
aws s3 cp tah_lib.zip  s3://$bucket/$env/etl/tah_lib.zip
rm tah_lib.zip

