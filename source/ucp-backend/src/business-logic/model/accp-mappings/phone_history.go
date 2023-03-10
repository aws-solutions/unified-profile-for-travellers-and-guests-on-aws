package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildPhoneHistoryMapping() []customerprofiles.FieldMapping {
	return []customerprofiles.FieldMapping{
		//METADATA
		{
			Type:   "STRING",
			Source: "_source.last_updated_by",
			Target: "_order.Attributes.last_updated_by",
		},
		{
			Type:   "STRING",
			Source: "_source.model_version",
			Target: "_order.Attributes.model_version",
		},
		//PROFILE DATA
		{
			Type:    "STRING",
			Source:  "_source.traveller_id",
			Target:  "_profile.profileId",
			Indexes: []string{"PROFILE"},
		},
		//ORDER DATA
		{
			Type: "STRING",
			Source: "_source.number	",
			Target:  "_order.Attributes.number",
			Indexes: []string{"UNIQUE", "ORDER"},
		},
		{
			Type:   "STRING",
			Source: "_source.last_updated",
			Target: "_order.Attributes.last_updated",
		},
		{
			Type:   "STRING",
			Source: "_source.object_type",
			Target: "_order.Attributes.object_type",
		},
		{
			Type:   "STRING",
			Source: "_source.country_code",
			Target: "_order.Attributes.country_code",
		},
		{
			Type: "STRING",
			Source: "_source.type	",
			Target: "_order.Attributes.type",
		},
	}
}
