#!/bin/bash

SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
PIPELINE_DIR="$SCRIPT_DIR/pipeline_tests"
LOG_DIR="$SCRIPT_DIR/zlogs"
MAX_CONCURRENT_TESTS=4
EXPENSIVE_TEST="TestRefreshCacheEvent" # Runs separately because it creates 100 profiles and tests cache rebuilding with some rules
CHANGE_TEST="TestChangeEvent" # Runs separately to ensure EB and SQS integration results in correct object counts (and is not affected bu any other test running in parallel)
ID_RES_SPLIT="TestIdResMergeSplit" # Runs separately to ensure test doesn't fail on fargate task failing to trigger

mkdir -p "$LOG_DIR"

if [ -d "$PIPELINE_DIR" ]; then
  cd "$PIPELINE_DIR" || exit

  export GOWORK=off # prevent issue when running on pipeline, since tah-core is not committed
  export UCP_REGION=$(aws configure get region)

  TESTS=$(go test -list '.*' | grep -E '^Test' | grep -v "$EXPENSIVE_TEST" | grep -v "$CHANGE_TEST" | grep -v "$ID_RES_SPLIT" | awk '{print $1}')
  running_tests=0
  failed_tests=()

  # Function to wait for any running test to finish
  wait_for_slot() {
    while [ $(jobs -r | wc -l) -ge $MAX_CONCURRENT_TESTS ]; do
      sleep 1
    done
  }

  for TEST in $TESTS; do
    wait_for_slot

    echo "Running $TEST..."
    go test -run "$TEST" -v > "$LOG_DIR/${TEST}.log" 2>&1 &
    running_tests=$((running_tests + 1))
  done

  # Wait for all remaining tests to complete
  wait
  echo "All tests completed. Checking for failures..."

  for TEST in $TESTS; do
    if grep -q "FAIL" "$LOG_DIR/${TEST}.log"; then
      echo "Test failed: $TEST"
      cat "$LOG_DIR/${TEST}.log"
      echo "----"
      failed_tests+=("$TEST")
    fi
  done

  echo "Running TestChangeEvent: $CHANGE_TEST..."
  go test -run "$CHANGE_TEST" -v > "$LOG_DIR/${CHANGE_TEST}.log" 2>&1

  if grep -q "FAIL" "$LOG_DIR/${CHANGE_TEST}.log"; then
    echo "Test failed: $CHANGE_TEST"
    cat "$LOG_DIR/${CHANGE_TEST}.log"
    echo "----"
    failed_tests+=("$CHANGE_TEST")
  fi

  echo "Running the expensive test: $EXPENSIVE_TEST..."
  go test -run "$EXPENSIVE_TEST" -v > "$LOG_DIR/${EXPENSIVE_TEST}.log" 2>&1

  if grep -q "FAIL" "$LOG_DIR/${EXPENSIVE_TEST}.log"; then
    echo "Test failed: $EXPENSIVE_TEST"
    cat "$LOG_DIR/${EXPENSIVE_TEST}.log"
    echo "----"
    failed_tests+=("$EXPENSIVE_TEST")
  fi

  echo "Running TestIdResMergeSplit: $ID_RES_SPLIT..."
  go test -run "$ID_RES_SPLIT" -v > "$LOG_DIR/${ID_RES_SPLIT}.log" 2>&1

  if grep -q "FAIL" "$LOG_DIR/${ID_RES_SPLIT}.log"; then
    echo "Test failed: $ID_RES_SPLIT"
    cat "$LOG_DIR/${ID_RES_SPLIT}.log"
    echo "----"
    failed_tests+=("$ID_RES_SPLIT")
  fi

  echo "Cleaning up log directory..."
  rm -rf "$LOG_DIR"

  # Output the list of failed test names
  if [ ${#failed_tests[@]} -gt 0 ]; then
    echo "List of failed tests:"
    for test in "${failed_tests[@]}"; do
      echo "$test"
    done
    exit 1
  else
    echo "All tests passed."
  fi

else
  echo "The directory 'pipeline_tests' does not exist in the script's location."
  exit 1
fi