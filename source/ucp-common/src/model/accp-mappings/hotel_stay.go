// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	customerprofiles "tah/upt/source/storage"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

func BuildHotelStayObjectMapping() customerprofiles.ObjectMapping {
	return customerprofiles.ObjectMapping{
		Name:    constants.ACCP_RECORD_HOTEL_STAY_MAPPING,
		Version: "1.0",
		Fields:  BuildHotelStayMapping(),
	}
}

func BuildHotelStayMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
		// Profile Object Unique Key
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.id",
			Target:  "id",
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
			Source:  "_source.date",
			Target:  "date",
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
			Source:  "_source.type",
			Target:  "type",
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
			Source:  "_source.amount",
			Target:  "amount",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.currency_symbol",
			Target:  "currency_symbol",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.booking_id",
			Target:  "booking_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.description",
			Target:  "description",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.currency_name",
			Target:  "currency_name",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.currency_code",
			Target:  "currency_code",
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
			Source:  "_source.start_date",
			Target:  "start_date",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_updated_partition",
			Target:  "last_updated_partition",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.tx_id",
			Target:  "tx_id",
			KeyOnly: true,
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.extended_data",
			Target: "extended_data",
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
			Type:        "STRING",
			Source:      "_source.first_name",
			Target:      "_profile.FirstName",
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.last_name",
			Target:      "_profile.LastName",
			Searcheable: true,
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
	}
}
