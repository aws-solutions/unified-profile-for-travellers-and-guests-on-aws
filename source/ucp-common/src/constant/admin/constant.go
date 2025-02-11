// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"encoding/binary"
	model "tah/upt/source/ucp-common/src/model/admin"
	"tah/upt/source/ucp-common/src/utils/utils"
)

const DOMAIN_TAG_ENV_NAME = "envName"

const BIZ_OBJECT_AIR_BOOKING = "air_booking"
const BIZ_OBJECT_HOTEL_BOOKING = "hotel_booking"
const BIZ_OBJECT_GUEST_PROFILE = "guest_profile"
const BIZ_OBJECT_PAX_PROFILE = "pax_profile"
const BIZ_OBJECT_CLICKSTREAM = "clickstream"
const BIZ_OBJECT_STAY_REVENUE = "hotel_stay_revenue"
const BIZ_OBJECT_CSI = "customer_service_interaction"

var BUSINESS_OBJECTS = []model.BusinessObject{
	{Name: BIZ_OBJECT_HOTEL_BOOKING},
	{Name: BIZ_OBJECT_AIR_BOOKING},
	{Name: BIZ_OBJECT_PAX_PROFILE},
	{Name: BIZ_OBJECT_GUEST_PROFILE},
	{Name: BIZ_OBJECT_CLICKSTREAM},
	{Name: BIZ_OBJECT_STAY_REVENUE},
	{Name: BIZ_OBJECT_CSI},
}

const ACCP_RECORD_AIR_BOOKING = "air_booking"
const ACCP_RECORD_EMAIL_HISTORY = "email_history"
const ACCP_RECORD_PHONE_HISTORY = "phone_history"
const ACCP_RECORD_AIR_LOYALTY = "air_loyalty"
const ACCP_RECORD_LOYALTY_TX = "loyalty_transaction"
const ACCP_RECORD_CLICKSTREAM = "clickstream"
const ACCP_RECORD_GUEST_PROFILE = "guest_profile"
const ACCP_RECORD_HOTEL_LOYALTY = "hotel_loyalty"
const ACCP_RECORD_HOTEL_BOOKING = "hotel_booking"
const ACCP_RECORD_PAX_PROFILE = "pax_profile"
const ACCP_RECORD_HOTEL_STAY_MAPPING = "hotel_stay_revenue_items"
const ACCP_RECORD_CSI = "customer_service_interaction"
const ACCP_RECORD_ANCILLARY = "ancillary_service"
const ACCP_RECORD_ALTERNATE_PROFILE_ID = "alternate_profile_id"

var ACCP_RECORDS = []model.AccpRecord{
	{Name: ACCP_RECORD_AIR_BOOKING},
	{Name: ACCP_RECORD_EMAIL_HISTORY},
	{Name: ACCP_RECORD_PHONE_HISTORY},
	{Name: ACCP_RECORD_AIR_LOYALTY},
	{Name: ACCP_RECORD_LOYALTY_TX},
	{Name: ACCP_RECORD_CLICKSTREAM},
	{Name: ACCP_RECORD_GUEST_PROFILE},
	{Name: ACCP_RECORD_HOTEL_LOYALTY},
	{Name: ACCP_RECORD_HOTEL_BOOKING},
	{Name: ACCP_RECORD_PAX_PROFILE},
	{Name: ACCP_RECORD_HOTEL_STAY_MAPPING},
	{Name: ACCP_RECORD_CSI},
	{Name: ACCP_RECORD_ANCILLARY},
	{Name: ACCP_RECORD_ALTERNATE_PROFILE_ID},
}

const SQS_MES_ATTR_UCP_ERROR_TYPE = "UcpErrorType"
const SQS_MES_ATTR_BUSINESS_OBJECT_TYPE_NAME = "BusinessObjectTypeName"
const SQS_MES_ATTR_MESSAGE = "Message"
const SQS_MES_ATTR_TX_ID = "TransactionId"
const SQS_MES_ATTR_DOMAIN = "DomainName"

