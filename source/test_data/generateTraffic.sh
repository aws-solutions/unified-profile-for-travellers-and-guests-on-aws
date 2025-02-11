###########################
# This script Generate randomized traffic to Kinesis
########################

echo "***********************************************"
echo "* Generate real time traffic for UCP stream "
echo "***********************************************"

bucket=$1
env=$2
tahCommonVersion=$3
n_objects=$4
n_profiles=$5
business_object=$6
domain=$7
nLines=$8

aws s3 cp s3://$bucket/config/ucp-config-$env.json ./ucp-config.json
export KINESIS_NAME_REAL_TIME=$(jq -r .kinesisStreamNameRealTime ./ucp-config.json)

echo "1 - Generating test data locally"
rm -rf examples
mkdir temp temp/temp # Workaround because the data generator built was for specific folder depth
    cd temp/temp
    aws s3 cp s3://$bucket/$env/$tahCommonVersion/main ./main
    chmod +x main
    ./main $n_objects $n_profiles local $domain all $nLines
    cd ../../
    rm -rf temp

echo "2 - Updating folder names"
mv examples/airBooking examples/air_booking
mv examples/guestClicktream examples/clickstream
cp -r examples/passengerClicktream/* examples/clickstream/*
mv examples/hotelGuest examples/guest_profile
mv examples/hotelBooking examples/hotel_booking
mv examples/hotelStay examples/hotel_stay
mv examples/passengerProfile examples/pax_profile
mv examples/guestCustomerServiceInteraction examples/customer_service_interaction
cp -r examples/passengerCustomerServiceInteraction/* examples/customer_service_interaction/*


for bo in "air_booking" "clickstream" "guest_profile" "hotel_booking" "hotel_stay" "pax_profile" "customer_service_interaction"; do
    if [ $business_object == "all" ] || [ $business_object == $bo ]; then
        echo "3 - Send business object $bo to kinesis stream for domain $domain"
        #list all the leaf objects under air booking folder
        find "examples/$bo" -type f | while read -r file; do
            # Do something with each file
            echo "Processing file: $file"
            # put the content of the file in kinesis stream
            #we only take the first line in case this is a jonl file
            while read json; do
                #type BusinessObjectRecord struct {
                #    Domain       string        `json:"domain"`
                #    ObjectType   string        `json:"objectType"`
                #    ModelVersion string        `json:"modelVersion"`
                #    Data         interface{}   `json:"data"`
                #}
                parent_json="{\"domain\": \"$domain\",\"objectType\":\"$bo\", \"mode\":\"\", \"data\":$json}"
                echo "Sending payload to kinesis stream:  $parent_json"
                #base64 ecnode the payload
                parent_json_base64=$(echo $parent_json | base64)
                aws kinesis put-record --stream-name $KINESIS_NAME_REAL_TIME --data $parent_json_base64 --partition-key $file &
            done < $file
        done
    fi
done


#mv examples/airBooking examples/air_booking
#mv examples/guestClicktream examples/clickstream
#mv examples/hotelGuest examples/guest_profile
#mv examples/hotelBooking examples/hotel_booking
#mv examples/hotelStay examples/hotel_stay
#mv examples/passengerProfile examples/pax_profile
#rm -rf examples
