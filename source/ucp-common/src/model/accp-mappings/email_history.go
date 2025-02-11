// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	customerprofiles "tah/upt/source/storage"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

func BuildEmailHistoryObjectMapping() customerprofiles.ObjectMapping {
	return customerprofiles.ObjectMapping{
		Name:    constants.ACCP_RECORD_EMAIL_HISTORY,
		Version: "1.0",
		Fields:  BuildEmailHistoryMapping(),
	}
}

func BuildEmailHistoryMapping() customerprofiles.FieldMappings {
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
			Source:  "_source.address",
			Target:  "address",
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
			Source:  "_source.last_updated",
			Target:  "last_updated",
			KeyOnly: true,
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.type",
			Target:  "type",
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
			Type:    customerprofiles.MappingTypeString,
			Source:  "_source.is_verified",
			Target:  "is_verified",
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
	}
}
