// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	customerprofiles "tah/upt/source/storage"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

func BuildGuestProfileObjectMapping() customerprofiles.ObjectMapping {
	return customerprofiles.ObjectMapping{
		Name:    constants.ACCP_RECORD_GUEST_PROFILE,
		Version: "1.0",
		Fields:  BuildGuestProfileMapping(),
	}
}

func BuildGuestProfileMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
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
			Source:  "_source.last_updated",
			Target:  "last_updated",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.created_on",
			Target:  "created_on",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.created_by",
			Target:  "created_by",
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
		{
			Type:    "STRING",
			Source:  "_source.address_billing_is_primary",
			Target:  "address_billing_is_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.address_mailing_is_primary",
			Target:  "address_mailing_is_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.email_primary",
			Target:  "email_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.phone_primary",
			Target:  "phone_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.address_business_is_primary",
			Target:  "address_business_is_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.address_is_primary",
			Target:  "address_is_primary",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.tx_id",
			Target:  "tx_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_updated_partition",
			Target:  "last_updated_partition",
			KeyOnly: true,
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.extended_data",
			Target: "extended_data",
		},
		// Profile Data
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
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
			Target: "_profile.BusinessName",
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
			Source: "_source.email_business",
			Target: "_profile.BusinessEmailAddress",
		},
		{
			Type:   "STRING",
			Source: "_source.phone_home",
			Target: "_profile.HomePhoneNumber",
		},
		{
			Type:   "STRING",
			Source: "_source.phone_mobile",
			Target: "_profile.MobilePhoneNumber",
		},
		{
			Type:   "STRING",
			Source: "_source.phone_business",
			Target: "_profile.BusinessPhoneNumber",
		},
		{
			Type:   "STRING",
			Source: "_source.honorific",
			Target: "_profile.Attributes.honorific",
		},
		{
			Type:   "STRING",
			Source: "_source.first_name",
			Target: "_profile.FirstName",
		},
		{
			Type:   "STRING",
			Source: "_source.middle_name",
			Target: "_profile.MiddleName",
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
		{
			Type:   "STRING",
			Source: "_source.address_state",
			Target: "_profile.Address.State",
		},
		{
			Type:   "STRING",
			Source: "_source.address_province",
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
		// Billing Address
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
		{
			Type:   "STRING",
			Source: "_source.address_billing_state",
			Target: "_profile.BillingAddress.State",
		},
		{
			Type:   "STRING",
			Source: "_source.address_billing_province",
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
		{
			Type:   "STRING",
			Source: "_source.address_mailing_state",
			Target: "_profile.MailingAddress.State",
		},
		{
			Type:   "STRING",
			Source: "_source.address_mailing_province",
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
			Target: "_profile.ShippingAddress.Address1",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line2",
			Target: "_profile.ShippingAddress.Address2",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line3",
			Target: "_profile.ShippingAddress.Address3",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_line4",
			Target: "_profile.ShippingAddress.Address4",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_city",
			Target: "_profile.ShippingAddress.City",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_state",
			Target: "_profile.ShippingAddress.State",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_province",
			Target: "_profile.ShippingAddress.Province",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_postal_code",
			Target: "_profile.ShippingAddress.PostalCode",
		},
		{
			Type:   "STRING",
			Source: "_source.address_business_country",
			Target: "_profile.ShippingAddress.Country",
		},
	}
}
