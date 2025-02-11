#!/bin/bash

# Exit on error
set -e

echo "Running tests for UCP Portal..."

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "Node.js is required but not installed. Please install Node.js first."
    exit 1
fi

echo "Installing dependencies..."
if ! npm install; then
    echo "Failed to install dependencies"
    exit 1
fi

echo "Running unit tests"
if ! npm run test; then
    echo "Unit tests failed"
    exit 1
fi

exit 0
