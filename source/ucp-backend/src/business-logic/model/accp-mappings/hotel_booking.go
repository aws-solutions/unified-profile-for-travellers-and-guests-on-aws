package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildHotelBookingMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
		// Profile Object Unique Key
		{
			Type:    "STRING",
			Source:  "_source.booking_id",
			Target:  "booking_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.totalAfterTax",
			Target:  "totalAfterTax",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.email_type",
			Target:  "email_type",
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
			Source:  "_source.n_nights",
			Target:  "n_nights",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.product_id",
			Target:  "product_id",
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
			Source:  "_source.n_guests",
			Target:  "n_guests",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.room_type_description",
			Target:  "room_type_description",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.room_type_name",
			Target:  "room_type_name",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.totalBeforeTax",
			Target:  "totalBeforeTax",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.attribute_names",
			Target:  "attribute_names",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.attribute_codes",
			Target:  "attribute_codes",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.attribute_descriptions",
			Target:  "attribute_descriptions",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.room_type_code",
			Target:  "room_type_code",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.check_in_date",
			Target:  "check_in_date",
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
			Source:  "_source.model_version",
			Target:  "model_version",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.address_type",
			Target:  "address_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.phone_type",
			Target:  "phone_type",
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
		{
			Type:   "STRING",
			Source: "_source.company",
			Target: "_profile.Attributes.company",
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
		// Home Address
		{
			Type:   "STRING",
			Source: "_source.address_line1",
			Target: "_profile.Address.Address1",
		},
		{
			Type:   "STRING",
			Source: "_source.address_line2",
			Target: "_profile.Address.Address2",
		},
		{
			Type:   "STRING",
			Source: "_source.address_line3",
			Target: "_profile.Address.Address3",
		},
		{
			Type:   "STRING",
			Source: "_source.address_line4",
			Target: "_profile.Address.Address4",
		},
		{
			Type:   "STRING",
			Source: "_source.address_city",
			Target: "_profile.Address.City",
		},
		// TODO: determine when to use state vs province
		{
			Type:   "STRING",
			Source: "_source.address_state_province",
			Target: "_profile.Address.Province",
		},
		{
			Type:   "STRING",
			Source: "_source.address_postal_code",
			Target: "_profile.Address.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.address_country",
			Target: "_profile.Address.Country",
		},
		// Business Address
		{
			Type:   "STRING",
			Source: "_source.address_billing_line1",
			Target: "_profile.BillingAddress.Address1",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_line2",
			Target: "_profile.BillingAddress.Address2",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_line3",
			Target: "_profile.BillingAddress.Address3",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_line4",
			Target: "_profile.BillingAddress.Address4",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_city",
			Target: "_profile.BillingAddress.City",
		},
		// TODO: determine when to use state vs province
		{
			Type:   "STRING",
			Source: "_source.address_billing_state_province",
			Target: "_profile.BillingAddress.Province",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_postal_code",
			Target: "_profile.BillingAddress.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_country",
			Target: "_profile.BillingAddress.Country",
		},
		// Mailing Address
		{
			Type:   "STRING",
			Source: "_source.address_mailing_line1",
			Target: "_profile.MailingAddress.Address1",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_line2",
			Target: "_profile.MailingAddress.Address2",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_line3",
			Target: "_profile.MailingAddress.Address3",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_line4",
			Target: "_profile.MailingAddress.Address4",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_city",
			Target: "_profile.MailingAddress.City",
		},
		// TODO: determine when to use state vs province
		{
			Type:   "STRING",
			Source: "_source.address_mailing_state_province",
			Target: "_profile.MailingAddress.Province",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_postal_code",
			Target: "_profile.MailingAddress.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_country",
			Target: "_profile.MailingAddress.Country",
		},
		// Business
		{
			Type:   "STRING",
			Source: "_source.address_business_line1",
			Target: "_profile.Attributes.Address1",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line2",
			Target: "_profile.Attributes.Address2",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line3",
			Target: "_profile.Attributes.Address3",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line4",
			Target: "_profile.Attributes.Address4",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_city",
			Target: "_profile.Attributes.City",
		},
		// TODO: determine when to use state vs province
		{
			Type:   "STRING",
			Source: "_source.address_business_state_province",
			Target: "_profile.Attributes.Province",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_postal_code",
			Target: "_profile.Attributes.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_country",
			Target: "_profile.Attributes.Country",
		},
		{
			Type:   "STRING",
			Source: "_source.email_business",
			Target: "_profile.Attributes.email_business",
		},
		{
			Type:   "STRING",
			Source: "_source.phone_home",
			Target: "_profile.Attributes.phone_home",
		},
		{
			Type:   "STRING",
			Source: "_source.phone_mobile",
			Target: "_profile.Attributes.phone_mobile",
		},
		{
			Type:   "STRING",
			Source: "_source.phone_business",
			Target: "_profile.Attributes.phone_business",
		},
	}
}
