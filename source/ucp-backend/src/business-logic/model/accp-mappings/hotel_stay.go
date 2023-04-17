package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildHotelStayMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
		// Profile Object Unique Key
		{
			Type:    "STRING",
			Source:  "_source.id",
			Target:  "id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.first_name",
			Target:  "first_name",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_updated_by",
			Target:  "last_updated_by",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.date",
			Target:  "date",
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
			Source:  "_source.type",
			Target:  "type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.created_on",
			Target:  "created_on",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.created_by",
			Target:  "created_by",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.amount",
			Target:  "amount",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.currency_symbol",
			Target:  "currency_symbol",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.booking_id",
			Target:  "booking_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.description",
			Target:  "description",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.currency_name",
			Target:  "currency_name",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.currency_code",
			Target:  "currency_code",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.email",
			Target:  "email",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_name",
			Target:  "last_name",
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
			Source:  "_source.object_type",
			Target:  "object_type",
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
			Source:  "_source.start_date",
			Target:  "start_date",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.phone",
			Target:  "phone",
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
