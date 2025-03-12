// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	"errors"
	"time"
)

/************************************************************
* Service Limits
*******************************************************/
const MAX_INTERACTION_BY_TYPE = 100000
const MAX_OBJECT_TYPES_PER_DOMAIN = 100
const MAX_VALUES_PER_SEARCH_QUERY = 100
const MAX_PROFILES_TO_FETCH = 1000
const MAX_CONDITIONS_PER_RULES = 100
const MAX_FIELDS_IN_OBJECT_TYPE = 200
const MAX_DOMAIN_NAME_LENGTH = 26
const MIN_DOMAIN_NAME_LENGTH = 3
const MAX_MAPPING_NAME_LENGTH = 28 // Current limit is our longest mapping (customer_service_interaction)
const MAX_RULE_SET_SIZE = 10000    // Max number of rules in a single rule set

// Profile Retrieve
const MAX_PROFILE_ID_PER_CONNECT_ID = 100000 //max profile ids for a given connect ID
const MAX_INTERACTIONS_PER_PROFILE = 100000  //max interactions per profile

// Profile Search
const MAX_SEARCH_RESPONSE = 1000     // Max size of profile search response
const MIN_CHAR_FOR_APPROX_SEARCH = 3 // at least 3 char for approx search

/************************************************************
* Identity Resolution
*******************************************************/
const RULE_OP_EQUALS = "equals"
const RULE_OP_EQUALS_VALUE = "equals_value"
const RULE_OP_NOT_EQUALS = "not_equals"
const RULE_OP_NOT_EQUALS_VALUE = "not_equals_value"
const RULE_OP_MATCHES_REGEXP = "matches_regexp"
const RULE_OP_WITHIN_SECONDS = "within_seconds"

const CONDITION_TYPE_MATCH = "match"
const CONDITION_TYPE_FILTER = "filter"
const CONDITION_TYPE_SKIP = "skip"

const CONDITION_VALUE_TYPE_INT = "int"
const CONDITION_VALUE_TYPE_STRING = "string"

const FILTER_FIELD_PREFIX = "field_"

const RULE_SET_TYPE_ACTIVE = "active"
const RULE_SET_TYPE_DRAFT = "draft"

const (
	RULE_SK_PREFIX       string = "rule_"
	RULE_SK_CACHE_PREFIX string = "cache_"
)

const configTablePk = "pk"
const configTableSk = "sk"

/************************************************************
* Profile Cache
*******************************************************/

type CacheMode uint8

// This constant is also defined in the front end (see source/ucp-portal/ucp-react/src/models/domain.ts)
const (
	CUSTOMER_PROFILES_MODE CacheMode = 1 << iota
	DYNAMO_MODE
)

const MAX_CACHE_BIT = CUSTOMER_PROFILES_MODE | DYNAMO_MODE

const DDB_GSI_NAME = "TravelerIndex"
const DDB_GSI_PK = "profileId"

/************************************************************
* Data Plane
*******************************************************/

const PROFILE_OBJECT_FIELD_NAME_TIMESTAMP = "timestamp"

// For depth of 1000, the unmerge operation takes 8 seconds (quip doc with benchmark data to be linked here)
const PROFILE_HISTORY_DEPTH = 1000

/************************************************************
* Control Plane
*******************************************************/
const PROFILE_OBJECT_TYPE_PREFIX = "profile_"

// name used for the profile object type
const PROFILE_OBJECT_TYPE_NAME = "_profile"

// reserved fields that cannot be used at object type name
const RESERVED_FIELD_PROFILE_ID = "profile_id"

// not sure what this field is?
const RESERVED_FIELD_CONNECT_ID = "ProfileId"
const RESERVED_FIELD_FIRST_NAME = "FirstName"

const RESERVED_FIELD_MIDDLE_NAME = "MiddleName"

const RESERVED_FIELD_LAST_NAME = "LastName"

const RESERVED_FIELD_PHONE_NUMBER = "PhoneNumber"

