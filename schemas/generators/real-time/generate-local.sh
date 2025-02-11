#!/bin/bash

envJson="./env.json"
numProfiles=$(jq -r .numProfiles $envJson)
if [ -z "$numProfiles" ] || [ "$numProfiles" == "null" ]; then
    echo "numProfiles not provided"
    exit 1
fi
businessObject=$(jq -r .businessObject $envJson)
if [ -z "$businessObject" ] || [ "$businessObject" == "null" ]; then
    echo "businessObject not provided"
    exit 1
fi
traffic=$(jq -r .traffic $envJson)
if [ -z "$traffic" ] || [ "$traffic" == "null" ]; then
    echo "traffic not provided"
    exit 1
fi
domain=$(jq -r .domain $envJson)
if [ -z "$domain" ] || [ "$domain" == "null" ]; then
    echo "domain not provided"
    exit 1
fi

sh ./generate.sh "$numProfiles" "$businessObject" "$traffic" "$domain"
