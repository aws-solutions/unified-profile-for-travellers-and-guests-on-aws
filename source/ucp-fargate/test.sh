# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

export scriptPath=$(pwd -P -- "$0")
export coverageDir="$scriptPath"/../z-coverage/ucp-fargate
mkdir -p "$coverageDir"
export coveragePath="$coverageDir"/cover.out
export GOWORK=off
go test ./... -failfast -timeout 30m -coverprofile "$coveragePath"