package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildPhoneHistoryMapping() customerprofiles.FieldMappings {
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
			Source:  "_source.number",
			Target:  "_order.Attributes.number",
			Indexes: []string{"UNIQUE", "ORDER"},
		},
		{
			Type:   "STRING",
			Source: "_source.country_code",
			Target: "_order.Attributes.country_code",
		},
		{
			Type:   "STRING",
			Source: "_source.type",
			Target: "_order.Attributes.type",
		},
	}
}
