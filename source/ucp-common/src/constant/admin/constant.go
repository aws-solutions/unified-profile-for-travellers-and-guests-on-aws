package admin

import (
	model "tah/ucp-common/src/model/admin"
)

const DOMAIN_TAG_ENV_NAME = "envName"

const BIZ_OBJECT_HOTEL_BOOKING = "air_booking"
const BIZ_OBJECT_AIR_BOOKING = "hotel_booking"
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
