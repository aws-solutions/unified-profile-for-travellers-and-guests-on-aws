// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	customerprofiles "tah/upt/source/storage"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

func BuildHotelLoyaltyObjectMapping() customerprofiles.ObjectMapping {
	return customerprofiles.ObjectMapping{
		Name:    constants.ACCP_RECORD_HOTEL_LOYALTY,
		Version: "1.0",
		Fields:  BuildHotelLoyaltyMapping(),
	}
}

func BuildHotelLoyaltyMapping() customerprofiles.FieldMappings {
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
			Source:  "_source.level",
			Target:  "level",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.program_name",
			Target:  "program_name",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.points",
			Target:  "points",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.points_to_next_level",
			Target:  "points_to_next_level",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.units",
			Target:  "units",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.joined",
			Target:  "joined",
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
			Type:        "STRING",
			Source:      "_source.traveller_id",
			Target:      "_profile.Attributes.profile_id",
			Searcheable: true,
			Indexes:     []string{"PROFILE"},
		},
		{
			Type:   "STRING",
			Source: "_source.id",
			Target: "_profile.Attributes.loyalty_id",
		},
	}
}
