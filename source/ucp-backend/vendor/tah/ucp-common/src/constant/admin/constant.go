package admin

import (
	model "tah/ucp-common/src/model/admin"
)

const DOMAIN_TAG_ENV_NAME = "envName"

const BIZ_OBJECT_AIR_BOOKING = "air_booking"
const BIZ_OBJECT_HOTEL_BOOKING = "hotel_booking"
const BIZ_OBJECT_GUEST_PROFILE = "guest_profile"
const BIZ_OBJECT_PAX_PROFILE = "pax_profile"
const BIZ_OBJECT_CLICKSTREAM = "clickstream"
const BIZ_OBJECT_STAY_REVENUE = "hotel_stay_revenue"

var BUSINESS_OBJECTS = []model.BusinessObject{
	model.BusinessObject{Name: BIZ_OBJECT_HOTEL_BOOKING},
	model.BusinessObject{Name: BIZ_OBJECT_AIR_BOOKING},
	model.BusinessObject{Name: BIZ_OBJECT_PAX_PROFILE},
	model.BusinessObject{Name: BIZ_OBJECT_GUEST_PROFILE},
	model.BusinessObject{Name: BIZ_OBJECT_CLICKSTREAM},
	model.BusinessObject{Name: BIZ_OBJECT_STAY_REVENUE},
}

var ACCP_RECORD_AIR_BOOKING = "air_booking"
var ACCP_RECORD_EMAIL_HISTORY = "email_history"
var ACCP_RECORD_PHONE_HISTORY = "phone_history"
var ACCP_RECORD_AIR_LOYALTY = "air_loyalty"
var ACCP_RECORD_CLICKSTREAM = "clickstream"
var ACCP_RECORD_GUEST_PROFILE = "guest_profile"
var ACCP_RECORD_HOTEL_LOYALTY = "hotel_loyalty"
var ACCP_RECORD_HOTEL_BOOKING = "hotel_booking"
var ACCP_RECORD_PAX_PROFILE = "pax_profile"
var ACCP_RECORD_HOTEL_STAY_MAPPING = "hotel_stay_revenue_items"

var ACCP_RECORDS = []model.AccpRecord{
	model.AccpRecord{Name: ACCP_RECORD_AIR_BOOKING},
	model.AccpRecord{Name: ACCP_RECORD_EMAIL_HISTORY},
	model.AccpRecord{Name: ACCP_RECORD_PHONE_HISTORY},
	model.AccpRecord{Name: ACCP_RECORD_AIR_LOYALTY},
	model.AccpRecord{Name: ACCP_RECORD_CLICKSTREAM},
	model.AccpRecord{Name: ACCP_RECORD_GUEST_PROFILE},
	model.AccpRecord{Name: ACCP_RECORD_HOTEL_LOYALTY},
	model.AccpRecord{Name: ACCP_RECORD_HOTEL_BOOKING},
	model.AccpRecord{Name: ACCP_RECORD_PAX_PROFILE},
	model.AccpRecord{Name: ACCP_RECORD_HOTEL_STAY_MAPPING},
}
