#!/bin/bash

# Exit on error
set -e

echo "Running unit tests"

# Clean up and recreate coverage directories
rm -rf ../z-coverage/tests
mkdir -p ../z-coverage/tests/coverage-reports

# Set paths
coverage_report_path=../z-coverage/tests/coverage-reports/etls-coverage.coverage.xml
source_dir="$(cd $PWD/..; pwd -P)"

echo "Installing required packages..."
python3 -m pip install coverage --break-system-packages

echo "Installing tah_lib..."
if ! python3 -m pip install -e ../tah_lib --break-system-packages; then
    echo "tah_lib install failed"
    exit 1
fi

echo "Running tests with coverage..."
if ! python3 -m coverage run -m unittest discover; then
    echo "Changes have been detected in the transformation code"
    echo "To accept these changes, run sh update-test-data.sh"
    exit 1
fi

# Generate coverage report
python3 -m coverage xml
cp coverage.xml $coverage_report_path

# Update paths in coverage report
sed -i -e "s,<source>$source_dir,<source>source,g" $coverage_report_path
sed -i -e "s,filename=\"$source_dir,filename=\"source,g" $coverage_report_path

# Clean up
rm coverage.xml

echo "Tests completed successfully"
exit 0
