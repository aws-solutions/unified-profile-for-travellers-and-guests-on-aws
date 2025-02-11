// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	customerprofiles "tah/upt/source/storage"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

var ACCP_OBJECT_MAPPINGS = map[string]func() customerprofiles.ObjectMapping{
	constants.ACCP_RECORD_AIR_BOOKING:          BuildAirBookingObjectMapping,
	constants.ACCP_RECORD_EMAIL_HISTORY:        BuildEmailHistoryObjectMapping,
	constants.ACCP_RECORD_PHONE_HISTORY:        BuildPhoneHistoryObjectMapping,
	constants.ACCP_RECORD_AIR_LOYALTY:          BuildAirLoyaltyObjectMapping,
	constants.ACCP_RECORD_CLICKSTREAM:          BuildClickstreamObjectMapping,
	constants.ACCP_RECORD_GUEST_PROFILE:        BuildGuestProfileObjectMapping,
	constants.ACCP_RECORD_HOTEL_LOYALTY:        BuildHotelLoyaltyObjectMapping,
	constants.ACCP_RECORD_HOTEL_BOOKING:        BuildHotelBookingObjectMapping,
	constants.ACCP_RECORD_PAX_PROFILE:          BuildPassengerProfileObjectMapping,
	constants.ACCP_RECORD_HOTEL_STAY_MAPPING:   BuildHotelStayObjectMapping,
	constants.ACCP_RECORD_CSI:                  BuildCustomerServiceInteractionObjectMapping,
	constants.ACCP_RECORD_LOYALTY_TX:           BuildLoyaltyTransactionObjectMapping,
	constants.ACCP_RECORD_ANCILLARY:            BuildAncillaryServiceObjectMapping,
	constants.ACCP_RECORD_ALTERNATE_PROFILE_ID: BuildAlternateProfileIdObjectMapping,
}
