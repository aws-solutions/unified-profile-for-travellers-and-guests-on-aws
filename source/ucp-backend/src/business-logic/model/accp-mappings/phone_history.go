package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildPhoneHistoryMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
		// Profile Object Unique Key
		{
			Type:    "STRING",
			Source:  "_source.number",
			Target:  "number",
			Indexes: []string{"UNIQUE"},
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
