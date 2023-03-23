package accpmappings

import customerprofiles "tah/core/customerprofiles"

func BuildAirBookingMapping() customerprofiles.FieldMappings {
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
			Source: "_source.pss_id",
			Target: "_profile.Attributes.pss_id",
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

		//Order Data
		{
			Type:        "STRING",
			Source:      "_source.booking_id",
			Target:      "_order.Attributes.booking_id",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.segment_id",
			Target:      "_order.Attributes.segment_id",
			Indexes:     []string{"UNIQUE", "ORDER"},
			Searcheable: true,
		},
		{
			Type:   "STRING",
			Source: "_source.from",
			Target: "_order.Attributes.from",
		},
		{
			Type:   "STRING",
			Source: "_source.to",
			Target: "_order.Attributes.to",
		},
		{
			Type:   "STRING",
			Source: "_source.flight_number",
			Target: "_order.Attributes.flight_number",
		},
		{
			Type:   "STRING",
			Source: "_source.departure_date",
			Target: "_order.Attributes.departure_date",
		},
		{
			Type:   "STRING",
			Source: "_source.departure_time",
			Target: "_order.Attributes.departure_time",
		},
		{
			Type:   "STRING",
			Source: "_source.arrival_date",
			Target: "_order.Attributes.arrival_date",
		},
		{
			Type:   "STRING",
			Source: "_source.arrival_time",
			Target: "_order.Attributes.arrival_time",
		},
		{
			Type:   "STRING",
			Source: "_source.channel",
			Target: "_order.Attributes.channel",
		},
		{
			Type:   "STRING",
			Source: "_source.status",
			Target: "_order.Attributes.status",
		},
		{
			Type:   "STRING",
			Source: "_source.price",
			Target: "_order.Attributes.price",
		},
	}
}
