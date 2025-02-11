#!/bin/bash

# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

[[ $DEBUG ]] && set -x
set -eo pipefail

generate_password() {
    local length=${1:-8}
    # Use max of input length or 8
    if [ "$length" -lt 8 ]; then
        length=8
    fi

    lowercase="abcdefghijklmnopqrstuvwxyz"
    uppercase="ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    numbers="0123456789"
    special="!#$%^&*()_+"
    allchars="${lowercase}${uppercase}${numbers}${special}"

    # Get one character from each required set
    l_char=${lowercase:$((RANDOM % ${#lowercase})):1}
    u_char=${uppercase:$((RANDOM % ${#uppercase})):1}
    n_char=${numbers:$((RANDOM % ${#numbers})):1}
    s_char=${special:$((RANDOM % ${#special})):1}

    # Generate remaining characters
    password="${l_char}${u_char}${n_char}${s_char}"
    
    for ((i=4; i<length; i++)); do
        rand_index=$((RANDOM % ${#allchars}))
        password="${password}${allchars:$rand_index:1}"
    done
    
    echo "$password"
}

log_if_debug() {
    if [[ $DEBUG ]]; then
        echo "$1"
    fi
}

# Validate Docker is running, otherwise starting container will fail
if ! docker info >/dev/null 2>&1; then
    echo "Docker is not running. Please start Docker and try again."
    exit 1
fi

CONTAINER_NAME="postgres_aurora"
if docker inspect $CONTAINER_NAME >/dev/null 2>&1; then
    echo "Aurora container found. Starting."
    docker start $CONTAINER_NAME
    exit 0
else
    echo "Aurora container not found. Creating."
fi

# Validate that there are valid certificates
if ! [ -f tmp/certs/server.crt ] || ! [ -f tmp/certs/server.key ]; then
    echo "No valid certificates found. Running creation process."
    # Validate that openssl is installed
    if ! openssl version >/dev/null 2>&1; then
        echo "openssl is not installed. Please install openssl and try again."
        exit 1
    fi
    mkdir -p tmp/certs
    pushd tmp/certs || {
        echo "Could not find certs directory. Please create it and try again."
        exit 1
    }
    openssl req -new -x509 -days 3650 -nodes -subj "/CN=localhost" -out server.crt -keyout server.key
    popd
fi

# Validate that env.json exists at the same level of this script
if ! [ -f env.json ]; then
    echo "No env.json found. Please create it based off of env.example.json and try again."
    exit 1
fi

dbName=$(jq -r .auroraDbName env.json)
if [ -z "$dbName" ]; then
    echo "Key: auroraDbName not found. Please add it to env.json and try again."
    exit 1
fi

auroraClusterName=$(jq -r .auroraClusterName env.json)
if [ -z "$auroraClusterName" ]; then
    echo "Key: auroraClusterName not found. Please add it to env.json and try again."
    exit 1
fi

auroraSecretName=$(jq -r .auroraSecretName env.json)
if [ -z "$auroraSecretName" ]; then
    echo "Key: auroraSecretName not found. Please add it to env.json and try again."
    exit 1
fi

if [ ! $auroraSecretName = "aurora-$auroraClusterName" ]; then
    echo: "Key: auroraSecretName must be in the format aurora-<auroraClusterName>. Please update it and try again."
    exit 1
fi

dbUname="tah_core_user"
dbPwd=""
secretArn=$(aws secretsmanager list-secrets --filter Key=name,Values=$auroraSecretName --query "SecretList[]" | jq --arg exact_name "$auroraSecretName" '.[] | select(.Name==$exact_name).ARN')
if [ -z "$secretArn" ]; then
    echo "Aurora secret not found. Creating."
    dbPwd=$(generate_password 40)
    aws secretsmanager create-secret --name $auroraSecretName --secret-string "{\"username\":\"$dbUname\",\"password\":\"$dbPwd\"}" >/dev/null
else
    echo "Aurora secret found.  Retrieving password."
    dbPwd=$(aws secretsmanager get-secret-value --secret-id $auroraSecretName --query SecretString --output text | jq -r .password)
fi

if [ -z $dbPwd ]; then
    echo "Failed to get DB Password from Aurora secret. Please create it manually and try again."
    exit 1
fi

# Ensure container data directory exists
mkdir -p tmp/postgres

# Run Docker Container
docker run \
    -d \
    --name $CONTAINER_NAME \
    -p 5432:5432 \
    -e POSTGRES_USER=$dbUname \
    -e POSTGRES_PASSWORD=$dbPwd \
    -e POSTGRES_DB=$dbName \
    -v $(pwd)/tmp/postgres:/var/lib/postgresql/data \
    -v $(pwd)/tmp/certs/server.crt:/var/lib/postgresql/server.crt \
    -v $(pwd)/tmp/certs/server.key:/var/lib/postgresql/server.key \
    postgres:16-alpine \
    postgres -c ssl=on -c fsync=off -c ssl_cert_file=/var/lib/postgresql/server.crt -c ssl_key_file=/var/lib/postgresql/server.key
    
echo "Aurora Container Started. Use \"docker stop $CONTAINER_NAME\" to stop"