const RESERVED_FIELD_EMAIL_ADDRESS = "EmailAddress"
const RESERVED_FIELD_TIMESTAMP = "timestamp"

const RESERVED_FIELDS_ADDRESS = "Address"
const RESERVED_FIELDS_SHIPPING_ADDRESS = "ShippingAddress"
const RESERVED_FIELDS_MAILING_ADDRESS = "MailingAddress"
const RESERVED_FIELDS_BILLING_ADDRESS = "BillingAddress"

const RESERVED_FIELD_CONNECT = "connect_id"
const RESERVED_FIELD_ORIGINAL_CONNECT_ID = "original_connect_id"

var RESERVED_FIELDS = []string{
	//refers to the profile_id
	RESERVED_FIELD_PROFILE_ID,
	//refers to the connectID
	RESERVED_FIELD_CONNECT_ID,
	"AccountNumber",
	"AdditionalInformation",
	"PartyType",
	"BusinessName",
	RESERVED_FIELD_FIRST_NAME,
	RESERVED_FIELD_MIDDLE_NAME,
	RESERVED_FIELD_LAST_NAME,
	"BirthDate",
	"Gender",
	RESERVED_FIELD_PHONE_NUMBER,
	"MobilePhoneNumber",
	"HomePhoneNumber",
	"BusinessPhoneNumber",
	RESERVED_FIELD_EMAIL_ADDRESS,
	"BusinessEmailAddress",
	"PersonalEmailAddress",
	RESERVED_FIELDS_ADDRESS,
	RESERVED_FIELDS_SHIPPING_ADDRESS,
	RESERVED_FIELDS_MAILING_ADDRESS,
	RESERVED_FIELDS_BILLING_ADDRESS,
	"Attributes",
	RESERVED_FIELD_TIMESTAMP,
}

var ADDRESS_FIELDS = []string{
	"Address1",
	"Address2",
	"Address3",
	"Address4",
	"City",
	"State",
	"Country",
	"PostalCode",
	"Province",
}

// key supported in the Domain option update
const DOMAIN_CONFIG_OPTION_AI_IDENTITY_RESOLUTION_ENABLED = "aiIdResolutionOn"

const PRIO_FUNCTION_PREFIX = "obj_rank"

// Mapped field to check for date of last object update; leveraged to enable in-order ingestion
const LAST_UPDATED_FIELD = "last_updated"

/***************
* Event stream
*******************/
const (
	EventTypeCreated  string = "CREATED"
	EventTypeUpdated  string = "UPDATED"
	EventTypeMerged   string = "MERGED"
	EventTypeUnmerged string = "UNMERGED"
	EventTypeDeleted  string = "DELETED"
)

const (
	MergeTypeRule       string = "rule"
	MergeTypeAI         string = "ai"
	MergeTypeManual     string = "manual"
	MergeTypeUnmerge    string = "unmerge"
	OperationTypeDelete string = "delete"
)

func GetAllMergeTypesExceptUnmerge() []string {
	return []string{MergeTypeRule, MergeTypeAI, MergeTypeManual, OperationTypeDelete}
}

/*****************
* Utils
*******************/
const TIMESTAMP_FORMAT = "2006-01-02 15:04:05.000"
const INGEST_TIMESTAMP_FORMAT = "2006-01-02T15:04:05.000000Z"

const DAY_DURATION = time.Hour * 24
const WEEK_DURATION = DAY_DURATION * 7
const YEAR_DURATION = DAY_DURATION * 365
const CP_CREATE_MAPPING_MAX_RETRY = 5
const CP_CREATE_MAPPING_INTERVAL = 10
const CP_CREATE_MAPPING_MAX_INTERVAL = 180

var ErrDomainNotFound = errors.New("domain not found")

const UPDATE_SET_TIMESTAMP_LOCALTIMESTAMP = "timestamp = LOCALTIMESTAMP"
const SENSITIVE_LOG_MASK = "********"
