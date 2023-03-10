package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildClickstreamMapping() []customerprofiles.FieldMapping {
	return []customerprofiles.FieldMapping{
		// Metadata
		{
			Type:   "STRING",
			Source: "_source.model_version",
			Target: "_order.Attributes.model_version",
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
			Type:    "STRING",
			Source:  "_source.timestamp",
			Target:  "_order.Attributes.timestamp",
			Indexes: []string{"UNIQUE", "ORDER"},
		},
		{
			Type:   "STRING",
			Source: "_source.location",
			Target: "_order.Attributes.location",
		},
		{
			Type:   "STRING",
			Source: "_source.nGuests",
			Target: "_order.Attributes.nGuests",
		},
		{
			Type:    "STRING",
			Source:  "_source.profile",
			Target:  "_order.Attributes.loyaltyId",
			Indexes: []string{"PROFILE"},
		},
		{
			Type:   "STRING",
			Source: "_source.roomType",
			Target: "_order.Attributes.roomType",
		},
		{
			Type:   "STRING",
			Source: "_source.session_id",
			Target: "_order.Attributes.sessionId",
		},
		{
			Type:        "STRING",
			Source:      "_source.startDate",
			Target:      "_order.Attributes.startDate",
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.tags",
			Target: "_order.Attributes.tags",
		},
		{
			Type:   "STRING",
			Source: "_source.action",
			Target: "_order.Attributes.action",
		},
		{
			Type:        "STRING",
			Source:      "_source.hotel",
			Target:      "_order.Attributes.hotelCode",
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.nRooms",
			Target: "_order.Attributes.nRooms",
		},
		{
			Type:   "STRING",
			Source: "_source.nNights",
			Target: "_order.Attributes.nNights",
		},
	}
}
