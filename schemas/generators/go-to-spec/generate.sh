echo "*************************************************"
echo "* AWS Basics for Industry: Schema Generator      *"
echo "**************************************************"

envName=$(jq -r .localEnvName ../../env.json)
artifactBucket=$(jq -r .artifactBucket ../../env.json)
awsprofile=$(jq -r .awsProfile ../../env.json)

echo "Getting tah dependencies"

echo "Getting tah-core version"
rm -rf src/tah-core
cp -r ../../../source/tah-core src/.

echo "2-Cleaning up existing example"
rm -rf ../../examples/*
rm -rf ../../schemas/*

echo "3-Copy Source Code"
cp -r ../../src/tah-common src/.

echo "4-Organze dependencies"
go mod tidy -e

echo "5-Generate schema and Examples"
go build -o main src/main/main.go
rc=$?
if [ $rc -ne 0 ]; then
    echo "Go mod tidy: Exiting Build with status $rc" >&2
    exit $rc
fi

if [ -z "$awsprofile" ] || [ "$awsprofile" = "null" ]; then
  echo "AWS Profile not provided: running default CLI profile"
else 
  export AWS_PROFILE=$awsprofile
  echo "AWS Profile provided: running profile $AWS_PROFILE"
fi
export TAH_COMMON_REGION=$(jq -r .region ../../env.json)
export TAH_COMMON_BUCKET=$(jq -r .loadBucket ../../env.json)
export TAH_COMMON_S3_HOTEL_BOOKING_BUCKET=$(jq -r .customerBuckethotelbooking ../../env.json)
export TAH_COMMON_S3_HOTEL_STAY_BUCKET=$(jq -r .customerBuckethotelstay ../../env.json)
export TAH_COMMON_S3_GUEST_PROFILE_BUCKET=$(jq -r .customerBucketguestprofile ../../env.json)
export TAH_COMMON_S3_CLICKSTREAM_BUCKET=$(jq -r .customerBucketclickstream ../../env.json)
export TAH_COMMON_S3_CSI_BUCKET=$(jq -r .customerBucketcsi ../../env.json)

echo "TAH_COMMON_REGION=$TAH_COMMON_REGION"
echo "TAH_COMMON_BUCKET=$TAH_COMMON_BUCKET" 
./main
rc=$?
if [ $rc -ne 0 ]; then
    echo "Go mod tidy: Exiting Build with status $rc" >&2
    exit $rc
fi



