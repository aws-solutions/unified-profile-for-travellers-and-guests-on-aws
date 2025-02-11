###########################
# This script Generate partial data traffic to Kinesis
########################

echo "***********************************************"
echo "* Generate real time traffic for UCP stream "
echo "***********************************************"

bucket=$1
env=$2
business_object=$3
domain=$4
object_type=$5
traveller_id=$6
object_id=$7
key=$8
objectkey=$9
value=${10}

aws s3 cp s3://$bucket/config/ucp-config-$env.json ./ucp-config.json
export KINESIS_NAME_REAL_TIME=$(jq -r .kinesisStreamNameRealTime ./ucp-config.json)

json_file=$business_object/empty.json
object_prefix=""

if [ $object_type == "hotel_loyalty" ]; then
    object_prefix=.loyaltyPrograms[0]
fi

echo "Generate real time traffic for UCP stream $key key $value value"

#new_value=$(jq --arg id "$traveller_id" --arg object_id "$object_id" --arg object_prefix "$object_prefix" '.id = $id | $object_prefix = $object_id' "$json_file")
new_value=$(jq --arg id "$traveller_id" --arg object_id "$object_id" --arg value "$value" ".id = \$id | $object_prefix.id = \$object_id | $object_prefix.$key = \$value" "$json_file")

#type BusinessObjectRecord struct {
#    Domain       string        `json:"domain"`
#    ObjectType   string        `json:"objectType"`
#    ModelVersion string        `json:"modelVersion"`
#    Data         interface{}   `json:"data"`
#}

file=temp.json
echo $new_value > $file

while read json; do

    #type BusinessObjectRecord struct {
	#    TravellerID        string                    `json:"travellerId"`
	#    TxID               string                    `json:"transactionId"`
	#    Domain             string                    `json:"domain"`
	#    Mode               string                    `json:"mode"`
	#    MergeModeProfileID string                    `json:"mergeModeProfileId"`
	#    PartialModeOptions PartialModeOptions        `json:"partialModeOptions"`
	#    ObjectType         string                    `json:"objectType"`
	#    ModelVersion       string                    `json:"modelVersion"`
	#    Data               []interface{}             `json:"data"`
	#    KinesisRecord      events.KinesisEventRecord // original record required when sending to error queue
    #}

    #type PartialModeOptions struct {
	#    Fields []string `json:"fields"`
    #}
    parent_json="{\"domain\": \"$domain\",\"objectType\":\"$business_object\", \"data\": $json, \"mode\": \"partial\", \"partialModeOptions\":{\"fields\": [\"$object_type.$objectkey\"]}}"
    echo "Sending payload to kinesis stream:  $parent_json"
    #base64 ecnode the payload
    parent_json_base64=$(echo $parent_json | base64)
    aws kinesis put-record --stream-name $KINESIS_NAME_REAL_TIME --data $parent_json_base64 --partition-key $file &
done < $file

rm -rf $file