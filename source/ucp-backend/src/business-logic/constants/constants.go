// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

const ACCP_DOMAIN_NAME_ENV_VAR = "ACCP_DOMAIN_NAME"

const BEDROCK_TEMPERATURE = 0.5

const BEDROCK_TOP_P = 1

const BEDROCK_TOP_K = 250

const BEDROCK_MAX_TOKENS_TO_SAMPLE = 250

const PROMPT_REQUIREMENTS = "You are to generate the summary by only using the information in the json object. If JSON object lacks information do not assume details. The response should only contain the information present in the json. Do not add a preface. Here is the json object: "