const ACCP_OBJECT_TYPE_KEY = "object_type"
const ACCP_OBJECT_ID_KEY = "accp_object_id"
const ACCP_PROFILE_ID_KEY = "profile_id"
const ACCP_OBJECT_TRAVELLER_ID_KEY = "traveller_id"

const CONFIG_DB_LINKED_CONNECTORS_PK = "linked_domains"
const CONFIG_DB_LINKED_CONNECTORS_SK = "industry_connector"

// event detail string to send as CLoudwatch events when manually triggering the UCP SYNC function
const CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER = "ucp_manual_trigger"
const INGESTION_MODE_PARTIAL = "partial"
const INGESTION_MODE_FULL = "full"
const INGESTION_MODE_MERGE = "merge"

const INVALID_OPERATOR_ERROR = "invalid operator in expression"

const DATA_ACCESS_PERMISSION_ALL = "*/*"

const MATCH_INDEX_NAME = "matchesByConfidenceScore"
const MATCH_INDEX_PK_NAME = "runId"
const MATCH_INDEX_SK_NAME = "scoreTargetId"
const MATCH_LATEST_RUN = "latest"
const MATCH_PREFIX = "match_"
const FALSE_POSITIVE_SUFFIX = "_false_positive"
const FALSE_POSITIVE_PREFIX = "false_positive_"
const ACCP_EVENT_STREAM_PREFIX = "ucp_accp_event_stream_"

const INTERACTION_CREATED = "INTERACTION_CREATED"
const MERGE_TYPE_ON_RULE = "MERGE_ON_RULE"
const MERGE_TYPE_ON_AI_MATCH = "MERGE_ON_AI_MATCH"
const MERGE_TYPE_MANUAL = "MERGE_MANUALLY_BY_OPERATOR"

type AppPermission uint32

var AppPermissionSize = binary.Size(AppPermission(0)) * 8

const (
	SearchProfilePermission       AppPermission = 1 << 0
	DeleteProfilePermission       AppPermission = 1 << 1
	MergeProfilePermission        AppPermission = 1 << 2
	UnmergeProfilePermission      AppPermission = 1 << 3
	CreateDomainPermission        AppPermission = 1 << 4
	DeleteDomainPermission        AppPermission = 1 << 5
	ConfigureGenAiPermission      AppPermission = 1 << 6
	SaveHyperlinkPermission       AppPermission = 1 << 7
	SaveRuleSetPermission         AppPermission = 1 << 8
	ActivateRuleSetPermission     AppPermission = 1 << 9
	RunGlueJobsPermission         AppPermission = 1 << 10
	ClearAllErrorsPermission      AppPermission = 1 << 11
	RebuildCachePermission        AppPermission = 1 << 12
	IndustryConnectorPermission   AppPermission = 1 << 13
	ListPrivacySearchPermission   AppPermission = 1 << 14
	GetPrivacySearchPermission    AppPermission = 1 << 15
	CreatePrivacySearchPermission AppPermission = 1 << 16
	DeletePrivacySearchPermission AppPermission = 1 << 17
	PrivacyDataPurgePermission    AppPermission = 1 << 18
	ListRuleSetPermission         AppPermission = 1 << 19
	PublicAccessPermission        AppPermission = AppPermission(0)
	AdminPermission               AppPermission = ^AppPermission(0)
)
const AppAccessPrefix = "app"
const DataAccessPrefix = "admin"

// Utility methods
func AccpRecordsNames() []string {
	var names = []string{}
	for _, record := range ACCP_RECORDS {
		names = append(names, record.Name)
	}
	return names
}

func AccpRecordsMap() map[string]model.AccpRecord {
	m := map[string]model.AccpRecord{}
	for _, r := range ACCP_RECORDS {
		m[r.Name] = r
	}
	return m
}

func IsValidAccpRecord(recName string) bool {
	return utils.ContainsString(AccpRecordsNames(), recName)
}

func PrependSourceField(fieldName string) string {
	return "_source." + fieldName
}
