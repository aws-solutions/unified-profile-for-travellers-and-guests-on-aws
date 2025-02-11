#!/bin/bash
[[ $DEBUG ]] && set -x
set -eo pipefail

# Array of directories to test
dirs=("storage" "ucp-async" "ucp-backend" "ucp-batch" "ucp-change-processor"  "ucp-common"
      "ucp-cp-indexer" "ucp-cp-writer" "ucp-error" "ucp-etls" "ucp-fargate" "ucp-infra" 
      "ucp-match" "ucp-merger" "ucp-portal/ucp-react" "ucp-real-time-transformer"
      "ucp-retry" "ucp-s3-excise-queue-processor" "ucp-sync")       


cd ../source
# Store original directory
ORIGINAL_DIR=$(pwd)

# Initialize counters and arrays
passed=0
failed=0
failed_dirs=()
skipped_dirs=()

for dir in "${dirs[@]}"; do
    if [ -d "$dir" ]; then
        echo "===================================="
        echo "Running tests in $dir"
        echo "===================================="
        
        cd "$dir" || continue
        
        # Run the test.sh script in each directory
        if sh ./test.sh; then
            ((passed++))
            echo "Tests passed in $dir"
        else
            rc=$?
            ((failed++))
            failed_dirs+=("$dir")
            echo "Tests failed in $dir with status $rc" >&2
        fi
        
        cd "$ORIGINAL_DIR" || exit 1
    else
        echo "Directory $dir not found, skipping..."
        skipped_dirs+=("$dir")
    fi
done

echo "===================================="
echo "Test Summary"
echo "===================================="
echo "Directories tested: $((passed + failed))"
echo "Passed: $passed"
echo "Failed: $failed"

# Print failed directories if any
if [ ${#failed_dirs[@]} -ne 0 ]; then
    echo -e "\nFailed directories:"
    printf '  - %s\n' "${failed_dirs[@]}"
fi

# Print skipped directories if any
if [ ${#skipped_dirs[@]} -ne 0 ]; then
    echo -e "\nSkipped directories:"
    printf '  - %s\n' "${skipped_dirs[@]}"
fi

# Exit with failure if any tests failed
if [ $failed -ne 0 ]; then
    echo -e "\nSome tests failed. Check the logs above for details." >&2
    exit 1
fi

echo -e "\nAll test suites passed successfully!"
exit 0
