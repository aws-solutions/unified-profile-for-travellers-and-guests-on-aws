path=$1
domain=$2

awsprofile=$(jq -r .awsProfile ../../env.json)

if [ -z "$awsprofile" ]; then
  echo "AWS Profile not provided: running default CLI profile"
else 
  export AWS_PROFILE=$awsprofile
  echo "AWS Profile provided: running profile $AWS_PROFILE"
fi

echo "******************************************************"
echo "** Dispatch generated file in biz object buckets ***"
echo "******************************************************"
loadBucket=$(jq -r .loadBucket ../../env.json)
hotelBooking=$(jq -r .customerBuckethotelbooking ../../env.json)
airbooking=$(jq -r .customerBucketairbooking ../../env.json)
guestProfile=$(jq -r .customerBucketguestprofile ../../env.json)
paxProfile=$(jq -r .customerBucketpaxprofile ../../env.json)
clicstream=$(jq -r .customerBucketclickstream ../../env.json)
hotelStay=$(jq -r .customerBuckethotelstay ../../env.json)

aws s3 rm s3://$hotelBooking/$path/hotelBooking  --recursive 
aws s3 rm s3://$hotelStay/$path/hotelStay  --recursive 
aws s3 rm s3://$clicstream/$path/hotelBooking  --recursive 

aws s3 cp s3://$loadBucket/$path/hotelBooking s3://$hotelBooking/$domain --recursive 
aws s3 cp s3://$loadBucket/$path/hotelStay s3://$hotelStay/$domain --recursive 
aws s3 cp s3://$loadBucket/$path/hotelBooking s3://$clicstream/$domain --recursive 
