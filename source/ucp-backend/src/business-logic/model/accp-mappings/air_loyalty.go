package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildAirLoyaltyMapping() customerprofiles.FieldMappings {
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
			Source:  "_source.last_updated_by",
			Target:  "last_updated_by",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.level",
			Target:  "level",
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
			Source:  "_source.program_name",
			Target:  "program_name",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.miles",
			Target:  "miles",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.miles_to_next_level",
			Target:  "miles_to_next_level",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.joined",
			Target:  "joined",
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
			Source:  "_source.model_version",
			Target:  "model_version",
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
