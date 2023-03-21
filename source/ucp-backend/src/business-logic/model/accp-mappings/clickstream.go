package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildClickstreamMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
		// Metadata
		{
			Type:   "STRING",
			Source: "_source.model_version",
			Target: "_order.Attributes.model_version",
		},
		{
			Type:   "STRING",
			Source: "_source.object_type",
			Target: "_order.Attributes.object_type",
		},
		{
			Type:   "STRING",
			Source: "_source.last_updated",
			Target: "_order.Attributes.last_updated",
		},

		// Profile Data
		{
			Type:        "STRING",
			Source:      "_source.traveller_id",
			Target:      "_profile.profileId",
			Searcheable: true,
			Indexes:     []string{"PROFILE"},
		},

		//Order Data
		{
			Type:   "STRING",
			Source: "_source.session_id",
			Target: "_order.Attributes.session_id",
		},
		{
			Type:    "STRING",
			Source:  "_source.event_timestamp",
			Target:  "_order.Attributes.event_timestamp",
			Indexes: []string{"UNIQUE", "ORDER"},
		},
		{
			Type:   "STRING",
			Source: "_source.event_type",
			Target: "_order.Attributes.event_type",
		},
		{
			Type:   "STRING",
			Source: "_source.event_version",
			Target: "_order.Attributes.event_version",
		},
		{
			Type:   "STRING",
			Source: "_source.arrival_timestamp",
			Target: "_order.Attributes.arrival_timestamp",
		},
		{
			Type:   "STRING",
			Source: "_source.user_agent",
			Target: "_order.Attributes.user_agent",
		},
		{
			Type:   "STRING",
			Source: "_source.products",
			Target: "_order.Attributes.products",
		},
		{
			Type:   "STRING",
			Source: "_source.fare_class",
			Target: "_order.Attributes.fare_class",
		},
		{
			Type:   "STRING",
			Source: "_source.fare_type",
			Target: "_order.Attributes.fare_type",
		},
		{
			Type:   "STRING",
			Source: "_source.flight_segments_departure_date_time",
			Target: "_order.Attributes.flight_segments_departure_date_time",
		},
		{
			Type:   "STRING",
			Source: "_source.flight_numbers",
			Target: "_order.Attributes.flight_numbers",
		},
		{
			Type:   "STRING",
			Source: "_source.flight_market",
			Target: "_order.Attributes.flight_market",
		},
		{
			Type:   "STRING",
			Source: "_source.flight_type",
			Target: "_order.Attributes.flight_type",
		},
		{
			Type:   "STRING",
			Source: "_source.origin_date",
			Target: "_order.Attributes.origin_date",
		},
		{
			Type:   "STRING",
			Source: "_source.origin_date_time",
			Target: "_order.Attributes.origin_date_time",
		},
		{
			Type:   "STRING",
			Source: "_source.return_date_time",
			Target: "_order.Attributes.return_date_time",
		},
		{
			Type:   "STRING",
			Source: "_source.return_date",
			Target: "_order.Attributes.return_date",
		},
		{
			Type:   "STRING",
			Source: "_source.return_flight_route",
			Target: "_order.Attributes.return_flight_route",
		},
		{
			Type:   "STRING",
			Source: "_source.num_pax_adults",
			Target: "_order.Attributes.num_pax_adults",
		},
		{
			Type:   "STRING",
			Source: "_source.num_pax_inf",
			Target: "_order.Attributes.num_pax_inf",
		},
		{
			Type:   "STRING",
			Source: "_source.num_pax_children",
			Target: "_order.Attributes.num_pax_children",
		},
		{
			Type:   "STRING",
			Source: "_source.pax_type",
			Target: "_order.Attributes.pax_type",
		},
		{
			Type:   "STRING",
			Source: "_source.total_passengers",
			Target: "_order.Attributes.total_passengers",
		},
		{
			Type:   "STRING",
			Source: "_source.aws_account_id",
			Target: "_order.Attributes.aws_account_id",
		},
	}
}
