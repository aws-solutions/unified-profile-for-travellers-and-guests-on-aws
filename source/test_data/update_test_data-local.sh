###########################
# This scripts replace the test data use by the transformer uint tests (an other unit test within the solution)
########################


#!/bin/bash
env=$(jq -r .localEnvName ../env.json)
bucket=$(jq -r .artifactBucket ../env.json)

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
    echo "Error: environment, bucket  must be specified."
    echo "Usage: sh generate.sh <env> <bucket>"
    exit 1
fi

# start=`date +%s`
echo "1 - Generating test data"
mkdir temp temp/temp # Workaround because the data generator built was for specific folder depth
cp config.json temp/temp/config.json
cd temp/temp
aws s3 cp s3://$bucket/$env/$tahCommonVersion/main ./main
chmod +x main
./main 10 5 local test_domain
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

rm -rf examples
