#!/bin/bash

env=$1
bucket=$2
tahCommonVersion=$(jq -r '."tah-common"' ../../tah.json)

echo "WARNING - this will delete existing business object buckets"
echo "and replace it with newly generated test data."
read -p "Do you want to proceed? (y/n) " yn
case $yn in
    y ) ;;
    n ) echo "Exiting..."
        exit;;
    * ) echo "Invalid response. Please respond 'y' or 'n'.";
        exit 1;;
esac

echo "***************************************"
echo "* Generating test data for env '$env' *"
echo "***************************************"

if [ -z "$env" ] || [ -z "$bucket" ]; then
    echo "Error: environment and bucket must be specified."
    echo "Usage: sh generate.sh <env> <bucket>"
fi

# start=`date +%s`
echo "1 - Getting stack information"
aws s3 cp s3://$bucket/config/ucp-config-$env.json ./ucp-config.json
export BUCKET_AIR_BOOKING=$(jq -r .customerBucketairbooking ./ucp-config.json)
export BUCKET_HOTEL_BOOKINGS=$(jq -r .customerBuckethotelbooking ./ucp-config.json)
export BUCKET_PAX_PROFILES=$(jq -r .customerBucketpaxprofile ./ucp-config.json)
export BUCKET_GUEST_PROFILES=$(jq -r .customerBucketguestprofile ./ucp-config.json)
export BUCKET_STAY_REVENUE=$(jq -r .customerBuckethotelstay ./ucp-config.json)
export BUCKET_CLICKSTREAM=$(jq -r .customerBucketclickstream ./ucp-config.json)

echo "2 - Generating test data"
mkdir temp temp/temp # Workaround because the data generator built was for specific folder depth
cd temp/temp
aws s3 cp s3://$bucket/$env/$tahCommonVersion/main ./main
chmod +x main
./main 50 10
cd ../../
rm -rf temp

# TODO: experiment with concurrent deletion https://stackoverflow.com/questions/24843570/concurrency-in-shell-scripts
echo "3 - Clearing existing bucket data"
aws s3 rm s3://$BUCKET_AIR_BOOKING --recursive --quiet
aws s3 rm s3://$BUCKET_HOTEL_BOOKINGS --recursive --quiet
aws s3 rm s3://$BUCKET_PAX_PROFILES --recursive --quiet
aws s3 rm s3://$BUCKET_GUEST_PROFILES --recursive --quiet
aws s3 rm s3://$BUCKET_STAY_REVENUE --recursive --quiet
aws s3 rm s3://$BUCKET_CLICKSTREAM --recursive --quiet

echo "4 - Uploading new data"
aws s3 cp examples/airBooking s3://$BUCKET_AIR_BOOKING --recursive --quiet
aws s3 cp examples/hotelBooking s3://$BUCKET_HOTEL_BOOKINGS --recursive --quiet
aws s3 cp examples/passengerProfile s3://$BUCKET_PAX_PROFILES --recursive --quiet
aws s3 cp examples/hotelGuest s3://$BUCKET_GUEST_PROFILES --recursive --quiet
aws s3 cp examples/hotelStay s3://$BUCKET_STAY_REVENUE --recursive --quiet
aws s3 cp examples/guestClicktream s3://$BUCKET_CLICKSTREAM --recursive --quiet

echo "5 - Cleaning up resources"
rm ucp-config.json
rm -rf examples
rm -rf schemas

echo "Successfully replaced test data!"
# end=`date +%s`
# runtime=$((end-start))
# echo "Runtime: $runtime"
