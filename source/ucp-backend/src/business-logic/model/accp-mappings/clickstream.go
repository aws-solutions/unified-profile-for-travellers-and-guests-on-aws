package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildClickstreamMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
		// Profile Object Unique Key
		{
			Type:    "STRING",
			Source:  "_source.session_id",
			Target:  "session_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.num_pax_children",
			Target:  "num_pax_children",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.flight_type",
			Target:  "flight_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_updated",
			Target:  "last_updated",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.origin_date",
			Target:  "origin_date",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.return_date_time",
			Target:  "return_date_time",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.fare_type",
			Target:  "fare_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.flight_market",
			Target:  "flight_market",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.num_pax_adults",
			Target:  "num_pax_adults",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.products",
			Target:  "products",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.fare_class",
			Target:  "fare_class",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.total_passengers",
			Target:  "total_passengers",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.origin_date_time",
			Target:  "origin_date_time",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.event_timestamp",
			Target:  "event_timestamp",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.aws_account_id",
			Target:  "aws_account_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.return_flight_route",
			Target:  "return_flight_route",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.flight_segments_departure_date_time",
			Target:  "flight_segments_departure_date_time",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.event_type",
			Target:  "event_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.return_date",
			Target:  "return_date",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.arrival_timestamp",
			Target:  "arrival_timestamp",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.model_version",
			Target:  "model_version",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.pax_type",
			Target:  "pax_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.object_type",
			Target:  "object_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.num_pax_inf",
			Target:  "num_pax_inf",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.flight_numbers",
			Target:  "flight_numbers",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.event_version",
			Target:  "event_version",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.user_agent",
			Target:  "user_agent",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.rate_plan",
			Target:  "rate_plan",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.checkin_date",
			Target:  "checkin_date",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.checkout_date",
			Target:  "checkout_date",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.hotel_code_list",
			Target:  "hotel_code_list",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.hotel_code",
			Target:  "hotel_code",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.room_type",
			Target:  "room_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.num_nights",
			Target:  "num_nights",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.destination",
			Target:  "destination",
			KeyOnly: true,
		},
		// Profile Data
		{
			Type:        "STRING",
			Source:      "_source.traveller_id",
			Target:      "_profile.Attributes.profile_id",
			Searcheable: true,
			Indexes:     []string{"PROFILE"},
		},
	}
}
