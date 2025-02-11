# Ssi Core



## Description

This repository contains common golang code for AWS SSI solutions

## Setup Steps

### running tests for all components

run  `sh deploy-local.sh` to run all tests

### running tests for a single components

run  `sh deploy-local.sh <comp>` to run tests for s specific component

note that when using this, all components will be add to the archive and uploaded to S3. if you then use this version in a solution
you will update all components (possible some that have not been tested)

 
## Service Specific Notes

### Aurora

- Make sure you are on VPN. 
- the first test will create the instance.
- update the network security group with the prefix list giving 
- make teh aurora instance public

### Redshift

- the first run will create a namespace and a secret manager secret
- add the secret Arn to your env.json file as documented in redshift_test.go

### Bedrock

- make sure you activate all the relevant models as described in bedrock_test.go

### Fargate

- this tets is really long since we have to wait for the data to be successfully ingested (can take up to 3 mins). We should have an option to make it optional
