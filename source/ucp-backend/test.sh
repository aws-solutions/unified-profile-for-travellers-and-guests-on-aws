export UCP_REGION=$(aws configure get region)
echo "Running solution functional tests in region $UCP_REGION"
echo ""
echo "Setting up environmental variables for use in usecase test"
export S3_Booking_Name=$(jq -r .S3_Booking_Name ucp-config.json)
export S3_Clickstream_Name=$(jq -r .S3_Clickstream_Name ucp-config.json)
export S3_AirBooking_Name=$(jq -r .S3_AirBooking_Name ucp-config.json)
export S3_HotelStayRevenue_Name=$(jq -r .S3_HotelStayRevenue_Name ucp-config.json)
export S3_GuestProfile_Name=$(jq -r .S3_GuestProfile_Name ucp-config.json)
export S3_PassengerProfile_Name=$(jq -r .S3_PassengerProfile_Name ucp-config.json)
export KMS=$(jq -r .KMS ucp-config.json)

go test -v -failfast src/business-logic/usecase/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi
go test -v -failfast src/main/*
rc=$?
if [ $rc -ne 0 ]; then
  echo "GO Unit Testing failed." >&2
  exit $rc
fi