// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

var DOMAIN_TAG_ENV_NAME = "envName"
var ACCP_SUB_FOLDER_AIR_BOOKING = "air_booking"
var ACCP_SUB_FOLDER_EMAIL_HISTORY = "email_history"
var ACCP_SUB_FOLDER_PHONE_HISTORY = "phone_history"
var ACCP_SUB_FOLDER_AIR_LOYALTY = "air_loyalty"
var ACCP_SUB_FOLDER_CLICKSTREAM = "clickstream"
var ACCP_SUB_FOLDER_GUEST_PROFILE = "guest_profile"
var ACCP_SUB_FOLDER_HOTEL_LOYALTY = "hotel_loyalty"
var ACCP_SUB_FOLDER_HOTEL_BOOKING = "hotel_booking"
var ACCP_SUB_FOLDER_PAX_PROFILE = "pax_profile"
var ACCP_SUB_FOLDER_HOTEL_STAY_MAPPING = "hotel_stay_revenue_items"

var CONNECTOR_NAME_HAPI = "hapi"
var CONNECTOR_NAME_TEALIUM = "tealium"

var BUSINESS_OBJECT_HOTEL_BOOKING = "hotel_booking"
var BUSINESS_OBJECT_HOTEL_STAY = "hotel_stay"
var BUSINESS_OBJECT_GUEST_PROFILE = "guest_profile"
var BUSINESS_OBJECT_CLICKSTREAM = "clickstream"

var ERROR_PK = "ucp_ingestion_error"
var ERROR_SK_PREFIX = "error_"

var PORTAL_CONFIG_HYPERLINKS_PK = "hyperlink_mappings"
var PORTAL_CONFIG_HYPERLINKS_SK_PREFIX = "mapping_"
var PORTAL_CONFIG_PROMPT_PK = "ai_prompt"
var PORTAL_CONFIG_MATCH_THRESHOLD_PK = "match_threshold"

var RULE_TABLE_SK_PREFIX = "rule_"

var CONNECTORS_MAP = map[string][]string{
	CONNECTOR_NAME_HAPI: {
		BUSINESS_OBJECT_HOTEL_BOOKING,
		BUSINESS_OBJECT_HOTEL_STAY,
		BUSINESS_OBJECT_GUEST_PROFILE,
	},
	CONNECTOR_NAME_TEALIUM: {
		BUSINESS_OBJECT_CLICKSTREAM,
	},
}
