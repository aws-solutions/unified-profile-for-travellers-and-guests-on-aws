package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildAirLoyaltyMapping() []customerprofiles.FieldMapping {
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
			Type:   "STRING",
			Source: "_source.id",
			Target: "_order.Attributes.loyaltyId",
			//TODO: this index should go on a dedicated customer ID field
			Indexes: []string{"UNIQUE", "ORDER"},
		},
		{
			Type:   "STRING",
			Source: "_source.city",
			Target: "_profile.Address.City",
		},
		{
			Type:   "STRING",
			Source: "_source.state",
			Target: "_profile.Address.State",
		},
		{
			Type:        "STRING",
			Source:      "_source.firstName",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.ageGroup",
			Target: "_profile.Attributes.ageGroup",
		},
		{
			Type:   "STRING",
			Source: "_source.honorific",
			Target: "_profile.Attributes.honorific",
		},
		{
			Type:   "STRING",
			Source: "_source.parentCompany",
			Target: "_profile.BusinessName",
		},
		{
			Type:   "STRING",
			Source: "_source.points",
			Target: "_order.Attributes.points",
		},
		{
			Type:   "STRING",
			Source: "_source.gender",
			Target: "_profile.Gender",
		},
		{
			Type:        "STRING",
			Source:      "_source.lastName",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.program",
			Target: "_order.Attributes.program",
		},
		{
			Type:   "STRING",
			Source: "_source.jobTitle",
			Target: "_profile.Attributes.jobTitle",
		},
		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.EmailAddress",
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.postalCode",
			Target: "_profile.Address.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.addressLine1",
			Target: "_profile.Address.Address1",
		},
		{
			Type:   "STRING",
			Source: "_source.birthdate",
			Target: "_profile.BirthDate",
		},
		{
			Type:   "STRING",
			Source: "_source.status",
			Target: "_order.Attributes.status",
		},
		{
			Type:   "STRING",
			Source: "_source.joined",
			Target: "_order.Attributes.joined",
		},
		{
			Type:        "STRING",
			Source:      "_source.phone",
			Target:      "_profile.PhoneNumber",
			Searcheable: true,
		},
	}
}
