// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	customerprofiles "tah/upt/source/storage"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

func BuildAirLoyaltyObjectMapping() customerprofiles.ObjectMapping {
	return customerprofiles.ObjectMapping{
		Name:    constants.ACCP_RECORD_AIR_LOYALTY,
		Version: "1.0",
		Fields:  BuildAirLoyaltyMapping(),
	}
}

func BuildAirLoyaltyMapping() customerprofiles.FieldMappings {
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
			Source:  "_source.last_updated_by",
			Target:  "last_updated_by",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.level",
			Target:  "level",
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
			Source:  "_source.program_name",
			Target:  "program_name",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.miles",
			Target:  "miles",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.miles_to_next_level",
			Target:  "miles_to_next_level",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.joined",
			Target:  "joined",
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
			Source:  "_source.model_version",
			Target:  "model_version",
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
			Source:  "_source.last_updated_partition",
			Target:  "last_updated_partition",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.enrollment_source",
			Target:  "enrollment_source",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.currency",
			Target:  "currency",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.amount",
			Target:  "amount",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.account_status",
			Target:  "account_status",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.reason_for_close",
			Target:  "reason_for_close",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.language_preference",
			Target:  "language_preference",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.display_preference",
			Target:  "display_preference",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.meal_preference",
			Target:  "meal_preference",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.seat_preference",
			Target:  "seat_preference",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.home_airport",
			Target:  "home_airport",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.date_time_format_preference",
			Target:  "date_time_format_preference",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.cabin_preference",
			Target:  "cabin_preference",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.fare_type_preference",
			Target:  "fare_type_preference",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.expert_mode",
			Target:  "expert_mode",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.privacy_indicator",
			Target:  "privacy_indicator",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.car_preference_vendor",
			Target:  "car_preference_vendor",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.car_preference_type",
			Target:  "car_preference_type",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.special_accommodation_1",
			Target:  "special_accommodation_1",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.special_accommodation_2",
			Target:  "special_accommodation_2",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.special_accommodation_3",
			Target:  "special_accommodation_3",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.special_accommodation_4",
			Target:  "special_accommodation_4",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.special_accommodation_5",
			Target:  "special_accommodation_5",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.marketing_opt_ins",
			Target:  "marketing_opt_ins",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.renew_date",
			Target:  "renew_date",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.next_bill_amount",
			Target:  "next_bill_amount",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.clear_enroll_date",
			Target:  "clear_enroll_date",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.clear_renew_date",
			Target:  "clear_renew_date",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.clear_tier_level",
			Target:  "clear_tier_level",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.clear_next_bill_amount",
			Target:  "clear_next_bill_amount",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.clear_is_active",
			Target:  "clear_is_active",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.clear_auto_renew",
			Target:  "clear_auto_renew",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.clear_has_biometrics",
			Target:  "clear_has_biometrics",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.clear_has_partner_pricing",
			Target:  "clear_has_partner_pricing",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.tsa_type",
			Target:  "tsa_type",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.tsa_seq_num",
			Target:  "tsa_seq_num",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.tsa_number",
			Target:  "tsa_number",
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
			Source: "_source.id",
			Target: "_profile.Attributes.loyalty_id",
		},
	}
}
