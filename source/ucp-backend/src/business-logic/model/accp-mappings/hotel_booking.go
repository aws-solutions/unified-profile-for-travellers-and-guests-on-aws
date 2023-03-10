package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildHotelBookingMapping() []customerprofiles.FieldMapping {
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

		// Order Data
		{
			Type:        "STRING",
			Source:      "_source.booking_id",
			Target:      "_order.Attributes.booking_id",
			Searcheable: true,
			Indexes:     []string{"UNIQUE", "ORDER"},
		},
		{
			Type:   "STRING",
			Source: "_source.last_updated",
			Target: "_order.UpdatedDate",
		},
		{
			Type:   "STRING",
			Source: "_source.hotel_code",
			Target: "_order.Attributes.hotel_code",
		},
		{
			Type:   "STRING",
			Source: "_source.n_nights",
			Target: "_order.Attributes.n_nights",
		},
		{
			Type:   "STRING",
			Source: "_source.n_guests",
			Target: "_order.Attributes.n_guests",
		},
		{
			Type:   "STRING",
			Source: "_source.product_id",
			Target: "_order.Attributes.product_id",
		},
		{
			Type:   "STRING",
			Source: "_source.check_in_date",
			Target: "_order.Attributes.check_in_date",
		},
		{
			Type:   "STRING",
			Source: "_source.room_type_code",
			Target: "_order.Attributes.room_type_code",
		},
		{
			Type:   "STRING",
			Source: "_source.room_type_name",
			Target: "_order.Attributes.room_type_name",
		},
		{
			Type:   "STRING",
			Source: "_source.room_type_description",
			Target: "_order.Attributes.room_type_description",
		},
		{
			Type:   "STRING",
			Source: "_source.attribute_codes",
			Target: "_order.Attributes.attribute_codes",
		},
		{
			Type:   "STRING",
			Source: "_source.attribute_names",
			Target: "_order.Attributes.attribute_names",
		},
		{
			Type:   "STRING",
			Source: "_source.attribute_descriptions",
			Target: "_order.Attributes.attribute_descriptions",
		},

		{
			Type:        "STRING",
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.phone",
			Target:      "_profile.PhoneNumber",
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.honorific",
			Target: "_profile.Attributes.honorific",
		},
		{
			Type:        "STRING",
			Source:      "_source.first_name",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.middle_name",
			Target:      "_profile.MiddleName",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.last_name",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.gender",
			Target: "_profile.Gender",
		},
		{
			Type:   "STRING",
			Source: "_source.pronoun",
			Target: "_profile.Attributes.pronoun",
		},
		{
			Type:   "STRING",
			Source: "_source.date_of_birth",
			Target: "_profile.BirthDate",
		},
		{
			Type:   "STRING",
			Source: "_source.job_title",
			Target: "_profile.Attributes.job_title",
		},
		{
			Type:   "STRING",
			Source: "_source.nationality_code",
			Target: "_profile.Attributes.nationality_code",
		},
		{
			Type:   "STRING",
			Source: "_source.nationality_name",
			Target: "_profile.Attributes.nationality_name",
		},
		{
			Type:   "STRING",
			Source: "_source.language_code",
			Target: "_profile.Attributes.language_code",
		},
		{
			Type:   "STRING",
			Source: "_source.language_name",
			Target: "_profile.Attributes.language_name",
		},
		{
			Type:   "STRING",
			Source: "_source.pms_id",
			Target: "_profile.Attributes.pms_id",
		},
		{
			Type:   "STRING",
			Source: "_source.crs_id",
			Target: "_profile.Attributes.crs_id",
		},
		{
			Type:   "STRING",
			Source: "_source.gds_id",
			Target: "_profile.Attributes.gds_id",
		},
		{
			Type:   "STRING",
			Source: "_source.payment_type",
			Target: "_profile.Attributes.payment_type",
		},
		{
			Type:   "STRING",
			Source: "_source.cc_token",
			Target: "_profile.Attributes.cc_token",
		},
		{
			Type:   "STRING",
			Source: "_source.cc_type",
			Target: "_profile.Attributes.cc_type",
		},
		{
			Type:   "STRING",
			Source: "_source.cc_exp",
			Target: "_profile.Attributes.cc_exp",
		},
		{
			Type:   "STRING",
			Source: "_source.cc_cvv",
			Target: "_profile.Attributes.cc_cvv",
		},
		{
			Type:   "STRING",
			Source: "_source.cc_name",
			Target: "_profile.Attributes.cc_name",
		},

		// {
		// 	Type:   "STRING",
		// 	Source: "_source.address_billing_line1",
		// 	Target: "_profile.BillingAddress.Address1",
		// },
		// {
		// 	Type:   "STRING",
		// 	Source: "_source.address_billing_line2",
		// 	Target: "_profile.BillingAddress.Address2",
		// },
		// {
		// 	Type:   "STRING",
		// 	Source: "_source.address_billing_line3",
		// 	Target: "_profile.BillingAddress.Address3",
		// },
		// {
		// 	Type:   "STRING",
		// 	Source: "_source.address_billing_line4",
		// 	Target: "_profile.BillingAddress.Address4",
		// },
		// {
		// 	Type:   "STRING",
		// 	Source: "_source.address_billing_city",
		// 	Target: "_profile.BillingAddress.City",
		// },
		// {
		// 	Type:   "STRING",
		// 	Source: "_source.address_billing_state_province",
		// TODO: determine when to use state vs province
		// 	Target: "_profile.BillingAddress.Province",
		// },
		// {
		// 	Type:   "STRING",
		// 	Source: "_source.address_billing_postal_code",
		// 	Target: "_profile.BillingAddress.PostalCode",
		// },
		// {
		// 	Type:   "STRING",
		// 	Source: "_source.address_billing_country",
		// 	Target: "_profile.BillingAddress.Country",
		// },
	}
}
