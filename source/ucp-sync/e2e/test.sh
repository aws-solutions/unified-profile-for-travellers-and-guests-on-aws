#!/bin/sh

env=$1
echo "**********************************************"
echo "* Ent To Ent Testing Cloudrack Availability for environment '$env' "
echo "***********************************************"
if [ -z "$env" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh test.sh <env>"
else

echo "Not Test Defined for The moment $env"
fi
