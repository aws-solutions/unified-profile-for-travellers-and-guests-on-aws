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
			Source:  "_source.id",
			Target:  "_order.Attributes.confirmationNumber",
			Indexes: []string{"UNIQUE", "ORDER"},
		},
		{
			Type:   "STRING",
			Source: "_source.creationChannelId",
			Target: "_order.Name",
		},

		{
			Type:   "STRING",
			Source: "_source.nGuests",
			Target: "_order.Attributes.nGuests",
		},
		{
			Type:        "STRING",
			Source:      "_source.lastName",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.hotel_name",
			Target: "_order.Attributes.hotelName",
		},
		{
			Type:   "STRING",
			Source: "_source.city",
			Target: "_profile.Address.City",
		},
		{
			Type:        "STRING",
			Source:      "_source.firstName",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.total_price",
			Target: "_order.TotalPrice",
		},
		{
			Type:   "STRING",
			Source: "_source.startDate",
			Target: "_order.Attributes.startDate",
		},
		{
			Type:   "STRING",
			Source: "_source.country",
			Target: "_profile.Address.Country",
		},
		{
			Type:   "STRING",
			Source: "_source.hotel_code",
			Target: "_order.Attributes.hotelCode",
		},
		{
			Type:   "STRING",
			Source: "_source.nNight",
			Target: "_order.Attributes.nNights",
		},
		{
			Type:   "STRING",
			Source: "_source.products",
			Target: "_order.Attributes.products",
		},
		{
			Type:   "STRING",
			Source: "_source.status",
			Target: "_order.Status",
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.loyaltyId",
			Target:      "_profile.AccountNumber",
			Searcheable: true,
			//TODO: this index should go on a dedicated customer ID field
			Indexes: []string{"PROFILE"},
		},
		{
			Type:        "STRING",
			Source:      "_source.phone",
			Target:      "_profile.PhoneNumber",
			Searcheable: true,
		},
	}
}
