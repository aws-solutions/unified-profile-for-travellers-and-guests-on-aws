#!/bin/bash

env=$1

echo "Running ssi-core unit tests in env $env"

lock_file=".test_lock"

echo "Cleaning up previous test output"
rm -f $lock_file
rm -r zlogs/*
mkdir zlogs
rm -r ../z-coverage/ssi-core/*
mkdir ../z-coverage/ssi-core

components=$(ls -d */ | grep -v zlogs | tr -d '/')
echo "Starting tests for the following compoenents: $components"
for component in $components; do
    (
        output_file="zlogs/$component.log"
        echo "Starting tests for $component"
        echo "****************************" >> "$output_file"
        echo "Running tests for $component" >> "$output_file"
        echo "****************************" >> "$output_file"
        echo "" >> "$output_file"
        go test -v  -timeout 30m -coverprofile ../z-coverage/ssi-core/"$component".out -failfast "$component"/*>> "$output_file" 2>&1
        if [ $? -ne 0 ]; then
            echo "Failed tests for $component. See zlogs/$component.log for detailed logs"
            cat "$output_file"
            echo "$component tests failed" > "$lock_file"
        else
            echo "Successfully completed tests for $component. See zlogs/$component.log for detailed logs"
            cat zlogs/"$component".log | grep coverage
        fi
    ) &
done
wait

if [ -f "$lock_file" ]; then
    echo "Lockfile indicates the following tests have failed:"
    cat "$lock_file"
    rm -f "$lock_file"
    exit 1
fi

echo "All tests completed successfully"