#!/bin/bash

# Exit on error
set -e

echo "Running tests for UCP Infrastructure..."

# Clean TypeScript build
echo "Cleaning TypeScript build..."
npx tsc --build --clean

# Run main tests
echo "Running main tests..."
if ! npm test 2>&1; then
    echo "Main tests failed"
    echo "Inspect your code changes or run \`npm test -- -u\` to update snapshots."
    exit 1
fi

# Copy coverage reports
cp -r coverage/ ../z-coverage/ucp-infra
rm -r coverage

# Run custom resource tests
echo "Running custom resource tests..."
cd custom_resource || exit
if ! npm run package; then
    echo "Custom resource tests failed"
    exit 1
fi

# Copy custom resource coverage
cp -r coverage/ ../../z-coverage/ucp-infra/custom_resource
rm -r coverage

echo "All tests completed successfully"
exit 0
