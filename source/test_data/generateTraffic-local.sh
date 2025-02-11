###########################
# This script Generate randomized traffic to Kinesis
########################
n_objects=$1
n_profiles=$2
business_object=$3
domain=$4
nLines=$5

envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)
sh generateTraffic.sh $artifactBucket $envName latest $n_objects $n_profiles $business_object $domain $nLines
