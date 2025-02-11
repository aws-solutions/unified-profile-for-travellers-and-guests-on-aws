// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	customerprofiles "tah/upt/source/storage"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

func BuildAirBookingObjectMapping() customerprofiles.ObjectMapping {
	return customerprofiles.ObjectMapping{
		Name:    constants.ACCP_RECORD_AIR_BOOKING,
		Version: "1.0",
		Fields:  BuildAirBookingMapping(),
	}
}

func BuildAirBookingMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
		// Profile Object Unique Key
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.segment_id",
			Target:  "segment_id",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.from",
			Target:  "from",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.arrival_date",
			Target:  "arrival_date",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.to",
			Target:  "to",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.booking_id",
			Target:  "booking_id",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.departure_time",
			Target:  "departure_time",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.status",
			Target:  "status",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.object_type",
			Target:  "object_type",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.arrival_time",
			Target:  "arrival_time",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.model_version",
			Target:  "model_version",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.channel",
			Target:  "channel",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.address_type",
			Target:  "address_type",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.departure_date",
			Target:  "departure_date",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.flight_number",
			Target:  "flight_number",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.phone_type",
			Target:  "phone_type",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.email_type",
			Target:  "email_type",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.last_updated_by",
			Target:  "last_updated_by",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.price",
			Target:  "price",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.last_updated",
			Target:  "last_updated",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.creation_channel_id",
			Target:  "creation_channel_id",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.last_update_channel_id",
			Target:  "last_update_channel_id",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.address_business_is_primary",
			Target:  "address_business_is_primary",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.address_billing_is_primary",
			Target:  "address_billing_is_primary",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.address_is_primary",
			Target:  "address_is_primary",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.tx_id",
			Target:  "tx_id",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.email_primary",
			Target:  "email_primary",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.phone_primary",
			Target:  "phone_primary",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.last_updated_partition",
			Target:  "last_updated_partition",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.address_mailing_is_primary",
			Target:  "address_mailing_is_primary",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.traveller_price",
			Target:  "traveller_price",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.booker_id",
			Target:  "booker_id",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.day_of_travel_email",
			Target:  "day_of_travel_email",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.day_of_travel_phone",
			Target:  "day_of_travel_phone",
			KeyOnly: true,
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.extended_data",
			Target: "extended_data",
		},
		// Profile Data
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.traveller_id",
			Target:      "_profile.Attributes.profile_id",
			Searcheable: true,
			Indexes:     []string{"PROFILE"},
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.company",
			Target: "_profile.BusinessName",
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.phone",
			Target:      "_profile.PhoneNumber",
			Searcheable: true,
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.honorific",
			Target: "_profile.Attributes.honorific",
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.first_name",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.middle_name",
			Target:      "_profile.MiddleName",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.last_name",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.gender",
			Target: "_profile.Gender",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.pronoun",
			Target: "_profile.Attributes.pronoun",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.date_of_birth",
			Target: "_profile.BirthDate",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.job_title",
			Target: "_profile.Attributes.job_title",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.nationality_code",
			Target: "_profile.Attributes.nationality_code",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.nationality_name",
			Target: "_profile.Attributes.nationality_name",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.language_code",
			Target: "_profile.Attributes.language_code",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.language_name",
			Target: "_profile.Attributes.language_name",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.pss_id",
			Target: "_profile.Attributes.pss_id",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.gds_id",
			Target: "_profile.Attributes.gds_id",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.payment_type",
			Target: "_profile.Attributes.payment_type",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.cc_token",
			Target: "_profile.Attributes.cc_token",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.cc_type",
			Target: "_profile.Attributes.cc_type",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.cc_exp",
			Target: "_profile.Attributes.cc_exp",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.cc_cvv",
			Target: "_profile.Attributes.cc_cvv",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.cc_name",
			Target: "_profile.Attributes.cc_name",
		},
		// Home Address
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_line1",
			Target: "_profile.Address.Address1",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_line2",
			Target: "_profile.Address.Address2",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_line3",
			Target: "_profile.Address.Address3",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_line4",
			Target: "_profile.Address.Address4",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_city",
			Target: "_profile.Address.City",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_state",
			Target: "_profile.Address.State",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_province",
			Target: "_profile.Address.Province",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_postal_code",
			Target: "_profile.Address.PostalCode",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_country",
			Target: "_profile.Address.Country",
		},
		// Business Address
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_billing_line1",
			Target: "_profile.BillingAddress.Address1",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_billing_line2",
			Target: "_profile.BillingAddress.Address2",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_billing_line3",
			Target: "_profile.BillingAddress.Address3",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_billing_line4",
			Target: "_profile.BillingAddress.Address4",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_billing_city",
			Target: "_profile.BillingAddress.City",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_billing_state",
			Target: "_profile.BillingAddress.State",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_billing_province",
			Target: "_profile.BillingAddress.Province",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_billing_postal_code",
			Target: "_profile.BillingAddress.PostalCode",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_billing_country",
			Target: "_profile.BillingAddress.Country",
		},
		// Mailing Address
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_mailing_line1",
			Target: "_profile.MailingAddress.Address1",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_mailing_line2",
			Target: "_profile.MailingAddress.Address2",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_mailing_line3",
			Target: "_profile.MailingAddress.Address3",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_mailing_line4",
			Target: "_profile.MailingAddress.Address4",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_mailing_city",
			Target: "_profile.MailingAddress.City",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_mailing_state",
			Target: "_profile.MailingAddress.State",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_mailing_province",
			Target: "_profile.MailingAddress.Province",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_mailing_postal_code",
			Target: "_profile.MailingAddress.PostalCode",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_mailing_country",
			Target: "_profile.MailingAddress.Country",
		},
		// Business
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_business_line1",
			Target: "_profile.ShippingAddress.Address1",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_business_line2",
			Target: "_profile.ShippingAddress.Address2",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_business_line3",
			Target: "_profile.ShippingAddress.Address3",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_business_line4",
			Target: "_profile.ShippingAddress.Address4",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_business_city",
			Target: "_profile.ShippingAddress.City",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_business_state",
			Target: "_profile.ShippingAddress.State",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_business_province",
			Target: "_profile.ShippingAddress.Province",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_business_postal_code",
			Target: "_profile.ShippingAddress.PostalCode",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.address_business_country",
			Target: "_profile.ShippingAddress.Country",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.email_business",
			Target: "_profile.BusinessEmailAddress",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.phone_home",
			Target: "_profile.HomePhoneNumber",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.phone_mobile",
			Target: "_profile.MobilePhoneNumber",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.phone_business",
			Target: "_profile.BusinessPhoneNumber",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.last_booking_id",
			Target: "_profile.Attributes.last_booking_id",
		},
	}
}
