###########################
# This script Clears the data in the solutiono source buckets
#############################

#!/bin/bash

env=$1
bucket=$2
domain=$3

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

if [ -z "$env" ] || [ -z "$bucket" ] || [ -z "$domain" ]; then
    echo "Error: environment, bucket and domain must be specified."
    echo "Usage: sh cleanup.sh <env> <bucket> <domain>"
    exit 1
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


# TODO: experiment with concurrent deletion https://stackoverflow.com/questions/24843570/concurrency-in-shell-scripts
echo "3 - Clearing existing bucket data"
aws s3 rm s3://$BUCKET_AIR_BOOKING/$domain --recursive --quiet
aws s3 rm s3://$BUCKET_HOTEL_BOOKINGS/$domain --recursive --quiet
aws s3 rm s3://$BUCKET_PAX_PROFILES/$domain --recursive --quiet
aws s3 rm s3://$BUCKET_GUEST_PROFILES/$domain --recursive --quiet
aws s3 rm s3://$BUCKET_STAY_REVENUE/$domain --recursive --quiet
aws s3 rm s3://$BUCKET_CLICKSTREAM/$domain --recursive --quiet

echo "Successfully clean-up test data!"

rm ucp-config.json

