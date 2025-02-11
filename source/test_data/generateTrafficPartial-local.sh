###########################
# This script Generate randomized traffic to Kinesis
########################
business_object=$1
domain=$2
object_type=$3
traveller_id=$4
object_id=$5
key=$6
objectkey=$7
value=$8

envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)
sh generateTrafficPartial.sh $artifactBucket $envName $business_object $domain $object_type $traveller_id $object_id $key $objectkey $value