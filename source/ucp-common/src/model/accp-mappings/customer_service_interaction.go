// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	customerprofiles "tah/upt/source/storage"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

func BuildCustomerServiceInteractionObjectMapping() customerprofiles.ObjectMapping {
	return customerprofiles.ObjectMapping{
		Name:    constants.ACCP_RECORD_CSI,
		Version: "1.0",
		Fields:  BuildCustomerServiceInteractionMapping(),
	}
}

func BuildCustomerServiceInteractionMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.model_version",
			Target:  "model_version",
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
			Source:  "_source.last_updated",
			Target:  "last_updated",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.accp_object_id",
			Indexes: []string{"UNIQUE"},
			Target:  "accp_object_id",
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
			Source:  "_source.interaction_type",
			Target:  "interaction_type",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.start_time",
			Target:  "start_time",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.end_time",
			Target:  "end_time",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.duration",
			Target:  "duration",
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
			Source:  "_source.language_code",
			Target:  "language_code",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.language_name",
			Target:  "language_name",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeText,
			Source:  "_source.summary",
			Target:  "summary",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeText,
			Source:  "_source.conversation",
			Target:  "conversation",
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
			Source:  "_source.tx_id",
			Target:  "tx_id",
			KeyOnly: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.loyalty_id",
			Target:      "loyalty_id",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.booking_id",
			Target:      "booking_id",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.related_loyalty_id1",
			Target:      "related_loyalty_id1",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.related_loyalty_id2",
			Target:      "related_loyalty_id2",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.related_loyalty_id3",
			Target:      "related_loyalty_id3",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.related_booking_id1",
			Target:      "related_booking_id1",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.related_booking_id2",
			Target:      "related_booking_id2",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.related_booking_id3",
			Target:      "related_booking_id3",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.cart_id",
			Target:      "cart_id",
			Searcheable: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.campaign_job_id",
			Target:  "campaign_job_id",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.campaign_strategy",
			Target:  "campaign_strategy",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.campaign_program",
			Target:  "campaign_program",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.campaign_product",
			Target:  "campaign_product",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.campaign_name",
			Target:  "campaign_name",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.category",
			Target:  "category",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.subject",
			Target:  "subject",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.loyalty_program_name",
			Target:  "loyalty_program_name",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.is_voice_otp",
			Target:  "is_voice_otp",
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
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.tsa_precheck_number",
			Target:      "_profile.Attributes.tsa_precheck_number",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.first_name",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.last_name",
			Target:      "_profile.LastName",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.email",
			Target:      "_profile.PersonalEmailAddress",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.email_business",
			Target:      "_profile.BusinessEmailAddress",
			Searcheable: true,
		},
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      "_source.phone_number",
			Target:      "_profile.PhoneNumber",
			Searcheable: true,
		},
	}
}
