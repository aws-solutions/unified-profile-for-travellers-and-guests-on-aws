#!/bin/sh

env=$1
bucket=$2

echo "**********************************************"
echo "* Error Ingector '$env' "
echo "***********************************************"
if [ -z "$env" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh test.sh <env>"
else

echo "1-Getting Stack information"
aws s3 cp s3://$bucket/config/ucp-config-$env.json ./ucp-config.json
cat ./ucp-config.json

KINESIS_NAME_REAL_TIME=$(jq -r .kinesisStreamNameRealTime ./ucp-config.json)
ACCP_DLQ=$(jq -r .accpDomainErrorQueue ./ucp-config.json)

aws kinesis put-record --data "aW52YWxpZCByZWNvcmQgZm9yIHRlc3Rpbmc=" --partition-key "test-patition" --stream-name $KINESIS_NAME_REAL_TIME
aws sqs send-message --queue-url $ACCP_DLQ \
 --message-body '{"model_version":"1.0","object_type":"hotel_stay_revenue","last_updated":"2023-03-10T20:47:22.186296521Z","last_updated_by":"Kathlyn Dietrich","created_on":"2023-03-10T16:54:54.525225581Z","created_by":"Lysanne Sporer","traveller_id":"8006113981","id":"TVZFB3XMWW","booking_id":"UZTLKV4YDP","currency_code":"SLL","currency_name":"Sierra Leone Leone","currency_symbol":"","first_name":"Reyes","last_name":"Paucek","email":"","phone":"3478.272.9525","start_date":"2023-03-06","hotel_code":"TOK09","type":"Room Charge 2023-03-06","description":"Suite with Sea View ( Xmas special rate)","amount":"11785.533","date":"2023-03-06T00:00:00Z"}' \
 --message-attributes '{"DomainName": {"StringValue":"test-domain","DataType":"String"},"Message": {"StringValue":"test message","DataType":"String"},"ObjectTypeName": {"StringValue":"test-domain","DataType":"String"}}'

fi