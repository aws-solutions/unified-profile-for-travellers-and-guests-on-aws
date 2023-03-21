package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildHotelStayMapping() customerprofiles.FieldMappings {
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
		{
			Type:   "STRING",
			Source: "_source.last_updated_by",
			Target: "_order.Attributes.last_updated_by",
		},
		{
			Type:   "STRING",
			Source: "_source.created_on",
			Target: "_order.Attributes.created_on",
		},
		{
			Type:   "STRING",
			Source: "_source.created_by",
			Target: "_order.Attributes.created_by",
		},

		// Profile Data
		{
			Type:        "STRING",
			Source:      "_source.traveller_id",
			Target:      "_profile.Attributes.profile_id",
			Searcheable: true,
			Indexes:     []string{"PROFILE"},
		},

		//Order Data
		{
			Type:        "STRING",
			Source:      "_source.id",
			Target:      "_order.Attributes.id",
			Searcheable: true,
			Indexes:     []string{"UNIQUE", "ORDER"},
		},
		{
			Type:        "STRING",
			Source:      "_source.booking_id",
			Searcheable: true,
			Target:      "_order.Attributes.booking_id",
		},
		{
			Type:   "STRING",
			Source: "_source.currency_code",
			Target: "_order.Attributes.currency_code",
		},
		{
			Type:   "STRING",
			Source: "_source.currency_name",
			Target: "_order.Attributes.currency_name",
		},
		{
			Type:   "STRING",
			Source: "_source.currency_symbol",
			Target: "_order.Attributes.currency_symbol",
		},
		{
			Type:   "STRING",
			Source: "_source.first_name",
			Target: "_order.Attributes.first_name",
		},
		{
			Type:   "STRING",
			Source: "_source.last_name",
			Target: "_order.Attributes.last_name",
		},
		{
			Type:   "STRING",
			Source: "_source.email",
			Target: "_order.Attributes.email",
		},
		{
			Type:   "STRING",
			Source: "_source.phone",
			Target: "_order.Attributes.phone",
		},
		{
			Type:   "STRING",
			Source: "_source.start_date",
			Target: "_order.Attributes.start_date",
		},
		{
			Type:   "STRING",
			Source: "_source.hotel_code",
			Target: "_order.Attributes.hotel_code",
		},
		{
			Type:   "STRING",
			Source: "_source.type",
			Target: "_order.Attributes.type",
		},
		{
			Type:   "STRING",
			Source: "_source.description",
			Target: "_order.Attributes.description",
		},
		{
			Type:   "STRING",
			Source: "_source.amount",
			Target: "_order.Attributes.amount",
		},
		{
			Type:   "STRING",
			Source: "_source.date",
			Target: "_order.Attributes.date",
		},
	}
}
