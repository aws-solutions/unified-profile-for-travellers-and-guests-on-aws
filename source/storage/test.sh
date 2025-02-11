#!/usr/bin/env bash
# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
[[ $DEBUG ]] && set -x
set -euo pipefail

main() {
    local scriptPath=$(pwd -P -- "$0")
    local coverageDir="$scriptPath"/../z-coverage/storage
    mkdir -p "$coverageDir"
    local coveragePath="$coverageDir"/cover.out

    export GOWORK=off

    go test -failfast -timeout 30m -coverprofile "$coveragePath"
}

main "$@"
