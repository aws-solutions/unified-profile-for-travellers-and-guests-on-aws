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
echo "0 - Getting stack information"
aws s3 cp s3://$bucket/config/ucp-config-$env.json ./ucp-config.json
export BUCKET_AIR_BOOKING=$(jq -r .customerBucketairbooking ./ucp-config.json)
export BUCKET_HOTEL_BOOKINGS=$(jq -r .customerBuckethotelbooking ./ucp-config.json)
export BUCKET_PAX_PROFILES=$(jq -r .customerBucketpaxprofile ./ucp-config.json)
export BUCKET_GUEST_PROFILES=$(jq -r .customerBucketguestprofile ./ucp-config.json)
export BUCKET_STAY_REVENUE=$(jq -r .customerBuckethotelstay ./ucp-config.json)
export BUCKET_CLICKSTREAM=$(jq -r .customerBucketclickstream ./ucp-config.json)

echo "1 - Generating test data"
mkdir temp temp/temp # Workaround because the data generator built was for specific folder depth
cd temp/temp
aws s3 cp s3://$bucket/$env/$tahCommonVersion/main ./main
chmod +x main
./main 10 5
cd ../../
rm -rf temp

echo "2 - Updating folder names"
mv examples/airBooking examples/air_booking
mv examples/guestClicktream examples/clickstream
mv examples/hotelGuest examples/guest_profile
mv examples/hotelBooking examples/hotel_booking
mv examples/hotelStay examples/hotel_stay
mv examples/passengerProfile examples/pax_profile

sh ./update_test_data.sh

# TODO: experiment with concurrent deletion https://stackoverflow.com/questions/24843570/concurrency-in-shell-scripts
echo "3 - Clearing existing bucket data"
aws s3 rm s3://$BUCKET_AIR_BOOKING --recursive --quiet
aws s3 rm s3://$BUCKET_HOTEL_BOOKINGS --recursive --quiet
aws s3 rm s3://$BUCKET_PAX_PROFILES --recursive --quiet
aws s3 rm s3://$BUCKET_GUEST_PROFILES --recursive --quiet
aws s3 rm s3://$BUCKET_STAY_REVENUE --recursive --quiet
aws s3 rm s3://$BUCKET_CLICKSTREAM --recursive --quiet

echo "4 - Uploading new data"
aws s3 cp examples/air_booking s3://$BUCKET_AIR_BOOKING --recursive --quiet
aws s3 cp examples/hotel_booking s3://$BUCKET_HOTEL_BOOKINGS --recursive --quiet
aws s3 cp examples/pax_profile s3://$BUCKET_PAX_PROFILES --recursive --quiet
aws s3 cp examples/guest_profile s3://$BUCKET_GUEST_PROFILES --recursive --quiet
aws s3 cp examples/hotel_stay s3://$BUCKET_STAY_REVENUE --recursive --quiet
aws s3 cp examples/clickstream s3://$BUCKET_CLICKSTREAM --recursive --quiet

echo "5 - Cleaning up resources"
rm ucp-config.json
rm -rf examples
rm -rf schemas

echo "Successfully replaced test data!"
# end=`date +%s`
# runtime=$((end-start))
# echo "Runtime: $runtime"
