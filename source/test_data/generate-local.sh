###########################
# This script replaces the daata in the solutiono S3 Bucket with 
#  fresh randomly generated data. NOte that existing data is cleared
#############################

envName=$(jq -r .localEnvName ../env.json)
artifactBucket=$(jq -r .artifactBucket ../env.json)
sh generate.sh $envName $artifactBucket $1
