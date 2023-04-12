package assetsSchema

import (
	glue "tah/core/glue"
	constant "tah/ucp-common/src/constant/admin"
	commonModel "tah/ucp-common/src/model/admin"
)

var SchemaPathMap = map[string]string{
	constant.BIZ_OBJECT_AIR_BOOKING:   "tah-common-glue-schemas/air_booking.glue.json",
	constant.BIZ_OBJECT_CLICKSTREAM:   "tah-common-glue-schemas/clickevent.glue.json",
	constant.BIZ_OBJECT_GUEST_PROFILE: "tah-common-glue-schemas/guest_profile.glue.json",
	constant.BIZ_OBJECT_HOTEL_BOOKING: "tah-common-glue-schemas/hotel_booking.glue.json",
	constant.BIZ_OBJECT_STAY_REVENUE:  "tah-common-glue-schemas/hotel_stay_revenue.glue.json",
	constant.BIZ_OBJECT_PAX_PROFILE:   "tah-common-glue-schemas/pax_profile.glue.json",
}

func LoadSchema(bizObject commonModel.BusinessObject) (glue.Schema, error) {
	schemaBytes, err := Asset(SchemaPathMap[bizObject.Name])
	if err != nil {
		return glue.Schema{}, err
	}
	schema, err := glue.ParseSchema(string(schemaBytes))
	if err != nil {
		return glue.Schema{}, err
	}
	return schema, nil
}
