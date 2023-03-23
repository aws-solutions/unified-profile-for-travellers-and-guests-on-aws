package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildAirLoyaltyMapping() customerprofiles.FieldMappings {
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
			Type:    "STRING",
			Source:  "_source.id",
			Target:  "_order.Attributes.loyalty_id",
			Indexes: []string{"UNIQUE", "ORDER"},
		},
		{
			Type:   "STRING",
			Source: "_source.program_name",
			Target: "_profile.Attributes.program_name",
		},
		{
			Type:   "STRING",
			Source: "_source.miles",
			Target: "_profile.Attributes.miles",
		},
		{
			Type:   "STRING",
			Source: "_source.miles_to_next_level",
			Target: "_profile.Attributes.miles_to_next_level",
		},
		{
			Type:   "STRING",
			Source: "_source.level",
			Target: "_profile.Attributes.level",
		},
		{
			Type:   "STRING",
			Source: "_source.joined",
			Target: "_profile.Attributes.joined",
		},
	}
}
